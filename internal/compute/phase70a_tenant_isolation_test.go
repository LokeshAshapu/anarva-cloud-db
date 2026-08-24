package compute_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computeDelivery "github.com/anarva-cloud/anarva-cloud-db/internal/compute/delivery/http"
	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	computeUseCase "github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

func TestPhase70A_ComputeTenantIsolation(t *testing.T) {
	prov := computeProvider.NewLocalDockerComputeProvider()
	uc := computeUseCase.NewComputeUseCase(nil, nil, prov)
	handler := computeDelivery.NewComputeHandler(uc, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	ctx := context.Background()

	// 1. Create Instance for Tenant A (org-tenant-a / proj-tenant-a)
	instA, err := uc.CreateInstance(ctx, &computeDomain.ComputeInstance{
		ID:             "acu-inst-tenant-a",
		OrganizationID: "org-tenant-a",
		ProjectID:      "proj-tenant-a",
		Name:           "Tenant A Worker",
		ACU:            1.0,
		RegionID:       "us-east-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "org-tenant-a", instA.OrganizationID)
	assert.Equal(t, "proj-tenant-a", instA.ProjectID)

	t.Run("1. Tenant A can GET its own compute instance", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/compute/instances/"+instA.ID, nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-a")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-a")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), instA.ID)
	})

	t.Run("2. Tenant B GET Tenant A instance -> 403 TENANT_ISOLATION_VIOLATION", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/compute/instances/"+instA.ID, nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-b")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-b")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("3. Tenant B START Tenant A instance -> 403 TENANT_ISOLATION_VIOLATION", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/compute/instances/"+instA.ID+"/start", nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-b")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-b")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("4. Tenant B STOP Tenant A instance -> 403 TENANT_ISOLATION_VIOLATION", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/compute/instances/"+instA.ID+"/stop", nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-b")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-b")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("5. Tenant B RESTART Tenant A instance -> 403 TENANT_ISOLATION_VIOLATION", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/compute/instances/"+instA.ID+"/restart", nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-b")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-b")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("6. Tenant B DELETE Tenant A instance -> 403 TENANT_ISOLATION_VIOLATION", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/compute/instances/"+instA.ID, nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-b")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-b")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("7. Cross-Project Isolation (Same Org, Different Project) -> 403 TENANT_ISOLATION_VIOLATION", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/compute/instances/"+instA.ID, nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-a")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-other")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("8. Instance Listing returns ONLY instances for authenticated TenantContext", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/compute/instances?projectId=proj-tenant-a", nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-b")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-b")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var list []computeDomain.ComputeInstance
		err := json.Unmarshal(rec.Body.Bytes(), &list)
		require.NoError(t, err)

		// Tenant B listing must NOT return Tenant A's instance
		for _, item := range list {
			assert.NotEqual(t, instA.ID, item.ID, "Tenant B instance list must not contain Tenant A instance")
		}
	})

	t.Run("9. Create Request Cannot Override Authenticated TenantContext", func(t *testing.T) {
		payload := `{"name":"Attacker Worker","organizationId":"org-tenant-a","projectId":"proj-tenant-a","acu":1.0}`
		req := httptest.NewRequest("POST", "/api/v1/compute/instances", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-b")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-b")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "TENANT_ISOLATION_VIOLATION")
	})
}
