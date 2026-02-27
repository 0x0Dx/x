package hash

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

type Hash string

func Parse(s string) (Hash, error) {
	if len(s) != 64 {
		return "", fmt.Errorf("invalid hash length: %d", len(s))
	}
	return Hash(s), nil
}

func FromString(s string) Hash {
	h := sha256.Sum256([]byte(s))
	return Hash(fmt.Sprintf("%x", h))
}

func FromFile(path string) (Hash, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return Hash(fmt.Sprintf("%x", h.Sum(nil))), nil
}

func (h Hash) String() string {
	return string(h)
}

func (h Hash) IsValid() bool {
	return len(h) == 64
}
