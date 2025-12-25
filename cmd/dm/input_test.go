package main

import (
	"strings"
	"testing"
)

// TestEmptyInputHandling tests that various forms of empty input are properly detected.
func TestEmptyInputHandling(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool // true if should be considered empty
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "Single space",
			input:    " ",
			expected: true,
		},
		{
			name:     "Multiple spaces",
			input:    "     ",
			expected: true,
		},
		{
			name:     "Single tab",
			input:    "\t",
			expected: true,
		},
		{
			name:     "Multiple tabs",
			input:    "\t\t\t",
			expected: true,
		},
		{
			name:     "Mixed spaces and tabs",
			input:    "  \t  \t  ",
			expected: true,
		},
		{
			name:     "Newline",
			input:    "\n",
			expected: true,
		},
		{
			name:     "Carriage return",
			input:    "\r",
			expected: true,
		},
		{
			name:     "Mixed whitespace",
			input:    " \t\n\r ",
			expected: true,
		},
		{
			name:     "Valid input with leading/trailing spaces",
			input:    "  hello  ",
			expected: false,
		},
		{
			name:     "Valid input with leading/trailing tabs",
			input:    "\thello\t",
			expected: false,
		},
		{
			name:     "Single character",
			input:    "a",
			expected: false,
		},
		{
			name:     "Normal command",
			input:    "exit",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trimmed := strings.TrimSpace(tc.input)
			isEmpty := trimmed == ""

			if isEmpty != tc.expected {
				t.Errorf("Input %q: expected isEmpty=%v, got isEmpty=%v (trimmed: %q)",
					tc.input, tc.expected, isEmpty, trimmed)
			}
		})
	}
}
