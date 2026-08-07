package domain

import (
	"time"

	"github.com/google/uuid"
)

type VerificationToken struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	UserID    string    `gorm:"index;not null;type:uuid"`
	Token     string    `gorm:"uniqueIndex;not null;type:varchar(255)"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func NewVerificationToken(userID, token string, expiryDuration time.Duration) *VerificationToken {
	return &VerificationToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(expiryDuration),
		CreatedAt: time.Now(),
	}
}

func (v *VerificationToken) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}
