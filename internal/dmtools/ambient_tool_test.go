package dmtools

import (
	"testing"

	"dungeons/internal/adventure"
)

func newAmbientToolWithAdv(t *testing.T) (*SetAmbientMusicTool, *adventure.Adventure) {
	t.Helper()
	adv := &adventure.Adventure{}
	adv.SetBasePath(t.TempDir())
	// No LLM client: the guardrail must short-circuit BEFORE any provider
	// call, so the suppressed path needs no network. The pass-through path
	// will fail at the client call, which is exactly how we detect the
	// guardrail let it through.
	return NewSetAmbientMusicTool(nil, "", adv), adv
}

func TestAmbientGuardrail_SuppressesCombatWhenNoCombat(t *testing.T) {
	tool, _ := newAmbientToolWithAdv(t)

	for _, mood := range []string{"combat", "danger"} {
		res, err := tool.Execute(map[string]interface{}{
			"scene_description": "intense fight",
			"mood":              mood,
		})
		if err != nil {
			t.Fatalf("Execute(%s): unexpected error %v", mood, err)
		}
		m := res.(map[string]interface{})
		if m["suppressed"] != true {
			t.Fatalf("mood=%s with no active combat should be suppressed, got %v", mood, m)
		}
		if _, hasPrompt := m["lyria_prompt"]; hasPrompt {
			t.Fatalf("suppressed cue must not carry a lyria_prompt, got %v", m)
		}
	}
}

func TestAmbientGuardrail_AllowsCombatWhenCombatActive(t *testing.T) {
	tool, adv := newAmbientToolWithAdv(t)
	if _, err := adv.StartCombat(); err != nil {
		t.Fatalf("StartCombat: %v", err)
	}

	res, err := tool.Execute(map[string]interface{}{
		"scene_description": "intense fight",
		"mood":              "combat",
	})
	if err != nil {
		t.Fatalf("Execute: unexpected error %v", err)
	}
	m := res.(map[string]interface{})
	// With combat active the guardrail must NOT suppress; it falls through to the
	// Claude call, which fails here (empty key) — proving the cue was not blocked.
	if m["suppressed"] == true {
		t.Fatal("mood=combat with active combat must not be suppressed by the guardrail")
	}
}

func TestAmbientGuardrail_NeverSuppressesNonCombatMoods(t *testing.T) {
	tool, _ := newAmbientToolWithAdv(t)

	res, err := tool.Execute(map[string]interface{}{
		"scene_description": "cozy tavern",
		"mood":              "tavern",
	})
	if err != nil {
		t.Fatalf("Execute: unexpected error %v", err)
	}
	m := res.(map[string]interface{})
	if m["suppressed"] == true {
		t.Fatal("non-combat mood must never be suppressed by the anti-spoil guardrail")
	}
}
