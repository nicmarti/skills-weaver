package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captureChatClient spins a stub server that records each request body as
// JSON and answers with a valid chat result.
func captureChatClient(t *testing.T) (*OpenRouterClient, *[]string) {
	t.Helper()
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readRequestBody(t, r)
		data, err := json.Marshal(body)
		if err != nil {
			t.Errorf("marshal captured body: %v", err)
		}
		bodies = append(bodies, string(data))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(chatResultFixture()))
	}))
	t.Cleanup(server.Close)

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}
	return client, &bodies
}

// TestReasoningEffort_WireFormat proves the neutral reasoning-effort control
// reaches the provider as reasoning.effort — the campaign-plan generation
// depends on it to keep reasoning models inside the 5-minute budget
// (docs/issues-2026-09-24.md).
func TestReasoningEffort_WireFormat(t *testing.T) {
	client, bodies := captureChatClient(t)

	if _, err := client.Complete(context.Background(), Request{
		Model:           ModelOpus55,
		ReasoningEffort: "medium",
		Messages:        []Message{UserMessage("génère le plan")},
	}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if len(*bodies) == 0 {
		t.Fatal("no request body captured")
	}
	body := (*bodies)[0]
	if !strings.Contains(body, `"reasoning"`) || !strings.Contains(body, `"effort":"medium"`) {
		t.Errorf("reasoning effort not sent, got body: %s", body)
	}
}

// TestReasoningEffort_EmptyLeavesProviderDefault: no control sent when unset.
func TestReasoningEffort_EmptyLeavesProviderDefault(t *testing.T) {
	client, bodies := captureChatClient(t)

	if _, err := client.Complete(context.Background(), Request{
		Model:    ModelSonnet5,
		Messages: []Message{UserMessage("bonjour")},
	}); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if body := (*bodies)[0]; strings.Contains(body, `"reasoning"`) {
		t.Errorf("empty effort must leave the provider default, got: %s", body)
	}
}

// TestReasoningEffort_InvalidRejected: a mistyped effort fails loudly instead
// of silently sending the provider default.
func TestReasoningEffort_InvalidRejected(t *testing.T) {
	_, err := buildChatRequest(Request{
		Model:           ModelOpus55,
		ReasoningEffort: "médium",
	}, false)
	if err == nil || !strings.Contains(err.Error(), "invalid reasoning effort") {
		t.Errorf("mistyped effort must be rejected, got err: %v", err)
	}
}
