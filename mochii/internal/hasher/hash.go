// Package hasher provides SHA-256 hashing functionality.
package hasher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// Hash represents a SHA-256 content hash (64 hex characters).
type Hash string

// Parse validates a hash string (must be exactly 64 hex characters).
func Parse(s string) (Hash, error) {
	if len(s) != 64 {
		return "", fmt.Errorf("invalid hash length: %d", len(s))
	}
	return Hash(s), nil
}

// FromString computes SHA-256 hash of a string.
func FromString(s string) Hash {
	h := sha256.Sum256([]byte(s))
	return Hash(fmt.Sprintf("%x", h))
}

// FromFile computes SHA-256 hash of a file's contents.
func FromFile(path string) (Hash, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close file: %w", err)
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("copy file: %w", err)
	}

	return Hash(fmt.Sprintf("%x", h.Sum(nil))), nil
}

// String returns the hash as a string.
func (h Hash) String() string {
	return string(h)
}

// IsValid checks if the hash is the correct length (64 characters).
func (h Hash) IsValid() bool {
	return len(h) == 64
}
