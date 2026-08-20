package domain

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type BackupType string

const (
	BackupSnapshot   BackupType = "SNAPSHOT"
	BackupAutomated  BackupType = "AUTOMATED"
	BackupManual     BackupType = "MANUAL"
	BackupWALArchive BackupType = "WAL_ARCHIVE"
)

type BackupStatus string

const (
	StatusRequested  BackupStatus = "REQUESTED"
	StatusQueued     BackupStatus = "QUEUED"
	StatusRunning    BackupStatus = "RUNNING"
	StatusUploading  BackupStatus = "UPLOADING"
	StatusCompleted  BackupStatus = "COMPLETED"
	StatusFailed     BackupStatus = "FAILED"
	StatusVerifying  BackupStatus = "VERIFYING"
	StatusVerified   BackupStatus = "VERIFIED"
	StatusExpired    BackupStatus = "EXPIRED"
	StatusDeleting   BackupStatus = "DELETING"
	StatusDeleted    BackupStatus = "DELETED"
)

type IntegrityStatus string

const (
	IntegrityValid     IntegrityStatus = "VALID"
	IntegrityInvalid   IntegrityStatus = "INVALID"
	IntegrityUnverified IntegrityStatus = "UNVERIFIED"
)

type BackupRecord struct {
	ID                  string          `json:"id"`
	ResourceID          string          `json:"resourceId"`
	OrganizationID      string          `json:"organizationId"`
	ProjectID           string          `json:"projectId"`
	DatabaseID          string          `json:"databaseId"`
	DatabaseName        string          `json:"databaseName"`
	Name                string          `json:"name"`
	Type                BackupType      `json:"type"`
	Status              BackupStatus    `json:"status"`
	Integrity           IntegrityStatus `json:"integrity"`
	SizeBytes           int64           `json:"sizeBytes"`
	RetentionDays       int             `json:"retentionDays"`
	StorageBucket       string          `json:"storageBucket"`
	StoragePath         string          `json:"storagePath"`
	Checksum            string          `json:"checksum"`
	StartedAt           time.Time       `json:"startedAt"`
	CompletedAt         *time.Time      `json:"completedAt,omitempty"`
	ExpiresAt           time.Time       `json:"expiresAt"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

type RestoreJob struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organizationId"`
	ProjectID        string    `json:"projectId"`
	SourceDatabaseID string    `json:"sourceDatabaseId"`
	BackupID         string    `json:"backupId"`
	TargetDBName     string    `json:"targetDbName"`
	TargetRegionID   string    `json:"targetRegionId"`
	RestoreType      string    `json:"restoreType"` // SNAPSHOT, POINT_IN_TIME, CLONE
	Status           string    `json:"status"`      // REQUESTED, VALIDATING, QUEUED, RESTORING, COMPLETED, FAILED
	StartedAt        time.Time `json:"startedAt"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
}

type BackupConfig struct {
	DatabaseID          string `json:"databaseId"`
	Enabled             bool   `json:"enabled"`
	RetentionDays       int    `json:"retentionDays"`
	BackupWindow        string `json:"backupWindow"`
	PitrEnabled         bool   `json:"pitrEnabled"`
	ProviderStatus      string `json:"providerStatus"` // CONFIGURED, PROVIDER_NOT_CONNECTED
}

func GenerateBackupARNV(regionID, projectID, dbName, backupName string) string {
	return arnv.GenerateARNV("BACKUP", regionID, projectID, fmt.Sprintf("database/%s/backup/%s", dbName, backupName))
}

func GenerateBackupStoragePath(orgID, projectID, databaseID, backupID string) string {
	if orgID == "" {
		orgID = "org-default"
	}
	if projectID == "" {
		projectID = "proj-default"
	}
	return fmt.Sprintf("backups/organizations/%s/projects/%s/databases/%s/backups/%s/backup.dump",
		orgID, projectID, databaseID, backupID)
}
