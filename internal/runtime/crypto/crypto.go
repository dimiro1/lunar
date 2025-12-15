package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"

	"github.com/google/uuid"
)

// MD5 computes the MD5 hash of the input string and returns it as a hex-encoded string.
func MD5(input string) string {
	return hashString(md5.New(), input)
}

// SHA1 computes the SHA1 hash of the input string and returns it as a hex-encoded string.
func SHA1(input string) string {
	return hashString(sha1.New(), input)
}

// SHA256 computes the SHA256 hash of the input string and returns it as a hex-encoded string.
func SHA256(input string) string {
	return hashString(sha256.New(), input)
}

// SHA512 computes the SHA512 hash of the input string and returns it as a hex-encoded string.
func SHA512(input string) string {
	return hashString(sha512.New(), input)
}

// HMACSHA1 computes the HMAC-SHA1 of a message with a secret key.
func HMACSHA1(message, key string) string {
	return hmacString(sha1.New, message, key)
}

// HMACSHA256 computes the HMAC-SHA256 of a message with a secret key.
func HMACSHA256(message, key string) string {
	return hmacString(sha256.New, message, key)
}

// HMACSHA512 computes the HMAC-SHA512 of a message with a secret key.
func HMACSHA512(message, key string) string {
	return hmacString(sha512.New, message, key)
}

// UUID generates a new UUID v4.
func UUID() string {
	return uuid.New().String()
}

// hashString is a helper function to compute hash of a string.
func hashString(h hash.Hash, input string) string {
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

// hmacString is a helper function to compute HMAC of a string.
func hmacString(hashFunc func() hash.Hash, message, key string) string {
	h := hmac.New(hashFunc, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
