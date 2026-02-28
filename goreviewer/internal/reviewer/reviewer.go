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
	"time"

	"github.com/0x0Dx/x/goreviewer/internal/github"
)

var safeIconRegex = regexp.MustCompile(`^[\p{L}\p{N}\p{Po}\p{S}\s]+$`)
var avatarRegex = regexp.MustCompile(`^<img src="[^"]+" alt="[^"]+" width="\d+" height="\d+" ?/?>$`)

func isValidBotIcon(icon string) bool {
	if icon == "" || len(icon) > 100 {
		return false
	}
	if avatarRegex.MatchString(icon) {
		return true
	}
	return safeIconRegex.MatchString(icon)
}

const (
	reviewHeader    = "## AI Code Review"
	reviewFooter    = "*Review by [GoReviewer](https://github.com/0x0Dx/x/tree/main/goreviewer)*"
	maxDiffSize     = 5000000
	defaultModel    = "minimax/minimax-m2.5"
	defaultBaseURL  = "https://openrouter.ai/api/v1"
	defaultLanguage = "en-US"
)

// Config holds reviewer configuration.
type Config struct {
	Debug               bool
	MaxFiles            int
	ReviewSimpleChanges bool
	ReviewCommentLGTM   bool
	PathFilters         string
	DisableReview       bool
	DisableReleaseNotes bool
	OpenAIBaseURL       string
	LightModel          string
	HeavyModel          string
	Temperature         float64
	Retries             int
	TimeoutMS           int
	ConcurrencyLimit    int
	SystemMessage       string
	SummarizePrompt     string
	ReleaseNotesPrompt  string
	Language            string
	BotIcon             string

	// Legacy options (for backward compatibility)
	Model     string
	MaxTokens int
}

// ReviewResponse represents the AI review response.
type ReviewResponse struct {
	Review           string   `json:"review"`
	FailPassWorkflow string   `json:"fail_pass_workflow"`
	LabelsAdded      []string `json:"labels_added"`
	ReleaseNotes     string   `json:"release_notes,omitempty"`
}

type fileChange struct {
	Files   string `json:"files"`
	Summary string `json:"summary"`
}

// SummaryResponse represents the summary response.
type SummaryResponse struct {
	Walkthrough string       `json:"walkthrough"`
	Changes     []fileChange `json:"changes"`
	Poem        string       `json:"poem"`
}

// ReviewCommentRequest represents a request to respond to a review comment.
type ReviewCommentRequest struct {
	Comment   string
	DiffHunk  string
	Path      string
	Line      int
	PRNumber  int
	RepoOwner string
	RepoName  string
}

// Reviewer handles AI code reviews.
type Reviewer struct {
	cfg        Config
	httpClient *http.Client
	ghClient   *github.Client
}

// New creates a new Reviewer.
func New(cfg Config, ghClient *github.Client) *Reviewer {
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.TimeoutMS) * time.Millisecond,
	}
	return &Reviewer{
		cfg:        cfg,
		httpClient: httpClient,
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
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return errorResponse("Missing OPENROUTER_API_KEY or OPENAI_API_KEY environment variable"), errors.New("missing API key")
	}

	// Determine which model to use
	model := r.cfg.HeavyModel
	if model == "" {
		model = r.cfg.Model
	}
	if model == "" {
		model = defaultModel
	}

	// Build the prompt
	prompt := r.buildReviewPrompt(diffContent)

	repoFull := os.Getenv("REPO_FULL_NAME")
	referer := fmt.Sprintf("https://github.com/%s", repoFull)

	// Call the API
	baseURL := r.cfg.OpenAIBaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	body, err := r.callAPI(ctx, apiKey, baseURL, referer, model, prompt)
	if err != nil {
		return errorResponse(fmt.Sprintf("API call failed: %v", err)), err
	}

	return r.parseResponse(body)
}

// Summarize generates a summary of the diff.
func (r *Reviewer) Summarize(ctx context.Context, diffContent string) (SummaryResponse, error) {
	if diffContent == "" {
		return SummaryResponse{}, errors.New("empty diff content")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return SummaryResponse{}, errors.New("missing API key")
	}

	model := r.cfg.LightModel
	if model == "" {
		model = defaultModel
	}

	prompt := r.buildSummarizePrompt(diffContent)

	repoFull := os.Getenv("REPO_FULL_NAME")
	referer := fmt.Sprintf("https://github.com/%s", repoFull)

	baseURL := r.cfg.OpenAIBaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	body, err := r.callAPI(ctx, apiKey, baseURL, referer, model, prompt)
	if err != nil {
		return SummaryResponse{}, err
	}

	return r.parseSummaryResponse(body)
}

