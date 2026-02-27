package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/0x0Dx/x/mochii/internal/util"
)

// Fetcher handles downloading files from URLs.
type Fetcher struct {
	SourcesDir string
}

// New creates a new Fetcher with the given sources directory.
func New(sourcesDir string) *Fetcher {
	return &Fetcher{SourcesDir: sourcesDir}
}

// FetchURL downloads a file from a URL to the sources directory.
// Returns the local path if already cached.
func (f *Fetcher) FetchURL(url string) (string, error) {
	filename := util.BaseNameOf(url)
	fullname := f.SourcesDir + "/" + filename

	// Return cached file if it exists
	if util.FileExists(fullname) {
		return fullname, nil
	}

	fmt.Printf("fetching %s\n", url)

	if err := util.EnsureDir(f.SourcesDir); err != nil {
		return "", err
	}

	// Download to temp file first
	tmpFile := fullname + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	if err := resp.Body.Close(); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}

	if err := out.Close(); err != nil {
		return "", err
	}

	// Atomically rename to final location
	if err := os.Rename(tmpFile, fullname); err != nil {
		return "", err
	}

	return fullname, nil
}

// FetchHash fetches a file by hash (not yet implemented).
func (f *Fetcher) FetchHash(hash string) (string, error) {
	return "", fmt.Errorf("FetchHash not implemented")
}

// IsURL checks if a string is an HTTP/HTTPS URL.
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
