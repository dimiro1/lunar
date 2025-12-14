package json

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  string
		expectErr bool
	}{
		{"nil", nil, "null", false},
		{"bool true", true, "true", false},
		{"bool false", false, "false", false},
		{"int", 42, "42", false},
		{"float", 3.14, "3.14", false},
		{"string", "hello", `"hello"`, false},
		{"empty array", []any{}, "[]", false},
		{"array", []any{1, 2, 3}, "[1,2,3]", false},
		{"empty object", map[string]any{}, "{}", false},
		{"object", map[string]any{"key": "value"}, `{"key":"value"}`, false},
		{"nested", map[string]any{"arr": []any{1, 2}}, `{"arr":[1,2]}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Encode(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Encode(%v) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Encode(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  any
		expectErr bool
	}{
		{"null", "null", nil, false},
		{"bool true", "true", true, false},
		{"bool false", "false", false, false},
		{"int", "42", float64(42), false},
		{"float", "3.14", 3.14, false},
		{"string", `"hello"`, "hello", false},
		{"empty array", "[]", []any{}, false},
		{"array", "[1,2,3]", []any{float64(1), float64(2), float64(3)}, false},
		{"empty object", "{}", map[string]any{}, false},
		{"object", `{"key":"value"}`, map[string]any{"key": "value"}, false},
		{"invalid json", "not json", nil, true},
		{"incomplete", `{"key":`, nil, true},
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
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Decode(%q) = %v (%T), want %v (%T)", tt.input, result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []any{
		nil,
		true,
		false,
		float64(42),
		"hello world",
		[]any{float64(1), float64(2), float64(3)},
		map[string]any{"name": "test", "value": float64(123)},
	}

	for _, input := range inputs {
		encoded, err := Encode(input)
		if err != nil {
			t.Errorf("Encode(%v) failed: %v", input, err)
			continue
		}
		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Decode(%q) failed: %v", encoded, err)
			continue
		}
		if !reflect.DeepEqual(decoded, input) {
			t.Errorf("Round trip failed: %v -> %q -> %v", input, encoded, decoded)
		}
	}
}
