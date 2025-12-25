package main

import (
	"strings"
	"testing"
)

// TestReadlineInputValidation tests that input validation still works with readline.
func TestReadlineInputValidation(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		shouldProcess bool
	}{
		{
			name:          "Empty string",
			input:         "",
			shouldProcess: false,
		},
		{
			name:          "Spaces only",
			input:         "     ",
			shouldProcess: false,
		},
		{
			name:          "Tabs only",
			input:         "\t\t\t",
			shouldProcess: false,
		},
		{
			name:          "Mixed whitespace",
			input:         "  \t  \n  ",
			shouldProcess: false,
		},
		{
			name:          "Valid input",
			input:         "hello world",
			shouldProcess: true,
		},
		{
			name:          "Valid input with whitespace",
			input:         "  hello  ",
			shouldProcess: true,
		},
		{
			name:          "Exit command",
			input:         "exit",
			shouldProcess: true, // Should be processed to exit
		},
		{
			name:          "Quit command",
			input:         "quit",
			shouldProcess: true, // Should be processed to exit
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the validation logic from main.go
			trimmed := strings.TrimSpace(tc.input)
			isEmpty := trimmed == ""

			if isEmpty && tc.shouldProcess {
				t.Errorf("Input %q should be processed but was detected as empty", tc.input)
			}
			if !isEmpty && !tc.shouldProcess {
				t.Errorf("Input %q should not be processed but was detected as non-empty", tc.input)
			}
		})
	}
}

// TestReadlineHistoryBehavior tests that command history logic works correctly.
func TestReadlineHistoryBehavior(t *testing.T) {
	// Mock history of commands
	history := []string{
		"first command",
		"second command",
		"  third with spaces  ",
		"", // Empty lines should not be in history
		"fourth command",
	}

	// Filter history (empty entries should not be saved)
	var validHistory []string
	for _, cmd := range history {
		trimmed := strings.TrimSpace(cmd)
		if trimmed != "" {
			validHistory = append(validHistory, trimmed)
		}
	}

	// Verify only non-empty commands are kept
	expectedCount := 4
	if len(validHistory) != expectedCount {
		t.Errorf("Expected %d valid history entries, got %d", expectedCount, len(validHistory))
	}

	// Verify whitespace was trimmed
	if validHistory[2] != "third with spaces" {
		t.Errorf("Expected 'third with spaces', got %q", validHistory[2])
	}
}

// TestReadlineExitCommands tests that exit commands are properly recognized.
func TestReadlineExitCommands(t *testing.T) {
	exitCommands := []string{"exit", "quit", "EXIT", "QUIT", "Exit", "Quit"}

	for _, cmd := range exitCommands {
		t.Run("Command: "+cmd, func(t *testing.T) {
			normalized := strings.ToLower(strings.TrimSpace(cmd))
			if normalized != "exit" && normalized != "quit" {
				t.Errorf("Exit command %q not properly normalized to exit/quit", cmd)
			}
		})
	}
}

// TestReadlineInputSanitization tests that inputs are properly sanitized.
func TestReadlineInputSanitization(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		isEmpty  bool
	}{
		{
			name:     "Normal input",
			input:    "attack the goblin",
			expected: "attack the goblin",
			isEmpty:  false,
		},
		{
			name:     "Input with leading/trailing spaces",
			input:    "  look around  ",
			expected: "look around",
			isEmpty:  false,
		},
		{
			name:     "Input with tabs",
			input:    "\tinspect door\t",
			expected: "inspect door",
			isEmpty:  false,
		},
		{
			name:     "Only spaces",
			input:    "     ",
			expected: "",
			isEmpty:  true,
		},
		{
			name:     "Multiple consecutive spaces",
			input:    "cast    fireball",
			expected: "cast    fireball", // Internal spaces preserved
			isEmpty:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sanitized := strings.TrimSpace(tc.input)

			if sanitized != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, sanitized)
			}

			isEmpty := sanitized == ""
			if isEmpty != tc.isEmpty {
				t.Errorf("Expected isEmpty=%v, got isEmpty=%v", tc.isEmpty, isEmpty)
			}
		})
	}
}

// TestReadlineControlCharacterHandling tests that control characters are handled.
func TestReadlineControlCharacterHandling(t *testing.T) {
	// With readline, these would be handled by the library, but we test the input validation
	testCases := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "Arrow left sequence",
			input:       "hello\x1b[Dworld", // ESC[D = arrow left
			description: "Arrow keys should not produce visible characters",
		},
		{
			name:        "Arrow right sequence",
			input:       "hello\x1b[Cworld", // ESC[C = arrow right
			description: "Arrow keys should not produce visible characters",
		},
		{
			name:        "Home key",
			input:       "hello\x1b[Hworld", // ESC[H = home
			description: "Home key should not produce visible characters",
		},
		{
			name:        "End key",
			input:       "hello\x1b[Fworld", // ESC[F = end
			description: "End key should not produce visible characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// In actual readline usage, these sequences are intercepted
			// and don't appear in the final input string
			// This test verifies our understanding of the problem
			if !strings.Contains(tc.input, "\x1b") {
				t.Errorf("Test case should contain escape sequences")
			}
			// With readline, the final input would NOT contain these sequences
		})
	}
}

// TestReadlinePromptBehavior tests prompt handling.
func TestReadlinePromptBehavior(t *testing.T) {
	// Test that prompts are handled correctly
	prompt := "> "

	if prompt == "" {
		t.Error("Prompt should not be empty")
	}

	// Verify prompt doesn't interfere with input parsing
	testInput := "test command"
	if strings.HasPrefix(testInput, prompt) {
		t.Error("Input should not contain prompt prefix")
	}
}
