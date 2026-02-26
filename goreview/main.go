package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("GoReview v0.1.0\n")
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)

	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg := GetConfig()
	ctx := context.Background()

	githubToken := GetGitHubToken()
	openaiAPIKey := GetOpenAIAPIKey()

	slog.Info("starting goreview",
		"github-repo", cfg.GitHubRepo,
		"github-sha", cfg.GitHubSHA,
		"pr-number", cfg.PRNumber,
		"model", cfg.OpenAIModel,
	)

	ghClient := NewGitHubClient(githubToken)
	if err := ghClient.SetRepo(cfg.GitHubRepo); err != nil {
		slog.Error("failed to set repo", "error", err)
		os.Exit(1)
	}

	pr, err := ghClient.GetPR(ctx, cfg.PRNumber)
	if err != nil {
		slog.Error("failed to get PR", "error", err)
		os.Exit(1)
	}

	files, err := ghClient.GetPRFiles(ctx, cfg.PRNumber)
	if err != nil {
		slog.Error("failed to get PR files", "error", err)
		os.Exit(1)
	}

	openaiClient := NewOpenAIClient(openaiAPIKey, cfg.OpenAIAPIBase, cfg.OpenAIModel)

	prFiles := make([]*File, len(files))
	for i, f := range files {
		prFiles[i] = &File{
			Filename: f.GetFilename(),
			Status:   f.GetStatus(),
		}
	}

	userPrompt, err := buildPrompt(prFiles, pr.GetTitle(), pr.GetBody())
	if err != nil {
		slog.Error("failed to build prompt", "error", err)
		os.Exit(1)
	}

	review, err := openaiClient.CreateReview(ctx, systemPrompt, userPrompt)
	if err != nil {
		slog.Error("failed to create review", "error", err)
		os.Exit(1)
	}

	fmt.Println(review)
	fmt.Println("Review completed successfully")
}

func buildPrompt(files []*File, title, body string) (string, error) {
	var buf bytes.Buffer

	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	data := struct {
		Title string
		Body  string
		Files []*File
	}{
		Title: title,
		Body:  body,
		Files: files,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

type File struct {
	Filename string
	Status   string
}

//go:embed prompt.tmpl
var promptTemplate string

//go:embed systemprompt.txt
var systemPrompt string
