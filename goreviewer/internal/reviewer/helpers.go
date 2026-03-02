package reviewer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	safeIconRegex = regexp.MustCompile(`^[\p{L}\p{N}\p{Po}\p{S}\s]+$`)
	avatarRegex   = regexp.MustCompile(`^<img src="[^"]+" alt="[^"]+" width="\d+" height="\d+" ?/?>$`)
)

func isValidBotIcon(icon string) bool {
	if icon == "" || len(icon) > 100 {
		return false
	}
	if avatarRegex.MatchString(icon) {
		return true
	}
	return safeIconRegex.MatchString(icon)
}

func removeThinking(content string) string {
	re := regexp.MustCompile(`(?s)<thinking>.*?</thinking>\s*`)
	return re.ReplaceAllString(content, "")
}

func extractJSON(content string) string {
	re := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
	matches := re.FindAllString(content, -1)

	var jsonStr string
	for _, m := range matches {
		if isValidJSON(m) {
			jsonStr = m
			break
		}
	}

	if jsonStr == "" && isValidJSON(content) {
		jsonStr = content
	}

	return jsonStr
}

func isValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func errorResponse(msg string) ReviewResponse {
	footer := buildFooter("")
	return ReviewResponse{
		Review:           reviewHeader + "\n\n❌ **Error**: " + msg + footer,
		FailPassWorkflow: "uncertain",
		LabelsAdded:      []string{},
	}
}

func buildFooter(botIcon string) string {
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	footer := fmt.Sprintf("\n\n---\n%s\n*Last updated: %s*\n", reviewFooter, timestamp)
	if isValidBotIcon(botIcon) {
		footer = fmt.Sprintf("\n\n---\n%s %s\n*Last updated: %s*\n", botIcon, reviewFooter, timestamp)
	}
	return footer
}

// computeReviewHash computes a hash of the review content (excluding timestamps and footers)
// to detect if the actual review content has changed.
func computeReviewHash(review string) string {
	// Remove the footer and timestamp to compare only the actual review content
	content := review

	// Remove everything after "---" (footer section)
	if idx := strings.Index(content, "\n---\n"); idx != -1 {
		content = content[:idx]
	}

	// Remove "Last updated:" lines
	re := regexp.MustCompile(`(?m)^\*Last updated:.*\*\s*$`)
	content = re.ReplaceAllString(content, "")

	// Normalize whitespace
	content = strings.TrimSpace(content)

	// Compute SHA256 hash
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// extractReviewHash extracts the review hash from a comment body if it exists.
func extractReviewHash(body string) string {
	re := regexp.MustCompile(`<!-- review-hash: ([a-f0-9]{64}) -->`)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// addReviewHash adds a hidden hash comment to the review body for future comparison.
func addReviewHash(review string) string {
	hash := computeReviewHash(review)
	return review + fmt.Sprintf("\n<!-- review-hash: %s -->", hash)
}
