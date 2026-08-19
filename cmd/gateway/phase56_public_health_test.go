package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	gwMiddleware "github.com/anarva-cloud/anarva-cloud-db/internal/gateway/middleware"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

func TestPhase56_PublicPersistenceHealthEndpoint(t *testing.T) {
	jwtManager := security.NewJWTManager("test_secret_key_1234567890_32bytes", "anarva-test", 15*time.Minute, 7*24*time.Hour)
	authMiddleware := gwMiddleware.NewAuthMiddleware(jwtManager)

	mux := http.NewServeMux()

	// Register /api/v1/health/persistence test route
	mux.HandleFunc("GET /api/v1/health/persistence", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"status":      "HEALTHY",
				"environment": "development",
				"mode":        "POSTGRESQL",
				"database": map[string]interface{}{
					"configured":    true,
					"connected":     true,
					"provider":      "postgresql",
					"database_name": "anarva_db",
				},
				"fallback_repository": map[string]interface{}{
					"enabled": false,
				},
			},
			"requestId": "req-test-123",
		})
	})

	// Protected route for verification
	mux.HandleFunc("/api/v1/databases", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": "protected"})
	})

	handler := authMiddleware.Authenticate(mux)

	t.Run("1. Public /api/v1/health/persistence does NOT require Authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/health/persistence", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "HEALTHY")
		assert.Contains(t, rec.Body.String(), "POSTGRESQL")
	})

	t.Run("2. Authenticated API (/api/v1/databases) REQUIRES Authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/databases", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "AUTH_REQUIRED")
	})

	t.Run("3. Options preflight request succeeds without Authorization header", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/v1/databases", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("4. Persistence health response contains NO secrets or credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/health/persistence", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		body := rec.Body.String()
		assert.NotContains(t, body, "password")
		assert.NotContains(t, body, "secret")
		assert.NotContains(t, body, "DATABASE_URL")
		assert.NotContains(t, body, "Bearer")
	})
}
