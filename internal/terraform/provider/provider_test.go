package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/terraform/client"
)

func TestProvider_DatabaseLifecycle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer anarva_live_testkey123", r.Header.Get("Authorization"))

		switch r.URL.Path {
		case "/api/v1/databases":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data":{"id":"res-rds-test-db","status":"AVAILABLE"},"requestId":"req_create_01"}`))
		case "/api/v1/databases/res-rds-test-db":
			if r.Method == "GET" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"id":"res-rds-test-db","name":"test-db","status":"AVAILABLE"}}`))
			} else if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	p := NewProvider(client.Config{
		APIKey:    "anarva_live_testkey123",
		APIURL:    ts.URL,
		ProjectID: "proj-default",
	})

	// Create
	plan := &DatabaseResourceState{
		Name:      "test-db",
		Engine:    "POSTGRESQL",
		StorageGB: 25,
		ACUUnits:  2.0,
		MultiAZ:   true,
	}
	state, err := p.CreateDatabase(context.Background(), plan)
	require.NoError(t, err)
	assert.Equal(t, "res-rds-test-db", state.ID)
	assert.Equal(t, "AVAILABLE", state.Status)

	// Read
	readState, err := p.ReadDatabase(context.Background(), state.ID)
	require.NoError(t, err)
	assert.NotNil(t, readState)
	assert.Equal(t, "test-db", readState.Name)

	// Delete
	err = p.DeleteDatabase(context.Background(), state.ID)
	require.NoError(t, err)
}

func TestProvider_Read404_RemovesFromState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"RESOURCE_NOT_FOUND","message":"Database not found"}}`))
	}))
	defer ts.Close()

	p := NewProvider(client.Config{
		APIKey: "anarva_live_testkey123",
		APIURL: ts.URL,
	})

	state, err := p.ReadDatabase(context.Background(), "res-missing")
	require.NoError(t, err)
	assert.Nil(t, state) // 404 removes resource from state
}

func TestProvider_SecretRedaction(t *testing.T) {
	os.Setenv("ANARVA_API_KEY", "anarva_live_secretkey999")
	defer os.Unsetenv("ANARVA_API_KEY")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":"ACCESS_DENIED","message":"Forbidden for key anarva_live_secretkey999"}}`))
	}))
	defer ts.Close()

	p := NewProvider(client.Config{
		APIKey: "anarva_live_secretkey999",
		APIURL: ts.URL,
	})

	_, err := p.CreateDatabase(context.Background(), &DatabaseResourceState{Name: "fail-db"})
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "anarva_live_secretkey999")
	assert.Contains(t, err.Error(), "[REDACTED_API_KEY]")
}
