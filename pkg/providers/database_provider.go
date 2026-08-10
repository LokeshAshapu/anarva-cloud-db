package providers

import (
	"context"

	dbDomain "github.com/anarva-cloud/anarva-cloud-db/internal/database/domain"
)

// DatabaseProvisionOpts encapsulates parameters for provisioning a managed database instance.
type DatabaseProvisionOpts struct {
	ID            string
	ProjectID     string
	Name          string
	Engine        dbDomain.EngineType
	Version       string
	Region        string
	MultiAZ       bool
	MinACU        float64
	MaxACU        float64
	StorageSizeGB int
}

// DatabaseProvider defines the provider-independent abstraction interface for managed database workloads.
type DatabaseProvider interface {
	ProvisionInstance(ctx context.Context, opts DatabaseProvisionOpts) (*dbDomain.DatabaseInstance, error)
	StartInstance(ctx context.Context, instanceID string) error
	StopInstance(ctx context.Context, instanceID string) error
	TerminateInstance(ctx context.Context, instanceID string) error
	ScaleACU(ctx context.Context, instanceID string, minACU, maxACU float64) error
	CreateReadReplica(ctx context.Context, primaryInstanceID, region string) (*dbDomain.DatabaseInstance, error)
	GetInstanceHealth(ctx context.Context, instanceID string) (string, error)
}
