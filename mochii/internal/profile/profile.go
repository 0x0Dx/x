package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/0x0Dx/x/mochii/internal/hash"
)

type Profile struct {
	Path string
}

func New(path string) *Profile {
	return &Profile{Path: path}
}

type Generation struct {
	Num  int
	Link string
	Hash string
	Path string
}

func (p *Profile) Switch(h hash.Hash, pkgPath string) error {
	if err := os.MkdirAll(p.Path, 0755); err != nil {
		return err
	}

	num := p.nextNum()

	genDir := fmt.Sprintf("%s/%d", p.Path, num)
	binDir := genDir + "/bin"

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	if err := symlinkExecutables(pkgPath, binDir); err != nil {
		return err
	}

	hashFile := genDir + ".hash"
	f, err := os.Create(hashFile)
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "%s\n", h.String())
	f.Close()

	current := p.Path + "/current"
	tmpLink := p.Path + "/new_current"

	if err := os.Symlink(genDir, tmpLink); err != nil {
		return err
	}

	if err := os.Rename(tmpLink, current); err != nil {
		return err
	}

	oldLink, err := os.Readlink(current)
	if err == nil && oldLink != genDir {
		os.RemoveAll(oldLink)
		os.Remove(oldLink + ".hash")
	}

	fmt.Printf("switched to %s\n", current)
	return nil
}

func symlinkExecutables(pkgPath, binDir string) error {
	var link func(string, string) error
	link = func(src, dst string) error {
		info, err := os.Lstat(dst)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				os.Remove(dst)
			}
		}
		return os.Symlink(src, dst)
	}

	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return err
	}

	for _, e := range entries {
		name := e.Name()
		src := filepath.Join(pkgPath, name)

		if e.IsDir() {
			subdir := filepath.Join(binDir, name)
			if err := os.MkdirAll(subdir, 0755); err != nil {
				continue
			}
			subEntries, err := os.ReadDir(src)
			if err != nil {
				continue
			}
			for _, se := range subEntries {
				seName := se.Name()
				seSrc := filepath.Join(src, seName)
				if !se.IsDir() {
					info, err := se.Info()
					if err != nil {
						continue
					}
					if info.Mode()&0111 != 0 {
						link(seSrc, filepath.Join(subdir, seName))
					}
				}
			}
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0111 != 0 && name != "builder" {
			link(src, filepath.Join(binDir, name))
		}
	}

	return nil
}

func (p *Profile) nextNum() int {
	entries, err := os.ReadDir(p.Path)
	if err != nil {
		return 0
	}

	max := 0
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".hash") {
			var num int
			fmt.Sscanf(name, "%d.hash", &num)
			if num > max {
				max = num
			}
		}
	}
	return max + 1
}

func (p *Profile) ListGenerations() ([]Generation, error) {
	entries, err := os.ReadDir(p.Path)
	if err != nil {
		return nil, err
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
			fmt.Sscanf(genName, "%d", &num)

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

func (p *Profile) Current() (string, error) {
	current := p.Path + "/current"
	link, err := os.Readlink(current)
	if err != nil {
		return "", err
	}
	return link + "/bin", nil
}

func (p *Profile) DeleteGeneration(num int) error {
	gens, err := p.ListGenerations()
	if err != nil {
		return err
	}

	for _, g := range gens {
		if g.Num == num {
			os.Remove(g.Link)
			os.Remove(g.Link + ".hash")
			return nil
		}
	}

	return fmt.Errorf("generation %d not found", num)
}
