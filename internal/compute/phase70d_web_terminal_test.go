package compute_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computeDelivery "github.com/anarva-cloud/anarva-cloud-db/internal/compute/delivery/http"
	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	computeUseCase "github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
	mapping "github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

func TestPhase70D_RealComputeWebTerminalExecution(t *testing.T) {
	ctx := context.Background()

	cRepo := newMemComputeRepo()
	mapRepo := mapping.NewInMemoryMappingRepository()
	prov := computeProvider.NewLocalDockerComputeProvider()
	uc := computeUseCase.NewComputeUseCase(cRepo, nil, prov)
	uc.SetMappingRepository(mapRepo)

	handler := computeDelivery.NewComputeHandler(uc, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Check if host Docker is available
	_, dockerErr := exec.LookPath("docker")
	isRealDocker := (dockerErr == nil)

	// 1. Provision Compute Instance for Tenant A
	instName := "terminal-worker-70d"
	instA, err := uc.CreateInstance(ctx, &computeDomain.ComputeInstance{
		ID:             "acu-inst-terminal-70d",
		OrganizationID: "org-tenant-a",
		ProjectID:      "proj-tenant-a",
		Name:           instName,
		ACU:            1.0,
		RegionID:       "us-east-1",
		DockerImage:    "nginx:alpine",
	})
	require.NoError(t, err)

	defer func() {
		_ = uc.DeleteInstance(ctx, instA.ID)
		if isRealDocker {
			_ = exec.Command("docker", "rm", "-f", "anarva-acu-"+instName).Run()
		}
	}()

	t.Run("1. Real Execute API Returns Real Container Output", func(t *testing.T) {
		cmdPayload := `{"command":"echo ANARVA_PHASE70D_SUCCESS"}`
		req := httptest.NewRequest("POST", "/api/v1/compute/instances/"+instA.ID+"/execute", strings.NewReader(cmdPayload))
		req.Header.Set("Content-Type", "application/json")
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-a")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-a")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var res computeDomain.CommandExecutionResult
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)

		assert.Equal(t, 0, res.ExitCode)
		assert.Contains(t, res.Stdout, "ANARVA_PHASE70D_SUCCESS")
		t.Logf("[Phase 70D Execution Output]: %s", strings.TrimSpace(res.Stdout))
	})

	t.Run("2. Empty Command Payload Rejected with HTTP 400 Bad Request", func(t *testing.T) {
		cmdPayload := `{"command":""}`
		req := httptest.NewRequest("POST", "/api/v1/compute/instances/"+instA.ID+"/execute", strings.NewReader(cmdPayload))
		req.Header.Set("Content-Type", "application/json")
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-a")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-a")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "command payload cannot be empty")
	})

	t.Run("3. Unauthorized Tenant B Execute Request Rejected with HTTP 403", func(t *testing.T) {
		cmdPayload := `{"command":"whoami"}`
		req := httptest.NewRequest("POST", "/api/v1/compute/instances/"+instA.ID+"/execute", strings.NewReader(cmdPayload))
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
