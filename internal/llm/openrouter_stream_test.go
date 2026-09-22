package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// sseChunk formats one SSE data line carrying a ChatStreamChunk directly
// (standard OpenAI wire format; the SDK wraps it internally).
func sseChunk(chunkJSON string) string {
	return fmt.Sprintf("data: %s\n\n", chunkJSON)
}

func chunkBase(id, model string, choices string) string {
	return fmt.Sprintf(`{"id":%q,"created":1700000000,"object":"chat.completion.chunk","model":%q,"choices":%s}`, id, model, choices)
}

var streamFixtures = ": OPENROUTER PROCESSING\n\n" +
	sseChunk(chunkBase("chatcmpl-stream-1", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"role":"assistant","content":"Tu "},"finish_reason":null}]`)) +
	sseChunk(chunkBase("chatcmpl-stream-1", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"content":"lances"},"finish_reason":null}]`)) +
	// interleaved parallel tool call fragments
	sseChunk(chunkBase("chatcmpl-stream-1", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_a","type":"function","function":{"name":"roll_dice","arguments":"{\"not"}}]},"finish_reason":null}]`)) +
	sseChunk(chunkBase("chatcmpl-stream-1", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"tool_calls":[{"index":1,"id":"call_b","type":"function","function":{"name":"get_monster"}}]},"finish_reason":null}]`)) +
	sseChunk(chunkBase("chatcmpl-stream-1", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"ation\":\"2d6\"}"}}]},"finish_reason":null}]`)) +
	sseChunk(chunkBase("chatcmpl-stream-1", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"tool_calls":[{"index":1,"function":{"arguments":"{\"cr\":5}"}}]},"finish_reason":null}]`)) +
	// final usage chunk repeats the terminal finish reason
	sseChunk(`{"id":"chatcmpl-stream-1","created":1700000000,"object":"chat.completion.chunk","model":"anthropic/claude-sonnet-5","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":500,"completion_tokens":40,"total_tokens":540,"prompt_tokens_details":{"cached_tokens":400},"cost":0.0031,"server_tool_use_details":{"tool_calls_requested":2,"tool_calls_executed":2}}}`) +
	sseChunk(`{"id":"chatcmpl-stream-1","created":1700000000,"object":"chat.completion.chunk","model":"anthropic/claude-sonnet-5","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":500,"completion_tokens":40,"total_tokens":540,"prompt_tokens_details":{"cached_tokens":400},"cost":0.0031,"server_tool_use_details":{"tool_calls_requested":2,"tool_calls_executed":2}}}`) +
	"data: [DONE]\n\n"

type recordingObserver struct {
	deltas []string
	mu     sync.Mutex
}

func (o *recordingObserver) OnTextDelta(text string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.deltas = append(o.deltas, text)
}

func TestStream_TextAndInterleavedToolCalls(t *testing.T) {
	var capturedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody = readRequestBody(t, r)
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		_, _ = w.Write([]byte(streamFixtures))
		flusher.Flush()
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	obs := &recordingObserver{}
	resp, err := client.Stream(context.Background(), Request{
		Model: ModelSonnet5,
		Messages: []Message{
			UserMessage("Lance un dé"),
		},
	}, obs)
	if err != nil {
		t.Fatalf("Stream() error: %v", err)
	}

	if got := capturedBody["stream"].(bool); !got {
		t.Error("stream should be true for Stream()")
	}

	if got := obs.deltas[0]; got != "Tu " {
		t.Errorf("first delta = %q", got)
	}
	if got := obs.deltas[1]; got != "lances" {
		t.Errorf("second delta = %q", got)
	}
	if len(obs.deltas) != 2 {
		t.Errorf("deltas = %v, want exactly the two text fragments", obs.deltas)
	}

	if len(resp.Message.ToolCalls) != 2 {
		t.Fatalf("tool calls = %+v, want 2 assembled calls", resp.Message.ToolCalls)
	}
	first := resp.Message.ToolCalls[0]
	if first.ID != "call_a" || first.Name != "roll_dice" {
		t.Errorf("call A = %+v", first)
	}
	if string(first.Arguments) != `{"notation":"2d6"}` {
		t.Errorf("call A arguments = %s (want fragments joined in order)", first.Arguments)
	}
	second := resp.Message.ToolCalls[1]
	if second.ID != "call_b" || second.Name != "get_monster" {
		t.Errorf("call B = %+v", second)
	}
	if string(second.Arguments) != `{"cr":5}` {
		t.Errorf("call B arguments = %s", second.Arguments)
	}

	if resp.FinishReason != FinishToolCalls {
		t.Errorf("finish reason = %q, want tool_calls", resp.FinishReason)
	}
	if resp.ID != "chatcmpl-stream-1" || resp.Model != ModelSonnet5 {
		t.Errorf("resp id/model = %q/%q", resp.ID, resp.Model)
	}
	if !resp.Usage.Present {
		t.Fatal("usage should be captured from the final chunk")
	}
	if resp.Usage.PromptTokens != 500 || resp.Usage.CachedTokens != 400 || resp.Usage.Cost != 0.0031 {
		t.Errorf("usage = %+v", resp.Usage)
	}
	if resp.Usage.ServerToolCallsRequested != 2 || resp.Usage.ServerToolCallsExecuted != 2 {
		t.Errorf("server tool usage = %+v", resp.Usage)
	}
	if resp.Message.Text != "Tu lances" {
		t.Errorf("accumulated text = %q", resp.Message.Text)
	}
}

