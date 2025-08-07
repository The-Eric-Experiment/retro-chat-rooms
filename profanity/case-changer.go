package profanity

import (
	"regexp"
	"strings"
)

// splitSingleCamelCase splits a single camelCase word into separate words
func splitSingleCamelCase(word string) string {
	// First, handle the tricky case: sequences of single uppercase letters followed by a word
	// Like "ILike" should become "I Like", not "IL ike"
	// This regex finds: (uppercase letter)(uppercase letter followed by lowercase)
	// But we want to split BEFORE the last uppercase letter in a sequence
	result := regexp.MustCompile(`([A-Z])([A-Z][a-z])`).ReplaceAllString(word, `$1 $2`)

	// Then handle normal lowercase to uppercase transitions
	result = regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(result, `$1 $2`)

	// Convert to lowercase
	return strings.ToLower(result)
}

// SplitCamelCase breaks camelCase words into separate words
func splitCamelCase(message string) string {
	// Regex to find camelCase/PascalCase words
	// Matches words that have at least one lowercase-to-uppercase or uppercase-to-uppercase transition
	camelCaseRegex := regexp.MustCompile(`\b[a-zA-Z]*[a-z][A-Z][a-zA-Z]*\b|\b[A-Z][A-Z][a-zA-Z]*\b`)

	// Replace each camelCase word with its split version
	result := camelCaseRegex.ReplaceAllStringFunc(message, func(match string) string {
		return splitSingleCamelCase(match)
	})

	return result
}
