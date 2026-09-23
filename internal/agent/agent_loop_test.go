package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/llm"
)

// fakeLLMClient is a test-only llm.Client returning scripted responses.
type fakeLLMClient struct {
	mu        sync.Mutex
	responses []llm.Response
	calls     int
	requests  []llm.Request
	// block, when non-nil, stalls Stream until closed or ctx is cancelled.
	block chan struct{}
}

func (f *fakeLLMClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.Response{}, fmt.Errorf("not implemented in fake")
}

func (f *fakeLLMClient) Stream(ctx context.Context, req llm.Request, obs llm.StreamObserver) (llm.Response, error) {
	f.mu.Lock()
	f.requests = append(f.requests, req)
	call := f.calls
	f.calls++
	f.mu.Unlock()

	if f.block != nil {
		select {
		case <-ctx.Done():
			return llm.Response{}, ctx.Err()
		case <-f.block:
		}
	}
	if call >= len(f.responses) {
		return llm.Response{}, fmt.Errorf("fake client exhausted scripted responses")
	}
	resp := f.responses[call]
	if resp.Message.Text != "" {
		obs.OnTextDelta(resp.Message.Text)
	}
	return resp, nil
}

func (f *fakeLLMClient) requestCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

// recordingOutput captures output-handler events for assertions.
type recordingOutput struct {
	mu         sync.Mutex
	text       strings.Builder
	toolStarts []string
	toolDone   []string
	onError    []error
	completed  int
}

func (r *recordingOutput) OnTextChunk(text string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.text.WriteString(text)
}

func (r *recordingOutput) OnToolStart(toolName, toolID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.toolStarts = append(r.toolStarts, toolName)
}

func (r *recordingOutput) OnToolComplete(toolName string, result interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.toolDone = append(r.toolDone, toolName)
}

func (r *recordingOutput) OnAgentInvocationStart(agentName string) {}
func (r *recordingOutput) OnAgentInvocationComplete(agentName string, duration time.Duration) {
}

func (r *recordingOutput) OnError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onError = append(r.onError, err)
}

func (r *recordingOutput) OnComplete() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed++
}

func (r *recordingOutput) fullText() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.text.String()
}

// probeStub is a state-modifying tool recording its invocations.
type probeStub struct {
	mu    sync.Mutex
	calls []map[string]interface{}
}

func (p *probeStub) Name() string        { return "probe" }
func (p *probeStub) Description() string { return "probe tool" }
func (p *probeStub) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{"value": map[string]interface{}{"type": "integer"}},
		"required":   []string{"value"},
	}
}

func (p *probeStub) Execute(params map[string]interface{}) (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, params)
	return map[string]interface{}{"success": true, "echo": params["value"]}, nil
}

// loopStartSessionStub is a start_session stand-in with a schema valid for
// ToolDefinitions (the shared startSessionStub has no properties key).
type loopStartSessionStub struct {
	adv   *adventure.Adventure
	calls int
}

func (s *loopStartSessionStub) Name() string        { return "start_session" }
func (s *loopStartSessionStub) Description() string { return "start a session" }
func (s *loopStartSessionStub) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (s *loopStartSessionStub) Execute(params map[string]interface{}) (interface{}, error) {
	s.calls++
	if _, err := s.adv.StartSession(); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "system_brief": "BRIEF"}, nil
}

// setupMainLoopTest creates a chdir'd temp workspace with a dungeon-master
// persona and a fully wired Agent over a fake client. The returned cleanup
// restores the previous working directory.
func setupMainLoopTest(t *testing.T, responses []llm.Response) (*Agent, *fakeLLMClient, *recordingOutput, *probeStub, *loopStartSessionStub, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "dungeon-master.md"), []byte("Tu es le Maître du Donjon."), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	cleanup := func() { os.Chdir(origDir) }

	adv := adventure.New("Loop Test", "")
	adv.SetBasePath(t.TempDir())

	actx := &AdventureContext{
		Adventure: adv,
		Party:     &adventure.Party{Characters: []string{}},
		Inventory: &adventure.SharedInventory{Gold: 0},
		State:     &adventure.GameState{CurrentLocation: "Taverne"},
	}
	registry := NewToolRegistry(actx)
	probe := &probeStub{}
	registry.Register(probe)
	sessionStub := &loopStartSessionStub{adv: adv}
	registry.Register(sessionStub)

	fake := &fakeLLMClient{responses: responses}
	out := &recordingOutput{}

	agent := &Agent{
		client:          fake,
		model:           llm.DefaultModelDM,
		toolRegistry:    registry,
		conversationCtx: NewConversationContextWithLimit(mainAgentContextTokenLimit),
		adventureCtx:    actx,
		outputHandler:   out,
		personaLoader:   NewPersonaLoader(),
	}
	return agent, fake, out, probe, sessionStub, cleanup
}

