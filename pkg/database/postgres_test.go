package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
)

func TestNewPostgresDB_InvalidConfig(t *testing.T) {
	invalidCfg := config.DatabaseConfig{
		Host:     "invalid_host_address_9999",
		Port:     5432,
		User:     "invalid_user",
		Password: "invalid_password",
		DBName:   "invalid_db",
		SSLMode:  "disable",
	}

	_, err := NewPostgresDB(invalidCfg)
	assert.Error(t, err)
}
