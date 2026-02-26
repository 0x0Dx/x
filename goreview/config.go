// Package main provides an AI-powered code reviewer for GitHub Pull Requests.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Config struct {
	GitHubRepo    string
	GitHubSHA     string
	GitHubToken   string
	OpenAIAPIBase string
	OpenAIAPIKey  string
	OpenAIModel   string
	PRNumber      int
}

var cfg Config

var rootCmd = &cobra.Command{
	Use:   "goreview",
	Short: "AI-powered code reviewer for GitHub Pull Requests",
	Long:  "Reviews pull requests using AI and submits reviews to GitHub",
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfg.GitHubRepo == "" {
			return fmt.Errorf("GitHub repository is required (use --github-repo)")
		}
		if cfg.GitHubSHA == "" {
			return fmt.Errorf("GitHub commit SHA is required (use --github-sha)")
		}
		if cfg.OpenAIAPIKey == "" {
			return fmt.Errorf("OpenAI API key is required (use --openai-api-key)")
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.GitHubRepo, "github-repo", "r", "", "GitHub repository (owner/repo)")
	rootCmd.Flags().StringVarP(&cfg.GitHubSHA, "github-sha", "s", "", "GitHub commit SHA")
	rootCmd.Flags().StringVarP(&cfg.GitHubToken, "github-token", "t", "", "GitHub API token")
	rootCmd.Flags().StringVarP(&cfg.OpenAIAPIBase, "openai-api-base", "b", "", "OpenAI API base URL")
	rootCmd.Flags().StringVarP(&cfg.OpenAIAPIKey, "openai-api-key", "k", "", "OpenAI API key")
	rootCmd.Flags().StringVarP(&cfg.OpenAIModel, "openai-model", "m", "gpt-4o", "OpenAI model")
	rootCmd.Flags().IntVarP(&cfg.PRNumber, "pr-number", "p", 0, "Pull request number")
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}
	return nil
}

func GetConfig() Config {
	return cfg
}

func GetGitHubToken() string {
	if cfg.GitHubToken != "" {
		return cfg.GitHubToken
	}
	return os.Getenv("GITHUB_TOKEN")
}

func GetOpenAIAPIKey() string {
	if cfg.OpenAIAPIKey != "" {
		return cfg.OpenAIAPIKey
	}
	return os.Getenv("OPENAI_API_KEY")
}
