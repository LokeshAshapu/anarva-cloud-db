package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

// In-Memory Repository Implementations for Testing
type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() domain.UserRepository {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.users[user.Email] = user
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "user not found")
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "user not found")
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	m.users[user.Email] = user
	m.users[user.ID] = user
	return nil
}

type mockSessionRepo struct {
	sessions map[string]*domain.Session
}

func newMockSessionRepo() domain.SessionRepository {
	return &mockSessionRepo{sessions: make(map[string]*domain.Session)}
}

func (m *mockSessionRepo) Create(ctx context.Context, session *domain.Session) error {
	m.sessions[session.RefreshToken] = session
	return nil
}

func (m *mockSessionRepo) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	if s, ok := m.sessions[token]; ok {
		return s, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "session not found")
}

func (m *mockSessionRepo) Revoke(ctx context.Context, sessionID string) error {
	for _, s := range m.sessions {
		if s.ID == sessionID {
			s.Revoke()
		}
	}
	return nil
}

func (m *mockSessionRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	for _, s := range m.sessions {
		if s.UserID == userID {
			s.Revoke()
		}
	}
	return nil
}

type mockAPIKeyRepo struct {
	keys map[string]*domain.APIKey
}

func newMockAPIKeyRepo() domain.APIKeyRepository {
	return &mockAPIKeyRepo{keys: make(map[string]*domain.APIKey)}
}

func (m *mockAPIKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	m.keys[key.ID] = key
	return nil
}

func (m *mockAPIKeyRepo) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	if k, ok := m.keys[id]; ok {
		return k, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "API key not found")
}

func (m *mockAPIKeyRepo) GetByHashedKey(ctx context.Context, hashedKey string) (*domain.APIKey, error) {
	for _, k := range m.keys {
		if k.HashedKey == hashedKey {
			return k, nil
		}
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "API key not found")
}

func (m *mockAPIKeyRepo) ListByUserID(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	var list []*domain.APIKey
	for _, k := range m.keys {
		if k.UserID == userID {
			list = append(list, k)
		}
	}
	return list, nil
}

func (m *mockAPIKeyRepo) Revoke(ctx context.Context, id string) error {
	if k, ok := m.keys[id]; ok {
		k.Revoke()
	}
	return nil
}

func (m *mockAPIKeyRepo) Update(ctx context.Context, key *domain.APIKey) error {
	m.keys[key.ID] = key
	return nil
}

type mockTokenRepo struct {
	tokens map[string]*domain.VerificationToken
}

func newMockTokenRepo() domain.VerificationTokenRepository {
	return &mockTokenRepo{tokens: make(map[string]*domain.VerificationToken)}
}

func (m *mockTokenRepo) Create(ctx context.Context, token *domain.VerificationToken) error {
	m.tokens[token.Token] = token
	return nil
}

func (m *mockTokenRepo) GetByToken(ctx context.Context, tokenStr string) (*domain.VerificationToken, error) {
	if t, ok := m.tokens[tokenStr]; ok {
		return t, nil
	}
	return nil, appErrors.New(appErrors.CodeNotFound, "token not found")
}

func (m *mockTokenRepo) Delete(ctx context.Context, tokenStr string) error {
	delete(m.tokens, tokenStr)
	return nil
}

type mockAuditRepo struct {
	logs []*domain.AuditLog
}

func newMockAuditRepo() domain.AuditLogRepository {
	return &mockAuditRepo{}
}

func (m *mockAuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockAuditRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.AuditLog, error) {
	var res []*domain.AuditLog
	for _, l := range m.logs {
		if l.UserID == userID {
			res = append(res, l)
		}
	}
	return res, nil
}

func setupAuthUseCase() AuthUseCase {
	jwtManager := security.NewJWTManager("test-secret-key-for-auth-tests-32b", "anarva-test", 15*time.Minute, 24*time.Hour)
	return NewAuthUseCase(
		newMockUserRepo(),
		newMockSessionRepo(),
		newMockAPIKeyRepo(),
		newMockTokenRepo(),
		newMockAuditRepo(),
		jwtManager,
		15*time.Minute,
		24*time.Hour,
	)
}

func TestSignUp_Success(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	user, token, err := uc.SignUp(ctx, "dev@anarva.io", "Password123!", "Senior Dev")
	require.NoError(t, err)
	assert.Equal(t, "dev@anarva.io", user.Email)
	assert.Equal(t, domain.UserStatusPending, user.Status)
	assert.NotEmpty(t, token)
}

func TestLogin_SuccessAndVerify(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	user, token, err := uc.SignUp(ctx, "user@anarva.io", "SecretPass123", "John Doe")
	require.NoError(t, err)

	// Verify email
	err = uc.VerifyEmail(ctx, token)
	require.NoError(t, err)

	// Login
	accessTok, refreshTok, expiry, loggedInUser, err := uc.Login(ctx, "user@anarva.io", "SecretPass123", "Mozilla", "127.0.0.1")
	require.NoError(t, err)
	assert.NotEmpty(t, accessTok)
	assert.NotEmpty(t, refreshTok)
	assert.Equal(t, 15*time.Minute, expiry)
	assert.Equal(t, user.ID, loggedInUser.ID)
}

func TestAPIKeyLifecycle(t *testing.T) {
	uc := setupAuthUseCase()
	ctx := context.Background()

	user, _, err := uc.SignUp(ctx, "apikey@anarva.io", "Pass123!", "Key Master")
	require.NoError(t, err)

	rawKey, keyEntity, err := uc.CreateAPIKey(ctx, user.ID, "CLI Deployment Key", 30)
	require.NoError(t, err)
	assert.Contains(t, rawKey, "anarva_live_")
	assert.Equal(t, "CLI Deployment Key", keyEntity.Name)

	keys, err := uc.ListAPIKeys(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, keys, 1)

	err = uc.RevokeAPIKey(ctx, user.ID, keyEntity.ID)
	require.NoError(t, err)

	keysAfter, _ := uc.ListAPIKeys(ctx, user.ID)
	assert.True(t, keysAfter[0].IsRevoked)
}
