package anarva

import (
	"context"
	"encoding/json"
	"fmt"
)

type ProviderInfo struct {
	ID                  string           `json:"id"`
	Name                string           `json:"name"`
	Type                string           `json:"type"`
	Status              string           `json:"status"`
	CredentialReference string           `json:"credentialReference"`
	Capabilities        map[string]bool  `json:"capabilities"`
	Regions             []string         `json:"regions"`
	RealityLabel        string           `json:"realityLabel"`
}

type ProviderResourceMapping struct {
	AnarvaResourceID   string `json:"anarvaResourceId"`
	Provider           string `json:"provider"`
	ProviderResourceID string `json:"providerResourceId"`
	Region             string `json:"region"`
	Managed            bool   `json:"managed"`
}

type ProvidersService struct {
	client *Client
}

func (s *ProvidersService) List(ctx context.Context) ([]*ProviderInfo, error) {
	res, err := s.client.doRequest(ctx, "GET", "/api/v1/providers", nil, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []*ProviderInfo `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *ProvidersService) Verify(ctx context.Context, id, credRef string) (*ProviderInfo, error) {
	path := fmt.Sprintf("/api/v1/providers/%s/verify", id)
	body := map[string]string{"credentialReference": credRef}
	res, err := s.client.doRequest(ctx, "POST", path, body, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *ProviderInfo `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *ProvidersService) Import(ctx context.Context, provider, providerResourceID, resourceType, region string) (*ProviderResourceMapping, error) {
	body := map[string]string{
		"provider":           provider,
		"providerResourceId": providerResourceID,
		"resourceType":       resourceType,
		"region":             region,
	}
	res, err := s.client.doRequest(ctx, "POST", "/api/v1/resources/import", body, "")
	if err != nil {
		return nil, err
	}
	var response struct {
		Data *ProviderResourceMapping `json:"data"`
	}
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
