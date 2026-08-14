package billing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/billing/domain"
	"github.com/anarva-cloud/anarva-cloud-db/internal/billing/usecase"
)

func TestBillingEngine_CalculateRealUsageAndInvoice(t *testing.T) {
	uc := usecase.NewBillingUseCase()

	inv, err := uc.CalculateRealUsageAndInvoice(context.Background(), "org-alpha-101", "proj-101")
	require.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, "org-alpha-101", inv.OrganizationID)
	assert.Equal(t, "USD", inv.Currency)
	assert.Equal(t, domain.InvoiceDraft, inv.Status)
	assert.Equal(t, "anarva-v1.0.0", inv.PricingVersion)
	assert.Equal(t, "ANARVA_ESTIMATED_CUSTOMER_CHARGE", inv.RealityLabel)

	// Verify line items for EC2, RDS, and S3
	assert.Len(t, inv.Lines, 3)
	assert.Equal(t, "COMPUTE", inv.Lines[0].ResourceType)
	assert.Equal(t, "DATABASE", inv.Lines[1].ResourceType)
	assert.Equal(t, "STORAGE", inv.Lines[2].ResourceType)

	assert.Greater(t, inv.Total, 0.0)
}

func TestBillingEngine_FinalizeInvoice_Immutability(t *testing.T) {
	uc := usecase.NewBillingUseCase()

	inv, err := uc.CalculateRealUsageAndInvoice(context.Background(), "org-alpha-101", "proj-101")
	require.NoError(t, err)

	finalized, err := uc.FinalizeInvoice(context.Background(), inv.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.InvoiceFinalized, finalized.Status)

	// Verify status remains FINALIZED on subsequent calls
	refinalized, err := uc.FinalizeInvoice(context.Background(), inv.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.InvoiceFinalized, refinalized.Status)
}

func TestBillingEngine_TenantIsolation(t *testing.T) {
	uc := usecase.NewBillingUseCase()

	invA, err := uc.CalculateRealUsageAndInvoice(context.Background(), "org-a", "proj-a")
	require.NoError(t, err)

	invsOrgB := uc.ListInvoices(context.Background(), "org-b")
	for _, inv := range invsOrgB {
		assert.NotEqual(t, invA.ID, inv.ID)
	}
}
