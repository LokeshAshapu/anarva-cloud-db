package providers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	localProvider "github.com/anarva-cloud/anarva-cloud-db/internal/providers/local"
	"github.com/anarva-cloud/anarva-cloud-db/internal/providers/registry"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/providers"
)

func TestPhase51_UniversalProviderContract(t *testing.T) {
	ctx := context.Background()
	lp := localProvider.NewLocalReferenceProvider()

	assert.Equal(t, "local", lp.ID())
	assert.Equal(t, "Local Reference Provider", lp.Name())
	assert.NoError(t, lp.CheckHealth(ctx))

	// 1. Compute Lifecycle & Error Mapping
	compOpts := providers.ComputeInstanceOpts{
		ID:        "inst-phase51-1",
		ProjectID: "proj-prod-1",
		Name:      "prod-app-server",
		Region:    "us-east-1",
		MinACU:    1.0,
		MaxACU:    4.0,
	}

	instDetails, err := lp.LaunchInstance(ctx, compOpts)
	require.NoError(t, err)
	assert.Equal(t, "inst-phase51-1", instDetails.ID)
	assert.Equal(t, "RUNNING", instDetails.Status)

	// Duplicate launch returns ErrResourceAlreadyExists
	_, errDup := lp.LaunchInstance(ctx, compOpts)
	assert.ErrorIs(t, errDup, providers.ErrResourceAlreadyExists)

	// Get Metrics
	metrics, errM := lp.GetInstanceMetrics(ctx, "inst-phase51-1")
	require.NoError(t, errM)
	assert.Contains(t, metrics, "cpu_utilization")

	// Scale & Stop
	assert.NoError(t, lp.ScaleComputeACU(ctx, "inst-phase51-1", 2.0))
	assert.NoError(t, lp.StopInstance(ctx, "inst-phase51-1"))
	assert.NoError(t, lp.StartInstance(ctx, "inst-phase51-1"))
	assert.NoError(t, lp.TerminateInstance(ctx, "inst-phase51-1"))

	// Non-existent operations return ErrResourceNotFound
	assert.ErrorIs(t, lp.StopInstance(ctx, "inst-missing"), providers.ErrResourceNotFound)

	// 2. Database Lifecycle & Error Mapping
	dbOpts := providers.DatabaseProvisionOpts{
		ID:     "db-phase51-1",
		Name:   "primary-postgres",
		Engine: "postgres",
	}

	dbDetails, errDB := lp.ProvisionDatabase(ctx, dbOpts)
	require.NoError(t, errDB)
	assert.Equal(t, "db-phase51-1", dbDetails.ID)
	assert.Equal(t, "HEALTHY", dbDetails.Status)

	backupID, errB := lp.BackupDatabase(ctx, "db-phase51-1")
	require.NoError(t, errB)
	assert.NotEmpty(t, backupID)

	assert.NoError(t, lp.RestoreDatabase(ctx, backupID, "db-phase51-restored"))
	assert.NoError(t, lp.DeleteDatabase(ctx, "db-phase51-1"))
	assert.ErrorIs(t, lp.DeleteDatabase(ctx, "db-missing"), providers.ErrResourceNotFound)

	// 3. Storage Lifecycle
	bucket, errS := lp.CreateBucket(ctx, "bucket-phase51-1", "us-east-1")
	require.NoError(t, errS)
	assert.Equal(t, "bucket-phase51-1", bucket.Name)

	assert.NoError(t, lp.UploadObject(ctx, "bucket-phase51-1", "file.txt", []byte("hello")))
	data, errD := lp.DownloadObject(ctx, "bucket-phase51-1", "file.txt")
	require.NoError(t, errD)
	assert.Equal(t, []byte("local_data_buffer"), data)

	assert.NoError(t, lp.DeleteObject(ctx, "bucket-phase51-1", "file.txt"))
	assert.NoError(t, lp.DeleteBucket(ctx, "bucket-phase51-1"))

	// 4. Networking Lifecycle
	vpcOpts := providers.VPCOpts{
		ID:        "vpc-phase51-1",
		Name:      "anarva-main-vpc",
		CIDRBlock: "10.0.0.0/16",
		Region:    "us-east-1",
	}

	vpc, errV := lp.CreateVPC(ctx, vpcOpts)
	require.NoError(t, errV)
	assert.Equal(t, "vpc-phase51-1", vpc.ID)

	subnetID, errSub := lp.CreateSubnet(ctx, "vpc-phase51-1", "10.0.1.0/24", "us-east-1a")
	require.NoError(t, errSub)
	assert.NotEmpty(t, subnetID)

	sgID, errSG := lp.CreateSecurityGroup(ctx, "vpc-phase51-1", "app-sg")
	require.NoError(t, errSG)
	assert.NotEmpty(t, sgID)

	assert.NoError(t, lp.DeleteVPC(ctx, "vpc-phase51-1"))
}

func TestPhase51_ProviderRegistryAndCapabilities(t *testing.T) {
	ctx := context.Background()
	reg := registry.NewProviderRegistry()
	assert.NotNil(t, reg)

	providersList, err := reg.ListProviders(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, providersList)

	localInfo, err := reg.GetProvider(ctx, "provider-local-docker")
	require.NoError(t, err)
	assert.Equal(t, "LOCAL_DOCKER", string(localInfo.Type))
	assert.True(t, localInfo.Capabilities.Compute)
	assert.True(t, localInfo.Capabilities.PostgreSQL)

	// Unknown provider returns error
	_, errUnknown := reg.GetProvider(ctx, "UNKNOWN_PROVIDER")
	assert.Error(t, errUnknown)
}
