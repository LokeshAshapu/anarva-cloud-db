package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/service"
)

type NetworkingHandler struct {
	svc *service.NetworkingService
}

func NewNetworkingHandler(svc *service.NetworkingService) *NetworkingHandler {
	return &NetworkingHandler{svc: svc}
}

func (h *NetworkingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/networks", h.handleNetworks)
	mux.HandleFunc("/api/v1/networks/", h.handleNetworkSubroutes)
	mux.HandleFunc("/api/v1/subnets", h.handleSubnets)
	mux.HandleFunc("/api/v1/security-groups", h.handleSecurityGroups)
	mux.HandleFunc("/api/v1/security-groups/", h.handleSecurityGroupSubroutes)
	mux.HandleFunc("/api/v1/route-tables", h.handleRouteTables)
	mux.HandleFunc("/api/v1/network/connectivity-tests", h.handleConnectivityTests)
}

func (h *NetworkingHandler) handleNetworks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		orgID := r.URL.Query().Get("organizationId")
		projectID := r.URL.Query().Get("projectId")
		if orgID == "" {
			orgID = "org-default"
		}
		if projectID == "" {
			projectID = "proj-default"
		}
		nets, err := h.svc.ListNetworks(r.Context(), orgID, projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": nets,
			"meta": map[string]interface{}{"count": len(nets)},
		})

	case http.MethodPost:
		var req struct {
			OrganizationID string `json:"organizationId"`
			ProjectID      string `json:"projectId"`
			Name           string `json:"name"`
			RegionID       string `json:"regionId"`
			CIDR           string `json:"cidr"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OrganizationID == "" {
			req.OrganizationID = "org-default"
		}
		if req.ProjectID == "" {
			req.ProjectID = "proj-default"
		}

		vNet, err := h.svc.CreateNetwork(r.Context(), req.OrganizationID, req.ProjectID, req.Name, req.RegionID, req.CIDR)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": vNet})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *NetworkingHandler) handleNetworkSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/networks/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid network ID", http.StatusBadRequest)
		return
	}

	netID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		if r.Method == http.MethodGet {
			vNet, err := h.svc.GetNetwork(r.Context(), netID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": vNet})
		} else if r.Method == http.MethodDelete {
			if err := h.svc.DeleteNetwork(r.Context(), netID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "DELETED", "id": netID})
		}

	case "subnets":
		if r.Method == http.MethodGet {
			subnets, err := h.svc.ListSubnets(r.Context(), netID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": subnets})
		}

	default:
		http.Error(w, "Subroute not found", http.StatusNotFound)
	}
}

func (h *NetworkingHandler) handleSubnets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		vpcID := r.URL.Query().Get("vpcId")
		subnets, err := h.svc.ListSubnets(r.Context(), vpcID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": subnets})

	case http.MethodPost:
		var req struct {
			OrganizationID string           `json:"organizationId"`
			ProjectID      string           `json:"projectId"`
			VPCID          string           `json:"vpcId"`
			Name           string           `json:"name"`
			CIDR           string           `json:"cidr"`
			Zone           string           `json:"zone"`
			Type           domain.SubnetType `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OrganizationID == "" {
			req.OrganizationID = "org-default"
		}
		if req.ProjectID == "" {
			req.ProjectID = "proj-default"
		}
		if req.Type == "" {
			req.Type = domain.SubnetPrivate
		}

		sn, err := h.svc.CreateSubnet(r.Context(), req.OrganizationID, req.ProjectID, req.VPCID, req.Name, req.CIDR, req.Zone, req.Type)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": sn})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *NetworkingHandler) handleSecurityGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		vpcID := r.URL.Query().Get("vpcId")
		sgs, err := h.svc.ListSecurityGroups(r.Context(), vpcID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": sgs})

	case http.MethodPost:
		var req struct {
			OrganizationID string `json:"organizationId"`
			ProjectID      string `json:"projectId"`
			VPCID          string `json:"vpcId"`
			Name           string `json:"name"`
			Description    string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OrganizationID == "" {
			req.OrganizationID = "org-default"
		}
		if req.ProjectID == "" {
			req.ProjectID = "proj-default"
		}

		sg, err := h.svc.CreateSecurityGroup(r.Context(), req.OrganizationID, req.ProjectID, req.VPCID, req.Name, req.Description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": sg})
	}
}

func (h *NetworkingHandler) handleSecurityGroupSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/security-groups/")
	parts := strings.Split(path, "/")

	if len(parts) > 1 && parts[1] == "rules" && r.Method == http.MethodPost {
		sgID := parts[0]
		var rule domain.SecurityRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sg, err := h.svc.AddSecurityRule(r.Context(), sgID, rule)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": sg})
	}
}

func (h *NetworkingHandler) handleRouteTables(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		vpcID := r.URL.Query().Get("vpcId")
		rts, err := h.svc.ListRouteTables(r.Context(), vpcID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": rts})

	case http.MethodPost:
		var req struct {
			OrganizationID string `json:"organizationId"`
			ProjectID      string `json:"projectId"`
			VPCID          string `json:"vpcId"`
			Name           string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.OrganizationID == "" {
			req.OrganizationID = "org-default"
		}
		if req.ProjectID == "" {
			req.ProjectID = "proj-default"
		}

		rt, err := h.svc.CreateRouteTable(r.Context(), req.OrganizationID, req.ProjectID, req.VPCID, req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": rt})
	}
}

func (h *NetworkingHandler) handleConnectivityTests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var req struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
			Port        int    `json:"port"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Port == 0 {
			req.Port = 80
		}

		res, err := h.svc.TestConnectivity(r.Context(), req.Source, req.Destination, req.Port)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "data": res})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": res})
	}
}
