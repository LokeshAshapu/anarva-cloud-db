package compute_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
)

func TestPhase70J_RemoteComputeProviderClient(t *testing.T) {
	// Setup in-process mock worker HTTP server
	var lastIdempotencyKey string
	var lastAuthHeader string
	var lastWorkerTokenHeader string

	mockWorker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastIdempotencyKey = r.Header.Get("Idempotency-Key")
		lastAuthHeader = r.Header.Get("Authorization")
		lastWorkerTokenHeader = r.Header.Get("X-Anarva-Worker-Token")

		// Route mock worker endpoints
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/workloads":
			var req computeProvider.CreateWorkloadRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.WorkloadID == "force-409" {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"code":"CONFLICT","message":"Workload already exists"}`))
				return
			}
			if req.WorkloadID == "force-500" {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"code":"INTERNAL_ERROR","message":"Worker internal failure"}`))
				return
			}
			if req.WorkloadID == "force-malformed" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`invalid-json-response`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
				WorkloadID: "workload-" + req.WorkloadID,
				Status:     "RUNNING",
				Health:     "HEALTHY",
				PrivateIP:  "10.200.1.5",
				PublicIP:   "20.200.1.5",
				CreatedAt:  time.Now(),
			})

		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/workloads/") && strings.HasSuffix(r.URL.Path, "/metrics"):
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/workloads/"), "/metrics")
			if id == "force-404" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"code":"NOT_FOUND","message":"Workload not found"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkerMetricsResponse{
				WorkloadID: id,
				CPUUsage:   12.5,
				MemoryMB:   256,
				NetworkRx:  1024,
				NetworkTx:  2048,
			})

		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/execute"):
			var req computeProvider.ExecuteCommandWorkerRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(computeProvider.ExecuteCommandWorkerResponse{
				ExitCode: 0,
				Stdout:   "PHASE_70J_WORKER_STDOUT: " + req.Command,
				Stderr:   "",
				Executed: time.Now(),
			})

		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/start"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"RUNNING"}`))

		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/stop"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"STOPPED"}`))

		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/restart"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"RUNNING"}`))

		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/v1/workloads/"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/workloads/")
			if id == "force-404" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/workloads/"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/workloads/")
			if id == "force-404" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"code":"NOT_FOUND","message":"Workload not found"}`))
				return
			}
			if id == "force-401" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code":"UNAUTHORIZED","message":"Worker token invalid"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
				WorkloadID: id,
				Status:     "RUNNING",
				Health:     "HEALTHY",
				PrivateIP:  "10.200.1.5",
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockWorker.Close()

	clientToken := "sec-worker-auth-token-70j"
	provider, err := computeProvider.NewRemoteComputeProvider(mockWorker.URL, clientToken, mockWorker.Client())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("1. CreateInstance sends HTTP POST /v1/workloads with Idempotency-Key", func(t *testing.T) {
		inst := &computeDomain.ComputeInstance{
			ID:             "acu-70j-node-01",
			ResourceID:     "arnv:vm:us-east-1:proj-test:compute/node-01",
			Name:           "node-01",
			DockerImage:    "nginx:alpine",
			VCPU:           1.0,
			MemoryMB:       2048,
			OrganizationID: "org-test",
			ProjectID:      "proj-test",
			EnvVars:        map[string]string{"DB_SECRET": "my_super_secret_val"},
		}

		created, createErr := provider.CreateInstance(ctx, inst)
		require.NoError(t, createErr)

		assert.Equal(t, "workload-acu-70j-node-01", created.ProviderInstanceID)
		assert.Equal(t, computeDomain.StatusRunning, created.Status)
		assert.Equal(t, "arnv:vm:us-east-1:proj-test:compute/node-01", lastIdempotencyKey, "Idempotency-Key header must match inst.ResourceID")
		assert.Equal(t, "Bearer "+clientToken, lastAuthHeader)
		assert.Equal(t, clientToken, lastWorkerTokenHeader)
	})

	t.Run("2. GetInstance returns instance state", func(t *testing.T) {
		getInst, getErr := provider.GetInstance(ctx, "acu-70j-node-01")
		require.NoError(t, getErr)
		assert.Equal(t, "acu-70j-node-01", getInst.ID)
	})

	t.Run("3. StartInstance calls POST /v1/workloads/{id}/start", func(t *testing.T) {
		err := provider.StartInstance(ctx, "acu-70j-node-01")
		require.NoError(t, err)
	})

	t.Run("4. StopInstance calls POST /v1/workloads/{id}/stop", func(t *testing.T) {
		err := provider.StopInstance(ctx, "acu-70j-node-01")
		require.NoError(t, err)
	})

	t.Run("5. RestartInstance calls POST /v1/workloads/{id}/restart", func(t *testing.T) {
		err := provider.RestartInstance(ctx, "acu-70j-node-01")
		require.NoError(t, err)
	})

	t.Run("6. ExecuteCommand calls POST /v1/workloads/{id}/execute", func(t *testing.T) {
		req := &computeDomain.CommandExecutionRequest{Command: "uptime"}
		res, execErr := provider.ExecuteCommand(ctx, "acu-70j-node-01", req)
		require.NoError(t, execErr)
		assert.Equal(t, 0, res.ExitCode)
		assert.Contains(t, res.Stdout, "PHASE_70J_WORKER_STDOUT: uptime")
	})

	t.Run("7. GetInstanceMetrics calls GET /v1/workloads/{id}/metrics", func(t *testing.T) {
		metrics, mErr := provider.GetInstanceMetrics(ctx, "acu-70j-node-01")
		require.NoError(t, mErr)
		assert.Equal(t, 12.5, metrics["cpuUsagePercent"])
		assert.Equal(t, 256, metrics["memoryUsageMb"])
	})

	t.Run("8. DeleteInstance calls DELETE /v1/workloads/{id}", func(t *testing.T) {
		err := provider.DeleteInstance(ctx, "acu-70j-node-01")
		require.NoError(t, err)
	})

	t.Run("9. Worker Timeout Handling", func(t *testing.T) {
		slowWorker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
		}))
		defer slowWorker.Close()

		slowClient := slowWorker.Client()
		slowClient.Timeout = 50 * time.Millisecond
		slowProv, _ := computeProvider.NewRemoteComputeProvider(slowWorker.URL, "token", slowClient)

		shortCtx, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
		defer cancel()

		_, err := slowProv.CreateInstance(shortCtx, &computeDomain.ComputeInstance{ID: "timeout-node"})
		assert.ErrorIs(t, err, computeProvider.ErrRemoteWorkerTimeout)
	})

	t.Run("10. Worker Unavailable Handling", func(t *testing.T) {
		badProv, _ := computeProvider.NewRemoteComputeProvider("http://127.0.0.1:59999", "token", nil)
		_, err := badProv.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "unavail-node"})
		assert.ErrorIs(t, err, computeProvider.ErrRemoteWorkerUnavailable)
	})

	t.Run("11. HTTP 401 Unauthorized Handling", func(t *testing.T) {
		unauthProv, _ := computeProvider.NewRemoteComputeProvider(mockWorker.URL, "bad-token", mockWorker.Client())
		// Trigger GET /v1/workloads/force-401
		unauthProv.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "force-401", ProviderInstanceID: "force-401"})
		err := unauthProv.RehydrateInstance(ctx, &computeDomain.ComputeInstance{ID: "force-401", ProviderInstanceID: "force-401"})
		assert.NoError(t, err) // Rehydrate sets status to STOPPED on worker error
	})

	t.Run("12. HTTP 404 Not Found Handling", func(t *testing.T) {
		_, err := provider.GetInstanceMetrics(ctx, "force-404")
		assert.ErrorIs(t, err, computeProvider.ErrRemoteWorkerNotFound)
	})

	t.Run("13. HTTP 409 Conflict Handling", func(t *testing.T) {
		_, err := provider.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "force-409"})
		assert.ErrorIs(t, err, computeProvider.ErrRemoteWorkerConflict)
	})

	t.Run("14. HTTP 500 & Error Sanitization", func(t *testing.T) {
		_, err := provider.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "force-500"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "remote compute provider error")
		assert.NotContains(t, err.Error(), clientToken, "Error message MUST NOT contain worker auth token")
	})

	t.Run("15. Malformed Response Handling", func(t *testing.T) {
		_, err := provider.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "force-malformed"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse remote worker workload response")
	})

	t.Run("16. RehydrateInstance Reads Remote Worker State", func(t *testing.T) {
		inst := &computeDomain.ComputeInstance{
			ID:                 "rehydrate-node",
			ProviderInstanceID: "workload-rehydrate",
		}
		err := provider.RehydrateInstance(ctx, inst)
		require.NoError(t, err)
		assert.Equal(t, computeDomain.StatusRunning, inst.Status)
		assert.Equal(t, computeDomain.HealthHealthy, inst.Health)
	})
}
