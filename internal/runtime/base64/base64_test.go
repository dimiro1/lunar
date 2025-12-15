package base64

import "testing"

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"hello world", "hello world", "aGVsbG8gd29ybGQ="},
		{"simple text", "test", "dGVzdA=="},
		{"with special chars", "hello@world!", "aGVsbG9Ad29ybGQh"},
		{"unicode", "hello 世界", "aGVsbG8g5LiW55WM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Encode(tt.input)
			if result != tt.expected {
				t.Errorf("Encode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{"empty string", "", "", false},
		{"hello world", "aGVsbG8gd29ybGQ=", "hello world", false},
		{"simple text", "dGVzdA==", "test", false},
		{"with special chars", "aGVsbG9Ad29ybGQh", "hello@world!", false},
		{"unicode", "aGVsbG8g5LiW55WM", "hello 世界", false},
		{"invalid base64", "not-valid-base64!!!", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Decode(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Decode(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Decode(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Decode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{
		"",
		"hello",
		"hello world",
		"special chars: !@#$%^&*()",
		"unicode: 你好世界 🌍",
		"binary-like: \x00\x01\x02\xff",
	}

	for _, input := range inputs {
		encoded := Encode(input)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Round trip failed for %q: %v", input, err)
			continue
		}
		if decoded != input {
			t.Errorf("Round trip failed for %q: got %q", input, decoded)
		}
	}
}
