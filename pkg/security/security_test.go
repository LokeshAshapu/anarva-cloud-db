package security

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHashing(t *testing.T) {
	password := "SuperSecurePassword123!"
	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	assert.True(t, ComparePassword(password, hash))
	assert.False(t, ComparePassword("WrongPassword", hash))
}

func TestJWTManager(t *testing.T) {
	jwtManager := NewJWTManager("super-secret-key-32-bytes-length-x", "anarva-test", 15*time.Minute, 24*time.Hour)

	userID := "user-uuid-12345"
	email := "admin@anarva.io"
	role := "ADMIN"
	orgID := "org-uuid-99999"

	accessTok, refreshTok, err := jwtManager.GenerateTokenPair(userID, email, role, orgID)
	require.NoError(t, err)
	assert.NotEmpty(t, accessTok)
	assert.NotEmpty(t, refreshTok)

	claims, err := jwtManager.ValidateToken(accessTok)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, orgID, claims.OrgID)
	assert.Equal(t, "access", claims.TokenType)

	// Test invalid token
	_, err = jwtManager.ValidateToken("invalid.token.string")
	assert.Error(t, err)
}

func TestAES256GCMEncryption(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	originalSecret := "database_master_password_secret"

	encryptedText, err := Encrypt([]byte(originalSecret), key)
	require.NoError(t, err)
	assert.NotEmpty(t, encryptedText)
	assert.NotEqual(t, originalSecret, encryptedText)

	decryptedBytes, err := Decrypt(encryptedText, key)
	require.NoError(t, err)
	assert.Equal(t, originalSecret, string(decryptedBytes))
}

func TestGenerateAPIKey(t *testing.T) {
	rawKey, hashedKey, err := GenerateAPIKey("anarva_test")
	require.NoError(t, err)
	assert.Contains(t, rawKey, "anarva_test_")
	assert.NotEmpty(t, hashedKey)

	assert.True(t, VerifyAPIKey(rawKey, hashedKey))
	assert.False(t, VerifyAPIKey("anarva_test_wrongkey", hashedKey))
}
