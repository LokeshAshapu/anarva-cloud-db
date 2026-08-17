package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
	OrgIDKey  contextKey = "org_id"
)

type AuthMiddleware struct {
	jwtManager *security.JWTManager
}

func NewAuthMiddleware(jwtManager *security.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Allow CORS Preflight OPTIONS Requests
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Explicit Public Routes
		if path == "/health" || path == "/readiness" || path == "/metrics" || path == "/api/v1/security/status" ||
			strings.HasPrefix(path, "/api/v1/auth/login") ||
			strings.HasPrefix(path, "/api/v1/auth/signup") ||
			strings.HasPrefix(path, "/api/v1/auth/verify-email") {
			next.ServeHTTP(w, r)
			return
		}

		// Check Authorization Header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// Check X-API-Key
			apiKeyHeader := r.Header.Get("X-API-Key")
			if apiKeyHeader != "" {
				ctx := context.WithValue(r.Context(), UserIDKey, "api_key_user")
				ctx = context.WithValue(ctx, RoleKey, "ADMIN")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Dev Auth Guard: Explicitly requires APP_ENV=development AND ENABLE_DEV_AUTH=true
			if os.Getenv("APP_ENV") == "development" && os.Getenv("ENABLE_DEV_AUTH") == "true" {
				ctx := context.WithValue(r.Context(), UserIDKey, "usr-dev")
				ctx = context.WithValue(ctx, RoleKey, "ADMIN")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			respondAuthError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "missing Authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		if tokenStr == "" {
			respondAuthError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "empty Bearer token")
			return
		}

		// Dev Auth Guard for token payloads: Explicitly requires APP_ENV=development AND ENABLE_DEV_AUTH=true
		if os.Getenv("APP_ENV") == "development" && os.Getenv("ENABLE_DEV_AUTH") == "true" {
			if strings.HasPrefix(tokenStr, "dev-token-") {
				ctx := context.WithValue(r.Context(), UserIDKey, "usr-dev")
				ctx = context.WithValue(ctx, RoleKey, "ADMIN")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Strict JWT Parse & Signature Verification
		claims, err := m.jwtManager.ValidateToken(tokenStr)
		if err != nil {
			respondAuthError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired token")
			return
		}

		// Inject verified claims into request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		if claims.OrgID != "" {
			ctx = context.WithValue(ctx, OrgIDKey, claims.OrgID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func respondAuthError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "unauthorized",
		"code":    code,
		"message": message,
	})
}

func respondError(w http.ResponseWriter, err *appErrors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatusCode())
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}
