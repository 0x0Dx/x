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

type Fetcher struct {
	SourcesDir string
}

func New(sourcesDir string) *Fetcher {
	return &Fetcher{SourcesDir: sourcesDir}
}

func (f *Fetcher) FetchURL(url string) (string, error) {
	filename := util.BaseNameOf(url)
	fullname := f.SourcesDir + "/" + filename

	if util.FileExists(fullname) {
		return fullname, nil
	}

	fmt.Printf("fetching %s\n", url)

	if err := util.EnsureDir(f.SourcesDir); err != nil {
		return "", err
	}

	tmpFile := fullname + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}

	out.Close()

	if err := os.Rename(tmpFile, fullname); err != nil {
		return "", err
	}

	return fullname, nil
}

func (f *Fetcher) FetchHash(hash string) (string, error) {
	return "", fmt.Errorf("FetchHash not implemented")
}

func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
