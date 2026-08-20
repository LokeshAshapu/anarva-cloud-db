package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/activity"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/backup/usecase"
)

type BackupHandler struct {
	prov   provider.BackupProvider
	uc     usecase.BackupUseCase
	stream *activity.Stream
}

func NewBackupHandler(prov provider.BackupProvider, stream *activity.Stream) *BackupHandler {
	return &BackupHandler{
		prov:   prov,
		stream: stream,
	}
}

func NewBackupHandlerWithUseCase(prov provider.BackupProvider, uc usecase.BackupUseCase, stream *activity.Stream) *BackupHandler {
	return &BackupHandler{
		prov:   prov,
		uc:     uc,
		stream: stream,
	}
}

func (h *BackupHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/databases/{id}/backups", h.ListDatabaseBackups)
	mux.HandleFunc("POST /api/v1/databases/{id}/backups", h.CreateSnapshot)
	mux.HandleFunc("GET /api/v1/backups/{id}", h.GetBackup)
	mux.HandleFunc("GET /api/v1/backups/{id}/download", h.DownloadBackupArchive)
	mux.HandleFunc("DELETE /api/v1/backups/{id}", h.DeleteBackup)
	mux.HandleFunc("POST /api/v1/backups/{id}/restore", h.RestoreBackup)
	mux.HandleFunc("GET /api/v1/databases/{id}/recovery-points", h.GetRecoveryPoints)
	mux.HandleFunc("GET /api/v1/databases/{id}/backup-config", h.GetBackupConfig)
}

func (h *BackupHandler) ListDatabaseBackups(w http.ResponseWriter, r *http.Request) {
	dbID := r.PathValue("id")
	orgID := getOrgIDFromContext(r)
	backups, err := h.prov.ListBackups(r.Context(), orgID, dbID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, backups)
}

func (h *BackupHandler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	dbID := r.PathValue("id")
	orgID := getOrgIDFromContext(r)
	projID := r.Header.Get("X-Project-ID")
	if projID == "" {
		projID = "proj-default"
	}

	var req struct {
		Name          string `json:"name"`
		DatabaseName  string `json:"databaseName"`
		RetentionDays int    `json:"retentionDays"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
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
		OrganizationID: orgID,
		ProjectID:      projID,
		DatabaseID:     dbID,
		DatabaseName:   req.DatabaseName,
		Name:           req.Name,
		Type:           domain.BackupManual,
		RetentionDays:  req.RetentionDays,
	}

	created, err := h.prov.CreateBackup(r.Context(), rec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.stream != nil {
		h.stream.Record(&activity.ActivityEvent{
			OrganizationID: orgID,
			ProjectID:      projID,
			ResourceID:     created.ID,
			ActorID:        "operator@anarva.cloud",
			Action:         activity.ActionBackupCompleted,
			Metadata:       map[string]string{"name": created.Name},
		})
	}

	respondJSON(w, http.StatusCreated, created)
}

func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	orgID := getOrgIDFromContext(r)
	b, err := h.prov.GetBackup(r.Context(), id, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "authorization violation") {
			http.Error(w, `{"error":"Forbidden: cross-tenant access denied"}`, http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *BackupHandler) DownloadBackupArchive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	orgID := getOrgIDFromContext(r)

	if h.uc == nil {
		http.Error(w, `{"error":"Direct archive download requires BackupUseCase streaming engine"}`, http.StatusNotImplemented)
		return
	}

	stream, _, err := h.uc.RestoreBackup(r.Context(), orgID, id, snapshotToTargetDB(id))
	if err != nil {
		if strings.Contains(err.Error(), "authorization violation") {
			http.Error(w, `{"error":"Forbidden: cross-tenant access denied"}`, http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "application/x-tar")
	w.Header().Set("Content-Disposition", "attachment; filename="+id+".dump")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, stream)
}

func (h *BackupHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	orgID := getOrgIDFromContext(r)

	if err := h.prov.DeleteBackup(r.Context(), id, orgID); err != nil {
		if strings.Contains(err.Error(), "authorization violation") {
			http.Error(w, `{"error":"Forbidden: cross-tenant access denied"}`, http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	backupID := r.PathValue("id")
	orgID := getOrgIDFromContext(r)

	var req struct {
		TargetDBName   string `json:"targetDbName"`
		TargetRegionID string `json:"targetRegionId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.TargetDBName == "" {
		req.TargetDBName = "production-db-restored"
	}

	job, err := h.prov.CreateRestoreJob(r.Context(), &domain.RestoreJob{
		OrganizationID:   orgID,
		ProjectID:        "proj-default",
		SourceDatabaseID: "res-db-prod-1",
		BackupID:         backupID,
		TargetDBName:     req.TargetDBName,
		TargetRegionID:   req.TargetRegionID,
		RestoreType:      "SNAPSHOT",
	})
	if err != nil {
		if strings.Contains(err.Error(), "authorization violation") {
			http.Error(w, `{"error":"Forbidden: cross-tenant access denied"}`, http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusAccepted, job)
}

func (h *BackupHandler) GetRecoveryPoints(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pitrStatus":         "CONTROL_PLANE_ONLY",
		"mode":               "SNAPSHOT_RECOVERY",
		"message":            "WAL archival & continuous point-in-time recovery points require physical WAL archiving driver attachment",
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

func getOrgIDFromContext(r *http.Request) string {
	orgID := r.Header.Get("X-Organization-ID")
	if orgID == "" {
		orgID = "org-default"
	}
	return orgID
}

func snapshotToTargetDB(id string) string {
	return "target-db-" + id
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
