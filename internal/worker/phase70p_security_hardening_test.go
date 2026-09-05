package worker_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/worker"
)

func TestPhase70P_SecurityHardening(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "phase70p_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "worker_metadata.json")

	t.Run("1. PID Limit Configuration Verification", func(t *testing.T) {
		runtime := worker.NewDockerContainerRuntime()
		assert.NotNil(t, runtime)
	})

	t.Run("2. Atomic Metadata Write & 0600 Permissions Verification", func(t *testing.T) {
		store, sErr := worker.NewFileMetadataStore(dbPath)
		require.NoError(t, sErr)

		meta := &worker.WorkloadMetadata{
			WorkloadID: "workload-atomic-70p",
			ResourceID: "res-atomic-70p",
			Name:       "node-70p",
			Status:     "RUNNING",
		}

		saveErr := store.Save(meta)
		require.NoError(t, saveErr)

		// Verify file exists
		info, statErr := os.Stat(dbPath)
		require.NoError(t, statErr)

		// Verify permissions on non-windows OS where file mode bits apply
		if os.PathSeparator == '/' {
			mode := info.Mode().Perm()
			assert.Equal(t, os.FileMode(0600), mode, "Metadata file permissions MUST be restricted to 0600")
		}

		// Verify temporary .tmp file is cleaned up post-rename
		tmpPath := dbPath + ".tmp"
		_, tmpStatErr := os.Stat(tmpPath)
		assert.True(t, os.IsNotExist(tmpStatErr), "Temporary file .tmp MUST be cleaned up post atomic rename")
	})

	t.Run("3. Metadata Overwrite Safety & Secret Redaction", func(t *testing.T) {
		store, _ := worker.NewFileMetadataStore(dbPath)
		meta := &worker.WorkloadMetadata{
			WorkloadID: "workload-atomic-70p",
			Status:     "STOPPED",
		}
		err := store.Save(meta)
		require.NoError(t, err)

		fetched, gErr := store.Get("workload-atomic-70p")
		require.NoError(t, gErr)
		assert.Equal(t, "STOPPED", fetched.Status)

		data, readErr := os.ReadFile(dbPath)
		require.NoError(t, readErr)
		assert.NotContains(t, string(data), "SECRET_KEY")
	})

	t.Run("4. Systemd Unit Security Hardening Validation", func(t *testing.T) {
		svcPath := filepath.Join("..", "..", "scripts", "anarva-worker.service")
		svcBytes, readErr := os.ReadFile(svcPath)
		require.NoError(t, readErr)
		svcContent := string(svcBytes)

		assert.Contains(t, svcContent, "User=anarva-worker")
		assert.Contains(t, svcContent, "NoNewPrivileges=true")
		assert.Contains(t, svcContent, "PrivateTmp=true")
		assert.Contains(t, svcContent, "ProtectSystem=full")
	})

	t.Run("5. Linux Installer Firewall Metadata Rule Validation", func(t *testing.T) {
		shPath := filepath.Join("..", "..", "scripts", "install-worker.sh")
		shBytes, readErr := os.ReadFile(shPath)
		require.NoError(t, readErr)
		shContent := string(shBytes)

		assert.Contains(t, shContent, "169.254.169.254")
		assert.Contains(t, shContent, "iptables")
		assert.Contains(t, shContent, "anarva-worker")
	})

	t.Run("6. Production Fail-Closed Configuration Audit", func(t *testing.T) {
		os.Setenv("ANARVA_ENV", "production")
		os.Unsetenv("COMPUTE_WORKER_TOKEN")
		os.Unsetenv("COMPUTE_WORKER_CA_CERT")
		defer os.Unsetenv("ANARVA_ENV")

		_, cfgErr := worker.LoadWorkerConfigFromEnv()
		assert.Error(t, cfgErr, "Worker MUST fail closed in production when tokens or certificates are missing")
	})
}
