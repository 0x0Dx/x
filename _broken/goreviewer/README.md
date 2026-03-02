# goreviewer

AI-powered code reviewer using OpenAI-compatible APIs.

## Features

- **Code Review** - Analyzes security, performance, logic, and best practices
- **Summarization** - Generates walkthrough, file change table, and celebration poem
- **Release Notes** - Auto-generates release notes for PRs
- **Label Suggestions** - AI suggests relevant labels (security, bug, enhancement, etc.)
- **Review Comment Replies** - Responds to review comments conversationally
- **Multi-language** - Supports any language for responses

## CLI Usage

### Installation

```bash
go install github.com/0x0Dx/x/goreviewer@latest
```

### Commands

#### `review` - Review a diff

```bash
goreviewer review [flags]
```

#### `run` - Full workflow (review + summarize + post to GitHub)

```bash
goreviewer run [flags]
```

#### `comment` - Respond to a review comment

```bash
goreviewer comment [flags]
```

### CLI Options

| Flag | Default | Description |
|------|---------|-------------|
| `--post` | false | Post review as GitHub PR comment |
| `--pr` | - | PR number |
| `--repo` | - | Repository (owner/repo) |
| `--light-model` | gpt-3.5-turbo | Model for summarization |
| `--heavy-model` | gpt-4 | Model for code review |
| `--openai-base-url` | https://api.openai.com/v1 | OpenAI-compatible API endpoint |
| `--system-message` | - | Custom system prompt |
| `--language` | en-US | Response language (ISO code) |
| `--temperature` | 0.05 | Sampling temperature |
| `--max-tokens` | 64000 | Max response tokens |
| `--bot-icon` | - | Emoji icon for bot (e.g., 🐰) |
| `--disable-review` | false | Skip code review, only summarize |
| `--token` | - | GitHub token (or use GITHUB_TOKEN env) |
| `-v, --verbose` | false | Enable debug output |

### Examples

Basic review:
```bash
export OPENAI_API_KEY=sk-...
git diff | goreviewer review
```

Post to GitHub PR:
```bash
export OPENAI_API_KEY=sk-...
export GITHUB_TOKEN=ghp_...

git diff | goreviewer review --post --pr 123 --repo owner/repo
```

Use OpenRouter with custom model:
```bash
export OPENROUTER_API_KEY=sk-or-...
git diff | goreviewer review \
  --openai-base-url https://openrouter.ai/api/v1 \
  --heavy-model anthropic/claude-3.5-sonnet
```

Summarize only (no review):
```bash
git diff | goreviewer review --disable-review
```

---

## GitHub Action

Automated code review for every PR.

### Quick Start

Create `.github/workflows/goreviewer.yml`:

```yaml
name: GoReviewer

on:
  pull_request_target:
    types: [opened, synchronize, reopened]
  pull_request_review_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write

concurrency:
  group: ${{ github.repository }}-${{ github.event.number }}-${{ github.workflow }}
  cancel-in-progress: ${{ github.event_name != 'pull_request_review_comment' }}

jobs:
  goreviewer:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: 0x0Dx/x/goreviewer@main
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          openai_api_key: ${{ secrets.OPENAI_API_KEY }}
```

### Triggers

| Event | Description |
|-------|-------------|
| `pull_request_target` (opened, synchronize, reopened) | Runs on new/revised PRs |
| `pull_request_review_comment` (created) | Runs when AI receives a review comment |

### Configuration

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `github_token` | Yes | github.token | GitHub token |
| `openai_api_key` | Yes* | - | OpenAI API key |
| `openai_base_url` | No | https://api.openai.com/v1 | OpenAI-compatible API |
| `openai_light_model` | No | gpt-3.5-turbo | Model for summarization |
| `openai_heavy_model` | No | gpt-4 | Model for code review |
| `openai_model_temperature` | No | 0.05 | Sampling temperature |
| `openai_timeout_ms` | No | 360000 | API timeout (ms) |
| `openai_retries` | No | 5 | Retry attempts |
| `language` | No | en-US | Response language (ISO code) |
| `system_message` | No | - | Custom system prompt |
| `summarize` | No | - | Custom summarize prompt |
| `bot_icon` | No | - | Emoji icon (e.g., 🐰) |
| `fail_on_requested_changes` | No | false | Fail workflow if AI rejects |
| `debug` | No | false | Enable debug output |

### How It Works

1. **Trigger** - Runs automatically on PR events (no label needed!)
2. **Fetch** - Gets diff between base and head branch
3. **Review** - Sends diff to AI for code review
4. **Summarize** - Generates walkthrough + file changes + poem
5. **Release Notes** - Creates release notes (optional)
6. **Post** - Comments on PR with review + summary
7. **Labels** - Adds suggested labels (security, bug, etc.)

### Example: Using OpenRouter

```yaml
- uses: 0x0Dx/x/goreviewer@main
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    openai_api_key: ${{ secrets.OPENROUTER_API_KEY }}
    openai_base_url: https://openrouter.ai/api/v1
    openai_light_model: arcee-ai/trinity-mini:free
    openai_heavy_model: arcee-ai/trinity-large-preview:free
    debug: true
```

### Example: Custom Prompts

```yaml
- uses: 0x0Dx/x/goreviewer@main
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    openai_api_key: ${{ secrets.OPENAI_API_KEY }}
    system_message: |
      You are a senior software engineer reviewing code for a fintech app.
      Focus on security, compliance, and error handling.
    language: en-US
    fail_on_requested_changes: true
```

---

## Supported Models

### OpenAI

- `gpt-4o` (recommended for reviews)
- `gpt-4`
- `gpt-3.5-turbo` (fast, for summarization)

### OpenRouter

Any OpenAI-compatible model:

- `anthropic/claude-3.5-sonnet`
- `google/gemini-pro-1.5`
- `meta-llama/llama-3.1-405b-instruct`
- `minimax/minimax-m2.5`
- `arcee-ai/trinity-mini:free`
- `arcee-ai/trinity-large-preview:free`

See [openrouter.ai/models](https://openrouter.ai/models) for full list.

---
