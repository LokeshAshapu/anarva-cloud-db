package importengine

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
)

type ImportEngine struct {
	mappingRepo *mapping.MappingRepository
}

func NewImportEngine(mappingRepo *mapping.MappingRepository) *ImportEngine {
	return &ImportEngine{mappingRepo: mappingRepo}
}

func (e *ImportEngine) ImportResource(ctx context.Context, provider, providerResourceID, resourceType, region string) (*mapping.ProviderResourceMapping, error) {
	anarvaID := fmt.Sprintf("imported-%s-%d", providerResourceID, time.Now().UnixNano())

	m := &mapping.ProviderResourceMapping{
		AnarvaResourceID:     anarvaID,
		Provider:             provider,
		ProviderResourceID:   providerResourceID,
		ProviderResourceType: resourceType,
		Region:               region,
		Status:               "IMPORTED",
		Managed:              false, // Default: MANAGED = false until adopted
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := e.mappingRepo.SaveMapping(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (e *ImportEngine) AdoptResource(ctx context.Context, anarvaResourceID string) (*mapping.ProviderResourceMapping, error) {
	m, err := e.mappingRepo.GetMapping(anarvaResourceID)
	if err != nil {
		return nil, err
	}

	m.Managed = true
	m.Status = "ACTIVE"
	if err := e.mappingRepo.SaveMapping(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (e *ImportEngine) ReleaseResource(ctx context.Context, anarvaResourceID string) error {
	m, err := e.mappingRepo.GetMapping(anarvaResourceID)
	if err != nil {
		return err
	}

	// Release only removes Anarva management; does NOT destroy cloud resource
	m.Managed = false
	m.Status = "RELEASED"
	return e.mappingRepo.SaveMapping(m)
}
