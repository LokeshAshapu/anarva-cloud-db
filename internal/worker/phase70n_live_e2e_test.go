package worker_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computeDelivery "github.com/anarva-cloud/anarva-cloud-db/internal/compute/delivery/http"
	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	computeUseCase "github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
	mapping "github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
	"github.com/anarva-cloud/anarva-cloud-db/internal/worker"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/crypto"
)

type memoryComputeRepository struct {
	mu        sync.RWMutex
	instances map[string]*computeDomain.ComputeInstance
}

func newMemoryComputeRepository() *memoryComputeRepository {
	return &memoryComputeRepository{
		instances: make(map[string]*computeDomain.ComputeInstance),
	}
}

func (r *memoryComputeRepository) Create(ctx context.Context, inst *computeDomain.ComputeInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = inst.BeforeSave(nil)
	r.instances[inst.ID] = inst
	return nil
}

func (r *memoryComputeRepository) GetByID(ctx context.Context, id string) (*computeDomain.ComputeInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inst, ok := r.instances[id]
	if !ok || inst.DeletedAt != nil {
		return nil, fmt.Errorf("instance '%s' not found", id)
	}
	_ = inst.AfterFind(nil)
	return inst, nil
}

func (r *memoryComputeRepository) GetTenantScopedByID(ctx context.Context, orgID, projID, id string) (*computeDomain.ComputeInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inst, ok := r.instances[id]
	if !ok || inst.DeletedAt != nil {
		return nil, fmt.Errorf("instance '%s' not found", id)
	}
	if inst.OrganizationID != orgID || inst.ProjectID != projID {
		return nil, fmt.Errorf("TENANT_ISOLATION_VIOLATION: unauthorized instance access for org '%s' and project '%s'", orgID, projID)
	}
	_ = inst.AfterFind(nil)
	return inst, nil
}

func (r *memoryComputeRepository) ListByProjectID(ctx context.Context, projectID string) ([]*computeDomain.ComputeInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*computeDomain.ComputeInstance
	for _, inst := range r.instances {
		if inst.ProjectID == projectID && inst.DeletedAt == nil {
			_ = inst.AfterFind(nil)
			list = append(list, inst)
		}
	}
	return list, nil
}

func (r *memoryComputeRepository) Update(ctx context.Context, inst *computeDomain.ComputeInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = inst.BeforeSave(nil)
	r.instances[inst.ID] = inst
	return nil
}

func (r *memoryComputeRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	inst, ok := r.instances[id]
	if !ok {
		return fmt.Errorf("instance '%s' not found", id)
	}
	now := time.Now()
	inst.DeletedAt = &now
	return nil
}

func withTenantCtx(req *http.Request, orgID, projID string) *http.Request {
	req.Header.Set("X-Organization-Id", orgID)
	req.Header.Set("X-Project-Id", projID)
	return req
}

