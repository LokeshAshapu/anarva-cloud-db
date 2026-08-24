package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

type ComputeHandler struct {
	uc     *usecase.ComputeUseCase
	stream *activity.Stream
}

func NewComputeHandler(uc *usecase.ComputeUseCase, stream *activity.Stream) *ComputeHandler {
	return &ComputeHandler{
		uc:     uc,
		stream: stream,
	}
}

func (h *ComputeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/compute/plans", h.handlePlans)
	mux.HandleFunc("/api/v1/compute/images", h.handleImages)
	mux.HandleFunc("/api/v1/compute/instances", h.handleInstances)
	mux.HandleFunc("/api/v1/compute/instances/", h.handleInstanceSubroutes)
}

func (h *ComputeHandler) handlePlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	plans := h.uc.ListPlans()
	respondJSON(w, http.StatusOK, plans)
}

func (h *ComputeHandler) handleImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	images := h.uc.ListImages()
	respondJSON(w, http.StatusOK, images)
}

func (h *ComputeHandler) handleInstances(w http.ResponseWriter, r *http.Request) {
	tc := security.GetTenantContext(r.Context())

	switch r.Method {
	case http.MethodGet:
		projectID := r.URL.Query().Get("projectId")
		if tc.ProjectID != "" {
			projectID = tc.ProjectID
		} else if projectID == "" {
			projectID = "proj-default"
		}
		orgID := tc.OrganizationID
		if orgID == "" {
			orgID = "org-default"
		}

		list, err := h.uc.ListInstancesForTenant(r.Context(), orgID, projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		redactedList := make([]*domain.ComputeInstance, len(list))
		for i, item := range list {
			redactedList[i] = item.RedactSecrets()
		}
		respondJSON(w, http.StatusOK, redactedList)

	case http.MethodPost:
		var req domain.ComputeInstance
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		// Enforce TenantContext authorization on creation
		if tc.OrganizationID != "" && req.OrganizationID != "" && req.OrganizationID != tc.OrganizationID {
			respondError(w, http.StatusForbidden, fmt.Sprintf("TENANT_ISOLATION_VIOLATION: Organization '%s' is prohibited from creating compute instance for Organization '%s'", tc.OrganizationID, req.OrganizationID))
			return
		}
		if tc.ProjectID != "" && req.ProjectID != "" && req.ProjectID != tc.ProjectID {
			respondError(w, http.StatusForbidden, fmt.Sprintf("TENANT_ISOLATION_VIOLATION: Project '%s' is prohibited from creating compute instance for Project '%s'", tc.ProjectID, req.ProjectID))
			return
		}

		if tc.OrganizationID != "" {
			req.OrganizationID = tc.OrganizationID
		} else if req.OrganizationID == "" {
			req.OrganizationID = "org-default"
		}

		if tc.ProjectID != "" {
			req.ProjectID = tc.ProjectID
		} else if req.ProjectID == "" {
			req.ProjectID = "proj-default"
		}

		if req.RegionID == "" {
			req.RegionID = "us-east-1"
		}
		if req.ACU == 0 {
			req.ACU = 1.0
		}

		created, err := h.uc.CreateInstance(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		if h.stream != nil {
			h.stream.Record(&activity.ActivityEvent{
				OrganizationID: req.OrganizationID,
				ProjectID:      req.ProjectID,
				ResourceID:     created.ID,
				ActorID:        "lokeshashapu@gmail.com",
				Action:         activity.ActionComputeCreated,
				Metadata:       map[string]string{"name": created.Name},
			})
		}

		respondJSON(w, http.StatusCreated, created.RedactSecrets())

	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *ComputeHandler) handleInstanceSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/compute/instances/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		respondError(w, http.StatusBadRequest, "Instance ID required")
		return
	}

	id := parts[0]
	tc := security.GetTenantContext(r.Context())

	// Enforce TenantContext authorization BEFORE any subroute operation
	inst, err := h.uc.GetInstanceForTenant(r.Context(), tc.OrganizationID, tc.ProjectID, id)
	if err != nil {
		if strings.Contains(err.Error(), "TENANT_ISOLATION_VIOLATION") {
			respondError(w, http.StatusForbidden, err.Error())
			return
		}
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, http.StatusOK, inst.RedactSecrets())

		case http.MethodDelete:
			if err := h.uc.DeleteInstance(r.Context(), id); err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if h.stream != nil {
				h.stream.Record(&activity.ActivityEvent{
					OrganizationID: inst.OrganizationID,
					ProjectID:      inst.ProjectID,
					ResourceID:     id,
					ActorID:        "lokeshashapu@gmail.com",
					Action:         activity.ActionComputeDeleted,
				})
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	action := parts[1]
	switch action {
	case "start":
		if err := h.uc.StartInstance(r.Context(), id); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "STARTED"})

	case "stop":
		if err := h.uc.StopInstance(r.Context(), id); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "STOPPED"})

	case "restart":
		if err := h.uc.RestartInstance(r.Context(), id); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "RESTARTED"})

	case "execute":
		var req domain.CommandExecutionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}
		res, err := h.uc.ExecuteCommand(r.Context(), id, &req)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, res)

	case "metrics":
		metrics, err := h.uc.GetInstanceMetrics(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, metrics)

	default:
		respondError(w, http.StatusNotFound, "Unknown compute action")
	}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	var code string
	switch status {
	case http.StatusForbidden:
		code = "FORBIDDEN"
	case http.StatusNotFound:
		code = "NOT_FOUND"
	case http.StatusBadRequest:
		code = "BAD_REQUEST"
	default:
		code = "INTERNAL_SERVER_ERROR"
	}
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    code,
		"error":   message,
		"message": message,
	})
}
