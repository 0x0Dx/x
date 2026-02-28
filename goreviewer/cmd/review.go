// Package cmd provides CLI commands for goreviewer.
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/0x0Dx/x/goreviewer/internal/github"
	"github.com/0x0Dx/x/goreviewer/internal/reviewer"
	"github.com/spf13/cobra"
)

var (
	postToGitHub    bool
	ghToken         string
	prNumber        int
	repoFullName    string
	disableReview   bool
	lightModel      string
	heavyModel      string
	systemMessage   string
	summarizePrompt string
	language        string
	openAIBaseURL   string
	botIcon         string
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
			Debug:       verbose,

			LightModel:    lightModel,
			HeavyModel:    heavyModel,
			SystemMessage: systemMessage,
			Language:      language,
			OpenAIBaseURL: openAIBaseURL,
			BotIcon:       botIcon,

			DisableReview: disableReview,
		}

		ghClient := getGitHubClientForPR()

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
			postReview(ghClient, result.Review, result.LabelsAdded)
		}

		fmt.Println(jsonOut)
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run full review workflow",
	Long:  "Runs the complete review workflow including diff generation, review, and posting to GitHub.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg := reviewer.Config{
			Model:           model,
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

		ghClient := getGitHubClient()
		r := reviewer.New(cfg, ghClient)

		diffContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}

		if prNumber == 0 || repoFullName == "" {
			return fmt.Errorf("PR number and repo are required")
		}

		parts := splitRepo(repoFullName)
		if len(parts) != 2 {
			return fmt.Errorf("invalid repo format: expected owner/repo")
		}

		result, err := r.Review(context.Background(), string(diffContent))
		if err != nil {
			if ghClient != nil {
				_ = ghClient.PostReview(context.Background(), result.Review)
			}
			return fmt.Errorf("review failed: %w", err)
		}

		if ghClient != nil {
			postReview(ghClient, result.Review, result.LabelsAdded)
		}

		fmt.Println("✅ Review completed")
		return nil
	},
}

func splitRepo(s string) []string {
	return strings.SplitN(s, "/", 2)
}

func getGitHubClientForPR() *github.Client {
	if ghToken == "" && github.GetEnvToken() == "" {
		return nil
	}

	ghClient := github.NewClient()
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

	return ghClient
}

func postReview(ghClient *github.Client, review string, labels []string) {
	if ghClient == nil {
		return
	}
	_ = ghClient.PostReview(context.Background(), review)
	for _, label := range labels {
		if label != "" {
			_ = ghClient.AddLabel(context.Background(), label)
		}
	}
}

func init() {
	reviewCmd.Flags().BoolVar(&postToGitHub, "post", false, "Post review as GitHub PR comment")
	reviewCmd.Flags().StringVar(&ghToken, "token", "", "GitHub token (or use GITHUB_TOKEN env var)")
	reviewCmd.Flags().IntVar(&prNumber, "pr", 0, "PR number")
	reviewCmd.Flags().StringVar(&repoFullName, "repo", "", "Repository (owner/repo)")
	reviewCmd.Flags().StringVar(&lightModel, "light-model", "", "Model for light tasks")
	reviewCmd.Flags().StringVar(&heavyModel, "heavy-model", "", "Model for heavy tasks")
	reviewCmd.Flags().StringVar(&systemMessage, "system-message", "", "System message")
	reviewCmd.Flags().StringVar(&language, "language", "en-US", "Response language")
	reviewCmd.Flags().StringVar(&openAIBaseURL, "openai-base-url", "", "OpenAI base URL")
	reviewCmd.Flags().StringVar(&botIcon, "bot-icon", "", "Bot icon (emoji only)")
	RootCmd.AddCommand(reviewCmd)

	runCmd.Flags().IntVar(&prNumber, "pr", 0, "PR number")
	runCmd.Flags().StringVar(&repoFullName, "repo", "", "Repository (owner/repo)")
	runCmd.Flags().StringVar(&lightModel, "light-model", "", "Model for light tasks")
	runCmd.Flags().StringVar(&heavyModel, "heavy-model", "", "Model for heavy tasks")
	runCmd.Flags().StringVar(&systemMessage, "system-message", "", "System message")
	runCmd.Flags().StringVar(&summarizePrompt, "summarize-prompt", "", "Summarize prompt")
	runCmd.Flags().StringVar(&language, "language", "en-US", "Response language")
	runCmd.Flags().StringVar(&openAIBaseURL, "openai-base-url", "", "OpenAI base URL")
	runCmd.Flags().StringVar(&botIcon, "bot-icon", "", "Bot icon (emoji only)")
	runCmd.Flags().BoolVar(&disableReview, "disable-review", false, "Only provide summary")
	RootCmd.AddCommand(runCmd)
}

var (
	commentID   int64
	commentBody string
	diffHunk    string
	commentFile string
	commentLine int
)

var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Respond to a review comment",
	Long:  "Responds to a GitHub PR review comment using AI.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg := reviewer.Config{
			Model:         model,
			Temperature:   temperature,
			MaxTokens:     maxTokens,
			Debug:         verbose,
			LightModel:    lightModel,
			HeavyModel:    heavyModel,
			SystemMessage: systemMessage,
			Language:      language,
			OpenAIBaseURL: openAIBaseURL,
		}

		ghClient := getGitHubClientForPR()

		r := reviewer.New(cfg, ghClient)

		req := reviewer.ReviewCommentRequest{
			Comment:   commentBody,
			DiffHunk:  diffHunk,
			Path:      commentFile,
			Line:      commentLine,
			PRNumber:  prNumber,
			RepoOwner: repoFullName,
			RepoName:  "",
		}

		resp, err := r.RespondToReviewComment(context.Background(), req)
		if err != nil {
			return fmt.Errorf("respond to comment failed: %w", err)
		}

		if ghClient != nil && commentID > 0 {
			if err := ghClient.ReplyToReviewComment(context.Background(), commentID, resp); err != nil {
				return fmt.Errorf("reply to comment: %w", err)
			}
			fmt.Println("Replied to comment")
		} else {
			fmt.Println(resp)
		}

		return nil
	},
}

func init() {
	commentCmd.Flags().Int64Var(&commentID, "comment-id", 0, "Comment ID to reply to")
	commentCmd.Flags().StringVar(&commentBody, "comment", "", "Comment body")
	commentCmd.Flags().StringVar(&diffHunk, "diff-hunk", "", "Diff hunk context")
	commentCmd.Flags().StringVar(&commentFile, "file", "", "File path")
	commentCmd.Flags().IntVar(&commentLine, "line", 0, "Line number")
	commentCmd.Flags().StringVar(&repoFullName, "repo", "", "Repository (owner/repo)")
	commentCmd.Flags().IntVar(&prNumber, "pr", 0, "PR number")
	commentCmd.Flags().StringVar(&lightModel, "light-model", "", "Model for chat")
	commentCmd.Flags().StringVar(&heavyModel, "heavy-model", "", "Model for chat")
	commentCmd.Flags().StringVar(&language, "language", "en-US", "Response language")
	commentCmd.Flags().StringVar(&openAIBaseURL, "openai-base-url", "", "OpenAI base URL")
	commentCmd.Flags().StringVar(&ghToken, "token", "", "GitHub token")
	RootCmd.AddCommand(commentCmd)
}
