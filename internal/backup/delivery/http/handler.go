package http

import (
	"encoding/json"
	"net/http"

	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/usecase"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
)

type BackupHandler struct {
	useCase usecase.BackupUseCase
}

func NewBackupHandler(useCase usecase.BackupUseCase) *BackupHandler {
	return &BackupHandler{useCase: useCase}
}

type createBackupReq struct {
	DatabaseID string            `json:"database_id"`
	ProjectID  string            `json:"project_id"`
	Name       string            `json:"name"`
	Type       domain.BackupType `json:"type"`
}

type restoreBackupReq struct {
	TargetDatabaseID string `json:"target_database_id"`
}

func (h *BackupHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/backups", h.CreateBackup)
	mux.HandleFunc("GET /api/v1/backups/{id}", h.GetBackup)
	mux.HandleFunc("GET /api/v1/databases/{database_id}/backups", h.ListBackups)
	mux.HandleFunc("POST /api/v1/backups/{id}/restore", h.RestoreBackup)
	mux.HandleFunc("DELETE /api/v1/backups/{id}", h.DeleteBackup)
}

func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	var req createBackupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request payload"))
		return
	}

	snapshot, err := h.useCase.CreateBackup(r.Context(), req.DatabaseID, req.ProjectID, req.Name, req.Type)
	if err != nil {
		respondError(w, err)
		return
	}

	metrics.RecordHTTPRequest(http.StatusCreated, r.Method, r.URL.Path, 0)
	respondJSON(w, http.StatusCreated, snapshot)
}

func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	snapshot, err := h.useCase.GetBackup(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, snapshot)
}

func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	databaseID := r.PathValue("database_id")
	backups, err := h.useCase.ListBackups(r.Context(), databaseID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"backups": backups})
}

func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req restoreBackupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, appErrors.New(appErrors.CodeInvalidInput, "invalid request payload"))
		return
	}

	if err := h.useCase.RestoreBackup(r.Context(), id, req.TargetDatabaseID); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *BackupHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.useCase.DeleteBackup(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
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
