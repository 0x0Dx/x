package cmd

import (
	"fmt"

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
}
