package http

import (
	"encoding/json"
	"net/http"

	"github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/database/usecase"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
)

type DatabaseHandler struct {
	useCase usecase.DatabaseUseCase
}

func NewDatabaseHandler(useCase usecase.DatabaseUseCase) *DatabaseHandler {
	return &DatabaseHandler{useCase: useCase}
}

type createDBReq struct {
	ProjectID string            `json:"project_id"`
	Name      string            `json:"name"`
	Engine    domain.EngineType `json:"engine"`
	StorageGB int               `json:"storage_gb"`
}

func (h *DatabaseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/databases", h.CreateDatabase)
	mux.HandleFunc("GET /api/v1/databases/{id}", h.GetDatabase)
	mux.HandleFunc("GET /api/v1/projects/{project_id}/databases", h.ListDatabases)
	mux.HandleFunc("POST /api/v1/databases/{id}/start", h.StartDatabase)
	mux.HandleFunc("POST /api/v1/databases/{id}/stop", h.StopDatabase)
	mux.HandleFunc("DELETE /api/v1/databases/{id}", h.DeleteDatabase)
	mux.HandleFunc("GET /api/v1/databases/{id}/connection-string", h.GetConnectionString)
}

func (h *DatabaseHandler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req createDBReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request body"))
		return
	}

	instance, connStr, err := h.useCase.CreateDatabase(r.Context(), req.ProjectID, req.Name, req.Engine, req.StorageGB)
	if err != nil {
		respondError(w, err)
		return
	}

	metrics.RecordHTTPRequest(http.StatusCreated, r.Method, r.URL.Path, 0)
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"instance":          instance,
		"connection_string": connStr,
	})
}

func (h *DatabaseHandler) GetDatabase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	instance, err := h.useCase.GetDatabase(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, instance)
}

func (h *DatabaseHandler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	instances, err := h.useCase.ListDatabases(r.Context(), projectID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"databases": instances})
}

func (h *DatabaseHandler) StartDatabase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.useCase.StartDatabase(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *DatabaseHandler) StopDatabase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.useCase.StopDatabase(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *DatabaseHandler) DeleteDatabase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.useCase.DeleteDatabase(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *DatabaseHandler) GetConnectionString(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	connStr, err := h.useCase.GetConnectionString(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"connection_string": connStr})
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*appErrors.AppError); ok {
		respondJSON(w, appErr.HTTPStatusCode(), appErr)
		return
	}
	respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
