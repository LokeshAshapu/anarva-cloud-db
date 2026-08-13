package gateway

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type GatewayService struct{}

func NewGatewayService() *GatewayService {
	return &GatewayService{}
}

func (s *GatewayService) CreateInternetGateway(networkID string) *domain.InternetGateway {
	return &domain.InternetGateway{
		ID:                 fmt.Sprintf("igw-%s", networkID),
		NetworkID:          networkID,
		Status:             "AVAILABLE",
		ProviderResourceId: fmt.Sprintf("docker-igw-sim-%s", networkID),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}
