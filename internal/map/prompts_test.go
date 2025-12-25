package mapgen

import (
	"dungeons/internal/world"
	"strings"
	"testing"
)

func TestBuildCityMapPrompt(t *testing.T) {
	// Create test data
	loc := &world.Location{
		Name:        "Cordova",
		Type:        "port majeur",
		Kingdom:     "Valdorine",
		Description: "Ville portuaire scintillante, commerce actif",
		KeyLocations: []string{
			"Taverne du Voile Écarlate",
			"Docks Marchands",
		},
	}

	kingdom := &world.Kingdom{
		ID:     "valdorine",
		Name:   "Royaume de Valdorine",
		Colors: []string{"bleu", "or"},
	}

	ctx := &MapContext{
		Location: loc,
		Kingdom:  kingdom,
		Scale:    "medium",
	}

	opts := PromptOptions{
		Features: []string{"Villa de Valorian"},
		Style:    "illustrated",
	}

	// Build prompt
	prompt := BuildCityMapPrompt(ctx, opts)

	// Verify prompt contains key elements
	if !strings.Contains(prompt, "Cordova") {
		t.Error("Prompt should contain city name")
	}

	if !strings.Contains(prompt, "portuaire") {
		t.Error("Prompt should mention port nature")
	}

	if !strings.Contains(prompt, "Taverne du Voile Écarlate") {
		t.Error("Prompt should include POIs from location")
	}

	if !strings.Contains(prompt, "Villa de Valorian") {
		t.Error("Prompt should include extra features")
	}

	if !strings.Contains(prompt, "valdorin maritime") {
		t.Error("Prompt should mention kingdom architectural style")
	}

	if len(prompt) < 100 {
		t.Errorf("Prompt seems too short (%d characters)", len(prompt))
	}
}

func TestBuildRegionalMapPrompt(t *testing.T) {
	region := &world.Region{
		Name:        "Côte Occidentale",
		Kingdom:     "Valdorine",
		Description: "Région côtière prospère",
		Cities: []world.Location{
			{Name: "Cordova"},
			{Name: "Port-de-Lune"},
		},
	}

	kingdom := &world.Kingdom{
		ID:   "valdorine",
		Name: "Royaume de Valdorine",
	}

	ctx := &MapContext{
		Region:  region,
		Kingdom: kingdom,
		Scale:   "large",
	}

	opts := PromptOptions{}

	prompt := BuildRegionalMapPrompt(ctx, opts)

	if !strings.Contains(prompt, "Côte Occidentale") {
		t.Error("Prompt should contain region name")
	}

	if !strings.Contains(prompt, "Cordova") {
		t.Error("Prompt should list cities")
	}

	if !strings.Contains(prompt, "Routes commerciales") {
		t.Error("Prompt should mention trade routes")
	}
}

func TestBuildDungeonMapPrompt(t *testing.T) {
	prompt := BuildDungeonMapPrompt("La Crypte des Ombres", 1, PromptOptions{
		Features: []string{"Salle du trône", "Crypte"},
	})

	if !strings.Contains(prompt, "La Crypte des Ombres") {
		t.Error("Prompt should contain dungeon name")
	}

	if !strings.Contains(prompt, "Niveau 1") {
		t.Error("Prompt should mention level")
	}

	if !strings.Contains(prompt, "Salle du trône") {
		t.Error("Prompt should include specific features")
	}

	if !strings.Contains(prompt, "Pièges") {
		t.Error("Prompt should mention traps")
	}

	if !strings.Contains(prompt, "Grille") {
		t.Error("Prompt should mention grid")
	}
}

func TestBuildTacticalMapPrompt(t *testing.T) {
	prompt := BuildTacticalMapPrompt("forêt", "Combat dans la forêt dense", PromptOptions{
		Features: []string{"Ruisseau", "Pont de bois"},
	})

	if !strings.Contains(prompt, "forêt") {
		t.Error("Prompt should mention terrain")
	}

	if !strings.Contains(prompt, "Arbres") {
		t.Error("Prompt should include forest-specific features")
	}

	if !strings.Contains(prompt, "Ruisseau") {
		t.Error("Prompt should include extra features")
	}

	if !strings.Contains(prompt, "Grille de combat") {
		t.Error("Prompt should mention combat grid")
	}
}

func TestGetKingdomArchitecturalStyle(t *testing.T) {
	style := getKingdomArchitecturalStyle("valdorine")
	if !strings.Contains(style, "maritime") {
		t.Errorf("Valdorine style should mention maritime, got: %s", style)
	}

	style = getKingdomArchitecturalStyle("karvath")
	if !strings.Contains(style, "militaire") {
		t.Errorf("Karvath style should mention military, got: %s", style)
	}
}

func TestGetTacticalTerrainFeatures(t *testing.T) {
	features := getTacticalTerrainFeatures("forest")
	if len(features) == 0 {
		t.Error("Should return features for forest terrain")
	}

	hasTreeFeature := false
	for _, f := range features {
		if strings.Contains(strings.ToLower(f), "arbre") {
			hasTreeFeature = true
			break
		}
	}
	if !hasTreeFeature {
		t.Error("Forest features should include trees")
	}
}
