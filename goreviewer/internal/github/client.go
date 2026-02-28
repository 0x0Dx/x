// Package github provides GitHub API client functionality for goreviewer.
package github

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v70/github"
)

// Client wraps GitHub API client and configuration.
type Client struct {
	owner          string
	repo           string
	prNumber       int
	ghClient       *github.Client
	Token          string
	IncludePrev    bool
	IncludeHuman   bool
	IncludeChecks  bool
	IncludeLabels  bool
	IncludeDesc    bool
	IncludeCommits bool
}

// NewClient creates a new GitHub client with default settings.
func NewClient() *Client {
	return &Client{
		ghClient:       github.NewClient(nil),
		IncludePrev:    true,
		IncludeHuman:   true,
		IncludeChecks:  true,
		IncludeLabels:  true,
		IncludeDesc:    true,
		IncludeCommits: true,
	}
}

// SetToken sets the GitHub token for authentication.
func (c *Client) SetToken(token string) {
	c.ghClient = github.NewClient(nil).WithAuthToken(token)
	c.Token = token
}

// SetPR sets the PR owner, repo, and number for context fetching.
func (c *Client) SetPR(owner, repo string, prNumber int) {
	c.owner = owner
	c.repo = repo
	c.prNumber = prNumber
}

// FetchContext fetches GitHub context for a PR.
func (c *Client) FetchContext(ctx context.Context) (Context, error) {
	var ctxData Context
	var errs []string

	if c.owner == "" || c.repo == "" || c.prNumber == 0 {
		return ctxData, nil
	}

	if c.IncludeChecks {
		if status, err := c.fetchCheckRuns(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("checks: %v", err))
		} else {
			ctxData.CheckRuns = status
		}
	}

	if c.IncludeLabels {
		if labels, err := c.fetchLabels(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("labels: %v", err))
		} else {
			ctxData.Labels = labels
		}
	}

	if c.IncludeDesc {
		if desc, err := c.fetchPRDescription(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("pr desc: %v", err))
		} else {
			ctxData.PRDescription = desc
		}
	}

	if c.IncludeCommits {
		if commits, err := c.fetchCommits(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("commits: %v", err))
		} else {
			ctxData.Commits = commits
		}
	}

	if c.IncludePrev {
		if prev, err := c.fetchPreviousReviews(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("prev reviews: %v", err))
		} else {
			ctxData.PreviousReview = prev
		}
	}

	if c.IncludeHuman {
		if comments, err := c.fetchHumanComments(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("human comments: %v", err))
		} else {
			ctxData.HumanComments = comments
		}
	}

	if len(errs) > 0 {
		return ctxData, fmt.Errorf("errors fetching context: %s", strings.Join(errs, "; "))
	}

	return ctxData, nil
}

// Context holds GitHub PR context information.
type Context struct {
	CheckRuns      string
	Labels         string
	PRDescription  string
	Commits        string
	PreviousReview string
	HumanComments  string
}

func (c *Client) fetchCheckRuns(ctx context.Context) (string, error) {
	pr, _, err := c.ghClient.PullRequests.Get(ctx, c.owner, c.repo, c.prNumber)
	if err != nil {
		return "", fmt.Errorf("get PR: %w", err)
	}

	status, _, err := c.ghClient.Repositories.GetCombinedStatus(ctx, c.owner, c.repo, pr.GetHead().GetSHA(), nil)
	if err != nil {
		return "", fmt.Errorf("get combined status: %w", err)
	}

	var lines []string
	for _, s := range status.Statuses {
		state := s.GetState()
		lines = append(lines, fmt.Sprintf("- **%s**: %s", s.GetContext(), state))
	}

	return strings.Join(lines, "\n"), nil
}

func (c *Client) fetchLabels(ctx context.Context) (string, error) {
	labels, _, err := c.ghClient.Issues.ListLabels(ctx, c.owner, c.repo, nil)
	if err != nil {
		return "", fmt.Errorf("list labels: %w", err)
	}

	var lines []string
	for _, label := range labels {
		desc := label.GetDescription()
		if desc == "" {
			desc = "No description"
		}
		lines = append(lines, fmt.Sprintf("- **%s**: %s (color: #%s)", label.GetName(), desc, label.GetColor()))
	}

	return strings.Join(lines, "\n"), nil
}

