package utils // dnywonnt.me/alerts2incidents/internal/utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JoinWithCommas takes a slice of strings and joins them into a single string separated by commas.
func JoinWithCommas(items []string) string {
	return strings.Join(items, ", ")
}

// EscapeMarkdownV2 escapes special Markdown V2 characters in the text to prevent formatting issues.
func EscapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		".", "\\.",
		"!", "\\!",
		"|", "\\|",
		"~", "\\~",
		">", "\\>",
		"=", "\\=",
		"\\", "\\\\",
	)
	return replacer.Replace(text)
}

// DerefStr safely dereferences a pointer to a string, returning an empty string if the pointer is nil.
func DerefStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// PrettyJSON takes a JSON string and returns a formatted version with indents for easier readability.
func PrettyJSON(jsonStr string) string {
	jsonObj := interface{}(nil)                      // Create an empty interface to hold the JSON object
	err := json.Unmarshal([]byte(jsonStr), &jsonObj) // Parse JSON string into the interface
	if err != nil {
		return "" // Return an empty string in case of parsing error
	}

	prettyJSON, err := json.MarshalIndent(jsonObj, "", "    ") // Format the JSON object with 4-space indents
	if err != nil {
		return "" // Return an empty string in case of formatting error
	}

	return string(prettyJSON) // Convert the formatted JSON back to string and return
}

// Function to format a number with commas
func FormatNumberWithCommas(num int) string {
	// Convert the number to a string
	numStr := fmt.Sprintf("%d", num)

	// Split the string into parts of three digits from the end
	result := strings.Builder{}
	length := len(numStr)
	for i, digit := range numStr {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteByte(byte(digit))
	}

	return result.String()
}
