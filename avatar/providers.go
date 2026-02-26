package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

const (
	formatJPEG = ".jpg"
	formatGIF  = ".gif"
	formatWEBP = ".webp"
	formatPNG  = ".png"
)

type Provider interface {
	Name() string
	ChangeAvatar(imageData []byte, format string, token string) error
}

type GitHubProvider struct{}

func (p GitHubProvider) Name() string { return "github" }

func (p GitHubProvider) ChangeAvatar(imageData []byte, _ string, token string) error {
	if token == "" {
		return fmt.Errorf("token is required for GitHub. Set AVATAR_TOKEN or use -t")
	}

	encoded := base64.StdEncoding.EncodeToString(imageData)

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

func (p DiscordProvider) ChangeAvatar(imageData []byte, format string, token string) error {
	if token == "" {
		return fmt.Errorf("token is required for Discord. Set AVATAR_TOKEN or use -t")
	}

	encoded := base64.StdEncoding.EncodeToString(imageData)

	jsonBody := fmt.Sprintf(`{"avatar":"data:image/%s;base64,%s"}`, formatToDiscord(format), encoded)
	ctx, cancel := context.WithTimeout(context.Background(), 30e9)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "PATCH", "https://discord.com/api/v10/users/@me", bytes.NewBufferString(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

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

func (p SteamProvider) ChangeAvatar(imageData []byte, format string, token string) error {
	if token == "" {
		return fmt.Errorf("steam session cookies required (sessionid;steamLoginSecure[;steamCountry])")
	}

	parts := strings.Split(token, ";")
	sessionID := strings.TrimSpace(parts[0])
	steamLoginSecure := ""
	steamCountry := ""
	if len(parts) > 1 {
		steamLoginSecure = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		steamCountry = strings.TrimSpace(parts[2])
	}

	if sessionID == "" || steamLoginSecure == "" {
		return fmt.Errorf("invalid token: need at least sessionid and steamLoginSecure")
	}

	steamID := ""
	if idx := strings.Index(steamLoginSecure, "%7C%7C"); idx != -1 {
		steamID = steamLoginSecure[:idx]
	} else if decoded, err := url.QueryUnescape(steamLoginSecure); err == nil {
		if splitIdx := strings.Index(decoded, "||"); splitIdx != -1 {
			steamID = decoded[:splitIdx]
		}
	}

	if steamID == "" {
		return fmt.Errorf("could not extract SteamID from token")
	}

	img, err := decodeImage(imageData)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	resized := resize.Resize(256, 256, img, resize.Lanczos3)

	var imgData bytes.Buffer
	if format == formatPNG {
		if err := png.Encode(&imgData, resized); err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
	} else {
		if err := jpeg.Encode(&imgData, resized, &jpeg.Options{Quality: 90}); err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
	}

	filename := "avatar.jpg"
	if format == formatPNG {
		filename = "avatar.png"
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("type", "player_avatar_image"); err != nil {
		return fmt.Errorf("failed to write type field: %w", err)
	}
	if err := writer.WriteField("sessionid", sessionID); err != nil {
		return fmt.Errorf("failed to write sessionid field: %w", err)
	}
	if err := writer.WriteField("doSub", "1"); err != nil {
		return fmt.Errorf("failed to write doSub field: %w", err)
	}
	part, err := writer.CreateFormFile("avatar", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(imgData.Bytes()); err != nil {
		return fmt.Errorf("failed to write avatar data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	cookie := fmt.Sprintf("sessionid=%s; steamLoginSecure=%s", sessionID, steamLoginSecure)
	if steamCountry != "" {
		cookie += "; steamCountry=" + steamCountry
	}

	uploadURL := fmt.Sprintf("https://steamcommunity.com/actions/FileUploader?type=player_avatar_image&sId=%s", steamID)
	ctx, cancel := context.WithTimeout(context.Background(), 30e9)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Origin", "https://steamcommunity.com")
	req.Header.Set("Referer", "https://steamcommunity.com/my/edit")

	//nolint:gosec // G704: URL is hardcoded with steamID from user cookie
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("steam API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if len(respBody) > 0 && !bytes.Contains(respBody, []byte("success")) && !bytes.Contains(respBody, []byte("OK")) && !bytes.Contains(respBody, []byte("http")) {
		return fmt.Errorf("steam API returned unexpected response: %s", string(respBody))
	}

	return nil
}

func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}

func formatToDiscord(ext string) string {
	switch ext {
	case formatJPEG, ".jpeg":
		return "jpeg"
	case formatGIF:
		return "gif"
	case formatWEBP:
		return "webp"
	default:
		return "png"
	}
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
