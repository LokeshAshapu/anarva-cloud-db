package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	APIKey         string
	APIURL         string
	OrganizationID string
	ProjectID      string
}

type Client struct {
	HTTPClient *http.Client
	Config     Config
}

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

type APIResponse struct {
	Data      interface{} `json:"data,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

func NewClient(cfg Config) *Client {
	if cfg.APIURL == "" {
		if envURL := os.Getenv("ANARVA_API_URL"); envURL != "" {
			cfg.APIURL = envURL
		} else {
			cfg.APIURL = "https://anarva-cloud-db-api.onrender.com"
		}
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("ANARVA_API_KEY")
	}

	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Config:     cfg,
	}
}

func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body interface{}) (*APIResponse, int, error) {
	url := strings.TrimRight(c.Config.APIURL, "/") + endpoint

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 1, fmt.Errorf("failed to encode request body: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 1, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 503, fmt.Errorf("Anarva API unavailable: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	reqID := resp.Header.Get("X-Request-Id")

	var apiRes APIResponse
	_ = json.Unmarshal(respBytes, &apiRes)

	if resp.StatusCode >= 400 {
		errMsg := "API operation failed"
		if apiRes.Error != nil && apiRes.Error.Message != "" {
			errMsg = apiRes.Error.Message
			if apiRes.Error.RequestID != "" {
				reqID = apiRes.Error.RequestID
			}
		} else if len(respBytes) > 0 {
			errMsg = string(respBytes)
		}

		// Redact secrets if present in error message
		errMsg = redactSecrets(errMsg)

		if reqID != "" {
			return &apiRes, resp.StatusCode, fmt.Errorf("[%s] %s (Request ID: %s)", resp.Status, errMsg, reqID)
		}
		return &apiRes, resp.StatusCode, fmt.Errorf("[%s] %s", resp.Status, errMsg)
	}

	return &apiRes, resp.StatusCode, nil
}

func redactSecrets(input string) string {
	// Replaces any raw API secret formats with REDACTED
	return strings.ReplaceAll(input, os.Getenv("ANARVA_API_KEY"), "[REDACTED_API_KEY]")
}
