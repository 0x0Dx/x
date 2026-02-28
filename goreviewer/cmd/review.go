// Package cmd provides CLI commands for goreviewer.
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/0x0Dx/x/goreviewer/internal/github"
	"github.com/0x0Dx/x/goreviewer/internal/reviewer"
	"github.com/spf13/cobra"
)

var (
	postToGitHub bool
	ghToken      string
	prNumber     int
	repoFullName string
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review code diff using AI",
	Long:  "Analyzes a code diff and provides an AI-powered code review. Reads diff from stdin by default.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		diffContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}

		cfg := reviewer.Config{
			Model:       model,
			Temperature: temperature,
			MaxTokens:   maxTokens,
			Verbose:     verbose,
		}

		var ghClient *github.Client
		if ghToken != "" || (ghToken == "" && github.GetEnvToken() != "") {
			ghClient = github.NewClient()
			token := ghToken
			if token == "" {
				token = github.GetEnvToken()
			}
			ghClient.SetToken(token)

			if prNumber > 0 && repoFullName != "" {
				parts := splitRepo(repoFullName)
				if len(parts) == 2 {
					ghClient.SetPR(parts[0], parts[1], prNumber)
				}
			}
		}

		r := reviewer.New(cfg, ghClient)
		result, err := r.Review(context.Background(), string(diffContent))
		if err != nil {
			if !postToGitHub {
				fmt.Fprintf(os.Stderr, "Review failed: %v\n", err)
			}
			jsonOut, _ := result.ToJSON()
			fmt.Println(jsonOut)
			return err
		}

		jsonOut, err := result.ToJSON()
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}

		if postToGitHub && ghClient != nil && prNumber > 0 && repoFullName != "" {
			if err := ghClient.PostReview(context.Background(), result.Review); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to post review to GitHub: %v\n", err)
			}
		}

		fmt.Println(jsonOut)
		return nil
	},
}

func splitRepo(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func init() {
	reviewCmd.Flags().BoolVar(&postToGitHub, "post", false, "Post review as GitHub PR comment")
	reviewCmd.Flags().StringVar(&ghToken, "token", "", "GitHub token (or use GITHUB_TOKEN env var)")
	reviewCmd.Flags().IntVar(&prNumber, "pr", 0, "PR number")
	reviewCmd.Flags().StringVar(&repoFullName, "repo", "", "Repository (owner/repo)")
	RootCmd.AddCommand(reviewCmd)
}
