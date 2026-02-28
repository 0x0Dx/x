# goreviewer

AI-powered code reviewer using OpenRouter API.

## CLI Usage

### Installation

```bash
go install github.com/0x0Dx/x/goreviewer@latest
```

### Basic Usage

Review a diff from a file:
```bash
git diff HEAD~1 --no-color | goreviewer review
```

Or from a file:
```bash
cat diff.txt | goreviewer review
```

### Options

```bash
goreviewer review [flags]

Flags:
      --max-tokens int      Maximum tokens in response (default 64000)
      --model string        AI model to use (default "minimax/minimax-m2.5")
      --post                Post review as GitHub PR comment
      --pr int             PR number
      --repo string        Repository (owner/repo)
      --temperature float  Sampling temperature (default 0.1)
      --token string       GitHub token (or use GITHUB_TOKEN env var)
  -v, --verbose           Enable verbose output
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `OPENROUTER_API_KEY` | Yes | Get from [openrouter.ai](https://openrouter.ai) |
| `GITHUB_TOKEN` | For `--post` | GitHub token for posting comments |
| `MODEL` | No | AI model (default: minimax/minimax-m2.5) |
| `TEMPERATURE` | No | Sampling temperature (default: 0.1) |

### Examples

Basic review:
```bash
export OPENROUTER_API_KEY=your-key-here
git diff | goreviewer review
```

Post to GitHub PR:
```bash
export OPENROUTER_API_KEY=your-key-here
export GITHUB_TOKEN=ghp_xxx

git diff | goreviewer review --post --pr 123 --repo owner/repo
```

Use different model:
```bash
git diff | goreviewer review --model anthropic/claude-3.5-sonnet
```

---

## GitHub Action

An AI code reviewer action for GitHub workflows.

### Quick Start

```yaml
name: GoReviewer

on:
  pull_request:
    types: [labeled]

jobs:
  goreviewer:
    name: GoReviewer
    runs-on: ubuntu-latest
    if: github.event.action == 'labeled' && github.event.label.name == 'ai_code_review'
    permissions:
      contents: read
      pull-requests: write
      issues: write
    steps:
      - name: AI Code Review
        uses: 0x0Dx/x/goreviewer@main
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          openrouter_api_key: ${{ secrets.OPENROUTER_API_KEY }}
```

### Configuration

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `github_token` | Yes | `github.token` | GitHub token |
| `openrouter_api_key` | Yes | - | OpenRouter API key |
| `model` | No | `minimax/minimax-m2.5` | AI model |
| `temperature` | No | `0.1` | Sampling temperature |
| `max_tokens` | No | `64000` | Max response tokens |
| `max_diff_size` | No | `800000` | Max diff size (bytes) |
| `exclude_patterns` | No | `*.lock,*.min.js,*.min.css,package-lock.json,yarn.lock` | Files to skip |
| `fail_on_requested_changes` | No | `false` | Fail workflow if AI rejects |
| `debug_mode` | No | `false` | Debug output |

### How It Works

1. Add `ai_code_review` label to a PR
2. Action fetches diff and sends to OpenRouter
3. AI reviews code (security, performance, quality)
4. Review posted as PR comment
5. Suggested labels added
6. Trigger label removed for re-triggering

### Secrets

- `OPENROUTER_API_KEY` - Get from [openrouter.ai](https://openrouter.ai)

### Example with Options

```yaml
- name: GoReviewer
  uses: 0x0Dx/x/goreviewer@main
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    openrouter_api_key: ${{ secrets.OPENROUTER_API_KEY }}
    model: anthropic/claude-3.5-sonnet
    temperature: 0.1
    max_tokens: 64000
    fail_on_requested_changes: true
```

---

## Supported Models

Popular models on OpenRouter:

- `minimax/minimax-m2.5` (default, fast)
- `anthropic/claude-3.5-sonnet`
- `openai/gpt-4o`
- `google/gemini-pro-1.5`
- `meta-llama/llama-3.1-405b-instruct`

See [openrouter.ai/models](https://openrouter.ai/models) for full list.
