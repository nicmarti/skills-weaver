package image

import (
	"strings"
	"testing"

	"dungeons/internal/npc"
)

// TestBuildNPCPromptRaceDescriptor verifies the race->visual-descriptor lookup
// is keyed by the canonical French race and tolerates the legacy English value.
func TestBuildNPCPromptRaceDescriptor(t *testing.T) {
	french := &npc.NPC{Name: "Dora", Race: "nain", Gender: "female", Occupation: "forgeronne"}
	if got := BuildNPCPrompt(french, StyleIllustrated); !strings.Contains(got, "dwarven, stocky") {
		t.Errorf("BuildNPCPrompt(nain) = %q, want it to contain \"dwarven, stocky\"", got)
	}

	legacy := &npc.NPC{Name: "Bram", Race: "dwarf", Gender: "male", Occupation: "forgeron"}
	if got := BuildNPCPrompt(legacy, StyleIllustrated); !strings.Contains(got, "dwarven, stocky") {
		t.Errorf("BuildNPCPrompt(legacy dwarf) = %q, want descriptor via normalization", got)
	}
}
