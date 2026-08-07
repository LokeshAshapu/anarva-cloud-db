package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationMember struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	OrgID     string    `gorm:"index;not null;type:uuid"`
	UserID    string    `gorm:"index;not null;type:uuid"`
	Role      string    `gorm:"not null;type:varchar(50);default:'DEVELOPER'"`
	JoinedAt  time.Time `gorm:"autoCreateTime"`
}

func NewOrganizationMember(orgID, userID, role string) *OrganizationMember {
	if role == "" {
		role = "DEVELOPER"
	}
	return &OrganizationMember{
		ID:       uuid.New().String(),
		OrgID:    orgID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}
}
