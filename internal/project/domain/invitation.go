package domain

import (
	"time"

	"github.com/google/uuid"
)

type Invitation struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	OrgID     string    `gorm:"index;not null;type:uuid"`
	Email     string    `gorm:"not null;type:varchar(255)"`
	Role      string    `gorm:"not null;type:varchar(50)"`
	Token     string    `gorm:"uniqueIndex;not null;type:varchar(255)"`
	InvitedBy string    `gorm:"not null;type:uuid"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func NewInvitation(orgID, email, role, invitedBy string, expiryDuration time.Duration) *Invitation {
	return &Invitation{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		Email:     email,
		Role:      role,
		Token:     uuid.New().String(),
		InvitedBy: invitedBy,
		ExpiresAt: time.Now().Add(expiryDuration),
		CreatedAt: time.Now(),
	}
}

func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}
