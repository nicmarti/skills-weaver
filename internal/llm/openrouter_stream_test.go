package llm

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/OpenRouterTeam/go-sdk/retry"
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

// closeTrackingHTTPClient wraps responses so tests can observe that the
// adapter closes every response body (no SSE stream is ever left open).
type closeTrackingHTTPClient struct {
	inner  *http.Client
	mu     sync.Mutex
	bodies []*trackingBody
}

func (c *closeTrackingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.inner.Do(req)
	if resp != nil && resp.Body != nil {
		body := newTrackingBody()
		body.ReadCloser = resp.Body
		c.mu.Lock()
		c.bodies = append(c.bodies, body)
		c.mu.Unlock()
		resp.Body = body
	}
	return resp, err
}

type trackingBody struct {
	io.ReadCloser
	closed chan struct{}
	once   sync.Once
}

func newTrackingBody() *trackingBody {
	return &trackingBody{closed: make(chan struct{})}
}

func (b *trackingBody) Close() error {
	b.once.Do(func() { close(b.closed) })
	return b.ReadCloser.Close()
}

// TestStream_BodyClosedOnAllExitPaths proves the adapter closes the SSE
// response body on every exit path: success, mid-stream error, malformed or
// incomplete tool calls, and a genuine connection drop mid-stream. A leaked
// body would keep the HTTP connection (and its buffers) alive forever.
func TestStream_BodyClosedOnAllExitPaths(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		rst     bool // abruptly reset the TCP connection mid-stream
	}{
		{
			name: "success",
			payload: sseChunk(chunkBase("chatcmpl-close-ok", "anthropic/claude-sonnet-5",
				`[{"index":0,"delta":{"content":"ok"},"finish_reason":"stop"}]`)) +
				"data: [DONE]\n\n",
		},
		{
			name: "mid-stream error",
			payload: sseChunk(chunkBase("chatcmpl-close-err", "anthropic/claude-sonnet-5",
				`[{"index":0,"delta":{"content":"Hel"},"finish_reason":null}]`)) +
				`data: {"id":"x","choices":[],"error":{"code":502,"message":"upstream failed"}}` + "\n\n",
		},
		{
			name: "incomplete tool call",
			payload: sseChunk(chunkBase("chatcmpl-close-inc", "anthropic/claude-sonnet-5",
				`[{"index":0,"delta":{"tool_calls":[{"index":0,"type":"function","function":{"name":"roll_dice"}}]},"finish_reason":"tool_calls"}]`)) +
				"data: [DONE]\n\n",
		},
		{
			name: "malformed tool arguments",
			payload: sseChunk(chunkBase("chatcmpl-close-bad", "anthropic/claude-sonnet-5",
				`[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_x","type":"function","function":{"name":"roll_dice","arguments":"{not json"}}]},"finish_reason":"tool_calls"}]`)) +
				"data: [DONE]\n\n",
		},
		{
			name: "connection reset mid-stream",
			payload: sseChunk(chunkBase("chatcmpl-close-rst", "anthropic/claude-sonnet-5",
				`[{"index":0,"delta":{"content":"Tu lances le d"},"finish_reason":null}]`)),
			rst: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.rst {
					// Bypass net/http and kill the TCP connection with a
					// RST mid-stream: a genuine provider-incident shape.
					hj := w.(http.Hijacker)
					conn, rw, err := hj.Hijack()
					if err != nil {
						return
					}
					defer conn.Close()
					_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\n\r\n")
					_, _ = rw.WriteString(tc.payload)
					_ = rw.Flush()
					time.Sleep(50 * time.Millisecond)
					if tcp, ok := conn.(*net.TCPConn); ok {
						_ = tcp.SetLinger(0)
					}
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte(tc.payload))
			}))
			defer server.Close()

			tracking := &closeTrackingHTTPClient{inner: server.Client()}
			client, err := newOpenRouterClientWithHTTP(testConfig(t), nil, server.URL, tracking)
			if err != nil {
				t.Fatalf("newOpenRouterClientWithHTTP: %v", err)
			}

			obs := &recordingObserver{}
			resp, err := client.Stream(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}}, obs)
			if tc.name != "success" && err == nil {
				t.Fatalf("Stream() should report an error on path %q, got none (resp=%+v)", tc.name, resp)
			}

			tracking.mu.Lock()
			bodies := append([]*trackingBody(nil), tracking.bodies...)
			tracking.mu.Unlock()
			if len(bodies) == 0 {
				t.Fatal("no response body was observed; test client wiring is broken")
			}
			for i, body := range bodies {
				select {
				case <-body.closed:
				case <-time.After(5 * time.Second):
					t.Fatalf("response body %d was never closed after Stream() returned: SSE stream leaked", i)
				}
			}
		})
	}
}

