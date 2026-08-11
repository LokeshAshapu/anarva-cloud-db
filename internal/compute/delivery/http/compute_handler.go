package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
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
	switch r.Method {
	case http.MethodGet:
		projectID := r.URL.Query().Get("projectId")
		if projectID == "" {
			projectID = "proj-default"
		}
		list, err := h.uc.ListInstances(r.Context(), projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, list)

	case http.MethodPost:
		var req domain.ComputeInstance
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if req.OrganizationID == "" {
			req.OrganizationID = "org-default"
		}
		if req.ProjectID == "" {
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
			http.Error(w, err.Error(), http.StatusBadRequest)
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

		respondJSON(w, http.StatusCreated, created)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ComputeHandler) handleInstanceSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/compute/instances/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	id := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			inst, err := h.uc.GetInstance(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			respondJSON(w, http.StatusOK, inst)

		case http.MethodDelete:
			if err := h.uc.DeleteInstance(r.Context(), id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if h.stream != nil {
				h.stream.Record(&activity.ActivityEvent{
					OrganizationID: "org-default",
					ProjectID:      "proj-default",
					ResourceID:     id,
					ActorID:        "lokeshashapu@gmail.com",
					Action:         activity.ActionComputeDeleted,
				})
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	action := parts[1]
	switch action {
	case "start":
		if err := h.uc.StartInstance(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "STARTED"})

	case "stop":
		if err := h.uc.StopInstance(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "STOPPED"})

	case "restart":
		if err := h.uc.RestartInstance(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "RESTARTED"})

	case "execute":
		var req domain.CommandExecutionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
		res, err := h.uc.ExecuteCommand(r.Context(), id, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		respondJSON(w, http.StatusOK, res)

	case "metrics":
		metrics, err := h.uc.GetInstanceMetrics(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, metrics)

	default:
		http.Error(w, "Unknown compute action", http.StatusNotFound)
	}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
