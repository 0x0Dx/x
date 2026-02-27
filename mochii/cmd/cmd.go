// Package cmd provides the CLI commands for mochii.
package cmd

import (
	"fmt"
	"os"

	"github.com/0x0Dx/x/mochii/internal/builder"
	"github.com/0x0Dx/x/mochii/internal/db"
	"github.com/0x0Dx/x/mochii/internal/hasher"
	"github.com/0x0Dx/x/mochii/internal/helper"
	"github.com/0x0Dx/x/mochii/internal/prebuilts"
	"github.com/0x0Dx/x/mochii/internal/profile"
)

// Config holds mochii configuration paths.
type Config struct {
	HomeDir      string // Root mochii directory
	SourcesDir   string // Source tarballs
	LogDir       string // Build logs
	DBPath       string // SQLite database
	ProfileDir   string // Profile generations
	PrebuiltsDir string // Prebuilt packages
}

// DefaultConfig creates a Config with default paths.
// Uses MOCHII_HOME env var or /var/mochii as base.
func DefaultConfig() *Config {
	home := helper.Getenv("MOCHII_HOME", "/var/mochii")
	return &Config{
		HomeDir:      home,
		SourcesDir:   home + "/sources",
		LogDir:       home + "/logs",
		DBPath:       home + "/pkginfo.db",
		ProfileDir:   home + "/profiles",
		PrebuiltsDir: home + "/prebuilts",
	}
}

// EnsureDirs creates all required directories.
func EnsureDirs(cfg *Config) error {
	dirs := []string{cfg.HomeDir, cfg.SourcesDir, cfg.LogDir, cfg.ProfileDir, cfg.PrebuiltsDir}
	for _, d := range dirs {
		if err := helper.EnsureDir(d); err != nil {
			return fmt.Errorf("ensure dir %s: %w", d, err)
		}
	}
	return nil
}

// CLI wraps all mochii components.
type CLI struct {
	cfg      *Config
	db       *db.DB
	builder  *builder.Builder
	profiler *profile.Profile
	prebuilt *prebuilts.Prebuilt
}

// New creates a new CLI with initialized components.
func New() (*CLI, error) {
	cfg := DefaultConfig()

	if err := EnsureDirs(cfg); err != nil {
		return nil, fmt.Errorf("ensure dirs: %w", err)
	}

	database, err := db.EnsureDB(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("ensure db: %w", err)
	}

	b := builder.New(database, cfg.SourcesDir, cfg.HomeDir+"/pkg", cfg.LogDir)
	p := profile.New(cfg.ProfileDir)
	pre := prebuilts.New(b, cfg.PrebuiltsDir)

	return &CLI{cfg, database, b, p, pre}, nil
}

// Close closes the CLI and underlying resources.
func (c *CLI) Close() error {
	if err := c.db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}
	return nil
}

// GetPkg builds and installs a package by hash.
func (c *CLI) GetPkg(h string) error {
	pkgHash, err := hasher.Parse(h)
	if err != nil {
		return fmt.Errorf("parse hash: %w", err)
	}

	path, err := c.builder.Install(pkgHash)
	if err != nil {
		return fmt.Errorf("install package: %w", err)
	}

	fmt.Println(path)
	return nil
}

// Run executes an installed package.
func (c *CLI) Run(h string, args []string) error {
	pkgHash, err := hasher.Parse(h)
	if err != nil {
		return fmt.Errorf("parse hash: %w", err)
	}

	if err := c.builder.Run(pkgHash, args); err != nil {
		return fmt.Errorf("run package: %w", err)
	}
	return nil
}

// Delete removes an installed package.
func (c *CLI) Delete(h string) error {
	pkgHash, err := hasher.Parse(h)
	if err != nil {
		return fmt.Errorf("parse hash: %w", err)
	}

	if err := c.builder.Delete(pkgHash); err != nil {
		return fmt.Errorf("delete package: %w", err)
	}
	return nil
}

// ListInstalled lists all installed packages.
func (c *CLI) ListInstalled() error {
	pkgs, err := c.builder.ListInstalled()
	if err != nil {
		return fmt.Errorf("list installed: %w", err)
	}

	for h, path := range pkgs {
		fmt.Printf("%s %s\n", h, path)
	}

	return nil
}

