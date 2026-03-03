// Package github provides GitHub API helpers.
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CreateRelease creates a release via GitHub API.
func CreateRelease(token, owner, repo, tag string, draft bool, assets []string) error {
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
		if err := UploadAsset(token, owner, repo, result.ID, path); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
	}

	return nil
}

// UploadAsset uploads a file to a GitHub release.
func UploadAsset(token, owner, repo string, releaseID int, path string) error {
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
