package observability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gwMiddleware "github.com/anarva-cloud/anarva-cloud-db/internal/gateway/middleware"
	healthService "github.com/anarva-cloud/anarva-cloud-db/internal/health"
	reliabilityDomain "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/domain"
	reliabilityUsecase "github.com/anarva-cloud/anarva-cloud-db/internal/reliability/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/config"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/metrics"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

// 1. Request ID Propagation & Correlation Middleware Test
func TestPhase46_RequestIDPropagation(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID, ok := r.Context().Value(gwMiddleware.RequestIDKey).(string)
		assert.True(t, ok)
		assert.NotEmpty(t, reqID)
		w.WriteHeader(http.StatusOK)
	})

	h := gwMiddleware.CorrelationMiddleware(dummyHandler)

	// Case A: Custom request ID preserved
	reqA := httptest.NewRequest("GET", "/api/v1/compute/instances", nil)
	reqA.Header.Set("X-Request-ID", "req-custom-trace-101")
	wA := httptest.NewRecorder()
	h.ServeHTTP(wA, reqA)

	assert.Equal(t, http.StatusOK, wA.Code)
	assert.Equal(t, "req-custom-trace-101", wA.Header().Get("X-Request-ID"))

	// Case B: Missing request ID auto-generated
	reqB := httptest.NewRequest("GET", "/api/v1/compute/instances", nil)
	wB := httptest.NewRecorder()
	h.ServeHTTP(wB, reqB)

	assert.Equal(t, http.StatusOK, wB.Code)
	assert.NotEmpty(t, wB.Header().Get("X-Request-ID"))
	assert.Contains(t, wB.Header().Get("X-Request-ID"), "req-")
}

// 2. Operation Timeline Step Recording Test
func TestPhase46_OperationTimelineEvents(t *testing.T) {
	ctx := context.Background()
	uc := reliabilityUsecase.NewReliabilityUseCase()

	op, err := uc.DispatchOperation(ctx, "org-trace-alpha", "proj-trace", "res-comp-trace", reliabilityDomain.OpCreateCompute, "idemp-trace-1", "payload", "req_trace_101")
	require.NoError(t, err)

	assert.Equal(t, "org-trace-alpha", op.OrganizationID)
	assert.NotEmpty(t, op.Timeline)
	assert.GreaterOrEqual(t, len(op.Timeline), 3)

	// Timeline events contain expected steps
	assert.Equal(t, "Validate Authorization & Quota", op.Timeline[0].Name)
	assert.Equal(t, "Acquire Resource Lock Lease", op.Timeline[1].Name)
}

// 3. Audit Investigation Correlation Query Under Tenant Isolation
func TestPhase46_AuditCorrelationAndTenantIsolation(t *testing.T) {
	uc := reliabilityUsecase.NewReliabilityUseCase()

	// Org A audit log query
	auditsOrgA := uc.ListAuditEvents("org-default", "proj-default")
	assert.NotEmpty(t, auditsOrgA)

	// Org B attempting to query Org A audit logs -> returns empty (isolated)
	auditsOrgB := uc.ListAuditEvents("org-attacker-beta", "proj-default")
	assert.Empty(t, auditsOrgB)
}

// 4. Resource Health Intelligence & Readiness Probe
func TestPhase46_ResourceHealthCalculation(t *testing.T) {
	cfg := &config.Config{Environment: "production"}
	uc := reliabilityUsecase.NewReliabilityUseCase()

	hSvc := healthService.NewHealthService(nil, cfg, nil, uc, "0.1.0")
	checks := hSvc.CheckReadiness(context.Background())

	// Without PostgreSQL in production -> StatusUnavailable
	assert.Equal(t, healthService.StatusUnavailable, checks.Database)

	// Status constants check
	assert.Equal(t, healthService.ComponentStatus("HEALTHY"), healthService.StatusHealthy)
	assert.Equal(t, healthService.ComponentStatus("RECOVERING"), healthService.StatusRecovering)
}

// 5. Prometheus Metrics Instrumentation Test
func TestPhase46_PrometheusMetricsInstrumentation(t *testing.T) {
	metrics.RecordHTTPRequest(http.StatusOK, "GET", "/api/v1/compute/instances", 0.025)
	metrics.RecordDatabaseQuery("SELECT_INSTANCES", "success", 0.005)
	metrics.RecordOperationEvent("CREATE_COMPUTE", "SUCCEEDED")
	metrics.RecordOperationRecovery("SUCCEEDED", 1.2)
	metrics.RecordLockConflict("res-compute-01")
	metrics.RecordLockExpiration("res-db-01")

	handler := metrics.Handler()
	assert.NotNil(t, handler)
}

// 6. Normalized Error Categorization & Secret Redaction
func TestPhase46_ErrorCategorizationAndRedaction(t *testing.T) {
	rawSecret := "anarva_live_ak_1234567890abcdef"
	appErr := appErrors.New(appErrors.CodeAuthenticationError, "Invalid authentication key "+rawSecret)

	assert.Equal(t, appErrors.CodeAuthenticationError, appErr.Code)
	redactedMessage := security.RedactSecrets(appErr.Error())

	assert.NotContains(t, redactedMessage, rawSecret)
	assert.Contains(t, redactedMessage, "[REDACTED_API_KEY]")
}
