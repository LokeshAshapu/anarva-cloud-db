package waf

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type WAFService struct{}

func NewWAFService() *WAFService {
	return &WAFService{}
}

func (s *WAFService) CreateWAFPolicy(lbID string, mode domain.WAFMode, rules []string) *domain.WAFPolicy {
	if len(rules) == 0 {
		rules = []string{"SQL_INJECTION", "XSS_PROTECTION", "PATH_TRAVERSAL"}
	}
	return &domain.WAFPolicy{
		ID:             fmt.Sprintf("waf-%d", time.Now().UnixNano()),
		LoadBalancerID: lbID,
		Rules:          rules,
		Mode:           mode,
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
	}
}
