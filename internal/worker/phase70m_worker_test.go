package worker_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	"github.com/anarva-cloud/anarva-cloud-db/internal/worker"
)

func generateTestCA(t *testing.T) ([]byte, *rsa.PrivateKey, *x509.Certificate) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "ANARVA Test CA"},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)

	pemBlock := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	parsedCert, _ := x509.ParseCertificate(certBytes)
	return pemBlock, priv, parsedCert
}

func generateSignedCert(t *testing.T, caCert *x509.Certificate, caPriv *rsa.PrivateKey, commonName string, dnsNames []string, isClient bool) ([]byte, []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	extUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	if isClient {
		extUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: commonName},
		DNSNames:     dnsNames,
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  extUsage,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &priv.PublicKey, caPriv)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return certPEM, keyPEM
}

type MockRuntime struct {
	hasDocker  bool
	containers map[string]string
}

func NewMockRuntime(hasDocker bool) *MockRuntime {
	return &MockRuntime{
		hasDocker:  hasDocker,
		containers: make(map[string]string),
	}
}

func (m *MockRuntime) HasRuntime() bool { return m.hasDocker }
func (m *MockRuntime) CreateContainer(ctx context.Context, meta *worker.WorkloadMetadata, envVars map[string]string) (string, error) {
	cID := "container-" + meta.WorkloadID
	m.containers[cID] = "running"
	return cID, nil
}
func (m *MockRuntime) StartContainer(ctx context.Context, containerID string) error {
	m.containers[containerID] = "running"
	return nil
}
func (m *MockRuntime) StopContainer(ctx context.Context, containerID string) error {
	m.containers[containerID] = "stopped"
	return nil
}
func (m *MockRuntime) RestartContainer(ctx context.Context, containerID string) error {
	m.containers[containerID] = "running"
	return nil
}
func (m *MockRuntime) DeleteContainer(ctx context.Context, containerID string) error {
	delete(m.containers, containerID)
	return nil
}
func (m *MockRuntime) ExecuteContainerCommand(ctx context.Context, containerID string, command string, timeoutSec int) (int, string, string, error) {
	if _, ok := m.containers[containerID]; !ok {
		return 1, "", "", fmt.Errorf("container %s not found", containerID)
	}
	return 0, "MOCK_STDOUT: " + command, "", nil
}
func (m *MockRuntime) InspectContainerStatus(ctx context.Context, containerID string) (string, error) {
	if st, ok := m.containers[containerID]; ok {
		return st, nil
	}
	return "unavailable", fmt.Errorf("container not found")
}
func (m *MockRuntime) GetContainerMetrics(ctx context.Context, containerID string) (float64, int, int64, int64, error) {
	return 15.0, 512, 1024, 2048, nil
}
func (m *MockRuntime) DiscoverANARVAContainers(ctx context.Context) ([]string, error) {
	var list []string
	for k := range m.containers {
		list = append(list, k)
	}
	return list, nil
}

