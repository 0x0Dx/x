package reviewer

import (
	"encoding/json"
	"fmt"
	"regexp"
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
	footer := fmt.Sprintf("\n\n---\n%s\n", reviewFooter)
	if isValidBotIcon(botIcon) {
		footer = fmt.Sprintf("\n\n---\n%s %s\n", botIcon, reviewFooter)
	}
	return footer
}
