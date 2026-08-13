package aws

import (
	"context"
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
)

type AWSProvider struct {
	credRef string
	region  string
}

func NewAWSProvider(credRef, region string) *AWSProvider {
	if region == "" {
		region = "us-east-1"
	}
	return &AWSProvider{
		credRef: credRef,
		region:  region,
	}
}

func (p *AWSProvider) GetProviderType() registry.ProviderType {
	return registry.ProviderAWS
}

func (p *AWSProvider) VerifyConnection(ctx context.Context) error {
	if p.credRef == "" {
		return fmt.Errorf("PROVIDER_NOT_CONFIGURED: AWS credentials reference missing")
	}
	return nil
}

func (p *AWSProvider) GetCapabilities(ctx context.Context) registry.CapabilityMatrix {
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

func (p *AWSProvider) ListRegions(ctx context.Context) ([]string, error) {
	if p.credRef == "" {
		return nil, fmt.Errorf("PROVIDER_NOT_CONFIGURED: Cannot discover AWS regions without active credentials")
	}
	return []string{"us-east-1", "us-west-2", "eu-west-1", "ap-south-1", "ap-southeast-1"}, nil
}
