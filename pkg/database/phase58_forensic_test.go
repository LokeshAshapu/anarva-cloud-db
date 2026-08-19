package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	pkgDatabase "github.com/anarva-cloud/anarva-cloud-db/pkg/database"
)

func TestPhase58_ForensicDiagnostics(t *testing.T) {
	t.Run("1. Parse postgres:// and postgresql:// URLs with sslmode=require", func(t *testing.T) {
		cfg1 := config.DatabaseConfig{
			URL: "postgres://user:secret_pass_123@dpg-cxxxx.render.com:5432/anarva_prod_db?sslmode=require",
		}
		diag1 := pkgDatabase.PerformForensicDiagnostics(cfg1)
		assert.Equal(t, "postgres", diag1.Scheme)
		assert.Equal(t, "dpg-cxxxx.render.com", diag1.Hostname)
		assert.Equal(t, 5432, diag1.Port)
		assert.Equal(t, "anarva_prod_db", diag1.Database)
		assert.Equal(t, "require", diag1.SSLMode)
		assert.NotContains(t, fmt.Sprintf("%+v", diag1), "secret_pass_123")

		cfg2 := config.DatabaseConfig{
			URL: "postgresql://anarva_user:super_secret@dpg-cyyyy.render.com:5432/anarva_db?sslmode=require",
		}
		diag2 := pkgDatabase.PerformForensicDiagnostics(cfg2)
		assert.Equal(t, "postgresql", diag2.Scheme)
		assert.Equal(t, "dpg-cyyyy.render.com", diag2.Hostname)
		assert.Equal(t, "anarva_db", diag2.Database)
		assert.Equal(t, "require", diag2.SSLMode)
		assert.NotContains(t, fmt.Sprintf("%+v", diag2), "super_secret")
	})

	t.Run("2. Detect DNS Failure for Non-Existent Domain", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			URL: "postgres://user:pass@nonexistent-domain-name-anarva-xyz-12345.render.com:5432/anarva_db",
		}
		diag := pkgDatabase.PerformForensicDiagnostics(cfg)
		assert.Equal(t, "FAIL", diag.DNSResolution)
		assert.Equal(t, "DATABASE_DNS_FAILURE", diag.ConnectionErrorClass)
	})

	t.Run("3. Detect Connection Refused for Unreachable Port", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:    "127.0.0.1",
			Port:    59999, // Unused port
			DBName:  "test_db",
			SSLMode: "disable",
		}
		diag := pkgDatabase.PerformForensicDiagnostics(cfg)
		assert.Equal(t, "PASS", diag.DNSResolution)
		assert.Equal(t, "FAIL", diag.TCPConnection)
		assert.Contains(t, []string{"DATABASE_CONNECTION_REFUSED", "DATABASE_CONNECTION_TIMEOUT"}, diag.ConnectionErrorClass)
	})

	t.Run("4. Live PostgreSQL Identity Verification (if available)", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "anarva_admin",
			Password:        "anarva_secret_pass",
			DBName:          "anarva_db",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		}
		dbPool, err := pkgDatabase.NewPostgresDB(cfg)
		if err != nil {
			t.Skipf("Skipping live database identity test (no PostgreSQL listening locally: %v)", err)
			return
		}
		defer dbPool.Close()

		ctx := context.Background()
		identity, errId := dbPool.GetDatabaseIdentity(ctx)
		require.NoError(t, errId)
		assert.True(t, identity.ServerReachable)
		assert.NotEmpty(t, identity.Database)
		assert.NotEmpty(t, identity.User)
	})
}
