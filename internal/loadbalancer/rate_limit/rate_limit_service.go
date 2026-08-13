package rate_limit

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/loadbalancer/domain"
)

type RateLimitService struct{}

func NewRateLimitService() *RateLimitService {
	return &RateLimitService{}
}

func (s *RateLimitService) CreatePolicy(projectID, scope string, requests, windowSec int, action string) *domain.RateLimitPolicy {
	return &domain.RateLimitPolicy{
		ID:        fmt.Sprintf("rlp-%d", time.Now().UnixNano()),
		ProjectID: projectID,
		Scope:     scope,
		Requests:  requests,
		WindowSec: windowSec,
		Action:    action,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}
}
