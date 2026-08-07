package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	appErrors "github.com/anarva-cloud/anarva-cloud-db/pkg/errors"
)

var (
	ErrInvalidToken = appErrors.New(appErrors.CodeUnauthorized, "invalid or expired token")
)

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	OrgID    string `json:"org_id,omitempty"`
	TokenType string `json:"token_type"` // access or refresh
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret        []byte
	issuer        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTManager(secret, issuer string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		issuer:        issuer,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair generates an Access Token and Refresh Token for a user.
func (m *JWTManager) GenerateTokenPair(userID, email, role, orgID string) (accessToken string, refreshToken string, err error) {
	now := time.Now()

	// Access Token
	accessClaims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		OrgID:     orgID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = token.SignedString(m.secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	refreshClaims := &Claims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	rToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = rToken.SignedString(m.secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken parses and validates a JWT token string.
func (m *JWTManager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
