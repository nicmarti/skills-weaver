package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"dungeons/internal/llm"
)

// nestedFakeClient is a test-only neutral llm.Client for nested-agent tests.
// It replaces the pre-migration Anthropic mock (mock_anthropic.go). Responses
// are selected from the last user message: exact match first, then substring
// match (case-insensitive), then built-in heuristics, then a generic fallback.
type nestedFakeClient struct {
	mu            sync.Mutex
	Responses     map[string]string // question substring → response text
	Script        []llm.Response    // optional scripted sequence (takes precedence)
	scriptIdx     int
	CallCount     int
	LastRequest   *llm.Request
	history       []llm.Request
	SimulateError bool
	ErrorMessage  string
}

func newNestedFakeClient() *nestedFakeClient {
	return &nestedFakeClient{Responses: make(map[string]string)}
}

// SetResponse registers a response for a question substring.
func (f *nestedFakeClient) SetResponse(question, response string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Responses[question] = response
}

func (f *nestedFakeClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	f.mu.Lock()
	f.CallCount++
	last := req
	f.history = append(f.history, last)
	f.LastRequest = &last
	scriptIdx := f.scriptIdx
	script := f.Script
	f.scriptIdx++
	simulateErr := f.SimulateError
	errMsg := f.ErrorMessage
	responses := f.Responses
	question := lastUserText(req.Messages)
	f.mu.Unlock()

	if simulateErr {
		if errMsg == "" {
			errMsg = "mock API error"
		}
		return llm.Response{}, fmt.Errorf("%s", errMsg)
	}

	if scriptIdx < len(script) {
		return script[scriptIdx], nil
	}

	usage := llm.Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		Present:          true,
		Cost:             0.001,
		CachedTokens:     10,
		ReasoningTokens:  2,
	}
	return llm.Response{
		ID:           fmt.Sprintf("fake-gen-%d", f.callCount()),
		Model:        req.Model,
		FinishReason: llm.FinishStop,
		Message:      llm.AssistantMessage(fakeNestedResponse(question, responses)),
		Usage:        usage,
	}, nil
}

func (f *nestedFakeClient) Stream(ctx context.Context, req llm.Request, obs llm.StreamObserver) (llm.Response, error) {
	return llm.Response{}, fmt.Errorf("nested fake client does not implement Stream")
}

func (f *nestedFakeClient) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.CallCount
}

// requests returns the requests seen so far.
func (f *nestedFakeClient) requests() []llm.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]llm.Request, len(f.history))
	copy(out, f.history)
	return out
}

// lastUserText returns the most recent textual user message, mirroring how the
// pre-migration mock keyed its canned responses.
func lastUserText(messages []llm.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == llm.RoleUser && messages[i].Text != "" {
			return messages[i].Text
		}
	}
	return ""
}

func fakeNestedResponse(question string, responses map[string]string) string {
	if response, ok := responses[question]; ok {
		return response
	}

	lower := strings.ToLower(question)
	for key, response := range responses {
		if strings.Contains(lower, strings.ToLower(key)) {
			return response
		}
	}

	switch {
	case strings.Contains(lower, "armor class") || strings.Contains(lower, "ac"):
		return "Armor Class (AC) is calculated as 10 + Dexterity modifier + armor bonus + shield bonus."
	case strings.Contains(lower, "saving throw"):
		return "Saving throws are d20 + ability modifier + proficiency bonus (if proficient)."
	case strings.Contains(lower, "character"):
		return "To create a character, choose species, class, roll stats (4d6 keep highest 3), select skills, and pick equipment."
	case strings.Contains(lower, "world") || strings.Contains(lower, "lore"):
		return "The world has four major factions: Valdorine (maritime), Karvath (military), Lumenciel (religious), and Astrene (scholarly)."
	default:
		return "This is a mock response. The question was: " + question
	}
}
