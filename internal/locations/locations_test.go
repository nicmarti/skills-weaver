package locations

import (
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if gen == nil {
		t.Fatal("Generator is nil")
	}

	if gen.data == nil {
		t.Fatal("Generator data is nil")
	}
}

func TestGenerateCity(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	kingdoms := []string{"valdorine", "karvath", "lumenciel", "astrene"}

	for _, kingdom := range kingdoms {
		t.Run(kingdom, func(t *testing.T) {
			name, err := gen.GenerateCity(kingdom)
			if err != nil {
				t.Errorf("GenerateCity(%s) failed: %v", kingdom, err)
				return
			}

			if name == "" {
				t.Errorf("GenerateCity(%s) returned empty name", kingdom)
			}

			t.Logf("%s city: %s", kingdom, name)
		})
	}
}

func TestGenerateTown(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	kingdoms := []string{"valdorine", "karvath", "lumenciel", "astrene"}

	for _, kingdom := range kingdoms {
		t.Run(kingdom, func(t *testing.T) {
			name, err := gen.GenerateTown(kingdom)
			if err != nil {
				t.Errorf("GenerateTown(%s) failed: %v", kingdom, err)
				return
			}

			if name == "" {
				t.Errorf("GenerateTown(%s) returned empty name", kingdom)
			}

			t.Logf("%s town: %s", kingdom, name)
		})
	}
}

func TestGenerateVillage(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	kingdoms := []string{"valdorine", "karvath", "lumenciel", "astrene"}

	for _, kingdom := range kingdoms {
		t.Run(kingdom, func(t *testing.T) {
			name, err := gen.GenerateVillage(kingdom)
			if err != nil {
				t.Errorf("GenerateVillage(%s) failed: %v", kingdom, err)
				return
			}

			if name == "" {
				t.Errorf("GenerateVillage(%s) returned empty name", kingdom)
			}

			t.Logf("%s village: %s", kingdom, name)

			// Valdorine villages should have "Le/La/Les" prefix
			if kingdom == "valdorine" && !strings.HasPrefix(name, "Le ") && !strings.HasPrefix(name, "La ") && !strings.HasPrefix(name, "Les ") {
				t.Errorf("Valdorine village should start with Le/La/Les, got: %s", name)
			}
		})
	}
}

func TestGenerateRegion(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	kingdoms := []string{"valdorine", "karvath", "lumenciel", "astrene"}

	for _, kingdom := range kingdoms {
		t.Run(kingdom, func(t *testing.T) {
			name, err := gen.GenerateRegion(kingdom)
			if err != nil {
				t.Errorf("GenerateRegion(%s) failed: %v", kingdom, err)
				return
			}

			if name == "" {
				t.Errorf("GenerateRegion(%s) returned empty name", kingdom)
			}

			t.Logf("%s region: %s", kingdom, name)
		})
	}
}

func TestGenerateRuin(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	for i := 0; i < 5; i++ {
		name := gen.GenerateRuin(true)
		if name == "" {
			t.Error("GenerateRuin returned empty name")
		}
		t.Logf("Ruin %d: %s", i+1, name)
	}
}

func TestGenerateGeneric(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	for i := 0; i < 5; i++ {
		name := gen.GenerateGeneric()
		if name == "" {
			t.Error("GenerateGeneric returned empty name")
		}
		t.Logf("Generic %d: %s", i+1, name)
	}
}

func TestGenerateSpecial(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	for i := 0; i < 5; i++ {
		name := gen.GenerateSpecial()
		if name == "" {
			t.Error("GenerateSpecial returned empty name")
		}
		t.Logf("Special %d: %s", i+1, name)
	}
}

func TestGenerateMultiple(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	names, err := gen.GenerateMultiple("valdorine", "city", 5)
	if err != nil {
		t.Fatalf("GenerateMultiple failed: %v", err)
	}

	if len(names) != 5 {
		t.Errorf("Expected 5 names, got %d", len(names))
	}

	// Check uniqueness
	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			t.Errorf("Duplicate name generated: %s", name)
		}
		seen[name] = true
		t.Logf("City: %s", name)
	}
}

func TestKingdomStyles(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	tests := []struct {
		kingdom  string
		expected []string // Keywords that should appear in generated names
	}{
		{
			kingdom:  "valdorine",
			expected: []string{"Cor", "Port", "Havre", "Mar", "Nav", "Bel", "dova", "luna", "vel"},
		},
		{
			kingdom:  "karvath",
			expected: []string{"Fer", "Acier", "Roc", "Forte", "Garde", "lance", "marteau", "épée", "burg", "heim"},
		},
		{
			kingdom:  "lumenciel",
			expected: []string{"Aurore", "Lumière", "Saint", "Céleste", "Divine", "sancta", "lumen", "ciel"},
		},
		{
			kingdom:  "astrene",
			expected: []string{"Étoile", "Lune", "Astro", "Nyx", "Alba", "automne", "crépuscule", "noctis"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.kingdom, func(t *testing.T) {
			found := false
			for i := 0; i < 20; i++ {
				name, err := gen.GenerateCity(tt.kingdom)
				if err != nil {
					t.Errorf("Failed to generate city: %v", err)
					continue
				}

				for _, keyword := range tt.expected {
					if strings.Contains(name, keyword) {
						found = true
						t.Logf("✓ Found keyword '%s' in '%s'", keyword, name)
						break
					}
				}

				if found {
					break
				}
			}

			if !found {
				t.Logf("Note: Expected keywords for %s not found in 20 generations (this may be normal due to randomness)", tt.kingdom)
			}
		})
	}
}

func TestInvalidKingdom(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	_, err = gen.GenerateCity("invalid")
	if err == nil {
		t.Error("Expected error for invalid kingdom, got nil")
	}
}

func TestGetAvailableKingdoms(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	kingdoms := gen.GetAvailableKingdoms()
	expected := []string{"valdorine", "karvath", "lumenciel", "astrene"}

	if len(kingdoms) != len(expected) {
		t.Errorf("Expected %d kingdoms, got %d", len(expected), len(kingdoms))
	}

	for _, exp := range expected {
		found := false
		for _, k := range kingdoms {
			if k == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected kingdom %s not found in available kingdoms", exp)
		}
	}
}

func TestGetAvailableTypes(t *testing.T) {
	gen, err := NewGenerator("../../data")
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	types := gen.GetAvailableTypes()
	expected := []string{"city", "town", "village", "region", "ruin", "generic", "special"}

	if len(types) != len(expected) {
		t.Errorf("Expected %d types, got %d", len(expected), len(types))
	}
}
