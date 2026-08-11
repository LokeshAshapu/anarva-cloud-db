package http

import (
	"encoding/json"
	"net/http"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/provider"
)

type BackupHandler struct {
	prov   provider.BackupProvider
	stream *activity.Stream
}

func NewBackupHandler(prov provider.BackupProvider, stream *activity.Stream) *BackupHandler {
	return &BackupHandler{
		prov:   prov,
		stream: stream,
	}
}

func (h *BackupHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/databases/{id}/backups", h.ListDatabaseBackups)
	mux.HandleFunc("POST /api/v1/databases/{id}/backups", h.CreateSnapshot)
	mux.HandleFunc("GET /api/v1/backups/{id}", h.GetBackup)
	mux.HandleFunc("DELETE /api/v1/backups/{id}", h.DeleteBackup)
	mux.HandleFunc("POST /api/v1/backups/{id}/restore", h.RestoreBackup)
	mux.HandleFunc("GET /api/v1/databases/{id}/recovery-points", h.GetRecoveryPoints)
	mux.HandleFunc("GET /api/v1/databases/{id}/backup-config", h.GetBackupConfig)
}

func (h *BackupHandler) ListDatabaseBackups(w http.ResponseWriter, r *http.Request) {
	dbID := r.PathValue("id")
	backups, err := h.prov.ListBackups(r.Context(), "org-default", dbID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, backups)
}

func (h *BackupHandler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	dbID := r.PathValue("id")
	var req struct {
		Name          string `json:"name"`
		DatabaseName  string `json:"databaseName"`
		RetentionDays int    `json:"retentionDays"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid payload"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = "manual-snapshot"
	}
	if req.DatabaseName == "" {
		req.DatabaseName = "production-db"
	}

	rec := &domain.BackupRecord{
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		DatabaseID:     dbID,
		DatabaseName:   req.DatabaseName,
		Name:           req.Name,
		Type:           domain.BackupManual,
		RetentionDays:  req.RetentionDays,
		SizeBytes:      14589000,
	}

	created, err := h.prov.CreateBackup(r.Context(), rec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.stream.Record(&activity.ActivityEvent{
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		ResourceID:     created.ID,
		ActorID:        "lokeshashapu@gmail.com",
		Action:         activity.ActionBackupCompleted,
		Metadata:       map[string]string{"name": created.Name},
	})

	respondJSON(w, http.StatusCreated, created)
}

func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	b, err := h.prov.GetBackup(r.Context(), id, "org-default")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *BackupHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.prov.DeleteBackup(r.Context(), id, "org-default"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	backupID := r.PathValue("id")
	var req struct {
		TargetDBName   string `json:"targetDbName"`
		TargetRegionID string `json:"targetRegionId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.TargetDBName == "" {
		req.TargetDBName = "production-db-restored"
	}

	job, err := h.prov.CreateRestoreJob(r.Context(), &domain.RestoreJob{
		OrganizationID:   "org-default",
		ProjectID:        "proj-default",
		SourceDatabaseID: "res-db-prod-1",
		BackupID:         backupID,
		TargetDBName:     req.TargetDBName,
		TargetRegionID:   req.TargetRegionID,
		RestoreType:      "SNAPSHOT",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusAccepted, job)
}

func (h *BackupHandler) GetRecoveryPoints(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pitrStatus":        "PROVIDER_NOT_CONNECTED",
		"message":           "WAL archival & continuous recovery points require bare-metal PostgreSQL replication driver attachment",
		"availableSnapshots": 1,
	})
}

func (h *BackupHandler) GetBackupConfig(w http.ResponseWriter, r *http.Request) {
	dbID := r.PathValue("id")
	respondJSON(w, http.StatusOK, &domain.BackupConfig{
		DatabaseID:     dbID,
		Enabled:        true,
		RetentionDays:  7,
		BackupWindow:   "02:00 UTC - 03:00 UTC",
		PitrEnabled:    false,
		ProviderStatus: "CONFIGURED",
	})
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
