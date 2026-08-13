package domain

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type EngineType string

const (
	EnginePostgreSQL EngineType = "POSTGRESQL"
	EngineMySQL      EngineType = "MYSQL"
)

type MySQLStatus string

const (
	StatusCreating   MySQLStatus = "CREATING"
	StatusAvailable  MySQLStatus = "AVAILABLE"
	StatusUpdating   MySQLStatus = "UPDATING"
	StatusRestarting MySQLStatus = "RESTARTING"
	StatusStopping   MySQLStatus = "STOPPING"
	StatusStopped    MySQLStatus = "STOPPED"
	StatusDeleting   MySQLStatus = "DELETING"
	StatusDeleted    MySQLStatus = "DELETED"
	StatusFailed     MySQLStatus = "FAILED"
	StatusDegraded   MySQLStatus = "DEGRADED"
	StatusUnknown    MySQLStatus = "UNKNOWN"
)

type MySQLUserRole string

const (
	RoleOwner     MySQLUserRole = "OWNER"
	RoleAdmin     MySQLUserRole = "ADMIN"
	RoleReadWrite MySQLUserRole = "READ_WRITE"
	RoleReadOnly  MySQLUserRole = "READ_ONLY"
	RoleCustom    MySQLUserRole = "CUSTOM"
)

type AvailabilityMode string

const (
	AvailabilitySingle        AvailabilityMode = "SINGLE"
	AvailabilityPrimaryStandby AvailabilityMode = "PRIMARY_STANDBY"
	AvailabilityMultiAZ       AvailabilityMode = "MULTI_ZONE"
	AvailabilityMultiRegion   AvailabilityMode = "MULTI_REGION"
)

type MySQLVersion struct {
	Version   string `json:"version"`
	Supported bool   `json:"supported"`
	Status    string `json:"status"`
	EndOfLife string `json:"endOfLife,omitempty"`
}

type MySQLInstance struct {
	ID                  string           `json:"id"`
	OrganizationID      string           `json:"organizationId"`
	ProjectID           string           `json:"projectId"`
	Name                string           `json:"name"`
	Provider            string           `json:"provider"`
	Version             string           `json:"version"`
	Status              MySQLStatus      `json:"status"`
	RegionID            string           `json:"regionId"`
	ZoneID              string           `json:"zoneId"`
	CPU                 int              `json:"cpu"`
	MemoryMB            int              `json:"memoryMb"`
	StorageGB           int              `json:"storageGb"`
	StorageType         string           `json:"storageType"`
	NetworkID           string           `json:"networkId"`
	SubnetID            string           `json:"subnetId"`
	AvailabilityMode    AvailabilityMode `json:"availabilityMode"`
	BackupMode          string           `json:"backupMode"`
	MaintenanceWindow   string           `json:"maintenanceWindow"`
	ProviderResourceID  string           `json:"providerResourceId"`
	Port                int              `json:"port"`
	RealityLabel        string           `json:"realityLabel"`
	CreatedAt           time.Time        `json:"createdAt"`
	UpdatedAt           time.Time        `json:"updatedAt"`
}

type MySQLDatabase struct {
	ID        string    `json:"id"`
	InstanceID string   `json:"instanceId"`
	Name      string    `json:"name"`
	Charset   string    `json:"charset"`
	Collation string    `json:"collation"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type MySQLUser struct {
	ID                  string        `json:"id"`
	InstanceID          string        `json:"instanceId"`
	Username            string        `json:"username"`
	HostPattern         string        `json:"hostPattern"`
	Role                MySQLUserRole `json:"role"`
	Status              string        `json:"status"`
	CredentialReference string        `json:"credentialReference"`
	CreatedAt           time.Time     `json:"createdAt"`
	UpdatedAt           time.Time     `json:"updatedAt"`
}

type MySQLPrivilege struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	DatabaseName string    `json:"databaseName"`
	TableName    string    `json:"tableName"`
	Privilege    string    `json:"privilege"` // SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, DROP, EXECUTE
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

type StorageAutoscalingPolicy struct {
	MinStorageGB int    `json:"minStorageGb"`
	MaxStorageGB int    `json:"maxStorageGb"`
	Threshold    int    `json:"threshold"`
	IncrementGB  int    `json:"incrementGb"`
	Status       string `json:"status"`
}

type MySQLHealth struct {
	InstanceID      string    `json:"instanceId"`
	Status          string    `json:"status"`
	ActiveConns     int       `json:"activeConns"`
	ThreadsRunning  int       `json:"threadsRunning"`
	BufferPoolUsage float64   `json:"bufferPoolUsage"`
	SlowQueries     int64     `json:"slowQueries"`
	UptimeSec       int64     `json:"uptimeSec"`
	Timestamp       time.Time `json:"timestamp"`
}

type MySQLBackup struct {
	ID           string    `json:"id"`
	InstanceID   string    `json:"instanceId"`
	Name         string    `json:"name"`
	SizeBytes    int64     `json:"sizeBytes"`
	Status       string    `json:"status"`
	StoragePath  string    `json:"storagePath"`
	RealityLabel string    `json:"realityLabel"`
	CreatedAt    time.Time `json:"createdAt"`
}

type MySQLReplica struct {
	ID                 string    `json:"id"`
	SourceInstanceID   string    `json:"sourceInstanceId"`
	ProviderResourceID string    `json:"providerResourceId"`
	RegionID           string    `json:"regionId"`
	ZoneID             string    `json:"zoneId"`
	Status             string    `json:"status"`
	ReplicationLagSec  float64   `json:"replicationLagSec"`
	CreatedAt          time.Time `json:"createdAt"`
}

func GenerateMySQLARNV(regionID, projectID, instanceName string) string {
	return arnv.GenerateARNV("MYSQL", regionID, projectID, fmt.Sprintf("database/%s", instanceName))
}
