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
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return "", fmt.Errorf("invalid hash character: %c", c)
		}
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
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("copy file: %w", err)
	}

	return Hash(fmt.Sprintf("%x", h.Sum(nil))), nil
}

// FromReader computes SHA-256 hash of a reader.
func FromReader(r io.Reader) (Hash, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("copy reader: %w", err)
	}
	return Hash(fmt.Sprintf("%x", h.Sum(nil))), nil
}

// String returns the hash as a string.
func (h Hash) String() string {
	return string(h)
}

// IsValid checks if the hash is the correct length (64 characters).
func (h Hash) IsValid() bool {
	if len(h) != 64 {
		return false
	}
	for _, c := range h {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// IsHash checks if a string is a valid hash.
func IsHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
