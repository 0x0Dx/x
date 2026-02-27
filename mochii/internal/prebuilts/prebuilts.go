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

// Prebuilt handles pulling and pushing binary substitutes.
type Prebuilt struct {
	Builder *build.Builder
	Dir     string // Local prebuilts directory
}

// New creates a new Prebuilt handler.
func New(b *build.Builder, dir string) *Prebuilt {
	return &Prebuilt{Builder: b, Dir: dir}
}

// PrebuiltConfig holds configuration for pulling prebuilts.
type PrebuiltConfig struct {
	URLs []string // URLs or directories to pull from
}

// PrebuiltEntry represents a prebuilt package entry.
type PrebuiltEntry struct {
	Filename  string
	URL       string
	PackageID string
	PKGHash   string
	Hash      string
}

// Pull fetches prebuilt packages from URLs or directories.
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

// pullFromURL fetches prebuilt index from a URL or local directory.
func (p *Prebuilt) pullFromURL(url string) error {
	fmt.Printf("obtaining prebuilt list from %s...\n", url)

	// Handle local directories
	if strings.HasPrefix(url, "/") {
		return p.pullFromDir(url)
	}

	// Download index file
	tmpFile := p.Dir + "/prebuilts.tmp"

	cmd := exec.Command("wget", "-q", "-O", tmpFile, url)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wget failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return err
	}

	// Parse HTML for links matching pattern: name-HASH-HASH.tar.bz2
	re := regexp.MustCompile(`href="([^"]*-([[:xdigit:]]{32})-([[:xdigit:]]{32})\.tar\.bz2)"`)
	matches := re.FindAllStringSubmatch(string(data), -1)

	baseURL := url
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	// Register each prebuilt
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

// pullFromDir reads prebuilt packages from a local directory.
func (p *Prebuilt) pullFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Match files: name-HASH-HASH.tar.bz2
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

// Push exports installed packages that don't have prebuilts.
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

	// Export packages without prebuilts
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

// fetchURL downloads a file from a URL (not currently used).
func fetchURL(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	return err
}
