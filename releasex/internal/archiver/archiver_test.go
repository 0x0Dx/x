package archiver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "releasex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}

	zipFile := filepath.Join(tmpDir, "test.zip")
	if err := Create([]string{testFile}, "zip", zipFile); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := os.Stat(zipFile); os.IsNotExist(err) {
		t.Error("zip file was not created")
	}
}

func TestCreateTarGz(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "releasex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}

	tarFile := filepath.Join(tmpDir, "test.tar.gz")
	if err := Create([]string{testFile}, "tar.gz", tarFile); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := os.Stat(tarFile); os.IsNotExist(err) {
		t.Error("tar.gz file was not created")
	}
}

func TestCreateDetectsFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "releasex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	tarFile := filepath.Join(tmpDir, "test.tar.gz")
	if err := Create([]string{testFile}, "tar.gz", tarFile); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := os.Stat(tarFile); os.IsNotExist(err) {
		t.Error("tar.gz file was not created")
	}
}
