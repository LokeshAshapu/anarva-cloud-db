package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type APIKeyStatus string

const (
	KeyStatusActive    APIKeyStatus = "ACTIVE"
	KeyStatusRevoked   APIKeyStatus = "REVOKED"
	KeyStatusExpired   APIKeyStatus = "EXPIRED"
	KeyStatusSuspended APIKeyStatus = "SUSPENDED"
)

type APIKeyEnvironment string

const (
	EnvLive APIKeyEnvironment = "LIVE"
	EnvTest APIKeyEnvironment = "TEST"
)

type APIKey struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organizationId"`
	ProjectID      string            `json:"projectId"`
	Name           string            `json:"name"`
	KeyPrefix      string            `json:"keyPrefix"`
	KeyHash        string            `json:"-"` // Never exposed in JSON
	Environment    APIKeyEnvironment `json:"environment"`
	Status         APIKeyStatus      `json:"status"`
	Permissions    []string          `json:"permissions"`
	CreatedBy      string            `json:"createdBy"`
	ExpiresAt      *time.Time        `json:"expiresAt,omitempty"`
	LastUsedAt     *time.Time        `json:"lastUsedAt,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	RevokedAt      *time.Time        `json:"revokedAt,omitempty"`
}

type ServiceAccount struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	ProjectID      string    `json:"projectId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	Role           string    `json:"role"`
	CreatedBy      string    `json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type APIUsageRecord struct {
	ID               string    `json:"id"`
	ApiKeyID         string    `json:"apiKeyId,omitempty"`
	ServiceAccountID string    `json:"serviceAccountId,omitempty"`
	OrganizationID   string    `json:"organizationId"`
	ProjectID        string    `json:"projectId"`
	Endpoint         string    `json:"endpoint"`
	Method           string    `json:"method"`
	StatusCode       int       `json:"statusCode"`
	ResponseTimeMs   float64   `json:"responseTimeMs"`
	RequestID        string    `json:"requestId"`
	Timestamp        time.Time `json:"timestamp"`
}

// GenerateAPIKey creates a secure random API key with anarva_live_ / anarva_test_ prefix
func GenerateAPIKey(isLive bool) (secretKey string, keyHash string, keyPrefix string, keyID string) {
	bytes := make([]byte, 24)
	_, _ = rand.Read(bytes)
	randomHex := hex.EncodeToString(bytes)

	prefix := "anarva_test_"
	if isLive {
		prefix = "anarva_live_"
	}

	secretKey = fmt.Sprintf("%s%s", prefix, randomHex)
	keyPrefix = fmt.Sprintf("%s%s...", prefix, randomHex[:6])
	keyID = fmt.Sprintf("key_%d", time.Now().UnixNano()/1e6)

	hashBytes := sha256.Sum256([]byte(secretKey))
	keyHash = hex.EncodeToString(hashBytes[:])

	return secretKey, keyHash, keyPrefix, keyID
}

func HashSecret(secret string) string {
	hashBytes := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hashBytes[:])
}
