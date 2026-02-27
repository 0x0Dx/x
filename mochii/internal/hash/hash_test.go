package hash

import (
	"testing"
)

func TestHashFromString(t *testing.T) {
	h := FromString("hello")
	if len(h) != 64 {
		t.Errorf("expected hash length 64, got %d", len(h))
	}
}

func TestHashFromFile(t *testing.T) {
	h := FromString("test content")
	h2 := FromString("test content")
	if h != h2 {
		t.Errorf("same content should produce same hash")
	}
}

func TestHashParse(t *testing.T) {
	h := FromString("test")

	parsed, err := Parse(h.String())
	if err != nil {
		t.Errorf("parse failed: %v", err)
	}
	if parsed != h {
		t.Errorf("parsed hash doesn't match")
	}
}

func TestHashParseInvalid(t *testing.T) {
	_, err := Parse("too-short")
	if err == nil {
		t.Error("expected error for invalid hash")
	}
}

func TestHashIsValid(t *testing.T) {
	h := FromString("test")
	if !h.IsValid() {
		t.Error("expected valid hash")
	}

	invalid := Hash("short")
	if invalid.IsValid() {
		t.Error("expected invalid hash")
	}
}
