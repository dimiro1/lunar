// Package token provides secure token generation and hashing utilities.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// Generate creates a cryptographically secure random token.
// Returns a 64-character hex string (32 random bytes).
func Generate() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

// Hash returns the hex-encoded SHA-256 hash of a token string.
func Hash(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}
