package random

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"testing"
)

func TestInt(t *testing.T) {
	t.Run("valid range", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			result, err := Int(1, 10)
			if err != nil {
				t.Fatalf("Int(1, 10) returned error: %v", err)
			}
			if result < 1 || result > 10 {
				t.Errorf("Int(1, 10) = %d, want 1-10", result)
			}
		}
	})

	t.Run("same min and max", func(t *testing.T) {
		result, err := Int(5, 5)
		if err != nil {
			t.Fatalf("Int(5, 5) returned error: %v", err)
		}
		if result != 5 {
			t.Errorf("Int(5, 5) = %d, want 5", result)
		}
	})

	t.Run("invalid range", func(t *testing.T) {
		_, err := Int(10, 1)
		if err == nil {
			t.Error("Int(10, 1) expected error, got nil")
		}
	})

	t.Run("negative range", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			result, err := Int(-10, -1)
			if err != nil {
				t.Fatalf("Int(-10, -1) returned error: %v", err)
			}
			if result < -10 || result > -1 {
				t.Errorf("Int(-10, -1) = %d, want -10 to -1", result)
			}
		}
	})
}

func TestFloat(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := Float()
		if result < 0.0 || result >= 1.0 {
			t.Errorf("Float() = %f, want 0.0 <= x < 1.0", result)
		}
	}
}

func TestString(t *testing.T) {
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	t.Run("valid length", func(t *testing.T) {
		for _, length := range []int{1, 10, 32, 64} {
			result, err := String(length)
			if err != nil {
				t.Fatalf("String(%d) returned error: %v", length, err)
			}
			if len(result) != length {
				t.Errorf("String(%d) returned length %d", length, len(result))
			}
			if !alphanumericRegex.MatchString(result) {
				t.Errorf("String(%d) = %q, not alphanumeric", length, result)
			}
		}
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := String(0)
		if err == nil {
			t.Error("String(0) expected error, got nil")
		}

		_, err = String(-1)
		if err == nil {
			t.Error("String(-1) expected error, got nil")
		}
	})

	t.Run("uniqueness", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			s, _ := String(32)
			if seen[s] {
				t.Errorf("String(32) produced duplicate: %s", s)
			}
			seen[s] = true
		}
	})
}

func TestBytes(t *testing.T) {
	t.Run("valid length", func(t *testing.T) {
		for _, length := range []int{1, 16, 32} {
			result, err := Bytes(length)
			if err != nil {
				t.Fatalf("Bytes(%d) returned error: %v", length, err)
			}
			// Result is base64 encoded, so decode it
			decoded, err := base64.StdEncoding.DecodeString(result)
			if err != nil {
				t.Errorf("Bytes(%d) returned invalid base64: %v", length, err)
				continue
			}
			if len(decoded) != length {
				t.Errorf("Bytes(%d) decoded to %d bytes", length, len(decoded))
			}
		}
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := Bytes(0)
		if err == nil {
			t.Error("Bytes(0) expected error, got nil")
		}
	})
}

func TestHex(t *testing.T) {
	t.Run("valid length", func(t *testing.T) {
		for _, length := range []int{1, 16, 32} {
			result, err := Hex(length)
			if err != nil {
				t.Fatalf("Hex(%d) returned error: %v", length, err)
			}
			// Result is hex encoded, so decode it
			decoded, err := hex.DecodeString(result)
			if err != nil {
				t.Errorf("Hex(%d) returned invalid hex: %v", length, err)
				continue
			}
			if len(decoded) != length {
				t.Errorf("Hex(%d) decoded to %d bytes", length, len(decoded))
			}
		}
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := Hex(0)
		if err == nil {
			t.Error("Hex(0) expected error, got nil")
		}
	})
}

func TestID(t *testing.T) {
	t.Run("format", func(t *testing.T) {
		result := ID()
		// XID is 20 characters
		if len(result) != 20 {
			t.Errorf("ID() = %q, want 20 characters, got %d", result, len(result))
		}
	})

	t.Run("uniqueness", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := ID()
			if seen[id] {
				t.Errorf("ID() produced duplicate: %s", id)
			}
			seen[id] = true
		}
	})
}
