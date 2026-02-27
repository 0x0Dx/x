package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0x0Dx/x/mochii/internal/helper"
)

func TestDB(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "mochii-test.db")
	os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	db, err := New(tmpFile)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Test Set and Get using refs table
	if err := db.Set("refs", "key", "value"); err != nil {
		t.Errorf("Set failed: %v", err)
	}

	val, ok, err := db.Get("refs", "key")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if !ok {
		t.Error("expected key to exist")
	}
	if val != "value" {
		t.Errorf("expected 'value', got %q", val)
	}
}

func TestDBGetNotFound(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "mochii-test2.db")
	os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	db, err := New(tmpFile)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	_, ok, err := db.Get("refs", "nonexistent")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if ok {
		t.Error("expected key to not exist")
	}
}

func TestDBDelete(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "mochii-test3.db")
	os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	db, err := New(tmpFile)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.Set("refs", "key", "value")
	db.Delete("refs", "key")

	_, ok, _ := db.Get("refs", "key")
	if ok {
		t.Error("expected key to be deleted")
	}
}

func TestDBList(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "mochii-test4.db")
	os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	db, err := New(tmpFile)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.Set("refs", "key1", "val1")
	db.Set("refs", "key2", "val2")

	m, err := db.List("refs")
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(m) != 2 {
		t.Errorf("expected 2 entries, got %d", len(m))
	}
	if m["key1"] != "val1" {
		t.Errorf("expected val1, got %q", m["key1"])
	}
}

func TestEnsureDB(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "mochii-test-ensure")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "test.db")

	db, err := EnsureDB(tmpFile)
	if err != nil {
		t.Fatalf("EnsureDB failed: %v", err)
	}
	defer db.Close()

	if !helper.FileExists(tmpFile) {
		t.Error("expected db file to exist")
	}
}
