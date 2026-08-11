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
	RoleOwner         RoleType = "OWNER"
	RoleAdmin         RoleType = "ADMIN"
	RoleDeveloper     RoleType = "DEVELOPER"
	RoleDatabaseAdmin RoleType = "DATABASE_ADMIN"
	RoleStorageAdmin  RoleType = "STORAGE_ADMIN"
	RoleViewer        RoleType = "VIEWER"
	RoleBillingAdmin  RoleType = "BILLING_ADMIN"
	RoleSecurityAdmin RoleType = "SECURITY_ADMIN"
)

type Permission string

const (
	PermOrgRead      Permission = "organization.read"
	PermOrgUpdate    Permission = "organization.update"
	PermOrgDelete    Permission = "organization.delete"
	PermProjRead     Permission = "project.read"
	PermProjCreate   Permission = "project.create"
	PermProjUpdate   Permission = "project.update"
	PermProjDelete   Permission = "project.delete"
	PermDBRead       Permission = "database.read"
	PermDBCreate     Permission = "database.create"
	PermDBUpdate     Permission = "database.update"
	PermDBDelete     Permission = "database.delete"
	PermDBQuery      Permission = "database.query"
	PermStorageRead  Permission = "storage.read"
	PermStorageCreate Permission = "storage.create"
	PermStorageUpload Permission = "storage.upload"
	PermStorageDelete Permission = "storage.delete"
	PermIAMRead      Permission = "iam.read"
	PermIAMManage    Permission = "iam.manage"
	PermSecurityManage Permission = "security.manage"
)

type APIKey struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId"`
	ProjectID      string     `json:"projectId"`
	Name           string     `json:"name"`
	KeyPrefix      string     `json:"keyPrefix"`
	HashedSecret   string     `json:"-"` // Never exposed in JSON
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
	Status         string    `json:"status"` // ACTIVE, DISABLED
	Role           RoleType  `json:"role"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Team struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
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
	Status         string    `json:"status"` // ACTIVE, INVITED, SUSPENDED
	JoinedAt       time.Time `json:"joinedAt"`
}

// HashSecret computes SHA-256 hash of API secret
func HashSecret(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(h[:])
}

// GenerateRawAPIKey creates key prefix and full secret
func GenerateRawAPIKey(orgID, name string) (prefix, fullSecret, hash string) {
	rawToken := fmt.Sprintf("ak_%d_%s", time.Now().UnixNano(), strings.ReplaceAll(name, " ", "_"))
	prefix = rawToken[:8]
	fullSecret = fmt.Sprintf("anarva_live_%s", rawToken)
	hash = HashSecret(fullSecret)
	return prefix, fullSecret, hash
}
