package providers

import (
	"context"
	"time"
)

// ComputeInstanceOpts defines parameters for provisioning Anarva Compute Units (ACUs).
type ComputeInstanceOpts struct {
	ID        string
	ProjectID string
	Name      string
	Region    string
	Zone      string
	MinACU    float64
	MaxACU    float64
	CPUCores  float64
	MemoryMB  int
	Image     string
}

// ComputeInstanceDetails represents compute instance state.
type ComputeInstanceDetails struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	CurrentACU  float64   `json:"current_acu"`
	PublicIP    string    `json:"public_ip"`
	PrivateIP   string    `json:"private_ip"`
	CreatedAt   time.Time `json:"created_at"`
}

// ComputeProvider defines provider-independent interface for compute workloads and ACU scaling.
type ComputeProvider interface {
	LaunchInstance(ctx context.Context, opts ComputeInstanceOpts) (*ComputeInstanceDetails, error)
	StartInstance(ctx context.Context, instanceID string) error
	StopInstance(ctx context.Context, instanceID string) error
	TerminateInstance(ctx context.Context, instanceID string) error
	ScaleComputeACU(ctx context.Context, instanceID string, targetACU float64) error
	GetInstanceMetrics(ctx context.Context, instanceID string) (map[string]float64, error)
}
