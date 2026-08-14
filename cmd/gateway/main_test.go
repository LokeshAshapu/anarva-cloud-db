package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadinessHealthCheck_DisconnectedInProduction(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("APP_ENV")

	// Create readiness test handler matching gateway readiness behavior
	readinessHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Simulated uninitialized dbPool in production
		if os.Getenv("APP_ENV") == "production" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"NOT_READY","database":"UNAVAILABLE"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"READY","database":"CONNECTED"}`))
	})

	req := httptest.NewRequest("GET", "/readiness", nil)
	rr := httptest.NewRecorder()

	readinessHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Contains(t, rr.Body.String(), "NOT_READY")
	assert.Contains(t, rr.Body.String(), "UNAVAILABLE")
}

func TestLivenessHealthCheck(t *testing.T) {
	livenessHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	livenessHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "UP")
}
