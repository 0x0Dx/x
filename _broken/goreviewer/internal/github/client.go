// Package github provides GitHub API client functionality for goreviewer.
package github

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

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
		IncludePrev:    false,
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

	if c.owner == "" || c.repo == "" || c.prNumber == 0 {
		return ctxData, nil
	}

	type result struct {
		key   string
		value string
		err   error
	}

	ch := make(chan result, 6)

	var wg sync.WaitGroup

	if c.IncludeChecks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, err := c.fetchCheckRuns(ctx)
			ch <- result{"CheckRuns", status, err}
		}()
	}

	if c.IncludeLabels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			labels, err := c.fetchLabels(ctx)
			ch <- result{"Labels", labels, err}
		}()
	}

	if c.IncludeDesc {
		wg.Add(1)
		go func() {
			defer wg.Done()
			desc, err := c.fetchPRDescription(ctx)
			ch <- result{"PRDescription", desc, err}
		}()
	}

	if c.IncludeCommits {
		wg.Add(1)
		go func() {
			defer wg.Done()
			commits, err := c.fetchCommits(ctx)
			ch <- result{"Commits", commits, err}
		}()
	}

	if c.IncludePrev {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prev, err := c.fetchPreviousReviews(ctx)
			ch <- result{"PreviousReview", prev, err}
		}()
	}

	if c.IncludeHuman {
		wg.Add(1)
		go func() {
			defer wg.Done()
			comments, err := c.fetchHumanComments(ctx)
			ch <- result{"HumanComments", comments, err}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var errs []string
	for r := range ch {
		if r.err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", r.key, r.err))
			continue
		}
		switch r.key {
		case "CheckRuns":
			ctxData.CheckRuns = r.value
		case "Labels":
			ctxData.Labels = r.value
		case "PRDescription":
			ctxData.PRDescription = r.value
		case "Commits":
			ctxData.Commits = r.value
		case "PreviousReview":
			ctxData.PreviousReview = r.value
		case "HumanComments":
			ctxData.HumanComments = r.value
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

// findExistingBotComment finds the most recent comment posted by the bot.
func (c *Client) findExistingBotComment(ctx context.Context) (*github.IssueComment, error) {
	comments, _, err := c.ghClient.Issues.ListComments(ctx, c.owner, c.repo, c.prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}

	// Find the most recent AI Code Review comment
	for i := len(comments) - 1; i >= 0; i-- {
		body := comments[i].GetBody()
		if strings.HasPrefix(body, "## AI Code Review") {
			return comments[i], nil
		}
	}

	return nil, nil
}

// extractReviewHash extracts the review hash from a comment body if it exists.
func extractReviewHash(body string) string {
	re := regexp.MustCompile(`<!-- review-hash: ([a-f0-9]{64}) -->`)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// PostReview posts a new review comment to the PR or updates an existing one.
// It compares the review content hash to avoid posting duplicate reviews.
func (c *Client) PostReview(ctx context.Context, body string) error {
	// Check if there's an existing bot comment
	existingComment, err := c.findExistingBotComment(ctx)
	if err != nil {
		// If we can't check for existing comments, just create a new one
		return c.createNewComment(ctx, body)
	}

	if existingComment != nil {
		// Extract hash from existing comment
		existingHash := extractReviewHash(existingComment.GetBody())
		newHash := extractReviewHash(body)

		// If hashes match, the content is identical - skip update
		if existingHash != "" && newHash != "" && existingHash == newHash {
			// Content is identical, no need to update
			return nil
		}

		// Update the existing comment with new content
		_, _, err := c.ghClient.Issues.EditComment(ctx, c.owner, c.repo, existingComment.GetID(), &github.IssueComment{
			Body: &body,
		})
		if err != nil {
			return fmt.Errorf("update comment: %w", err)
		}
		return nil
	}

	// No existing comment found, create a new one
	return c.createNewComment(ctx, body)
}

// createNewComment creates a new comment on the PR.
func (c *Client) createNewComment(ctx context.Context, body string) error {
	_, _, err := c.ghClient.Issues.CreateComment(ctx, c.owner, c.repo, c.prNumber, &github.IssueComment{
		Body: &body,
	})
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

// ReplyToReviewComment replies to a specific review comment.
func (c *Client) ReplyToReviewComment(ctx context.Context, commentID int64, body string) error {
	_, _, err := c.ghClient.PullRequests.CreateCommentInReplyTo(ctx, c.owner, c.repo, c.prNumber, body, commentID)
	if err != nil {
		return fmt.Errorf("reply to comment: %w", err)
	}
	return nil
}

// AddLabel adds a label to the PR.
func (c *Client) AddLabel(ctx context.Context, label string) error {
	if c.owner == "" || c.repo == "" || c.prNumber == 0 {
		return nil
	}
	_, _, err := c.ghClient.Issues.AddLabelsToIssue(ctx, c.owner, c.repo, c.prNumber, []string{label})
	if err != nil {
		return fmt.Errorf("add label: %w", err)
	}
	return nil
}

// RemoveLabel removes a label from the PR.
func (c *Client) RemoveLabel(ctx context.Context, label string) error {
	if c.owner == "" || c.repo == "" || c.prNumber == 0 {
		return nil
	}
	_, err := c.ghClient.Issues.RemoveLabelForIssue(ctx, c.owner, c.repo, c.prNumber, label)
	if err != nil {
		return fmt.Errorf("remove label: %w", err)
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
