// Package hasher provides SHA-256 hashing functionality.
package hasher

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
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

// FromPath computes SHA-256 hash of a path (file or directory).
// This is the key content-addressable hash - it includes all contents
// recursively, making the hash depend on the full path structure.
func FromPath(path string) (Hash, error) {
	h := sha256.New()
	sink := &hashSink{h: h}

	if err := dumpPath(path, sink); err != nil {
		return "", fmt.Errorf("dump path: %w", err)
	}

	return Hash(fmt.Sprintf("%x", h.Sum(nil))), nil
}

type hashSink struct {
	h hash.Hash
}

func (s *hashSink) Write(data []byte) (int, error) {
	s.h.Write(data)
	return len(data), nil
}

func dumpPath(path string, sink io.Writer) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("lstat %s: %w", path, err)
	}

	writeString("(", sink)
	writeString("type", sink)

	if info.IsDir() {
		writeString("directory", sink)
		writeString("entries", sink)
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("read dir %s: %w", path, err)
		}

		// Sort entries by name
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		sort.Strings(names)

		writeString("(", sink)
		for i, name := range names {
			if i > 0 {
				writeString(",", sink)
			}
			writeString("(", sink)
			writeString("name", sink)
			writeString(name, sink)
			writeString("node", sink)
			subPath := filepath.Join(path, name)
			if err := dumpPath(subPath, sink); err != nil {
				return err
			}
			writeString(")", sink)
		}
		writeString(")", sink)
	} else if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(path)
		if err != nil {
			return fmt.Errorf("readlink %s: %w", path, err)
		}
		writeString("symlink", sink)
		writeString("target", sink)
		writeString(link, sink)
	} else {
		writeString("regular", sink)
		writeString("contents", sink)
		contentsHash, err := hashFileContents(path)
		if err != nil {
			return err
		}
		writeHash(contentsHash, sink)
	}

	writeString(")", sink)
	return nil
}

func hashFileContents(path string) (Hash, error) {
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

func writeString(s string, w io.Writer) {
	writeInt(len(s), w)
	writePad(len(s), w)
	io.WriteString(w, s)
}

func writeInt(n int, w io.Writer) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(n))
	w.Write(buf[:])
}

func writePad(n int, w io.Writer) {
	pad := (8 - (n % 8)) % 8
	if pad > 0 {
		var buf [8]byte
		w.Write(buf[:pad])
	}
}

func writeHash(h Hash, w io.Writer) {
	io.WriteString(w, string(h))
}
