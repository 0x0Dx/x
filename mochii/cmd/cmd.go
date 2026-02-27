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

type Config struct {
	HomeDir      string
	SourcesDir   string
	LogDir       string
	DBPath       string
	ProfileDir   string
	PrebuiltsDir string
}

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

func EnsureDirs(cfg *Config) error {
	dirs := []string{cfg.HomeDir, cfg.SourcesDir, cfg.LogDir, cfg.ProfileDir, cfg.PrebuiltsDir}
	for _, d := range dirs {
		if err := util.EnsureDir(d); err != nil {
			return err
		}
	}
	return nil
}

type CLI struct {
	cfg      *Config
	db       *db.DB
	builder  *build.Builder
	profiler *profile.Profile
	prebuilt *prebuilts.Prebuilt
}

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

func (c *CLI) Close() error {
	return c.db.Close()
}

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

func (c *CLI) Run(h string, args []string) error {
	hash, err := hash.Parse(h)
	if err != nil {
		return err
	}

	return c.builder.Run(hash, args)
}

func (c *CLI) Delete(h string) error {
	hash, err := hash.Parse(h)
	if err != nil {
		return err
	}

	return c.builder.Delete(hash)
}

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

func (c *CLI) RegisterFile(path string) error {
	h, err := c.builder.RegisterFile(path)
	if err != nil {
		return err
	}

	fmt.Println(h)
	return nil
}

func (c *CLI) RegisterURL(hStr, url string) error {
	h, err := hash.Parse(hStr)
	if err != nil {
		return err
	}

	return c.builder.RegisterURL(h, url)
}

func (c *CLI) Fetch(url string) error {
	fmt.Println(url)
	return nil
}

func (c *CLI) Verify() error {
	pkgs, err := c.builder.ListInstalled()
	if err != nil {
		return err
	}

	for h, path := range pkgs {
		if !util.DirExists(path) {
			fmt.Printf("package %s at %s is missing\n", h, path)
			c.builder.Delete(hash.Hash(h))
		}
	}

	return nil
}

func (c *CLI) SwitchProfile(h, pkgPath string) error {
	hash, err := hash.Parse(h)
	if err != nil {
		return err
	}

	return c.profiler.Switch(hash, pkgPath)
}

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

func (c *CLI) PullPrebuilts(urls []string) error {
	config := &prebuilts.PrebuiltConfig{URLs: urls}
	return c.prebuilt.Pull(config)
}

func (c *CLI) PushPrebuilts(exportDir string) error {
	return c.prebuilt.Push(exportDir)
}

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
