package tarot

import (
	"strings"
	"testing"
)

// deckPath points at the repo's real deck so the tests also guard the JSON.
const deckPath = "../../data/tarot-deck.json"

func loadTestDeck(t *testing.T) *Deck {
	t.Helper()
	d, err := LoadDeck(deckPath)
	if err != nil {
		t.Fatalf("LoadDeck: %v", err)
	}
	return d
}

// TestAggregateDeterministic verifies that a fixed set of chosen cards always
// produces the exact same brief — the core "voyante is deterministic" guarantee.
func TestAggregateDeterministic(t *testing.T) {
	d := loadTestDeck(t)

	chosen := []string{"la_maison_dieu", "le_roi_deniers", "l_ermite", "la_mort", "le_pendu"}
	brief, err := d.Aggregate(chosen)
	if err != nil {
		t.Fatalf("Aggregate: %v", err)
	}

	// Exact match on the canonical token axes.
	tokens := map[string]string{
		"tone":           brief.Tone,
		"adventure_type": brief.AdventureType,
		"danger_level":   brief.DangerLevel,
		"stakes":         brief.Stakes,
		"pacing":         brief.Pacing,
	}
	wantTokens := map[string]string{
		"tone":           "grim",
		"adventure_type": "diplomacy", // from le_roi_deniers (Valdorine magnate)
		"danger_level":   "deadly",
		"stakes":         "kingdom",
		"pacing":         "slow_burn",
	}
	for k, got := range tokens {
		if got != wantTokens[k] {
			t.Errorf("%s = %q, want %q", k, got, wantTokens[k])
		}
	}

	// The lore-rich descriptive axes are checked by substring (they carry
	// concrete world-of-the-Four-Kingdoms context for the LLM).
	if !strings.Contains(brief.AntagonistNature, "Valdorine") {
		t.Errorf("antagonist_nature = %q, want it to reference Valdorine", brief.AntagonistNature)
	}
	if !strings.Contains(brief.Region, "Montagnes de Fer") {
		t.Errorf("region = %q, want it to reference Montagnes de Fer", brief.Region)
	}

	// Re-running must yield an identical brief.
	brief2, _ := d.Aggregate(chosen)
	if brief2 != brief {
		t.Errorf("Aggregate not deterministic: %+v vs %+v", brief2, brief)
	}
}

func TestAggregateUnknownCard(t *testing.T) {
	d := loadTestDeck(t)
	if _, err := d.Aggregate([]string{"does_not_exist"}); err == nil {
		t.Fatal("expected error for unknown card id")
	}
}

// TestDeckIntegrity guards against drift between questions, cards and art files.
func TestDeckIntegrity(t *testing.T) {
	d := loadTestDeck(t)

	cardByID := map[string]Card{}
	for _, c := range d.Cards {
		if c.ID == "" {
			t.Errorf("card with empty id: %+v", c)
		}
		if c.Art == "" {
			t.Errorf("card %q has empty art", c.ID)
		}
		if c.Name == "" {
			t.Errorf("card %q has empty name", c.ID)
		}
		if _, dup := cardByID[c.ID]; dup {
			t.Errorf("duplicate card id %q", c.ID)
		}
		cardByID[c.ID] = c
	}

	for _, q := range d.Questions {
		if len(q.CardIDs) < 3 {
			t.Errorf("question %q has only %d cards (need >=3 to draw 3)", q.ID, len(q.CardIDs))
		}
		for _, id := range q.CardIDs {
			c, ok := cardByID[id]
			if !ok {
				t.Errorf("question %q references unknown card %q", q.ID, id)
				continue
			}
			if c.Question != q.ID {
				t.Errorf("card %q claims question %q but is listed under %q", id, c.Question, q.ID)
			}
		}
	}
}

func TestPromptSection(t *testing.T) {
	d := loadTestDeck(t)
	brief, _ := d.Aggregate([]string{"le_soleil", "l_empereur"})
	brief.CoherenceText = "## World Coherence\nAvoid reusing: Kess, Sirène"

	out := brief.PromptSection()
	for _, want := range []string{"Ton : héroïque", "Karvath", "Avoid reusing"} {
		if !strings.Contains(out, want) {
			t.Errorf("PromptSection missing %q in:\n%s", want, out)
		}
	}

	if (CreativeBrief{}).PromptSection() != "" {
		t.Error("empty brief should render empty prompt section")
	}
}