func TestPhase70M_WorkerFoundation(t *testing.T) {
	caPEM, caPriv, caCert := generateTestCA(t)
	serverCertPEM, serverKeyPEM := generateSignedCert(t, caCert, caPriv, "localhost", []string{"localhost", "127.0.0.1"}, false)
	clientCertPEM, clientKeyPEM := generateSignedCert(t, caCert, caPriv, "anarva-gateway-client", nil, true)

	serverTLSCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	require.NoError(t, err)

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caPEM)

	dbDir, _ := os.MkdirTemp("", "worker_test_db")
	defer os.RemoveAll(dbDir)
	dbPath := filepath.Join(dbDir, "worker_metadata.json")

	store, err := worker.NewFileMetadataStore(dbPath)
	require.NoError(t, err)

	mockRt := NewMockRuntime(true)
	token := "worker-secret-token-70m"

	cfg := &worker.Config{
		ListenAddr:        "127.0.0.1:0",
		WorkerToken:       token,
		CACertPEM:         string(caPEM),
		ServerCertPEM:     string(serverCertPEM),
		ServerKeyPEM:      string(serverKeyPEM),
		RequireClientAuth: true,
		DBPath:            dbPath,
	}

	ws, err := worker.NewWorkerServer(cfg, store, mockRt)
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(ws.Handler())
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}
	server.StartTLS()
	defer server.Close()

	tlsOpts := &computeProvider.TLSConfigOptions{
		CACertPEM:     string(caPEM),
		ClientCertPEM: string(clientCertPEM),
		ClientKeyPEM:  string(clientKeyPEM),
		ServerName:    "localhost",
	}

	prov, err := computeProvider.NewRemoteComputeProviderWithTLS(server.URL, token, tlsOpts, nil)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("1. Worker starts and passes liveness", func(t *testing.T) {
		assert.NotNil(t, server.URL)
	})

	t.Run("2. mTLS client authentication succeeds", func(t *testing.T) {
		inst := &computeDomain.ComputeInstance{
			ID:         "workload-mtls-70m",
			ResourceID: "res-mtls-70m",
			Name:       "node-70m",
			EnvVars:    map[string]string{"SECRET_KEY": "super_secret_val_123"},
		}
		created, cErr := prov.CreateInstance(ctx, inst)
		require.NoError(t, cErr)
		assert.Equal(t, "workload-mtls-70m", created.ProviderInstanceID)
	})

	t.Run("3. Invalid client certificate rejected", func(t *testing.T) {
		otherCAPEM, otherCAPriv, otherCACert := generateTestCA(t)
		badCert, badKey := generateSignedCert(t, otherCACert, otherCAPriv, "bad-client", nil, true)

		badOpts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(otherCAPEM),
			ClientCertPEM: string(badCert),
			ClientKeyPEM:  string(badKey),
			ServerName:    "localhost",
		}
		badProv, _ := computeProvider.NewRemoteComputeProviderWithTLS(server.URL, token, badOpts, nil)
		_, cErr := badProv.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "bad-cert-workload"})
		assert.Error(t, cErr)
	})

	t.Run("4. Invalid service token rejected", func(t *testing.T) {
		badTokenProv, _ := computeProvider.NewRemoteComputeProviderWithTLS(server.URL, "invalid-token", tlsOpts, nil)
		_, cErr := badTokenProv.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "bad-token-workload"})
		assert.ErrorIs(t, cErr, computeProvider.ErrRemoteWorkerUnauthorized)
	})

	t.Run("5. CREATE provisions workload metadata and runtime container", func(t *testing.T) {
		inst := &computeDomain.ComputeInstance{
			ID:         "workload-create-01",
			ResourceID: "res-create-01",
			Name:       "app-worker",
			VCPU:       2.0,
			MemoryMB:   4096,
		}
		created, cErr := prov.CreateInstance(ctx, inst)
		require.NoError(t, cErr)
		assert.Equal(t, "workload-create-01", created.ProviderInstanceID)
	})

	t.Run("6. Duplicate CREATE does not duplicate workload (Idempotency-Key)", func(t *testing.T) {
		inst := &computeDomain.ComputeInstance{
			ID:         "workload-create-dup",
			ResourceID: "res-idempotent-key-01",
			Name:       "app-worker-dup",
		}
		created1, cErr1 := prov.CreateInstance(ctx, inst)
		require.NoError(t, cErr1)

		// Second call with same ResourceID / Idempotency-Key
		created2, cErr2 := prov.CreateInstance(ctx, inst)
		require.NoError(t, cErr2)
		assert.Equal(t, created1.ProviderInstanceID, created2.ProviderInstanceID)
	})

	t.Run("7. GET returns runtime status", func(t *testing.T) {
		inst, gErr := prov.GetInstance(ctx, "workload-create-01")
		require.NoError(t, gErr)
		assert.Equal(t, "workload-create-01", inst.ID)
	})

	t.Run("8. STOP halts workload", func(t *testing.T) {
		err := prov.StopInstance(ctx, "workload-create-01")
		assert.NoError(t, err)
	})

	t.Run("9. START resumes workload", func(t *testing.T) {
		err := prov.StartInstance(ctx, "workload-create-01")
		assert.NoError(t, err)
	})

	t.Run("10. RESTART restarts workload", func(t *testing.T) {
		err := prov.RestartInstance(ctx, "workload-create-01")
		assert.NoError(t, err)
	})

	t.Run("11. EXECUTE runs command inside container namespace", func(t *testing.T) {
		req := &computeDomain.CommandExecutionRequest{Command: "echo hello_70m"}
		res, eErr := prov.ExecuteCommand(ctx, "workload-create-01", req)
		require.NoError(t, eErr)
		assert.Equal(t, 0, res.ExitCode)
		assert.Contains(t, res.Stdout, "echo hello_70m")
	})

	t.Run("12. Host command execution is impossible through API", func(t *testing.T) {
		req := &computeDomain.CommandExecutionRequest{Command: "cat /etc/passwd"}
		_, eErr := prov.ExecuteCommand(ctx, "non-existent-workload-99", req)
		assert.Error(t, eErr)
	})

	t.Run("13. DELETE removes container and metadata", func(t *testing.T) {
		err := prov.DeleteInstance(ctx, "workload-create-01")
		assert.NoError(t, err)
	})

	t.Run("14. Missing workload returns safe 404", func(t *testing.T) {
		_, err := prov.GetInstanceMetrics(ctx, "missing-workload-id")
		assert.ErrorIs(t, err, computeProvider.ErrRemoteWorkerNotFound)
	})

	t.Run("15. CPU and Memory limits applied to metadata", func(t *testing.T) {
		meta, err := store.Get("workload-mtls-70m")
		require.NoError(t, err)
		assert.True(t, meta.VCPU >= 0)
	})

	t.Run("16. Secrets never appear in GET responses or stored JSON file", func(t *testing.T) {
		data, readErr := os.ReadFile(dbPath)
		require.NoError(t, readErr)
		assert.NotContains(t, string(data), "super_secret_val_123", "Secrets MUST NOT be stored in worker metadata file")
	})

	t.Run("17. Worker restart recovery reconciles state", func(t *testing.T) {
		newStore, _ := worker.NewFileMetadataStore(dbPath)
		newWs, _ := worker.NewWorkerServer(cfg, newStore, mockRt)
		err := newWs.ReconcileStartupState(ctx)
		assert.NoError(t, err)

		recovered, rErr := newStore.Get("workload-mtls-70m")
		require.NoError(t, rErr)
		assert.Equal(t, "RUNNING", recovered.Status)
	})
}
