package ambient

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"dungeons/internal/llm"
)

// fakeLyriaClient scripts neutral responses for Lyria parameter tests.
type fakeLyriaClient struct {
	response string
	err      error
	request  llm.Request
}

func (f *fakeLyriaClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	snapshot := req
	f.request = snapshot
	if f.err != nil {
		return llm.Response{}, f.err
	}
	return llm.Response{
		FinishReason: llm.FinishStop,
		Message:      llm.AssistantMessage(f.response),
	}, nil
}

func (f *fakeLyriaClient) Stream(ctx context.Context, req llm.Request, obs llm.StreamObserver) (llm.Response, error) {
	return llm.Response{}, fmt.Errorf("not implemented")
}

// TestGenerateLyriaPrompt_JSONAndClamping proves the parameter call requests
// structured JSON with the Fast-role model and clamps out-of-range values.
func TestGenerateLyriaPrompt_JSONAndClamping(t *testing.T) {
	fake := &fakeLyriaClient{
		response: `{"prompt_en": "lively tavern folk music, lutes", "bpm": 500, "temperature": 9.9, "scene_name": "Taverne"}`,
	}

	params, err := GenerateLyriaPrompt(fake, "", "taverne bruyante et joyeuse")
	if err != nil {
		t.Fatalf("GenerateLyriaPrompt: %v", err)
	}
	if params.Prompt != "lively tavern folk music, lutes" {
		t.Errorf("prompt = %q", params.Prompt)
	}
	if params.BPM != 160 {
		t.Errorf("BPM not clamped: %d, want 160", params.BPM)
	}
	if params.Temperature != 1.3 {
		t.Errorf("temperature not clamped: %v, want 1.3", params.Temperature)
	}
	if params.DisplayName != "Taverne" {
		t.Errorf("scene name = %q", params.DisplayName)
	}

	if fake.request.Model != llm.DefaultModelFast {
		t.Errorf("model = %q, want Fast role default", fake.request.Model)
	}
	if fake.request.JSONResponse == nil {
		t.Error("Lyria parameter generation must request structured JSON output")
	}
	if !fake.request.RequireParameters {
		t.Error("structured-JSON requests must require endpoint parameter support")
	}
	if fake.request.System != lyriaSystemPrompt {
		t.Error("system prompt not carried on the neutral request")
	}
}

// TestGenerateLyriaPrompt_TolerantParsing proves fenced JSON still parses.
func TestGenerateLyriaPrompt_TolerantParsing(t *testing.T) {
	fake := &fakeLyriaClient{
		response: "```json\n{\"prompt_en\": \"epic battle orchestra\", \"bpm\": 140, \"temperature\": 1.1, \"scene_name\": \"Combat\"}\n```",
	}

	params, err := GenerateLyriaPrompt(fake, "", "combat épique")
	if err != nil {
		t.Fatalf("GenerateLyriaPrompt with fenced JSON: %v", err)
	}
	if params.BPM != 140 {
		t.Errorf("BPM = %d", params.BPM)
	}
}

// TestGenerateLyriaPrompt_ClientFailure proves provider errors propagate.
func TestGenerateLyriaPrompt_ClientFailure(t *testing.T) {
	fake := &fakeLyriaClient{err: fmt.Errorf("provider exploded")}

	if _, err := GenerateLyriaPrompt(fake, "", "scene"); err == nil {
		t.Fatal("expected provider error to propagate")
	}
}

// TestGenerateLyriaPrompt_NilClient proves the missing-client guard.
func TestGenerateLyriaPrompt_NilClient(t *testing.T) {
	if _, err := GenerateLyriaPrompt(nil, "", "scene"); err == nil {
		t.Fatal("nil client must be rejected")
	}
}

// TestGenerateLyriaPrompt_Defaults proves empty fields fall back safely.
func TestGenerateLyriaPrompt_Defaults(t *testing.T) {
	fake := &fakeLyriaClient{response: `{"bpm": 90, "temperature": 1.0}`}
	params, err := GenerateLyriaPrompt(fake, "", "forêt mystérieuse")
	if err != nil {
		t.Fatalf("GenerateLyriaPrompt: %v", err)
	}
	if !strings.Contains(params.Prompt, "medieval fantasy ambient music") {
		t.Errorf("missing prompt fallback: %q", params.Prompt)
	}
	if params.DisplayName != "forêt mystérieuse" {
		t.Errorf("scene name fallback = %q", params.DisplayName)
	}
}
