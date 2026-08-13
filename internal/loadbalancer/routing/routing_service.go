package routing

import (
	"fmt"
	"strings"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type RoutingService struct{}

func NewRoutingService() *RoutingService {
	return &RoutingService{}
}

func (s *RoutingService) ValidateRule(rule *domain.RoutingRule, existingRules []domain.RoutingRule) error {
	if rule.Host == "" && rule.Path == "" {
		return fmt.Errorf("routing rule must specify at least a Host or Path pattern")
	}

	for _, existing := range existingRules {
		if existing.ListenerID == rule.ListenerID && existing.Priority == rule.Priority && existing.ID != rule.ID {
			return fmt.Errorf("priority conflict: priority %d is already assigned to rule '%s'", rule.Priority, existing.ID)
		}
	}
	return nil
}

func (s *RoutingService) MatchRoute(host, path string, rules []domain.RoutingRule) (*domain.RoutingRule, error) {
	var bestMatch *domain.RoutingRule
	highestPriority := -1

	for i := range rules {
		rule := &rules[i]
		if s.matchesHost(host, rule.Host) && s.matchesPath(path, rule.Path) {
			if rule.Priority > highestPriority {
				highestPriority = rule.Priority
				bestMatch = rule
			}
		}
	}

	if bestMatch != nil {
		return bestMatch, nil
	}
	return nil, fmt.Errorf("no routing rule matched host '%s' path '%s'", host, path)
}

func (s *RoutingService) matchesHost(requestHost, ruleHost string) bool {
	if ruleHost == "" || ruleHost == "*" {
		return true
	}
	return strings.EqualFold(requestHost, ruleHost)
}

func (s *RoutingService) matchesPath(requestPath, rulePath string) bool {
	if rulePath == "" || rulePath == "/*" {
		return true
	}
	if strings.HasSuffix(rulePath, "/*") {
		prefix := strings.TrimSuffix(rulePath, "/*")
		return strings.HasPrefix(requestPath, prefix)
	}
	return requestPath == rulePath
}
