package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestParseMarkdownTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "simple bold",
			input: "Hello **world**!",
			expected: []Token{
				{Text: "Hello ", Bold: false, Italic: false},
				{Text: "world", Bold: true, Italic: false},
				{Text: "!", Bold: false, Italic: false},
			},
		},
		{
			name:  "simple italic",
			input: "Hello *world*!",
			expected: []Token{
				{Text: "Hello ", Bold: false, Italic: false},
				{Text: "world", Bold: false, Italic: true},
				{Text: "!", Bold: false, Italic: false},
			},
		},
		{
			name:  "bold inside italic",
			input: "*text with **bold** inside*",
			expected: []Token{
				{Text: "text with ", Bold: false, Italic: true},
				{Text: "bold", Bold: true, Italic: true},
				{Text: " inside", Bold: false, Italic: true},
			},
		},
		{
			name:  "character name and action",
			input: "**Gareth** *(looking around)*",
			expected: []Token{
				{Text: "Gareth", Bold: true, Italic: false},
				{Text: " ", Bold: false, Italic: false},
				{Text: "(looking around)", Bold: false, Italic: true},
			},
		},
		{
			name:  "narration with emphasis",
			input: "*You enter the **dark crypt**.*",
			expected: []Token{
				{Text: "You enter the ", Bold: false, Italic: true},
				{Text: "dark crypt", Bold: true, Italic: true},
				{Text: ".", Bold: false, Italic: true},
			},
		},
		{
			name:  "multiple bold words",
			input: "The **red dragon** and **black orc**",
			expected: []Token{
				{Text: "The ", Bold: false, Italic: false},
				{Text: "red dragon", Bold: true, Italic: false},
				{Text: " and ", Bold: false, Italic: false},
				{Text: "black orc", Bold: true, Italic: false},
			},
		},
		{
			name:  "no markdown",
			input: "Plain text without formatting",
			expected: []Token{
				{Text: "Plain text without formatting", Bold: false, Italic: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := parseMarkdownTokens(tt.input)

			if len(tokens) != len(tt.expected) {
				t.Errorf("parseMarkdownTokens() got %d tokens, expected %d", len(tokens), len(tt.expected))
				t.Logf("Got tokens: %+v", tokens)
				t.Logf("Expected: %+v", tt.expected)
				return
			}

			for i, token := range tokens {
				expected := tt.expected[i]
				if token.Text != expected.Text {
					t.Errorf("Token %d: text = %q, want %q", i, token.Text, expected.Text)
				}
				if token.Bold != expected.Bold {
					t.Errorf("Token %d: bold = %v, want %v", i, token.Bold, expected.Bold)
				}
				if token.Italic != expected.Italic {
					t.Errorf("Token %d: italic = %v, want %v", i, token.Italic, expected.Italic)
				}
			}
		})
	}
}

func TestRenderMarkdownV2(t *testing.T) {
	baseStyle := lipgloss.NewStyle()

	tests := []struct {
		name          string
		input         string
		checkContains []string
		checkMissing  []string
	}{
		{
			name:          "bold text",
			input:         "Hello **world**!",
			checkContains: []string{"Hello", "world", "!"},
			checkMissing:  []string{"**"},
		},
		{
			name:          "italic text",
			input:         "Hello *world*!",
			checkContains: []string{"Hello", "world", "!"},
			checkMissing:  []string{"*"},
		},
		{
			name:          "nested bold in italic",
			input:         "*text with **bold** inside*",
			checkContains: []string{"text with", "bold", "inside"},
			checkMissing:  []string{"**", "*"},
		},
		{
			name:          "character dialogue",
			input:         "**Gareth** *(regardant Elara)* :",
			checkContains: []string{"Gareth", "regardant Elara", ":"},
			checkMissing:  []string{"**", "*"},
		},
		{
			name:          "narration with emphasis",
			input:         "*Vous vous retrouvez dans la **nuit noire**.*",
			checkContains: []string{"Vous vous retrouvez dans la", "nuit noire", "."},
			checkMissing:  []string{"**", "*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderMarkdownV2(tt.input, baseStyle)

			// Check that markdown symbols are removed
			for _, missing := range tt.checkMissing {
				if strings.Contains(result, missing) {
					t.Errorf("RenderMarkdownV2() still contains %q: %v", missing, result)
				}
			}

			// Check that content is present
			for _, content := range tt.checkContains {
				if !strings.Contains(result, content) {
					t.Errorf("RenderMarkdownV2() missing %q in result: %v", content, result)
				}
			}
		})
	}
}

func TestRenderDMTextV2(t *testing.T) {
	input := "*Vous entrez dans la **crypte sombre**.*"
	result := RenderDMTextV2(input)

	// Should not contain markdown symbols
	if strings.Contains(result, "**") || strings.Contains(result, "*") {
		t.Errorf("RenderDMTextV2() still contains markdown symbols")
	}

	// Should contain the content
	if !strings.Contains(result, "crypte sombre") {
		t.Errorf("RenderDMTextV2() missing content")
	}
}

func TestComplexDialogue(t *testing.T) {
	// Test the actual example from the user
	dialogue := `**Gareth** *(regardant Elara)* :
— *"On doit partir. Maintenant."*`

	result := RenderDMTextV2(dialogue)

	// Check all content is present
	expected := []string{"Gareth", "regardant Elara", ":", "On doit partir", "Maintenant"}
	for _, content := range expected {
		if !strings.Contains(result, content) {
			t.Errorf("Complex dialogue missing %q", content)
		}
	}

	// Check markdown symbols removed
	if strings.Contains(result, "**") {
		t.Errorf("Complex dialogue still contains **")
	}
}
