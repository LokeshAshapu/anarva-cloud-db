package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EngineType string

const (
	EnginePostgreSQL EngineType = "postgres"
	EngineMySQL      EngineType = "mysql"
)

type InstanceStatus string

const (
	StatusProvisioning InstanceStatus = "PROVISIONING"
	StatusRunning      InstanceStatus = "RUNNING"
	StatusStopped      InstanceStatus = "STOPPED"
	StatusFailed       InstanceStatus = "FAILED"
	StatusTerminated   InstanceStatus = "TERMINATED"
)

type DatabaseInstance struct {
	ID                string         `gorm:"primaryKey;type:uuid"`
	ProjectID         string         `gorm:"index;not null;type:uuid"`
	Name              string         `gorm:"not null;type:varchar(255)"`
	Engine            EngineType     `gorm:"not null;type:varchar(50);default:'postgres'"`
	Status            InstanceStatus `gorm:"not null;type:varchar(50);default:'PROVISIONING'"`
	Host              string         `gorm:"not null;type:varchar(255)"`
	Port              int            `gorm:"not null"`
	DBName            string         `gorm:"not null;type:varchar(255)"`
	Username          string         `gorm:"not null;type:varchar(255)"`
	PasswordEncrypted string         `gorm:"not null;type:text"`
	StorageSizeGB     int            `gorm:"not null;default:10"`
	CPUCores          float64        `gorm:"not null;default:1.0"`
	MemoryMB          int            `gorm:"not null;default:1024"`
	ContainerID       string         `gorm:"type:varchar(255)"`
	CreatedAt         time.Time      `gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"`
}

func NewDatabaseInstance(projectID, name string, engine EngineType, host string, port int, dbName, username, passwordEncrypted string, storageGB int) *DatabaseInstance {
	if engine == "" {
		engine = EnginePostgreSQL
	}
	if storageGB <= 0 {
		storageGB = 10
	}
	return &DatabaseInstance{
		ID:                uuid.New().String(),
		ProjectID:         projectID,
		Name:              name,
		Engine:            engine,
		Status:            StatusProvisioning,
		Host:              host,
		Port:              port,
		DBName:            dbName,
		Username:          username,
		PasswordEncrypted: passwordEncrypted,
		StorageSizeGB:     storageGB,
		CPUCores:          1.0,
		MemoryMB:          1024,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func (i *DatabaseInstance) FormatConnectionString(rawPassword string) string {
	switch i.Engine {
	case EnginePostgreSQL:
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			i.Username, rawPassword, i.Host, i.Port, i.DBName)
	case EngineMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			i.Username, rawPassword, i.Host, i.Port, i.DBName)
	default:
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			i.Username, rawPassword, i.Host, i.Port, i.DBName)
	}
}
