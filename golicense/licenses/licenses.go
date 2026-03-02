// Package licenses provides license file templates and generation.
package licenses

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"strings"
	"text/template"
	"time"
)

//go:embed data/*.txt
var licenses embed.FS

// List returns all available license names.
func List() ([]string, error) {
	var result []string

	if err := fs.WalkDir(licenses, "data", func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if path == "data" {
			return nil
		}

		fname := strings.TrimSuffix(path, ".txt")
		fname = strings.TrimPrefix(fname, "data/")

		result = append(result, fname)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk licenses directory: %w", err)
	}

	return result, nil
}

// Has returns true if the given license exists.
func Has(license string) bool {
	fin, err := licenses.Open("data/" + license + ".txt")
	if err != nil {
		return false
	}
	if err := fin.Close(); err != nil {
		log.Printf("failed to close license file: %v", err)
	}

	return true
}

// Hydrate fills in the template variables for the given license.
func Hydrate(license, name, email string, sink io.Writer) error {
	tmpl, err := template.ParseFS(licenses, "data/"+license+".txt")
	if err != nil {
		return fmt.Errorf("failed to parse license template: %w", err)
	}

	if err := tmpl.Execute(sink, struct {
		Name  string
		Email string
		Year  int
	}{
		Name:  name,
		Email: email,
		Year:  time.Now().Year(),
	}); err != nil {
		return fmt.Errorf("failed to execute license template: %w", err)
	}

	return nil
}
