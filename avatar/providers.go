package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Provider interface {
	Name() string
	ChangeAvatar(imagePath string, token string) error
}

type GitHubProvider struct{}

func (p GitHubProvider) Name() string { return "github" }

func (p GitHubProvider) ChangeAvatar(imagePath string, token string) error {
	if token == "" {
		return fmt.Errorf("token is required for GitHub. Set AVATAR_TOKEN or use -t")
	}

	data, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	body := fmt.Sprintf(`{"avatar_base64":"%s"}`, encoded)
	ctx, cancel := context.WithTimeout(context.Background(), 30e9)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "PATCH", "https://api.github.com/user", bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	//nolint:gosec // G704: URL is hardcoded, not user-controlled
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github API error: %s", string(respBody))
	}

	return nil
}

type DiscordProvider struct{}

func (p DiscordProvider) Name() string { return "discord" }

func (p DiscordProvider) ChangeAvatar(imagePath string, token string) error {
	if token == "" {
		return fmt.Errorf("token is required for Discord. Set AVATAR_TOKEN or use -t")
	}

	data, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	ext := filepath.Ext(imagePath)
	format := "png"
	if ext == ".jpg" || ext == ".jpeg" {
		format = "jpeg"
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "avatar."+format)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("failed to write avatar data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30e9)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "PATCH", "https://discord.com/api/v10/users/@me", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	//nolint:gosec // G704: URL is hardcoded, not user-controlled
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error: %s", string(respBody))
	}

	return nil
}

type SteamProvider struct{}

func (p SteamProvider) Name() string { return "steam" }

func (p SteamProvider) ChangeAvatar(_ string, _ string) error {
	return fmt.Errorf("steam does not provide a public API to change avatar programmatically")
}

func GetProvider(name string) (Provider, error) {
	switch name {
	case "github":
		return GitHubProvider{}, nil
	case "discord":
		return DiscordProvider{}, nil
	case "steam":
		return SteamProvider{}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
