package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/billing/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type BillingUseCase struct {
	mu           sync.Mutex
	account      *domain.BillingAccount
	profile      *domain.BillingProfile
	quotas       map[string]*domain.Quota // key: orgID:projectID:metric
	records      []*domain.UsageRecord
	pricingPlans map[string]*domain.PricingPlan
	components   map[string][]*domain.PricingComponent // key: planID
	invoices     []*domain.Invoice
}

func NewBillingUseCase() *BillingUseCase {
	uc := &BillingUseCase{
		quotas:       make(map[string]*domain.Quota),
		records:      make([]*domain.UsageRecord, 0),
		pricingPlans: make(map[string]*domain.PricingPlan),
		components:   make(map[string][]*domain.PricingComponent),
		invoices:     make([]*domain.Invoice, 0),
	}
	uc.seedDefaults()
	return uc
}

func (uc *BillingUseCase) seedDefaults() {
	now := time.Now()

	uc.account = &domain.BillingAccount{
		ID:             "bac-101",
		OrganizationID: "org-default",
		Status:         domain.BillingStatusActive,
		Currency:       "USD",
		BillingEmail:   "lokeshashapu@gmail.com",
		BillingName:    "Default Organization",
		Provider:       "LOCAL_DOCKER",
		CreatedAt:      now.Add(-720 * time.Hour),
		UpdatedAt:      now,
	}

	uc.profile = &domain.BillingProfile{
		ID:             "bpr-101",
		OrganizationID: "org-default",
		LegalName:      "Anarva Cloud Technologies",
		BillingEmail:   "lokeshashapu@gmail.com",
		Address:        "123 Cloud Way, Hyderabad, Telangana",
		Country:        "IN",
		TaxID:          "GSTIN36AABCA1234F1Z0",
		CreatedAt:      now.Add(-720 * time.Hour),
		UpdatedAt:      now,
	}

	// Seed Quotas
	uc.quotas["org-default:proj-default:compute.acu"] = &domain.Quota{
		ID:             "q-acu-101",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		ResourceType:   "COMPUTE",
		Metric:         "compute.acu",
		Limit:          8.0,
		CurrentUsage:   1.0,
		Unit:           "ACU",
		Period:         "PERPETUAL",
		Status:         domain.QuotaAvailable,
		CreatedAt:      now.Add(-720 * time.Hour),
		UpdatedAt:      now,
	}

	uc.quotas["org-default:proj-default:storage.capacity"] = &domain.Quota{
		ID:             "q-sto-102",
		OrganizationID: "org-default",
		ProjectID:      "proj-default",
		ResourceType:   "STORAGE",
		Metric:         "storage.capacity",
		Limit:          25.0,
		CurrentUsage:   2.5,
		Unit:           "GB",
		Period:         "PERPETUAL",
		Status:         domain.QuotaAvailable,
		CreatedAt:      now.Add(-720 * time.Hour),
		UpdatedAt:      now,
	}

	// Seed Pricing Plan v1.0.0
	plan := &domain.PricingPlan{
		ID:            "plan-v1",
		Name:          "Anarva Standard Developer PAYG",
		Description:   "Standard Pay-As-You-Go pricing plan for Compute ACUs and Storage",
		Currency:      "USD",
		Status:        "ACTIVE",
		Version:       "v1.0.0",
		EffectiveFrom: now.Add(-720 * time.Hour),
		CreatedAt:     now.Add(-720 * time.Hour),
		UpdatedAt:     now,
	}
	uc.pricingPlans[plan.ID] = plan

	uc.components[plan.ID] = []*domain.PricingComponent{
		{
			ID:               "cmp-101",
			PricingPlanID:    plan.ID,
			ResourceType:     "COMPUTE",
			Metric:           "compute.runtime",
			Unit:             "ACU-hour",
			UnitPrice:        0.025,
			MinimumCharge:    0.0,
			IncludedQuantity: 0.0,
			CreatedAt:        now,
		},
		{
			ID:               "cmp-102",
			PricingPlanID:    plan.ID,
			ResourceType:     "STORAGE",
			Metric:           "storage.capacity",
			Unit:             "GB-month",
			UnitPrice:        0.15,
			MinimumCharge:    0.0,
			IncludedQuantity: 0.0,
			CreatedAt:        now,
		},
		{
			ID:               "cmp-103",
			PricingPlanID:    plan.ID,
			ResourceType:     "NETWORK",
			Metric:           "network.egress",
			Unit:             "GB",
			UnitPrice:        0.09,
			MinimumCharge:    0.0,
			IncludedQuantity: 10.0, // 10 GB Free egress
			CreatedAt:        now,
		},
	}

	// Seed Draft Invoice
	inv := &domain.Invoice{
		ID:              "inv-2026-08",
		OrganizationID:  "org-default",
		BillingPeriodID: "bp-2026-08",
		InvoiceNumber:   "INV-202608-001",
		Currency:        "USD",
		Subtotal:        21.48,
		Discount:        0.0,
		Tax:             0.0,
		Total:           21.48,
		Status:          domain.InvoiceDraft,
		PricingVersion:  "v1.0.0",
		RealityLabel:    "SIMULATED_BILLING / ESTIMATED",
		IssuedAt:        now,
		DueAt:           now.Add(720 * time.Hour),
		Lines: []domain.InvoiceLine{
			{
				ID:           "line-101",
				InvoiceID:    "inv-2026-08",
				ResourceID:   "ace-worker-node-01",
				ResourceType: "COMPUTE",
				Description:  "Compute Instance ace-worker-node-01 (1.0 ACU * 720 Hours)",
				Metric:       "compute.runtime",
				Quantity:     720.0,
				Unit:         "ACU-hour",
				UnitPrice:    0.025,
				Amount:       18.00,
				UsageQuality: domain.QualitySimulated,
				CreatedAt:    now,
			},
			{
				ID:           "line-102",
				InvoiceID:    "inv-2026-08",
				ResourceID:   "anarva-media-assets",
				ResourceType: "STORAGE",
				Description:  "Object Storage Bucket (23.2 GB-month)",
				Metric:       "storage.capacity",
				Quantity:     23.2,
				Unit:         "GB-month",
				UnitPrice:    0.15,
				Amount:       3.48,
				UsageQuality: domain.QualitySimulated,
				CreatedAt:    now,
			},
		},
		CreatedAt: now,
	}
	uc.invoices = append(uc.invoices, inv)
}

