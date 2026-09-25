package llm

import (
	"context"
	"strings"
	"testing"
)

// TestCacheBreakpoint_WireFormat proves a breakpoint user message reaches the
// provider as an Anthropic-style cache_control directive on the text part —
// any multi-call conversation (e.g. the DM loop) can rely on it to cache the
// shared prefix between requests.
func TestCacheBreakpoint_WireFormat(t *testing.T) {
	client, bodies := captureChatClient(t)

	req := Request{
		Model: ModelOpus55,
		Messages: []Message{
			CachedUserMessage("contexte persistant du monde"),
			UserMessage("question"),
		},
	}
	if _, err := client.Complete(context.Background(), req); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	body := (*bodies)[0]
	if !strings.Contains(body, `"cache_control"`) || !strings.Contains(body, `"ephemeral"`) {
		t.Errorf("cache breakpoint not sent, got body: %s", body)
	}
}

// TestCacheBreakpoint_ImageInsideCachedPrefix: when the breakpoint message
// carries an image, the image part must precede the marked text part so the
// cached prefix covers it.
func TestCacheBreakpoint_ImageInsideCachedPrefix(t *testing.T) {
	client, bodies := captureChatClient(t)

	req := Request{
		Model: ModelOpus55,
		Messages: []Message{
			CachedUserMessage("contexte", ImagePart{MediaType: "image/png", Base64: "QUJD"}),
		},
	}
	if _, err := client.Complete(context.Background(), req); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	body := (*bodies)[0]
	imageIdx := strings.Index(body, `"image_url"`)
	textIdx := strings.Index(body, `"cache_control"`)
	if imageIdx == -1 || textIdx == -1 {
		t.Fatalf("image or cache_control missing, got: %s", body)
	}
	if imageIdx > textIdx {
		t.Errorf("image part must come before the breakpoint text part, got: %s", body)
	}
}

// TestCacheBreakpoint_PlainMessageUntouched: messages without the flag keep
// the plain string wire form.
func TestCacheBreakpoint_PlainMessageUntouched(t *testing.T) {
	client, bodies := captureChatClient(t)

	req := Request{
		Model:    ModelOpus55,
		Messages: []Message{UserMessage("question")},
	}
	if _, err := client.Complete(context.Background(), req); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	body := (*bodies)[0]
	if strings.Contains(body, "cache_control") {
		t.Errorf("plain message must not carry cache_control, got: %s", body)
	}
}