// TestStream_NoRetryAfterStreamOpens proves a stream that starts successfully
// (200 + emitted text delta) is never replayed when the connection dies
// mid-flight: a genuine provider mid-stream incident surfaces as an error on
// the partially accumulated response, not as a silent replay.
func TestStream_NoRetryAfterStreamOpens(t *testing.T) {
	var requests int32
	payload := sseChunk(chunkBase("chatcmpl-abrupt", "anthropic/claude-sonnet-5",
		`[{"index":0,"delta":{"content":"Tu lances le dé et..."},"finish_reason":null}]`))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		hj := w.(http.Hijacker)
		conn, rw, err := hj.Hijack()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\n\r\n")
		_, _ = rw.WriteString(payload)
		_ = rw.Flush()
		// Let the client read the delta, then kill the connection (RST).
		time.Sleep(50 * time.Millisecond)
		if tcp, ok := conn.(*net.TCPConn); ok {
			_ = tcp.SetLinger(0)
		}
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	obs := &recordingObserver{}
	resp, err := client.Stream(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}}, obs)
	if err == nil {
		t.Fatalf("Stream() must fail on a mid-stream connection reset (resp=%+v)", resp)
	}
	// The SDK's EventStream swallows terminal read errors; the adapter's
	// read-error holder must surface the dead connection as mid_stream.
	llmErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T, want *llm.Error", err)
	}
	if llmErr.Kind != ErrKindMidStream {
		t.Errorf("kind = %q, want mid_stream", llmErr.Kind)
	}
	if llmErr.Retryable {
		t.Error("a stream that already emitted deltas must never be replayed")
	}

	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Errorf("requests = %d, want exactly 1 (no replay after emitted deltas)", got)
	}
	if len(obs.deltas) == 0 || obs.deltas[0] != "Tu lances le dé et..." {
		t.Errorf("emitted text delta lost before the drop: %v", obs.deltas)
	}
	if resp.Message.Text != "Tu lances le dé et..." {
		t.Errorf("accumulated text = %q", resp.Message.Text)
	}
}

// TestComplete_NoRetryOn401And402WithProductionRetry proves 4XX
// authentication failures are never replayed even with the production retry
// configuration enabled (previous classification tests disabled retries).
func TestComplete_NoRetryOn401And402WithProductionRetry(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		wantErr ErrorKind
	}{
		{"401", http.StatusUnauthorized, `{"error":{"code":401,"message":"Invalid API key"}}`, ErrKindAuth},
		{"402", http.StatusPaymentRequired, `{"error":{"code":402,"message":"Insufficient credits"}}`, ErrKindPayment},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var requests int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requests, 1)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			// nil retries = production defaultRetry() configuration.
			client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
			if err != nil {
				t.Fatalf("newOpenRouterClient: %v", err)
			}

			_, err = client.Complete(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}})
			if err == nil {
				t.Fatal("Complete() should fail on 4XX")
			}
			llmErr, ok := err.(*Error)
			if !ok {
				t.Fatalf("error type = %T, want *llm.Error", err)
			}
			if llmErr.Kind != tc.wantErr {
				t.Errorf("kind = %q, want %q", llmErr.Kind, tc.wantErr)
			}
			if got := atomic.LoadInt32(&requests); got != 1 {
				t.Errorf("requests = %d, want exactly 1 (4XX must not be retried with production retry config)", got)
			}
		})
	}
}

// TestComplete_RetryExhaustionIsBounded proves the adapter gives up after a
// bounded number of attempts when the provider is persistently unavailable,
// instead of retrying forever.
func TestComplete_RetryExhaustionIsBounded(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"code":503,"message":"provider unavailable"}}`))
	}))
	defer server.Close()

	// Fast backoff so exhaustion happens in milliseconds; the bound being
	// proven is that attempts stop, not the specific count.
	fast := retry.Config{
		Strategy: "backoff",
		Backoff: &retry.BackoffStrategy{
			InitialInterval: 1,
			MaxInterval:     2,
			Exponent:        1.5,
			MaxElapsedTime:  15,
		},
		RetryConnectionErrors: true,
	}
	client, err := newOpenRouterClient(testConfig(t), &fast, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	_, err = client.Complete(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}})
	if err == nil {
		t.Fatal("Complete() should fail when the provider is persistently unavailable")
	}
	llmErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T, want *llm.Error", err)
	}
	if llmErr.Status != 503 || llmErr.Kind != ErrKindUnavailable {
		t.Errorf("error = %+v, want 503/unavailable", llmErr)
	}

	attempts := atomic.LoadInt32(&requests)
	if attempts < 2 {
		t.Errorf("attempts = %d, want at least one retry (503 is retryable)", attempts)
	}
	if attempts > 50 {
		t.Errorf("attempts = %d, retry exhaustion must stay bounded", attempts)
	}
}
