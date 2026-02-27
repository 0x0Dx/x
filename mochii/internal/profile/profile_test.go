package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0x0Dx/x/mochii/internal/hasher"
)

func TestProfileNextNum(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "mochii-profile-test")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(tmpDir, 0o755)

	p := New(tmpDir)

	num := p.nextNum()
	if num != 1 {
		t.Errorf("expected first generation to be 1, got %d", num)
	}

	// Create hash files
	os.WriteFile(filepath.Join(tmpDir, "1.hash"), []byte("abc"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "3.hash"), []byte("def"), 0o644)

	num = p.nextNum()
	if num != 4 {
		t.Errorf("expected next generation to be 4, got %d", num)
	}
}

func TestProfileListGenerations(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "mochii-profile-test2")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(tmpDir, 0o755)

	p := New(tmpDir)

	gens, err := p.ListGenerations()
	if err != nil {
		t.Fatalf("ListGenerations failed: %v", err)
	}
	if len(gens) != 0 {
		t.Errorf("expected 0 generations, got %d", len(gens))
	}

	// Create generation files
	os.MkdirAll(filepath.Join(tmpDir, "1"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "1.hash"), []byte("abc123"), 0o644)

	gens, err = p.ListGenerations()
	if err != nil {
		t.Fatalf("ListGenerations failed: %v", err)
	}
	if len(gens) != 1 {
		t.Errorf("expected 1 generation, got %d", len(gens))
	}
	if gens[0].Hash != "abc123" {
		t.Errorf("expected hash abc123, got %s", gens[0].Hash)
	}
}

func TestProfileCurrent(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "mochii-profile-test3")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(tmpDir, 0o755)

	p := New(tmpDir)

	_, err := p.Current()
	if err == nil {
		t.Error("expected error when no current generation")
	}

	// Create current symlink
	os.Symlink(filepath.Join(tmpDir, "1"), filepath.Join(tmpDir, "current"))

	current, err := p.Current()
	if err != nil {
		t.Fatalf("Current failed: %v", err)
	}
	if current != filepath.Join(tmpDir, "1", "bin") {
		t.Errorf("expected %s, got %s", filepath.Join(tmpDir, "1", "bin"), current)
	}
}

func TestProfileDeleteGeneration(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "mochii-profile-test4")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(tmpDir, 0o755)

	p := New(tmpDir)

	// Create generation
	os.MkdirAll(filepath.Join(tmpDir, "1"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "1.hash"), []byte("abc123"), 0o644)

	err := p.DeleteGeneration(1)
	if err != nil {
		t.Errorf("DeleteGeneration failed: %v", err)
	}

	gens, _ := p.ListGenerations()
	if len(gens) != 0 {
		t.Errorf("expected 0 generations after delete, got %d", len(gens))
	}

	// Test delete non-existent
	err = p.DeleteGeneration(999)
	if err == nil {
		t.Error("expected error deleting non-existent generation")
	}
}

func TestProfileSwitch(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "mochii-profile-test5")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(tmpDir, 0o755)

	// Create a mock package directory with executable
	pkgDir := filepath.Join(tmpDir, "mock-pkg")
	os.MkdirAll(filepath.Join(pkgDir, "bin"), 0o755)
	os.WriteFile(filepath.Join(pkgDir, "bin", "test"), []byte("test"), 0o755)

	p := New(tmpDir)

	h := hasher.Hash("abc123def456abc123def456abc123def456abc123def456abc123def456abc")
	err := p.Switch(h, pkgDir)
	if err != nil {
		t.Fatalf("Switch failed: %v", err)
	}

	// Check generation was created
	gens, err := p.ListGenerations()
	if err != nil {
		t.Fatalf("ListGenerations failed: %v", err)
	}
	if len(gens) != 1 {
		t.Errorf("expected 1 generation, got %d", len(gens))
	}

	// Check current symlink
	current, err := p.Current()
	if err != nil {
		t.Fatalf("Current failed: %v", err)
	}
	if current == "" {
		t.Error("expected current to be set")
	}
}
