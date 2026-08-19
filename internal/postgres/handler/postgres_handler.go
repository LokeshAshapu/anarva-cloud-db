package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
)

type PostgresHandler struct {
	postgresService *service.PostgresService
	sqlService      *service.SQLService
}

func NewPostgresHandler(ps *service.PostgresService, ss *service.SQLService) *PostgresHandler {
	return &PostgresHandler{
		postgresService: ps,
		sqlService:      ss,
	}
}

func (h *PostgresHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/databases", h.handleDatabases)
	mux.HandleFunc("/api/v1/databases/", h.handleDatabaseSubroutes)
}

func (h *PostgresHandler) handleDatabases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		orgID := r.URL.Query().Get("organizationId")
		projectID := r.URL.Query().Get("projectId")
		instances, err := h.postgresService.ListInstances(r.Context(), orgID, projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": instances,
			"meta": map[string]interface{}{"count": len(instances)},
		})

	case http.MethodPost:
		var req struct {
			OrganizationID string  `json:"organizationId"`
			ProjectID      string  `json:"projectId"`
			Name           string  `json:"name"`
			Version        string  `json:"version"`
			RegionID       string  `json:"regionId"`
			NetworkID      string  `json:"networkId"`
			CPU            float64 `json:"cpu"`
			MemoryMB       int     `json:"memoryMb"`
			StorageGB      int     `json:"storageGb"`
			PublicAccess   bool    `json:"publicAccess"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		inst, err := h.postgresService.CreateInstance(r.Context(), req.OrganizationID, req.ProjectID, req.Name, req.Version, req.RegionID, req.NetworkID, req.CPU, req.MemoryMB, req.StorageGB, req.PublicAccess)
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

func (h *PostgresHandler) handleDatabaseSubroutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/databases/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid database instance ID", http.StatusBadRequest)
		return
	}

	instanceID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		if r.Method == http.MethodGet {
			inst, err := h.postgresService.GetInstance(r.Context(), instanceID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": inst})
		} else if r.Method == http.MethodDelete {
			if err := h.postgresService.DeleteInstance(r.Context(), instanceID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "DELETED", "id": instanceID})
		}

	case "start":
		if err := h.postgresService.StartInstance(r.Context(), instanceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "STARTED", "id": instanceID})

	case "stop":
		if err := h.postgresService.StopInstance(r.Context(), instanceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "STOPPED", "id": instanceID})

	case "restart":
		if err := h.postgresService.RestartInstance(r.Context(), instanceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "RESTARTED", "id": instanceID})

	case "health":
		health, err := h.postgresService.GetHealth(r.Context(), instanceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": health})

	case "metrics":
		metrics, err := h.postgresService.GetMetrics(r.Context(), instanceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": metrics})

	case "logs":
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		logs, err := h.postgresService.GetLogs(r.Context(), instanceID, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": logs})

	case "connection":
		conn, err := h.postgresService.GetConnectionInfo(r.Context(), instanceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": conn})

	case "test-connection":
		res, err := h.postgresService.TestConnection(r.Context(), instanceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": res})

	case "branch":
		var req struct {
			BranchID string `json:"branchId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BranchID == "" {
			http.Error(w, "missing branchId", http.StatusBadRequest)
			return
		}
		res, err := h.sqlService.BranchDatabase(instanceID, req.BranchID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": res})

	case "sql", "query":
		var req struct {
			SQL   string `json:"sql"`
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondStructuredError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		sqlText := req.SQL
		if sqlText == "" {
			sqlText = req.Query
		}

		if sqlText == "" {
			respondStructuredError(w, http.StatusBadRequest, "EMPTY_QUERY", "empty SQL query statement", r.Header.Get("X-Request-ID"))
			return
		}

		res, err := h.sqlService.ExecuteQuery(r.Context(), instanceID, sqlText)
		if err != nil {
			respondStructuredError(w, http.StatusBadRequest, "SQL_EXECUTION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": res})

	default:
		respondStructuredError(w, http.StatusNotFound, "NOT_FOUND", "Subroute not found", r.Header.Get("X-Request-ID"))
	}
}

func respondStructuredError(w http.ResponseWriter, status int, code, msg, reqID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if reqID == "" {
		reqID = "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":      msg,
		"code":       code,
		"request_id": reqID,
		"details":    msg,
	})
}
