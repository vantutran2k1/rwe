package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const KeyPrefix = "rwe_sk_"

func GenerateAPIKey() (string, string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}

	randomPart := hex.EncodeToString(bytes)

	rawKey := fmt.Sprintf("%s%s", KeyPrefix, randomPart)

	hashedKey := HashKey(rawKey)

	return rawKey, hashedKey, KeyPrefix, nil
}

func HashKey(rawKey string) string {
	h := sha256.New()
	h.Write([]byte(rawKey))
	return hex.EncodeToString(h.Sum(nil))
}
