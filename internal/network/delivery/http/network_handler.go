package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/network/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/network/usecase"
)

type NetworkHandler struct {
	uc     *usecase.NetworkUseCase
	stream *activity.Stream
}

func NewNetworkHandler(uc *usecase.NetworkUseCase, stream *activity.Stream) *NetworkHandler {
	return &NetworkHandler{
		uc:     uc,
		stream: stream,
	}
}

func (h *NetworkHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/networks", h.handleNetworks)
	mux.HandleFunc("/api/v1/networks/", h.handleNetworkSubroutes)
}

func (h *NetworkHandler) handleNetworks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		projectID := r.URL.Query().Get("projectId")
		if projectID == "" {
			projectID = "proj-default"
		}
		list, err := h.uc.ListNetworks(r.Context(), projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{"data": list})

	case http.MethodPost:
		var req domain.Network
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

		created, err := h.uc.CreateNetwork(r.Context(), &req)
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
				Action:         activity.ActionNetworkCreated,
				Metadata:       map[string]string{"name": created.Name, "cidr": created.CIDR},
			})
		}

		respondJSON(w, http.StatusCreated, map[string]interface{}{"data": created})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *NetworkHandler) handleNetworkSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/networks/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Network ID required", http.StatusBadRequest)
		return
	}

	id := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			net, err := h.uc.GetNetwork(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			respondJSON(w, http.StatusOK, net)

		case http.MethodDelete:
			if err := h.uc.DeleteNetwork(r.Context(), id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if h.stream != nil {
				h.stream.Record(&activity.ActivityEvent{
					OrganizationID: "org-default",
					ProjectID:      "proj-default",
					ResourceID:     id,
					ActorID:        "lokeshashapu@gmail.com",
					Action:         activity.ActionNetworkDeleted,
				})
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	subroute := parts[1]
	switch subroute {
	case "subnets":
		if r.Method == http.MethodPost {
			var sub domain.Subnet
			if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
				http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
				return
			}
			sub.NetworkID = id
			created, err := h.uc.CreateSubnet(r.Context(), &sub)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			respondJSON(w, http.StatusCreated, created)
		} else {
			respondJSON(w, http.StatusOK, []interface{}{})
		}

	case "dns":
		zones, _ := h.uc.ListDNSZones(r.Context(), "proj-default")
		respondJSON(w, http.StatusOK, zones)

	case "load-balancers":
		if r.Method == http.MethodPost {
			var lb domain.LoadBalancer
			if err := json.NewDecoder(r.Body).Decode(&lb); err != nil {
				http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
				return
			}
			lb.NetworkID = id
			created, err := h.uc.CreateLoadBalancer(r.Context(), &lb)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			respondJSON(w, http.StatusCreated, created)
		} else {
			respondJSON(w, http.StatusOK, []interface{}{})
		}

	default:
		http.Error(w, "Unknown network route", http.StatusNotFound)
	}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
