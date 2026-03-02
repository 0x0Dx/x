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

// NewReviewCmd returns the review command.
func NewReviewCmd() *cobra.Command {
	var postToGitHub bool
	var ghToken string
	var prNumber int
	var repoFullName string
	var lightModel string
	var heavyModel string
	var systemMessage string
	var language string
	var openAIBaseURL string
	var botIcon string

	cmd := &cobra.Command{
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
				Temperature: temperature,
				MaxTokens:   maxTokens,
				Debug:       verbose,

				LightModel:    lightModel,
				HeavyModel:    heavyModel,
				SystemMessage: systemMessage,
				Language:      language,
				OpenAIBaseURL: openAIBaseURL,
				BotIcon:       botIcon,
			}

			ghClient, err := getGitHubClient(ghToken, prNumber, repoFullName)
			if err != nil {
				return fmt.Errorf("failed to create GitHub client: %w", err)
			}

			r := reviewer.New(cfg, ghClient)
			result, err := r.Review(context.Background(), string(diffContent))
			if err != nil {
				jsonOut, _ := result.ToJSON()
				fmt.Println(jsonOut)
				return fmt.Errorf("review failed: %w", err)
			}

			jsonOut, err := result.ToJSON()
			if err != nil {
				return fmt.Errorf("marshal result: %w", err)
			}

			if postToGitHub && ghClient != nil && prNumber > 0 && repoFullName != "" {
				postReview(ghClient, result.Review, result.LabelsAdded, string(diffContent))
			}

			fmt.Println(jsonOut)
			return nil
		},
	}

	cmd.Flags().BoolVar(&postToGitHub, "post", false, "Post review as GitHub PR comment")
	cmd.Flags().StringVar(&ghToken, "token", "", "GitHub token (or use GITHUB_TOKEN env var)")
	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number")
	cmd.Flags().StringVar(&repoFullName, "repo", "", "Repository (owner/repo)")
	cmd.Flags().StringVar(&lightModel, "light-model", "", "Model for light tasks")
	cmd.Flags().StringVar(&heavyModel, "heavy-model", "", "Model for heavy tasks")
	cmd.Flags().StringVar(&systemMessage, "system-message", "", "System message")
	cmd.Flags().StringVar(&language, "language", "en-US", "Response language")
	cmd.Flags().StringVar(&openAIBaseURL, "openai-base-url", "", "OpenAI base URL")
	cmd.Flags().StringVar(&botIcon, "bot-icon", "", "Bot icon (emoji only)")

	return cmd
}

func postReview(ghClient *github.Client, review string, labels []string, diff string) {
	if ghClient == nil {
		return
	}
	if err := ghClient.PostReview(context.Background(), review, diff); err != nil {
		fmt.Fprintln(os.Stderr, "Warning: failed to post review:", err)
	}
	for _, label := range labels {
		if label != "" {
			if err := ghClient.AddLabel(context.Background(), label); err != nil {
				fmt.Fprintln(os.Stderr, "Warning: failed to add label:", err)
			}
		}
	}
}
