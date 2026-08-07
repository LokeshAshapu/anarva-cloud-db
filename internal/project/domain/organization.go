package domain

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	OwnerID   string    `gorm:"index;not null;type:uuid"`
	Name      string    `gorm:"not null;type:varchar(255)"`
	Slug      string    `gorm:"uniqueIndex;not null;type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func NewOrganization(ownerID, name, slug string) *Organization {
	return &Organization{
		ID:        uuid.New().String(),
		OwnerID:   ownerID,
		Name:      name,
		Slug:      slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
