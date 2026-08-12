package provider

import "time"

type ProviderHealthState string

const (
	StateConnected     ProviderHealthState = "CONNECTED"
	StateDegraded      ProviderHealthState = "DEGRADED"
	StateDisconnected  ProviderHealthState = "DISCONNECTED"
	StateNotConfigured ProviderHealthState = "NOT_CONFIGURED"
	StateUnknown       ProviderHealthState = "UNKNOWN"
)

type ProviderStatus struct {
	Name            string              `json:"name"`
	Category        string              `json:"category"` // DATABASE, STORAGE, COMPUTE, NETWORK, BACKUP, DNS, LOAD_BALANCER
	State           ProviderHealthState `json:"state"`
	ProviderType    string              `json:"providerType"` // LOCAL_DOCKER, LOCAL_STORAGE, AWS_EC2, AWS_S3, GCP, AZURE
	RealityLabel    string              `json:"realityLabel"` // REAL, LOCAL DEVELOPMENT, PROVIDER CONNECTED, CONFIGURED, NOT CONNECTED, PLANNED
	Details         string              `json:"details"`
	LastHealthCheck time.Time           `json:"lastHealthCheck"`
}

type ProviderRegistry struct {
	providers map[string]*ProviderStatus
}

func NewProviderRegistry() *ProviderRegistry {
	now := time.Now()
	return &ProviderRegistry{
		providers: map[string]*ProviderStatus{
			"COMPUTE": {
				Name:            "Anarva Compute Engine (ACE)",
				Category:        "COMPUTE",
				State:           StateConnected,
				ProviderType:    "LOCAL_DOCKER",
				RealityLabel:    "LOCAL DEVELOPMENT PROVIDER",
				Details:         "Docker Container Runtime available with cgroup limits",
				LastHealthCheck: now,
			},
			"DATABASE": {
				Name:            "Anarva Database Engine",
				Category:        "DATABASE",
				State:           StateConnected,
				ProviderType:    "LOCAL_POSTGRES",
				RealityLabel:    "LOCAL DEVELOPMENT PROVIDER",
				Details:         "PostgreSQL 17.2 Local Driver active",
				LastHealthCheck: now,
			},
			"STORAGE": {
				Name:            "Anarva Object Storage (AOS)",
				Category:        "STORAGE",
				State:           StateConnected,
				ProviderType:    "LOCAL_FS",
				RealityLabel:    "LOCAL DEVELOPMENT PROVIDER",
				Details:         "Local filesystem bucket storage active",
				LastHealthCheck: now,
			},
			"NETWORK": {
				Name:            "Anarva Network Engine (VPC)",
				Category:        "NETWORK",
				State:           StateConnected,
				ProviderType:    "LOCAL_DOCKER_BRIDGE",
				RealityLabel:    "LOCAL DEVELOPMENT PROVIDER",
				Details:         "Docker Bridge Network driver active",
				LastHealthCheck: now,
			},
			"BACKUP": {
				Name:            "Anarva Backup Control Plane",
				Category:        "BACKUP",
				State:           StateConnected,
				ProviderType:    "CONTROL_PLANE_STORAGE",
				RealityLabel:    "CONFIGURED",
				Details:         "Control-plane backup registry active",
				LastHealthCheck: now,
			},
			"DNS": {
				Name:            "Anarva Public DNS Provider",
				Category:        "DNS",
				State:           StateNotConfigured,
				ProviderType:    "AWS_ROUTE53",
				RealityLabel:    "PROVIDER NOT CONNECTED",
				Details:         "Public DNS resolution requires cloud provider credentials",
				LastHealthCheck: now,
			},
			"LOAD_BALANCER": {
				Name:            "Anarva Global Load Balancer",
				Category:        "LOAD_BALANCER",
				State:           StateConnected,
				ProviderType:    "LOCAL_PROXY",
				RealityLabel:    "LOCAL DEVELOPMENT PROVIDER",
				Details:         "Application proxy routing active",
				LastHealthCheck: now,
			},
		},
	}
}

func (r *ProviderRegistry) ListStatuses() []*ProviderStatus {
	var list []*ProviderStatus
	for _, p := range r.providers {
		list = append(list, p)
	}
	return list
}
