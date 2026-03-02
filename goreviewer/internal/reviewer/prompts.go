package reviewer

import (
	"context"
	"fmt"
	"strings"
)

func (r *Reviewer) buildReviewPrompt(diffContent string) string {
	var b strings.Builder

	systemMsg := r.cfg.SystemMessage
	if systemMsg == "" {
		systemMsg = `You are @goreviewer, a language model trained by OpenAI. 
Your purpose is to act as a highly experienced software engineer and provide a thorough review of the code hunks and suggest code snippets to improve key areas such:
- Logic
- Security
- Performance
- Data races
- Consistency
- Error handling
- Maintainability
- Modularity
- Complexity
- Optimization
- Best practices: DRY, SOLID, KISS

Do not comment on minor code style issues, missing comments/documentation. 
Identify and resolve significant concerns to improve overall code quality while deliberately disregarding minor issues.`
	}

	lang := r.cfg.Language
	if lang == "" {
		lang = defaultLanguage
	}
	systemMsg += fmt.Sprintf("\n\nYour entire response must be in the language with ISO code: %s", lang)

	b.WriteString(systemMsg)
	b.WriteString("\n\n")

	b.WriteString("Please analyze this code diff and provide a comprehensive review in markdown format.\n\n")
	b.WriteString("Focus on security, performance, code quality, and best practices.\n\n")
	b.WriteString("Keep the review scannable and grouped by importance. Lead with critical issues if any exist.\n\n")

	if !r.cfg.DisableReleaseNotes {
		b.WriteString("Your response MUST be a valid JSON object with these fields:\n")
		b.WriteString("- `review`: The complete markdown review content\n")
		b.WriteString("- `fail_pass_workflow`: Either \"pass\", \"fail\", or \"uncertain\"\n")
		b.WriteString("- `labels_added`: Array of label strings (e.g., [\"bug\", \"security\"])\n")
		b.WriteString("- `release_notes`: Brief release notes for this PR (50-100 words)\n\n")
	} else {
		b.WriteString("Your response MUST be a valid JSON object with these fields:\n")
		b.WriteString("- `review`: The complete markdown review content\n")
		b.WriteString("- `fail_pass_workflow`: Either \"pass\", \"fail\", or \"uncertain\"\n")
		b.WriteString("- `labels_added`: Array of label strings (e.g., [\"bug\", \"security\"])\n\n")
	}

	b.WriteString("Respond ONLY with the JSON object, no other text.\n")

	if r.ghClient != nil && r.ghClient.Token != "" {
		r.addGitHubContext(&b)
	}

	b.WriteString("\nCode diff to analyze:\n\n")
	b.WriteString(diffContent)

	return b.String()
}

func (r *Reviewer) buildSummarizePrompt(diffContent string) string {
	var b strings.Builder

	prompt := r.cfg.SummarizePrompt
	if prompt == "" {
		prompt = `Provide your final response in markdown with the following content:
- **Walkthrough**: A high-level summary of the overall change within 80 words.
- **Changes**: A markdown table of files and their summaries. Group files with similar changes together.
- **Poem**: Below the changes, include a whimsical short poem written by a rabbit to celebrate the changes. Format as a quote using ">" and use emojis.

Avoid additional commentary as this summary will be added as a comment on the GitHub pull request. Use titles "Walkthrough" and "Changes" as H2.`
	}

	lang := r.cfg.Language
	if lang == "" {
		lang = defaultLanguage
	}
	prompt += fmt.Sprintf("\n\nYour entire response must be in the language with ISO code: %s", lang)

	b.WriteString(prompt)
	b.WriteString("\n\nCode diff to summarize:\n\n")
	b.WriteString(diffContent)

	return b.String()
}

func (r *Reviewer) buildReviewCommentPrompt(req ReviewCommentRequest) string {
	var b strings.Builder

	systemMsg := r.cfg.SystemMessage
	if systemMsg == "" {
		systemMsg = `You are @goreviewer, a helpful AI assistant for code reviews.`
	}

	lang := r.cfg.Language
	if lang == "" {
		lang = defaultLanguage
	}
	systemMsg += fmt.Sprintf("\n\nYour entire response must be in the language with ISO code: %s", lang)

	b.WriteString(systemMsg)
	b.WriteString("\n\n")

	b.WriteString("A user has left a review comment on a pull request. Please respond to their comment helpfully.\n\n")

	b.WriteString("Comment details:\n")
	fmt.Fprintf(&b, "- File: %s\n", req.Path)
	fmt.Fprintf(&b, "- Line: %d\n", req.Line)
	fmt.Fprintf(&b, "- Diff hunk:\n%s\n\n", req.DiffHunk)
	fmt.Fprintf(&b, "User's comment:\n%s\n\n", req.Comment)

	b.WriteString("Provide a helpful response to their comment. This could be:\n")
	b.WriteString("- Answering a question\n")
	b.WriteString("- Explaining code changes\n")
	b.WriteString("- Acknowledging suggestions\n")
	b.WriteString("- Or responding appropriately\n\n")

	b.WriteString("Be concise, helpful, and conversational. Respond directly to their feedback.")

	return b.String()
}

func (r *Reviewer) buildQuestionPrompt(diffContent, question string) string {
	var b strings.Builder

	systemMsg := r.cfg.SystemMessage
	if systemMsg == "" {
		systemMsg = `You are @goreviewer, a helpful AI assistant for code reviews.`
	}

	lang := r.cfg.Language
	if lang == "" {
		lang = defaultLanguage
	}
	systemMsg += fmt.Sprintf("\n\nYour entire response must be in the language with ISO code: %s", lang)

	b.WriteString(systemMsg)
	b.WriteString("\n\n")

	b.WriteString("A user has asked a question about a code review. Please answer their specific question based on the diff provided.\n\n")

	if r.ghClient != nil && r.ghClient.Token != "" {
		ghCtx, err := r.ghClient.FetchContext(context.Background())
		if err == nil && ghCtx.PreviousReview != "" {
			b.WriteString("Previous AI Review for context:\n")
			b.WriteString(ghCtx.PreviousReview)
			b.WriteString("\n\n")
		}
	}

	b.WriteString("Code diff:\n")
	b.WriteString(diffContent)
	b.WriteString("\n\n")

	b.WriteString("User's question: ")
	b.WriteString(question)
	b.WriteString("\n\n")

	b.WriteString("Answer the question directly and specifically. If the question refers to something from your previous review, reference it. Be helpful and concise.")

	return b.String()
}

func (r *Reviewer) addGitHubContext(b *strings.Builder) {
	ghCtx, err := r.ghClient.FetchContext(context.Background())
	if err != nil {
		return
	}

	if ghCtx.CheckRuns != "" {
		b.WriteString("\nGitHub Actions Check Status:\n")
		b.WriteString(ghCtx.CheckRuns)
		b.WriteString("\n\nPlease consider any failed or pending checks in your review.\n")
	}

	if ghCtx.Labels != "" {
		b.WriteString("\nAvailable Repository Labels:\n")
		b.WriteString(ghCtx.Labels)
		b.WriteString("\n\n")
	}

	if ghCtx.PRDescription != "" {
		b.WriteString("\nPR Context:\n")
		b.WriteString(ghCtx.PRDescription)
		b.WriteString("\n")
	}

	if ghCtx.Commits != "" {
		b.WriteString("\nCommit History:\n")
		b.WriteString(ghCtx.Commits)
		b.WriteString("\n")
	}

	if ghCtx.HumanComments != "" {
		b.WriteString("\nHuman Comments:\n")
		b.WriteString(ghCtx.HumanComments)
		b.WriteString("\n")
	}
}
