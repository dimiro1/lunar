package random

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	mathrand "math/rand"

	"github.com/rs/xid"
)

const alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Int generates a cryptographically secure random integer between min and max (inclusive).
// Falls back to math/rand on crypto/rand errors.
func Int(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min (%d) must be less than or equal to max (%d)", min, max)
	}

	rangeSize := int64(max - min + 1)
	n, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
	if err != nil {
		// Fallback to math/rand
		return mathrand.Intn(int(rangeSize)) + min, nil
	}
	return int(n.Int64()) + min, nil
}

// Float generates a random float64 between 0.0 and 1.0.
func Float() float64 {
	return mathrand.Float64()
}

// String generates a random alphanumeric string of the specified length.
func String(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive, got %d", length)
	}

	bytes := make([]byte, length)
	for i := range bytes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumericCharset))))
		if err != nil {
			// Fallback to math/rand
			bytes[i] = alphanumericCharset[mathrand.Intn(len(alphanumericCharset))]
		} else {
			bytes[i] = alphanumericCharset[n.Int64()]
		}
	}
	return string(bytes), nil
}

// Bytes generates random bytes and returns them as a base64-encoded string.
func Bytes(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive, got %d", length)
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Hex generates random bytes and returns them as a hex-encoded string.
func Hex(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive, got %d", length)
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ID generates a globally unique sortable ID using xid.
// Returns a 20-character string that is smaller than UUID and sortable by creation time.
func ID() string {
	return xid.New().String()
}
