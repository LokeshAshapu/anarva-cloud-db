package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/provisioning/usecase"
)

type ProvisioningHandler struct {
	uc       *usecase.ProvisioningUseCase
	registry *provider.ProviderRegistry
	stream   *activity.Stream
}

func NewProvisioningHandler(uc *usecase.ProvisioningUseCase, registry *provider.ProviderRegistry, stream *activity.Stream) *ProvisioningHandler {
	return &ProvisioningHandler{
		uc:       uc,
		registry: registry,
		stream:   stream,
	}
}

func (h *ProvisioningHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/provisioning/plan", h.handlePlan)
	mux.HandleFunc("/api/v1/provisioning/apply", h.handleApply)
	mux.HandleFunc("/api/v1/provisioning/requests", h.handleRequests)
	mux.HandleFunc("/api/v1/provisioning/requests/", h.handleRequestByID)
	mux.HandleFunc("/api/v1/providers", h.handleProviders)
	mux.HandleFunc("/api/v1/resources/", h.handleResourceSubroutes)
}

func (h *ProvisioningHandler) handlePlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.ProvisioningRequest
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
	if req.RequestedBy == "" {
		req.RequestedBy = "lokeshashapu@gmail.com"
	}

	plan, err := h.uc.CreatePlan(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.stream != nil {
		h.stream.Record(&activity.ActivityEvent{
			OrganizationID: req.OrganizationID,
			ProjectID:      req.ProjectID,
			ResourceID:     plan.ResourceID,
			ActorID:        req.RequestedBy,
			Action:         activity.ActionProvisioningPlanCreated,
			Metadata:       map[string]string{"provider": plan.Provider, "resourceType": string(plan.ResourceType)},
		})
	}

	respondJSON(w, http.StatusOK, plan)
}

func (h *ProvisioningHandler) handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.RequestID == "" {
		http.Error(w, "requestId is required", http.StatusBadRequest)
		return
	}

	res, err := h.uc.ApplyRequest(r.Context(), payload.RequestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.stream != nil {
		h.stream.Record(&activity.ActivityEvent{
			OrganizationID: res.OrganizationID,
			ProjectID:      res.ProjectID,
			ResourceID:     res.ResourceID,
			ActorID:        res.RequestedBy,
			Action:         activity.ActionProvisioningCompleted,
			Metadata:       map[string]string{"status": string(res.Status)},
		})
	}

	respondJSON(w, http.StatusOK, res)
}

func (h *ProvisioningHandler) handleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	projectID := r.URL.Query().Get("projectId")
	list := h.uc.ListRequests(projectID)
	respondJSON(w, http.StatusOK, list)
}

func (h *ProvisioningHandler) handleRequestByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/provisioning/requests/")
	if id == "" {
		http.Error(w, "Request ID required", http.StatusBadRequest)
		return
	}
	req, err := h.uc.GetRequest(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, req)
}

func (h *ProvisioningHandler) handleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	caps := h.registry.GetCapabilities("LOCAL_DOCKER")
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"provider":     "LOCAL_DOCKER",
		"realityLabel": "LOCAL DEVELOPMENT PROVIDER",
		"capabilities": caps,
	})
}

func (h *ProvisioningHandler) handleResourceSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/resources/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.Error(w, "Resource route invalid", http.StatusBadRequest)
		return
	}

	resourceID := parts[0]
	subroute := parts[1]

	switch subroute {
	case "drift":
		drift, err := h.uc.GetDrift(resourceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		respondJSON(w, http.StatusOK, drift)

	case "reconcile":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		reconciled, err := h.uc.ReconcileResource(r.Context(), resourceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if h.stream != nil {
			h.stream.Record(&activity.ActivityEvent{
				OrganizationID: "org-default",
				ProjectID:      "proj-default",
				ResourceID:     resourceID,
				ActorID:        "lokeshashapu@gmail.com",
				Action:         activity.ActionResourceReconciled,
			})
		}

		respondJSON(w, http.StatusOK, reconciled)

	default:
		http.Error(w, "Subroute not found", http.StatusNotFound)
	}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
