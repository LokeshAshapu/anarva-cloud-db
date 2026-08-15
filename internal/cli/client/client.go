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

	"github.com/anarva-cloud/anarva-cloud-db/internal/cli/config"
)

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

type Client struct {
	HTTPClient *http.Client
	Profile    *config.Profile
	Debug      bool
	OutputFmt  string
	Quiet      bool
	NoColor    bool
}

func NewClient(prof *config.Profile, debug bool, outputFmt string, quiet bool, noColor bool) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Profile:    prof,
		Debug:      debug,
		OutputFmt:  outputFmt,
		Quiet:      quiet,
		NoColor:    noColor,
	}
}

func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body interface{}) (*APIResponse, int, error) {
	url := strings.TrimRight(c.Profile.APIURL, "/") + endpoint

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

	// Get token from env or profile
	key := os.Getenv("ANARVA_API_KEY")
	if key == "" {
		key = c.Profile.APIKey
	}

	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	if c.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, url)
		if key != "" {
			fmt.Fprintf(os.Stderr, "[DEBUG] Authorization: Bearer REDACTED\n")
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 6, fmt.Errorf("API unavailable: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 1, fmt.Errorf("failed to read response: %w", err)
	}

	if c.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response Status: %d\n", resp.StatusCode)
	}

	var apiRes APIResponse
	_ = json.Unmarshal(respBytes, &apiRes)

	exitCode := c.MapStatusCodeToExitCode(resp.StatusCode)
	if resp.StatusCode >= 400 {
		errMsg := "API operation failed"
		reqID := ""
		if apiRes.Error != nil {
			errMsg = apiRes.Error.Message
			reqID = apiRes.Error.RequestID
		} else if len(respBytes) > 0 {
			errMsg = string(respBytes)
		}

		if reqID != "" {
			return &apiRes, exitCode, fmt.Errorf("Error: %s (Request ID: %s)", errMsg, reqID)
		}
		return &apiRes, exitCode, fmt.Errorf("Error: %s", errMsg)
	}

	return &apiRes, 0, nil
}

func (c *Client) MapStatusCodeToExitCode(statusCode int) int {
	switch statusCode {
	case 200, 201, 202, 204:
		return 0
	case 400:
		return 2 // Invalid usage
	case 401:
		return 3 // Auth failure
	case 403:
		return 4 // Authorization failure
	case 404:
		return 5 // Resource not found
	case 429:
		return 7 // Rate limited
	case 502, 503, 504:
		return 6 // API unavailable
	default:
		return 1 // General error
	}
}
