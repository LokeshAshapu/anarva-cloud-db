package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/cli/client"
	"github.com/anarva-cloud/anarva-cloud-db/internal/cli/config"
)

func TestCLI_ConfigAndProfileManagement(t *testing.T) {
	os.Setenv("ANARVA_API_KEY", "anarva_test_1234567890abcdef")
	defer os.Unsetenv("ANARVA_API_KEY")

	cfg := config.DefaultConfig()
	assert.Equal(t, "default", cfg.ActiveProfile)

	prof := cfg.GetCurrentProfile()
	assert.NotNil(t, prof)
	assert.Equal(t, "anarva_test_1234567890abcdef", prof.APIKey)
}

func TestCLI_ClientExitCodeMapping(t *testing.T) {
	prof := config.DefaultProfile()
	c := client.NewClient(prof, false, "table", false, false)

	assert.Equal(t, 0, c.MapStatusCodeToExitCode(200))
	assert.Equal(t, 2, c.MapStatusCodeToExitCode(400))
	assert.Equal(t, 3, c.MapStatusCodeToExitCode(401))
	assert.Equal(t, 4, c.MapStatusCodeToExitCode(403))
	assert.Equal(t, 5, c.MapStatusCodeToExitCode(404))
	assert.Equal(t, 6, c.MapStatusCodeToExitCode(503))
	assert.Equal(t, 7, c.MapStatusCodeToExitCode(429))
}

func TestCLI_ClientDoRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer anarva_live_testkey123", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"status":"OK"},"requestId":"req_test_01"}`))
	}))
	defer ts.Close()

	prof := &config.Profile{
		APIURL:         ts.URL,
		OrganizationID: "org-test",
		ProjectID:      "proj-test",
		APIKey:         "anarva_live_testkey123",
	}

	c := client.NewClient(prof, true, "json", false, false)
	res, code, err := c.DoRequest(context.Background(), "GET", "/api/v1/health", nil)

	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.NotNil(t, res)
}
