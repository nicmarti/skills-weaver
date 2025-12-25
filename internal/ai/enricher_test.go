package ai

import (
	"dungeons/internal/world"
	"strings"
	"testing"
)

func TestBuildBaseMapPrompt_City(t *testing.T) {
	e := &Enricher{model: "test"}

	loc := &world.Location{
		Name:        "Cordova",
		Type:        "port majeur",
		Kingdom:     "Valdorine",
		Description: "Ville portuaire scintillante",
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

	req := MapPromptRequest{
		MapType:  "city",
		Scale:    "medium",
		Features: []string{"Villa de Valorian"},
	}

	prompt, err := e.buildBaseMapPrompt(req, loc, kingdom)
	if err != nil {
		t.Fatalf("buildBaseMapPrompt failed: %v", err)
	}

	// Validate prompt contains key elements
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

func TestBuildBaseMapPrompt_Region(t *testing.T) {
	e := &Enricher{model: "test"}

	loc := &world.Location{
		Name:        "Cordova",
		Kingdom:     "Valdorine",
		Description: "Région côtière prospère",
	}

	kingdom := &world.Kingdom{
		ID:   "valdorine",
		Name: "Royaume de Valdorine",
	}

	req := MapPromptRequest{
		MapType: "region",
		Scale:   "large",
	}

	prompt, err := e.buildBaseMapPrompt(req, loc, kingdom)
	if err != nil {
		t.Fatalf("buildBaseMapPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "région") {
		t.Error("Prompt should contain 'région'")
	}

	if !strings.Contains(prompt, "Routes commerciales") {
		t.Error("Prompt should mention trade routes")
	}
}

func TestBuildBaseMapPrompt_Dungeon(t *testing.T) {
	e := &Enricher{model: "test"}

	req := MapPromptRequest{
		MapType:      "dungeon",
		LocationName: "La Crypte des Ombres",
		DungeonLevel: 1,
		Features:     []string{"Salle du trône", "Crypte"},
	}

	prompt, err := e.buildBaseMapPrompt(req, nil, nil)
	if err != nil {
		t.Fatalf("buildBaseMapPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "La Crypte des Ombres") {
		t.Error("Prompt should contain dungeon name")
	}

	if !strings.Contains(prompt, "Niveau 1") {
		t.Error("Prompt should mention level")
	}

	if !strings.Contains(prompt, "Salle du trône") {
		t.Error("Prompt should include specific features")
	}

	if !strings.Contains(prompt, "Grille") {
		t.Error("Prompt should mention grid")
	}
}

func TestBuildBaseMapPrompt_Tactical(t *testing.T) {
	e := &Enricher{model: "test"}

	req := MapPromptRequest{
		MapType:   "tactical",
		Terrain:   "forêt",
		SceneDesc: "Combat dans la forêt dense",
		Features:  []string{"Ruisseau", "Pont de bois"},
	}

	prompt, err := e.buildBaseMapPrompt(req, nil, nil)
	if err != nil {
		t.Fatalf("buildBaseMapPrompt failed: %v", err)
	}

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

func TestBuildBaseMapPrompt_InvalidType(t *testing.T) {
	e := &Enricher{model: "test"}

	req := MapPromptRequest{
		MapType: "invalid",
	}

	_, err := e.buildBaseMapPrompt(req, nil, nil)
	if err == nil {
		t.Error("Should return error for invalid map type")
	}

	if !strings.Contains(err.Error(), "unknown map type") {
		t.Errorf("Error should mention unknown type, got: %v", err)
	}
}

func TestBuildBaseMapPrompt_CityRequiresLocation(t *testing.T) {
	e := &Enricher{model: "test"}

	req := MapPromptRequest{
		MapType: "city",
	}

	_, err := e.buildBaseMapPrompt(req, nil, nil)
	if err == nil {
		t.Error("Should return error when location is nil for city map")
	}

	if !strings.Contains(err.Error(), "location required") {
		t.Errorf("Error should mention location required, got: %v", err)
	}
}

func TestGetStyleHints(t *testing.T) {
	e := &Enricher{model: "test"}

	// Test Valdorine kingdom
	kingdom := &world.Kingdom{
		ID: "valdorine",
	}

	req := MapPromptRequest{
		MapType: "city",
		Style:   "illustrated",
	}

	hints := e.getStyleHints(req, kingdom)

	if !strings.Contains(hints, "maritime") {
		t.Errorf("Valdorine hints should mention maritime, got: %s", hints)
	}

	if !strings.Contains(hints, "blue/gold") {
		t.Errorf("Valdorine hints should mention blue/gold colors, got: %s", hints)
	}

	if !strings.Contains(hints, "aerial view") {
		t.Errorf("City hints should mention aerial view, got: %s", hints)
	}

	// Test Karvath kingdom
	kingdom.ID = "karvath"
	hints = e.getStyleHints(req, kingdom)

	if !strings.Contains(hints, "militaristic") {
		t.Errorf("Karvath hints should mention militaristic, got: %s", hints)
	}

	// Test dark fantasy style
	req.Style = "dark_fantasy"
	hints = e.getStyleHints(req, kingdom)

	if !strings.Contains(hints, "dark atmosphere") {
		t.Errorf("Dark fantasy hints should mention dark atmosphere, got: %s", hints)
	}
}

func TestBuildMapEnrichmentPrompt(t *testing.T) {
	e := &Enricher{model: "test"}

	loc := &world.Location{
		Name:         "Cordova",
		Type:         "port majeur",
		Kingdom:      "Valdorine",
		Description:  "Ville portuaire",
		KeyLocations: []string{"Taverne"},
	}

	kingdom := &world.Kingdom{
		ID:     "valdorine",
		Name:   "Royaume de Valdorine",
		Colors: []string{"bleu", "or"},
		Symbol: "Trident doré",
		Values: []string{"Commerce", "Prospérité"},
	}

	req := MapPromptRequest{
		MapType: "city",
		Scale:   "medium",
		Style:   "illustrated",
	}

	basePrompt := "Cette carte montre Cordova..."

	enrichPrompt := e.buildMapEnrichmentPrompt(req, basePrompt, loc, kingdom)

	// Validate enrichment prompt structure
	if !strings.Contains(enrichPrompt, "MAP TYPE: city") {
		t.Error("Should contain map type")
	}

	if !strings.Contains(enrichPrompt, "Cordova") {
		t.Error("Should contain location name")
	}

	if !strings.Contains(enrichPrompt, "Royaume de Valdorine") {
		t.Error("Should contain kingdom name")
	}

	if !strings.Contains(enrichPrompt, "bleu") {
		t.Error("Should contain kingdom colors")
	}

	if !strings.Contains(enrichPrompt, "BASE PROMPT") {
		t.Error("Should include base prompt section")
	}

	if !strings.Contains(enrichPrompt, "GUIDELINES") {
		t.Error("Should include guidelines section")
	}

	if !strings.Contains(enrichPrompt, "100-200 words") {
		t.Error("Should specify target length")
	}

	if !strings.Contains(enrichPrompt, "French") {
		t.Error("Should specify French output")
	}
}
