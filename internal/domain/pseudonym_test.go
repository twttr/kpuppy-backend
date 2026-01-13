package domain

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePseudonym_Deterministic(t *testing.T) {
	hash := "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"

	result1 := GeneratePseudonym(hash)
	result2 := GeneratePseudonym(hash)

	assert.Equal(t, result1, result2)
	assert.NotEqual(t, "Anonymous User", result1)
}

func TestGeneratePseudonym_DifferentHashes(t *testing.T) {
	hash1 := "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"
	hash2 := "b8c4d3e2f9e0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b6"

	result1 := GeneratePseudonym(hash1)
	result2 := GeneratePseudonym(hash2)

	assert.NotEqual(t, result1, result2)
}

func TestGeneratePseudonym_InvalidHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"empty hash", ""},
		{"short hash", "abc"},
		{"invalid hex", "gggg"},
		{"too short after decode", "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GeneratePseudonym(tt.hash)
			assert.Equal(t, "Anonymous User", result)
		})
	}
}

func TestGeneratePseudonym_Format(t *testing.T) {
	hash := "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"

	result := GeneratePseudonym(hash)

	pattern := regexp.MustCompile(`^[A-Z][a-z]+ [A-Z][a-z]+ \d{1,2}$`)
	assert.True(t, pattern.MatchString(result), "Expected format 'Adjective Animal N', got: %s", result)
}

func TestGeneratePseudonym_Distribution(t *testing.T) {
	hashes := []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"1f00000000000000000000000000000000000000000000000000000000000000",
		"001f000000000000000000000000000000000000000000000000000000000000",
		"00001f0000000000000000000000000000000000000000000000000000000000",
		"ff00000000000000000000000000000000000000000000000000000000000000",
		"00ff000000000000000000000000000000000000000000000000000000000000",
	}

	results := make(map[string]bool)
	for _, hash := range hashes {
		result := GeneratePseudonym(hash)
		results[result] = true
	}

	assert.Greater(t, len(results), 1, "Expected different pseudonyms for different hashes")
}

func TestGeneratePseudonym_SuffixRange(t *testing.T) {
	hash1 := "0000000000000000000000000000000000000000000000000000000000000000"
	hash2 := "0000630000000000000000000000000000000000000000000000000000000000"

	result1 := GeneratePseudonym(hash1)
	result2 := GeneratePseudonym(hash2)

	pattern := regexp.MustCompile(`\d{1,2}$`)
	match1 := pattern.FindString(result1)
	match2 := pattern.FindString(result2)

	assert.NotEmpty(t, match1)
	assert.NotEmpty(t, match2)
}
