package npc

import (
	"strings"
	"testing"
)

// TestNPCDisplayNormalizesLegacyRace guards the "normalize on read" promise:
// NPCs persisted with the legacy English race ("human") must still render in
// French, and canonical French races render too.
func TestNPCDisplayNormalizesLegacyRace(t *testing.T) {
	legacy := &NPC{Name: "Bram", Race: "human", Gender: "male", Occupation: "forgeron"}
	if got := legacy.ToShortDescription(); !strings.Contains(got, "humain") {
		t.Errorf("ToShortDescription(legacy human) = %q, want it to contain \"humain\"", got)
	}
	if got := legacy.ToMarkdown(); !strings.Contains(got, "Humain") {
		t.Errorf("ToMarkdown(legacy human) = %q, want it to contain \"Humain\"", got)
	}

	french := &NPC{Name: "Dora", Race: string(RaceNain), Gender: "female", Occupation: "mineuse"}
	if got := french.ToShortDescription(); !strings.Contains(got, "nain") {
		t.Errorf("ToShortDescription(nain) = %q, want it to contain \"nain\"", got)
	}
	if got := french.ToMarkdown(); !strings.Contains(got, "Nain") {
		t.Errorf("ToMarkdown(nain) = %q, want it to contain \"Nain\"", got)
	}
}

// TestGenerateRaceFrenchAppearance verifies the French race switch actually
// applies the racial build (proves the canonical-French appearance path).
func TestGenerateRaceFrenchAppearance(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	n, err := gen.Generate(WithRace("nain"))
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if n.Race != string(RaceNain) {
		t.Errorf("race = %q, want %q", n.Race, RaceNain)
	}
	dwarfBuilds := map[string]bool{"trapu": true, "musclé": true, "robuste": true, "râblé": true}
	if !dwarfBuilds[n.Appearance.Build] {
		t.Errorf("nain Build = %q, want one of trapu/musclé/robuste/râblé", n.Appearance.Build)
	}
}
