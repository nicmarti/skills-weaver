package campaign

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/agent"
	"dungeons/internal/llm"
)

// fakeStreamClient scripts llm.Client.Stream responses for unit tests: it
// never touches the network. It can stream text deltas, return a normalized
// provider error (timeout, 5xx, 429, mid-stream death...), or die after
// emitting partial text — mirroring how the real OpenRouter client surfaces
// failures.
type fakeStreamClient struct {
	deltas   []string    // text fragments streamed in order
	usage    *llm.Usage  // optional usage attached to the response
	model    string      // routed model reported in the response
	err      error       // error returned by Stream (before any delta)
	captured llm.Request // request captured for assertions
	blockCtx bool        // simulate a stuck provider honoring ctx cancellation
}

func (f *fakeStreamClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.Response{}, errors.New("Complete must not be used by campaign generation")
}

func (f *fakeStreamClient) Stream(ctx context.Context, req llm.Request, obs llm.StreamObserver) (llm.Response, error) {
	f.captured = req

	if f.blockCtx {
		<-ctx.Done()
		return llm.Response{}, &llm.Error{
			Kind:      llm.ErrKindTimeout,
			Message:   "request deadline exceeded",
			Retryable: true,
		}
	}
	if f.err != nil {
		return llm.Response{}, f.err
	}

	var text strings.Builder
	for _, d := range f.deltas {
		if obs != nil {
			obs.OnTextDelta(d)
		}
		text.WriteString(d)
	}

	resp := llm.Response{
		FinishReason: llm.FinishStop,
		Message:      llm.AssistantMessage(text.String()),
	}
	if f.model != "" {
		resp.Model = f.model
	}
	if f.usage != nil {
		resp.Usage = *f.usage
	}
	return resp, nil
}

// minimalPlanJSON is a small but structurally valid campaign plan payload.
const minimalPlanJSON = `{
  "version": "1.0.0",
  "metadata": {"campaign_title": "Le Sang des Rois", "theme": "trahison"},
  "narrative_structure": {"objective": "Empêcher la guerre", "hook": "Une lettre scellée", "acts": []}
}`

func testRequest() *Request {
	return &Request{
		Model:  llm.DefaultModelCampaign,
		System: "persona DM",
		Prompt: "Génère un plan de campagne",
	}
}

// recordingObs captures text deltas in arrival order.
type recordingObs struct{ deltas []string }

func (o *recordingObs) OnTextDelta(text string) { o.deltas = append(o.deltas, text) }

func TestGenerate_Success_StreamsParsesAndMeasures(t *testing.T) {
	usage := &llm.Usage{
		Present:          true,
		PromptTokens:     4200,
		CompletionTokens: 3100,
		TotalTokens:      7300,
		ReasoningTokens:  900,
		Cost:             0.024,
	}
	client := &fakeStreamClient{
		deltas: []string{`{"version":"1.0.0","metadata":{"campaign_title":"Le`, ` Sang des Rois"},"narrative_structure":{"objective":"Empêcher la guerre"}}`},
		usage:  usage,
		model:  "anthropic/claude-sonnet-5",
	}
	obs := &recordingObs{}

	result, err := Generate(context.Background(), client, testRequest(), obs)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if result.Plan.Metadata.CampaignTitle != "Le Sang des Rois" {
		t.Errorf("campaign title = %q", result.Plan.Metadata.CampaignTitle)
	}
	if result.Plan.NarrativeStructure.Objective != "Empêcher la guerre" {
		t.Errorf("objective = %q", result.Plan.NarrativeStructure.Objective)
	}
	if len(obs.deltas) != 2 {
		t.Errorf("observer deltas = %d, want 2 (streaming must forward deltas)", len(obs.deltas))
	}
	if result.Elapsed <= 0 {
		t.Error("elapsed must be measured")
	}
	if result.TimeToFirstDelta <= 0 || result.TimeToFirstDelta > result.Elapsed {
		t.Errorf("time to first delta = %v, elapsed = %v", result.TimeToFirstDelta, result.Elapsed)
	}
	if result.Response.Model != "anthropic/claude-sonnet-5" {
		t.Errorf("routed model = %q", result.Response.Model)
	}
	if !result.Response.Usage.Present || result.Response.Usage.Cost != 0.024 || result.Response.Usage.ReasoningTokens != 900 {
		t.Errorf("usage not carried: %+v", result.Response.Usage)
	}
	// The request must carry the campaign-plan JSON contract.
	if client.captured.Model != llm.DefaultModelCampaign {
		t.Errorf("request model = %q", client.captured.Model)
	}
	if client.captured.JSONResponse == nil || client.captured.JSONResponse.Name != "campaign_plan" {
		t.Error("request must ask for campaign_plan structured JSON")
	}
}

