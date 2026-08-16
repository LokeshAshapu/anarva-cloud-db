package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := LoadConfig("")
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 9090, cfg.Server.GRPCPort)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "anarva_admin", cfg.Database.User)
	assert.Equal(t, "anarva_password", cfg.Database.Password)
	assert.Equal(t, "anarva_cloud_db", cfg.Database.DBName)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
}

func TestDatabaseConfig_DSN(t *testing.T) {
	dbCfg := DatabaseConfig{
		Host:     "db.example.com",
		Port:     5432,
		User:     "user",
		Password: "secretpassword",
		DBName:   "mydb",
		SSLMode:  "require",
	}

	expectedDSN := "host=db.example.com port=5432 user=user password=secretpassword dbname=mydb sslmode=require"
	assert.Equal(t, expectedDSN, dbCfg.DSN())
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("JWT_SECRET", "super_secure_custom_prod_jwt_key_12345")
	os.Setenv("DATABASE_HOST", "prod-db.internal")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DATABASE_HOST")
	}()

	cfg, err := LoadConfig("")
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, 9000, cfg.Server.Port)
}
