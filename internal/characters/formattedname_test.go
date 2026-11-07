package characters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFormattedAdjectives(t *testing.T) {
	// Setup: populate adjectiveSwaps with test data
	adjectiveSwaps = map[string]string{
		"downed":       "♥friend",
		"downed-short": "♥",
		"hidden":       "hidden",
		"hidden-short": "?",
		"zombie":       "zOmBie",
	}

	tests := []struct {
		name         string
		excludeShort bool
		expected     []string
	}{
		{
			name:         "Include short adjectives",
			excludeShort: false,
			expected:     []string{"downed", "downed-short", "hidden", "hidden-short", "zombie"},
		},
		{
			name:         "Exclude short adjectives",
			excludeShort: true,
			expected:     []string{"downed", "hidden", "zombie"},
		},
		{
			name:         "Empty adjectiveSwaps",
			excludeShort: false,
			expected:     []string{},
		},
	}

	for _, tt := range tests[:2] {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFormattedAdjectives(tt.excludeShort)
			assert.ElementsMatch(t, tt.expected, got)
		})
	}

	// Test with empty adjectiveSwaps
	adjectiveSwaps = map[string]string{}
	t.Run(tests[2].name, func(t *testing.T) {
		got := GetFormattedAdjectives(tests[2].excludeShort)
		assert.Empty(t, got)
	})
}
func TestGetFormattedAdjective(t *testing.T) {
	// Setup: populate adjectiveSwaps with test data
	adjectiveSwaps = map[string]string{
		"downed":       "♥friend",
		"downed-short": "♥",
		"hidden":       "hidden",
		"hidden-short": "?",
		"zombie":       "zOmBie",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Existing adjective",
			input:    "downed",
			expected: "♥friend",
		},
		{
			name:     "Existing short adjective",
			input:    "downed-short",
			expected: "♥",
		},
		{
			name:     "Non-existing adjective returns input",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "Another existing adjective",
			input:    "zombie",
			expected: "zOmBie",
		},
		{
			name:     "Another short adjective",
			input:    "hidden-short",
			expected: "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFormattedAdjective(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}

	// Test with empty adjectiveSwaps
	adjectiveSwaps = map[string]string{}
	t.Run("Empty adjectiveSwaps returns input", func(t *testing.T) {
		got := GetFormattedAdjective("downed")
		assert.Equal(t, "downed", got)
	})
}