func TestGenerate_FencedJSONStillParses(t *testing.T) {
	client := &fakeStreamClient{
		deltas: []string{"```json\n" + minimalPlanJSON + "\n```"},
	}
	result, err := Generate(context.Background(), client, testRequest(), nil)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if result.Plan.Metadata.CampaignTitle != "Le Sang des Rois" {
		t.Errorf("campaign title = %q, want fenced JSON stripped", result.Plan.Metadata.CampaignTitle)
	}
}

// TestGenerate_Timeout simulates the production incident: the context
// deadline fires while the model is still generating (docs/issues-2026-09-24.md).
// The normalized timeout error must propagate unwrapped enough for callers
// to classify it, and FriendlyError must turn it into a French blocking
// message.
func TestGenerate_Timeout(t *testing.T) {
	client := &fakeStreamClient{err: &llm.Error{
		Kind:      llm.ErrKindTimeout,
		Message:   "request deadline exceeded",
		Retryable: true,
	}}

	_, err := Generate(context.Background(), client, testRequest(), nil)
	if err == nil {
		t.Fatal("timeout must fail generation")
	}
	var llmErr *llm.Error
	if !errors.As(err, &llmErr) || llmErr.Kind != llm.ErrKindTimeout {
		t.Fatalf("error must preserve the normalized timeout kind, got %v", err)
	}
	if !strings.Contains(err.Error(), "API call failed") {
		t.Errorf("error should carry generation context: %v", err)
	}

	msg := FriendlyError(err)
	for _, want := range []string{"temps imparti", "Aucune aventure"} {
		if !strings.Contains(msg, want) {
			t.Errorf("FriendlyError(timeout) = %q, want %q", msg, want)
		}
	}
}

// TestGenerate_RealContextDeadline proves a raw context deadline (the 300s
// budget expiring on a stuck provider) also fails generation instead of
// hanging or returning a truncated plan.
func TestGenerate_RealContextDeadline(t *testing.T) {
	client := &fakeStreamClient{blockCtx: true}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := Generate(ctx, client, testRequest(), nil)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("stuck provider past the deadline must fail generation")
		}
		// The real client normalizes context deadlines into an llm.Error
		// with kind=timeout (wrapContextError); the fake mirrors that.
		var llmErr *llm.Error
		if !errors.As(err, &llmErr) || llmErr.Kind != llm.ErrKindTimeout {
			t.Errorf("error should classify as timeout, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Generate did not honor the context deadline")
	}
}

