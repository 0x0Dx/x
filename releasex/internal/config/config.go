// Package config handles releasex configuration.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the releasex configuration.
type Config struct {
	Project   string            `yaml:"project"`
	Version   string            `yaml:"version"`
	Builds    []Build           `yaml:"builds"`
	Archives  []Archive         `yaml:"archives"`
	Checksums []ChecksumsConfig `yaml:"checksums"`
	GitHub    []GitHubConfig    `yaml:"github"`
}

// Build defines a binary to build.
type Build struct {
	ID      string   `yaml:"id"`
	GoOS    []string `yaml:"goos"`
	GoArch  []string `yaml:"goarch"`
	Main    string   `yaml:"main"`
	Binary  string   `yaml:"binary"`
	Env     []string `yaml:"env"`
	LdFlags string   `yaml:"ldflags"`
}

// Archive defines an archive to create.
type Archive struct {
	ID     string   `yaml:"id"`
	Builds []string `yaml:"builds"`
	Format string   `yaml:"format"`
	Files  []string `yaml:"files"`
}

// ChecksumsConfig defines checksums to generate.
type ChecksumsConfig struct {
	IDs     []string `yaml:"ids"`
	Outputs []string `yaml:"outputs"`
}

// GitHubConfig defines GitHub release settings.
type GitHubConfig struct {
	Owner   string `yaml:"owner"`
	Repo    string `yaml:"repo"`
	Version string `yaml:"version"`
	Draft   bool   `yaml:"draft"`
}

// Load reads and parses the configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return &cfg, nil
}
