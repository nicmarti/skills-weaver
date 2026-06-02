package npc

import (
	"testing"
)

func TestGenerateOccupationSpecific(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		name       string
		occupation string
		expected   string
	}{
		{
			name:       "specific occupation - innkeeper",
			occupation: "aubergiste",
			expected:   "aubergiste",
		},
		{
			name:       "specific occupation - merchant",
			occupation: "marchand",
			expected:   "marchand",
		},
		{
			name:       "specific occupation - priest",
			occupation: "prêtre",
			expected:   "prêtre",
		},
		{
			name:       "specific occupation - guard",
			occupation: "garde de ville",
			expected:   "garde de ville",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc, err := gen.Generate(WithOccupationType(tt.occupation))
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if npc.Occupation != tt.expected {
				t.Errorf("Generate() occupation = %v, want %v", npc.Occupation, tt.expected)
			}
		})
	}
}

func TestGenerateOccupationCategory(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		name       string
		category   string
		validOccupations []string
	}{
		{
			name:     "category - commoner",
			category: "commoner",
			validOccupations: []string{
				"fermier", "pêcheur", "bûcheron", "mineur", "berger", "meunier",
				"boulanger", "boucher", "tanneur", "tisserand", "potier", "charpentier",
				"maçon", "forgeron", "cordonnier", "tailleur", "aubergiste", "cuisinier",
				"serveur", "palefrenier", "porteur", "mendiant", "fossoyeur", "balayeur",
			},
		},
		{
			name:     "category - skilled",
			category: "skilled",
			validOccupations: []string{
				"marchand", "apothicaire", "herboriste", "guérisseur", "sage-femme",
				"scribe", "cartographe", "bibliothécaire", "tuteur", "musicien",
				"acteur", "jongleur", "acrobate", "artiste", "sculpteur", "orfèvre",
				"horloger", "armurier", "sellier", "navigateur", "ingénieur",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc, err := gen.Generate(WithOccupationType(tt.category))
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Check if occupation is in the valid list
			found := false
			for _, validOcc := range tt.validOccupations {
				if npc.Occupation == validOcc {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Generate() occupation = %v, not in category %v", npc.Occupation, tt.category)
			}
		})
	}
}

func TestGenerateNPCWithAllOptions(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	npc, err := gen.Generate(
		WithRace("human"),
		WithGender("f"),
		WithOccupationType("aubergiste"),
		WithAttitude("friendly"),
	)

	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Legacy English input "human" is normalized to the canonical French race.
	if npc.Race != string(RaceHumain) {
		t.Errorf("Generate() race = %v, want %v", npc.Race, RaceHumain)
	}

	if npc.Gender != "female" {
		t.Errorf("Generate() gender = %v, want %v", npc.Gender, "female")
	}

	if npc.Occupation != "aubergiste" {
		t.Errorf("Generate() occupation = %v, want %v", npc.Occupation, "aubergiste")
	}

	// Attitude should be from the positive pool
	validAttitudes := []string{
		"amical et serviable", "curieux et bavard", "respectueux et poli",
		"enthousiaste et accueillant", "protecteur et bienveillant",
		"admiratif et impressionné", "reconnaissant", "confiant",
	}

	found := false
	for _, att := range validAttitudes {
		if npc.Attitude == att {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Generate() attitude = %v, not in friendly pool", npc.Attitude)
	}
}

// TestFacialFeaturePoolFiltersFemaleFacialHair guards the gender-aware fix:
// female NPCs must never be offered facial-hair features (beard, moustache,
// goatee, sideburns), while other genders keep the full pool.
func TestFacialFeaturePoolFiltersFemaleFacialHair(t *testing.T) {
	pool := []string{
		"une barbe fournie",
		"une barbe soigneusement taillée",
		"une moustache imposante",
		"un bouc",
		"des favoris",
		"un nez aquilin",
		"une cicatrice au visage",
		"un grain de beauté",
	}

	female := facialFeaturePool(pool, "female")
	for _, f := range female {
		if isFacialHair(f) {
			t.Errorf("female pool must not contain facial hair, found %q", f)
		}
	}
	if len(female) != 3 {
		t.Errorf("expected 3 non-facial-hair features, got %d: %v", len(female), female)
	}

	if got := facialFeaturePool(pool, "male"); len(got) != len(pool) {
		t.Errorf("male pool should keep all %d features, got %d", len(pool), len(got))
	}
	if got := facialFeaturePool(pool, ""); len(got) != len(pool) {
		t.Errorf("unknown gender should keep the full pool, got %d", len(got))
	}

	// Degenerate pool (only facial hair) must fall back to the full list rather
	// than return an empty pool that would break randomChoice.
	allHair := []string{"une barbe fournie", "un bouc"}
	if got := facialFeaturePool(allHair, "female"); len(got) != len(allHair) {
		t.Errorf("all-facial-hair pool should fall back to full, got %v", got)
	}
}

// TestGenerateFemaleNPCHasNoFacialHair is an integration check on the generator:
// over many samples, a female NPC never gets facial hair.
func TestGenerateFemaleNPCHasNoFacialHair(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	for i := 0; i < 300; i++ {
		n, err := gen.Generate(WithGender("female"))
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if isFacialHair(n.Appearance.FacialFeature) {
			t.Fatalf("female NPC #%d has facial hair: %q", i, n.Appearance.FacialFeature)
		}
	}
}

// TestGenerateMaleNPCCanHaveFacialHair ensures the fix does not over-filter:
// males still draw from the full pool, so facial hair appears across samples.
func TestGenerateMaleNPCCanHaveFacialHair(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	found := false
	for i := 0; i < 300 && !found; i++ {
		n, err := gen.Generate(WithGender("male"))
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if isFacialHair(n.Appearance.FacialFeature) {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one male NPC with facial hair across 300 samples")
	}
}
