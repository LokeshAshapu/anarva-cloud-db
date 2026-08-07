package domain

import "context"

type ProvisionParams struct {
	InstanceID string
	Engine     EngineType
	DBName     string
	Username   string
	Password   string
	Port       int
}

type ProvisionerDriver interface {
	ProvisionInstance(ctx context.Context, params ProvisionParams) (containerID string, err error)
	StartInstance(ctx context.Context, containerID string) error
	StopInstance(ctx context.Context, containerID string) error
	TerminateInstance(ctx context.Context, containerID string) error
	CheckHealth(ctx context.Context, containerID string) (bool, error)
}