func TestGenerate_ProviderErrors(t *testing.T) {
	cases := []struct {
		name      string
		err       *llm.Error
		wantInMsg string
	}{
		{
			name: "provider 5xx",
			err: &llm.Error{
				Kind:    llm.ErrKindProvider,
				Status:  502,
				Message: "API error (status 502): upstream provider failure",
			},
			wantInMsg: "pas pu générer le plan",
		},
		{
			name: "rate limit 429",
			err: &llm.Error{
				Kind:    llm.ErrKindRateLimit,
				Status:  429,
				Message: "rate limited",
			},
			wantInMsg: "Limite de requêtes",
		},
		{
			name: "auth 401",
			err: &llm.Error{
				Kind:    llm.ErrKindAuth,
				Status:  401,
				Message: "invalid api key",
			},
			wantInMsg: "OPENROUTER_API_KEY",
		},
		{
			name: "payment 402",
			err: &llm.Error{
				Kind:    llm.ErrKindPayment,
				Status:  402,
				Message: "insufficient credits",
			},
			wantInMsg: "Crédits OpenRouter insuffisants",
		},
		{
			name: "overloaded 529",
			err: &llm.Error{
				Kind:    llm.ErrKindOverloaded,
				Status:  529,
				Message: "provider overloaded",
			},
			wantInMsg: "surchargé",
		},
		{
			name: "mid-stream death",
			err: &llm.Error{
				Kind:    llm.ErrKindMidStream,
				Message: "stream terminated by a network error",
			},
			wantInMsg: "interrompue",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeStreamClient{err: tc.err}

			_, err := Generate(context.Background(), client, testRequest(), nil)
			if err == nil {
				t.Fatal("provider error must fail generation")
			}
			var llmErr *llm.Error
			if !errors.As(err, &llmErr) || llmErr.Kind != tc.err.Kind {
				t.Fatalf("normalized kind not preserved, got %v", err)
			}
			if msg := FriendlyError(err); !strings.Contains(msg, tc.wantInMsg) {
				t.Errorf("FriendlyError = %q, want %q", msg, tc.wantInMsg)
			}
		})
	}
}

// TestGenerate_MalformedOutput guards the blocking-error contract: a model
// answering prose or truncated JSON must never yield a plan.
func TestGenerate_MalformedOutput(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		client := &fakeStreamClient{deltas: []string{"Voici votre plan de campagne, mes amis ! Il commence par..."}}
		if _, err := Generate(context.Background(), client, testRequest(), nil); err == nil {
			t.Fatal("non-JSON output must fail")
		}
	})
	t.Run("truncated JSON", func(t *testing.T) {
		client := &fakeStreamClient{deltas: []string{`{"version":"1.0.0","metadata":{"campaign_title":"Interrom`}}
		if _, err := Generate(context.Background(), client, testRequest(), nil); err == nil {
			t.Fatal("truncated JSON must fail")
		}
	})
	t.Run("empty response", func(t *testing.T) {
		client := &fakeStreamClient{deltas: nil}
		if _, err := Generate(context.Background(), client, testRequest(), nil); err == nil {
			t.Fatal("empty response must fail")
		}
	})
	t.Run("nil client", func(t *testing.T) {
		if _, err := Generate(context.Background(), nil, testRequest(), nil); err == nil {
			t.Fatal("nil client must fail")
		}
	})
}

func TestParsePlanJSON(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		plan, err := ParsePlanJSON(minimalPlanJSON)
		if err != nil {
			t.Fatalf("plain JSON: %v", err)
		}
		if plan.Version != "1.0.0" {
			t.Errorf("version = %q", plan.Version)
		}
	})
	t.Run("fenced with language tag", func(t *testing.T) {
		_, err := ParsePlanJSON("```json\n" + minimalPlanJSON + "\n```")
		if err != nil {
			t.Fatalf("fenced JSON: %v", err)
		}
	})
	t.Run("fenced without language tag", func(t *testing.T) {
		_, err := ParsePlanJSON("```\n" + minimalPlanJSON + "\n```")
		if err != nil {
			t.Fatalf("fenced JSON without tag: %v", err)
		}
	})
	t.Run("garbage", func(t *testing.T) {
		if _, err := ParsePlanJSON("not json at all"); err == nil {
			t.Fatal("garbage must fail")
		}
	})
}

// TestRequest_LLMRequest_Multimodal proves the world-map image rides along
// the prompt, and that the JSON contract is applied.
func TestRequest_LLMRequest_Multimodal(t *testing.T) {
	req := &Request{
		Model:  "openai/gpt-test",
		System: "persona",
		Prompt: "le prompt",
		WorldResources: &agent.WorldResources{
			MapImageBase64:    "cXVhdHJl",
			MapImageMediaType: "image/png",
		},
	}
	llmReq := req.LLMRequest()

	if llmReq.Model != "openai/gpt-test" || llmReq.System != "persona" {
		t.Errorf("model/system lost: %q / %q", llmReq.Model, llmReq.System)
	}
	if len(llmReq.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(llmReq.Messages))
	}
	msg := llmReq.Messages[0]
	if msg.Text != "le prompt" {
		t.Errorf("prompt lost: %q", msg.Text)
	}
	if len(msg.Images) != 1 || msg.Images[0].Base64 != "cXVhdHJl" {
		t.Errorf("world map image not attached: %+v", msg.Images)
	}
	if !llmReq.RequireParameters {
		t.Error("RequireParameters must be set for structured JSON")
	}
	if llmReq.MaxCompletionTokens != MaxCompletionTokens {
		t.Errorf("max completion tokens = %d, want %d", llmReq.MaxCompletionTokens, MaxCompletionTokens)
	}
}

