package worker

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	ListenAddr     string
	WorkerToken    string
	CACertPEM      string
	ServerCertPEM  string
	ServerKeyPEM   string
	RequireClientAuth bool
	DBPath         string
}

func LoadWorkerConfigFromEnv() (*Config, error) {
	addr := strings.TrimSpace(os.Getenv("WORKER_LISTEN_ADDR"))
	if addr == "" {
		addr = ":8443"
	}

	token := strings.TrimSpace(os.Getenv("COMPUTE_WORKER_TOKEN"))
	ca := strings.TrimSpace(os.Getenv("COMPUTE_WORKER_CA_CERT"))
	cert := strings.TrimSpace(os.Getenv("COMPUTE_WORKER_SERVER_CERT"))
	key := strings.TrimSpace(os.Getenv("COMPUTE_WORKER_SERVER_KEY"))
	dbPath := strings.TrimSpace(os.Getenv("WORKER_DB_PATH"))
	if dbPath == "" {
		dbPath = "worker_data.json"
	}

	requireAuth := true
	if os.Getenv("WORKER_REQUIRE_CLIENT_AUTH") == "false" {
		requireAuth = false
	}

	cfg := &Config{
		ListenAddr:        addr,
		WorkerToken:       token,
		CACertPEM:         ca,
		ServerCertPEM:     cert,
		ServerKeyPEM:      key,
		RequireClientAuth: requireAuth,
		DBPath:            dbPath,
	}

	if os.Getenv("ANARVA_ENV") == "production" {
		if token == "" {
			return nil, errors.New("COMPUTE_WORKER_TOKEN is required in production")
		}
		if ca == "" || cert == "" || key == "" {
			return nil, errors.New("mTLS certificate configuration (CA, Server Cert, Server Key) is required in production")
		}
	}

	return cfg, nil
}
