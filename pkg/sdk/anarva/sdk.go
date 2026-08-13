package anarva

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client

	Compute      *ComputeService
	Database     *DatabaseService
	Databases    *PostgresService
	Storage      *StorageService
	Network      *NetworkService
	Networks     *NetworksService
	LoadBalancers *LoadBalancersService
	Applications *ApplicationsService
	Provisioning *ProvisioningService
	Projects     *ProjectsService
	Providers    *ProvidersService
}

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"requestId"`
	StatusCode int    `json:"statusCode"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Anarva API Error [%s]: %s (RequestID: %s, Status: %d)", e.Code, e.Message, e.RequestID, e.StatusCode)
}

func NewClient(apiKey string, baseURLs ...string) *Client {
	url := "http://localhost:8080"
	if len(baseURLs) > 0 && baseURLs[0] != "" {
		url = baseURLs[0]
	}

	c := &Client{
		BaseURL: url,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	c.Compute = &ComputeService{client: c}
	c.Database = &DatabaseService{client: c}
	c.Databases = &PostgresService{client: c}
	c.Storage = &StorageService{client: c}
	c.Network = &NetworkService{client: c}
	c.Networks = &NetworksService{client: c}
	c.LoadBalancers = &LoadBalancersService{client: c}
	c.Applications = &ApplicationsService{client: c}
	c.Provisioning = &ProvisioningService{client: c}
	c.Projects = &ProjectsService{client: c}
	c.Providers = &ProvidersService{client: c}

	return c
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, idempotencyKey string) ([]byte, error) {
	fullURL := fmt.Sprintf("%s%s", c.BaseURL, path)

	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	req.Header.Set("X-Request-ID", fmt.Sprintf("sdk_req_%d", time.Now().UnixNano()/1e6))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error struct {
				Code      string `json:"code"`
				Message   string `json:"message"`
				RequestID string `json:"requestId"`
			} `json:"error"`
		}
		if err := json.Unmarshal(respBytes, &errResp); err == nil && errResp.Error.Code != "" {
			return nil, &APIError{
				Code:       errResp.Error.Code,
				Message:    errResp.Error.Message,
				RequestID:  errResp.Error.RequestID,
				StatusCode: resp.StatusCode,
			}
		}
		return nil, &APIError{
			Code:       "HTTP_ERROR",
			Message:    string(respBytes),
			RequestID:  resp.Header.Get("X-Request-ID"),
			StatusCode: resp.StatusCode,
		}
	}

	return respBytes, nil
}

// Service definitions
type ComputeService struct{ client *Client }

func (s *ComputeService) List(ctx context.Context, projectID string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/api/v1/compute?projectId=%s", projectID)
	res, err := s.client.doRequest(ctx, "GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	var wrapped struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(res, &wrapped); err == nil && wrapped.Data != nil {
		return wrapped.Data, nil
	}
	var direct []map[string]interface{}
	_ = json.Unmarshal(res, &direct)
	return direct, nil
}

type DatabaseService struct{ client *Client }
type StorageService struct{ client *Client }
type NetworkService struct{ client *Client }
type ProvisioningService struct{ client *Client }
type ProjectsService struct{ client *Client }
type ProvidersService struct{ client *Client }
