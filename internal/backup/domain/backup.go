package domain

import (
	"time"

	"github.com/google/uuid"
)

type BackupType string

const (
	BackupTypeSnapshot BackupType = "SNAPSHOT"
	BackupTypeWAL      BackupType = "WAL"
)

type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "PENDING"
	BackupStatusInProgress BackupStatus = "IN_PROGRESS"
	BackupStatusCompleted  BackupStatus = "COMPLETED"
	BackupStatusFailed     BackupStatus = "FAILED"
)

type BackupSnapshot struct {
	ID          string       `gorm:"primaryKey;type:uuid"`
	DatabaseID  string       `gorm:"index;not null;type:uuid"`
	ProjectID   string       `gorm:"index;not null;type:uuid"`
	Name        string       `gorm:"not null;type:varchar(255)"`
	StoragePath string       `gorm:"not null;type:varchar(512)"`
	SizeBytes   int64        `gorm:"not null;default:0"`
	BackupType  BackupType   `gorm:"not null;type:varchar(50);default:'SNAPSHOT'"`
	Status      BackupStatus `gorm:"not null;type:varchar(50);default:'PENDING'"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
}

func NewBackupSnapshot(databaseID, projectID, name, storagePath string, backupType BackupType) *BackupSnapshot {
	return &BackupSnapshot{
		ID:          uuid.New().String(),
		DatabaseID:  databaseID,
		ProjectID:   projectID,
		Name:        name,
		StoragePath: storagePath,
		SizeBytes:   0,
		BackupType:  backupType,
		Status:      BackupStatusPending,
		CreatedAt:   time.Now(),
	}
}
