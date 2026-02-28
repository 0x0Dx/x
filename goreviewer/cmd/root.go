package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/0x0Dx/x/goreviewer/internal/github"
	"github.com/spf13/cobra"
)

var (
	verbose     bool
	model       string
	temperature float64
	maxTokens   int
)

const (
	defaultModel       = "minimax/minimax-m2.5"
	defaultTemperature = 0.1
	defaultMaxTokens   = 64000
)

// RootCmd is the root cobra command.
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
	RootCmd.PersistentFlags().StringVar(&model, "model", defaultModel, "AI model to use")
	RootCmd.PersistentFlags().Float64Var(&temperature, "temperature", defaultTemperature, "Sampling temperature")
	RootCmd.PersistentFlags().IntVar(&maxTokens, "max-tokens", defaultMaxTokens, "Maximum tokens in response")

	RootCmd.AddCommand(reviewCmd)
	RootCmd.AddCommand(commentCmd)
}

func getGitHubClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil
	}

	client := github.NewClient()
	client.SetToken(token)

	owner := os.Getenv("REPO_FULL_NAME")
	if owner == "" {
		return client
	}

	parts := splitRepo(owner)
	if len(parts) != 2 {
		return client
	}

	prNumStr := os.Getenv("PR_NUMBER")
	if prNumStr == "" {
		return client
	}

	prNum, err := strconv.Atoi(prNumStr)
	if err != nil {
		return client
	}

	client.SetPR(parts[0], parts[1], prNum)
	return client
}
