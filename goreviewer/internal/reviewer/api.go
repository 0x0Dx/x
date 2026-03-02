package reviewer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type apiRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiResponse struct {
	Choices []choice `json:"choices"`
	Error   apiError `json:"error"`
}

type choice struct {
	Message message `json:"message"`
}

type apiError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (r *Reviewer) callAPI(ctx context.Context, apiKey, baseURL, referer, model, prompt string) ([]byte, error) {
	temp := r.cfg.Temperature
	if temp == 0 {
		temp = 0.05
	}

	maxTokens := r.cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 64000
	}

	retries := r.cfg.Retries
	if retries == 0 {
		retries = 5
	}

	var lastErr error
	for i := 0; i < retries; i++ {
		body, err := r.doRequest(ctx, apiKey, baseURL, referer, model, prompt, temp, maxTokens)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if i < retries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	return nil, lastErr
}

func (r *Reviewer) doRequest(ctx context.Context, apiKey, baseURL, referer, model, prompt string, temp float64, maxTokens int) ([]byte, error) {
	reqBody := apiRequest{
		Model: model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		Temperature: temp,
		MaxTokens:   maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", referer)

	resp, err := r.httpClient.Do(req) //nolint:gosec // G704: intentionally calling known API endpoint
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (r *Reviewer) parseResponse(body []byte) (ReviewResponse, error) {
	if r.cfg.Debug {
		fmt.Printf("DEBUG: API response body: %s\n", string(body))
	}

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errorResponse(fmt.Sprintf("Failed to parse API response - the response wasn't valid JSON. This usually means the API is down, rate limited, or returning an error page. Check your API key and try again later. Details: %v", err)), fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error.Message != "" {
		return errorResponse(fmt.Sprintf("The API returned an error: '%s'. This usually means your API key is invalid, expired, or you've hit a rate limit. Check your API key and try again.", resp.Error.Message)), errors.New(resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return errorResponse("The API returned no response choices. This usually means the model is rate limited or unavailable. Try again later."), errors.New("empty response")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return errorResponse("The API returned an empty response. This usually means the model is rate limited or had an error. Try again later."), errors.New("empty content")
	}

	content = removeThinking(content)
	content = strings.TrimSpace(content)

	jsonMatch := extractJSON(content)
	if jsonMatch == "" {
		truncated := content
		if len(truncated) > 500 {
			truncated = truncated[:500] + "..."
		}
		return errorResponse(fmt.Sprintf("The API returned text but it wasn't valid JSON. The model might be confused or rate limited. Here's what we got: %s", truncated)), errors.New("no JSON in response")
	}

	var result ReviewResponse
	if err := json.Unmarshal([]byte(jsonMatch), &result); err != nil {
		return errorResponse(fmt.Sprintf("Invalid JSON in response: %v", err)), err
	}

	if result.Review == "" {
		return errorResponse("Missing review field"), errors.New("missing review")
	}

	result.Review = stripExistingFooter(result.Review)

	result.Review += buildFooter(r.cfg.BotIcon)
	result.Review = addReviewHash(result.Review)

	return result, nil
}

func (r *Reviewer) parseSummaryResponse(body []byte) (SummaryResponse, error) {
	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return SummaryResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error.Message != "" {
		return SummaryResponse{}, errors.New(resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return SummaryResponse{}, errors.New("empty response")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return SummaryResponse{}, errors.New("empty content")
	}

	content = removeThinking(content)
	content = strings.TrimSpace(content)

	var result SummaryResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		result.Walkthrough = content
	}

	return result, nil
}

func (r *Reviewer) parseSimpleResponse(body []byte) (string, error) {
	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error.Message != "" {
		return "", errors.New(resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("empty response")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return "", errors.New("empty content")
	}

	content = removeThinking(content)
	return strings.TrimSpace(content), nil
}

func jsonMarshal(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return b, nil
}
