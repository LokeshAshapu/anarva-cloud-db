package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleOwner     Role = "OWNER"
	RoleAdmin     Role = "ADMIN"
	RoleDeveloper Role = "DEVELOPER"
	RoleViewer    Role = "VIEWER"
)

type UserStatus string

const (
	UserStatusPending   UserStatus = "PENDING"
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

type User struct {
	ID            string     `gorm:"primaryKey;type:uuid"`
	Email         string     `gorm:"uniqueIndex;not null;type:varchar(255)"`
	PasswordHash  string     `gorm:"not null;type:varchar(255)"`
	FullName      string     `gorm:"not null;type:varchar(255)"`
	Role          Role       `gorm:"not null;type:varchar(50);default:'DEVELOPER'"`
	Status        UserStatus `gorm:"not null;type:varchar(50);default:'PENDING'"`
	EmailVerified bool       `gorm:"not null;default:false"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime"`
}

func NewUser(email, passwordHash, fullName string) *User {
	return &User{
		ID:            uuid.New().String(),
		Email:         email,
		PasswordHash:  passwordHash,
		FullName:      fullName,
		Role:          RoleDeveloper,
		Status:        UserStatusPending,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (u *User) Activate() {
	u.Status = UserStatusActive
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
}
