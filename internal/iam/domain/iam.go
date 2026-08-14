package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type RoleType string

const (
	RoleOwner        RoleType = "OWNER"
	RoleAdmin        RoleType = "ADMIN"
	RoleDeveloper    RoleType = "DEVELOPER"
	RoleViewer       RoleType = "VIEWER"
	RoleBillingAdmin RoleType = "BILLING_ADMIN"
	RoleAuditor      RoleType = "AUDITOR"
)

type Permission string

const (
	PermOrgRead             Permission = "organization.read"
	PermOrgUpdate           Permission = "organization.update"
	PermOrgDelete           Permission = "organization.delete"
	PermOrgMembersRead      Permission = "organization.members.read"
	PermOrgMembersInvite    Permission = "organization.members.invite"
	PermOrgMembersRemove    Permission = "organization.members.remove"
	PermProjCreate          Permission = "project.create"
	PermProjRead            Permission = "project.read"
	PermProjUpdate          Permission = "project.update"
	PermProjDelete          Permission = "project.delete"
	PermProjMembersManage   Permission = "project.members.manage"
	PermComputeRead         Permission = "compute.read"
	PermComputeCreate       Permission = "compute.create"
	PermComputeUpdate       Permission = "compute.update"
	PermComputeDelete       Permission = "compute.delete"
	PermDatabaseRead        Permission = "database.read"
	PermDatabaseCreate      Permission = "database.create"
	PermDatabaseUpdate      Permission = "database.update"
	PermDatabaseDelete      Permission = "database.delete"
	PermStorageRead         Permission = "storage.read"
	PermStorageCreate       Permission = "storage.create"
	PermStorageUpload       Permission = "storage.upload"
	PermStorageDelete       Permission = "storage.delete"
	PermMetricsRead         Permission = "metrics.read"
	PermBillingRead         Permission = "billing.read"
	PermBillingManage       Permission = "billing.manage"
	PermBillingPricingManage Permission = "billing.pricing.manage"
	PermAuditRead           Permission = "audit.read"
	PermIAMManage           Permission = "iam.manage"
)

type Organization struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	OwnerUserID string    `json:"ownerUserId"`
	Status      string    `json:"status"` // ACTIVE, SUSPENDED
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Status         string    `json:"status"` // ACTIVE, ARCHIVED
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type OrgMember struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	UserID         string    `json:"userId"`
	UserEmail      string    `json:"userEmail"`
	UserName       string    `json:"userName"`
	Role           RoleType  `json:"role"`
	Status         string    `json:"status"` // ACTIVE, INVITED, SUSPENDED, REMOVED
	JoinedAt       time.Time `json:"joinedAt"`
}

type Invitation struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId"`
	Email          string     `json:"email"`
	Role           RoleType   `json:"role"`
	TokenHash      string     `json:"-"` // Never expose in JSON
	Status         string     `json:"status"` // PENDING, ACCEPTED, EXPIRED, REVOKED
	ExpiresAt      time.Time  `json:"expiresAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
}

type PolicyRule struct {
	Role        RoleType     `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// NormalizeSlug converts name to URL-safe slug
func NormalizeSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}

// HashInvitationToken computes SHA-256 hash of single-use invitation token
func HashInvitationToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Legacy APIKey stub for Phase 36 compatibility
type APIKey struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId"`
	ProjectID      string     `json:"projectId"`
	Name           string     `json:"name"`
	KeyPrefix      string     `json:"keyPrefix"`
	HashedSecret   string     `json:"-"`
	Permissions    []string   `json:"permissions"`
	LastUsedAt     *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
}

type ServiceAccount struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	Role           RoleType  `json:"role"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func HashSecret(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(h[:])
}

func GenerateRawAPIKey(orgID, name string) (string, string, string) {
	prefix := "anarva_" + strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()%10000))
	secret := fmt.Sprintf("ak_secret_%d", time.Now().UnixNano())
	hash := HashSecret(secret)
	return prefix, secret, hash
}
