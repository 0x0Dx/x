// Package helper provides utility functions.
package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BaseNameOf extracts the filename from a URL or path.
func BaseNameOf(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

// AbsPath returns the absolute path, converting relative paths to absolute.
func AbsPath(p string) (string, error) {
	if filepath.IsAbs(p) {
		return p, nil
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}
	return abs, nil
}

// FileExists checks if a file exists at the given path.
func FileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// DirExists checks if a directory exists at the given path.
func DirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// EnsureDir creates a directory and all parent directories if they don't exist.
func EnsureDir(p string) error {
	if err := os.MkdirAll(p, 0o750); err != nil {
		return fmt.Errorf("mkdir all: %w", err)
	}
	return nil
}

// Getenv returns the value of an environment variable, or a default if not set.
func Getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// FileName returns the base name of a file path.
func FileName(p string) string {
	return filepath.Base(p)
}

// DirName returns the directory name of a path.
func DirName(p string) string {
	return filepath.Dir(p)
}

// Join joins path elements into a single path.
func Join(p ...string) string {
	return filepath.Join(p...)
}

// Error represents a custom error with a message.
type Error struct {
	msg string
}

// Error returns the error message.
func (e Error) Error() string {
	return e.msg
}

// Errorf creates a new Error with a formatted message.
func Errorf(format string, args ...any) Error {
	return Error{msg: fmt.Sprintf(format, args...)}
}
