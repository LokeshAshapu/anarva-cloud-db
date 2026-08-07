package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// Encrypt encrypts plaintext using AES-256-GCM with a 32-byte key.
func Encrypt(plaintext []byte, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be exactly 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64 encoded ciphertext using AES-256-GCM.
func Decrypt(cryptoText string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode cipher text: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	return plaintext, nil
}

// HashAPIKey computes a SHA-256 hash of a raw API key.
func HashAPIKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}

// VerifyAPIKey verifies a raw API key against a hex SHA-256 hash.
func VerifyAPIKey(rawKey, hashedKey string) bool {
	expectedHash := HashAPIKey(rawKey)
	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(hashedKey)) == 1
}

// GenerateAPIKey generates a cryptographically secure random API key with prefix.
func GenerateAPIKey(prefix string) (rawKey string, hashedKey string, err error) {
	bytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	if prefix == "" {
		prefix = "anarva_live"
	}

	rawKey = fmt.Sprintf("%s_%s_%s", prefix, uuid.New().String()[:8], hex.EncodeToString(bytes))
	hashedKey = HashAPIKey(rawKey)

	return rawKey, hashedKey, nil
}
