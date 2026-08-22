package domain

import "context"

type ComputeRepository interface {
	Create(ctx context.Context, inst *ComputeInstance) error
	GetByID(ctx context.Context, id string) (*ComputeInstance, error)
	GetTenantScopedByID(ctx context.Context, orgID, projID, id string) (*ComputeInstance, error)
	ListByProjectID(ctx context.Context, projectID string) ([]*ComputeInstance, error)
	Update(ctx context.Context, inst *ComputeInstance) error
	Delete(ctx context.Context, id string) error
}

type VolumeRepository interface {
	CreateVolume(ctx context.Context, vol *Volume) error
	GetVolumeByID(ctx context.Context, id string) (*Volume, error)
	GetTenantScopedVolumeByID(ctx context.Context, orgID, projID, id string) (*Volume, error)
	ListVolumesByProjectID(ctx context.Context, projectID string) ([]*Volume, error)
	UpdateVolume(ctx context.Context, vol *Volume) error
	DeleteVolume(ctx context.Context, id string) error
}
