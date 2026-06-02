package agent

import (
	"strings"
	"testing"
)

// TestLoadMapDescriptionFrench guards the geography source change: the map
// description fed to the LLMs must be the FRENCH layout/distances section, not
// the English image-generation directives.
func TestLoadMapDescriptionFrench(t *testing.T) {
	desc, err := loadMapDescription("../../ai/world-map-prompt.md")
	if err != nil {
		t.Fatalf("loadMapDescription: %v", err)
	}

	// Must contain the French geography layout section.
	if !strings.Contains(desc, "Schema de disposition") {
		t.Errorf("expected the French geography layout section in the description")
	}

	// Must NOT leak the English image-generation directives.
	for _, leaked := range []string{"300 DPI", "raised bump", "color wash", "STYLE DIRECTIVES"} {
		if strings.Contains(desc, leaked) {
			t.Errorf("map description leaked English image directive %q", leaked)
		}
	}
}
