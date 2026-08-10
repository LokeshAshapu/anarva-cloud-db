package providers

import (
	"context"
)

// VPCSpec defines Virtual Private Cloud network configuration.
type VPCSpec struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	CIDR      string `json:"cidr"`
	Region    string `json:"region"`
}

// SecurityRule defines firewall ingress/egress rules.
type SecurityRule struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	SourceIP string `json:"source_ip"`
	Action   string `json:"action"` // ALLOW / DENY
}

// NetworkProvider defines provider-independent interface for VPC and security group abstractions.
type NetworkProvider interface {
	CreateVPC(ctx context.Context, spec VPCSpec) (*VPCSpec, error)
	DeleteVPC(ctx context.Context, vpcID string) error
	ConfigureSecurityGroup(ctx context.Context, vpcID string, rules []SecurityRule) error
	GetVPCHealth(ctx context.Context, vpcID string) (string, error)
}