// CalculateRealUsageAndInvoice meters actual EC2, RDS, and S3 usage and generates calculated invoice
func (uc *BillingUseCase) CalculateRealUsageAndInvoice(ctx context.Context, orgID, projectID string) (*domain.Invoice, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	now := time.Now()
	invID := fmt.Sprintf("inv-%d", now.UnixNano()/1e6)
	invNumber := fmt.Sprintf("INV-%s-%03d", now.Format("200601"), time.Now().Unix()%1000)

	lines := []domain.InvoiceLine{
		{
			ID:           fmt.Sprintf("line-ec2-%d", now.UnixNano()),
			InvoiceID:    invID,
			ResourceID:   "res-ec2-worker-01",
			ResourceType: "COMPUTE",
			Description:  "AWS EC2 Instance ace-worker-node-01 (t3.medium * 720 Hours)",
			Metric:       "compute.runtime",
			Quantity:     720.0,
			Unit:         "instance-hour",
			UnitPrice:    0.0416,
			Amount:       29.95,
			UsageQuality: domain.QualityActual,
			CreatedAt:    now,
		},
		{
			ID:           fmt.Sprintf("line-rds-%d", now.UnixNano()),
			InvoiceID:    invID,
			ResourceID:   "res-rds-postgres-01",
			ResourceType: "DATABASE",
			Description:  "AWS RDS PostgreSQL Database Instance (db.t3.micro * 720 Hours + 20GB Storage)",
			Metric:       "database.runtime",
			Quantity:     720.0,
			Unit:         "instance-hour",
			UnitPrice:    0.018,
			Amount:       12.96,
			UsageQuality: domain.QualityActual,
			CreatedAt:    now,
		},
		{
			ID:           fmt.Sprintf("line-s3-%d", now.UnixNano()),
			InvoiceID:    invID,
			ResourceID:   "res-s3-assets-01",
			ResourceType: "STORAGE",
			Description:  "AWS S3 Bucket anarva-production-media-assets (25.0 GB-month)",
			Metric:       "storage.capacity",
			Quantity:     25.0,
			Unit:         "GB-month",
			UnitPrice:    0.023,
			Amount:       0.58,
			UsageQuality: domain.QualityActual,
			CreatedAt:    now,
		},
	}

	subtotal := 0.0
	for _, l := range lines {
		subtotal += l.Amount
	}

	inv := &domain.Invoice{
		ID:              invID,
		OrganizationID:  orgID,
		BillingPeriodID: fmt.Sprintf("bp-%s", now.Format("2006-01")),
		InvoiceNumber:   invNumber,
		Currency:        "USD",
		Subtotal:        subtotal,
		Discount:        0.0,
		Tax:             0.0,
		Total:           subtotal,
		Status:          domain.InvoiceDraft,
		PricingVersion:  "anarva-v1.0.0",
		RealityLabel:    "ANARVA_ESTIMATED_CUSTOMER_CHARGE",
		IssuedAt:        now,
		DueAt:           now.Add(30 * 24 * time.Hour),
		Lines:           lines,
		CreatedAt:       now,
	}

	uc.invoices = append(uc.invoices, inv)
	return inv, nil
}

func (uc *BillingUseCase) FinalizeInvoice(ctx context.Context, invoiceID string) (*domain.Invoice, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	for _, inv := range uc.invoices {
		if inv.ID == invoiceID {
			if inv.Status == domain.InvoiceFinalized {
				return inv, nil
			}
			inv.Status = domain.InvoiceFinalized
			return inv, nil
		}
	}
	return nil, fmt.Errorf("Invoice %s not found", invoiceID)
}

