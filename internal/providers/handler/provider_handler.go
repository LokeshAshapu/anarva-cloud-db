package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/service"
)

type ProviderHandler struct {
	svc *service.ProviderService
}

func NewProviderHandler(svc *service.ProviderService) *ProviderHandler {
	return &ProviderHandler{svc: svc}
}

func (h *ProviderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/providers", h.handleProviders)
	mux.HandleFunc("GET /api/v1/providers/", h.handleProviderSubroutes)
	mux.HandleFunc("POST /api/v1/providers/", h.handleProviderSubroutes)
	mux.HandleFunc("POST /api/v1/providers/resources/import", h.handleImportResource)
	mux.HandleFunc("POST /api/v1/providers/resources/adopt", h.handleAdoptResource)
	mux.HandleFunc("POST /api/v1/providers/resources/release", h.handleReleaseResource)
	mux.HandleFunc("POST /api/v1/providers/drift", h.handleDrift)
}

func (h *ProviderHandler) handleProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	providers, err := h.svc.ListProviders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": providers})
}

func (h *ProviderHandler) handleProviderSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/providers/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	providerID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		p, err := h.svc.GetProvider(r.Context(), providerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": p})

	case "verify":
		if r.Method == http.MethodPost {
			var req struct {
				CredentialReference string `json:"credentialReference"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			p, err := h.svc.VerifyProvider(r.Context(), providerID, req.CredentialReference)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": p})
		}
	}
}

func (h *ProviderHandler) handleImportResource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Provider           string `json:"provider"`
		ProviderResourceID string `json:"providerResourceId"`
		ResourceType       string `json:"resourceType"`
		Region             string `json:"region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := h.svc.ImportResource(r.Context(), req.Provider, req.ProviderResourceID, req.ResourceType, req.Region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": m})
}

func (h *ProviderHandler) handleAdoptResource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		AnarvaResourceID string `json:"anarvaResourceId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	m, err := h.svc.AdoptResource(r.Context(), req.AnarvaResourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": m})
}

func (h *ProviderHandler) handleReleaseResource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		AnarvaResourceID string `json:"anarvaResourceId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.ReleaseResource(r.Context(), req.AnarvaResourceID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "RELEASED", "id": req.AnarvaResourceID})
}

func (h *ProviderHandler) handleDrift(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		AnarvaResourceID string `json:"anarvaResourceId"`
		DesiredState     string `json:"desiredState"`
		ObservedState    string `json:"observedState"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	d, err := h.svc.DetectDrift(r.Context(), req.AnarvaResourceID, req.DesiredState, req.ObservedState)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"data": d})
}
