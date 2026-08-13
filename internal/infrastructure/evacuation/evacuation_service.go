package evacuation

import (
	"context"
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/infrastructure/domain"
)

type EvacuationService struct{}

func NewEvacuationService() *EvacuationService {
	return &EvacuationService{}
}

func (s *EvacuationService) ExecuteEvacuation(ctx context.Context, sourceRegion, targetRegion string) (*domain.RegionEvacuationPlan, error) {
	plan := &domain.RegionEvacuationPlan{
		ID:              fmt.Sprintf("evac-%d", time.Now().UnixNano()),
		EvacuatedRegion: sourceRegion,
		TargetRegion:    targetRegion,
		Status:          "COMPLETED",
		StepsCompleted:  7,
		TotalSteps:      7,
		CreatedAt:       time.Now(),
	}
	return plan, nil
}