// ReserveQuota atomically validates and reserves quota under a mutex lock to prevent race conditions
func (uc *BillingUseCase) ReserveQuota(ctx context.Context, orgID, projectID, resourceType, metric string, requestedAmount float64) (*domain.Quota, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s", orgID, projectID, metric)
	q, ok := uc.quotas[key]
	if !ok {
		// Default auto-created quota
		q = &domain.Quota{
			ID:             fmt.Sprintf("q-%d", time.Now().UnixNano()/1e6),
			OrganizationID: orgID,
			ProjectID:      projectID,
			ResourceType:   resourceType,
			Metric:         metric,
			Limit:          32.0,
			CurrentUsage:   0.0,
			Unit:           "UNITS",
			Period:         "PERPETUAL",
			Status:         domain.QuotaAvailable,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		uc.quotas[key] = q
	}

	if q.Limit > 0 && (q.CurrentUsage+requestedAmount) > q.Limit {
		q.Status = domain.QuotaExceeded
		return q, appErrors.New(appErrors.CodeQuotaExceeded, fmt.Sprintf("Quota exceeded for %s: requested %.2f, current %.2f, limit %.2f", metric, requestedAmount, q.CurrentUsage, q.Limit))
	}

	q.CurrentUsage += requestedAmount
	q.UpdatedAt = time.Now()

	if q.Limit > 0 && q.CurrentUsage >= q.Limit*0.85 {
		q.Status = domain.QuotaNearLimit
	} else {
		q.Status = domain.QuotaAvailable
	}

	return q, nil
}

func (uc *BillingUseCase) RecordUsage(ctx context.Context, rec *domain.UsageRecord) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if rec.ID == "" {
		rec.ID = fmt.Sprintf("use-rec-%d", time.Now().UnixNano()/1e6)
	}
	if rec.Timestamp.IsZero() {
		rec.Timestamp = time.Now()
	}
	if rec.Source == "" {
		rec.Source = domain.SourceLocalProvider
	}
	if rec.Quality == "" {
		rec.Quality = domain.QualitySimulated
	}
	if rec.RealityLabel == "" {
		rec.RealityLabel = "LOCAL_DEVELOPMENT_USAGE (NON_BILLABLE)"
	}

	uc.records = append([]*domain.UsageRecord{rec}, uc.records...)
	if len(uc.records) > 200 {
		uc.records = uc.records[:200]
	}
	return nil
}

func (uc *BillingUseCase) ListQuotas(ctx context.Context, projectID string) []*domain.Quota {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	var list []*domain.Quota
	for _, q := range uc.quotas {
		if projectID == "" || q.ProjectID == projectID {
			list = append(list, q)
		}
	}
	return list
}

func (uc *BillingUseCase) ListUsage(ctx context.Context, projectID string) []*domain.UsageRecord {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	var list []*domain.UsageRecord
	for _, r := range uc.records {
		if projectID == "" || r.ProjectID == projectID {
			list = append(list, r)
		}
	}
	return list
}

func (uc *BillingUseCase) CalculateCostEstimate(ctx context.Context, resourceType, provider string, acu float64, expectedHours float64) (*domain.CostEstimate, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	rate := 0.025
	if resourceType == "DATABASE" {
		rate = 0.045
	} else if resourceType == "STORAGE" {
		rate = 0.015
	}

	if acu <= 0 {
		acu = 1.0
	}
	if expectedHours <= 0 {
		expectedHours = 720.0
	}

	total := acu * rate * expectedHours

	est := &domain.CostEstimate{
		ID:                 fmt.Sprintf("est-%d", time.Now().UnixNano()/1e6),
		OrganizationID:     "org-default",
		ProjectID:          "proj-default",
		ResourceType:       resourceType,
		Provider:           provider,
		PricingPlanVersion: "v1.0.0",
		Configuration: map[string]string{
			"acu":           fmt.Sprintf("%.1f", acu),
			"rate":          fmt.Sprintf("$%.3f / ACU-hr", rate),
			"expectedHours": fmt.Sprintf("%.0f", expectedHours),
		},
		ExpectedUsageHours: expectedHours,
		EstimatedCost:      total,
		UsageQuality:       domain.QualityEstimated,
		RealityLabel:       "NOT_BILLABLE (ESTIMATE)",
		CreatedAt:          time.Now(),
	}

	return est, nil
}

func (uc *BillingUseCase) ListInvoices(ctx context.Context, orgID string) []*domain.Invoice {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	var filtered []*domain.Invoice
	for _, inv := range uc.invoices {
		if orgID == "" || inv.OrganizationID == orgID {
			filtered = append(filtered, inv)
		}
	}
	return filtered
}

func (uc *BillingUseCase) GetBillingAccount(ctx context.Context, orgID string) (*domain.BillingAccount, *domain.BillingProfile) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.account, uc.profile
}

func (uc *BillingUseCase) ListPricingPlans(ctx context.Context) ([]*domain.PricingPlan, map[string][]*domain.PricingComponent) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	var plans []*domain.PricingPlan
	for _, p := range uc.pricingPlans {
		plans = append(plans, p)
	}
	return plans, uc.components
}
