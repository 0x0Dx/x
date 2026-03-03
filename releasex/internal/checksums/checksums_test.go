package checksums

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "releasex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}

	checksumFile := filepath.Join(tmpDir, "checksums.txt")
	if err := Generate([]string{testFile}, checksumFile, tmpDir); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	data, err := os.ReadFile(checksumFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "test.txt") {
		t.Error("checksum file should contain filename")
	}
}

func TestFileHash(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "releasex-hash-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("test content")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	hash, err := fileHash(tmpFile.Name())
	if err != nil {
		t.Fatalf("fileHash() error = %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("SHA256 hash should be 64 chars, got %d", len(hash))
	}
}
