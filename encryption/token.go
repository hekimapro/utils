package encryption

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/hekimapro/utils/helpers"
)

// GenerateRefreshToken generates a cryptographically secure random token
func GenerateRefreshToken() (string, error) {
	refreshTokenLength := helpers.GetENVIntValue("REFRESH_TOKEN_LENGTH", 12)
	b := make([]byte, refreshTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

// HashToken creates a SHA-256 hash of a token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ValidateTokenHash compares a token with its hash
func ValidateTokenHash(token, hash string) bool {
	return HashToken(token) == hash
}
