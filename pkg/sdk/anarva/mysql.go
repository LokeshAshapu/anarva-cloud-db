package anarva

import (
	"context"
	"encoding/json"
	"fmt"
)

type MySQLInstance struct {
	ID                 string `json:"id"`
	OrganizationID     string `json:"organizationId"`
	ProjectID          string `json:"projectId"`
	Name               string `json:"name"`
	Provider           string `json:"provider"`
	Version            string `json:"version"`
	Status             string `json:"status"`
	RegionID           string `json:"regionId"`
	ZoneID             string `json:"zoneId"`
	CPU                int    `json:"cpu"`
	MemoryMB           int    `json:"memoryMb"`
	StorageGB          int    `json:"storageGb"`
	StorageType        string `json:"storageType"`
	NetworkID          string `json:"networkId"`
	SubnetID           string `json:"subnetId"`
	AvailabilityMode   string `json:"availabilityMode"`
	BackupMode         string `json:"backupMode"`
	MaintenanceWindow  string `json:"maintenanceWindow"`
	ProviderResourceID string `json:"providerResourceId"`
	Port               int    `json:"port"`
	RealityLabel       string `json:"realityLabel"`
	CreatedAt          string `json:"createdAt"`
}

type MySQLService struct {
	client *Client
}

func (s *MySQLService) List(ctx context.Context, orgID, projectID string) ([]*MySQLInstance, error) {
	path := fmt.Sprintf("/api/v1/mysql/databases?organizationId=%s&projectId=%s", orgID, projectID)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*MySQLInstance `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *MySQLService) Get(ctx context.Context, id string) (*MySQLInstance, error) {
	path := fmt.Sprintf("/api/v1/mysql/databases/%s", id)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *MySQLInstance `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *MySQLService) Start(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/mysql/databases/%s/start", id)
	_, err := s.client.doRequest(ctx, "POST", path, nil, "")
	return err
}

func (s *MySQLService) Stop(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/mysql/databases/%s/stop", id)
	_, err := s.client.doRequest(ctx, "POST", path, nil, "")
	return err
}

func (s *MySQLService) Restart(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/mysql/databases/%s/restart", id)
	_, err := s.client.doRequest(ctx, "POST", path, nil, "")
	return err
}