func (c *Client) fetchPRDescription(ctx context.Context) (string, error) {
	pr, _, err := c.ghClient.PullRequests.Get(ctx, c.owner, c.repo, c.prNumber)
	if err != nil {
		return "", fmt.Errorf("get PR: %w", err)
	}

	title := pr.GetTitle()
	body := pr.GetBody()
	if body == "" {
		body = "No description provided"
	}

	return fmt.Sprintf("**PR Title**: %s\n\n**Description**:\n%s", title, body), nil
}

func (c *Client) fetchCommits(ctx context.Context) (string, error) {
	commits, _, err := c.ghClient.PullRequests.ListCommits(ctx, c.owner, c.repo, c.prNumber, nil)
	if err != nil {
		return "", fmt.Errorf("list commits: %w", err)
	}

	var lines []string
	count := 0
	for i := len(commits) - 1; i >= 0 && count < 15; i-- {
		msg := commits[i].GetCommit().GetMessage()
		if !strings.HasPrefix(msg, "Merge") {
			firstLine := strings.Split(msg, "\n")[0]
			lines = append(lines, "- "+firstLine)
			count++
		}
	}

	return strings.Join(lines, "\n"), nil
}

func (c *Client) fetchPreviousReviews(ctx context.Context) (string, error) {
	comments, _, err := c.ghClient.Issues.ListComments(ctx, c.owner, c.repo, c.prNumber, nil)
	if err != nil {
		return "", fmt.Errorf("list comments: %w", err)
	}

	for i := len(comments) - 1; i >= 0; i-- {
		body := comments[i].GetBody()
		if strings.HasPrefix(body, "## AI Code Review") {
			return fmt.Sprintf("### Previous AI Review (%s):\n%s\n---\n", comments[i].GetCreatedAt().Format("2006-01-02"), body), nil
		}
	}

	return "", nil
}

func (c *Client) fetchHumanComments(ctx context.Context) (string, error) {
	comments, _, err := c.ghClient.Issues.ListComments(ctx, c.owner, c.repo, c.prNumber, nil)
	if err != nil {
		return "", fmt.Errorf("list comments: %w", err)
	}

	var lines []string
	for _, comment := range comments {
		body := comment.GetBody()
		if !strings.HasPrefix(body, "## AI Code Review") {
			lines = append(lines, fmt.Sprintf("**%s** (%s):\n%s", comment.GetUser().GetLogin(), comment.GetCreatedAt().Format("2006-01-02"), body))
		}
	}

	return strings.Join(lines, "\n\n---\n\n"), nil
}

// PostReview posts a review comment to the PR.
func (c *Client) PostReview(ctx context.Context, body string) error {
	_, _, err := c.ghClient.Issues.CreateComment(ctx, c.owner, c.repo, c.prNumber, &github.IssueComment{
		Body: &body,
	})
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

// ReplyToComment replies to a specific comment.
func (c *Client) ReplyToComment(ctx context.Context, owner, repo string, _ int64, body string) error {
	_, _, err := c.ghClient.Issues.CreateComment(ctx, owner, repo, c.prNumber, &github.IssueComment{
		Body: &body,
	})
	if err != nil {
		return fmt.Errorf("reply to comment: %w", err)
	}
	return nil
}

// GetEnvToken returns the GitHub token from environment.
func GetEnvToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

// GetPRFromEnv returns PR details from environment variables.
func GetPRFromEnv() (owner, repo string, prNumber int, err error) {
	repoFull := os.Getenv("REPO_FULL_NAME")
	if repoFull == "" {
		return "", "", 0, nil
	}

	parts := strings.Split(repoFull, "/")
	if len(parts) != 2 {
		return "", "", 0, fmt.Errorf("invalid REPO_FULL_NAME: %s", repoFull)
	}

	prNumStr := os.Getenv("PR_NUMBER")
	if prNumStr == "" {
		return "", "", 0, nil
	}

	prNumber, err = strconv.Atoi(prNumStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid PR_NUMBER: %s", prNumStr)
	}
	return parts[0], parts[1], prNumber, nil
}
