package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LocationStyle for location names (bold + cyan color)
var LocationStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(Cyan)

// StreamingMarkdownRenderer maintains state across streaming text chunks
// to properly handle markdown formatting that spans multiple chunks.
type StreamingMarkdownRenderer struct {
	buffer       strings.Builder
	bold         bool
	italic       bool
	pendingStyle lipgloss.Style
}

// NewStreamingMarkdownRenderer creates a new streaming renderer
func NewStreamingMarkdownRenderer() *StreamingMarkdownRenderer {
	return &StreamingMarkdownRenderer{
		pendingStyle: DMStyle,
	}
}

// AddChunk processes a chunk of text and returns any complete content to render.
// It buffers incomplete markdown tokens until they're complete.
func (r *StreamingMarkdownRenderer) AddChunk(chunk string) string {
	var output strings.Builder

	for _, char := range chunk {
		r.buffer.WriteRune(char)

		// Check if we have a complete line (newline) or enough buffer to render
		if char == '\n' {
			line := r.buffer.String()
			r.buffer.Reset()
			output.WriteString(r.renderLine(line))
		}
	}

	// For chunks without newlines, check if we can safely render partial content
	// We'll render if buffer doesn't end with an incomplete markdown token
	bufStr := r.buffer.String()
	if bufStr != "" && !strings.HasSuffix(bufStr, "*") && !strings.HasSuffix(bufStr, "**") {
		// Check if we have enough characters to safely render (avoid breaking mid-word)
		if len(bufStr) > 10 || strings.HasSuffix(bufStr, " ") {
			safeToRender := bufStr
			r.buffer.Reset()
			output.WriteString(r.renderWithState(safeToRender))
		}
	}

	return output.String()
}

// Flush renders any remaining buffered content
func (r *StreamingMarkdownRenderer) Flush() string {
	if r.buffer.Len() == 0 {
		return ""
	}

	line := r.buffer.String()
	r.buffer.Reset()
	return r.renderLine(line)
}

// renderLine renders a complete line with markdown parsing
func (r *StreamingMarkdownRenderer) renderLine(line string) string {
	return r.renderWithState(line)
}

// renderWithState maintains bold/italic state across chunks
func (r *StreamingMarkdownRenderer) renderWithState(text string) string {
	var output strings.Builder

	i := 0
	for i < len(text) {
		// Check for **
		if i+1 < len(text) && text[i:i+2] == "**" {
			r.bold = !r.bold
			i += 2
			continue
		}

		// Check for single *
		if text[i] == '*' {
			r.italic = !r.italic
			i++
			continue
		}

		// Collect regular text until next markdown token
		start := i
		for i < len(text) && text[i] != '*' {
			i++
		}

		if i > start {
			content := text[start:i]
			style := r.getStyle()
			output.WriteString(style.Render(content))
		}
	}

	return output.String()
}

// getStyle returns the appropriate style based on current state
func (r *StreamingMarkdownRenderer) getStyle() lipgloss.Style {
	// Bold text is treated as locations (cyan)
	if r.bold && !r.italic {
		return LocationStyle
	}

	// Bold + italic
	if r.bold && r.italic {
		return DMStyle.Copy().Bold(true).Italic(true)
	}

	// Italic only
	if r.italic {
		return DMStyle.Copy().Italic(true)
	}

	// Normal text
	return DMStyle
}

// Reset clears all state (useful between agent responses)
func (r *StreamingMarkdownRenderer) Reset() {
	r.buffer.Reset()
	r.bold = false
	r.italic = false
}
