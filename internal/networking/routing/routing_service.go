package routing

import (
	"fmt"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type RoutingService struct{}

func NewRoutingService() *RoutingService {
	return &RoutingService{}
}

func (s *RoutingService) CreateDefaultRouteTable(networkID, cidr string) *domain.RouteTable {
	rtID := fmt.Sprintf("rt-default-%s", networkID)
	return &domain.RouteTable{
		ID:        rtID,
		NetworkID: networkID,
		Name:      "default-route-table",
		Status:    "AVAILABLE",
		Routes: []domain.Route{
			{
				ID:           fmt.Sprintf("route-local-%s", networkID),
				RouteTableID: rtID,
				Destination:  cidr,
				Target:       "local",
				TargetType:   domain.TargetLocal,
				Status:       "ACTIVE",
				CreatedAt:    time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *RoutingService) ValidateRoute(route *domain.Route) error {
	switch route.TargetType {
	case domain.TargetLocal, domain.TargetInternetGateway, domain.TargetNatGateway, domain.TargetNetworkInterface, domain.TargetPeering:
		return nil
	default:
		return fmt.Errorf("unsupported route target type '%s'", route.TargetType)
	}
}
