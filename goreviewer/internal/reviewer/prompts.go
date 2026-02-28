// Package reviewer provides AI code review functionality.
package reviewer

// PromptSummarizeFileDiff is the prompt for summarizing a file diff.
const PromptSummarizeFileDiff = `## GitHub PR Title

$title 

## Description

$description

## Diff

$file_diff

## Instructions

I would like you to succinctly summarize the diff within 100 words.
If applicable, your summary should include a note about alterations 
to the signatures of exported functions, global data structures and 
variables, and any changes that might affect the external interface or 
behavior of the code.
`

// PromptTriageFileDiff is the prompt for triaging a file diff.
const PromptTriageFileDiff = `Below the summary, I would also like you to triage the diff as NEEDS_REVIEW or APPROVED based on the following criteria:

- If the diff involves any modifications to the logic or functionality, even if they 
  seem minor, triage it as NEEDS_REVIEW. This includes changes to control structures, 
  function calls, or variable assignments that might impact the behavior of the code.
- If the diff only contains very minor changes that don't affect the code logic, such as 
  fixing typos, formatting, or renaming variables for clarity, triage it as APPROVED.

Please evaluate the diff thoroughly and take into account factors such as the number of 
lines changed, the potential impact on the overall system, and the likelihood of 
introducing new bugs or security vulnerabilities. 
When in doubt, always err on the side of caution and triage the diff as NEEDS_REVIEW.

You must strictly follow the format below for triaging the diff:
[TRIAGE]: <NEEDS_REVIEW or APPROVED>

Important:
- In your summary do not mention that the file needs a through review or caution about potential issues.
- Do not provide any reasoning why you triaged the diff as NEEDS_REVIEW or APPROVED.
- Do not mention that these changes affect the logic or functionality of the code in the summary. 
  You must only use the triage status format above to indicate that.
`

// PromptSummarizeChangesets is the prompt for summarizing changesets.
const PromptSummarizeChangesets = `Provided below are changesets in this pull request. Changesets 
are in chronological order and new changesets are appended to the
end of the list. The format consists of filename(s) and the summary 
of changes for those files. There is a separator between each changeset.
Your task is to deduplicate and group together files with
related/similar changes into a single changeset. Respond with the updated 
changesets using the same format as the input. 

$raw_summary
`

// PromptSummarizeShort is the prompt for a short summary.
const PromptSummarizeShort = `Your task is to provide a concise summary of the changes. This 
summary will be used as a prompt while reviewing each file and must be very clear for 
the AI bot to understand. 

Instructions:

- Focus on summarizing only the changes in the PR and stick to the facts.
- Do not provide any instructions to the bot on how to perform the review.
- Do not mention that files need a through review or caution about potential issues.
- Do not mention that these changes affect the logic or functionality of the code.
- The summary should not exceed 500 words.
`
