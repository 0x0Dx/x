// Package config provides declarative package management configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0x0Dx/x/mochii/internal/hasher"
)

type Package struct {
	Expression interface{} `json:"expression,omitempty"`
	Hash       string      `json:"hash,omitempty"`
}

type Config struct {
	Packages map[string]Package `json:"packages"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Packages == nil {
		cfg.Packages = make(map[string]Package)
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (c *Config) Hash() (hasher.Hash, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	return hasher.FromString(string(data)), nil
}

func (c *Config) PackageHashes() ([]hasher.Hash, error) {
	var hashes []hasher.Hash
	for _, pkg := range c.Packages {
		if pkg.Hash != "" {
			h, err := hasher.Parse(pkg.Hash)
			if err != nil {
				return nil, fmt.Errorf("parse hash %s: %w", pkg.Hash, err)
			}
			hashes = append(hashes, h)
		}
	}
	return hashes, nil
}
