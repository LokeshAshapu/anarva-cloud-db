package anarva

import (
	"context"
	"encoding/json"
)

type Region struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Code              string  `json:"code"`
	Provider          string  `json:"provider"`
	Status            string  `json:"status"`
	LatitudeReference float64 `json:"latitudeReference"`
	CountryCode       string  `json:"countryCode"`
	CapacityStatus    string  `json:"capacityStatus"`
	RealityLabel      string  `json:"realityLabel"`
}

type GlobalHealth struct {
	Status        string `json:"status"`
	TotalRegions  int    `json:"totalRegions"`
	ActiveRegions int    `json:"activeRegions"`
	DegradedCount int    `json:"degradedCount"`
	OutageCount   int    `json:"outageCount"`
}

type InfrastructureIncident struct {
	ID        string `json:"id"`
	Severity  string `json:"severity"`
	RegionID  string `json:"regionId"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Summary   string `json:"summary"`
	StartedAt string `json:"startedAt"`
}

type InfrastructureService struct {
	client *Client
}

func (s *InfrastructureService) ListRegions(ctx context.Context) ([]*Region, error) {
	res, err := s.client.doRequest(ctx, "GET", "/api/v1/regions", nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*Region `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *InfrastructureService) GetGlobalHealth(ctx context.Context) (*GlobalHealth, error) {
	res, err := s.client.doRequest(ctx, "GET", "/api/v1/infrastructure/global-health", nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *GlobalHealth `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *InfrastructureService) ListIncidents(ctx context.Context) ([]*InfrastructureIncident, error) {
	res, err := s.client.doRequest(ctx, "GET", "/api/v1/incidents", nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*InfrastructureIncident `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
