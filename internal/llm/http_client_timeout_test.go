package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestDefaultHTTPClient_NoTotalBodyCap guards the campaign-plan timeout fix
// (docs/issues-2026-09-24.md): the production client must NOT set
// http.Client.Timeout — that cap covers reading the response body too and
// killed any stream longer than 60s mid-flight. The stuck-provider guard is
// the transport's ResponseHeaderTimeout instead.
func TestDefaultHTTPClient_NoTotalBodyCap(t *testing.T) {
	client := defaultHTTPClient()

	if client.Timeout != 0 {
		t.Errorf("whole-request timeout = %v, want 0 (streamed bodies may take minutes)", client.Timeout)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport = %T, want *http.Transport", client.Transport)
	}
	if transport.ResponseHeaderTimeout != defaultResponseHeaderTimeout {
		t.Errorf("ResponseHeaderTimeout = %v, want %v", transport.ResponseHeaderTimeout, defaultResponseHeaderTimeout)
	}
}

// TestStream_SlowBodyOutlivesHeaderTimeout proves the mechanism end to end:
// a stream that drips chunks over a duration LONGER than the header budget
// must complete, while the old design (http.Client.Timeout) would have cut
// it mid-flight. The same dripping server is probed twice: once through a
// client shaped like the production default (header-only budget), once
// through the old total-cap client.
func TestStream_SlowBodyOutlivesHeaderTimeout(t *testing.T) {
	const budget = 150 * time.Millisecond
	const drip = 250 * time.Millisecond

	sse := func() string {
		return sseChunk(chunkBase("chatcmpl-slow", "anthropic/claude-sonnet-5",
			`[{"index":0,"delta":{"role":"assistant","content":"première "},"finish_reason":null}]`))
	}
	final := sseChunk(`{"id":"chatcmpl-slow","created":1700000000,"object":"chat.completion.chunk","model":"anthropic/claude-sonnet-5","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`) +
		"data: [DONE]\n\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		// Headers arrive immediately, then the body drips well past the budget.
		_, _ = w.Write([]byte(sse()))
		flusher.Flush()
		time.Sleep(drip)
		_, _ = w.Write([]byte(sseChunk(chunkBase("chatcmpl-slow", "anthropic/claude-sonnet-5",
			`[{"index":0,"delta":{"content":"partie"},"finish_reason":null}]`))))
		flusher.Flush()
		time.Sleep(drip)
		_, _ = w.Write([]byte(final))
		flusher.Flush()
	}))
	defer server.Close()

	headerOnly := http.DefaultTransport.(*http.Transport).Clone()
	headerOnly.ResponseHeaderTimeout = budget
	productionLike := &http.Client{Transport: headerOnly}

	client, err := newOpenRouterClientWithHTTP(testConfig(t), newRetryNone(), server.URL, productionLike)
	if err != nil {
		t.Fatalf("newOpenRouterClientWithHTTP: %v", err)
	}
	resp, err := client.Stream(context.Background(), Request{
		Model:    ModelSonnet5,
		Messages: []Message{UserMessage("lent")},
	}, nil)
	if err != nil {
		t.Fatalf("slow body must outlive the header budget: %v", err)
	}
	if resp.Message.Text != "première partie" {
		t.Errorf("accumulated text = %q, want the full dripped stream", resp.Message.Text)
	}

	// Contrast: the old design (whole-request cap) dies on the same stream.
	oldStyle := &http.Client{Timeout: budget}
	oldClient, err := newOpenRouterClientWithHTTP(testConfig(t), newRetryNone(), server.URL, oldStyle)
	if err != nil {
		t.Fatalf("newOpenRouterClientWithHTTP (old style): %v", err)
	}
	if _, err := oldClient.Stream(context.Background(), Request{
		Model:    ModelSonnet5,
		Messages: []Message{UserMessage("lent")},
	}, nil); err == nil {
		t.Fatal("old total-cap client should have been killed mid-stream on a slow body")
	}
}
