// Package licenses provides license file templates and generation.
package licenses

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path"
	"strings"
	"text/template"
	"time"
)

//go:embed data/*.txt
var licenses embed.FS

const dataDir = "data"

// licenseData holds the template variables for license generation.
type licenseData struct {
	Name  string
	Email string
	Year  int
}

// List returns all available license names, sorted alphabetically.
func List() ([]string, error) {
	var result []string

	err := fs.WalkDir(licenses, dataDir, func(p string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == dataDir {
			return nil
		}
		name := strings.TrimSuffix(path.Base(p), ".txt")
		result = append(result, name)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk licenses directory: %w", err)
	}

	return result, nil
}

// Has returns true if the given license exists.
func Has(license string) bool {
	fin, err := licenses.Open(path.Join(dataDir, license+".txt"))
	if err != nil {
		return false
	}
	if err := fin.Close(); err != nil {
		log.Printf("close license file: %v", err)
	}
	return true
}

// Hydrate fills in the template variables for the given license
// and writes the result to sink.
func Hydrate(license, name, email string, sink io.Writer) error {
	tmpl, err := template.ParseFS(licenses, path.Join(dataDir, license+".txt"))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	data := licenseData{
		Name:  name,
		Email: email,
		Year:  time.Now().Year(),
	}

	if err := tmpl.Execute(sink, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
