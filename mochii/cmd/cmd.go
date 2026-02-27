package cmd

import (
	"fmt"
	"os"

	"github.com/0x0Dx/x/mochii/internal/build"
	"github.com/0x0Dx/x/mochii/internal/db"
	"github.com/0x0Dx/x/mochii/internal/hash"
	"github.com/0x0Dx/x/mochii/internal/prebuilts"
	"github.com/0x0Dx/x/mochii/internal/profile"
	"github.com/0x0Dx/x/mochii/internal/util"
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
	home := util.Getenv("MOCHII_HOME", "/var/mochii")
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
		if err := util.EnsureDir(d); err != nil {
			return err
		}
	}
	return nil
}

// CLI wraps all mochii components.
type CLI struct {
	cfg      *Config
	db       *db.DB
	builder  *build.Builder
	profiler *profile.Profile
	prebuilt *prebuilts.Prebuilt
}

// New creates a new CLI with initialized components.
func New() (*CLI, error) {
	cfg := DefaultConfig()

	if err := EnsureDirs(cfg); err != nil {
		return nil, err
	}

	database, err := db.EnsureDB(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	b := build.New(database, cfg.SourcesDir, cfg.HomeDir+"/pkg", cfg.LogDir)
	p := profile.New(cfg.ProfileDir)
	pre := prebuilts.New(b, cfg.PrebuiltsDir)

	return &CLI{cfg, database, b, p, pre}, nil
}

// Close closes the CLI and underlying resources.
func (c *CLI) Close() error {
	return c.db.Close()
}

// GetPkg builds and installs a package by hash.
func (c *CLI) GetPkg(h string) error {
	hash, err := hash.Parse(h)
	if err != nil {
		return err
	}

	path, err := c.builder.Install(hash)
	if err != nil {
		return err
	}

	fmt.Println(path)
	return nil
}

// Run executes an installed package.
func (c *CLI) Run(h string, args []string) error {
	hash, err := hash.Parse(h)
	if err != nil {
		return err
	}

	return c.builder.Run(hash, args)
}

// Delete removes an installed package.
func (c *CLI) Delete(h string) error {
	hash, err := hash.Parse(h)
	if err != nil {
		return err
	}

	return c.builder.Delete(hash)
}

// ListInstalled lists all installed packages.
func (c *CLI) ListInstalled() error {
	pkgs, err := c.builder.ListInstalled()
	if err != nil {
		return err
	}

	for h, path := range pkgs {
		fmt.Printf("%s %s\n", h, path)
	}

	return nil
}

// RegisterFile registers a source file by hash.
func (c *CLI) RegisterFile(path string) error {
	h, err := c.builder.RegisterFile(path)
	if err != nil {
		return err
	}

	fmt.Println(h)
	return nil
}

// RegisterURL registers a network URL for a package hash.
func (c *CLI) RegisterURL(hStr, url string) error {
	h, err := hash.Parse(hStr)
	if err != nil {
		return err
	}

	return c.builder.RegisterURL(h, url)
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
		return err
	}

	for h, path := range pkgs {
		if !util.DirExists(path) {
			fmt.Printf("package %s at %s is missing\n", h, path)
			if err := c.builder.Delete(hash.Hash(h)); err != nil {
				fmt.Fprintf(os.Stderr, "warning: delete failed: %v\n", err)
			}
		}
	}

	return nil
}

// SwitchProfile switches to a package in the profile.
func (c *CLI) SwitchProfile(h, pkgPath string) error {
	hash, err := hash.Parse(h)
	if err != nil {
		return err
	}

	return c.profiler.Switch(hash, pkgPath)
}

// ListProfiles lists all profile generations.
func (c *CLI) ListProfiles() error {
	gens, err := c.profiler.ListGenerations()
	if err != nil {
		return err
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
		return err
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
		if err := c.builder.Delete(hash.Hash(h)); err != nil {
			fmt.Fprintf(os.Stderr, "warning: delete failed: %v\n", err)
		}
	}

	return nil
}

// PullPrebuilts pulls prebuilt packages from URLs.
func (c *CLI) PullPrebuilts(urls []string) error {
	config := &prebuilts.PrebuiltConfig{URLs: urls}
	return c.prebuilt.Pull(config)
}

// PushPrebuilts exports installed packages as prebuilts.
func (c *CLI) PushPrebuilts(exportDir string) error {
	return c.prebuilt.Push(exportDir)
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
