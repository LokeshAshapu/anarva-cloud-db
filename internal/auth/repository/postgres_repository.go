package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return appErrors.New(appErrors.CodeAlreadyExists, "user with this email already exists")
		}
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to create user")
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "user not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch user by id")
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "user not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch user by email")
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update user")
	}
	return nil
}

// Session repository
type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) domain.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to create session")
	}
	return nil
}

func (r *sessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.WithContext(ctx).First(&session, "refresh_token = ?", refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "session not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch session")
	}
	return &session, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, sessionID string) error {
	if err := r.db.WithContext(ctx).Model(&domain.Session{}).Where("id = ?", sessionID).Update("is_revoked", true).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to revoke session")
	}
	return nil
}

func (r *sessionRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Model(&domain.Session{}).Where("user_id = ?", userID).Update("is_revoked", true).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to revoke user sessions")
	}
	return nil
}

// APIKey repository
type apiKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) domain.APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *domain.APIKey) error {
	if err := r.db.WithContext(ctx).Create(apiKey).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to create API key")
	}
	return nil
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	var key domain.APIKey
	if err := r.db.WithContext(ctx).First(&key, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "API key not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch API key")
	}
	return &key, nil
}

func (r *apiKeyRepository) GetByHashedKey(ctx context.Context, hashedKey string) (*domain.APIKey, error) {
	var key domain.APIKey
	if err := r.db.WithContext(ctx).First(&key, "hashed_key = ?", hashedKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "API key not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch API key by hash")
	}
	return &key, nil
}

func (r *apiKeyRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	var keys []*domain.APIKey
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&keys).Error; err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list API keys")
	}
	return keys, nil
}

func (r *apiKeyRepository) Revoke(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Model(&domain.APIKey{}).Where("id = ?", id).Update("is_revoked", true).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to revoke API key")
	}
	return nil
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *domain.APIKey) error {
	if err := r.db.WithContext(ctx).Save(apiKey).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update API key")
	}
	return nil
}

// VerificationToken repository
type verificationTokenRepository struct {
	db *gorm.DB
}

func NewVerificationTokenRepository(db *gorm.DB) domain.VerificationTokenRepository {
	return &verificationTokenRepository{db: db}
}

func (r *verificationTokenRepository) Create(ctx context.Context, token *domain.VerificationToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to create verification token")
	}
	return nil
}

func (r *verificationTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.VerificationToken, error) {
	var token domain.VerificationToken
	if err := r.db.WithContext(ctx).First(&token, "token = ?", tokenStr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.CodeNotFound, "verification token not found")
		}
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to fetch verification token")
	}
	return &token, nil
}

func (r *verificationTokenRepository) Delete(ctx context.Context, tokenStr string) error {
	if err := r.db.WithContext(ctx).Where("token = ?", tokenStr).Delete(&domain.VerificationToken{}).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to delete verification token")
	}
	return nil
}

// AuditLog repository
type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to create audit log")
	}
	return nil
}

func (r *auditLogRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list audit logs")
	}
	return logs, nil
}
