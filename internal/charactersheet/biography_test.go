package charactersheet

import (
	"context"
	"fmt"
	"testing"

	"dungeons/internal/character"
	"dungeons/internal/llm"
)

// fakeBioClient scripts neutral responses for biography tests.
type fakeBioClient struct {
	response string
	err      error
	request  llm.Request
}

func (f *fakeBioClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
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

func (f *fakeBioClient) Stream(ctx context.Context, req llm.Request, obs llm.StreamObserver) (llm.Response, error) {
	return llm.Response{}, fmt.Errorf("not implemented")
}

func testCharacter() *character.Character {
	return &character.Character{
		Name:      "Aldric",
		Species:   "human",
		Class:     "fighter",
		Level:     3,
		Abilities: character.AbilityScores{Strength: 16, Dexterity: 12, Constitution: 14, Intelligence: 10, Wisdom: 11, Charisma: 9},
	}
}

// TestBiography_AIOutput proves the AI path parses the structured response,
// uses the Fast-role model and requests JSON output.
func TestBiography_AIOutput(t *testing.T) {
	fake := &fakeBioClient{response: `{
		"origin": "Né dans les terres rudes du nord",
		"background": "Ancien soldat de la garde",
		"motivation": "Venger son village",
		"personality": "Rude mais loyal",
		"bond_name": "Capitaine Ilsa",
		"bond_description": "Son ancienne commandante",
		"bond_type": "person",
		"bond_sentiment": "ally",
		"secrets": ["Cache une cicatrice"]
	}`}
	gen := NewBiographyGenerator(fake, "")

	bio, err := gen.Generate(testCharacter(), "")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if bio.CharacterName != "Aldric" || bio.Origin == "" || len(bio.Bonds) != 1 || bio.Bonds[0].Name != "Capitaine Ilsa" {
		t.Fatalf("biography fields wrong: %+v", bio)
	}
	if len(bio.Secrets) != 1 {
		t.Errorf("secrets = %v", bio.Secrets)
	}

	if fake.request.Model != llm.DefaultModelFast {
		t.Errorf("model = %q, want Fast role default", fake.request.Model)
	}
	if fake.request.JSONResponse == nil {
		t.Error("biography generation must request structured JSON output")
	}
}

// TestBiography_TemplateFallbackWithoutClient proves a nil client keeps the
// generator functional with template biographies.
func TestBiography_TemplateFallbackWithoutClient(t *testing.T) {
	gen := NewBiographyGenerator(nil, "")

	bio, err := gen.Generate(testCharacter(), "")
	if err != nil {
		t.Fatalf("Generate (templates): %v", err)
	}
	if bio.Origin == "" || bio.Background == "" || bio.Personality == "" {
		t.Fatalf("template biography incomplete: %+v", bio)
	}
}

// TestBiography_TemplateFallbackOnClientError proves an AI failure degrades
// to templates instead of failing the whole generation.
func TestBiography_TemplateFallbackOnClientError(t *testing.T) {
	fake := &fakeBioClient{err: fmt.Errorf("provider exploded")}
	gen := NewBiographyGenerator(fake, "")

	bio, err := gen.Generate(testCharacter(), "")
	if err != nil {
		t.Fatalf("Generate must fall back to templates, got error: %v", err)
	}
	if bio.Origin == "" {
		t.Fatal("template fallback produced an empty biography")
	}
}
