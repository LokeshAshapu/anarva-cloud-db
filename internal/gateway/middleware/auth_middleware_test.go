package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

func TestAuthMiddleware(t *testing.T) {
	jwtSecret := "super-secret-production-key-32-bytes!"
	jwtManager := security.NewJWTManager(jwtSecret, "anarva-cloud-db", 1*time.Hour, 24*time.Hour)
	authMiddleware := NewAuthMiddleware(jwtManager)

	// Sample protected test handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/health" || path == "/metrics" || path == "/api/v1/security/status" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"PUBLIC"}`))
			return
		}

		userID, ok := r.Context().Value(UserIDKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"SUCCESS","user_id":"` + userID + `"}`))
	})

	handler := authMiddleware.Authenticate(protectedHandler)

	t.Run("P0 Regression Test: Missing Authorization Header on Protected Route Returns 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/organizations/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "AUTH_REQUIRED")
		assert.NotContains(t, rr.Body.String(), "usr-default")
		assert.NotContains(t, rr.Body.String(), "ADMIN")
	})

	t.Run("P0 Regression Test: Invalid JWT Token Returns 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/projects/", nil)
		req.Header.Set("Authorization", "Bearer invalid-garbage-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "INVALID_TOKEN")
		assert.NotContains(t, rr.Body.String(), "usr-default")
		assert.NotContains(t, rr.Body.String(), "ADMIN")
	})

	t.Run("P0 Regression Test: Malformed Bearer Header Returns 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/databases", nil)
		req.Header.Set("Authorization", "Bearer ")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "AUTH_REQUIRED")
	})

	t.Run("P0 Regression Test: Expired JWT Token Returns 401", func(t *testing.T) {
		// Generate an expired JWT manager
		expiredManager := security.NewJWTManager(jwtSecret, "anarva-cloud-db", -1*time.Hour, -1*time.Hour)
		expiredToken, _, err := expiredManager.GenerateTokenPair("usr-101", "test@anarva.io", "MEMBER", "org-101")
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/v1/databases", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "INVALID_TOKEN")
	})

	t.Run("Valid JWT Token Returns 200 with Identity", func(t *testing.T) {
		validToken, _, err := jwtManager.GenerateTokenPair("usr-prod-999", "admin@anarva.io", "ADMIN", "org-main")
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/v1/databases", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "usr-prod-999")
	})

	t.Run("Public Endpoints Allow Unauthenticated Access", func(t *testing.T) {
		publicPaths := []string{"/health", "/metrics", "/api/v1/security/status"}
		for _, path := range publicPaths {
			req := httptest.NewRequest("GET", path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code, "Expected 200 for public route: "+path)
		}
	})

	t.Run("Dev Auth Fails Closed in Production Environment", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Unsetenv("ENABLE_DEV_AUTH")
		defer os.Unsetenv("APP_ENV")

		req := httptest.NewRequest("GET", "/api/v1/organizations/", nil)
		req.Header.Set("Authorization", "Bearer dev-token-bypass")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "INVALID_TOKEN")
	})
}
