package security_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pgService "github.com/anarva-cloud/anarva-cloud-db/internal/postgres/service"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
	stgDomain "github.com/anarva-cloud/anarva-cloud-db/internal/storage/domain"
	stgProvider "github.com/anarva-cloud/anarva-cloud-db/internal/storage/provider"
	stgService "github.com/anarva-cloud/anarva-cloud-db/internal/storage/service"
)

func TestTenantIsolation_CompleteAttackSuite(t *testing.T) {
	ctxA := context.WithValue(context.Background(), security.OrgIDKey, "org-tenant-a")
	ctxA = context.WithValue(ctxA, security.ProjectIDKey, "proj-tenant-a")
	ctxA = context.WithValue(ctxA, security.UserIDKey, "usr-tenant-a")

	ctxB := context.WithValue(context.Background(), security.OrgIDKey, "org-tenant-b")
	ctxB = context.WithValue(ctxB, security.ProjectIDKey, "proj-tenant-b")
	ctxB = context.WithValue(ctxB, security.UserIDKey, "usr-tenant-b")

	tcA := security.GetTenantContext(ctxA)
	tcB := security.GetTenantContext(ctxB)

	t.Run("TEST 1 & 2: User B cannot query or access User A database instance", func(t *testing.T) {
		sqlSvc := pgService.NewSQLService()
		dbIDA := fmt.Sprintf("db-tenant-a-%d", time.Now().UnixNano())

		_, errA := sqlSvc.ExecuteQuery(ctxA, dbIDA, "CREATE TABLE secret_users (id INT, email TEXT);")
		require.NoError(t, errA)

		_, errAInsert := sqlSvc.ExecuteQuery(ctxA, dbIDA, "INSERT INTO secret_users (id, email) VALUES (1, 'secret_a@tenant-a.com');")
		require.NoError(t, errAInsert)

		errIsolation := tcB.EnforceOwnership("org-tenant-a", "proj-tenant-a")
		assert.Error(t, errIsolation)
		assert.Contains(t, errIsolation.Error(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("TEST 4 & 5: User B cannot access or download User A storage bucket object", func(t *testing.T) {
		stgProv := stgProvider.NewLocalStorageProvider(t.TempDir())

		bucketA, errBkt := stgProv.CreateBucket(ctxA, &stgDomain.Bucket{
			ID:             "bkt-tenant-a",
			OrganizationID: "org-tenant-a",
			ProjectID:      "proj-tenant-a",
			Name:           "bucket-a",
		})
		require.NoError(t, errBkt)

		errIsolation := tcB.EnforceOwnership(bucketA.OrganizationID, bucketA.ProjectID)
		assert.Error(t, errIsolation)
	})

	t.Run("TEST 6: Presigned URL HMAC Signature Verification Rejects Tampering", func(t *testing.T) {
		stgProv := stgProvider.NewLocalStorageProvider(t.TempDir())
		stgSignedUrlSvc := stgService.NewSignedURLService(stgProv)

		pURL, errGen := stgSignedUrlSvc.GenerateSignedURL(ctxA, "bkt-tenant-a", "secret.pdf", "GET", 300)
		require.NoError(t, errGen)
		assert.NotNil(t, pURL)

		errTampered := stgSignedUrlSvc.ValidateSignedURL("bkt-tenant-a", "secret.pdf", "GET", "invalid_sig_123", time.Now().Unix()+300)
		assert.Error(t, errTampered)
		assert.Contains(t, errTampered.Error(), "INVALID_SIGNATURE")
	})

	t.Run("TEST 14: Forged Organization ID in request context is rejected", func(t *testing.T) {
		errForged := tcB.EnforceOwnership(tcA.OrganizationID, tcA.ProjectID)
		assert.Error(t, errForged)
		assert.Contains(t, errForged.Error(), "TENANT_ISOLATION_VIOLATION")
	})

	t.Run("TEST 15: Cross-Project Isolation within same Organization", func(t *testing.T) {
		ctxProj1 := context.WithValue(context.Background(), security.OrgIDKey, "org-shared")
		ctxProj1 = context.WithValue(ctxProj1, security.ProjectIDKey, "proj-01")

		ctxProj2 := context.WithValue(context.Background(), security.OrgIDKey, "org-shared")
		ctxProj2 = context.WithValue(ctxProj2, security.ProjectIDKey, "proj-02")

		tcProj1 := security.GetTenantContext(ctxProj1)
		tcProj2 := security.GetTenantContext(ctxProj2)

		errCrossProj := tcProj2.EnforceOwnership(tcProj1.OrganizationID, tcProj1.ProjectID)
		assert.Error(t, errCrossProj)
		assert.Contains(t, errCrossProj.Error(), "Project access denied")
	})
}
