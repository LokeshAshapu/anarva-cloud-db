package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type correlationContextKey string

const (
	RequestIDKey      correlationContextKey = "requestID"
	CorrelationIDKey  correlationContextKey = "correlationID"
	IdempotencyKeyKey correlationContextKey = "idempotencyKey"
)

// CorrelationMiddleware injects requestId, correlationId, and Idempotency-Key into context
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = fmt.Sprintf("req-%d", time.Now().UnixNano()/1e5)
		}

		corrID := r.Header.Get("X-Correlation-ID")
		if corrID == "" {
			corrID = reqID
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")

		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		ctx = context.WithValue(ctx, CorrelationIDKey, corrID)
		if idempotencyKey != "" {
			ctx = context.WithValue(ctx, IdempotencyKeyKey, idempotencyKey)
		}

		w.Header().Set("X-Request-ID", reqID)
		w.Header().Set("X-Correlation-ID", corrID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
