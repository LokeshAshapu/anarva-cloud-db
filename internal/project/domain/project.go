package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID              string    `gorm:"primaryKey;type:uuid"`
	OrgID           string    `gorm:"index;not null;type:uuid"`
	Name            string    `gorm:"not null;type:varchar(255)"`
	Slug            string    `gorm:"uniqueIndex;not null;type:varchar(255)"`
	Region          string    `gorm:"not null;type:varchar(100);default:'us-east-1'"`
	DatabaseCount   int       `gorm:"not null;default:0"`
	MaxDatabases    int       `gorm:"not null;default:5"`
	MaxStorageBytes int64     `gorm:"not null;default:10737418240"` // 10 GB
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func NewProject(orgID, name, slug, region string) *Project {
	if region == "" {
		region = "us-east-1"
	}
	return &Project{
		ID:              uuid.New().String(),
		OrgID:           orgID,
		Name:            name,
		Slug:            slug,
		Region:          region,
		DatabaseCount:   0,
		MaxDatabases:    5,
		MaxStorageBytes: 10 * 1024 * 1024 * 1024,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func (p *Project) CanCreateDatabase() bool {
	return p.DatabaseCount < p.MaxDatabases
}

func (p *Project) IncrementDatabaseCount() {
	p.DatabaseCount++
	p.UpdatedAt = time.Now()
}

func (p *Project) DecrementDatabaseCount() {
	if p.DatabaseCount > 0 {
		p.DatabaseCount--
		p.UpdatedAt = time.Now()
	}
}
