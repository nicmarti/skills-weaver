package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"
)

// These are paid, opt-in contract checks against OpenRouter. Ordinary go test
// runs never contact the provider, even when a key is present in the shell.
func realOpenRouterClient(t *testing.T) *OpenRouterClient {
	t.Helper()
	if os.Getenv("RUN_REAL_API_TESTS") != "1" {
		t.Skip("set RUN_REAL_API_TESTS=1 to enable paid OpenRouter contract tests")
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewOpenRouterClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func realRequestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func logRealUsage(t *testing.T, resp Response) {
	t.Helper()
	t.Logf("model=%s generation=%s finish=%s tokens=%d cost_usd=%.6f server_tools=%d/%d",
		resp.Model, resp.ID, resp.FinishReason, resp.Usage.TotalTokens, resp.Usage.Cost,
		resp.Usage.ServerToolCallsExecuted, resp.Usage.ServerToolCallsRequested)
}

func TestRealOpenRouterStreamText(t *testing.T) {
	client := realOpenRouterClient(t)
	observer := &recordingObserver{}
	resp, err := client.Stream(realRequestContext(t), Request{
		Model: ModelSonnet5, MaxCompletionTokens: 256,
		Messages: []Message{UserMessage("Reply with one short greeting in French.")},
	}, observer)
	if err != nil {
		t.Fatal(err)
	}
	logRealUsage(t, resp)
	if len(observer.deltas) == 0 || resp.Message.Text == "" || resp.FinishReason != FinishStop {
		t.Fatalf("expected streamed text and a normal finish; deltas=%d finish=%s", len(observer.deltas), resp.FinishReason)
	}
	if !resp.Usage.Present || resp.Usage.TotalTokens <= 0 || resp.ID == "" || resp.Model == "" {
		t.Fatalf("missing generation metadata or usage: %+v", resp)
	}
}

func TestRealOpenRouterStreamParallelTools(t *testing.T) {
	client := realOpenRouterClient(t)
	tools := []ToolDefinition{
		{Name: "record_sun", Description: "Record the sun's code", Parameters: map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{"code": map[string]interface{}{"type": "string"}}, "required": []string{"code"},
		}},
		{Name: "record_moon", Description: "Record the moon's code", Parameters: map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{"code": map[string]interface{}{"type": "string"}}, "required": []string{"code"},
		}},
	}
	request := Request{
		Model: ModelSonnet5, MaxCompletionTokens: 512, Tools: tools,
		Messages: []Message{UserMessage("Call BOTH record_sun with code SUN and record_moon with code MOON now. Make both function calls in this response, not a written answer.")},
	}
	resp, err := client.Stream(realRequestContext(t), request, nil)
	if err != nil {
		t.Fatal(err)
	}
	logRealUsage(t, resp)
	if resp.FinishReason != FinishToolCalls || len(resp.Message.ToolCalls) != 2 {
		t.Fatalf("expected two parallel client tool calls; finish=%s calls=%d", resp.FinishReason, len(resp.Message.ToolCalls))
	}
	seen := make(map[string]bool)
	for _, call := range resp.Message.ToolCalls {
		if call.ID == "" || seen[call.Name] {
			t.Fatalf("missing ID or duplicate tool: %+v", call)
		}
		var args struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			t.Fatalf("incomplete tool arguments for %s: %v", call.Name, err)
		}
		want := map[string]string{"record_sun": "SUN", "record_moon": "MOON"}[call.Name]
		if want == "" || args.Code != want {
			t.Fatalf("unexpected tool arguments: name=%s code=%q", call.Name, args.Code)
		}
		seen[call.Name] = true
	}
	request.Messages = append(request.Messages, resp.Message)
	for _, call := range resp.Message.ToolCalls {
		request.Messages = append(request.Messages, ToolResultMessage(ToolResult{ToolCallID: call.ID, Content: `{"ok":true}`}))
	}
	request.Tools = nil
	followup, err := client.Complete(realRequestContext(t), request)
	if err != nil {
		t.Fatalf("tool-result follow-up: %v", err)
	}
	logRealUsage(t, followup)
	if followup.Message.Text == "" || followup.FinishReason != FinishStop {
		t.Fatalf("tool-result follow-up did not complete: finish=%s", followup.FinishReason)
	}
}

func TestRealOpenRouterImage(t *testing.T) {
	client := realOpenRouterClient(t)
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	resp, err := client.Complete(realRequestContext(t), Request{
		Model: ModelSonnet5, MaxCompletionTokens: 256,
		Messages: []Message{UserMessageWithImage("What solid color is this image? Answer briefly.",
			ImagePart{MediaType: "image/png", Base64: base64.StdEncoding.EncodeToString(encoded.Bytes())})},
	})
	if err != nil {
		t.Fatal(err)
	}
	logRealUsage(t, resp)
	if resp.FinishReason != FinishStop || resp.Message.Text == "" || !resp.Usage.Present {
		t.Fatalf("image input did not yield a complete response: finish=%s", resp.FinishReason)
	}
}

