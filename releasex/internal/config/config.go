package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project   string            `yaml:"project"`
	Version   string            `yaml:"version"`
	Builds    []Build           `yaml:"builds"`
	Archives  []Archive         `yaml:"archives"`
	Checksums []ChecksumsConfig `yaml:"checksums"`
	GitHub    []GitHubConfig    `yaml:"github"`
}

type Build struct {
	ID      string   `yaml:"id"`
	GoOS    []string `yaml:"goos"`
	GoArch  []string `yaml:"goarch"`
	Main    string   `yaml:"main"`
	Binary  string   `yaml:"binary"`
	Env     []string `yaml:"env"`
	LdFlags string   `yaml:"ldflags"`
}

type Archive struct {
	ID     string   `yaml:"id"`
	Builds []string `yaml:"builds"`
	Format string   `yaml:"format"`
	Files  []string `yaml:"files"`
}

type ChecksumsConfig struct {
	IDs     []string `yaml:"ids"`
	Outputs []string `yaml:"outputs"`
}

type GitHubConfig struct {
	Owner   string `yaml:"owner"`
	Repo    string `yaml:"repo"`
	Version string `yaml:"version"`
	Draft   bool   `yaml:"draft"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
