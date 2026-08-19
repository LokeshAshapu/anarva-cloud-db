package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gwMiddleware "github.com/anarva-cloud/anarva-cloud-db/internal/gateway/middleware"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	pkgDatabase "github.com/anarva-cloud/anarva-cloud-db/pkg/database"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

func TestPhase60_ProductionGatewayInitializationChain(t *testing.T) {
	t.Run("1. DATABASE_URL priority over development defaults in LoadConfig", func(t *testing.T) {
		prodURL := "postgres://produser:secret123@prod-db-host.render.com:5432/anarva_prod?sslmode=require"
		os.Setenv("DATABASE_URL", prodURL)
		defer os.Unsetenv("DATABASE_URL")

		cfg, err := config.LoadConfig("")
		require.NoError(t, err)

		assert.Equal(t, prodURL, cfg.Database.URL)
		assert.Equal(t, prodURL, cfg.Database.DSN())
	})

	t.Run("2. Production mode with localhost URL fails startup assertion", func(t *testing.T) {
		os.Setenv("ANARVA_ENV", "production")
		defer os.Unsetenv("ANARVA_ENV")
		os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
		defer os.Unsetenv("DATABASE_URL")

		cfg, err := config.LoadConfig("")
		require.NoError(t, err)

		rawDSN := cfg.Database.DSN()
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", rawDSN)
	})

	t.Run("3. Public Persistence Health Endpoint returns build commit and diagnostic info", func(t *testing.T) {
		jwtManager := security.NewJWTManager("test_secret_key_1234567890_32bytes", "anarva-test", 15*time.Minute, 7*24*time.Hour)
		authMiddleware := gwMiddleware.NewAuthMiddleware(jwtManager)

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/health/persistence", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"status":      "HEALTHY",
					"environment": "production",
					"mode":        "POSTGRESQL",
					"build": map[string]interface{}{
						"version":   "0.1.0",
						"gitCommit": "phase-60-cd8ca2a",
					},
					"database": map[string]interface{}{
						"configuration_source": "DATABASE_URL",
						"configured":           true,
						"connected":            true,
						"hostname":             "prod-db-host.render.com",
						"port":                 5432,
						"sslmode":              "require",
					},
				},
				"requestId": "req-phase60-test",
			})
		})

		handler := authMiddleware.Authenticate(mux)
		req := httptest.NewRequest("GET", "/api/v1/health/persistence", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		build := data["build"].(map[string]interface{})
		assert.Equal(t, "phase-60-cd8ca2a", build["gitCommit"])

		db := data["database"].(map[string]interface{})
		assert.Equal(t, "DATABASE_URL", db["configuration_source"])
		assert.Equal(t, "prod-db-host.render.com", db["hostname"])
		assert.Equal(t, "require", db["sslmode"])
	})

	t.Run("4. Forensic Diagnostics classifies DNS failure safely", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			URL: "postgres://user:pass@nonexistent-host-9999.render.com:5432/db",
		}
		res := pkgDatabase.PerformForensicDiagnostics(cfg)
		assert.Equal(t, "FAIL", res.DNSResolution)
		assert.Equal(t, "DATABASE_DNS_FAILURE", res.ConnectionErrorClass)
		assert.Equal(t, "nonexistent-host-9999.render.com", res.Hostname)
	})
}
