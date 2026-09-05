package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/postgres/domain"
)

type PostgresUserRepository interface {
	Create(ctx context.Context, user *domain.PostgresUser) error
	GetByInstanceAndUsername(ctx context.Context, instanceID, username string) (*domain.PostgresUser, error)
	Delete(ctx context.Context, instanceID, username string) error
	Update(ctx context.Context, user *domain.PostgresUser) error
}

type GormPostgresUserRepository struct {
	db *gorm.DB
}

func NewGormPostgresUserRepository(db *gorm.DB) PostgresUserRepository {
	return &GormPostgresUserRepository{db: db}
}

func (r *GormPostgresUserRepository) Create(ctx context.Context, user *domain.PostgresUser) error {
	if user == nil || user.ID == "" {
		return errors.New("invalid postgres user: ID is required")
	}

	now := time.Now()

	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}

	user.UpdatedAt = now
	
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormPostgresUserRepository) GetByInstanceAndUsername(
	ctx context.Context,
	instanceID, username string,
) (*domain.PostgresUser, error) {
	var user domain.PostgresUser

	err := r.db.WithContext(ctx).
		Where("instance_id = ? AND username = ?", instanceID, username).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(
				"postgres user '%s' for instance '%s' not found",
				username,
				instanceID,
			)
		}

		return nil, err
	}

	return &user, nil
}

func (r *GormPostgresUserRepository) Delete(
	ctx context.Context,
	instanceID, username string,
) error {
	result := r.db.WithContext(ctx).
		Where("instance_id = ? AND username = ?", instanceID, username).
		Delete(&domain.PostgresUser{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf(
			"postgres user '%s' for instance '%s' not found",
			username,
			instanceID,
		)
	}

	return nil
}

func (r *GormPostgresUserRepository) Update(
	ctx context.Context,
	user *domain.PostgresUser,
) error {
	if user == nil || user.ID == "" {
		return errors.New("invalid postgres user: ID is required")
	}

	user.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).Save(user).Error
}
