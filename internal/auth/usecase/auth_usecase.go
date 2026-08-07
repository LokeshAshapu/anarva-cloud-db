package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/anarva-cloud/anarva-cloud-db/internal/auth/domain"
	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/logger"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/security"
)

type AuthUseCase interface {
	SignUp(ctx context.Context, email, password, fullName string) (*domain.User, string, error)
	VerifyEmail(ctx context.Context, tokenStr string) error
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (accessToken, refreshToken string, accessExpiry time.Duration, user *domain.User, err error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (newAccessToken, newRefreshToken string, accessExpiry time.Duration, err error)
	CreateAPIKey(ctx context.Context, userID, name string, expiryDays int) (rawKey string, apiKey *domain.APIKey, err error)
	ListAPIKeys(ctx context.Context, userID string) ([]*domain.APIKey, error)
	RevokeAPIKey(ctx context.Context, userID, keyID string) error
	ValidateToken(ctx context.Context, tokenStr string) (*security.Claims, error)
}

type authUseCase struct {
	userRepo        domain.UserRepository
	sessionRepo     domain.SessionRepository
	apiKeyRepo      domain.APIKeyRepository
	tokenRepo       domain.VerificationTokenRepository
	auditRepo       domain.AuditLogRepository
	jwtManager      *security.JWTManager
	accessDuration  time.Duration
	refreshDuration time.Duration
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	apiKeyRepo domain.APIKeyRepository,
	tokenRepo domain.VerificationTokenRepository,
	auditRepo domain.AuditLogRepository,
	jwtManager *security.JWTManager,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) AuthUseCase {
	return &authUseCase{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		apiKeyRepo:      apiKeyRepo,
		tokenRepo:       tokenRepo,
		auditRepo:       auditRepo,
		jwtManager:      jwtManager,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

func (u *authUseCase) SignUp(ctx context.Context, email, password, fullName string) (*domain.User, string, error) {
	if email == "" || password == "" || fullName == "" {
		return nil, "", appErrors.New(appErrors.CodeInvalidInput, "email, password, and fullName are required")
	}

	existingUser, _ := u.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, "", appErrors.New(appErrors.CodeAlreadyExists, "user with this email already exists")
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return nil, "", appErrors.Wrap(err, appErrors.CodeInternal, "failed to hash password")
	}

	user := domain.NewUser(email, hashedPassword, fullName)
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate verification token
	verificationCode := uuid.New().String()
	token := domain.NewVerificationToken(user.ID, verificationCode, 24*time.Hour)
	if err := u.tokenRepo.Create(ctx, token); err != nil {
		return nil, "", err
	}

	u.auditRepo.Create(ctx, domain.NewAuditLog(user.ID, "USER_SIGNUP", "", "", "SUCCESS", "User account registered"))
	logger.Context(ctx).Info(fmt.Sprintf("New user registered: %s", email))

	return user, verificationCode, nil
}

func (u *authUseCase) VerifyEmail(ctx context.Context, tokenStr string) error {
	vToken, err := u.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return appErrors.New(appErrors.CodeInvalidInput, "invalid or expired verification token")
	}

	if vToken.IsExpired() {
		_ = u.tokenRepo.Delete(ctx, tokenStr)
		return appErrors.New(appErrors.CodeInvalidInput, "verification token expired")
	}

	user, err := u.userRepo.GetByID(ctx, vToken.UserID)
	if err != nil {
		return err
	}

	user.Activate()
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	_ = u.tokenRepo.Delete(ctx, tokenStr)
	u.auditRepo.Create(ctx, domain.NewAuditLog(user.ID, "EMAIL_VERIFICATION", "", "", "SUCCESS", "Email address verified"))

	return nil
}

func (u *authUseCase) Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, time.Duration, *domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		u.auditRepo.Create(ctx, domain.NewAuditLog("", "USER_LOGIN", ipAddress, userAgent, "FAILED", fmt.Sprintf("Email %s not found", email)))
		return "", "", 0, nil, appErrors.New(appErrors.CodeUnauthorized, "invalid email or password")
	}

	if !security.ComparePassword(password, user.PasswordHash) {
		u.auditRepo.Create(ctx, domain.NewAuditLog(user.ID, "USER_LOGIN", ipAddress, userAgent, "FAILED", "Incorrect password"))
		return "", "", 0, nil, appErrors.New(appErrors.CodeUnauthorized, "invalid email or password")
	}

	if user.Status == domain.UserStatusSuspended {
		return "", "", 0, nil, appErrors.New(appErrors.CodeForbidden, "user account suspended")
	}

	accessToken, refreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role), "")
	if err != nil {
		return "", "", 0, nil, appErrors.Wrap(err, appErrors.CodeInternal, "failed to generate token pair")
	}

	session := domain.NewSession(user.ID, refreshToken, userAgent, ipAddress, time.Now().Add(u.refreshDuration))
	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return "", "", 0, nil, err
	}

	u.auditRepo.Create(ctx, domain.NewAuditLog(user.ID, "USER_LOGIN", ipAddress, userAgent, "SUCCESS", "User logged in successfully"))
	return accessToken, refreshToken, u.accessDuration, user, nil
}

