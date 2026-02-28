// Package values provides value management (addValue, queryValuePath).
package values

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/0x0Dx/x/mochii/internal/db"
	"github.com/0x0Dx/x/mochii/internal/hasher"
)

// Manager manages values in the Nix store.
type Manager struct {
	DB        *db.DB
	ValuesDir string
}

// New creates a new Values manager.
func New(db *db.DB, valuesDir string) *Manager {
	return &Manager{
		DB:        db,
		ValuesDir: valuesDir,
	}
}

// AddValue adds a value (file) to the store and registers it.
// Returns the hash of the value.
func (m *Manager) AddValue(path string) (hasher.Hash, error) {
	hash, err := hasher.FromFile(path)
	if err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}

	// Check if already registered
	_, ok, err := m.DB.Get("refs", hash.String())
	if err != nil {
		return "", fmt.Errorf("query refs: %w", err)
	}
	if ok {
		return hash, nil
	}

	// Copy to values directory
	baseName := filepath.Base(path)
	targetName := fmt.Sprintf("%s-%s", hash, baseName)
	targetPath := filepath.Join(m.ValuesDir, targetName)

	if err := m.copyFile(path, targetPath); err != nil {
		return "", fmt.Errorf("copy file: %w", err)
	}

	// Register in database
	if err := m.DB.Set("refs", hash.String(), targetName); err != nil {
		return "", fmt.Errorf("set ref: %w", err)
	}

	return hash, nil
}

// QueryValuePath gets the path of a value by hash.
// Returns the absolute path to the value.
func (m *Manager) QueryValuePath(hash hasher.Hash) (string, error) {
	name, ok, err := m.DB.Get("refs", hash.String())
	if err != nil {
		return "", fmt.Errorf("query refs: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("value with hash %s not found", hash)
	}

	path := filepath.Join(m.ValuesDir, name)

	// Verify hash hasn't changed
	currentHash, err := hasher.FromFile(path)
	if err != nil {
		return "", fmt.Errorf("verify hash: %w", err)
	}

	if currentHash != hash {
		return "", fmt.Errorf("value %s is stale (hash mismatch)", path)
	}

	return path, nil
}

// AddValueFromReader adds a value from a reader.
func (m *Manager) AddValueFromReader(name string, r io.Reader) (hasher.Hash, error) {
	// Create temp file
	tmp, err := os.CreateTemp("", "mochii-value-*")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(tmp.Name())

	// Copy to temp and hash
	h := sha256.New()
	w := io.MultiWriter(tmp, h)

	if _, err := io.Copy(w, r); err != nil {
		tmp.Close()
		return "", fmt.Errorf("copy: %w", err)
	}
	tmp.Close()

	hash := hasher.Hash(fmt.Sprintf("%x", h.Sum(nil)))

	// Copy to values directory
	targetName := fmt.Sprintf("%s-%s", hash, name)
	targetPath := filepath.Join(m.ValuesDir, targetName)

	tmp2, err := os.Open(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("open temp: %w", err)
	}
	defer tmp2.Close()

	dest, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create dest: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, tmp2); err != nil {
		return "", fmt.Errorf("copy to dest: %w", err)
	}

	// Register
	m.DB.Set("refs", hash.String(), targetName)

	return hash, nil
}

// DeleteValue removes a value from the store.
func (m *Manager) DeleteValue(hash hasher.Hash) error {
	name, ok, err := m.DB.Get("refs", hash.String())
	if err != nil {
		return fmt.Errorf("query refs: %w", err)
	}
	if !ok {
		return nil
	}

	path := filepath.Join(m.ValuesDir, name)
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove value: %w", err)
	}

	if err := m.DB.Delete("refs", hash.String()); err != nil {
		return fmt.Errorf("delete ref: %w", err)
	}

	return nil
}

func (m *Manager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	return nil
}
