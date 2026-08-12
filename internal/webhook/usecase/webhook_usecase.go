package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/webhook/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type WebhookUseCase struct {
	mu         sync.RWMutex
	endpoints  map[string]*domain.WebhookEndpoint
	deliveries []*domain.WebhookDelivery
}

func NewWebhookUseCase() *WebhookUseCase {
	uc := &WebhookUseCase{
		endpoints:  make(map[string]*domain.WebhookEndpoint),
		deliveries: make([]*domain.WebhookDelivery, 0),
	}
	uc.seedDefaults()
	return uc
}

func (uc *WebhookUseCase) seedDefaults() {
	now := time.Now()
	secret := "whsec_live_9f82a1bc3d4e5f67"
	hash := domain.ComputeHMACSignature([]byte("seed"), secret)

	ep := &domain.WebhookEndpoint{
		ID:             "whe-101",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		URL:            "https://api.anarva.io/v1/webhooks/receive",
		Description:    "Production Deployment Webhook Notification",
		Status:         "ACTIVE",
		SecretPrefix:   "whsec_live_9f...",
		SecretHash:     hash,
		Events:         []string{"resource.created", "provisioning.completed", "resource.drift_detected"},
		CreatedAt:      now.Add(-12 * time.Hour),
		UpdatedAt:      now,
	}
	uc.endpoints[ep.ID] = ep

	del := &domain.WebhookDelivery{
		ID:           "del-101",
		EndpointID:   "whe-101",
		EventID:      "evt-88f1",
		EventType:    "provisioning.completed",
		Status:       "SUCCESS",
		Attempts:     1,
		ResponseCode: 200,
		DeliveredAt:  now.Add(-10 * time.Minute),
		CreatedAt:    now.Add(-10 * time.Minute),
	}
	uc.deliveries = append(uc.deliveries, del)
}

func (uc *WebhookUseCase) CreateEndpoint(ctx context.Context, targetURL, description, orgID, projectID string, events []string) (*domain.WebhookEndpoint, string, error) {
	if err := domain.ValidateWebhookURL(targetURL); err != nil {
		return nil, "", appErrors.New(appErrors.CodeInvalidInput, err.Error())
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	bytes := make([]byte, 20)
	_, _ = rand.Read(bytes)
	rawSecret := fmt.Sprintf("whsec_live_%s", hex.EncodeToString(bytes))
	secretHash := domain.ComputeHMACSignature([]byte(rawSecret), rawSecret)
	prefix := fmt.Sprintf("whsec_live_%s...", hex.EncodeToString(bytes[:4]))

	if len(events) == 0 {
		events = []string{"resource.created", "resource.deleted", "provisioning.completed"}
	}

	now := time.Now()
	ep := &domain.WebhookEndpoint{
		ID:             fmt.Sprintf("whe-%d", now.UnixNano()/1e6),
		OrganizationID: orgID,
		ProjectID:      projectID,
		URL:            targetURL,
		Description:    description,
		Status:         "ACTIVE",
		SecretPrefix:   prefix,
		SecretHash:     secretHash,
		Events:         events,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	uc.endpoints[ep.ID] = ep
	return ep, rawSecret, nil
}

func (uc *WebhookUseCase) ListEndpoints(ctx context.Context, projectID string) []*domain.WebhookEndpoint {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	var list []*domain.WebhookEndpoint
	for _, ep := range uc.endpoints {
		if projectID == "" || ep.ProjectID == projectID {
			list = append(list, ep)
		}
	}
	return list
}

func (uc *WebhookUseCase) DeleteEndpoint(ctx context.Context, id string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if _, ok := uc.endpoints[id]; !ok {
		return appErrors.New(appErrors.CodeNotFound, "webhook endpoint not found")
	}
	delete(uc.endpoints, id)
	return nil
}

func (uc *WebhookUseCase) ListDeliveries(ctx context.Context, endpointID string) []*domain.WebhookDelivery {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	var list []*domain.WebhookDelivery
	for _, d := range uc.deliveries {
		if endpointID == "" || d.EndpointID == endpointID {
			list = append(list, d)
		}
	}
	return list
}
