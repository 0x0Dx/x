// Package profile manages profile generations.
package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/0x0Dx/x/mochii/internal/hasher"
)

// Profile manages profile generations.
type Profile struct {
	Path string
}

// New creates a new Profile at the given path.
func New(path string) *Profile {
	return &Profile{Path: path}
}

// Generation represents a single profile generation.
type Generation struct {
	Num  int
	Link string // Path to generation directory
	Hash string // Package hash
	Path string // Package path
}

// Switch creates a new generation pointing to a package.
// Creates symlinks to all executables in the package's bin directory.
func (p *Profile) Switch(h hasher.Hash, pkgPath string) error {
	if err := os.MkdirAll(p.Path, 0o750); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}

	num := p.nextNum()

	genDir := fmt.Sprintf("%s/%d", p.Path, num)
	binDir := genDir + "/bin"

	if err := os.MkdirAll(binDir, 0o750); err != nil {
		return fmt.Errorf("create bin dir: %w", err)
	}

	if err := symlinkExecutables(pkgPath, binDir); err != nil {
		return fmt.Errorf("symlink executables: %w", err)
	}

	hashFile := genDir + ".hash"
	f, err := os.Create(hashFile)
	if err != nil {
		return fmt.Errorf("create hash file: %w", err)
	}
	if _, err := fmt.Fprintf(f, "%s\n", h.String()); err != nil {
		return fmt.Errorf("write hash: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close hash file: %w", err)
	}

	current := p.Path + "/current"
	tmpLink := p.Path + "/new_current"

	if err := os.Symlink(genDir, tmpLink); err != nil {
		return fmt.Errorf("symlink new current: %w", err)
	}

	if err := os.Rename(tmpLink, current); err != nil {
		return fmt.Errorf("rename current: %w", err)
	}

	oldLink, err := os.Readlink(current)
	if err == nil && oldLink != genDir {
		if err := os.RemoveAll(oldLink); err != nil {
			fmt.Fprintf(os.Stderr, "warning: remove old link: %v\n", err)
		}
		if err := os.Remove(oldLink + ".hash"); err != nil {
			fmt.Fprintf(os.Stderr, "warning: remove old hash: %v\n", err)
		}
	}

	fmt.Printf("switched to %s\n", current)
	return nil
}

func symlinkExecutables(pkgPath, binDir string) error {
	link := func(src, dst string) error {
		info, err := os.Lstat(dst)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				if err := os.Remove(dst); err != nil {
					return fmt.Errorf("remove dst: %w", err)
				}
			}
		}
		if err := os.Symlink(src, dst); err != nil {
			return fmt.Errorf("symlink: %w", err)
		}
		return nil
	}

	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return fmt.Errorf("read pkg dir: %w", err)
	}

	for _, e := range entries {
		name := e.Name()
		src := filepath.Join(pkgPath, name)

		if !e.IsDir() {
			info, err := e.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0o111 != 0 && name != "builder" {
				if err := link(src, filepath.Join(binDir, name)); err != nil {
					fmt.Fprintf(os.Stderr, "warning: link failed: %v\n", err)
				}
			}
			continue
		}

		subdir := filepath.Join(binDir, name)
		if err := os.MkdirAll(subdir, 0o750); err != nil {
			continue
		}
		subEntries, err := os.ReadDir(src)
		if err != nil {
			continue
		}
		for _, se := range subEntries {
			if se.IsDir() {
				continue
			}
			seName := se.Name()
			seSrc := filepath.Join(src, seName)
			info, err := se.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0o111 != 0 {
				if err := link(seSrc, filepath.Join(subdir, seName)); err != nil {
					fmt.Fprintf(os.Stderr, "warning: link failed: %v\n", err)
				}
			}
		}
	}

	return nil
}

func (p *Profile) nextNum() int {
	entries, err := os.ReadDir(p.Path)
	if err != nil {
		return 0
	}

	maxNum := 0
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".hash") {
			var num int
			if _, err := fmt.Sscanf(name, "%d.hash", &num); err != nil {
				continue
			}
			if num > maxNum {
				maxNum = num
			}
		}
	}
	return maxNum + 1
}

// ListGenerations returns all profile generations, sorted by number.
func (p *Profile) ListGenerations() ([]Generation, error) {
	entries, err := os.ReadDir(p.Path)
	if err != nil {
		return nil, fmt.Errorf("read profile dir: %w", err)
	}

	var gens []Generation
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".hash") {
			path := filepath.Join(p.Path, name)
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			hashStr := strings.TrimSpace(string(data))

			genName := strings.TrimSuffix(name, ".hash")
			genPath := filepath.Join(p.Path, genName)

			var num int
			if _, err := fmt.Sscanf(genName, "%d", &num); err != nil {
				continue
			}

			gens = append(gens, Generation{
				Num:  num,
				Link: genPath,
				Hash: hashStr,
				Path: genPath,
			})
		}
	}

	sort.Slice(gens, func(i, j int) bool {
		return gens[i].Num < gens[j].Num
	})

	return gens, nil
}

// Current returns the path to the current generation's bin directory.
func (p *Profile) Current() (string, error) {
	current := p.Path + "/current"
	link, err := os.Readlink(current)
	if err != nil {
		return "", fmt.Errorf("readlink current: %w", err)
	}
	return link + "/bin", nil
}

// DeleteGeneration removes a profile generation.
func (p *Profile) DeleteGeneration(num int) error {
	gens, err := p.ListGenerations()
	if err != nil {
		return fmt.Errorf("list generations: %w", err)
	}

	for _, g := range gens {
		if g.Num == num {
			if err := os.Remove(g.Link); err != nil {
				return fmt.Errorf("remove link: %w", err)
			}
			if err := os.Remove(g.Link + ".hash"); err != nil {
				return fmt.Errorf("remove hash: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("generation %d not found", num)
}
