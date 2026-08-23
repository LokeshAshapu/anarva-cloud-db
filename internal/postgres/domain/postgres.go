package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Availability modes
type AvailabilityMode string

const (
	AvailabilitySingle         AvailabilityMode = "SINGLE"
	AvailabilityPrimaryStandby AvailabilityMode = "PRIMARY_STANDBY"
	AvailabilityMultiZone      AvailabilityMode = "MULTI_ZONE"
	AvailabilityMultiRegion    AvailabilityMode = "MULTI_REGION"
)

// Instance Statuses
type InstanceStatus string

const (
	StatusCreating   InstanceStatus = "CREATING"
	StatusAvailable  InstanceStatus = "AVAILABLE"
	StatusUpdating   InstanceStatus = "UPDATING"
	StatusRestarting InstanceStatus = "RESTARTING"
	StatusStopping   InstanceStatus = "STOPPING"
	StatusStopped    InstanceStatus = "STOPPED"
	StatusDeleting   InstanceStatus = "DELETING"
	StatusDeleted    InstanceStatus = "DELETED"
	StatusFailed     InstanceStatus = "FAILED"
	StatusDegraded   InstanceStatus = "DEGRADED"
	StatusUnknown    InstanceStatus = "UNKNOWN"
)

// User Roles
type UserRole string

const (
	RoleOwner     UserRole = "OWNER"
	RoleAdmin     UserRole = "ADMIN"
	RoleReadWrite UserRole = "READ_WRITE"
	RoleReadOnly  UserRole = "READ_ONLY"
	RoleCustom    UserRole = "CUSTOM"
)

