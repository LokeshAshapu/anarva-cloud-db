package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
	Revoke(ctx context.Context, sessionID string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *APIKey) error
	GetByID(ctx context.Context, id string) (*APIKey, error)
	GetByHashedKey(ctx context.Context, hashedKey string) (*APIKey, error)
	ListByUserID(ctx context.Context, userID string) ([]*APIKey, error)
	Revoke(ctx context.Context, id string) error
	Update(ctx context.Context, apiKey *APIKey) error
}

type VerificationTokenRepository interface {
	Create(ctx context.Context, token *VerificationToken) error
	GetByToken(ctx context.Context, token string) (*VerificationToken, error)
	Delete(ctx context.Context, token string) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*AuditLog, error)
}
