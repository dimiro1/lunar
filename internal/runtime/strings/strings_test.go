package strings

import (
	"reflect"
	"testing"
)

func TestTrim(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"\t\nhello\n\t", "hello"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		result := Trim(tt.input)
		if result != tt.expected {
			t.Errorf("Trim(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTrimLeft(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello  "},
		{"hello", "hello"},
		{"\t\nhello", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		result := TrimLeft(tt.input)
		if result != tt.expected {
			t.Errorf("TrimLeft(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTrimRight(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "  hello"},
		{"hello", "hello"},
		{"hello\t\n", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		result := TrimRight(tt.input)
		if result != tt.expected {
			t.Errorf("TrimRight(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		s        string
		sep      string
		expected []string
	}{
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{"hello world", " ", []string{"hello", "world"}},
		{"no-sep", ",", []string{"no-sep"}},
		{"", ",", []string{""}},
		{"a::b", ":", []string{"a", "", "b"}},
	}

	for _, tt := range tests {
		result := Split(tt.s, tt.sep)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("Split(%q, %q) = %v, want %v", tt.s, tt.sep, result, tt.expected)
		}
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		parts    []string
		sep      string
		expected string
	}{
		{[]string{"a", "b", "c"}, ",", "a,b,c"},
		{[]string{"hello", "world"}, " ", "hello world"},
		{[]string{"single"}, ",", "single"},
		{[]string{}, ",", ""},
		{[]string{"a", "", "b"}, ":", "a::b"},
	}

	for _, tt := range tests {
		result := Join(tt.parts, tt.sep)
		if result != tt.expected {
			t.Errorf("Join(%v, %q) = %q, want %q", tt.parts, tt.sep, result, tt.expected)
		}
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		s        string
		prefix   string
		expected bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", false},
		{"hello", "hello", true},
		{"hello", "hello world", false},
		{"", "", true},
		{"hello", "", true},
	}

	for _, tt := range tests {
		result := HasPrefix(tt.s, tt.prefix)
		if result != tt.expected {
			t.Errorf("HasPrefix(%q, %q) = %v, want %v", tt.s, tt.prefix, result, tt.expected)
		}
	}
}

func TestHasSuffix(t *testing.T) {
	tests := []struct {
		s        string
		suffix   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", false},
		{"hello", "hello", true},
		{"hello", "hello world", false},
		{"", "", true},
		{"hello", "", true},
	}

	for _, tt := range tests {
		result := HasSuffix(tt.s, tt.suffix)
		if result != tt.expected {
			t.Errorf("HasSuffix(%q, %q) = %v, want %v", tt.s, tt.suffix, result, tt.expected)
		}
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		s        string
		old      string
		new      string
		n        int
		expected string
	}{
		{"hello hello", "hello", "hi", -1, "hi hi"},
		{"hello hello", "hello", "hi", 1, "hi hello"},
		{"hello", "x", "y", -1, "hello"},
		{"aaa", "a", "b", 2, "bba"},
		{"", "a", "b", -1, ""},
	}

	for _, tt := range tests {
		result := Replace(tt.s, tt.old, tt.new, tt.n)
		if result != tt.expected {
			t.Errorf("Replace(%q, %q, %q, %d) = %q, want %q", tt.s, tt.old, tt.new, tt.n, result, tt.expected)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"hello", "hello"},
		{"", ""},
		{"123ABC", "123abc"},
	}

	for _, tt := range tests {
		result := ToLower(tt.input)
		if result != tt.expected {
			t.Errorf("ToLower(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"HELLO", "HELLO"},
		{"", ""},
		{"123abc", "123ABC"},
	}

	for _, tt := range tests {
		result := ToUpper(tt.input)
		if result != tt.expected {
			t.Errorf("ToUpper(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "xyz", false},
		{"hello", "hello world", false},
		{"", "", true},
		{"hello", "", true},
	}

	for _, tt := range tests {
		result := Contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("Contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestRepeat(t *testing.T) {
	tests := []struct {
		s        string
		count    int
		expected string
	}{
		{"ab", 3, "ababab"},
		{"x", 5, "xxxxx"},
		{"hello", 1, "hello"},
		{"hello", 0, ""},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := Repeat(tt.s, tt.count)
		if result != tt.expected {
			t.Errorf("Repeat(%q, %d) = %q, want %q", tt.s, tt.count, result, tt.expected)
		}
	}
}
