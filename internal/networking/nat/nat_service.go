package nat

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type NatService struct{}

func NewNatService() *NatService {
	return &NatService{}
}

func (s *NatService) CreateNatGateway(networkID, subnetID string) *domain.NatGateway {
	return &domain.NatGateway{
		ID:                     fmt.Sprintf("nat-%s", networkID),
		NetworkID:              networkID,
		SubnetID:               subnetID,
		Status:                 "SIMULATED",
		PublicAddressReference: "eip-alloc-sim-101",
		ProviderResourceId:     fmt.Sprintf("docker-nat-sim-%s", networkID),
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
}
