package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	UserID    string    `gorm:"index;type:uuid"`
	Action    string    `gorm:"not null;type:varchar(100)"`
	IPAddress string    `gorm:"type:varchar(100)"`
	UserAgent string    `gorm:"type:varchar(512)"`
	Status    string    `gorm:"not null;type:varchar(50)"` // SUCCESS or FAILED
	Details   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func NewAuditLog(userID, action, ipAddress, userAgent, status, details string) *AuditLog {
	return &AuditLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Status:    status,
		Details:   details,
		CreatedAt: time.Now(),
	}
}
