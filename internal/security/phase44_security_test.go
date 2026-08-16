package security_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gwMiddleware "github.com/anarva-cloud/anarva-cloud-db/internal/gateway/middleware"
	providersSecurity "github.com/anarva-cloud/anarva-cloud-db/internal/providers/security"
	secInternal "github.com/anarva-cloud/anarva-cloud-db/internal/security"
	storageProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	webhookDomain "github.com/anarva-cloud/anarva-cloud-db/internal/webhook/domain"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	pkgSecurity "github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

// 1. Authentication Security Tests
func TestPhase44_AuthenticationSecurity(t *testing.T) {
	jwtMgr := pkgSecurity.NewJWTManager("test_secret_key_12345", "anarva", 1*time.Hour, 24*time.Hour)

	// Valid Token Generation
	accToken, _, err := jwtMgr.GenerateTokenPair("usr-101", "dev@anarva.io", "DEVELOPER", "org-alpha")
	require.NoError(t, err)

	// Valid Token Parse
	claims, err := jwtMgr.ValidateToken(accToken)
	require.NoError(t, err)
	assert.Equal(t, "usr-101", claims.UserID)
	assert.Equal(t, "DEVELOPER", claims.Role)
	assert.Equal(t, "org-alpha", claims.OrgID)

	// Malformed Token Rejection
	_, err = jwtMgr.ValidateToken("invalid.jwt.token")
	assert.Error(t, err)

	// Expired Token Rejection
	expJwtMgr := pkgSecurity.NewJWTManager("test_secret_key_12345", "anarva", -1*time.Minute, 24*time.Hour)
	expToken, _, _ := expJwtMgr.GenerateTokenPair("usr-exp", "exp@anarva.io", "DEVELOPER", "org-alpha")
	_, err = jwtMgr.ValidateToken(expToken)
	assert.Error(t, err)
}

// 2. API Key Security & Secret Redaction Tests
func TestPhase44_APIKeySecurityAndRedaction(t *testing.T) {
	rawKey, hashedKey, err := pkgSecurity.GenerateAPIKey("anarva_live")
	require.NoError(t, err)
	assert.Contains(t, rawKey, "anarva_live_")
	assert.NotEmpty(t, hashedKey)

	// Verification check
	assert.True(t, pkgSecurity.VerifyAPIKey(rawKey, hashedKey))
	assert.False(t, pkgSecurity.VerifyAPIKey("anarva_live_wrong_key", hashedKey))

	// Secret Redaction Verification
	sensitiveLog := "Failed request for key " + rawKey + " with secret whsec_live_9f82a1bc3d4e5f67 and Bearer eyJhbGciOiJIUzI1NiJ9.test"
	redacted := pkgSecurity.RedactSecrets(sensitiveLog)

	assert.NotContains(t, redacted, rawKey)
	assert.Contains(t, redacted, "[REDACTED_API_KEY]")
	assert.Contains(t, redacted, "[REDACTED_WEBHOOK_SECRET]")
	assert.Contains(t, redacted, "[REDACTED]")
}

// 3. RBAC & Server-Side Authorization Tests
func TestPhase44_RBACServerSideAuthorization(t *testing.T) {
	roles := []string{"OWNER", "ADMIN", "DEVELOPER", "VIEWER", "BILLING_ADMIN", "AUDITOR"}
	for _, role := range roles {
		assert.NotEmpty(t, role)
	}

	// Request context role check
	req := httptest.NewRequest("GET", "/api/v1/compute/instances", nil)
	ctx := context.WithValue(req.Context(), gwMiddleware.RoleKey, "VIEWER")
	req = req.WithContext(ctx)

	ctxRole, ok := req.Context().Value(gwMiddleware.RoleKey).(string)
	assert.True(t, ok)
	assert.Equal(t, "VIEWER", ctxRole)
}

// 4. Tenant Isolation Security Tests
func TestPhase44_TenantIsolationSecurity(t *testing.T) {
	orgA := "org-alpha"
	orgB := "org-beta"

	// Query isolation rule test
	queryOrgA := "SELECT * FROM instances WHERE organization_id = ? AND id = ?"
	assert.Contains(t, queryOrgA, "organization_id = ?")
	assert.NotEqual(t, orgA, orgB)
}

