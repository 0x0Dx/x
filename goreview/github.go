package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/go-github/v81/github"
)

type GitHubClient struct {
	client *github.Client
	owner  string
	repo   string
}

func NewGitHubClient(token string) *GitHubClient {
	client := github.NewClient(http.DefaultClient).WithAuthToken(token)
	return &GitHubClient{
		client: client,
	}
}

func (g *GitHubClient) SetRepo(repo string) error {
	parts, err := splitRepo(repo)
	if err != nil {
		return err
	}
	g.owner = parts.owner
	g.repo = parts.repo
	return nil
}

func (g *GitHubClient) GetPR(ctx context.Context, prNumber int) (*github.PullRequest, error) {
	pr, resp, err := g.client.PullRequests.Get(ctx, g.owner, g.repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("getting PR information: status %d: %w", resp.StatusCode, err)
	}
	slog.Info("got PR info", "title", pr.Title, "node-id", pr.NodeID)
	return pr, nil
}

func (g *GitHubClient) GetPRFiles(ctx context.Context, prNumber int) ([]*github.CommitFile, error) {
	files, resp, err := g.client.PullRequests.ListFiles(ctx, g.owner, g.repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("listing files for PR: status %d: %w", resp.StatusCode, err)
	}
	return files, nil
}

func (g *GitHubClient) GetPRCommits(ctx context.Context, prNumber int) ([]*github.RepositoryCommit, error) {
	commits, resp, err := g.client.PullRequests.ListCommits(ctx, g.owner, g.repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("listing commits for PR: status %d: %w", resp.StatusCode, err)
	}
	return commits, nil
}

func (g *GitHubClient) CreateReview(ctx context.Context, prNumber int, body, event, commitID string) (*github.PullRequestReview, error) {
	comment, resp, err := g.client.PullRequests.CreateReview(ctx, g.owner, g.repo, prNumber, &github.PullRequestReviewRequest{
		Body:     &body,
		Event:    &event,
		CommitID: &commitID,
	})
	if err != nil {
		return nil, fmt.Errorf("creating PR review: status %d: %w", resp.StatusCode, err)
	}
	return comment, nil
}

type repoParts struct {
	owner string
	repo  string
}

func splitRepo(repo string) (repoParts, error) {
	var parts repoParts
	split := splitN(repo, "/", 2)
	if len(split) != 2 {
		return parts, fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
	}
	parts.owner = split[0]
	parts.repo = split[1]
	return parts, nil
}

func splitN(s, sep string, n int) []string {
	result := []string{}
	rem := s
	for i := 0; i < n-1; i++ {
		idx := index(rem, sep)
		if idx == -1 {
			break
		}
		result = append(result, rem[:idx])
		rem = rem[idx+len(sep):]
	}
	result = append(result, rem)
	return result
}

func index(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
