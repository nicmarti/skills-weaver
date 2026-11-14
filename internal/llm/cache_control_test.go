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

// TestCacheLastMessage_MarksFinalToolMessage: the conversational-caching
// pattern used by the agent tool loops — Request.CacheLastMessage with a
// tool-role final message must put the breakpoint on the last tool result,
// and must not mutate the caller's message slice (conversation contexts
// reuse it across a whole session).
func TestCacheLastMessage_MarksFinalToolMessage(t *testing.T) {
	client, bodies := captureChatClient(t)

	messages := []Message{
		AssistantToolCallMessage("", []ToolCall{{ID: "call_1", Name: "roll_dice", Arguments: []byte(`{"expression":"1d20"}`)}}),
		ToolResultMessage(ToolResult{ToolCallID: "call_1", Content: `{"total":18}`}),
	}
	req := Request{Model: ModelSonnet5, Messages: messages, CacheLastMessage: true}
	if _, err := client.Complete(context.Background(), req); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	body := (*bodies)[0]
	if !strings.Contains(body, `"cache_control"`) || !strings.Contains(body, `"ephemeral"`) {
		t.Errorf("cache breakpoint missing on final tool result, got: %s", body)
	}
	// The captured body is JSON-in-JSON: the tool content's quotes are
	// escaped, so check the escaped form of the payload.
	if !strings.Contains(body, `\"total\":18`) {
		t.Errorf("tool result content mangled by breakpoint form, got: %s", body)
	}
	if messages[1].CacheBreakpoint {
		t.Error("caller's message slice was mutated: CacheBreakpoint leaked into conversation history")
	}
}

// TestCacheLastMessage_MarksFinalUserMessage: with a user-role final message
// the first call of a turn must carry the breakpoint too (iteration 0 writes
// the cache that iteration 1 reads).
func TestCacheLastMessage_MarksFinalUserMessage(t *testing.T) {
	client, bodies := captureChatClient(t)

	messages := []Message{UserMessage("question")}
	if _, err := client.Complete(context.Background(), Request{
		Model: ModelSonnet5, Messages: messages, CacheLastMessage: true,
	}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	body := (*bodies)[0]
	if !strings.Contains(body, `"cache_control"`) {
		t.Errorf("cache breakpoint missing on final user message, got: %s", body)
	}
	if messages[0].CacheBreakpoint {
		t.Error("caller's message slice was mutated")
	}
}

// TestCacheLastMessage_AssistantTailUnmarked: a trailing assistant message
// completes the turn — no follow-up call could read that cache — so it must
// not pay the cache-write surcharge.
func TestCacheLastMessage_AssistantTailUnmarked(t *testing.T) {
	client, bodies := captureChatClient(t)

	if _, err := client.Complete(context.Background(), Request{
		Model:            ModelSonnet5,
		Messages:         []Message{UserMessage("question"), AssistantMessage("réponse")},
		CacheLastMessage: true,
	}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	body := (*bodies)[0]
	if strings.Contains(body, "cache_control") {
		t.Errorf("assistant tail must not carry cache_control, got: %s", body)
	}
}