func toolCallResp(calls ...llm.ToolCall) llm.Response {
	return llm.Response{
		FinishReason: llm.FinishToolCalls,
		Message:      llm.AssistantToolCallMessage("", calls),
	}
}

func textResp(text string, usage llm.Usage) llm.Response {
	return llm.Response{
		FinishReason: llm.FinishStop,
		Message:      llm.AssistantMessage(text),
		Usage:        usage,
	}
}

// TestMainLoop_StreamedTurnWithParallelTools proves a streamed main turn that
// executes parallel tool calls with matching IDs and JSON results, receives a
// final assistant message, aggregates usage, auto-starts the session, and
// completes the output.
func TestMainLoop_StreamedTurnWithParallelTools(t *testing.T) {
	agent, fake, out, probe, sessionStub, cleanup := setupMainLoopTest(t, []llm.Response{
		toolCallResp(
			llm.ToolCall{ID: "call-a", Name: "probe", Arguments: json.RawMessage(`{"value":1}`)},
			llm.ToolCall{ID: "call-b", Name: "probe", Arguments: json.RawMessage(`{"value":2}`)},
		),
		textResp("Le groupe avance.", llm.Usage{
			PromptTokens: 100, CompletionTokens: 20, TotalTokens: 120, Present: true, Cost: 0.001,
		}),
	})
	defer cleanup()

	if err := agent.ProcessUserMessage("j'avance"); err != nil {
		t.Fatalf("ProcessUserMessage: %v", err)
	}

	if sessionStub.calls != 1 {
		t.Errorf("session guard-rail: start_session called %d times, want 1", sessionStub.calls)
	}
	if got := out.fullText(); got != "Le groupe avance." {
		t.Errorf("streamed text = %q", got)
	}
	if out.completed != 1 {
		t.Errorf("OnComplete called %d times, want 1", out.completed)
	}
	if len(out.toolStarts) != 2 || out.toolStarts[0] != "probe" || out.toolStarts[1] != "probe" {
		t.Errorf("tool starts = %v", out.toolStarts)
	}
	if len(probe.calls) != 2 {
		t.Fatalf("probe executed %d times, want 2", len(probe.calls))
	}
	if probe.calls[0]["value"].(float64) != 1 || probe.calls[1]["value"].(float64) != 2 {
		t.Errorf("probe arguments = %v", probe.calls)
	}

	msgs := agent.conversationCtx.NeutralMessages()
	if len(msgs) != 5 {
		t.Fatalf("conversation has %d messages, want 5 (user, assistant+tools, result a, result b, final): %+v", len(msgs), msgs)
	}
	if msgs[1].Role != llm.RoleAssistant || len(msgs[1].ToolCalls) != 2 ||
		msgs[1].ToolCalls[0].ID != "call-a" || msgs[1].ToolCalls[1].ID != "call-b" {
		t.Fatalf("assistant tool-call message wrong: %+v", msgs[1])
	}
	if msgs[2].Role != llm.RoleTool || msgs[2].ToolResults[0].ToolCallID != "call-a" {
		t.Fatalf("tool result a wrong: %+v", msgs[2])
	}
	var resultA map[string]interface{}
	if err := json.Unmarshal([]byte(msgs[2].ToolResults[0].Content), &resultA); err != nil || resultA["success"] != true {
		t.Fatalf("tool result a not valid JSON: %+v", msgs[2].ToolResults[0])
	}
	if msgs[3].Role != llm.RoleTool || msgs[3].ToolResults[0].ToolCallID != "call-b" {
		t.Fatalf("tool result b wrong: %+v", msgs[3])
	}
	if msgs[4].Role != llm.RoleAssistant || msgs[4].Text != "Le groupe avance." {
		t.Fatalf("final assistant message wrong: %+v", msgs[4])
	}

	// Tool results must be matched back to the request the model saw.
	lastReq := fake.requests[len(fake.requests)-1]
	if len(lastReq.Messages) < 4 {
		t.Fatalf("follow-up request missing tool results: %+v", lastReq.Messages)
	}

	// Usage aggregation covers both provider calls of the turn.
	usage := agent.LastTurnUsage()
	if !usage.Present || usage.PromptTokens != 100 || usage.CompletionTokens != 20 || usage.TotalTokens != 120 {
		t.Errorf("aggregated usage = %+v", usage)
	}
}

