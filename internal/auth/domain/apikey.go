package domain

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         string     `gorm:"primaryKey;type:uuid"`
	UserID     string     `gorm:"index;not null;type:uuid"`
	Name       string     `gorm:"not null;type:varchar(255)"`
	Prefix     string     `gorm:"not null;type:varchar(50)"`
	HashedKey  string     `gorm:"uniqueIndex;not null;type:varchar(255)"`
	IsRevoked  bool       `gorm:"not null;default:false"`
	ExpiresAt  *time.Time `gorm:"type:timestamp"`
	LastUsedAt *time.Time `gorm:"type:timestamp"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
}

func NewAPIKey(userID, name, prefix, hashedKey string, expiresAt *time.Time) *APIKey {
	return &APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		Prefix:    prefix,
		HashedKey: hashedKey,
		IsRevoked: false,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}

func (k *APIKey) Revoke() {
	k.IsRevoked = true
}

func (k *APIKey) MarkUsed() {
	now := time.Now()
	k.LastUsedAt = &now
}
