package anarva

import (
	"context"
	"encoding/json"
	"fmt"
)

type VirtualNetwork struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	ProjectID      string `json:"projectId"`
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	RegionID       string `json:"regionId"`
	CIDR           string `json:"cidr"`
	Status         string `json:"status"`
	DNSEnabled     bool   `json:"dnsEnabled"`
	RealityLabel   string `json:"realityLabel"`
	CreatedAt      string `json:"createdAt"`
}

type ConnectivityTestResult struct {
	ID          string  `json:"id"`
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	Protocol    string  `json:"protocol"`
	Port        int     `json:"port"`
	Reachable   bool    `json:"reachable"`
	LatencyMs   float64 `json:"latencyMs"`
	Error       string  `json:"error,omitempty"`
}

type NetworksService struct {
	client *Client
}

func (s *NetworksService) List(ctx context.Context, orgID, projectID string) ([]*VirtualNetwork, error) {
	path := fmt.Sprintf("/api/v1/networks?organizationId=%s&projectId=%s", orgID, projectID)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*VirtualNetwork `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *NetworksService) Get(ctx context.Context, id string) (*VirtualNetwork, error) {
	path := fmt.Sprintf("/api/v1/networks/%s", id)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *VirtualNetwork `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *NetworksService) TestConnectivity(ctx context.Context, src, dest string, port int) (*ConnectivityTestResult, error) {
	path := "/api/v1/network/connectivity-tests"
	body := map[string]interface{}{
		"source":      src,
		"destination": dest,
		"port":        port,
	}
	res, err := s.client.doRequest(ctx, "POST", path, body, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *ConnectivityTestResult `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
