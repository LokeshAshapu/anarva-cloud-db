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
	mux.HandleFunc("/api/v1/providers", h.handleProviders)
	mux.HandleFunc("/api/v1/providers/", h.handleProviderSubroutes)
	mux.HandleFunc("/api/v1/resources/import", h.handleImportResource)
	mux.HandleFunc("/api/v1/resources/", h.handleResourceSubroutes)
	mux.HandleFunc("/api/v1/drift", h.handleDrift)
}

func (h *ProviderHandler) handleProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		providers, err := h.svc.ListProviders(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": providers})
	}
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
	if r.Method == http.MethodPost {
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
}

func (h *ProviderHandler) handleResourceSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/resources/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.Error(w, "Invalid resource action path", http.StatusBadRequest)
		return
	}

	resID := parts[0]
	action := parts[1]

	switch action {
	case "adopt":
		if r.Method == http.MethodPost {
			m, err := h.svc.AdoptResource(r.Context(), resID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": m})
		}
	case "release":
		if r.Method == http.MethodPost {
			if err := h.svc.ReleaseResource(r.Context(), resID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "RELEASED", "id": resID})
		}
	}
}

func (h *ProviderHandler) handleDrift(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
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
}
