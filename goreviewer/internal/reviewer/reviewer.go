// Package reviewer provides AI code review functionality.
package reviewer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/0x0Dx/x/goreviewer/internal/github"
)

const (
	reviewHeader = "## AI Code Review"
	reviewFooter = "*Review by [GoReviewer](https://github.com/0x0Dx/x/goreviewer)*"
	maxDiffSize  = 5000000
)

// Config holds reviewer configuration.
type Config struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Verbose     bool
}

// ReviewResponse represents the AI review response.
type ReviewResponse struct {
	Review           string   `json:"review"`
	FailPassWorkflow string   `json:"fail_pass_workflow"`
	LabelsAdded      []string `json:"labels_added"`
}

// Reviewer handles AI code reviews.
type Reviewer struct {
	cfg        Config
	httpClient *http.Client
	ghClient   *github.Client
}

// New creates a new Reviewer.
func New(cfg Config, ghClient *github.Client) *Reviewer {
	return &Reviewer{
		cfg:        cfg,
		httpClient: &http.Client{},
		ghClient:   ghClient,
	}
}

// Review performs an AI code review on the given diff.
func (r *Reviewer) Review(ctx context.Context, diffContent string) (ReviewResponse, error) {
	if diffContent == "" {
		return errorResponse("No diff content to analyze"), errors.New("empty diff content")
	}

	if len(diffContent) > maxDiffSize {
		return errorResponse(fmt.Sprintf("Diff is too large (%d bytes, max: %d bytes)", len(diffContent), maxDiffSize)), errors.New("diff too large")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return errorResponse("Missing OPENROUTER_API_KEY environment variable"), errors.New("missing API key")
	}

	prompt := r.buildPrompt(diffContent)

	repoFull := os.Getenv("REPO_FULL_NAME")
	referer := fmt.Sprintf("https://github.com/%s", repoFull)

	body, err := r.callAPI(ctx, apiKey, referer, prompt)
	if err != nil {
		return errorResponse(fmt.Sprintf("API call failed: %v", err)), err
	}

	return r.parseResponse(body)
}

func (r *Reviewer) buildPrompt(diffContent string) string {
	var b strings.Builder

	b.WriteString("Please analyze this code diff and provide a comprehensive review in markdown format.\n\n")
	b.WriteString("Focus on security, performance, code quality, and best practices.\n\n")
	b.WriteString("Keep the review scannable and grouped by importance. Lead with critical issues if any exist.\n\n")
	b.WriteString("Your response MUST be a valid JSON object with exactly these three fields:\n")
	b.WriteString("- `review`: The complete markdown review content\n")
	b.WriteString("- `fail_pass_workflow`: Either \"pass\", \"fail\", or \"uncertain\" - whether the code changes should block merging\n")
	b.WriteString("- `labels_added`: An array of label strings that best describe this PR (e.g., [\"bug\", \"security\"])\n\n")
	b.WriteString("Respond ONLY with the JSON object, no other text.\n")

	if r.ghClient != nil && r.ghClient.Token != "" {
		ghCtx, err := r.ghClient.FetchContext(context.Background())
		if err == nil {
			if ghCtx.CheckRuns != "" {
				b.WriteString("\nGitHub Actions Check Status:\n")
				b.WriteString(ghCtx.CheckRuns)
				b.WriteString("\n\nPlease consider any failed or pending checks in your review. If tests are failing, investigate whether the code changes might be the cause.\n")
			}

			if ghCtx.Labels != "" {
				b.WriteString("\nAvailable Repository Labels:\nPlease prefer using existing labels from this list over creating new ones:\n")
				b.WriteString(ghCtx.Labels)
				b.WriteString("\n\nIf none of these labels are appropriate for the changes, you may suggest new ones.\n")
			}

			if ghCtx.PRDescription != "" {
				b.WriteString("\nPull Request Context:\n")
				b.WriteString(ghCtx.PRDescription)
				b.WriteString("\n")
			}

			if ghCtx.Commits != "" {
				b.WriteString("\nCommit History (showing development journey):\n")
				b.WriteString(ghCtx.Commits)
				b.WriteString("\n\nPlease consider the commit history to understand what was tried, what issues were discovered, and how the solution evolved.\n")
			}

			if ghCtx.HumanComments != "" {
				b.WriteString("\nHuman Comments on this PR:\n")
				b.WriteString(ghCtx.HumanComments)
				b.WriteString("\n\nPlease consider these human comments when reviewing the code.\n")
			}

			if ghCtx.PreviousReview != "" {
				b.WriteString("\nPrevious AI Review (for context on what was already reviewed):\n")
				b.WriteString(ghCtx.PreviousReview)
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\nCode diff to analyze:\n\n")
	b.WriteString(diffContent)

	return b.String()
}

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

func (r *Reviewer) callAPI(ctx context.Context, apiKey, referer, prompt string) ([]byte, error) {
	reqBody := apiRequest{
		Model: r.cfg.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		Temperature: r.cfg.Temperature,
		MaxTokens:   r.cfg.MaxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", referer)

	resp, err := r.httpClient.Do(req)
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
	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errorResponse("Invalid JSON response from API"), fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error.Message != "" {
		return errorResponse(resp.Error.Message), errors.New(resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return errorResponse("No choices in API response"), errors.New("empty response")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return errorResponse("Empty response content"), errors.New("empty content")
	}

	content = removeThinking(content)
	content = strings.TrimSpace(content)

	jsonMatch := extractJSON(content)
	if jsonMatch == "" {
		return errorResponse("No valid JSON found in response"), errors.New("no JSON in response")
	}

	var result ReviewResponse
	if err := json.Unmarshal([]byte(jsonMatch), &result); err != nil {
		return errorResponse(fmt.Sprintf("Invalid JSON in response: %v", err)), err
	}

	if result.Review == "" {
		return errorResponse("Missing review field"), errors.New("missing review")
	}

	return result, nil
}

func removeThinking(content string) string {
	re := regexp.MustCompile(`(?s)<thinking>.*?</thinking>\s*`)
	return re.ReplaceAllString(content, "")
}

func extractJSON(content string) string {
	re := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
	matches := re.FindAllString(content, -1)

	var jsonStr string
	for _, m := range matches {
		if isValidJSON(m) {
			jsonStr = m
			break
		}
	}

	if jsonStr == "" && isValidJSON(content) {
		jsonStr = content
	}

	return jsonStr
}

func isValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func errorResponse(msg string) ReviewResponse {
	footer := fmt.Sprintf("\n\n---\n%s\n", reviewFooter)
	return ReviewResponse{
		Review:           reviewHeader + "\n\n❌ **Error**: " + msg + footer,
		FailPassWorkflow: "uncertain",
		LabelsAdded:      []string{},
	}
}

// ToJSON converts the ReviewResponse to JSON string.
func (r *ReviewResponse) ToJSON() (string, error) {
	b, err := json.Marshal(r)
	return string(b), err
}
