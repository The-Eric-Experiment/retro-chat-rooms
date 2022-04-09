package main

import (
	"bufio"
	"os"
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
var blocked []string

func LoadProfanityFilters() {
	var err error
	censored, err = readLines("profanity/censored-profanity.txt")
	if err != nil {
		panic(err)
	}

	blocked, err = readLines("profanity/fully-blocked-profanity.txt")
	if err != nil {
		panic(err)
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

func HasBlockedWords(input string) bool {
	words := strings.Split(input, " ")

	for _, word := range words {
		if word == "" {
			continue
		}
		if lo.Contains(blocked, strings.ToLower(word)) {
			return true
		}
	}

	return false
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
