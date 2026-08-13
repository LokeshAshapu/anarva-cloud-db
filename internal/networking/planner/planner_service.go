package planner

import (
	"fmt"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type NetworkPlan struct {
	NetworkCIDR            string   `json:"networkCidr"`
	Subnets                []string `json:"subnets"`
	Routes                 []string `json:"routes"`
	SecurityGroups         []string `json:"securityGroups"`
	Gateways               []string `json:"gateways"`
	DNSEnabled             bool     `json:"dnsEnabled"`
	EstimatedMonthlyCostUSD float64  `json:"estimatedMonthlyCostUsd"`
	Provider               string   `json:"provider"`
	RealityLabel           string   `json:"realityLabel"`
}

type PlannerService struct{}

func NewPlannerService() *PlannerService {
	return &PlannerService{}
}

func (s *PlannerService) GeneratePlan(net *domain.VirtualNetwork, subnets []string) *NetworkPlan {
	return &NetworkPlan{
		NetworkCIDR: net.CIDR,
		Subnets:     subnets,
		Routes:      []string{fmt.Sprintf("%s -> LOCAL", net.CIDR)},
		SecurityGroups: []string{
			"default (DENY inbound, ALLOW outbound)",
		},
		Gateways: []string{
			"InternetGateway (LOCAL_SIMULATED)",
		},
		DNSEnabled:              net.DNSEnabled,
		EstimatedMonthlyCostUSD: 0.0,
		Provider:               net.Provider,
		RealityLabel:           "LOCAL_NETWORK (LIMITED_CAPABILITIES)",
	}
}
