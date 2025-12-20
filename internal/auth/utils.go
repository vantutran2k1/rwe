package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

const keyPrefix = "rwe_sk_"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func GenerateAPIKey() (string, string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}

	randomPart := hex.EncodeToString(bytes)
	rawKey := fmt.Sprintf("%s%s", keyPrefix, randomPart)
	hashedKey := HashKey(rawKey)

	return rawKey, hashedKey, keyPrefix, nil
}

func HashKey(rawKey string) string {
	h := sha256.New()
	h.Write([]byte(rawKey))

	return hex.EncodeToString(h.Sum(nil))
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

func ValidateEmail(email string) error {
	if len(email) < 3 || len(email) > 254 {
		return fmt.Errorf("email length must be between 3 and 254 characters")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}
