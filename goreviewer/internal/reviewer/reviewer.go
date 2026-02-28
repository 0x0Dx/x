// Package reviewer provides AI code review functionality.
package reviewer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/0x0Dx/x/goreviewer/internal/github"
)

const (
	reviewHeader      = "## AI Code Review"
	reviewFooter      = "*Review by [GoReviewer](https://github.com/0x0Dx/x/tree/main/goreviewer)*"
	maxDiffSize       = 5000000
	defaultLightModel = "gpt-3.5-turbo"
	defaultHeavyModel = "gpt-4"
	defaultBaseURL    = "https://api.openai.com/v1"
	defaultLanguage   = "en-US"
)

// Config holds reviewer configuration.
type Config struct {
	Debug               bool
	DisableReview       bool
	DisableReleaseNotes bool
	OpenAIBaseURL       string
	LightModel          string
	HeavyModel          string
	Temperature         float64
	Retries             int
	TimeoutMS           int
	SystemMessage       string
	SummarizePrompt     string
	Language            string
	BotIcon             string
	MaxTokens           int
}

// ReviewResponse represents the AI review response.
type ReviewResponse struct {
	Review           string   `json:"review"`
	FailPassWorkflow string   `json:"fail_pass_workflow"`
	LabelsAdded      []string `json:"labels_added"`
	ReleaseNotes     string   `json:"release_notes,omitempty"`
}

// SummaryResponse represents the summary response.
type SummaryResponse struct {
	Walkthrough string       `json:"walkthrough"`
	Changes     []fileChange `json:"changes"`
	Poem        string       `json:"poem"`
}

type fileChange struct {
	Files   string `json:"files"`
	Summary string `json:"summary"`
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

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return errorResponse("Missing OPENAI_API_KEY environment variable"), errors.New("missing API key")
	}

	prompt := r.buildReviewPrompt(diffContent)

	repoFull := os.Getenv("REPO_FULL_NAME")
	referer := fmt.Sprintf("https://github.com/%s", repoFull)

	baseURL := r.cfg.OpenAIBaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	model := r.selectModel(true)
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

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return SummaryResponse{}, errors.New("missing API key")
	}

	model := r.selectModel(false)

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
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("missing API key")
	}

	model := r.selectModel(true)

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
			model = defaultHeavyModel
		}
		return model
	}
	model := r.cfg.LightModel
	if model == "" {
		model = defaultLightModel
	}
	return model
}

// ToJSON converts the ReviewResponse to JSON string.
func (r *ReviewResponse) ToJSON() (string, error) {
	b, err := jsonMarshal(r)
	return string(b), err
}
