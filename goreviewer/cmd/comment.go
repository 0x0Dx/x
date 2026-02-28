package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/0x0Dx/x/goreviewer/internal/reviewer"
	"github.com/spf13/cobra"
)

var (
	commentID  int64
	commentMsg string
)

var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Respond to a GitHub comment",
	Long:  "Generates AI response to a comment on a PR when /goreviewer is mentioned.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if commentMsg == "" {
			return fmt.Errorf("comment message is required")
		}

		cfg := reviewer.Config{
			Model:       model,
			Temperature: temperature,
			MaxTokens:   maxTokens,
			Verbose:     verbose,
		}

		ghClient := getGitHubClient()

		r := reviewer.New(cfg, ghClient)
		result, err := r.RespondToComment(context.Background(), commentMsg)
		if err != nil {
			return fmt.Errorf("generate response: %w", err)
		}

		if ghClient != nil && commentID > 0 {
			owner := os.Getenv("REPO_FULL_NAME")
			parts := strings.Split(owner, "/")
			if len(parts) == 2 {
				if err := ghClient.ReplyToComment(context.Background(), parts[0], parts[1], commentID, result.Response); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to post reply: %v\n", err)
				}
			}
		}

		fmt.Println(result.Response)
		return nil
	},
}

func init() {
	commentCmd.Flags().Int64Var(&commentID, "comment-id", 0, "ID of the comment to reply to")
	commentCmd.Flags().StringVar(&commentMsg, "message", "", "Message from the comment")
}
