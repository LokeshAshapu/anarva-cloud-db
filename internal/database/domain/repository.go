package domain

import "context"

type InstanceRepository interface {
	Create(ctx context.Context, instance *DatabaseInstance) error
	GetByID(ctx context.Context, id string) (*DatabaseInstance, error)
	ListByProjectID(ctx context.Context, projectID string) ([]*DatabaseInstance, error)
	Update(ctx context.Context, instance *DatabaseInstance) error
	Delete(ctx context.Context, id string) error
	CountByProjectID(ctx context.Context, projectID string) (int64, error)
}