func TestRealOpenRouterInvalidModelError(t *testing.T) {
	client := realOpenRouterClient(t)
	_, err := client.Complete(realRequestContext(t), Request{
		Model: "openrouter/invalid-contract-test-model", MaxCompletionTokens: 16,
		Messages: []Message{UserMessage("hello")},
	})
	var llmErr *Error
	if !errors.As(err, &llmErr) || llmErr.Kind != ErrKindInvalidRequest {
		t.Fatalf("expected classified invalid-model error, got %v", err)
	}
	t.Logf("invalid model classified as kind=%s status=%d", llmErr.Kind, llmErr.Status)
}

// TestRealOpenRouterPromptCachingBeforeAfter validates the conversational
// prompt-caching fix end to end (see internal/agent loops): replaying the
// same growing conversation prefix WITHOUT a cache breakpoint is served at
// full input price (cached=0 — the pre-fix behavior), while with
// Request.CacheLastMessage the follow-up call reads the prefix from cache
// (cached>0) at a fraction of the input cost.
func TestRealOpenRouterPromptCachingBeforeAfter(t *testing.T) {
	client := realOpenRouterClient(t)
	ctx := realRequestContext(t)

	// Sonnet-class models cache only prompts over ~1024 tokens; build a
	// deterministic context well above that threshold.
	var b strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&b, "Chapitre %d : la caravane de Cordova franchit le col de Val-Brume ; les reliques murmurantes sifflent dans leurs coffrets d'anches, le Guet consigne chaque lanterne du port, et la Légate Iolanthe attend le traité de Lumenciel dans son hôtel pavoisé.\n", i)
	}
	bigContext := b.String()
	system := "Tu es un archiviste de Cordova. Réponds en une seule phrase, très courte."
	question1 := UserMessage("Résume le contexte ci-dessus en une phrase.")
	question2 := UserMessage("Propose maintenant un titre court pour ce récit.")

	cacheAware := func(messages []Message) Request {
		return Request{
			Model: ModelSonnet5, System: system, MaxCompletionTokens: 64,
			Messages: messages, CacheLastMessage: true,
		}
	}
	cacheBlind := func(messages []Message) Request {
		return Request{
			Model: ModelSonnet5, System: system, MaxCompletionTokens: 64,
			Messages: messages,
		}
	}
	logPhase := func(label string, resp Response) {
		perK := 0.0
		if resp.Usage.PromptTokens > 0 {
			perK = resp.Usage.Cost / float64(resp.Usage.PromptTokens) * 1000
		}
		t.Logf("%s: prompt=%d completion=%d cached=%d cache_write=%d cost=%.6f (usd/1k prompt tokens=%.4f) generation=%s",
			label, resp.Usage.PromptTokens, resp.Usage.CompletionTokens,
			resp.Usage.CachedTokens, resp.Usage.CacheWriteTokens,
			resp.Usage.Cost, perK, resp.ID)
	}

	// --- AVANT (pre-fix behavior): no cache breakpoint anywhere ---
	first := []Message{UserMessage(bigContext + "\n\n" + question1.Text)}
	avant1, err := client.Complete(ctx, cacheBlind(first))
	if err != nil {
		t.Fatalf("avant call 1: %v", err)
	}
	logPhase("AVANT appel 1", avant1)
	avant2, err := client.Complete(ctx, cacheBlind([]Message{
		first[0], avant1.Message, question2,
	}))
	if err != nil {
		t.Fatalf("avant call 2: %v", err)
	}
	logPhase("AVANT appel 2", avant2)
	if avant2.Usage.CachedTokens != 0 {
		t.Errorf("AVANT (sans breakpoint) : cached=%d, attendu 0 — le fournisseur mettrait en cache sans directive ?",
			avant2.Usage.CachedTokens)
	}

	// --- APRÈS (the fix): CacheLastMessage on every call ---
	apres1, err := client.Complete(ctx, cacheAware([]Message{
		UserMessage(bigContext + "\n\n" + question1.Text),
	}))
	if err != nil {
		t.Fatalf("après call 1: %v", err)
	}
	logPhase("APRES appel 1 (écriture cache)", apres1)
	apres2, err := client.Complete(ctx, cacheAware([]Message{
		first[0], apres1.Message, question2,
	}))
	if err != nil {
		t.Fatalf("après call 2: %v", err)
	}
	logPhase("APRES appel 2 (lecture cache)", apres2)
	if apres2.Usage.CachedTokens <= 0 {
		t.Errorf("APRES (avec breakpoint) : cached=%d, attendu > 0 — le cache n'a pas été lu (routage autre fournisseur ?)", apres2.Usage.CachedTokens)
	}
}
