package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           string    `gorm:"primaryKey;type:uuid"`
	UserID       string    `gorm:"index;not null;type:uuid"`
	RefreshToken string    `gorm:"uniqueIndex;not null;type:text"`
	UserAgent    string    `gorm:"type:varchar(512)"`
	IPAddress    string    `gorm:"type:varchar(100)"`
	IsRevoked    bool      `gorm:"not null;default:false"`
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func NewSession(userID, refreshToken, userAgent, ipAddress string, expiresAt time.Time) *Session {
	return &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		IsRevoked:    false,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
}

func (s *Session) Revoke() {
	s.IsRevoked = true
}