func (u *authUseCase) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, time.Duration, error) {
	claims, err := u.jwtManager.ValidateToken(refreshTokenStr)
	if err != nil || claims.TokenType != "refresh" {
		return "", "", 0, appErrors.New(appErrors.CodeUnauthorized, "invalid refresh token")
	}

	session, err := u.sessionRepo.GetByRefreshToken(ctx, refreshTokenStr)
	if err != nil || session.IsRevoked || time.Now().After(session.ExpiresAt) {
		return "", "", 0, appErrors.New(appErrors.CodeUnauthorized, "session expired or revoked")
	}

	user, err := u.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return "", "", 0, err
	}

	// Revoke old session
	_ = u.sessionRepo.Revoke(ctx, session.ID)

	// Issue new token pair
	newAccessToken, newRefreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role), "")
	if err != nil {
		return "", "", 0, appErrors.Wrap(err, appErrors.CodeInternal, "failed to generate new token pair")
	}

	newSession := domain.NewSession(user.ID, newRefreshToken, session.UserAgent, session.IPAddress, time.Now().Add(u.refreshDuration))
	if err := u.sessionRepo.Create(ctx, newSession); err != nil {
		return "", "", 0, err
	}

	return newAccessToken, newRefreshToken, u.accessDuration, nil
}

func (u *authUseCase) CreateAPIKey(ctx context.Context, userID, name string, expiryDays int) (string, *domain.APIKey, error) {
	if name == "" {
		return "", nil, appErrors.New(appErrors.CodeInvalidInput, "API key name is required")
	}

	rawKey, hashedKey, err := security.GenerateAPIKey("anarva_live")
	if err != nil {
		return "", nil, appErrors.Wrap(err, appErrors.CodeInternal, "failed to generate API key")
	}

	var expiresAt *time.Time
	if expiryDays > 0 {
		exp := time.Now().AddDate(0, 0, expiryDays)
		expiresAt = &exp
	}

	apiKey := domain.NewAPIKey(userID, name, "anarva_live", hashedKey, expiresAt)
	if err := u.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return "", nil, err
	}

	u.auditRepo.Create(ctx, domain.NewAuditLog(userID, "CREATE_API_KEY", "", "", "SUCCESS", fmt.Sprintf("Created API key '%s'", name)))
	return rawKey, apiKey, nil
}

func (u *authUseCase) ListAPIKeys(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	return u.apiKeyRepo.ListByUserID(ctx, userID)
}

func (u *authUseCase) RevokeAPIKey(ctx context.Context, userID, keyID string) error {
	apiKey, err := u.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return err
	}

	if apiKey.UserID != userID {
		return appErrors.New(appErrors.CodeForbidden, "permission denied")
	}

	if err := u.apiKeyRepo.Revoke(ctx, keyID); err != nil {
		return err
	}

	u.auditRepo.Create(ctx, domain.NewAuditLog(userID, "REVOKE_API_KEY", "", "", "SUCCESS", fmt.Sprintf("Revoked API key '%s'", keyID)))
	return nil
}

func (u *authUseCase) ValidateToken(ctx context.Context, tokenStr string) (*security.Claims, error) {
	return u.jwtManager.ValidateToken(tokenStr)
}
