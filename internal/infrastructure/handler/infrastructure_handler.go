package handler

import (
	"encoding/json"
	"net/http"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/service"
)

type InfrastructureHandler struct {
	svc *service.InfrastructureService
}

func NewInfrastructureHandler(svc *service.InfrastructureService) *InfrastructureHandler {
	return &InfrastructureHandler{svc: svc}
}

func (h *InfrastructureHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/regions", h.handleRegions)
	mux.HandleFunc("/api/v1/infrastructure/global-health", h.handleGlobalHealth)
	mux.HandleFunc("/api/v1/failover/execute", h.handleExecuteFailover)
	mux.HandleFunc("/api/v1/incidents", h.handleIncidents)
	mux.HandleFunc("/api/v1/infrastructure/simulate-outage", h.handleSimulateOutage)
}

func (h *InfrastructureHandler) handleRegions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	regions, err := h.svc.ListRegions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": regions})
}

func (h *InfrastructureHandler) handleGlobalHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	hlth, err := h.svc.GetGlobalHealth(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": hlth})
}

func (h *InfrastructureHandler) handleExecuteFailover(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		var policy domain.FailoverPolicy
		if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		plan, err := h.svc.ExecuteFailover(r.Context(), &policy)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": plan})
	}
}

func (h *InfrastructureHandler) handleIncidents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	incidents, err := h.svc.ListIncidents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": incidents})
}

func (h *InfrastructureHandler) handleSimulateOutage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		var req struct {
			RegionID string `json:"regionId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.RegionID == "" {
			req.RegionID = "ap-hyderabad-1"
		}
		inc, err := h.svc.SimulateOutage(req.RegionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": inc})
	}
}
