package domain

import (
	"time"
)

type BillingAccountStatus string

const (
	BillingStatusActive        BillingAccountStatus = "ACTIVE"
	BillingStatusSuspended     BillingAccountStatus = "SUSPENDED"
	BillingStatusClosed        BillingAccountStatus = "CLOSED"
	BillingStatusNotConfigured BillingAccountStatus = "NOT_CONFIGURED"
)

type UsageQuality string

const (
	QualityActual    UsageQuality = "ACTUAL"
	QualityEstimated UsageQuality = "ESTIMATED"
	QualitySimulated UsageQuality = "SIMULATED"
	QualityUnknown   UsageQuality = "UNKNOWN"
)

type UsageSource string

const (
	SourceLocalProvider   UsageSource = "LOCAL_PROVIDER"
	SourceCloudProvider   UsageSource = "CLOUD_PROVIDER"
	SourceDatabaseProvider UsageSource = "DATABASE_PROVIDER"
	SourceStorageProvider UsageSource = "STORAGE_PROVIDER"
	SourceControlPlane    UsageSource = "CONTROL_PLANE"
	SourceEstimated       UsageSource = "ESTIMATED"
	SourceSimulated       UsageSource = "SIMULATED"
)

type QuotaStatus string

const (
	QuotaAvailable QuotaStatus = "AVAILABLE"
	QuotaNearLimit QuotaStatus = "NEAR_LIMIT"
	QuotaExceeded  QuotaStatus = "EXCEEDED"
	QuotaUnlimited QuotaStatus = "UNLIMITED"
)

type InvoiceStatus string

const (
	InvoiceDraft     InvoiceStatus = "DRAFT"
	InvoiceFinalized InvoiceStatus = "FINALIZED"
	InvoiceVoid      InvoiceStatus = "VOID"
	InvoicePaid      InvoiceStatus = "PAID"
	InvoiceUnpaid    InvoiceStatus = "UNPAID"
)

type BillingAccount struct {
	ID                 string               `json:"id"`
	OrganizationID     string               `json:"organizationId"`
	Status             BillingAccountStatus `json:"status"`
	Currency           string               `json:"currency"`
	BillingEmail       string               `json:"billingEmail"`
	BillingName        string               `json:"billingName"`
	TaxConfiguration   map[string]string    `json:"taxConfiguration,omitempty"`
	Provider           string               `json:"provider"`
	ProviderCustomerID string               `json:"providerCustomerId,omitempty"`
	CreatedAt          time.Time            `json:"createdAt"`
	UpdatedAt          time.Time            `json:"updatedAt"`
}

type BillingProfile struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	LegalName      string    `json:"legalName"`
	BillingEmail   string    `json:"billingEmail"`
	Address        string    `json:"address"`
	Country        string    `json:"country"`
	TaxID          string    `json:"taxId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type UsageRecord struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organizationId"`
	ProjectID      string            `json:"projectId"`
	ResourceID     string            `json:"resourceId"`
	ResourceType   string            `json:"resourceType"` // COMPUTE, DATABASE, STORAGE, NETWORK, BACKUP
	Provider       string            `json:"provider"`
	Metric         string            `json:"metric"`       // compute.runtime, storage.capacity, etc.
	Quantity       float64           `json:"quantity"`
	Unit           string            `json:"unit"`         // ACU-hour, GB-month, GB
	Source         UsageSource       `json:"source"`
	Quality        UsageQuality      `json:"quality"`
	RealityLabel   string            `json:"realityLabel"` // LOCAL_DEVELOPMENT_USAGE, NON_BILLABLE
	Timestamp      time.Time         `json:"timestamp"`
	PeriodStart    time.Time         `json:"periodStart"`
	PeriodEnd      time.Time         `json:"periodEnd"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
}

type Quota struct {
	ID             string      `json:"id"`
	OrganizationID string      `json:"organizationId"`
	ProjectID      string      `json:"projectId"`
	ResourceType   string      `json:"resourceType"`
	Metric         string      `json:"metric"`
	Limit          float64     `json:"limit"`
	CurrentUsage   float64     `json:"currentUsage"`
	Unit           string      `json:"unit"`
	Period         string      `json:"period"` // MONTHLY, HOURLY, PERPETUAL
	Status         QuotaStatus `json:"status"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

type PricingPlan struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"` // ACTIVE, DEPRECATED
	Version       string    `json:"version"` // v1.0.0, v2.0.0
	EffectiveFrom time.Time `json:"effectiveFrom"`
	EffectiveTo   *time.Time `json:"effectiveTo,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type PricingComponent struct {
	ID                string    `json:"id"`
	PricingPlanID     string    `json:"pricingPlanId"`
	ResourceType      string    `json:"resourceType"`
	Metric            string    `json:"metric"`
	Unit              string    `json:"unit"`
	UnitPrice         float64   `json:"unitPrice"`
	MinimumCharge     float64   `json:"minimumCharge"`
	IncludedQuantity  float64   `json:"includedQuantity"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type InvoiceLine struct {
	ID           string       `json:"id"`
	InvoiceID    string       `json:"invoiceId"`
	ResourceID   string       `json:"resourceId"`
	ResourceType string       `json:"resourceType"`
	Description  string       `json:"description"`
	Metric       string       `json:"metric"`
	Quantity     float64      `json:"quantity"`
	Unit         string       `json:"unit"`
	UnitPrice    float64      `json:"unitPrice"`
	Amount       float64      `json:"amount"`
	UsageQuality UsageQuality `json:"usageQuality"`
	CreatedAt    time.Time    `json:"createdAt"`
}

type Invoice struct {
	ID              string        `json:"id"`
	OrganizationID  string        `json:"organizationId"`
	BillingPeriodID string        `json:"billingPeriodId"`
	InvoiceNumber   string        `json:"invoiceNumber"`
	Currency        string        `json:"currency"`
	Subtotal        float64       `json:"subtotal"`
	Discount        float64       `json:"discount"`
	Tax             float64       `json:"tax"`
	Total           float64       `json:"total"`
	Status          InvoiceStatus `json:"status"`
	PricingVersion  string        `json:"pricingVersion"`
	RealityLabel    string        `json:"realityLabel"` // ESTIMATED_BILLING, SIMULATED
	IssuedAt        time.Time     `json:"issuedAt"`
	DueAt           time.Time     `json:"dueAt"`
	Lines           []InvoiceLine `json:"lines,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
}

type CostEstimate struct {
	ID                 string            `json:"id"`
	OrganizationID     string            `json:"organizationId"`
	ProjectID          string            `json:"projectId"`
	ResourceType       string            `json:"resourceType"`
	Provider           string            `json:"provider"`
	PricingPlanVersion string            `json:"pricingPlanVersion"`
	Configuration      map[string]string `json:"configuration"`
	ExpectedUsageHours float64           `json:"expectedUsageHours"`
	EstimatedCost      float64           `json:"estimatedCost"`
	UsageQuality       UsageQuality      `json:"usageQuality"`
	RealityLabel       string            `json:"realityLabel"` // NOT_BILLABLE, ESTIMATE
	CreatedAt          time.Time         `json:"createdAt"`
}
