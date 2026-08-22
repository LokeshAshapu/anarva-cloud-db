package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/arnv"
)

type InstanceStatus string

const (
	StatusProvisioning InstanceStatus = "PROVISIONING"
	StatusRunning      InstanceStatus = "RUNNING"
	StatusStopping     InstanceStatus = "STOPPING"
	StatusStopped      InstanceStatus = "STOPPED"
	StatusRestarting   InstanceStatus = "RESTARTING"
	StatusRebuilding   InstanceStatus = "REBUILDING"
	StatusScaling      InstanceStatus = "SCALING"
	StatusDeleting     InstanceStatus = "DELETING"
	StatusDeleted      InstanceStatus = "DELETED"
	StatusFailed       InstanceStatus = "FAILED"
	StatusUnknown      InstanceStatus = "UNKNOWN"
)

type InstanceHealth string

const (
	HealthHealthy     InstanceHealth = "HEALTHY"
	HealthDegraded    InstanceHealth = "DEGRADED"
	HealthUnavailable InstanceHealth = "UNAVAILABLE"
	HealthUnknown     InstanceHealth = "UNKNOWN"
)

type ProviderType string

const (
	ProviderLocalDocker  ProviderType = "LOCAL_DOCKER"
	ProviderControlPlane ProviderType = "CONTROL_PLANE"
	ProviderKubernetes   ProviderType = "KUBERNETES"
	ProviderAWS          ProviderType = "AWS"
	ProviderGCP          ProviderType = "GCP"
	ProviderAzure        ProviderType = "AZURE"
)

type ComputePlan struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	ACU                float64   `json:"acu"`
	VCPU               float64   `json:"vcpu"`
	MemoryMB           int       `json:"memoryMb"`
	StorageLimitGB     int       `json:"storageLimitGb"`
	NetworkLimitMbps   int       `json:"networkLimitMbps"`
	Description        string    `json:"description"`
	Status             string    `json:"status"` // ACTIVE, DEPRECATED
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type ComputeImage struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Version      string    `json:"version"`
	Type         string    `json:"type"` // DOCKER, OS_IMAGE
	Provider     string    `json:"provider"`
	Architecture string    `json:"architecture"` // x86_64, arm64
	Status       string    `json:"status"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"createdAt"`
}

type SecurityGroupRule struct {
	ID          string `json:"id"`
	Direction   string `json:"direction"` // INBOUND, OUTBOUND
	Protocol    string `json:"protocol"`  // TCP, UDP, ICMP
	FromPort    int    `json:"fromPort"`
	ToPort      int    `json:"toPort"`
	CidrBlock   string `json:"cidrBlock"`
	Description string `json:"description"`
}

type SecurityGroup struct {
	ID             string              `json:"id"`
	OrganizationID string              `json:"organizationId"`
	ProjectID      string              `json:"projectId"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Rules          []SecurityGroupRule `json:"rules"`
	CreatedAt      time.Time           `json:"createdAt"`
}

type InstanceSecurityPolicy struct {
	SSHKeyIDs       []string `json:"sshKeyIds,omitempty"`
	ServiceAccountID string   `json:"serviceAccountId,omitempty"`
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`
	IAMRole          string   `json:"iamRole,omitempty"`
	SecretRefs       []string `json:"secretRefs,omitempty"`
}

type Volume struct {
	ID               string     `json:"id" gorm:"primaryKey;column:id;type:varchar(255)"`
	OrganizationID   string     `json:"organizationId" gorm:"column:organization_id;type:varchar(255);index"`
	ProjectID        string     `json:"projectId" gorm:"column:project_id;type:varchar(255);index"`
	InstanceID       string     `json:"instanceId,omitempty" gorm:"column:instance_id;type:varchar(255);index"`
	Name             string     `json:"name" gorm:"column:name;type:varchar(255)"`
	SizeGB           int        `json:"sizeGb" gorm:"column:size_gb"`
	RegionID         string     `json:"regionId" gorm:"column:region_id;type:varchar(100)"`
	ZoneID           string     `json:"zoneId" gorm:"column:zone_id;type:varchar(100)"`
	Type             string     `json:"type" gorm:"column:type;type:varchar(50)"`
	ProviderVolumeID string     `json:"providerVolumeId,omitempty" gorm:"column:provider_volume_id;type:varchar(255)"`
	Status           string     `json:"status" gorm:"column:status;type:varchar(50)"`
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt        *time.Time `json:"deletedAt,omitempty" gorm:"column:deleted_at;index"`
}

func (Volume) TableName() string {
	return "compute_volumes"
}

