package domain

import (
	"fmt"
	"time"

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
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organizationId"`
	ProjectID        string    `json:"projectId"`
	InstanceID       string    `json:"instanceId,omitempty"`
	Name             string    `json:"name"`
	SizeGB           int       `json:"sizeGb"`
	RegionID         string    `json:"regionId"`
	ZoneID           string    `json:"zoneId"`
	Type             string    `json:"type"` // NVME_SSD, STANDARD_HDD
	ProviderVolumeID string    `json:"providerVolumeId,omitempty"`
	Status           string    `json:"status"` // ATTACHED, DETACHED, CREATING, DELETING
	CreatedAt        time.Time `json:"createdAt"`
}

type ComputeInstance struct {
	ID                 string                 `json:"id"`
	ResourceID         string                 `json:"resourceId"`
	OrganizationID     string                 `json:"organizationId"`
	ProjectID          string                 `json:"projectId"`
	Name               string                 `json:"name"`
	Slug               string                 `json:"slug"`
	RegionID           string                 `json:"regionId"`
	ZoneID             string                 `json:"zoneId"`
	Status             InstanceStatus         `json:"status"`
	Health             InstanceHealth         `json:"health"`
	PlanID             string                 `json:"planId"`
	ACU                float64                `json:"acu"`
	VCPU               float64                `json:"vcpu"`
	MemoryMB           int                    `json:"memoryMb"`
	StorageGB          int                    `json:"storageGb"`
	ImageID            string                 `json:"imageId"`
	DockerImage        string                 `json:"dockerImage,omitempty"`
	NetworkID          string                 `json:"networkId"`
	SubnetID           string                 `json:"subnetId"`
	PrivateIP          string                 `json:"privateIp,omitempty"`
	PublicIP           string                 `json:"publicIp,omitempty"`
	Provider           ProviderType           `json:"provider"`
	ProviderInstanceID string                 `json:"providerInstanceId,omitempty"`
	Security           InstanceSecurityPolicy `json:"security"`
	EnvVars            map[string]string      `json:"envVars,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	DeletedAt          *time.Time             `json:"deletedAt,omitempty"`
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
