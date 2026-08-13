package gcp

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
)

type GCPProvider struct {
	projectID string
	credRef   string
}

func NewGCPProvider(credRef, projectID string) *GCPProvider {
	return &GCPProvider{
		credRef:   credRef,
		projectID: projectID,
	}
}

func (p *GCPProvider) GetProviderType() registry.ProviderType {
	return registry.ProviderGCP
}

func (p *GCPProvider) VerifyConnection(ctx context.Context) error {
	if p.credRef == "" {
		return fmt.Errorf("PROVIDER_NOT_CONFIGURED: GCP Service Account JSON key reference missing")
	}
	return nil
}

func (p *GCPProvider) GetCapabilities(ctx context.Context) registry.CapabilityMatrix {
	return registry.CapabilityMatrix{
		Compute:       true,
		Containers:    true,
		Kubernetes:    true,
		PostgreSQL:    true,
		MySQL:         true,
		ObjectStorage: true,
		Networking:    true,
		LoadBalancer:  true,
		DNS:           true,
		TLS:           true,
		Monitoring:    true,
		Backup:        true,
		Replication:   true,
		Autoscaling:   true,
	}
}

func (p *GCPProvider) ListRegions(ctx context.Context) ([]string, error) {
	if p.credRef == "" {
		return nil, fmt.Errorf("PROVIDER_NOT_CONFIGURED: Cannot discover GCP regions without active Service Account credentials")
	}
	return []string{"us-central1", "us-east1", "europe-west1", "asia-south1", "asia-east1"}, nil
}
