package cmd

import (
	"fmt"
	"strings"

	"github.com/0x0Dx/x/goreviewer/internal/github"
	"github.com/spf13/cobra"
)

var (
	verbose     bool
	temperature float64
	maxTokens   int
)

const (
	defaultTemperature = 0.1
	defaultMaxTokens   = 64000
)

// RootCmd is the root command.
var RootCmd = &cobra.Command{
	Use:   "goreviewer",
	Short: "AI-powered code reviewer",
	Long:  "An AI code reviewer that analyzes code diffs and provides comprehensive reviews using OpenRouter API.",
}

// Execute runs the root command.
func Execute() error {
	err := RootCmd.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}
	return nil
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	RootCmd.PersistentFlags().Float64Var(&temperature, "temperature", defaultTemperature, "Sampling temperature")
	RootCmd.PersistentFlags().IntVar(&maxTokens, "max-tokens", defaultMaxTokens, "Maximum tokens in response")

	RootCmd.AddCommand(NewReviewCmd())
	RootCmd.AddCommand(NewRunCmd())
	RootCmd.AddCommand(NewCommentCmd())
}

func splitRepo(s string) []string {
	return strings.SplitN(s, "/", 2)
}

func getGitHubClient(ghToken string, prNumber int, repoFullName string) (*github.Client, error) {
	if ghToken == "" && github.GetEnvToken() == "" {
		return nil, nil
	}

	token := ghToken
	if token == "" {
		token = github.GetEnvToken()
	}

	if len(token) < 10 {
		return nil, fmt.Errorf("invalid GitHub token: token too short")
	}

	ghClient := github.NewClient()
	ghClient.SetToken(token)

	if prNumber > 0 && repoFullName != "" {
		parts := splitRepo(repoFullName)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repo format: expected owner/repo")
		}
		ghClient.SetPR(parts[0], parts[1], prNumber)
	}

	return ghClient, nil
}
