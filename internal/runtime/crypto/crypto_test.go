package crypto

import (
	"regexp"
	"testing"
)

func TestMD5(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"hello world", "5eb63bbbe01eeed093cb22bb8f5acdc3"},
	}

	for _, tt := range tests {
		result := MD5(tt.input)
		if result != tt.expected {
			t.Errorf("MD5(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSHA1(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"hello world", "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"},
	}

	for _, tt := range tests {
		result := SHA1(tt.input)
		if result != tt.expected {
			t.Errorf("SHA1(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSHA256(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"hello world", "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"},
	}

	for _, tt := range tests {
		result := SHA256(tt.input)
		if result != tt.expected {
			t.Errorf("SHA256(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSHA512(t *testing.T) {
	// Just test that it produces the right length (128 hex chars = 64 bytes)
	result := SHA512("hello")
	if len(result) != 128 {
		t.Errorf("SHA512 returned %d chars, want 128", len(result))
	}
}

func TestHMACSHA1(t *testing.T) {
	tests := []struct {
		message  string
		key      string
		expected string
	}{
		{"hello", "secret", "5112055c05f944f85755efc5cd8970e194e9f45b"},
	}

	for _, tt := range tests {
		result := HMACSHA1(tt.message, tt.key)
		if result != tt.expected {
			t.Errorf("HMACSHA1(%q, %q) = %q, want %q", tt.message, tt.key, result, tt.expected)
		}
	}
}

func TestHMACSHA256(t *testing.T) {
	tests := []struct {
		message  string
		key      string
		expected string
	}{
		{"hello", "secret", "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b"},
	}

	for _, tt := range tests {
		result := HMACSHA256(tt.message, tt.key)
		if result != tt.expected {
			t.Errorf("HMACSHA256(%q, %q) = %q, want %q", tt.message, tt.key, result, tt.expected)
		}
	}
}

func TestHMACSHA512(t *testing.T) {
	// Just test that it produces the right length (128 hex chars = 64 bytes)
	result := HMACSHA512("hello", "secret")
	if len(result) != 128 {
		t.Errorf("HMACSHA512 returned %d chars, want 128", len(result))
	}
}

func TestUUID(t *testing.T) {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	for i := 0; i < 10; i++ {
		result := UUID()
		if !uuidRegex.MatchString(result) {
			t.Errorf("UUID() = %q, not a valid UUID v4", result)
		}
	}

	// Test uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		u := UUID()
		if seen[u] {
			t.Errorf("UUID() produced duplicate: %s", u)
		}
		seen[u] = true
	}
}
