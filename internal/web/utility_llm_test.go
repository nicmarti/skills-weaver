package web

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"dungeons/internal/agent"
	"dungeons/internal/llm"
	"dungeons/internal/tarot"
)

// fakeWebLLMClient scripts neutral responses for web utility tests.
type fakeWebLLMClient struct {
	response string
	err      error
	request  llm.Request
}

func (f *fakeWebLLMClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
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

func (f *fakeWebLLMClient) Stream(ctx context.Context, req llm.Request, obs llm.StreamObserver) (llm.Response, error) {
	return llm.Response{}, fmt.Errorf("not implemented")
}

// TestCampaignLLMRequest_Multimodal proves the campaign request carries the
// world-map image alongside the prompt, targets the Campaign model, and asks
// for structured JSON with endpoint parameter support.
func TestCampaignLLMRequest_Multimodal(t *testing.T) {
	resources := &agent.WorldResources{
		MapDescription:    "Les Quatre Royaumes...",
		MapImageBase64:    "cXVhdHJl",
		MapImageMediaType: "image/png",
	}

	req := campaignLLMRequest(llm.DefaultModelCampaign, "persona DM", "Genère le plan", resources)
	if req.Model != llm.DefaultModelCampaign {
		t.Errorf("model = %q, want Campaign role default", req.Model)
	}
	if len(req.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(req.Messages))
	}
	msg := req.Messages[0]
	if msg.Text != "Genère le plan" {
		t.Errorf("prompt text lost: %q", msg.Text)
	}
	if len(msg.Images) != 1 || msg.Images[0].Base64 != "cXVhdHJl" || msg.Images[0].MediaType != "image/png" {
		t.Fatalf("world-map image not attached: %+v", msg.Images)
	}
	if req.JSONResponse == nil || req.JSONResponse.Name != "campaign_plan" {
		t.Error("campaign plan must request structured JSON output")
	}
	if !req.RequireParameters {
		t.Error("structured-JSON campaign requests must require endpoint parameter support")
	}
	if req.MaxCompletionTokens != 8192 {
		t.Errorf("max completion tokens = %d, want 8192", req.MaxCompletionTokens)
	}
}

// TestCampaignLLMRequest_NoImage proves a missing map degrades to text-only.
func TestCampaignLLMRequest_NoImage(t *testing.T) {
	req := campaignLLMRequest(llm.DefaultModelCampaign, "persona", "prompt", nil)
	if len(req.Messages) != 1 || len(req.Messages[0].Images) != 0 {
		t.Fatalf("expected text-only message, got %+v", req.Messages)
	}
	if req.Messages[0].Text != "prompt" {
		t.Errorf("prompt lost: %q", req.Messages[0].Text)
	}
}

// TestGenerateAdventureName_FastRole proves title suggestions run on the Fast
// role through the shared client and sanitize the raw output.
func TestGenerateAdventureName_FastRole(t *testing.T) {
	fake := &fakeWebLLMClient{response: "\"Le Sextant Magique\"\n"}
	s := &Server{llmCfg: llm.DefaultModels(), llmClient: fake}

	name, err := s.generateAdventureName(tarot.CreativeBrief{Tone: "mysterious"}, "océan")
	if err != nil {
		t.Fatalf("generateAdventureName: %v", err)
	}
	if name != "Le Sextant Magique" {
		t.Errorf("sanitized name = %q", name)
	}
	if fake.request.Model != llm.DefaultModelFast {
		t.Errorf("model = %q, want Fast role default", fake.request.Model)
	}
	if fake.request.MaxCompletionTokens != 64 {
		t.Errorf("max completion tokens = %d, want 64", fake.request.MaxCompletionTokens)
	}
	if strings.Contains(fake.request.System, "titres d'aventures") == false {
		t.Error("title system prompt not carried")
	}
}

// TestGenerateAdventureName_ClientError proves provider failures propagate to
// the best-effort caller (which returns an empty name).
func TestGenerateAdventureName_ClientError(t *testing.T) {
	fake := &fakeWebLLMClient{err: fmt.Errorf("provider exploded")}
	s := &Server{llmCfg: llm.DefaultModels(), llmClient: fake}

	if _, err := s.generateAdventureName(tarot.CreativeBrief{}, ""); err == nil {
		t.Fatal("expected provider error to propagate")
	}
}