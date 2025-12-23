package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles for dialogue elements
var (
	// CharacterNameStyle for character names (**Name**)
	CharacterNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Gold)

	// ActionStyle for character actions (*(doing something)*)
	ActionStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(Gray)

	// DialogueStyle for spoken text ("dialogue")
	DialogueStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(Text)

	// NarrationStyle for narrative text
	NarrationStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(Text)

	// EmphasisStyle for emphasized words within narration (**word**)
	EmphasisStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Text)
)

// Token represents a piece of styled text
type Token struct {
	Text   string
	Bold   bool
	Italic bool
}

// RenderMarkdownV2 parses markdown with standard syntax:
//   - **text** for bold
//   - *text* for italic
//   - Supports nesting: *text with **bold** inside*
func RenderMarkdownV2(text string, baseStyle lipgloss.Style) string {
	tokens := parseMarkdownTokens(text)
	return renderTokens(tokens, baseStyle)
}

// parseMarkdownTokens tokenizes markdown text into styled segments
func parseMarkdownTokens(text string) []Token {
	var tokens []Token
	var currentText strings.Builder
	italic := false
	bold := false

	i := 0
	for i < len(text) {
		// Check for **
		if i+1 < len(text) && text[i:i+2] == "**" {
			// Flush current token
			if currentText.Len() > 0 {
				tokens = append(tokens, Token{
					Text:   currentText.String(),
					Bold:   bold,
					Italic: italic,
				})
				currentText.Reset()
			}

			// Toggle bold
			bold = !bold
			i += 2
			continue
		}

		// Check for single *
		if text[i] == '*' {
			// Flush current token
			if currentText.Len() > 0 {
				tokens = append(tokens, Token{
					Text:   currentText.String(),
					Bold:   bold,
					Italic: italic,
				})
				currentText.Reset()
			}

			// Toggle italic
			italic = !italic
			i++
			continue
		}

		// Regular character
		currentText.WriteByte(text[i])
		i++
	}

	// Flush remaining text
	if currentText.Len() > 0 {
		tokens = append(tokens, Token{
			Text:   currentText.String(),
			Bold:   bold,
			Italic: italic,
		})
	}

	return tokens
}

// renderTokens applies styles to tokens and returns formatted string
func renderTokens(tokens []Token, baseStyle lipgloss.Style) string {
	var result strings.Builder

	for _, token := range tokens {
		if token.Text == "" {
			continue
		}

		style := baseStyle.Copy()

		if token.Bold {
			style = style.Bold(true)
		}
		if token.Italic {
			style = style.Italic(true)
		}

		result.WriteString(style.Render(token.Text))
	}

	return result.String()
}

// RenderDialogue renders a complete dialogue line with proper styling
func RenderDialogue(line string) string {
	return RenderMarkdownV2(line, DMStyle)
}

// RenderNarration renders narrative text with emphasis
func RenderNarration(line string) string {
	return RenderMarkdownV2(line, DMStyle)
}

// RenderDMTextV2 is the v2 renderer for DM text with standard markdown
func RenderDMTextV2(text string) string {
	return RenderMarkdownV2(text, DMStyle)
}