// TestRequest_LLMRequest_ReasoningEffort guards the 5-minute budget: the
// campaign request must constrain reasoning-model thinking (provider-default
// effort made Opus 5.5 exceed every deadline) and an explicit override wins.
func TestRequest_LLMRequest_ReasoningEffort(t *testing.T) {
	def := testRequest()
	if got := def.LLMRequest().ReasoningEffort; got != DefaultReasoningEffort {
		t.Errorf("default reasoning effort = %q, want %q", got, DefaultReasoningEffort)
	}

	over := testRequest()
	over.ReasoningEffort = "low"
	if got := over.LLMRequest().ReasoningEffort; got != "low" {
		t.Errorf("explicit reasoning effort = %q, want low", got)
	}
}

func TestRequest_LLMRequest_NoMapStaysTextOnly(t *testing.T) {
	req := testRequest()
	llmReq := req.LLMRequest()
	if len(llmReq.Messages) != 1 || len(llmReq.Messages[0].Images) != 0 {
		t.Errorf("expected text-only message, got %+v", llmReq.Messages)
	}
}

// TestBuildRequest_AssemblesWizardPrompt runs the real prompt assembly from
// the repository root (personas and world resources are read from relative
// paths, exactly like the running web server).
func TestBuildRequest_AssemblesWizardPrompt(t *testing.T) {
	t.Chdir("../..")

	adv := adventure.New("La Crypte des Ombres", "Une description")
	req, err := BuildRequest(adv, "Un culte s'infiltre à Cordova", "campaign", "mystery", nil)
	if err != nil {
		t.Fatalf("BuildRequest: %v", err)
	}

	if req.System == "" {
		t.Error("DM persona must be loaded into System")
	}
	if req.Duration.MaxSessions == 0 {
		t.Error("duration preset must be resolved")
	}
	for _, want := range []string{
		"**Nom de l'aventure** : La Crypte des Ombres",
		"Un culte s'infiltre à Cordova",
		"Génère un plan de campagne D&D 5e",
		// The foreshadow trame must be explicit: without a sub-schema models
		// invented incompatible shapes (docs/issues-2026-09-24.md).
		`"introduced_session"`,
		`"introduced_act"`,
		`"expected_payoff_act"`,
		`"linked_element"`,
		`"fsh_001"`,
	} {
		if !strings.Contains(req.Prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

// TestBuildRequest_ActsTemplateAdaptsToDuration guards the act structure
// against the selected duration preset.
func TestBuildRequest_ActsTemplateAdaptsToDuration(t *testing.T) {
	t.Chdir("../..")

	adv := adventure.New("Test", "")
	oneshot, err := BuildRequest(adv, "thème", "oneshot", "", nil)
	if err != nil {
		t.Fatalf("BuildRequest(oneshot): %v", err)
	}
	if !strings.Contains(oneshot.Prompt, `"number": 1`) {
		t.Error("oneshot prompt must contain the single-act template")
	}
	if strings.Contains(oneshot.Prompt, `"number": 2`) {
		t.Error("oneshot prompt must not contain multi-act templates")
	}

	campaignReq, err := BuildRequest(adv, "thème", "campaign", "", nil)
	if err != nil {
		t.Fatalf("BuildRequest(campaign): %v", err)
	}
	for _, n := range []string{`"number": 1`, `"number": 2`, `"number": 3`} {
		if !strings.Contains(campaignReq.Prompt, n) {
			t.Errorf("campaign prompt missing act template %s", n)
		}
	}
}