func TestStream_MidStreamError(t *testing.T) {
	payload := sseChunk(chunkBase("chatcmpl-stream-err", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"content":"Hel"},"finish_reason":null}]`)) +
		`data: {"id":"chatcmpl-stream-err","created":1,"object":"chat.completion.chunk","model":"anthropic/claude-sonnet-5","choices":[],"error":{"code":502,"message":"Provider returned error"}}` + "\n\n" +
		"data: [DONE]\n\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	obs := &recordingObserver{}
	resp, err := client.Stream(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}}, obs)
	if err == nil {
		t.Fatal("Stream() should fail on a mid-stream error chunk")
	}
	llmErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T, want *llm.Error", err)
	}
	if llmErr.Kind != ErrKindMidStream {
		t.Errorf("kind = %q, want mid_stream", llmErr.Kind)
	}
	if llmErr.Status != 502 {
		t.Errorf("status = %d, want 502", llmErr.Status)
	}
	if llmErr.Retryable {
		t.Error("mid-stream errors must never be retried")
	}
	if resp.Message.Text != "Hel" {
		t.Errorf("partial text should be preserved in response: %q", resp.Message.Text)
	}
}

func TestStream_IncompleteToolCall(t *testing.T) {
	// Tool call missing its ID: the adapter must reject it instead of
	// executing a call it cannot correlate.
	payload := sseChunk(chunkBase("chatcmpl-stream-inc", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"tool_calls":[{"index":0,"type":"function","function":{"name":"roll_dice"}}]},"finish_reason":"tool_calls"}]`)) +
		"data: [DONE]\n\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	_, err = client.Stream(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}}, nil)
	if err == nil {
		t.Fatal("Stream() should fail on an incomplete streamed tool call")
	}
	llmErr, ok := err.(*Error)
	if !ok || llmErr.Kind != ErrKindMalformed {
		t.Fatalf("error = %v, want malformed_stream", err)
	}
}

func TestStream_MalformedToolArguments(t *testing.T) {
	payload := sseChunk(chunkBase("chatcmpl-stream-bad", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_x","type":"function","function":{"name":"roll_dice","arguments":"{not json"}}]},"finish_reason":"tool_calls"}]`)) +
		"data: [DONE]\n\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	_, err = client.Stream(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}}, nil)
	if err == nil {
		t.Fatal("Stream() should fail on malformed streamed tool arguments")
	}
	llmErr, ok := err.(*Error)
	if !ok || llmErr.Kind != ErrKindMalformed {
		t.Fatalf("error = %v, want malformed_stream", err)
	}
}

func TestStream_CancelWhilePending(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		close(started)
		// Hold the stream open until the test releases it; never stream data.
		<-release
	}))
	defer func() {
		close(release)
		server.Close()
	}()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-started
		cancel()
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err = client.Stream(ctx, Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}}, nil)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Stream() did not return within 10s after context cancellation")
	}

	if err == nil {
		t.Fatal("Stream() should fail when the context is canceled")
	}
	llmErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T, want *llm.Error", err)
	}
	if llmErr.Kind != ErrKindCanceled {
		t.Errorf("kind = %q, want canceled", llmErr.Kind)
	}
}

func TestStream_SendsStreamTrue(t *testing.T) {
	var capturedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody = readRequestBody(t, r)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sseChunk(chunkBase("chatcmpl-stream-2", "anthropic/claude-sonnet-5",
			`[{"index":0,"delta":{"content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":2,"total_tokens":12}`)) +
			"data: [DONE]\n\n"))
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	resp, err := client.Stream(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}}, nil)
	if err != nil {
		t.Fatalf("Stream() error: %v", err)
	}
	if got := capturedBody["stream"].(bool); !got {
		t.Error("stream should be true for Stream()")
	}
	if resp.FinishReason != FinishStop {
		t.Errorf("finish reason = %q", resp.FinishReason)
	}
	if resp.Message.Text != "ok" {
		t.Errorf("text = %q", resp.Message.Text)
	}
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}
