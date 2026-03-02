package cmd

import (
	"context"
	"fmt"

	"github.com/0x0Dx/x/goreviewer/internal/reviewer"
	"github.com/spf13/cobra"
)

// NewCommentCmd returns the comment command.
func NewCommentCmd() *cobra.Command {
	var ghToken string
	var prNumber int
	var repoFullName string
	var commentID int64
	var commentBody string
	var diffHunk string
	var commentFile string
	var commentLine int
	var lightModel string
	var heavyModel string
	var language string
	var openAIBaseURL string

	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Respond to a review comment",
		Long:  "Responds to a GitHub PR review comment using AI.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := reviewer.Config{
				Temperature:   temperature,
				MaxTokens:     maxTokens,
				Debug:         verbose,
				LightModel:    lightModel,
				HeavyModel:    heavyModel,
				SystemMessage: "",
				Language:      language,
				OpenAIBaseURL: openAIBaseURL,
			}

			ghClient, err := getGitHubClient(ghToken, prNumber, repoFullName)
			if err != nil {
				return fmt.Errorf("failed to create GitHub client: %w", err)
			}

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

	cmd.Flags().Int64Var(&commentID, "comment-id", 0, "Comment ID to reply to")
	cmd.Flags().StringVar(&commentBody, "comment", "", "Comment body")
	cmd.Flags().StringVar(&diffHunk, "diff-hunk", "", "Diff hunk context")
	cmd.Flags().StringVar(&commentFile, "file", "", "File path")
	cmd.Flags().IntVar(&commentLine, "line", 0, "Line number")
	cmd.Flags().StringVar(&repoFullName, "repo", "", "Repository (owner/repo)")
	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number")
	cmd.Flags().StringVar(&lightModel, "light-model", "", "Model for chat")
	cmd.Flags().StringVar(&heavyModel, "heavy-model", "", "Model for chat")
	cmd.Flags().StringVar(&language, "language", "en-US", "Response language")
	cmd.Flags().StringVar(&openAIBaseURL, "openai-base-url", "", "OpenAI base URL")
	cmd.Flags().StringVar(&ghToken, "token", "", "GitHub token")

	return cmd
}
