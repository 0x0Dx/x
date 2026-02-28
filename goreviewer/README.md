# goreviewer

AI-powered code reviewer using OpenAI API.

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
      --disable-review        Only provide summary (default false)
      --disable-release-notes Disable release notes (default false)
      --language string       Response language (default "en-US")
      --light-model string    Model for light tasks (e.g., summarization)
      --heavy-model string    Model for heavy tasks (e.g., code review)
      --max-tokens int        Maximum tokens in response (default 64000)
      --model string          AI model to use (legacy option)
      --openai-base-url string OpenAI-compatible API base URL
      --post                  Post review as GitHub PR comment
      --pr int                PR number
      --repo string           Repository (owner/repo)
      --system-message string Custom system message
      --temperature float     Sampling temperature (default 0.05)
      --token string          GitHub token (or use GITHUB_TOKEN env var)
  -v, --verbose               Enable verbose output
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `OPENAI_API_KEY` | Yes | Get from [openai.com](https://platform.openai.com/api-keys) |
| `OPENROUTER_API_KEY` | No | Alternative to OpenAI (supports many models) |
| `GITHUB_TOKEN` | For `--post` | GitHub token for posting comments |
| `LIGHT_MODEL` | No | Model for light tasks (default: gpt-3.5-turbo) |
| `HEAVY_MODEL` | No | Model for heavy tasks (default: gpt-4) |
| `TEMPERATURE` | No | Sampling temperature (default: 0.05) |

### Examples

Basic review:
```bash
export OPENAI_API_KEY=your-key-here
git diff | goreviewer review
```

Post to GitHub PR:
```bash
export OPENAI_API_KEY=your-key-here
export GITHUB_TOKEN=ghp_xxx

git diff | goreviewer review --post --pr 123 --repo owner/repo
```

Use different model:
```bash
git diff | goreviewer review --heavy-model gpt-4o
```

Use custom OpenAI-compatible endpoint:
```bash
git diff | goreviewer review --openai-base-url https://openrouter.ai/api/v1 --heavy-model anthropic/claude-3.5-sonnet
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
          openai_api_key: ${{ secrets.OPENAI_API_KEY }}
```

### Configuration

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `github_token` | Yes | `github.token` | GitHub token |
| `openai_api_key` | Yes | - | OpenAI API key |
| `openai_base_url` | No | `https://api.openai.com/v1` | OpenAI-compatible API endpoint |
| `openai_light_model` | No | `gpt-3.5-turbo` | Model for summarization |
| `openai_heavy_model` | No | `gpt-4` | Model for code review |
| `openai_model_temperature` | No | `0.05` | Sampling temperature |
| `openai_timeout_ms` | No | `360000` | API timeout (ms) |
| `openai_retries` | No | `5` | Number of retries |
| `max_files` | No | `150` | Max files to review |
| `language` | No | `en-US` | Response language |
| `disable_review` | No | `false` | Skip code review, only summarize |
| `disable_release_notes` | No | `false` | Disable release notes generation |
| `path_filters` | No | See action.yaml | Files to include/exclude |
| `bot_icon` | No | - | Emoji icon for the bot |
| `fail_on_requested_changes` | No | `false` | Fail workflow if AI requests changes |
| `debug` | No | `false` | Debug output |

### How It Works

1. Add `ai_code_review` label to a PR
2. Action fetches diff and sends to AI
3. AI reviews code (security, performance, quality)
4. Review posted as PR comment
5. Suggested labels added
6. Trigger label removed for re-triggering

### Secrets

- `OPENAI_API_KEY` - Get from [platform.openai.com](https://platform.openai.com/api-keys)
- `OPENROUTER_API_KEY` - Alternative (supports many models) from [openrouter.ai](https://openrouter.ai)

### Example with Options

```yaml
- name: GoReviewer
  uses: 0x0Dx/x/goreviewer@main
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    openai_api_key: ${{ secrets.OPENAI_API_KEY }}
    openai_heavy_model: gpt-4o
    openai_model_temperature: 0.05
    language: en-US
    fail_on_requested_changes: true
```

---

## Supported Models

### Default (OpenAI)

- `gpt-3.5-turbo` (default for light tasks)
- `gpt-4` (default for heavy tasks)
- `gpt-4o`
- `gpt-4o-mini`

### OpenRouter (via openai_base_url)

Set `openai_base_url` to `https://openrouter.ai/api/v1` and use any model:

- `anthropic/claude-3.5-sonnet`
- `google/gemini-pro-1.5`
- `meta-llama/llama-3.1-405b-instruct`
- `minimax/minimax-m2.5`

See [openrouter.ai/models](https://openrouter.ai/models) for full list.