// RegisterFile registers a source file by 	hasher.
func (c *CLI) RegisterFile(path string) error {
	h, err := c.builder.RegisterFile(path)
	if err != nil {
		return fmt.Errorf("register file: %w", err)
	}

	fmt.Println(h)
	return nil
}

// RegisterURL registers a network URL for a package 	hasher.
func (c *CLI) RegisterURL(hStr, url string) error {
	h, err := hasher.Parse(hStr)
	if err != nil {
		return fmt.Errorf("parse hash: %w", err)
	}

	if err := c.builder.RegisterURL(h, url); err != nil {
		return fmt.Errorf("register url: %w", err)
	}
	return nil
}

// Fetch fetches a URL (not fully implemented).
func (c *CLI) Fetch(url string) error {
	fmt.Println(url)
	return nil
}

// Verify checks that all installed packages still exist.
func (c *CLI) Verify() error {
	pkgs, err := c.builder.ListInstalled()
	if err != nil {
		return fmt.Errorf("list installed: %w", err)
	}

	for h, path := range pkgs {
		if !helper.DirExists(path) {
			fmt.Printf("package %s at %s is missing\n", h, path)
			if err := c.builder.Delete(hasher.Hash(h)); err != nil {
				fmt.Fprintf(os.Stderr, "warning: delete failed: %v\n", err)
			}
		}
	}

	return nil
}

// SwitchProfile switches to a package in the profile.
func (c *CLI) SwitchProfile(h, pkgPath string) error {
	pkgHash, err := hasher.Parse(h)
	if err != nil {
		return fmt.Errorf("parse hash: %w", err)
	}

	if err := c.profiler.Switch(pkgHash, pkgPath); err != nil {
		return fmt.Errorf("switch profile: %w", err)
	}
	return nil
}

// ListProfiles lists all profile generations.
func (c *CLI) ListProfiles() error {
	gens, err := c.profiler.ListGenerations()
	if err != nil {
		return fmt.Errorf("list generations: %w", err)
	}

	current, _ := c.profiler.Current()

	for _, g := range gens {
		marker := " "
		if g.Link == current {
			marker = "*"
		}
		fmt.Printf("%s %d %s -> %s\n", marker, g.Num, g.Hash, g.Path)
	}

	return nil
}

// CollectGarbage removes packages not referenced by any profile.
func (c *CLI) CollectGarbage() error {
	toDelete, err := c.builder.CollectGarbage(c.cfg.ProfileDir)
	if err != nil {
		return fmt.Errorf("collect garbage: %w", err)
	}

	if len(toDelete) == 0 {
		fmt.Println("no garbage to collect")
		return nil
	}

	fmt.Println("garbage:")
	for _, h := range toDelete {
		fmt.Printf("  %s\n", h)
	}

	for _, h := range toDelete {
		fmt.Printf("deleting %s...\n", h)
		if err := c.builder.Delete(hasher.Hash(h)); err != nil {
			fmt.Fprintf(os.Stderr, "warning: delete failed: %v\n", err)
		}
	}

	return nil
}

// PullPrebuilts pulls prebuilt packages from URLs.
func (c *CLI) PullPrebuilts(urls []string) error {
	config := &prebuilts.PrebuiltConfig{URLs: urls}
	if err := c.prebuilt.Pull(config); err != nil {
		return fmt.Errorf("pull prebuilts: %w", err)
	}
	return nil
}

// PushPrebuilts exports installed packages as prebuilts.
func (c *CLI) PushPrebuilts(exportDir string) error {
	if err := c.prebuilt.Push(exportDir); err != nil {
		return fmt.Errorf("push prebuilts: %w", err)
	}
	return nil
}

// PrintUsage prints the CLI usage message.
func PrintUsage() {
	fmt.Fprintf(os.Stderr, `Usage: mochii COMMAND [OPTIONS]...

Commands:
  init                  Initialize the database
  verify                Verify installed packages
  getpkg HASH           Install and return package path
  delpkg HASH           Remove installed package
  listinst              List installed packages
  run HASH [ARGS]       Run package
  regfile FILE          Register file by hash
  regurl HASH URL       Register URL for hash
  fetch URL             Fetch URL and print hash
  profile               List profile generations
  switch HASH PATH      Switch to package in profile
  gc                    Collect garbage
  pull URL...           Pull prebuilt packages
  push DIR              Push prebuilt packages to directory

`)
}
