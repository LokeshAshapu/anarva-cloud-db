package compute_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	computeDelivery "github.com/anarva-cloud/anarva-cloud-db/internal/compute/delivery/http"
	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	computeRepo "github.com/anarva-cloud/anarva-cloud-db/internal/compute/repository"
	computeUseCase "github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
	mapping "github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
)

func TestPhase70C_EndToEndComputeWorkflow_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}

	var cRepo computeDomain.ComputeRepository
	var mapRepo mapping.MappingRepository

	if dsn != "" {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			_ = db.AutoMigrate(&computeDomain.ComputeInstance{}, &computeDomain.Volume{}, &mapping.ProviderResourceMapping{})
			cRepo = computeRepo.NewPostgresComputeRepository(db)
			mapRepo = mapping.NewPostgresMappingRepository(db)
			t.Logf("[Phase 70C Audit] Connected to PostgreSQL Control Plane DB at %s", dsn)
		}
	}

	if cRepo == nil {
		cRepo = newMemComputeRepo()
		mapRepo = mapping.NewInMemoryMappingRepository()
		t.Log("[Phase 70C Audit] Using In-Memory Repositories for Control Plane Metadata")
	}

	ctx := context.Background()
	_, dockerErr := exec.LookPath("docker")
	isRealDocker := (dockerErr == nil)
	t.Logf("[Phase 70C Audit] Real Docker Host Available: %v", isRealDocker)

	prov1 := computeProvider.NewLocalDockerComputeProvider()
	uc1 := computeUseCase.NewComputeUseCase(cRepo, nil, prov1)
	uc1.SetMappingRepository(mapRepo)

	handler1 := computeDelivery.NewComputeHandler(uc1, nil)
	mux1 := http.NewServeMux()
	handler1.RegisterRoutes(mux1)

	instID := "acu-e2e-phase70c-node"
	instName := "e2e-worker-70c"

	// Cleanup prior runs
	_ = uc1.DeleteInstance(ctx, instID)
	if isRealDocker {
		_ = exec.Command("docker", "rm", "-f", "anarva-acu-"+instName).Run()
	}

	// Helper to make authenticated HTTP requests with TenantContext
	doRequest := func(handler http.Handler, method, path string, orgID, projID string, body string) *httptest.ResponseRecorder {
		var req *http.Request
		if body != "" {
			req = httptest.NewRequest(method, path, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		if orgID != "" || projID != "" {
			reqCtx := context.WithValue(req.Context(), security.OrgIDKey, orgID)
			reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, projID)
			req = req.WithContext(reqCtx)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	// 1. CREATE COMPUTE INSTANCE
	t.Run("1. Create Compute Instance via HTTP API", func(t *testing.T) {
		createBody := `{"id":"` + instID + `","name":"` + instName + `","acu":1.0,"imageId":"img-alpine-320","dockerImage":"nginx:alpine","regionId":"us-east-1"}`
		rec := doRequest(mux1, "POST", "/api/v1/compute/instances", "org-tenant-a", "proj-tenant-a", createBody)

		require.Equal(t, http.StatusCreated, rec.Code, "Create response body: %s", rec.Body.String())

		var created computeDomain.ComputeInstance
		err := json.Unmarshal(rec.Body.Bytes(), &created)
		require.NoError(t, err)
		instID = created.ID
		assert.NotEmpty(t, instID)
		assert.Equal(t, "org-tenant-a", created.OrganizationID)
		assert.Equal(t, "proj-tenant-a", created.ProjectID)

		// Verify PostgreSQL control-plane DB record
		dbInst, getErr := cRepo.GetByID(ctx, instID)
		require.NoError(t, getErr)
		assert.Equal(t, instID, dbInst.ID)
		assert.Equal(t, "org-tenant-a", dbInst.OrganizationID)

		// Verify provider_resource_mappings DB record
		mapEntry, mapErr := mapRepo.GetMapping(instID)
		require.NoError(t, mapErr)
		assert.Equal(t, instID, mapEntry.AnarvaResourceID)
		assert.Equal(t, "org-tenant-a", mapEntry.OrganizationID)
		assert.Equal(t, "proj-tenant-a", mapEntry.ProjectID)

		// If real Docker is present, verify host Docker container existence
		if isRealDocker {
			out, inspectErr := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", "anarva-acu-"+instName).CombinedOutput()
			require.NoError(t, inspectErr, "Docker container 'anarva-acu-%s' must exist on host Docker engine", instName)
			t.Logf("[Phase 70C Live Docker] Host Container Status: %s", strings.TrimSpace(string(out)))
		}
	})

	// 2. READ COMPUTE INSTANCE
	t.Run("2. Read Compute Instance", func(t *testing.T) {
		rec := doRequest(mux1, "GET", "/api/v1/compute/instances/"+instID, "org-tenant-a", "proj-tenant-a", "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), instID)
	})

	// 3. STOP COMPUTE INSTANCE
	t.Run("3. Stop Compute Instance", func(t *testing.T) {
		rec := doRequest(mux1, "POST", "/api/v1/compute/instances/"+instID+"/stop", "org-tenant-a", "proj-tenant-a", "")
		require.Equal(t, http.StatusOK, rec.Code)

		// Verify DB status updated to STOPPED
		dbInst, err := cRepo.GetByID(ctx, instID)
		require.NoError(t, err)
		assert.Equal(t, computeDomain.StatusStopped, dbInst.Status)

		if isRealDocker {
			out, _ := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", "anarva-acu-"+instName).CombinedOutput()
			assert.Equal(t, "exited", strings.TrimSpace(string(out)))
		}
	})

	// 4. START COMPUTE INSTANCE
	t.Run("4. Start Compute Instance", func(t *testing.T) {
		rec := doRequest(mux1, "POST", "/api/v1/compute/instances/"+instID+"/start", "org-tenant-a", "proj-tenant-a", "")
		require.Equal(t, http.StatusOK, rec.Code)

		// Verify DB status updated to RUNNING
		dbInst, err := cRepo.GetByID(ctx, instID)
		require.NoError(t, err)
		assert.Equal(t, computeDomain.StatusRunning, dbInst.Status)

		if isRealDocker {
			out, _ := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", "anarva-acu-"+instName).CombinedOutput()
			assert.Equal(t, "running", strings.TrimSpace(string(out)))
		}
	})

	// 5. RESTART COMPUTE INSTANCE
	t.Run("5. Restart Compute Instance", func(t *testing.T) {
		rec := doRequest(mux1, "POST", "/api/v1/compute/instances/"+instID+"/restart", "org-tenant-a", "proj-tenant-a", "")
		require.Equal(t, http.StatusOK, rec.Code)

		dbInst, err := cRepo.GetByID(ctx, instID)
		require.NoError(t, err)
		assert.Equal(t, computeDomain.StatusRunning, dbInst.Status)
	})

	// 6. EXECUTE COMMAND INSIDE CONTAINER
	t.Run("6. Execute Command Inside Container", func(t *testing.T) {
		execBody := `{"command":"echo 'ANARVA_PHASE70C_SUCCESS'"}`
		rec := doRequest(mux1, "POST", "/api/v1/compute/instances/"+instID+"/execute", "org-tenant-a", "proj-tenant-a", execBody)
		require.Equal(t, http.StatusOK, rec.Code)

		var res computeDomain.CommandExecutionResult
		err := json.Unmarshal(rec.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, 0, res.ExitCode)
		assert.Contains(t, res.Stdout, "ANARVA_PHASE70C_SUCCESS")
	})

	// 7. GATEWAY RESTART SIMULATION & RECOVERY
	t.Run("7. Gateway Restart Simulation & Recovery", func(t *testing.T) {
		// Re-instantiate gateway handler with fresh provider (RAM cache erased)
		prov2 := computeProvider.NewLocalDockerComputeProvider()
		uc2 := computeUseCase.NewComputeUseCase(cRepo, nil, prov2)
		uc2.SetMappingRepository(mapRepo)

		handler2 := computeDelivery.NewComputeHandler(uc2, nil)
		mux2 := http.NewServeMux()
		handler2.RegisterRoutes(mux2)

		// A. GET Instance succeeds post-restart
		recGet := doRequest(mux2, "GET", "/api/v1/compute/instances/"+instID, "org-tenant-a", "proj-tenant-a", "")
		require.Equal(t, http.StatusOK, recGet.Code)
		assert.Contains(t, recGet.Body.String(), instID)

		// B. STOP Instance succeeds post-restart
		recStop := doRequest(mux2, "POST", "/api/v1/compute/instances/"+instID+"/stop", "org-tenant-a", "proj-tenant-a", "")
		require.Equal(t, http.StatusOK, recStop.Code)

		// C. START Instance succeeds post-restart
		recStart := doRequest(mux2, "POST", "/api/v1/compute/instances/"+instID+"/start", "org-tenant-a", "proj-tenant-a", "")
		require.Equal(t, http.StatusOK, recStart.Code)

		// D. RESTART Instance succeeds post-restart
		recRestart := doRequest(mux2, "POST", "/api/v1/compute/instances/"+instID+"/restart", "org-tenant-a", "proj-tenant-a", "")
		require.Equal(t, http.StatusOK, recRestart.Code)

		// E. EXECUTE Command succeeds post-restart
		execBody := `{"command":"uname -a"}`
		recExec := doRequest(mux2, "POST", "/api/v1/compute/instances/"+instID+"/execute", "org-tenant-a", "proj-tenant-a", execBody)
		require.Equal(t, http.StatusOK, recExec.Code)
		assert.Contains(t, recExec.Body.String(), `"exitCode":0`)
	})

	// 8. TENANT ISOLATION VERIFICATION POST-RESTART
	t.Run("8. Tenant Isolation Verification Post-Restart", func(t *testing.T) {
		prov3 := computeProvider.NewLocalDockerComputeProvider()
		uc3 := computeUseCase.NewComputeUseCase(cRepo, nil, prov3)
		uc3.SetMappingRepository(mapRepo)
		mux3 := http.NewServeMux()
		computeDelivery.NewComputeHandler(uc3, nil).RegisterRoutes(mux3)

		// Tenant B attempting GET Tenant A instance -> 403
		recBGet := doRequest(mux3, "GET", "/api/v1/compute/instances/"+instID, "org-tenant-b", "proj-tenant-b", "")
		assert.Equal(t, http.StatusForbidden, recBGet.Code)
		assert.Contains(t, recBGet.Body.String(), "TENANT_ISOLATION_VIOLATION")

		// Tenant B attempting STOP -> 403
		recBStop := doRequest(mux3, "POST", "/api/v1/compute/instances/"+instID+"/stop", "org-tenant-b", "proj-tenant-b", "")
		assert.Equal(t, http.StatusForbidden, recBStop.Code)

		// Tenant B attempting START -> 403
		recBStart := doRequest(mux3, "POST", "/api/v1/compute/instances/"+instID+"/start", "org-tenant-b", "proj-tenant-b", "")
		assert.Equal(t, http.StatusForbidden, recBStart.Code)

		// Tenant B attempting DELETE -> 403
		recBDel := doRequest(mux3, "DELETE", "/api/v1/compute/instances/"+instID, "org-tenant-b", "proj-tenant-b", "")
		assert.Equal(t, http.StatusForbidden, recBDel.Code)

		// Cross-Project Isolation (Same Org, Different Project) -> 403
		recCrossProj := doRequest(mux3, "GET", "/api/v1/compute/instances/"+instID, "org-tenant-a", "proj-other", "")
		assert.Equal(t, http.StatusForbidden, recCrossProj.Code)
	})

	// 9. INSTANCE TERMINATION & CLEANUP
	t.Run("9. Delete Instance & Provider Resource Cleanup", func(t *testing.T) {
		recDel := doRequest(mux1, "DELETE", "/api/v1/compute/instances/"+instID, "org-tenant-a", "proj-tenant-a", "")
		assert.Equal(t, http.StatusNoContent, recDel.Code)

		// Verify compute_instances is soft-deleted in DB
		_, err := cRepo.GetByID(ctx, instID)
		require.Error(t, err, "GetByID must return error for soft-deleted compute instance")

		// Verify provider_resource_mappings is deleted
		_, mapErr := mapRepo.GetMapping(instID)
		require.Error(t, mapErr, "provider_resource_mappings must be deleted for terminated compute instance")

		// If real Docker is present, verify host Docker container is removed (no orphan container)
		if isRealDocker {
			time.Sleep(300 * time.Millisecond)
			inspectErr := exec.Command("docker", "inspect", "anarva-acu-"+instName).Run()
			require.Error(t, inspectErr, "Docker container 'anarva-acu-%s' must be removed from host Docker engine", instName)
			t.Logf("[Phase 70C Orphan Check] Host Container 'anarva-acu-%s' cleanly destroyed. ZERO orphan containers remain.", instName)
		}
	})
}
