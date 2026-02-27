package prebuilts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/0x0Dx/x/mochii/internal/build"
	"github.com/0x0Dx/x/mochii/internal/hash"
)

type Prebuilt struct {
	Builder *build.Builder
	Dir     string
}

func New(b *build.Builder, dir string) *Prebuilt {
	return &Prebuilt{Builder: b, Dir: dir}
}

type PrebuiltConfig struct {
	URLs []string
}

type PrebuiltEntry struct {
	Filename  string
	URL       string
	PackageID string
	PKGHash   string
	Hash      string
}

func (p *Prebuilt) Pull(config *PrebuiltConfig) error {
	if err := os.MkdirAll(p.Dir, 0755); err != nil {
		return err
	}

	for _, url := range config.URLs {
		if err := p.pullFromURL(url); err != nil {
			fmt.Fprintf(os.Stderr, "warning: pull from %s failed: %v\n", url, err)
		}
	}

	return nil
}

func (p *Prebuilt) pullFromURL(url string) error {
	fmt.Printf("obtaining prebuilt list from %s...\n", url)

	if strings.HasPrefix(url, "/") {
		return p.pullFromDir(url)
	}

	tmpFile := p.Dir + "/prebuilts.tmp"

	cmd := exec.Command("wget", "-q", "-O", tmpFile, url)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wget failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`href="([^"]*-([[:xdigit:]]{32})-([[:xdigit:]]{32})\.tar\.bz2)"`)
	matches := re.FindAllStringSubmatch(string(data), -1)

	baseURL := url
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}

		filename := m[1]
		pkgHash := m[2]
		prebuiltHash := m[3]

		entryURL := baseURL + filename

		fmt.Printf("registering prebuilt %s => %s\n", pkgHash, prebuiltHash)

		pkgHashH, err := hash.Parse(pkgHash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: invalid pkg hash: %v\n", err)
			continue
		}

		prebuiltHashH, err := hash.Parse(prebuiltHash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: invalid prebuilt hash: %v\n", err)
			continue
		}

		if err := p.Builder.RegisterPrebuilt(pkgHashH, prebuiltHashH); err != nil {
			fmt.Fprintf(os.Stderr, "warning: register prebuilt failed: %v\n", err)
			continue
		}

		if err := p.Builder.RegisterURL(prebuiltHashH, entryURL); err != nil {
			fmt.Fprintf(os.Stderr, "warning: register url failed: %v\n", err)
		}
	}

	os.Remove(tmpFile)
	return nil
}

func (p *Prebuilt) pullFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`^(.+)-([[:xdigit:]]{32})-([[:xdigit:]]{32})\.tar\.bz2$`)

	for _, e := range entries {
		name := e.Name()
		m := re.FindStringSubmatch(name)
		if m == nil {
			continue
		}

		id := m[1]
		pkgHash := m[2]
		prebuiltHash := m[3]

		fmt.Printf("registering prebuilt %s => %s (%s)\n", pkgHash, prebuiltHash, id)

		pkgHashH, _ := hash.Parse(pkgHash)
		prebuiltHashH, _ := hash.Parse(prebuiltHash)

		p.Builder.RegisterPrebuilt(pkgHashH, prebuiltHashH)

		fullPath := filepath.Join(dir, name)
		p.Builder.RegisterFile(fullPath)
	}

	return nil
}

func (p *Prebuilt) Push(exportDir string) error {
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return err
	}

	installed, err := p.Builder.ListInstalled()
	if err != nil {
		return err
	}

	prebuilts, err := p.Builder.ListPrebuilts()
	if err != nil {
		return err
	}

	for pkgHash := range installed {
		if _, ok := prebuilts[pkgHash]; !ok {
			fmt.Printf("exporting %s...\n", pkgHash)

			path, ok := installed[pkgHash]
			if !ok {
				continue
			}

			output := exportDir + "/export-" + pkgHash + ".tar.gz"

			cmd := exec.Command("tar", "czf", output, "-C", path, ".")
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: tar failed: %v\n", err)
				continue
			}
		}
	}

	fmt.Printf("prebuilts exported to %s\n", exportDir)
	return nil
}

func fetchURL(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
