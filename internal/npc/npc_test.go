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

	if npc.Race != "human" {
		t.Errorf("Generate() race = %v, want %v", npc.Race, "human")
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
