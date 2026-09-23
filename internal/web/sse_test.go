package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"dungeons/internal/llm"
)

// TestWebOutput_SlowSubscriberDoesNotBlockAgentLoop proves a full event buffer
// (slow subscriber that stopped reading) only delays the producer briefly:
// after the bounded wait the event is dropped so the agent loop keeps running
// instead of blocking forever on a dead consumer.
func TestWebOutput_SlowSubscriberDoesNotBlockAgentLoop(t *testing.T) {
	output := NewWebOutput()
	defer output.Close()

	// Fill the buffered channel without ever reading it.
	for i := 0; i < 1000; i++ {
		output.OnTextChunk("x")
	}
	if got := len(output.eventChan); got != 1000 {
		t.Fatalf("precondition: buffer = %d, want 1000", got)
	}

	done := make(chan struct{})
	go func() {
		output.OnTextChunk("overflow")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("sendEvent blocked forever on a full buffer: the agent loop would stall on a slow subscriber")
	}

	// The overflowing event was dropped, not queued.
	if got := len(output.eventChan); got != 1000 {
		t.Errorf("buffer = %d after overflow, want 1000 (event must be dropped after the bounded wait)", got)
	}
}

// TestWebOutput_ErrorAlwaysCompletes proves an agent error still terminates
// the turn client-side: OnError emits "error" followed by "complete" so web
// consumers never wait for a lifecycle event that never comes.
func TestWebOutput_ErrorAlwaysCompletes(t *testing.T) {
	output := NewWebOutput()
	defer output.Close()

	output.OnError(fmt.Errorf("provider exploded"))

	first := <-output.Events()
	second := <-output.Events()

	if first.Event != "error" {
		t.Errorf("first event = %q, want error", first.Event)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(first.Data), &payload); err != nil {
		t.Fatalf("error event payload: %v", err)
	}
	if payload["error"] != "provider exploded" {
		t.Errorf("error payload = %v", payload)
	}
	if second.Event != "complete" {
		t.Errorf("second event = %q, want complete (terminal lifecycle event)", second.Event)
	}
}

// TestHandleStream_ClientDisconnectCancelsStream proves the SSE handler
// stops when the subscriber goes away: it must not keep the request (or the
// goroutine) alive after the client disconnects.
func TestHandleStream_ClientDisconnectCancelsStream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	output := NewWebOutput()
	defer output.Close()
	redirect := NewOutputRedirector()
	redirect.SetTarget(output)

	sm := NewSessionManager(llm.DefaultModels())
	defer sm.Stop()
	session := &Session{Slug: "sse-test", outputRedirect: redirect}
	sm.sessions["sse-test"] = session

	s := &Server{sessionManager: sm}
	engine := gin.New()
	engine.GET("/play/:slug/stream", s.handleStream)
	server := httptest.NewServer(engine)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/play/sse-test/stream", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	// Prime the stream with one buffered event before connecting.
	output.OnTextChunk("première ligne")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	// The subscriber disconnects mid-stream while the channel stays open.
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(io.Discard, resp.Body)
	}()

	select {
	case <-done:
		// Handler noticed the disconnect and returned: no leaked stream.
	case <-time.After(5 * time.Second):
		t.Fatal("SSE stream kept running after the client disconnected")
	}
}

// TestHandleStream_ChannelCloseSendsDone proves a closed event channel
// terminates the SSE response with a final "done" event.
func TestHandleStream_ChannelCloseSendsDone(t *testing.T) {
	gin.SetMode(gin.TestMode)

	output := NewWebOutput()
	redirect := NewOutputRedirector()
	redirect.SetTarget(output)

	sm := NewSessionManager(llm.DefaultModels())
	defer sm.Stop()
	session := &Session{Slug: "sse-done", outputRedirect: redirect}
	sm.sessions["sse-done"] = session

	s := &Server{sessionManager: sm}
	engine := gin.New()
	engine.GET("/play/:slug/stream", s.handleStream)
	server := httptest.NewServer(engine)
	defer server.Close()

	output.OnTextChunk("ligne")
	output.Close() // terminal condition: producer finished

	resp, err := http.Get(server.URL + "/play/sse-done/stream")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !contains(body, "event: done") {
		t.Errorf("SSE response missing terminal done event:\n%s", body)
	}
}

// TestWebOutput_DetachClosesAfterGrace proves a detached output (subscriber
// gone, nobody re-attached) closes itself after the grace period, so the
// producer's sends become immediate no-ops instead of paying the full-buffer
// 500ms wait on every event.
func TestWebOutput_DetachClosesAfterGrace(t *testing.T) {
	output := NewWebOutput()
	defer output.Close()

	output.Detach(10 * time.Millisecond)
	select {
	case <-output.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("detached output never closed after the grace period")
	}
	if !output.IsClosed() {
		t.Fatal("IsClosed must report true after the detach timer fired")
	}

	// Sends on a closed output must not block and must not queue anything.
	done := make(chan struct{})
	go func() {
		output.OnTextChunk("tombé à l'eau")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("send blocked on a closed output")
	}
	if got := len(output.eventChan); got != 0 {
		t.Errorf("closed output queued %d events, want 0", got)
	}
}

// TestWebOutput_AttachCancelsDetach proves a subscriber re-attaching within
// the grace period keeps the output alive: a page reload mid-turn must not
// lose the rest of the stream.
func TestWebOutput_AttachCancelsDetach(t *testing.T) {
	output := NewWebOutput()
	defer output.Close()

	output.Detach(30 * time.Millisecond)
	output.Attach()
	time.Sleep(80 * time.Millisecond)
	if output.IsClosed() {
		t.Fatal("attach must cancel a pending detach timer")
	}

	// A second Detach after re-attach starts a new countdown (idempotent per
	// pending timer only).
	output.Detach(30 * time.Millisecond)
	output.Attach()
	output.Detach(30 * time.Millisecond)
	select {
	case <-output.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("second detach never closed the output")
	}
}

// TestWebOutput_SendNeverBlocksForeverOnClosedOutput proves the slow path
// (full buffer) unblocks when the output closes mid-wait: no mutex is held
// and no channel-close race can panic.
func TestWebOutput_SendNeverBlocksForeverOnClosedOutput(t *testing.T) {
	output := NewWebOutput()

	for i := 0; i < 1000; i++ {
		output.OnTextChunk("x")
	}

	done := make(chan struct{})
	go func() {
		output.OnTextChunk("bloqué")
		close(done)
	}()

	// Give the slow-path send time to reach its grace-period wait, then close.
	time.Sleep(50 * time.Millisecond)
	output.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("full-buffer send did not unblock after Close")
	}
}

func contains(haystack []byte, needle string) bool {
	return len(needle) == 0 || indexOf(haystack, needle) >= 0
}

func indexOf(haystack []byte, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if string(haystack[i:i+len(needle)]) == needle {
			return i
		}
	}
	return -1
}