// PostgresVersion
type PostgresVersion struct {
	Version     string    `json:"version"`
	Status      string    `json:"status"` // ACTIVE, DEPRECATED, EOL
	Supported   bool      `json:"supported"`
	EndOfLifeAt string    `json:"endOfLifeAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// PostgresInstance
type PostgresInstance struct {
	ID                 string           `gorm:"primaryKey;column:id;type:varchar(255)" json:"id"`
	OrganizationID     string           `gorm:"column:organization_id;type:varchar(255);index" json:"organizationId"`
	ProjectID          string           `gorm:"column:project_id;type:varchar(255);index" json:"projectId"`
	Name               string           `gorm:"column:name;type:varchar(255)" json:"name"`
	Provider           string           `gorm:"column:provider;type:varchar(50);default:'LOCAL_POSTGRES'" json:"provider"`
	Version            string           `gorm:"column:version;type:varchar(50);default:'17'" json:"version"`
	Status             InstanceStatus   `gorm:"column:status;type:varchar(50);default:'CREATING'" json:"status"`
	RegionID           string           `gorm:"column:region_id;type:varchar(100)" json:"regionId"`
	ZoneId             string           `gorm:"column:zone_id;type:varchar(100)" json:"zoneId"`
	CPU                float64          `gorm:"column:cpu;default:1.0" json:"cpu"`
	MemoryMB           int              `gorm:"column:memory_mb;default:1024" json:"memoryMb"`
	StorageGB          int              `gorm:"column:storage_gb;default:25" json:"storageGb"`
	StorageType        string           `gorm:"column:storage_type;type:varchar(50);default:'SSD'" json:"storageType"`
	NetworkID          string           `gorm:"column:network_id;type:varchar(255)" json:"networkId"`
	SubnetID           string           `gorm:"column:subnet_id;type:varchar(255)" json:"subnetId"`
	AvailabilityMode   AvailabilityMode `gorm:"column:availability_mode;type:varchar(50);default:'SINGLE'" json:"availabilityMode"`
	BackupMode         string           `gorm:"column:backup_mode;type:varchar(50);default:'DAILY_SNAPSHOT'" json:"backupMode"`
	MaintenanceWindow  string           `gorm:"column:maintenance_window;type:varchar(100);default:'Sun:03:00'" json:"maintenanceWindow"`
	ProviderResourceId string           `gorm:"column:provider_resource_id;type:varchar(255)" json:"providerResourceId"`
	Host               string           `gorm:"column:host;type:varchar(255)" json:"host"`
	Port               int              `gorm:"column:port;default:5432" json:"port"`
	PublicAccess       bool             `gorm:"column:public_access;default:false" json:"publicAccess"`
	RealityLabel       string           `gorm:"column:reality_label;type:varchar(100);default:'LOCAL_POSTGRES'" json:"realityLabel"`
	CreatedAt          time.Time        `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt          time.Time        `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt          *time.Time       `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

func (PostgresInstance) TableName() string {
	return "postgres_instances"
}

// PostgresDatabase
type PostgresDatabase struct {
	ID             string    `gorm:"primaryKey;column:id;type:varchar(255)" json:"id"`
	InstanceID     string    `gorm:"column:instance_id;type:varchar(255);index" json:"instanceId"`
	Name           string    `gorm:"column:name;type:varchar(255)" json:"name"`
	OwnerReference string    `gorm:"column:owner_reference;type:varchar(255)" json:"ownerReference"`
	Encoding       string    `gorm:"column:encoding;type:varchar(50);default:'UTF8'" json:"encoding"`
	Collation      string    `gorm:"column:collation;type:varchar(50);default:'en_US.UTF-8'" json:"collation"`
	Status         string    `gorm:"column:status;type:varchar(50);default:'AVAILABLE'" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (PostgresDatabase) TableName() string {
	return "postgres_databases"
}

// CredentialReference - Password / Secrets wrapper (never plaintext in DB columns)
type CredentialReference struct {
	ID              string    `gorm:"primaryKey;column:id;type:varchar(255)" json:"id"`
	Provider        string    `gorm:"column:provider;type:varchar(50)" json:"provider"`
	SecretReference string    `gorm:"column:secret_reference;type:varchar(255)" json:"secretReference"`
	Type            string    `gorm:"column:type;type:varchar(50);default:'POSTGRES_USER_SECRET'" json:"type"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// PostgresUser
type PostgresUser struct {
	ID                  string    `gorm:"primaryKey;column:id;type:varchar(255)" json:"id"`
	InstanceID          string    `gorm:"column:instance_id;type:varchar(255);index" json:"instanceId"`
	Username            string    `gorm:"column:username;type:varchar(255)" json:"username"`
	Role                UserRole  `gorm:"column:role;type:varchar(50);default:'READ_WRITE'" json:"role"`
	Status              string    `gorm:"column:status;type:varchar(50);default:'ACTIVE'" json:"status"`
	CredentialReference string    `gorm:"column:credential_reference;type:varchar(255)" json:"credentialReference"`
	CreatedAt           time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (PostgresUser) TableName() string {
	return "postgres_users"
}

// ConnectionInfo
type ConnectionInfo struct {
	HostReference        string `json:"hostReference"`
	Port                 int    `json:"port"`
	Database             string `json:"database"`
	UsernameReference    string `json:"usernameReference"`
	PasswordSecretRef    string `json:"passwordSecretRef"`
	SSLMode              string `json:"sslMode"`
	ConnectionString     string `json:"connectionString"`
	CertificateReference string `json:"certificateReference,omitempty"`
}

// StorageAutoscalingPolicy
type StorageAutoscalingPolicy struct {
	MinStorageGB int    `json:"minStorageGb"`
	MaxStorageGB int    `json:"maxStorageGb"`
	ThresholdPct int    `json:"thresholdPct"`
	IncrementGB  int    `json:"incrementGb"`
	Status       string `json:"status"` // ACTIVE, DISABLED, NOT_SUPPORTED
}

// DatabaseHealth
type DatabaseHealth struct {
	InstanceID           string    `json:"instanceId"`
	ConnectionAvailable  bool      `json:"connectionAvailable"`
	ReplicationStatus    string    `json:"replicationStatus"` // SINGLE, STREAMING, LAGGING, PITR_NOT_CONFIGURED
	CPUPct               float64   `json:"cpuPct"`
	MemoryPct            float64   `json:"memoryPct"`
	StorageUsedGB        float64   `json:"storageUsedGb"`
	StorageAllocatedGB   float64   `json:"storageAllocatedGb"`
	ActiveConnections    int       `json:"activeConnections"`
	MaxConnections       int       `json:"maxConnections"`
	QueryLatencyMs       float64   `json:"queryLatencyMs"`
	CacheHitRatio        float64   `json:"cacheHitRatio"`
	SourceQuality        string    `json:"sourceQuality"` // ACTUAL, ESTIMATED, UNKNOWN
	Timestamp            time.Time `json:"timestamp"`
}

// PostgresLogEntry
type PostgresLogEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // INFO, WARN, ERROR, FATAL
	Component string    `json:"component"`
	Message   string    `json:"message"`
	Query     string    `json:"query,omitempty"`
}

// PostgresReplica
type PostgresReplica struct {
	ID                 string    `json:"id"`
	SourceInstanceID   string    `json:"sourceInstanceId"`
	ProviderResourceID string    `json:"providerResourceId"`
	RegionID           string    `json:"regionId"`
	ZoneID             string    `json:"zoneId"`
	Status             string    `json:"status"` // STREAMING, LAGGING, STOPPED, FAILED
	ReplicationLagMs   int64     `json:"replicationLagMs"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// MaintenanceWindow
type MaintenanceWindow struct {
	Day              string    `json:"day"` // Sun, Mon, Tue...
	StartTime        string    `json:"startTime"` // 03:00
	DurationMinutes  int       `json:"durationMinutes"`
	Timezone         string    `json:"timezone"`
	Status           string    `json:"status"` // SCHEDULED, IDLE, IN_PROGRESS
	NextScheduledAt  time.Time `json:"nextScheduledAt"`
}

// Custom Errors
var (
	ErrInstanceNotFound    = errors.New("postgres instance not found")
	ErrInvalidVersion      = errors.New("unsupported postgresql version")
	ErrQuotaExceeded       = errors.New("database quota limit exceeded")
	ErrPublicAccessDenied  = errors.New("public postgresql access requires explicit security confirmation")
	ErrInvalidCredentials  = errors.New("invalid postgresql credentials or secret reference")
)

func NewPostgresInstance(orgID, projectID, name, version, regionID, networkID string, cpu float64, memoryMB, storageGB int) *PostgresInstance {
	if version == "" {
		version = "17"
	}
	if cpu <= 0 {
		cpu = 1.0
	}
	if memoryMB <= 0 {
		memoryMB = 1024
	}
	if storageGB <= 0 {
		storageGB = 25
	}

	return &PostgresInstance{
		ID:               uuid.New().String(),
		OrganizationID:   orgID,
		ProjectID:        projectID,
		Name:             name,
		Provider:         "LOCAL_POSTGRES",
		Version:          version,
		Status:           StatusCreating,
		RegionID:         regionID,
		ZoneId:           regionID + "-a",
		CPU:              cpu,
		MemoryMB:         memoryMB,
		StorageGB:        storageGB,
		StorageType:      "SSD",
		NetworkID:        networkID,
		SubnetID:         networkID + "-sub-1",
		AvailabilityMode: AvailabilitySingle,
		BackupMode:       "DAILY_SNAPSHOT",
		MaintenanceWindow: "Sun:03:00",
		Host:             "localhost",
		Port:             5432,
		PublicAccess:     false,
		RealityLabel:     "LOCAL_POSTGRES",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}
