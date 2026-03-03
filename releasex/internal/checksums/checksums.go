// Package checksums generates SHA256 checksums for files.
package checksums

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Generate creates a checksums file for the given files.
func Generate(files []string, output, basePath string) error {
	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", output, err)
	}
	defer func() { _ = f.Close() }()

	absBase, _ := filepath.Abs(basePath)
	for _, file := range files {
		hash, err := fileHash(file)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", file, err)
		}
		relPath := file
		if absFile, err := filepath.Abs(file); err == nil {
			if after, ok := strings.CutPrefix(absFile, absBase+"/"); ok {
				relPath = after
			}
		}
		_, _ = fmt.Fprintf(f, "%s  %s\n", hash, relPath)
	}

	fmt.Printf("Created checksums: %s\n", output)
	return nil
}

func fileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash %s: %w", path, err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