// RespondToReviewComment responds to a review comment.
func (r *Reviewer) RespondToReviewComment(ctx context.Context, req ReviewCommentRequest) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return "", errors.New("missing API key")
	}

	model := r.cfg.HeavyModel
	if model == "" {
		model = r.cfg.Model
	}
	if model == "" {
		model = defaultModel
	}

	prompt := r.buildReviewCommentPrompt(req)

	repoFull := fmt.Sprintf("%s/%s", req.RepoOwner, req.RepoName)
	referer := fmt.Sprintf("https://github.com/%s", repoFull)

	baseURL := r.cfg.OpenAIBaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	body, err := r.callAPI(ctx, apiKey, baseURL, referer, model, prompt)
	if err != nil {
		return "", err
	}

	return r.parseSimpleResponse(body)
}

func (r *Reviewer) selectModel(isHeavy bool) string {
	if isHeavy {
		model := r.cfg.HeavyModel
		if model == "" {
			model = r.cfg.Model
		}
		if model == "" {
			model = defaultModel
		}
		return model
	}
	model := r.cfg.LightModel
	if model == "" {
		model = defaultModel
	}
	return model
}

// ChatRequest represents a chat request.
type ChatRequest struct {
	CodeContext string
	Message     string
	FilePath    string
	Line        int
}

// Chat responds to a chat message about code context.
func (r *Reviewer) Chat(ctx context.Context, req ChatRequest) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return "", errors.New("missing API key")
	}

	model := r.selectModel(true)

	prompt := r.buildChatPrompt(req)

	repoFull := os.Getenv("REPO_FULL_NAME")
	referer := fmt.Sprintf("https://github.com/%s", repoFull)

	baseURL := r.cfg.OpenAIBaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	body, err := r.callAPI(ctx, apiKey, baseURL, referer, model, prompt)
	if err != nil {
		return "", err
	}

	return r.parseSimpleResponse(body)
}

func (r *Reviewer) buildChatPrompt(req ChatRequest) string {
	var b strings.Builder

	systemMsg := r.cfg.SystemMessage
	if systemMsg == "" {
		systemMsg = `You are @goreviewer, a helpful AI assistant for code reviews.`
	}

	lang := r.cfg.Language
	if lang == "" {
		lang = defaultLanguage
	}
	systemMsg += fmt.Sprintf("\n\nYour entire response must be in the language with ISO code: %s", lang)

	b.WriteString(systemMsg)
	b.WriteString("\n\n")

	if req.FilePath != "" {
		b.WriteString(fmt.Sprintf("The user is asking about file: **%s**\n", req.FilePath))
		if req.Line > 0 {
			b.WriteString(fmt.Sprintf("Specifically about line: **%d**\n", req.Line))
		}
		b.WriteString("\n")
	}

	b.WriteString("Code context:\n\n")
	b.WriteString("```\n")
	b.WriteString(req.CodeContext)
	b.WriteString("\n```\n\n")

	b.WriteString(fmt.Sprintf("User's question: %s\n\n", req.Message))
	b.WriteString("Please answer the user's question based on the provided code context. Be helpful and specific.")

	return b.String()
}

func (r *Reviewer) buildReviewPrompt(diffContent string) string {
	var b strings.Builder

	// Use custom system message or default
	systemMsg := r.cfg.SystemMessage
	if systemMsg == "" {
		systemMsg = `You are @goreviewer, a language model trained by OpenAI. 
Your purpose is to act as a highly experienced software engineer and provide a thorough review of the code hunks and suggest code snippets to improve key areas such:
- Logic
- Security
- Performance
- Data races
- Consistency
- Error handling
- Maintainability
- Modularity
- Complexity
- Optimization
- Best practices: DRY, SOLID, KISS

Do not comment on minor code style issues, missing comments/documentation. 
Identify and resolve significant concerns to improve overall code quality while deliberately disregarding minor issues.`
	}

	// Add language requirement
	lang := r.cfg.Language
	if lang == "" {
		lang = defaultLanguage
	}
	systemMsg += fmt.Sprintf("\n\nYour entire response must be in the language with ISO code: %s", lang)

	b.WriteString(systemMsg)
	b.WriteString("\n\n")

	b.WriteString("Please analyze this code diff and provide a comprehensive review in markdown format.\n\n")
	b.WriteString("Focus on security, performance, code quality, and best practices.\n\n")
	b.WriteString("Keep the review scannable and grouped by importance. Lead with critical issues if any exist.\n\n")

	// Check if we should include release notes
	if !r.cfg.DisableReleaseNotes {
		b.WriteString("Your response MUST be a valid JSON object with these fields:\n")
		b.WriteString("- `review`: The complete markdown review content\n")
		b.WriteString("- `fail_pass_workflow`: Either \"pass\", \"fail\", or \"uncertain\"\n")
		b.WriteString("- `labels_added`: Array of label strings (e.g., [\"bug\", \"security\"])\n")
		b.WriteString("- `release_notes`: Brief release notes for this PR (50-100 words)\n\n")
	} else {
		b.WriteString("Your response MUST be a valid JSON object with these fields:\n")
		b.WriteString("- `review`: The complete markdown review content\n")
		b.WriteString("- `fail_pass_workflow`: Either \"pass\", \"fail\", or \"uncertain\"\n")
		b.WriteString("- `labels_added`: Array of label strings (e.g., [\"bug\", \"security\"])\n\n")
	}

	b.WriteString("Respond ONLY with the JSON object, no other text.\n")

	if r.ghClient != nil && r.ghClient.Token != "" {
		r.addGitHubContext(&b)
	}

	b.WriteString("\nCode diff to analyze:\n\n")
	b.WriteString(diffContent)

	return b.String()
}

