package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0x0Dx/x/releasex/internal/config"
	"github.com/spf13/cobra"
)

var (
	githubToken string
	draft       bool
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Build and create GitHub release",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := buildCmd.RunE(buildCmd, nil); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.GitHub) == 0 {
			fmt.Println("No GitHub config found, skipping release")
			return nil
		}

		version := cfg.Version
		if cfg.GitHub[0].Version == "tag" {
			version = strings.TrimPrefix(version, "v")
		}

		tag := cfg.GitHub[0].Version
		if tag == "tag" {
			tag = version
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
			if err := createReleaseAPI(token, cfg.GitHub[0].Owner, cfg.GitHub[0].Repo, tag, draft, assets); err != nil {
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

func createReleaseAPI(token, owner, repo, tag string, draft bool, assets []string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

	body := fmt.Sprintf(`{"tag_name":"%s","draft":%t,"generate_release_notes":true}`, tag, draft)
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	//nolint:gosec // G704: false positive - hardcoded GitHub API URL
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %s %s", resp.Status, string(b))
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, path := range assets {
		if err := uploadAssetAPI(token, owner, repo, result.ID, path); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
	}

	return nil
}

func uploadAssetAPI(token, owner, repo string, releaseID int, path string) error {
	name := filepath.Base(path)
	url := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets?name=%s", owner, repo, releaseID, name)

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	//nolint:gosec // G704: false positive - hardcoded GitHub API URL
	req, err := http.NewRequest("POST", url, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")

	//nolint:gosec // G704: false positive - hardcoded GitHub API URL
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s %s", resp.Status, string(b))
	}
	return nil
}
