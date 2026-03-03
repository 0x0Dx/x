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

func Generate(files []string, output, basePath string) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	absBase, _ := filepath.Abs(basePath)
	for _, file := range files {
		hash, err := fileHash(file)
		if err != nil {
			return err
		}
		relPath := file
		if absFile, err := filepath.Abs(file); err == nil {
			if strings.HasPrefix(absFile, absBase+"/") {
				relPath = strings.TrimPrefix(absFile, absBase+"/")
			}
		}
		fmt.Fprintf(f, "%s  %s\n", hash, relPath)
	}

	fmt.Printf("Created checksums: %s\n", output)
	return nil
}

func fileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
