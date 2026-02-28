# AGENTS.md - goreviewer

AI-powered code reviewer using OpenAI-compatible APIs.

## Project Overview

- **Type**: CLI tool + GitHub Action
- **Entry**: `main.go`
- **Commands**: `cmd/root.go`, `cmd/review.go`, `cmd/run.go`, `cmd/comment.go`
- **Core**: `internal/reviewer/`, `internal/github/client.go`

## Commands

```bash
# Review a diff
goreviewer review [flags]

# Full workflow (review + summarize + post to GitHub)
goreviewer run [flags]

# Respond to a review comment
goreviewer comment [flags]
```

## Key Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--light-model` | gpt-3.5-turbo | Model for summarization |
| `--heavy-model` | gpt-4 | Model for code review |
| `--openai-base-url` | https://api.openai.com/v1 | API endpoint |
| `--language` | en-US | Response language |

## Environment Variables (CLI)

| Variable | Required | Description |
|----------|----------|-------------|
| `OPENAI_API_KEY` | Yes* | OpenAI API key |
| `OPENROUTER_API_KEY` | Yes* | OpenRouter API key |
| `GITHUB_TOKEN` | For --post | GitHub token |

*At least one required. OPENROUTER_API_KEY takes priority.

## Testing

```bash
cd goreviewer
go test -v -run TestName ./...
go test -cover ./...
```

## Code Structure

```
goreviewer/
├── main.go           # Entry point
├── cmd/
│   ├── root.go      # Root command & helpers
│   ├── review.go    # review command
│   ├── run.go       # run command
│   └── comment.go  # comment command
├── internal/
│   ├── reviewer/    # Core review logic
│   │   ├── reviewer.go  # Config, types, main methods
│   │   ├── api.go      # API calls & response parsing
│   │   ├── prompts.go  # Prompt building
│   │   └── helpers.go  # Helper functions
│   └── github/      # GitHub API client
│       └── client.go
└── action.yaml      # GitHub Action
```

## Important Files

- `internal/reviewer/reviewer.go` - Config, types, main review methods
- `internal/reviewer/api.go` - OpenAI API calls and response parsing
- `internal/reviewer/prompts.go` - Prompt building logic
- `internal/github/client.go` - GitHub API interactions
- `cmd/` - CLI command handlers
- `action.yaml` - GitHub Action definition

## Error Handling

- Use `fmt.Errorf` with `%w` for wrapped errors
- Return errors early, avoid nested handling
- Error responses include user-friendly messages
