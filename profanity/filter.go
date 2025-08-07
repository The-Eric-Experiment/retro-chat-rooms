package profanity

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/samber/lo"
)

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

var censored []string
var blockedRegexes []*regexp.Regexp

func LoadProfanityFilters() {
	var err error
	censored, err = readLines("profanity/censored-profanity.txt")
	if err != nil {
		panic(err)
	}

	blocked, err := readLines("profanity/fully-blocked-profanity.txt")
	if err != nil {
		panic(err)
	}

	for _, word := range blocked {
		// Using word boundaries for better matching
		pattern := `(?i)\b` + regexp.QuoteMeta(strings.TrimSpace(word)) + `\b`
		blockedRegexes = append(blockedRegexes, regexp.MustCompile(pattern))
	}
}

func ReplaceSensoredProfanity(input string) string {
	words := strings.Split(input, " ")
	words = lo.Map(words, func(word string, _ int) string {
		if word == "" {
			return word
		}
		exists := lo.Contains(censored, strings.ToLower(word))
		if !exists {
			return word
		}

		chars := strings.Split(word, "")
		chars = lo.Map(chars, func(c string, i int) string {
			if i == 0 || i >= (len(chars)-1) {
				return c
			}

			return "*"
		})

		return strings.Join(chars, "")
	})

	return strings.Join(words, " ")
}

// ReplaceUnderscores replaces all underscores with spaces
func replaceUnderscores(text string) string {
	return strings.ReplaceAll(text, "_", " ")
}

func hasBlockedWords(input string, blockedRegexes []*regexp.Regexp) bool {
	camelCase := normalizeTextPreserveCase(input)
	camelCase = replaceUnderscores(camelCase)
	camelCase = splitCamelCase(camelCase)
	input = normalizeText(input)
	input = replaceUnderscores(input)
	for _, regex := range blockedRegexes {
		// we are checking camel case separately in case the user was
		// typing the word with different cases and the slur would be
		// split into two words
		if regex.MatchString(input) || regex.MatchString(camelCase) {
			return true
		}
	}
	return false
}

func HasBlockedWords(input string) bool {
	return hasBlockedWords(input, blockedRegexes)
}

func IsProfaneNickname(input string) bool {
	lowered := strings.ToLower(input)

	if HasBlockedWords(lowered) {
		return true
	}

	words := strings.Split(input, " ")

	for _, word := range words {
		if word == "" {
			continue
		}
		if lo.Contains(censored, strings.ToLower(word)) {
			return true
		}
	}

	return false
}