// TestMainLoop_BoundedToolIterations proves the tool loop stops after
// mainAgentMaxToolIterations instead of looping forever.
func TestMainLoop_BoundedToolIterations(t *testing.T) {
	alwaysTools := toolCallResp(llm.ToolCall{ID: "call-x", Name: "probe", Arguments: json.RawMessage(`{"value":1}`)})
	responses := make([]llm.Response, mainAgentMaxToolIterations+1)
	for i := range responses {
		responses[i] = alwaysTools
	}
	agent, fake, out, _, _, cleanup := setupMainLoopTest(t, responses)
	defer cleanup()

	err := agent.ProcessUserMessage("encore")
	if err != nil {
		t.Fatalf("bounded loop must end the turn via output, not an error return: %v", err)
	}
	if got := fake.requestCount(); got != mainAgentMaxToolIterations {
		t.Errorf("provider called %d times, want %d", got, mainAgentMaxToolIterations)
	}
	if len(out.onError) == 0 || !strings.Contains(out.onError[0].Error(), "tool loop exceeded") {
		t.Errorf("expected tool-loop-exceeded error, got %v", out.onError)
	}
	if out.completed != 1 {
		t.Errorf("OnComplete called %d times, want 1 (terminal event after error)", out.completed)
	}
}

// TestMainLoop_CancelContext proves a cancelled context aborts the turn and
// no partial assistant history is committed.
func TestMainLoop_CancelContext(t *testing.T) {
	agent, _, out, probe, _, cleanup := setupMainLoopTest(t, nil)
	defer cleanup()
	blocking := &fakeLLMClient{block: make(chan struct{})}
	agent.client = blocking

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- agent.ProcessUserMessageContext(ctx, "long narration")
	}()
	cancel()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "context") {
			t.Fatalf("expected context error, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("turn did not react to cancellation")
	}
	if len(probe.calls) != 0 {
		t.Errorf("no tool should run after cancellation")
	}
	msgs := agent.conversationCtx.NeutralMessages()
	if len(msgs) != 1 || msgs[0].Role != llm.RoleUser {
		t.Fatalf("partial history committed after cancellation: %+v", msgs)
	}
	if out.completed != 0 {
		t.Errorf("no complete event expected on aborted turn, got %d", out.completed)
	}
}

// TestMainLoop_IncompleteFinishReason proves length/content_filter/error
// finishes end the turn without committing the partial assistant message.
func TestMainLoop_IncompleteFinishReason(t *testing.T) {
	for _, reason := range []llm.FinishReason{llm.FinishLength, llm.FinishContentFilter, llm.FinishError} {
		t.Run(string(reason), func(t *testing.T) {
			agent, _, out, _, _, cleanup := setupMainLoopTest(t, []llm.Response{
				{
					FinishReason: reason,
					Message:     llm.AssistantMessage("partial text that must not be committed"),
				},
			})
			defer cleanup()

			if err := agent.ProcessUserMessage("test"); err != nil {
				t.Fatalf("incomplete finish must end the turn, not return an error: %v", err)
			}
			if len(out.onError) == 0 {
				t.Error("expected OnError for incomplete finish reason")
			}
			if out.completed != 1 {
				t.Errorf("OnComplete = %d, want 1 (terminal event)", out.completed)
			}
			for _, msg := range agent.conversationCtx.NeutralMessages() {
				if msg.Role == llm.RoleAssistant {
					t.Fatalf("partial assistant message committed for finish %q: %+v", reason, msg)
				}
			}
		})
	}
}

// TestMainLoop_SetModelValidation proves invalid selections are rejected and
// aliases resolve through the neutral catalog.
func TestMainLoop_SetModelValidation(t *testing.T) {
	agent, _, _, _, _, cleanup := setupMainLoopTest(t, nil)
	defer cleanup()

	if err := agent.SetModel("bogus"); err == nil {
		t.Error("invalid model must be rejected, not silently replaced")
	}
	if agent.GetModel() != llm.DefaultModelDM {
		t.Errorf("rejected selection changed the model: %q", agent.GetModel())
	}
	if err := agent.SetModel("opus"); err != nil {
		t.Fatalf("opus alias: %v", err)
	}
	if agent.GetModel() != llm.ModelOpus5 {
		t.Errorf("opus alias resolved to %q", agent.GetModel())
	}
	if err := agent.SetModel("openai/gpt-6"); err != nil {
		t.Fatalf("provider-qualified ID: %v", err)
	}
	if agent.GetModel() != "openai/gpt-6" {
		t.Errorf("provider-qualified ID resolved to %q", agent.GetModel())
	}
}