type ComputeInstance struct {
	ID                 string                 `json:"id" gorm:"primaryKey;column:id;type:varchar(255)"`
	ResourceID         string                 `json:"resourceId" gorm:"column:resource_id;type:varchar(255);index"`
	OrganizationID     string                 `json:"organizationId" gorm:"column:organization_id;type:varchar(255);index"`
	ProjectID          string                 `json:"projectId" gorm:"column:project_id;type:varchar(255);index"`
	Name               string                 `json:"name" gorm:"column:name;type:varchar(255)"`
	Slug               string                 `json:"slug" gorm:"column:slug;type:varchar(255)"`
	RegionID           string                 `json:"regionId" gorm:"column:region_id;type:varchar(100);index"`
	ZoneID             string                 `json:"zoneId" gorm:"column:zone_id;type:varchar(100)"`
	Status             InstanceStatus         `json:"status" gorm:"column:status;type:varchar(50);index"`
	Health             InstanceHealth         `json:"health" gorm:"column:health;type:varchar(50)"`
	PlanID             string                 `json:"planId" gorm:"column:plan_id;type:varchar(100)"`
	ACU                float64                `json:"acu" gorm:"column:acu;type:numeric(10,2)"`
	VCPU               float64                `json:"vcpu" gorm:"column:vcpu;type:numeric(10,2)"`
	MemoryMB           int                    `json:"memoryMb" gorm:"column:memory_mb"`
	StorageGB          int                    `json:"storageGb" gorm:"column:storage_gb"`
	ImageID            string                 `json:"imageId" gorm:"column:image_id;type:varchar(100)"`
	DockerImage        string                 `json:"dockerImage,omitempty" gorm:"column:docker_image;type:varchar(255)"`
	NetworkID          string                 `json:"networkId" gorm:"column:network_id;type:varchar(255)"`
	SubnetID           string                 `json:"subnetId" gorm:"column:subnet_id;type:varchar(255)"`
	PrivateIP          string                 `json:"privateIp,omitempty" gorm:"column:private_ip;type:varchar(100)"`
	PublicIP           string                 `json:"publicIp,omitempty" gorm:"column:public_ip;type:varchar(100)"`
	Provider           ProviderType           `json:"provider" gorm:"column:provider;type:varchar(100);index"`
	ProviderInstanceID string                 `json:"providerInstanceId,omitempty" gorm:"column:provider_instance_id;type:varchar(255);index"`
	Security           InstanceSecurityPolicy `json:"security" gorm:"-"`
	SecurityJSON       string                 `json:"-" gorm:"column:security_json;type:text"`
	EnvVars            map[string]string      `json:"envVars,omitempty" gorm:"-"`
	EnvVarsJSON        string                 `json:"-" gorm:"column:env_vars_json;type:text"`
	CreatedAt          time.Time              `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt          time.Time              `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt          *time.Time             `json:"deletedAt,omitempty" gorm:"column:deleted_at;index"`
}

func (ComputeInstance) TableName() string {
	return "compute_instances"
}

func (c *ComputeInstance) BeforeSave(tx *gorm.DB) error {
	if secBytes, err := json.Marshal(c.Security); err == nil {
		c.SecurityJSON = string(secBytes)
	}
	if len(c.EnvVars) > 0 {
		if envBytes, err := json.Marshal(c.EnvVars); err == nil {
			c.EnvVarsJSON = string(envBytes)
		}
	} else {
		c.EnvVarsJSON = ""
	}
	return nil
}

func (c *ComputeInstance) AfterFind(tx *gorm.DB) error {
	if c.SecurityJSON != "" {
		var sec InstanceSecurityPolicy
		if err := json.Unmarshal([]byte(c.SecurityJSON), &sec); err == nil {
			c.Security = sec
		}
	}
	if c.EnvVarsJSON != "" {
		var envs map[string]string
		if err := json.Unmarshal([]byte(c.EnvVarsJSON), &envs); err == nil {
			c.EnvVars = envs
		}
	}
	return nil
}

type ComputeCapacity struct {
	Provider          string    `json:"provider"`
	Region            string    `json:"region"`
	Zone              string    `json:"zone"`
	AvailableACU      float64   `json:"availableAcu"`
	AvailableVCPU     float64   `json:"availableVcpu"`
	AvailableMemoryMB int       `json:"availableMemoryMb"`
	UpdatedAt         time.Time `json:"updatedAt"`
	Status            string    `json:"status"` // AVAILABLE, UNKNOWN, EXHAUSTED
}

type CommandExecutionRequest struct {
	Command string `json:"command"`
	Timeout int    `json:"timeoutSeconds,omitempty"`
}

type CommandExecutionResult struct {
	ExitCode int       `json:"exitCode"`
	Stdout   string    `json:"stdout"`
	Stderr   string    `json:"stderr"`
	Executed time.Time `json:"executedAt"`
}

var ValidACUs = []float64{0.5, 1.0, 2.0, 4.0, 8.0, 16.0, 32.0, 64.0, 128.0}

func IsValidACU(acu float64) bool {
	for _, v := range ValidACUs {
		if v == acu {
			return true
		}
	}
	return false
}

func GenerateComputeARNV(regionID, projectID, instanceName string) string {
	return arnv.GenerateARNV("COMPUTE", regionID, projectID, fmt.Sprintf("compute/%s", instanceName))
}
