package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultExt = ".png"

func GetImageData(source string) ([]byte, string, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return fetchImageFromURL(source)
	}
	return readLocalImage(source)
}

func fetchImageFromURL(source string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30e9)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	//nolint:gosec // G704: URL is provided by user, but we only do GET requests to fetch images
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	ext := extensionFromContentType(contentType)
	if ext == "" {
		parsedURL, _ := url.Parse(source)
		ext = extensionFromPath(parsedURL.Path)
	}

	return data, ext, nil
}

func readLocalImage(path string) ([]byte, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image: %w", err)
	}

	ext := extensionFromPath(path)
	return data, ext, nil
}

func extensionFromPath(path string) string {
	ext := strings.ToLower(path)
	if idx := strings.LastIndex(ext, "."); idx != -1 {
		return ext[idx:]
	}
	return defaultExt
}

func extensionFromContentType(ct string) string {
	ct = strings.ToLower(ct)
	switch {
	case strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "gif"):
		return ".gif"
	case strings.Contains(ct, "webp"):
		return ".webp"
	}
	return ""
}
