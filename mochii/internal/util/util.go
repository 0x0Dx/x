package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BaseNameOf(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func AbsPath(p string) (string, error) {
	if filepath.IsAbs(p) {
		return p, nil
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func FileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func DirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func EnsureDir(p string) error {
	return os.MkdirAll(p, 0755)
}

func Getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func FileName(p string) string {
	return filepath.Base(p)
}

func DirName(p string) string {
	return filepath.Dir(p)
}

func Join(p ...string) string {
	return filepath.Join(p...)
}

type Error struct {
	msg string
}

func (e Error) Error() string {
	return e.msg
}

func Errorf(format string, args ...interface{}) Error {
	return Error{msg: fmt.Sprintf(format, args...)}
}
