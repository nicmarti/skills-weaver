package ui

import (
	"strings"
	"testing"
)

func TestStreamingRenderer_CompleteLine(t *testing.T) {
	r := NewStreamingMarkdownRenderer()

	// Simulate streaming a complete line with bold text
	chunks := []string{"**La Taverne", " du Voile", " Écarlate**", " apparaît", ".\n"}

	var output strings.Builder
	for _, chunk := range chunks {
		output.WriteString(r.AddChunk(chunk))
	}

	result := output.String()
	if result == "" {
		t.Fatal("Expected rendered output, got empty string")
	}

	// The output should contain the text (we can't easily test ANSI codes)
	if !strings.Contains(result, "La Taverne du Voile Écarlate") {
		t.Errorf("Expected location name in output, got: %s", result)
	}
}

func TestStreamingRenderer_BoldAcrossChunks(t *testing.T) {
	r := NewStreamingMarkdownRenderer()

	// Simulate bold text arriving in chunks
	chunks := []string{"**", "Bold", " Text", "**", "\n"}

	var output strings.Builder
	for _, chunk := range chunks {
		output.WriteString(r.AddChunk(chunk))
	}

	result := output.String()
	if !strings.Contains(result, "Bold Text") {
		t.Errorf("Expected 'Bold Text' in output, got: %s", result)
	}
}

func TestStreamingRenderer_ItalicAcrossChunks(t *testing.T) {
	r := NewStreamingMarkdownRenderer()

	chunks := []string{"*Italic", " text*", "\n"}

	var output strings.Builder
	for _, chunk := range chunks {
		output.WriteString(r.AddChunk(chunk))
	}

	result := output.String()
	if !strings.Contains(result, "Italic text") {
		t.Errorf("Expected 'Italic text' in output, got: %s", result)
	}
}

func TestStreamingRenderer_Flush(t *testing.T) {
	r := NewStreamingMarkdownRenderer()

	// Add text without newline
	r.AddChunk("**Bold text**")

	// Flush should render remaining content
	result := r.Flush()
	if !strings.Contains(result, "Bold text") {
		t.Errorf("Expected 'Bold text' after flush, got: %s", result)
	}
}

func TestStreamingRenderer_Reset(t *testing.T) {
	r := NewStreamingMarkdownRenderer()

	// Set some state
	r.AddChunk("**Bold")
	r.Reset()

	// After reset, bold state should be cleared
	if r.bold {
		t.Error("Expected bold to be false after reset")
	}
	if r.buffer.Len() > 0 {
		t.Error("Expected buffer to be empty after reset")
	}
}

func TestStreamingRenderer_MixedFormatting(t *testing.T) {
	r := NewStreamingMarkdownRenderer()

	chunks := []string{"Normal ", "**bold** ", "and ", "*italic*", "\n"}

	var output strings.Builder
	for _, chunk := range chunks {
		output.WriteString(r.AddChunk(chunk))
	}

	result := output.String()
	if !strings.Contains(result, "Normal") || !strings.Contains(result, "bold") ||
		!strings.Contains(result, "italic") {
		t.Errorf("Expected mixed formatting in output, got: %s", result)
	}
}

func TestStreamingRenderer_IncompleteMarkdown(t *testing.T) {
	r := NewStreamingMarkdownRenderer()

	// Chunk ending with * should be buffered
	output1 := r.AddChunk("Some text *")
	if output1 != "" && strings.Contains(output1, "*") {
		t.Error("Expected incomplete markdown to be buffered")
	}

	// Complete it
	output2 := r.AddChunk("italic*\n")
	combined := output1 + output2
	if !strings.Contains(combined, "italic") {
		t.Errorf("Expected 'italic' after completing markdown, got: %s", combined)
	}
}
