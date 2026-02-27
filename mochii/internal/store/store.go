// Package store provides store query operations like nix-store.
package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0x0Dx/x/mochii/internal/db"
)

// Store provides access to the mochii store.
type Store struct {
	DB         *db.DB
	SourcesDir string
	InstallDir string
}

// New creates a new Store.
func New(db *db.DB, sourcesDir, installDir string) *Store {
	return &Store{
		DB:         db,
		SourcesDir: sourcesDir,
		InstallDir: installDir,
	}
}

// QueryResult represents a query result.
type QueryResult struct {
	Path string
	Hash string
}

// Query performs store queries.
func (s *Store) Query(args []string) ([]QueryResult, error) {
	var results []QueryResult

	for _, arg := range args {
		// Check if it's a hash
		if isHex(arg) && len(arg) == 64 {
			path, found, err := s.DB.Get("pkginst", arg)
			if err != nil {
				return nil, err
			}
			if found {
				results = append(results, QueryResult{
					Path: path,
					Hash: arg,
				})
				continue
			}
		}

		// Check if it's a path
		info, err := os.Stat(arg)
		if err == nil && info.IsDir() {
			pkgs, err := s.DB.List("pkginst")
			if err != nil {
				return nil, err
			}
			for h, p := range pkgs {
				if strings.HasPrefix(p, arg) {
					results = append(results, QueryResult{
						Path: p,
						Hash: h,
					})
				}
			}
		}
	}

	return results, nil
}

// QueryRequisites returns all requisites (dependencies) of a path.
func (s *Store) QueryRequisites(path string) ([]string, error) {
	var requisites []string
	visited := make(map[string]bool)

	var walk func(p string) error
	walk = func(p string) error {
		if visited[p] {
			return nil
		}
		visited[p] = true
		requisites = append(requisites, p)

		refsFile := p + "/.mochii-refs"
		if data, err := os.ReadFile(refsFile); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if isHex(line) && len(line) == 64 {
					walk(s.InstallDir + "/" + line)
				}
			}
		}
		return nil
	}

	if err := walk(path); err != nil {
		return nil, err
	}

	return requisites, nil
}

// QueryGraveyard returns unreachable store paths.
func (s *Store) QueryGraveyard(profilePath string) ([]string, error) {
	alive := make(map[string]bool)

	if entries, err := os.ReadDir(profilePath); err == nil {
		for _, e := range entries {
			name := e.Name()
			if strings.HasSuffix(name, ".hash") {
				data, _ := os.ReadFile(filepath.Join(profilePath, name))
				hashStr := strings.TrimSpace(string(data))
				alive[hashStr] = true
			}
		}
	}

	installed, err := s.DB.List("pkginst")
	if err != nil {
		return nil, err
	}

	var garbage []string
	for h, path := range installed {
		if !alive[h] {
			garbage = append(garbage, path)
		}
	}

	return garbage, nil
}

// QueryReferences returns references (dependencies) of a path.
func (s *Store) QueryReferences(path string) ([]string, error) {
	var refs []string
	refsFile := path + "/.mochii-refs"

	if data, err := os.ReadFile(refsFile); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				refs = append(refs, line)
			}
		}
	}

	return refs, nil
}

// Verify checks store consistency.
func (s *Store) Verify() ([]string, error) {
	var errors []string

	installed, err := s.DB.List("pkginst")
	if err != nil {
		return nil, err
	}

	for h, path := range installed {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("path %s (hash: %s) does not exist", path, h))
		}
	}

	return errors, nil
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
