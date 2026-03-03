package checksums

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func Generate(files []string, output string) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, file := range files {
		hash, err := fileHash(file)
		if err != nil {
			return err
		}
		fmt.Fprintf(f, "%s  %s\n", hash, file)
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
