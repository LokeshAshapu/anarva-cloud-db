package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/terraform/client"
)

func TestProvider_PlanApplyRefreshPlanNoPerpetualDiff(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"res-rds-stable","name":"stable-db","engine":"POSTGRESQL","storageGb":25,"acuUnits":2.0,"multiAz":true,"status":"AVAILABLE"}}`))
	}))
	defer ts.Close()

	p := NewProvider(client.Config{
		APIKey: "anarva_live_testkey123",
		APIURL: ts.URL,
	})

	state1, err := p.ReadDatabase(context.Background(), "res-rds-stable")
	require.NoError(t, err)

	state2, err := p.ReadDatabase(context.Background(), "res-rds-stable")
	require.NoError(t, err)

	// Assert no perpetual diff
	assert.Equal(t, state1.ID, state2.ID)
	assert.Equal(t, state1.Name, state2.Name)
	assert.Equal(t, state1.MultiAZ, state2.MultiAZ)
	assert.Equal(t, state1.StorageGB, state2.StorageGB)
}

func TestProvider_ExternalDriftDetection(t *testing.T) {
	multiAZVal := true
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if multiAZVal {
			_, _ = w.Write([]byte(`{"data":{"id":"res-rds-drift","name":"drift-db","multiAz":true,"status":"AVAILABLE"}}`))
		} else {
			_, _ = w.Write([]byte(`{"data":{"id":"res-rds-drift","name":"drift-db","multiAz":false,"status":"AVAILABLE"}}`))
		}
	}))
	defer ts.Close()

	p := NewProvider(client.Config{
		APIKey: "anarva_live_testkey123",
		APIURL: ts.URL,
	})

	// Initial Read: MultiAZ = true
	stateInitial, err := p.ReadDatabase(context.Background(), "res-rds-drift")
	require.NoError(t, err)
	assert.True(t, stateInitial.MultiAZ)

	// External change occurs in AWS: MultiAZ = false
	multiAZVal = false

	// Refresh Read detects drift
	stateDrifted, err := p.ReadDatabase(context.Background(), "res-rds-drift")
	require.NoError(t, err)
	assert.False(t, stateDrifted.MultiAZ)
}

func TestProvider_StaleObservationRetention(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"code":"OBSERVATION_UNAVAILABLE","message":"Observation Engine telemetry delayed"}}`))
	}))
	defer ts.Close()

	p := NewProvider(client.Config{
		APIKey: "anarva_live_testkey123",
		APIURL: ts.URL,
	})

	// Telemetry delayed should NOT return nil (should not remove state)
	state, err := p.ReadDatabase(context.Background(), "res-rds-telemetry-delay")
	assert.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "Observation Engine telemetry delayed")
}

func TestProvider_ConcurrentOperations(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"res-concurrent","status":"RUNNING"}}`))
	}))
	defer ts.Close()

	p := NewProvider(client.Config{
		APIKey: "anarva_live_testkey123",
		APIURL: ts.URL,
	})

	var wg sync.WaitGroup
	errCh := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := p.ReadCompute(context.Background(), "res-concurrent")
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err)
	}
}
