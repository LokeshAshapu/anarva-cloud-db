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
	KeyStatusActive  APIKeyStatus = "ACTIVE"
	KeyStatusRevoked APIKeyStatus = "REVOKED"
	KeyStatusExpired APIKeyStatus = "EXPIRED"
)

type APIKey struct {
	ID             string       `json:"id"`
	OrganizationID string       `json:"organizationId"`
	ProjectID      string       `json:"projectId"`
	Name           string       `json:"name"`
	KeyPrefix      string       `json:"keyPrefix"`
	KeyHash        string       `json:"-"`
	Status         APIKeyStatus `json:"status"`
	Permissions    []string     `json:"permissions"`
	CreatedBy      string       `json:"createdBy"`
	ExpiresAt      *time.Time   `json:"expiresAt,omitempty"`
	LastUsedAt     *time.Time   `json:"lastUsedAt,omitempty"`
	CreatedAt      time.Time    `json:"createdAt"`
	RevokedAt      *time.Time   `json:"revokedAt,omitempty"`
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
	Endpoint         string    `json:"endpoint"`
	Method           string    `json:"method"`
	StatusCode       int       `json:"statusCode"`
	ResponseTimeMs   float64   `json:"responseTimeMs"`
	RequestID        string    `json:"requestId"`
	Timestamp        time.Time `json:"timestamp"`
}

// GenerateAPIKey creates a secure random API key with a prefix (e.g., ank_live_...) and returns plaintext secret + hash
func GenerateAPIKey(isLive bool) (secretKey string, keyHash string, keyPrefix string) {
	bytes := make([]byte, 24)
	_, _ = rand.Read(bytes)
	randomHex := hex.EncodeToString(bytes)

	prefix := "ank_test_"
	if isLive {
		prefix = "ank_live_"
	}

	secretKey = fmt.Sprintf("%s%s", prefix, randomHex)
	keyPrefix = fmt.Sprintf("%s%s...", prefix, randomHex[:6])

	hashBytes := sha256.Sum256([]byte(secretKey))
	keyHash = hex.EncodeToString(hashBytes[:])

	return secretKey, keyHash, keyPrefix
}

func HashSecret(secret string) string {
	hashBytes := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hashBytes[:])
}
