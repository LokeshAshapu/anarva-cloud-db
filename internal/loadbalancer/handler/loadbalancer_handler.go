package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/service"
)

type LoadBalancerHandler struct {
	svc *service.LoadBalancerService
}

func NewLoadBalancerHandler(svc *service.LoadBalancerService) *LoadBalancerHandler {
	return &LoadBalancerHandler{svc: svc}
}

func (h *LoadBalancerHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/load-balancers", h.handleLoadBalancers)
	mux.HandleFunc("/api/v1/load-balancers/", h.handleLoadBalancerSubroutes)
	mux.HandleFunc("/api/v1/domains", h.handleDomains)
	mux.HandleFunc("/api/v1/certificates", h.handleCertificates)
	mux.HandleFunc("/api/v1/applications", h.handleApplications)
	mux.HandleFunc("/api/v1/applications/", h.handleApplicationSubroutes)
}

func (h *LoadBalancerHandler) handleLoadBalancers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		orgID := r.URL.Query().Get("organizationId")
		projectID := r.URL.Query().Get("projectId")
		lbs, err := h.svc.ListLoadBalancers(r.Context(), orgID, projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": lbs,
			"meta": map[string]interface{}{"count": len(lbs)},
		})

	case http.MethodPost:
		var req struct {
			OrganizationID string          `json:"organizationId"`
			ProjectID      string          `json:"projectId"`
			Name           string          `json:"name"`
			Type           domain.LBType   `json:"type"`
			Scheme         domain.LBScheme `json:"scheme"`
			NetworkID      string          `json:"networkId"`
			SubnetIDs      []string        `json:"subnetIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Type == "" {
			req.Type = domain.LBTypeApplication
		}
		if req.Scheme == "" {
			req.Scheme = domain.LBSchemePublic
		}

		lb, err := h.svc.CreateLoadBalancer(r.Context(), req.OrganizationID, req.ProjectID, req.Name, req.Type, req.Scheme, req.NetworkID, req.SubnetIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": lb})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *LoadBalancerHandler) handleLoadBalancerSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/load-balancers/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid load balancer ID", http.StatusBadRequest)
		return
	}

	lbID := parts[0]
	if r.Method == http.MethodGet {
		lb, err := h.svc.GetLoadBalancer(r.Context(), lbID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": lb})
	} else if r.Method == http.MethodDelete {
		if err := h.svc.DeleteLoadBalancer(r.Context(), lbID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "DELETED", "id": lbID})
	}
}

func (h *LoadBalancerHandler) handleApplications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		orgID := r.URL.Query().Get("organizationId")
		projectID := r.URL.Query().Get("projectId")
		apps, err := h.svc.ListApplications(r.Context(), orgID, projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": apps,
			"meta": map[string]interface{}{"count": len(apps)},
		})

	case http.MethodPost:
		var req struct {
			OrganizationID string `json:"organizationId"`
			ProjectID      string `json:"projectId"`
			Name           string `json:"name"`
			ContainerImage string `json:"containerImage"`
			NetworkID      string `json:"networkId"`
			DomainName     string `json:"domainName"`
			ACUCount       int    `json:"acuCount"`
			Port           int    `json:"port"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		app, err := h.svc.DeployApplication(r.Context(), req.OrganizationID, req.ProjectID, req.Name, req.ContainerImage, req.NetworkID, req.DomainName, req.ACUCount, req.Port)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": app})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *LoadBalancerHandler) handleApplicationSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/applications/")
	parts := strings.Split(path, "/")

	if len(parts) > 1 && parts[1] == "health" && r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "HEALTHY",
			"details": map[string]string{"lb": "HEALTHY", "container": "RUNNING", "db": "CONNECTED"},
		})
	}
}

func (h *LoadBalancerHandler) handleDomains(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		var req struct {
			OrganizationID string `json:"organizationId"`
			ProjectID      string `json:"projectId"`
			Name           string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		dom, err := h.svc.CreateDomain(req.OrganizationID, req.ProjectID, req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": dom})
	}
}

func (h *LoadBalancerHandler) handleCertificates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		var req struct {
			OrganizationID string `json:"organizationId"`
			ProjectID      string `json:"projectId"`
			Domain         string `json:"domain"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		cert, err := h.svc.RequestCertificate(req.OrganizationID, req.ProjectID, req.Domain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": cert})
	}
}
