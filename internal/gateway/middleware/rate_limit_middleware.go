package middleware

import (
	"net/http"
	"sync"
	"time"

	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type rateLimiter struct {
	tokens     int
	maxTokens  int
	lastRefill time.Time
	mu         sync.Mutex
}

type RateLimitMiddleware struct {
	limiters map[string]*rateLimiter
	mu       sync.RWMutex
	rate     int // tokens per minute
}

func NewRateLimitMiddleware(ratePerMinute int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiters: make(map[string]*rateLimiter),
		rate:     ratePerMinute,
	}
}

func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}

		m.mu.Lock()
		limiter, ok := m.limiters[ip]
		if !ok {
			limiter = &rateLimiter{
				tokens:     m.rate,
				maxTokens:  m.rate,
				lastRefill: time.Now(),
			}
			m.limiters[ip] = limiter
		}
		m.mu.Unlock()

		limiter.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(limiter.lastRefill)
		if elapsed >= time.Minute {
			limiter.tokens = limiter.maxTokens
			limiter.lastRefill = now
		}

		if limiter.tokens <= 0 {
			limiter.mu.Unlock()
			respondError(w, appErrors.New(appErrors.CodeTimeout, "rate limit exceeded. Try again in 1 minute."))
			return
		}

		limiter.tokens--
		limiter.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
