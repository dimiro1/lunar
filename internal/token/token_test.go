package token

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	tok, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if len(tok) != 64 {
		t.Errorf("expected 64-char hex string, got %d chars", len(tok))
	}

	// Should produce unique tokens
	tok2, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	if tok == tok2 {
		t.Error("Generate() produced identical tokens")
	}
}

func TestHash(t *testing.T) {
	hash := Hash("test-token")

	if len(hash) != 64 {
		t.Errorf("expected 64-char hex hash, got %d chars", len(hash))
	}

	// Same input should produce same hash
	if hash != Hash("test-token") {
		t.Error("Hash() is not deterministic")
	}

	// Different input should produce different hash
	if hash == Hash("other-token") {
		t.Error("Hash() produced same hash for different inputs")
	}
}
