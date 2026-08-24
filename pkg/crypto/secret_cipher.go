package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

var (
	ErrInvalidKeySize        = errors.New("encryption key must be exactly 32 bytes (64 hex characters)")
	ErrInvalidCiphertext     = errors.New("invalid ciphertext format")
	ErrDecryptionFailed      = errors.New("decryption failed: authentication tag mismatch or corrupted payload")
	ErrMissingEncryptionKey  = errors.New("COMPUTE_SECRET_ENCRYPTION_KEY is required but missing or invalid")
)

type SecretCipher interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertext string) ([]byte, error)
}

type AESGCMCipher struct {
	key     []byte
	version string
}

func NewAESGCMCipher(keyHex string, version string) (*AESGCMCipher, error) {
	keyHex = strings.TrimSpace(keyHex)
	if len(keyHex) == 0 {
		return nil, ErrMissingEncryptionKey
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		// Fallback: If raw key string is exactly 32 bytes
		if len([]byte(keyHex)) == 32 {
			key = []byte(keyHex)
		} else {
			return nil, ErrInvalidKeySize
		}
	}

	if version == "" {
		version = "v1"
	}

	return &AESGCMCipher{
		key:     key,
		version: version,
	}, nil
}

func (c *AESGCMCipher) Encrypt(plaintext []byte) (string, error) {
	if len(plaintext) == 0 {
		return "", nil
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM mode: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	sealed := gcm.Seal(nil, nonce, plaintext, nil)
	return fmt.Sprintf("anarva:%s:%s:%s", c.version, hex.EncodeToString(nonce), hex.EncodeToString(sealed)), nil
}

func (c *AESGCMCipher) Decrypt(ciphertextStr string) ([]byte, error) {
	ciphertextStr = strings.TrimSpace(ciphertextStr)
	if len(ciphertextStr) == 0 {
		return nil, nil
	}

	// Legacy Plaintext Fallback: Check if string is unencrypted JSON
	if strings.HasPrefix(ciphertextStr, "{") || strings.HasPrefix(ciphertextStr, "[") {
		return []byte(ciphertextStr), nil
	}

	parts := strings.Split(ciphertextStr, ":")
	if len(parts) != 4 || parts[0] != "anarva" {
		return nil, ErrInvalidCiphertext
	}

	// Version check (v1)
	// parts[1] is version (e.g. "v1")
	nonce, err := hex.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	sealed, err := hex.DecodeString(parts[3])
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	if len(nonce) != gcm.NonceSize() {
		return nil, ErrInvalidCiphertext
	}

	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// Global Cipher instance management for GORM hooks
var (
	globalCipher     SecretCipher
	globalCipherMu   sync.RWMutex
	defaultDevKeyHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 32 bytes hex
)

func SetGlobalCipher(c SecretCipher) {
	globalCipherMu.Lock()
	defer globalCipherMu.Unlock()
	globalCipher = c
}

func GetGlobalCipher() SecretCipher {
	globalCipherMu.RLock()
	defer globalCipherMu.RUnlock()
	if globalCipher == nil {
		// Default dev cipher fallback for tests / local dev
		devCipher, _ := NewAESGCMCipher(defaultDevKeyHex, "v1")
		return devCipher
	}
	return globalCipher
}
