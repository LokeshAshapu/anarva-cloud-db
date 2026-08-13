package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/mysql/service"
)

type MySQLHandler struct {
	svc    *service.MySQLService
	sqlSvc *service.SQLService
}

func NewMySQLHandler(svc *service.MySQLService, sqlSvc *service.SQLService) *MySQLHandler {
	return &MySQLHandler{svc: svc, sqlSvc: sqlSvc}
}

func (h *MySQLHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/mysql/databases", h.handleDatabases)
	mux.HandleFunc("/api/v1/mysql/databases/", h.handleDatabaseSubroutes)
}

func (h *MySQLHandler) handleDatabases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		orgID := r.URL.Query().Get("organizationId")
		projectID := r.URL.Query().Get("projectId")
		insts, err := h.svc.ListInstances(r.Context(), orgID, projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": insts,
			"meta": map[string]interface{}{"count": len(insts)},
		})

	case http.MethodPost:
		var req struct {
			OrganizationID string `json:"organizationId"`
			ProjectID      string `json:"projectId"`
			Name           string `json:"name"`
			Version        string `json:"version"`
			RegionID       string `json:"regionId"`
			NetworkID      string `json:"networkId"`
			ACUCount       int    `json:"acuCount"`
			StorageGB      int    `json:"storageGb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.ACUCount == 0 {
			req.ACUCount = 2
		}
		if req.StorageGB == 0 {
			req.StorageGB = 20
		}

		inst, err := h.svc.CreateInstance(r.Context(), req.OrganizationID, req.ProjectID, req.Name, req.Version, req.RegionID, req.NetworkID, req.ACUCount, req.StorageGB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": inst})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *MySQLHandler) handleDatabaseSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mysql/databases/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid instance ID", http.StatusBadRequest)
		return
	}

	instID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		if r.Method == http.MethodGet {
			inst, err := h.svc.GetInstance(r.Context(), instID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": inst})
		} else if r.Method == http.MethodDelete {
			if err := h.svc.DeleteInstance(r.Context(), instID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "DELETED", "id": instID})
		}

	case "health":
		hlth, err := h.svc.GetHealth(r.Context(), instID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": hlth})

	case "connection":
		conn, err := h.svc.GetConnectionInfo(r.Context(), instID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": conn})

	case "query":
		if r.Method == http.MethodPost {
			var req struct {
				SQL string `json:"sql"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			res, err := h.sqlSvc.ExecuteQuery(r.Context(), req.SQL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": res})
		}

	default:
		http.Error(w, "Subroute not found", http.StatusNotFound)
	}
}
