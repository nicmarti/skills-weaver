package agent

import (
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

// TestSerializeMessageWithImagePreservesText guards the fix for the noisy
// "Unknown content block type: image" warning. A message pairing text with an
// injected image (e.g. the world map given to world-keeper) must serialize
// cleanly: the text is preserved and the message still round-trips.
func TestSerializeMessageWithImagePreservesText(t *testing.T) {
	msg := anthropic.NewUserMessage(
		anthropic.NewTextBlock("Voici la carte du monde."),
		anthropic.NewImageBlockBase64("image/png", "QUJD"), // base64("ABC")
	)

	s, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage error: %v", err)
	}
	if s.TextContent != "Voici la carte du monde." {
		t.Errorf("text not preserved, got %q", s.TextContent)
	}
	// Must round-trip back to a valid (non-empty) message.
	if _, err := DeserializeMessage(s); err != nil {
		t.Errorf("DeserializeMessage error: %v", err)
	}
}

// TestSerializeImageOnlyMessageGetsPlaceholder ensures an image-only message
// (no accompanying text) does not serialize to an empty content array —
// DeserializeMessage rejects those as orphaned. A placeholder keeps it valid.
func TestSerializeImageOnlyMessageGetsPlaceholder(t *testing.T) {
	msg := anthropic.NewUserMessage(
		anthropic.NewImageBlockBase64("image/png", "QUJD"),
	)

	s, err := SerializeMessage(msg)
	if err != nil {
		t.Fatalf("SerializeMessage error: %v", err)
	}
	if s.TextContent == "" {
		t.Error("image-only message should receive a non-empty placeholder")
	}
	if _, err := DeserializeMessage(s); err != nil {
		t.Errorf("DeserializeMessage on image-only message error: %v", err)
	}
}
