package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	content := `project: test
version: v1.0.0

builds:
  - id: main
    goos: [linux, darwin]
    goarch: [amd64]
    main: ./cmd/app
    binary: myapp

archives:
  - id: default
    builds: [main]
    format: tar.gz

checksums:
  - ids: [default]
`

	tmpfile, err := os.CreateTemp("", "releasex-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Project != "test" {
		t.Errorf("Project = %q, want %q", cfg.Project, "test")
	}
	if cfg.Version != "v1.0.0" {
		t.Errorf("Version = %q, want %q", cfg.Version, "v1.0.0")
	}
	if len(cfg.Builds) != 1 {
		t.Errorf("Builds length = %d, want 1", len(cfg.Builds))
	}
	if cfg.Builds[0].ID != "main" {
		t.Errorf("Builds[0].ID = %q, want %q", cfg.Builds[0].ID, "main")
	}
	if len(cfg.Archives) != 1 {
		t.Errorf("Archives length = %d, want 1", len(cfg.Archives))
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Error("Load() expected error for nonexistent file")
	}
}
