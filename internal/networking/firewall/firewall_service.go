package firewall

import (
	"fmt"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type FirewallService struct{}

func NewFirewallService() *FirewallService {
	return &FirewallService{}
}

func (s *FirewallService) ValidateSecurityRule(rule *domain.SecurityRule) error {
	if rule.Direction != domain.DirectionIngress && rule.Direction != domain.DirectionEgress {
		return fmt.Errorf("invalid rule direction '%s': must be INGRESS or EGRESS", rule.Direction)
	}

	if rule.Action != domain.ActionAllow && rule.Action != domain.ActionDeny {
		return fmt.Errorf("invalid rule action '%s': must be ALLOW or DENY", rule.Action)
	}

	// Database Port 5432 Public Access Warning
	if rule.Direction == domain.DirectionIngress && rule.FromPort <= 5432 && rule.ToPort >= 5432 {
		if strings.TrimSpace(rule.Source) == "0.0.0.0/0" && rule.Action == domain.ActionAllow {
			return fmt.Errorf("SECURITY RISK: Inbound PostgreSQL port 5432 rule permits unrestricted public access (0.0.0.0/0). Explicit confirmation required")
		}
	}
	return nil
}

func (s *FirewallService) CreateDefaultSecurityGroup(networkID string) *domain.SecurityGroup {
	return &domain.SecurityGroup{
		ID:          fmt.Sprintf("sg-default-%s", networkID),
		NetworkID:   networkID,
		Name:        "default",
		Description: "Default security group: Deny inbound, allow outbound",
		Status:      "ACTIVE",
		Rules: []domain.SecurityRule{
			{
				ID:          fmt.Sprintf("rule-default-egress-%s", networkID),
				Direction:   domain.DirectionEgress,
				Protocol:    "ALL",
				FromPort:    0,
				ToPort:      65535,
				Destination: "0.0.0.0/0",
				Action:      domain.ActionAllow,
				Priority:    100,
				Description: "Allow all outbound traffic",
			},
		},
	}
}