func (r *Reviewer) buildSummarizePrompt(diffContent string) string {
	var b strings.Builder

	// Use custom prompt or default
	prompt := r.cfg.SummarizePrompt
	if prompt == "" {
		prompt = `Provide your final response in markdown with the following content:
- **Walkthrough**: A high-level summary of the overall change within 80 words.
- **Changes**: A markdown table of files and their summaries. Group files with similar changes together.
- **Poem**: Below the changes, include a whimsical short poem written by a rabbit to celebrate the changes. Format as a quote using ">" and use emojis.

Avoid additional commentary as this summary will be added as a comment on the GitHub pull request. Use titles "Walkthrough" and "Changes" as H2.`
	}

	lang := r.cfg.Language
	if lang == "" {
		lang = defaultLanguage
	}
	prompt += fmt.Sprintf("\n\nYour entire response must be in the language with ISO code: %s", lang)

	b.WriteString(prompt)
	b.WriteString("\n\nCode diff to summarize:\n\n")
	b.WriteString(diffContent)

	return b.String()
}

func (r *Reviewer) buildReviewCommentPrompt(req ReviewCommentRequest) string {
	var b strings.Builder

	systemMsg := r.cfg.SystemMessage
	if systemMsg == "" {
		systemMsg = `You are @goreviewer, a helpful AI assistant for code reviews.`
	}

	lang := r.cfg.Language
	if lang == "" {
		lang = defaultLanguage
	}
	systemMsg += fmt.Sprintf("\n\nYour entire response must be in the language with ISO code: %s", lang)

	b.WriteString(systemMsg)
	b.WriteString("\n\n")

	b.WriteString("A user has left a review comment on a pull request. Please respond to their comment helpfully.\n\n")

	b.WriteString("Comment details:\n")
	fmt.Fprintf(&b, "- File: %s\n", req.Path)
	fmt.Fprintf(&b, "- Line: %d\n", req.Line)
	fmt.Fprintf(&b, "- Diff hunk:\n%s\n\n", req.DiffHunk)
	fmt.Fprintf(&b, "User's comment:\n%s\n\n", req.Comment)

	b.WriteString("Provide a helpful response to their comment. This could be:\n")
	b.WriteString("- Answering a question\n")
	b.WriteString("- Explaining code changes\n")
	b.WriteString("- Acknowledging suggestions\n")
	b.WriteString("- Or responding appropriately\n\n")

	b.WriteString("Be concise, helpful, and conversational. Respond directly to their feedback.")

	return b.String()
}

func (r *Reviewer) addGitHubContext(b *strings.Builder) {
	ghCtx, err := r.ghClient.FetchContext(context.Background())
	if err != nil {
		return
	}

	if ghCtx.CheckRuns != "" {
		b.WriteString("\nGitHub Actions Check Status:\n")
		b.WriteString(ghCtx.CheckRuns)
		b.WriteString("\n\nPlease consider any failed or pending checks in your review.\n")
	}

	if ghCtx.Labels != "" {
		b.WriteString("\nAvailable Repository Labels:\n")
		b.WriteString(ghCtx.Labels)
		b.WriteString("\n\n")
	}

	if ghCtx.PRDescription != "" {
		b.WriteString("\nPR Context:\n")
		b.WriteString(ghCtx.PRDescription)
		b.WriteString("\n")
	}

	if ghCtx.Commits != "" {
		b.WriteString("\nCommit History:\n")
		b.WriteString(ghCtx.Commits)
		b.WriteString("\n")
	}

	if ghCtx.HumanComments != "" {
		b.WriteString("\nHuman Comments:\n")
		b.WriteString(ghCtx.HumanComments)
		b.WriteString("\n")
	}

	if ghCtx.PreviousReview != "" {
		b.WriteString("\nPrevious AI Review:\n")
		b.WriteString(ghCtx.PreviousReview)
		b.WriteString("\n")
	}
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

	footer := fmt.Sprintf("\n\n---\n%s\n", reviewFooter)
	if isValidBotIcon(r.cfg.BotIcon) {
		footer = fmt.Sprintf("\n\n---\n%s %s\n", r.cfg.BotIcon, reviewFooter)
	}
	result.Review += footer

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

	// Try to parse as JSON first
	var result SummaryResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// If not JSON, just return the content as walkthrough
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
