package helpers

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

func Now() time.Time {
	return time.Now().UTC()
}

func FormatTimestamp(t time.Time) string {
	return t.Format("03:04:05 PM")
}

func FormatTimestamp24H(t time.Time) string {
	return t.Format("15:04:05")
}

func GenerateUniqueID(input string) string {
	// Step 1: Generate a deterministic UUID using MD5 hashing
	uuid := uuid.NewMD5(uuid.NameSpaceOID, []byte(input)).String()

	// Step 2: Remove hyphens from the UUID to work with the raw characters
	uuid = strings.ReplaceAll(uuid, "-", "")

	// Step 3: Split UUID into parts and combine them into a single number
	part1 := uuid[0:8]
	part2 := uuid[8:16]
	part3 := uuid[16:24]
	part4 := uuid[24:32]

	// Convert parts to numbers and combine them using XOR
	var result uint64
	for _, part := range []string{part1, part2, part3, part4} {
		for _, c := range part {
			result = result*31 + uint64(c)
		}
	}

	// Step 4: Encode the result into base62 to get a short ID
	const base62Charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base := uint64(len(base62Charset))
	shortID := ""
	for result > 0 {
		shortID = string(base62Charset[result%base]) + shortID
		result /= base
	}

	// Ensure the result is at least 5 characters long
	for len(shortID) < 5 {
		shortID = string(base62Charset[0]) + shortID
	}

	return shortID
}
