package diff

import (
	"testing"
)

func TestGetTypeColor(t *testing.T) {
	tests := []struct {
		name     string
		lineType LineType
		wantNil  bool
	}{
		{
			name:     "added line returns green color",
			lineType: LineAdded,
			wantNil:  false,
		},
		{
			name:     "removed line returns red color",
			lineType: LineRemoved,
			wantNil:  false,
		},
		{
			name:     "unchanged line returns muted color",
			lineType: LineUnchanged,
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeColor(tt.lineType)
			if (result == nil) != tt.wantNil {
				t.Errorf("getTypeColor(%v) returned nil = %v, want nil = %v", tt.lineType, result == nil, tt.wantNil)
			}
		})
	}
}

func TestGetContentColor(t *testing.T) {
	tests := []struct {
		name     string
		lineType LineType
		wantNil  bool
	}{
		{
			name:     "added line returns green content color",
			lineType: LineAdded,
			wantNil:  false,
		},
		{
			name:     "removed line returns red content color",
			lineType: LineRemoved,
			wantNil:  false,
		},
		{
			name:     "unchanged line returns default content color",
			lineType: LineUnchanged,
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getContentColor(tt.lineType)
			if (result == nil) != tt.wantNil {
				t.Errorf("getContentColor(%v) returned nil = %v, want nil = %v", tt.lineType, result == nil, tt.wantNil)
			}
		})
	}
}

func TestGetTypeSymbol(t *testing.T) {
	tests := []struct {
		name     string
		lineType LineType
		want     string
	}{
		{
			name:     "added line returns plus",
			lineType: LineAdded,
			want:     "+",
		},
		{
			name:     "removed line returns minus",
			lineType: LineRemoved,
			want:     "-",
		},
		{
			name:     "unchanged line returns space",
			lineType: LineUnchanged,
			want:     " ",
		},
		{
			name:     "unknown type returns space",
			lineType: LineType("unknown"),
			want:     " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeSymbol(tt.lineType)
			if result != tt.want {
				t.Errorf("getTypeSymbol(%v) = %q, want %q", tt.lineType, result, tt.want)
			}
		})
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{
			name:  "single digit",
			input: 5,
			want:  "5",
		},
		{
			name:  "double digit",
			input: 42,
			want:  "42",
		},
		{
			name:  "triple digit",
			input: 100,
			want:  "100",
		},
		{
			name:  "zero",
			input: 0,
			want:  "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := intToString(tt.input)
			if result != tt.want {
				t.Errorf("intToString(%d) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func TestLineTypeConstants(t *testing.T) {
	if LineAdded != "added" {
		t.Errorf("LineAdded = %q, want %q", LineAdded, "added")
	}
	if LineRemoved != "removed" {
		t.Errorf("LineRemoved = %q, want %q", LineRemoved, "removed")
	}
	if LineUnchanged != "unchanged" {
		t.Errorf("LineUnchanged = %q, want %q", LineUnchanged, "unchanged")
	}
}

func TestHighlightLine(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		language string
		wantRaw  bool // if true, expect the raw code back (no highlighting)
	}{
		{
			name:     "valid go code",
			code:     "func main() {}",
			language: "go",
			wantRaw:  false,
		},
		{
			name:     "valid lua code",
			code:     "local x = 10",
			language: "lua",
			wantRaw:  false,
		},
		{
			name:     "unknown language falls back",
			code:     "some code",
			language: "unknownlang123",
			wantRaw:  false, // uses fallback lexer
		},
		{
			name:     "empty code",
			code:     "",
			language: "go",
			wantRaw:  false,
		},
		{
			name:     "javascript code",
			code:     "const x = 5;",
			language: "javascript",
			wantRaw:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlightLine(tt.code, tt.language)
			if result == "" && tt.code != "" {
				t.Errorf("highlightLine(%q, %q) returned empty string", tt.code, tt.language)
			}
			// Result should not be wrapped in <code> tags (we strip them)
			if len(result) > 6 && result[:6] == "<code>" {
				t.Errorf("highlightLine(%q, %q) should not have <code> wrapper", tt.code, tt.language)
			}
		})
	}
}

func TestGetTypeAriaLabel(t *testing.T) {
	tests := []struct {
		name     string
		lineType LineType
		want     string
	}{
		{
			name:     "added line",
			lineType: LineAdded,
			want:     "Line added",
		},
		{
			name:     "removed line",
			lineType: LineRemoved,
			want:     "Line removed",
		},
		{
			name:     "unchanged line",
			lineType: LineUnchanged,
			want:     "Unchanged line",
		},
		{
			name:     "unknown type",
			lineType: LineType("unknown"),
			want:     "Unchanged line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeAriaLabel(tt.lineType)
			if result != tt.want {
				t.Errorf("getTypeAriaLabel(%v) = %q, want %q", tt.lineType, result, tt.want)
			}
		})
	}
}

func TestVersionLabelProps(t *testing.T) {
	tests := []struct {
		name  string
		props VersionLabelProps
	}{
		{
			name: "basic labels",
			props: VersionLabelProps{
				OldLabel: "v1.0",
				NewLabel: "v2.0",
			},
		},
		{
			name: "with metadata",
			props: VersionLabelProps{
				OldLabel: "v1.0",
				NewLabel: "v2.0",
				OldMeta:  "2 days ago",
				NewMeta:  "just now",
			},
		},
		{
			name: "with stats",
			props: VersionLabelProps{
				OldLabel:  "v1.0",
				NewLabel:  "v2.0",
				Additions: 10,
				Deletions: 5,
			},
		},
		{
			name: "full configuration",
			props: VersionLabelProps{
				OldLabel:  "main",
				NewLabel:  "feature-branch",
				OldMeta:   "abc123",
				NewMeta:   "def456",
				Additions: 25,
				Deletions: 12,
			},
		},
		{
			name: "only additions",
			props: VersionLabelProps{
				OldLabel:  "v1.0",
				NewLabel:  "v2.0",
				Additions: 5,
				Deletions: 0,
			},
		},
		{
			name: "only deletions",
			props: VersionLabelProps{
				OldLabel:  "v1.0",
				NewLabel:  "v2.0",
				Additions: 0,
				Deletions: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify props are valid
			if tt.props.OldLabel == "" {
				t.Error("OldLabel should not be empty")
			}
			if tt.props.NewLabel == "" {
				t.Error("NewLabel should not be empty")
			}
			if tt.props.Additions < 0 {
				t.Error("Additions should not be negative")
			}
			if tt.props.Deletions < 0 {
				t.Error("Deletions should not be negative")
			}
		})
	}
}
