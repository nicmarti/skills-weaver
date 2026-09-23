package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"dungeons/internal/adventure"
	"dungeons/internal/llm"
)

// fakeEnrichClient scripts neutral responses for enrichment tests.
type fakeEnrichClient struct {
	mu        sync.Mutex
	responses []string
	calls     int
	requests  []llm.Request
}

func (f *fakeEnrichClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	snapshot := req
	f.requests = append(f.requests, snapshot)
	idx := f.calls
	f.calls++
	if idx >= len(f.responses) {
		return llm.Response{}, fmt.Errorf("fake exhausted")
	}
	return llm.Response{
		FinishReason: llm.FinishStop,
		Message:      llm.AssistantMessage(f.responses[idx]),
	}, nil
}

func (f *fakeEnrichClient) Stream(ctx context.Context, req llm.Request, obs llm.StreamObserver) (llm.Response, error) {
	return llm.Response{}, fmt.Errorf("not implemented")
}

func (f *fakeEnrichClient) lastRequest() llm.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests[len(f.requests)-1]
}

func newTestEnricher(responses ...string) (*Enricher, *fakeEnrichClient) {
	fake := &fakeEnrichClient{responses: responses}
	return &Enricher{client: fake, model: llm.DefaultModelFast}, fake
}

func testEnrichmentContext() *adventure.EnrichmentContext {
	return &adventure.EnrichmentContext{
		AdventureName: "Test Adventure",
		PartyMembers:  []string{"Aldric"},
		SessionInfo:   "Session 1",
	}
}

// TestEnrichEntry_JSONResponse proves the journal entry call requests
// structured JSON, requires endpoint parameter support, uses the Fast-role
// model, and parses the bilingual result.
func TestEnrichEntry_JSONResponse(t *testing.T) {
	enricher, fake := newTestEnricher(`{"description": "Aldric fights in a torch-lit corridor.", "description_fr": "Aldric combat dans un couloir éclairé aux torches."}`)

	result, err := enricher.EnrichEntry(adventure.JournalEntry{ID: 1, Type: "combat", Content: "Aldric attacks"}, testEnrichmentContext())
	if err != nil {
		t.Fatalf("EnrichEntry: %v", err)
	}
	if result.Description == "" || result.DescriptionFr == "" {
		t.Fatalf("incomplete result: %+v", result)
	}

	req := fake.lastRequest()
	if req.Model != llm.DefaultModelFast {
		t.Errorf("model = %q, want Fast role default", req.Model)
	}
	if req.JSONResponse == nil {
		t.Error("journal enrichment must request structured JSON output")
	}
	if !req.RequireParameters {
		t.Error("structured-JSON requests must require endpoint parameter support")
	}
}

// TestEnrichEntry_TolerantParsing proves markdown-fenced JSON still parses
// (fallback for endpoints without response-format support).
func TestEnrichEntry_TolerantParsing(t *testing.T) {
	enricher, _ := newTestEnricher("```json\n{\"description\": \"d\", \"description_fr\": \"f\"}\n```")

	result, err := enricher.EnrichEntry(adventure.JournalEntry{ID: 2, Type: "note", Content: "x"}, testEnrichmentContext())
	if err != nil {
		t.Fatalf("EnrichEntry with fenced JSON: %v", err)
	}
	if result.Description != "d" || result.DescriptionFr != "f" {
		t.Errorf("parsed fields = %+v", result)
	}
}

// TestEnrichEntry_RejectsIncomplete proves incomplete descriptions fail loudly.
func TestEnrichEntry_RejectsIncomplete(t *testing.T) {
	enricher, _ := newTestEnricher(`{"description": "only english"}`)

	_, err := enricher.EnrichEntry(adventure.JournalEntry{ID: 3, Type: "note", Content: "x"}, testEnrichmentContext())
	if err == nil || !strings.Contains(err.Error(), "incomplete descriptions") {
		t.Fatalf("expected incomplete-descriptions error, got %v", err)
	}
}

// TestEnrichMapPrompt_PlainTextNoJSON proves map-prompt enrichment stays
// free-form prose (no JSON response format) and passes word-count validation.
func TestEnrichMapPrompt_PlainTextNoJSON(t *testing.T) {
	longPrompt := strings.Repeat("Vue aérienne d'une cité portuaire aux toits de tuiles rouges, ruelles escarpées et marchés animés près des quais. ", 5)
	enricher, fake := newTestEnricher(longPrompt)

	result, err := enricher.EnrichMapPrompt(MapPromptRequest{MapType: "dungeon", LocationName: "La Crypte des Ombres", Scale: "medium"}, nil, nil)
	if err != nil {
		t.Fatalf("EnrichMapPrompt: %v", err)
	}
	if result.Prompt == "" {
		t.Fatal("empty enriched prompt")
	}
	if result.EnrichedAt == "" {
		t.Error("metadata not filled")
	}

	req := fake.lastRequest()
	if req.JSONResponse != nil {
		t.Error("map prompts are plain text — no JSON response format expected")
	}
	if req.RequireParameters {
		t.Error("plain-text requests must not require response-format support")
	}
}

// TestEnrichMapPrompt_TooShort proves the word-count guard rejects thin prompts.
func TestEnrichMapPrompt_TooShort(t *testing.T) {
	enricher, _ := newTestEnricher("Prompt trop court.")

	// Dungeon maps need no location record, so the call reaches enrichment.
	_, err := enricher.EnrichMapPrompt(MapPromptRequest{MapType: "dungeon", LocationName: "La Crypte des Ombres", Scale: "medium"}, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "too short") {
		t.Fatalf("expected too-short error, got %v", err)
	}
}

// TestNewEnricher_RequiresClient proves the dependency is explicit.
func TestNewEnricher_RequiresClient(t *testing.T) {
	if _, err := NewEnricher(nil, ""); err == nil {
		t.Fatal("nil client must be rejected")
	}
}