func TestPhase70N_LiveGatewayToWorkerE2E(t *testing.T) {
	testKeyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	cipher, cErr := crypto.NewAESGCMCipher(testKeyHex, "v1")
	require.NoError(t, cErr)
	crypto.SetGlobalCipher(cipher)

	// Step 2: Ephemeral Local Test PKI Generation
	caPEM, caPriv, caCert := generateTestCA(t)
	serverCertPEM, serverKeyPEM := generateSignedCert(t, caCert, caPriv, "localhost", []string{"localhost", "127.0.0.1"}, false)
	clientCertPEM, clientKeyPEM := generateSignedCert(t, caCert, caPriv, "anarva-gateway-client", nil, true)

	serverTLSCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	require.NoError(t, err)

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caPEM)

	dbDir, _ := os.MkdirTemp("", "worker_e2e_db")
	defer os.RemoveAll(dbDir)
	dbPath := filepath.Join(dbDir, "worker_metadata.json")

	store, err := worker.NewFileMetadataStore(dbPath)
	require.NoError(t, err)

	// Step 3: Start Real Worker Server
	realDockerRuntime := worker.NewDockerContainerRuntime()
	hasRealDocker := realDockerRuntime.HasRuntime()
	t.Logf("[Phase 70N Audit] Real Docker Host Available: %v", hasRealDocker)

	var activeRuntime worker.ContainerRuntime = realDockerRuntime
	if !hasRealDocker {
		t.Log("[Phase 70N Audit] Real Docker daemon unavailable. Using mock runtime adapter for live end-to-end flow.")
		activeRuntime = NewMockRuntime(true)
	}

	workerToken := "sec-token-phase70n-e2e-live"
	wCfg := &worker.Config{
		ListenAddr:        "127.0.0.1:0",
		WorkerToken:       workerToken,
		CACertPEM:         string(caPEM),
		ServerCertPEM:     string(serverCertPEM),
		ServerKeyPEM:      string(serverKeyPEM),
		RequireClientAuth: true,
		DBPath:            dbPath,
	}

	ws, err := worker.NewWorkerServer(wCfg, store, activeRuntime)
	require.NoError(t, err)

	workerServer := httptest.NewUnstartedServer(ws.Handler())
	workerServer.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}
	workerServer.StartTLS()
	defer workerServer.Close()

	// Step 4: Start Real Gateway with RemoteComputeProvider & mTLS
	tlsOpts := &computeProvider.TLSConfigOptions{
		CACertPEM:     string(caPEM),
		ClientCertPEM: string(clientCertPEM),
		ClientKeyPEM:  string(clientKeyPEM),
		ServerName:    "localhost",
	}

	remoteProvider, err := computeProvider.NewRemoteComputeProviderWithTLS(workerServer.URL, workerToken, tlsOpts, nil)
	require.NoError(t, err)

	compRepo := newMemoryComputeRepository()
	mappingRepo := mapping.NewInMemoryMappingRepository()

	compUC := computeUseCase.NewComputeUseCase(compRepo, nil, remoteProvider)
	compUC.SetMappingRepository(mappingRepo)
	compHandler := computeDelivery.NewComputeHandler(compUC, nil)

	gatewayMux := http.NewServeMux()
	compHandler.RegisterRoutes(gatewayMux)

	tenantMiddleware := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgHeader := r.Header.Get("X-Organization-Id")
		projHeader := r.Header.Get("X-Project-Id")
		ctx := r.Context()
		if orgHeader != "" {
			ctx = context.WithValue(ctx, security.OrgIDKey, orgHeader)
		}
		if projHeader != "" {
			ctx = context.WithValue(ctx, security.ProjectIDKey, projHeader)
		}
		gatewayMux.ServeHTTP(w, r.WithContext(ctx))
	})

	gatewayServer := httptest.NewServer(tenantMiddleware)
	defer gatewayServer.Close()

	ctx := context.Background()

	var createdInstanceID string
	var createdWorkloadID string

	orgA := "org-tenant-a"
	projA := "proj-tenant-a"
	orgB := "org-tenant-b"
	projB := "proj-tenant-b"

	// Step 5: Verify mTLS & Service Token Authentication Rejections
	t.Run("Step 5. mTLS & Service Token Security Boundary Verification", func(t *testing.T) {
		noCertOpts := &computeProvider.TLSConfigOptions{
			CACertPEM: string(caPEM),
		}
		noCertProv, _ := computeProvider.NewRemoteComputeProviderWithTLS(workerServer.URL, workerToken, noCertOpts, nil)
		_, cErr := noCertProv.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "no-cert-node"})
		assert.Error(t, cErr, "Worker MUST reject gateway request missing mTLS client cert")

		badTokenProv, _ := computeProvider.NewRemoteComputeProviderWithTLS(workerServer.URL, "invalid-token", tlsOpts, nil)
		_, tErr := badTokenProv.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "bad-token-node"})
		assert.ErrorIs(t, tErr, computeProvider.ErrRemoteWorkerUnauthorized, "Worker MUST reject request with invalid service token")
	})

	// Step 6: Real CREATE Instance via Gateway HTTP API
	t.Run("Step 6. CREATE Compute Instance via Gateway HTTP API", func(t *testing.T) {
		createPayload := map[string]interface{}{
			"name":        "e2e-worker-70n",
			"slug":        "e2e-worker-70n",
			"dockerImage": "busybox:stable",
			"command":     []string{"sh", "-c", "while true; do sleep 1; done"},
			"vcpu":        1.0,
			"memoryMb":    2048,
			"storageGb":   20,
			"acu":         1.0,
			"regionId":    "us-east-1",
			"envVars": map[string]string{
				"ANARVA_TEST_SECRET": "secret_val_70n_e2e_live",
			},
		}

		jsonBytes, _ := json.Marshal(createPayload)
		req, _ := http.NewRequest("POST", gatewayServer.URL+"/api/v1/compute/instances", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withTenantCtx(req, orgA, projA)

		resp, doErr := http.DefaultClient.Do(req)
		require.NoError(t, doErr)
		defer resp.Body.Close()

		respBytes, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "Response body on failure: %s", string(respBytes))

		var inst computeDomain.ComputeInstance
		require.NoError(t, json.Unmarshal(respBytes, &inst))

		createdInstanceID = inst.ID
		createdWorkloadID = inst.ProviderInstanceID

		assert.NotEmpty(t, createdInstanceID)
		assert.Equal(t, createdInstanceID, createdWorkloadID)
		assert.Equal(t, computeDomain.StatusRunning, inst.Status)
		assert.Nil(t, inst.EnvVars, "API response MUST redact EnvVars")
	})

	// Step 7: Real READ Instance via Gateway HTTP API
	t.Run("Step 7. READ Compute Instance via Gateway HTTP API", func(t *testing.T) {
		req, _ := http.NewRequest("GET", gatewayServer.URL+"/api/v1/compute/instances/"+createdInstanceID, nil)
		req = withTenantCtx(req, orgA, projA)

		resp, doErr := http.DefaultClient.Do(req)
		require.NoError(t, doErr)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var inst computeDomain.ComputeInstance
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&inst))

		assert.Equal(t, createdInstanceID, inst.ID)
		assert.Equal(t, computeDomain.StatusRunning, inst.Status)
		assert.Nil(t, inst.EnvVars, "API GET response MUST redact EnvVars")
	})

	// Step 8: Real STOP Instance via Gateway HTTP API
	t.Run("Step 8. STOP Compute Instance via Gateway HTTP API", func(t *testing.T) {
		req, _ := http.NewRequest("POST", gatewayServer.URL+"/api/v1/compute/instances/"+createdInstanceID+"/stop", nil)
		req = withTenantCtx(req, orgA, projA)

		resp, doErr := http.DefaultClient.Do(req)
		require.NoError(t, doErr)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Step 9: Real START Instance via Gateway HTTP API
	t.Run("Step 9. START Compute Instance via Gateway HTTP API", func(t *testing.T) {
		req, _ := http.NewRequest("POST", gatewayServer.URL+"/api/v1/compute/instances/"+createdInstanceID+"/start", nil)
		req = withTenantCtx(req, orgA, projA)

		resp, doErr := http.DefaultClient.Do(req)
		require.NoError(t, doErr)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Step 10: Real RESTART Instance via Gateway HTTP API
	t.Run("Step 10. RESTART Compute Instance via Gateway HTTP API", func(t *testing.T) {
		req, _ := http.NewRequest("POST", gatewayServer.URL+"/api/v1/compute/instances/"+createdInstanceID+"/restart", nil)
		req = withTenantCtx(req, orgA, projA)

		resp, doErr := http.DefaultClient.Do(req)
		require.NoError(t, doErr)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Step 11: Real EXECUTE Command via Gateway HTTP API
	t.Run("Step 11. EXECUTE Command via Gateway HTTP API", func(t *testing.T) {
		execPayload := map[string]interface{}{
			"command": "echo ANARVA_PHASE70N_SUCCESS",
		}
		jsonBytes, _ := json.Marshal(execPayload)
		req, _ := http.NewRequest("POST", gatewayServer.URL+"/api/v1/compute/instances/"+createdInstanceID+"/execute", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withTenantCtx(req, orgA, projA)

		resp, doErr := http.DefaultClient.Do(req)
		require.NoError(t, doErr)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res computeDomain.CommandExecutionResult
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))

		assert.Equal(t, 0, res.ExitCode)
		assert.Contains(t, res.Stdout, "ANARVA_PHASE70N_SUCCESS")
	})

	// Step 12: Real Secret Non-Leakage & Persistence Audit
	t.Run("Step 12. Secret Non-Leakage Verification across Storage & Metadata", func(t *testing.T) {
		if _, statErr := os.Stat(dbPath); statErr == nil {
			metaFileBytes, readErr := os.ReadFile(dbPath)
			require.NoError(t, readErr)
			assert.NotContains(t, string(metaFileBytes), "secret_val_70n_e2e_live", "Secret MUST NOT exist in worker_metadata.json file store")
		}
	})

	// Step 13: Real Tenant Isolation Verification (HTTP 403)
	t.Run("Step 13. Tenant B Unauthorized Access Rejected with HTTP 403", func(t *testing.T) {
		require.NotEmpty(t, createdInstanceID, "createdInstanceID must not be empty for Step 13")
		req, _ := http.NewRequest("GET", gatewayServer.URL+"/api/v1/compute/instances/"+createdInstanceID, nil)
		req = withTenantCtx(req, orgB, projB)

		resp, doErr := http.DefaultClient.Do(req)
		require.NoError(t, doErr)
		defer resp.Body.Close()

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)
		require.Equal(t, http.StatusForbidden, resp.StatusCode, "Tenant B MUST receive 403 Forbidden. Body: %s", buf.String())
	})

	// Step 15: Worker Restart Recovery Simulation
	t.Run("Step 15. Worker Restart Recovery State Reconciliation", func(t *testing.T) {
		recStore, _ := worker.NewFileMetadataStore(dbPath)
		recWs, _ := worker.NewWorkerServer(wCfg, recStore, activeRuntime)
		err := recWs.ReconcileStartupState(ctx)
		require.NoError(t, err)

		m, getErr := recStore.Get(createdWorkloadID)
		require.NoError(t, getErr)
		assert.Equal(t, createdWorkloadID, m.WorkloadID)
		assert.Equal(t, "RUNNING", m.Status)
	})

	// Step 23: Real DELETE Instance via Gateway HTTP API
	t.Run("Step 23. DELETE Compute Instance via Gateway HTTP API", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", gatewayServer.URL+"/api/v1/compute/instances/"+createdInstanceID, nil)
		req = withTenantCtx(req, orgA, projA)

		resp, doErr := http.DefaultClient.Do(req)
		require.NoError(t, doErr)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Confirm deleted in worker store
		_, getErr := store.Get(createdWorkloadID)
		assert.Error(t, getErr, "Deleted workload must not be found in worker store")
	})
}
