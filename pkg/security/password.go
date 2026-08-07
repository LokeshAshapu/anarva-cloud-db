package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultBcryptCost = 12

// HashPassword hashes a raw password using bcrypt with cost 12.
func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), defaultBcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

// ComparePassword verifies if a raw password matches a hashed password.
func ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