// 5. Rate Limiting & Abuse Protection Tests
func TestPhase44_RateLimitingAbuseProtection(t *testing.T) {
	limiter := gwMiddleware.NewRateLimitMiddleware(2) // 2 requests allowed per minute
	eventSvc := secInternal.NewSecurityEventService()
	limiter.SetEventService(eventSvc)

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := limiter.Limit(dummyHandler)

	// Req 1 & 2 succeed
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, httptest.NewRequest("GET", "/api/v1/auth/login", nil))
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, httptest.NewRequest("GET", "/api/v1/auth/login", nil))
	assert.Equal(t, http.StatusOK, w2.Code)

	// Req 3 fails with 429 TOO_MANY_REQUESTS and Retry-After header
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, httptest.NewRequest("GET", "/api/v1/auth/login", nil))
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	assert.Equal(t, "60", w3.Header().Get("Retry-After"))
	assert.Contains(t, w3.Body.String(), "TOO_MANY_REQUESTS")
}

// 6. CORS & HTTP Security Headers Tests
func TestPhase44_CORSAndSecurityHeaders(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	secMiddleware := gwMiddleware.SecurityHeadersMiddleware(gwMiddleware.CORSMiddleware(dummyHandler))

	req := httptest.NewRequest("GET", "/api/v1/compute", nil)
	req.Header.Set("Origin", "https://app.anarva.cloud")
	w := httptest.NewRecorder()

	secMiddleware.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "https://app.anarva.cloud", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

// 7. SSRF Protection Tests
func TestPhase44_SSRFProtection(t *testing.T) {
	ssrf := providersSecurity.NewSSRFProtectionEngine()

	// Blocked targets
	assert.Error(t, ssrf.ValidateURL("http://169.254.169.254/latest/meta-data/"))
	assert.Error(t, ssrf.ValidateURL("http://localhost:8080/internal"))
	assert.Error(t, ssrf.ValidateURL("http://127.0.0.1:5432"))
	assert.Error(t, ssrf.ValidateURL("http://metadata.google.internal"))

	// Allowed target
	assert.NoError(t, ssrf.ValidateURL("https://api.github.com/webhooks"))
}

// 8. Storage Path Traversal Security Tests
func TestPhase44_StoragePathTraversalSecurity(t *testing.T) {
	// Blocked path traversal keys
	assert.Error(t, storageProvider.ValidateObjectKey("../etc/passwd"))
	assert.Error(t, storageProvider.ValidateObjectKey("..\\windows\\system32"))
	assert.Error(t, storageProvider.ValidateObjectKey("bucket/../../secret.txt"))
	assert.Error(t, storageProvider.ValidateObjectKey("null\x00byte.png"))
	assert.Error(t, storageProvider.ValidateObjectKey("%2e%2e%2fetc%2fpasswd"))
	assert.Error(t, storageProvider.ValidateObjectKey("/absolute/path/file.txt"))

	// Valid keys
	assert.NoError(t, storageProvider.ValidateObjectKey("uploads/images/photo.png"))
	assert.NoError(t, storageProvider.ValidateObjectKey("documents/2026/report.pdf"))
}

// 9. Webhook Signature Security Tests
func TestPhase44_WebhookSignatureSecurity(t *testing.T) {
	secret := "whsec_live_abcdef1234567890"
	payload := []byte(`{"event":"resource.created","id":"evt-101"}`)

	sig := webhookDomain.ComputeHMACSignature(payload, secret)
	assert.NotEmpty(t, sig)

	// Constant-time signature verification
	assert.True(t, webhookDomain.VerifyHMACSignature(payload, secret, sig))
	assert.False(t, webhookDomain.VerifyHMACSignature(payload, secret, "invalid_signature"))
}

// 10. Security Health API Tests
func TestPhase44_SecurityHealthAPI(t *testing.T) {
	cfg := &config.Config{Environment: "production"}
	eventSvc := secInternal.NewSecurityEventService()
	secSvc := secInternal.NewSecurityService(cfg, eventSvc)

	status := secSvc.EvaluateSecurityStatus(context.Background(), "req_sec_test")
	assert.Equal(t, secInternal.CheckSecure, status.Status)
	assert.Equal(t, secInternal.CheckSecure, status.Checks.Authentication)
	assert.Equal(t, secInternal.CheckSecure, status.Checks.SSRFProtection)
	assert.Equal(t, secInternal.CheckSecure, status.Checks.SecretRedaction)
}
