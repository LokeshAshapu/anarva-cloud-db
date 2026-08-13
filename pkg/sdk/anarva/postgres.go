package anarva

import (
	"context"
	"encoding/json"
	"fmt"
)

type PostgresInstance struct {
	ID               string  `json:"id"`
	OrganizationID   string  `json:"organizationId"`
	ProjectID        string  `json:"projectId"`
	Name             string  `json:"name"`
	Provider         string  `json:"provider"`
	Version          string  `json:"version"`
	Status           string  `json:"status"`
	RegionID         string  `json:"regionId"`
	CPU              float64 `json:"cpu"`
	MemoryMB         int     `json:"memoryMb"`
	StorageGB        int     `json:"storageGb"`
	NetworkID        string  `json:"networkId"`
	AvailabilityMode string  `json:"availabilityMode"`
	Host             string  `json:"host"`
	Port             int     `json:"port"`
	PublicAccess     bool    `json:"publicAccess"`
	RealityLabel     string  `json:"realityLabel"`
	CreatedAt        string  `json:"createdAt"`
}

type DatabaseHealth struct {
	InstanceID          string  `json:"instanceId"`
	ConnectionAvailable bool    `json:"connectionAvailable"`
	ReplicationStatus   string  `json:"replicationStatus"`
	CPUPct              float64 `json:"cpuPct"`
	MemoryPct           float64 `json:"memoryPct"`
	StorageUsedGB       float64 `json:"storageUsedGb"`
	ActiveConnections   int     `json:"activeConnections"`
	QueryLatencyMs      float64 `json:"queryLatencyMs"`
	SourceQuality       string  `json:"sourceQuality"`
}

type PostgresService struct {
	client *Client
}

func (s *PostgresService) List(ctx context.Context, orgID, projectID string) ([]*PostgresInstance, error) {
	path := fmt.Sprintf("/api/v1/databases?organizationId=%s&projectId=%s", orgID, projectID)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*PostgresInstance `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *PostgresService) Get(ctx context.Context, id string) (*PostgresInstance, error) {
	path := fmt.Sprintf("/api/v1/databases/%s", id)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *PostgresInstance `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *PostgresService) Health(ctx context.Context, id string) (*DatabaseHealth, error) {
	path := fmt.Sprintf("/api/v1/databases/%s/health", id)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *DatabaseHealth `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
