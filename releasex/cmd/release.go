package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
			return err
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.GitHub) == 0 {
			fmt.Println("No GitHub config found, skipping release")
			return nil
		}

		token := githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}
		if token == "" {
			return fmt.Errorf("GITHUB_TOKEN not set")
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

		releaseID, err := createRelease(token, cfg.GitHub[0].Owner, cfg.GitHub[0].Repo, tag, draft)
		if err != nil {
			return fmt.Errorf("failed to create release: %w", err)
		}

		// Upload assets
		files, _ := os.ReadDir(buildDir)
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			path := filepath.Join(buildDir, f.Name())
			fmt.Printf("Uploading %s...\n", f.Name())
			if err := uploadAsset(token, cfg.GitHub[0].Owner, cfg.GitHub[0].Repo, releaseID, path); err != nil {
				return fmt.Errorf("failed to upload %s: %w", f.Name(), err)
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

func createRelease(token, owner, repo, tag string, draft bool) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

	body := fmt.Sprintf(`{"tag_name":"%s","draft":%t,"generate_release_notes":true}`, tag, draft)
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status: %s %s", resp.Status, string(b))
	}

	var result struct {
		ID int `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return fmt.Sprintf("%d", result.ID), nil
}

func uploadAsset(token, owner, repo, releaseID, path string) error {
	name := filepath.Base(path)
	url := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%s/assets?name=%s", owner, repo, releaseID, name)

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	req, _ := http.NewRequest("POST", url, file)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s %s", resp.Status, string(b))
	}
	return nil
}
