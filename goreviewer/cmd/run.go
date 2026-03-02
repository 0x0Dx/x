package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/0x0Dx/x/goreviewer/internal/reviewer"
	"github.com/spf13/cobra"
)

// NewRunCmd returns the run command.
func NewRunCmd() *cobra.Command {
	var ghToken string
	var prNumber int
	var repoFullName string
	var lightModel string
	var heavyModel string
	var systemMessage string
	var summarizePrompt string
	var language string
	var openAIBaseURL string
	var botIcon string
	var disableReview bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run full review workflow",
		Long:  "Runs the complete review workflow including diff generation, review, and posting to GitHub.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			diffContent, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}

			if prNumber == 0 || repoFullName == "" {
				return fmt.Errorf("PR number and repo are required")
			}

			cfg := reviewer.Config{
				Temperature:     temperature,
				MaxTokens:       maxTokens,
				Debug:           verbose,
				LightModel:      lightModel,
				HeavyModel:      heavyModel,
				SystemMessage:   systemMessage,
				SummarizePrompt: summarizePrompt,
				Language:        language,
				OpenAIBaseURL:   openAIBaseURL,
				BotIcon:         botIcon,
				DisableReview:   disableReview,
			}

			ghClient, err := getGitHubClient(ghToken, prNumber, repoFullName)
			if err != nil {
				return fmt.Errorf("failed to create GitHub client: %w", err)
			}

			r := reviewer.New(cfg, ghClient)
			result, err := r.Review(context.Background(), string(diffContent))
			if err != nil {
				return fmt.Errorf("review failed: %w", err)
			}

			if ghClient != nil {
				postReview(ghClient, result.Review, result.LabelsAdded)
				_ = ghClient.RemoveLabel(context.Background(), "ai_code_review")
			}

			fmt.Println("Review completed")
			return nil
		},
	}

	cmd.Flags().StringVar(&ghToken, "token", "", "GitHub token (or use GITHUB_TOKEN env var)")
	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number")
	cmd.Flags().StringVar(&repoFullName, "repo", "", "Repository (owner/repo)")
	cmd.Flags().StringVar(&lightModel, "light-model", "", "Model for light tasks")
	cmd.Flags().StringVar(&heavyModel, "heavy-model", "", "Model for heavy tasks")
	cmd.Flags().StringVar(&systemMessage, "system-message", "", "System message")
	cmd.Flags().StringVar(&summarizePrompt, "summarize-prompt", "", "Summarize prompt")
	cmd.Flags().StringVar(&language, "language", "en-US", "Response language")
	cmd.Flags().StringVar(&openAIBaseURL, "openai-base-url", "", "OpenAI base URL")
	cmd.Flags().StringVar(&botIcon, "bot-icon", "", "Bot icon (emoji only)")
	cmd.Flags().BoolVar(&disableReview, "disable-review", false, "Only provide summary")

	return cmd
}
