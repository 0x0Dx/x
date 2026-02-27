package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaseNameOf(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://example.com/file.tar.gz", "file.tar.gz"},
		{"/path/to/file.txt", "file.txt"},
		{"file", "file"},
	}

	for _, tt := range tests {
		got := BaseNameOf(tt.input)
		if got != tt.want {
			t.Errorf("BaseNameOf(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAbsPath(t *testing.T) {
	abs, err := AbsPath("/absolute/path")
	if err != nil {
		t.Errorf("AbsPath failed: %v", err)
	}
	if abs != "/absolute/path" {
		t.Errorf("AbsPath(/absolute/path) = %q, want /absolute/path", abs)
	}
}

func TestFileExists(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "mochii-test")

	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	if !FileExists(tmpFile) {
		t.Error("expected file to exist")
	}

	if FileExists("/nonexistent/file") {
		t.Error("expected file to not exist")
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := os.TempDir()

	if !DirExists(tmpDir) {
		t.Error("expected temp dir to exist")
	}

	if DirExists("/nonexistent/dir") {
		t.Error("expected dir to not exist")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "mochii-test-ensure")
	defer os.RemoveAll(tmpDir)

	if err := EnsureDir(tmpDir); err != nil {
		t.Errorf("EnsureDir failed: %v", err)
	}

	if !DirExists(tmpDir) {
		t.Error("expected dir to exist after EnsureDir")
	}
}

func TestGetenv(t *testing.T) {
	os.Setenv("MOCHII_TEST", "value")
	defer os.Unsetenv("MOCHII_TEST")

	if Getenv("MOCHII_TEST", "default") != "value" {
		t.Error("expected env value")
	}

	if Getenv("MOCHII_NOT_SET", "default") != "default" {
		t.Error("expected default value")
	}
}

func TestError(t *testing.T) {
	e := Error{msg: "test error"}
	if e.Error() != "test error" {
		t.Errorf("Error() = %q, want 'test error'", e.Error())
	}
}

func TestErrorf(t *testing.T) {
	e := Errorf("test %d", 123)
	if e.Error() != "test 123" {
		t.Errorf("Errorf() = %q, want 'test 123'", e.Error())
	}
}
