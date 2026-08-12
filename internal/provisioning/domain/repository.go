package domain

import "context"

type ProvisioningRepository interface {
	CreateRequest(ctx context.Context, req *ProvisioningRequest) error
	GetRequestByID(ctx context.Context, id string) (*ProvisioningRequest, error)
	GetRequestByIdempotencyKey(ctx context.Context, key string) (*ProvisioningRequest, error)
	ListRequests(ctx context.Context, projectID string) ([]*ProvisioningRequest, error)
	UpdateRequestStatus(ctx context.Context, id string, status ProvisioningStatus, errCode, errMsg string) error
}

type ResourceLockRepository interface {
	AcquireLock(ctx context.Context, lock *ResourceLock) error
	ReleaseLock(ctx context.Context, resourceID string) error
	GetLock(ctx context.Context, resourceID string) (*ResourceLock, error)
}

type DriftRepository interface {
	SaveDrift(ctx context.Context, drift *ResourceDrift) error
	GetDriftByResourceID(ctx context.Context, resourceID string) (*ResourceDrift, error)
	ListDrifts(ctx context.Context) ([]*ResourceDrift, error)
}
