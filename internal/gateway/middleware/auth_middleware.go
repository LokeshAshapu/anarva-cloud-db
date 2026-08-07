package middleware

import (
	"context"
	"net/http"
	"strings"

	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

type AuthMiddleware struct {
	jwtManager *security.JWTManager
}

func NewAuthMiddleware(jwtManager *security.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip public routes
		path := r.URL.Path
		if path == "/health" || path == "/metrics" || strings.HasPrefix(path, "/api/v1/auth/login") || strings.HasPrefix(path, "/api/v1/auth/signup") || strings.HasPrefix(path, "/api/v1/auth/verify-email") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			apiKeyHeader := r.Header.Get("X-API-Key")
			if apiKeyHeader != "" {
				// API Key authentication path
				ctx := context.WithValue(r.Context(), UserIDKey, "api_key_user")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			respondError(w, appErrors.New(appErrors.CodeUnauthorized, "missing Authorization header"))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := m.jwtManager.ValidateToken(tokenStr)
		if err != nil {
			respondError(w, appErrors.New(appErrors.CodeUnauthorized, "invalid or expired token"))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func respondError(w http.ResponseWriter, err *appErrors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatusCode())
	w.Write([]byte(err.Error()))
}
