package utils

import (
	"regexp"
	"strings"
)

var nonDigitRegex = regexp.MustCompile(`[^\d]`)

// SanitizePhoneNumber cleans a phone string and normalizes it to Indonesian standard (628xxx)
// Returns empty string if invalid or too short (< 8 digits).
func SanitizePhoneNumber(phone string) string {
	cleaned := strings.TrimSpace(phone)
	if cleaned == "" {
		return ""
	}

	// Remove all non-digits
	cleaned = nonDigitRegex.ReplaceAllString(cleaned, "")

	if len(cleaned) < 8 {
		return ""
	}

	// Normalize prefix
	if strings.HasPrefix(cleaned, "08") {
		cleaned = "628" + cleaned[2:]
	} else if strings.HasPrefix(cleaned, "+628") {
		cleaned = "628" + cleaned[4:]
	} else if strings.HasPrefix(cleaned, "8") {
		cleaned = "628" + cleaned[1:]
	}

	return cleaned
}
