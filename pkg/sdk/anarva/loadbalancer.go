package anarva

import (
	"context"
	"encoding/json"
	"fmt"
)

type LoadBalancer struct {
	ID                string   `json:"id"`
	OrganizationID    string   `json:"organizationId"`
	ProjectID         string   `json:"projectId"`
	Name              string   `json:"name"`
	Provider          string   `json:"provider"`
	Type              string   `json:"type"`
	Scheme            string   `json:"scheme"`
	NetworkID         string   `json:"networkId"`
	SubnetIDs         []string `json:"subnetIds"`
	Status            string   `json:"status"`
	IPReference       string   `json:"ipReference"`
	HostnameReference string   `json:"hostnameReference"`
	RealityLabel      string   `json:"realityLabel"`
	CreatedAt         string   `json:"createdAt"`
}

type Application struct {
	ID                  string `json:"id"`
	OrganizationID      string `json:"organizationId"`
	ProjectID           string `json:"projectId"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	NetworkID           string `json:"networkId"`
	DeploymentReference string `json:"deploymentReference"`
	LoadBalancerID      string `json:"loadBalancerId"`
	DomainReference     string `json:"domainReference"`
	ContainerImage      string `json:"containerImage"`
	ACUCount            int    `json:"acuCount"`
	Health              string `json:"health"`
	CreatedAt           string `json:"createdAt"`
}

type LoadBalancersService struct {
	client *Client
}

func (s *LoadBalancersService) List(ctx context.Context, orgID, projectID string) ([]*LoadBalancer, error) {
	path := fmt.Sprintf("/api/v1/load-balancers?organizationId=%s&projectId=%s", orgID, projectID)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*LoadBalancer `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *LoadBalancersService) Get(ctx context.Context, id string) (*LoadBalancer, error) {
	path := fmt.Sprintf("/api/v1/load-balancers/%s", id)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *LoadBalancer `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

type ApplicationsService struct {
	client *Client
}

func (s *ApplicationsService) List(ctx context.Context, orgID, projectID string) ([]*Application, error) {
	path := fmt.Sprintf("/api/v1/applications?organizationId=%s&projectId=%s", orgID, projectID)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*Application `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
