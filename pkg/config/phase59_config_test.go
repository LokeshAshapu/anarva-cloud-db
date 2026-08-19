package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	pkgDatabase "github.com/anarva-cloud/anarva-cloud-db/pkg/database"
)

func TestPhase59_DatabaseURLPriorityAndFailClosed(t *testing.T) {
	t.Run("TEST 1 & TEST 5 & TEST 6: postgres:// and postgresql:// URL parsing with sslmode=require", func(t *testing.T) {
		dbCfg := config.DatabaseConfig{
			URL: "postgres://produser:secret@production-host:5432/proddb?sslmode=require",
		}
		diag := pkgDatabase.PerformForensicDiagnostics(dbCfg)
		assert.Equal(t, "production-host", diag.Hostname)
		assert.Equal(t, 5432, diag.Port)
		assert.Equal(t, "proddb", diag.Database)
		assert.Equal(t, "require", diag.SSLMode)

		dbCfg2 := config.DatabaseConfig{
			URL: "postgresql://produser:secret@production-host:5432/proddb?sslmode=require",
		}
		diag2 := pkgDatabase.PerformForensicDiagnostics(dbCfg2)
		assert.Equal(t, "production-host", diag2.Hostname)
		assert.Equal(t, "require", diag2.SSLMode)
	})

	t.Run("TEST 2: DATABASE_URL WINS over localhost development defaults", func(t *testing.T) {
		dbCfg := config.DatabaseConfig{
			URL:      "postgres://produser:secret@real-prod-db.render.com:5432/anarva_prod?sslmode=require",
			Host:     "localhost",
			Port:     5432,
			User:     "anarva_admin",
			Password: "anarva_password",
			DBName:   "anarva_cloud_db",
			SSLMode:  "disable",
		}
		assert.Equal(t, "postgres://produser:secret@real-prod-db.render.com:5432/anarva_prod?sslmode=require", dbCfg.DSN())
	})

	t.Run("TEST 3: Missing DATABASE_URL in Production returns empty DSN (NEVER localhost)", func(t *testing.T) {
		os.Setenv("ANARVA_ENV", "production")
		defer os.Unsetenv("ANARVA_ENV")
		os.Unsetenv("DATABASE_URL")

		dbCfg := config.DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			DBName:  "anarva_cloud_db",
			SSLMode: "disable",
		}
		assert.Empty(t, dbCfg.DSN())
	})

	t.Run("TEST 4: Missing DATABASE_URL in Development allows local development configuration", func(t *testing.T) {
		os.Setenv("ANARVA_ENV", "development")
		defer os.Unsetenv("ANARVA_ENV")
		os.Unsetenv("DATABASE_URL")

		dbCfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "dev_user",
			Password: "dev_password",
			DBName:   "dev_db",
			SSLMode:  "disable",
		}
		assert.Contains(t, dbCfg.DSN(), "host=localhost")
		assert.Contains(t, dbCfg.DSN(), "dbname=dev_db")
	})
}
