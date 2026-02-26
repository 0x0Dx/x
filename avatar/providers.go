package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

func (p SteamProvider) ChangeAvatar(imagePath string, token string) error {
	if token == "" {
		return fmt.Errorf("steam session cookies required (sessionid and steamLoginSecure)")
	}

	data, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	sessionID, steamLoginSecure, ok := strings.Cut(token, ";")
	if !ok {
		return fmt.Errorf("invalid token format: expected 'sessionid;steamLoginSecure'")
	}
	sessionID = strings.TrimSpace(sessionID)
	steamLoginSecure = strings.TrimSpace(steamLoginSecure)

	steamID := ""
	if idx := strings.Index(steamLoginSecure, "%7C%7C"); idx != -1 {
		steamID = steamLoginSecure[:idx]
	} else if decoded, err := url.QueryUnescape(steamLoginSecure); err == nil {
		if parts := strings.Split(decoded, "||"); len(parts) > 0 {
			steamID = parts[0]
		}
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("type", "player_avatar_image"); err != nil {
		return fmt.Errorf("failed to write type field: %w", err)
	}
	if err := writer.WriteField("sId", steamID); err != nil {
		return fmt.Errorf("failed to write sId field: %w", err)
	}
	if err := writer.WriteField("sessionid", sessionID); err != nil {
		return fmt.Errorf("failed to write sessionid field: %w", err)
	}
	if err := writer.WriteField("doSub", "1"); err != nil {
		return fmt.Errorf("failed to write doSub field: %w", err)
	}
	part, err := writer.CreateFormFile("avatar", filepath.Base(imagePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("failed to write avatar data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	cookie := fmt.Sprintf("sessionid=%s; steamLoginSecure=%s;", sessionID, steamLoginSecure)
	ctx, cancel := context.WithTimeout(context.Background(), 30e9)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", "https://steamcommunity.com/actions/FileUploader", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", cookie)
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
		return fmt.Errorf("steam API error: %s", string(respBody))
	}

	return nil
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
