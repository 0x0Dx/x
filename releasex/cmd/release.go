// Package cmd provides releasex CLI commands.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0x0Dx/x/releasex/internal/config"
	"github.com/0x0Dx/x/releasex/internal/github"
	"github.com/spf13/cobra"
)

var (
	githubToken string
	draft       bool
)

// releaseCmd builds and creates a GitHub release.
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Build and create GitHub release",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := buildCmd.RunE(buildCmd, nil); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		cfg, err := config.Load(GetConfigPath())
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.GitHub) == 0 {
			fmt.Println("No GitHub config found, skipping release")
			return nil
		}

		tag := cfg.Version
		if cfg.GitHub[0].Version == "tag" {
			tag = strings.TrimPrefix(tag, "v")
			if !strings.HasPrefix(tag, "v") {
				tag = "v" + tag
			}
		}

		files, _ := os.ReadDir(buildDir)
		var assets []string
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			assets = append(assets, filepath.Join(buildDir, f.Name()))
		}

		if findGH() {
			fmt.Println("Using gh CLI for release...")
			if err := createReleaseGH(cfg.GitHub[0].Owner, cfg.GitHub[0].Repo, tag, draft, assets); err != nil {
				return fmt.Errorf("gh release failed: %w", err)
			}
		} else {
			fmt.Println("Using API for release...")
			token := githubToken
			if token == "" {
				token = os.Getenv("GITHUB_TOKEN")
			}
			if token == "" {
				return fmt.Errorf("GITHUB_TOKEN not set and gh not found")
			}
			if err := github.CreateRelease(token, cfg.GitHub[0].Owner, cfg.GitHub[0].Repo, tag, draft, assets); err != nil {
				return fmt.Errorf("API release failed: %w", err)
			}
		}

		fmt.Println("Release complete!")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(releaseCmd)
	releaseCmd.Flags().StringVar(&githubToken, "token", "", "GitHub token")
	releaseCmd.Flags().BoolVar(&draft, "draft", false, "Create draft release")
}

func findGH() bool {
	cmd := exec.Command("gh", "--version")
	return cmd.Run() == nil
}

func createReleaseGH(owner, repo, tag string, draft bool, assets []string) error {
	args := []string{"release", "create", tag, "-R", owner + "/" + repo}
	if draft {
		args = append(args, "--draft")
	}
	args = append(args, "--generate-notes")
	args = append(args, assets...)

	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh command failed: %w", err)
	}
	return nil
}
