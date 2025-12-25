package world

import (
	"testing"
)

func TestLoadGeography(t *testing.T) {
	geo, err := LoadGeography("../../data")
	if err != nil {
		t.Fatalf("Failed to load geography: %v", err)
	}

	if len(geo.Continents) == 0 {
		t.Error("Expected at least one continent")
	}

	// Check if Cordova exists
	loc, region, err := geo.GetLocationByName("Cordova")
	if err != nil {
		t.Errorf("Cordova should exist: %v", err)
	}

	if loc.Kingdom != "Valdorine" {
		t.Errorf("Cordova should belong to Valdorine, got %s", loc.Kingdom)
	}

	if region.Name != "Côte Occidentale" {
		t.Errorf("Cordova should be in Côte Occidentale, got %s", region.Name)
	}
}

func TestLoadFactions(t *testing.T) {
	factions, err := LoadFactions("../../data")
	if err != nil {
		t.Fatalf("Failed to load factions: %v", err)
	}

	if len(factions.Kingdoms) != 4 {
		t.Errorf("Expected 4 kingdoms, got %d", len(factions.Kingdoms))
	}

	// Check Valdorine exists
	valdorine, err := factions.GetKingdomByID("valdorine")
	if err != nil {
		t.Errorf("Valdorine should exist: %v", err)
	}

	if valdorine.Capital != "Cordova" {
		t.Errorf("Valdorine capital should be Cordova, got %s", valdorine.Capital)
	}
}

func TestGetLocationsByKingdom(t *testing.T) {
	geo, err := LoadGeography("../../data")
	if err != nil {
		t.Fatalf("Failed to load geography: %v", err)
	}

	valdorineLocs := geo.GetLocationsByKingdom("Valdorine")
	if len(valdorineLocs) == 0 {
		t.Error("Valdorine should have at least one location")
	}

	// Check all locations belong to Valdorine
	for _, loc := range valdorineLocs {
		if loc.Kingdom != "Valdorine" {
			t.Errorf("Location %s should belong to Valdorine, got %s", loc.Name, loc.Kingdom)
		}
	}
}

func TestGetSuggestions(t *testing.T) {
	geo, err := LoadGeography("../../data")
	if err != nil {
		t.Fatalf("Failed to load geography: %v", err)
	}

	// Test fuzzy matching
	suggestions := GetSuggestions("Cordov", geo, 5)
	if len(suggestions) == 0 {
		t.Error("Should return suggestions for 'Cordov'")
	}

	// First suggestion should be Cordova
	if suggestions[0].Location.Name != "Cordova" {
		t.Errorf("First suggestion should be Cordova, got %s", suggestions[0].Location.Name)
	}
}

func TestValidateLocationExists(t *testing.T) {
	geo, err := LoadGeography("../../data")
	if err != nil {
		t.Fatalf("Failed to load geography: %v", err)
	}

	// Existing location
	exists, loc, region, err := ValidateLocationExists("Cordova", geo)
	if !exists || err != nil {
		t.Error("Cordova should exist")
	}
	if loc == nil || region == nil {
		t.Error("Should return location and region for Cordova")
	}

	// Non-existing location
	exists, _, _, err = ValidateLocationExists("NonExistent", geo)
	if exists {
		t.Error("NonExistent should not exist")
	}
	if err == nil {
		t.Error("Should return error for non-existent location")
	}
}

func TestValidateKingdomExists(t *testing.T) {
	factions, err := LoadFactions("../../data")
	if err != nil {
		t.Fatalf("Failed to load factions: %v", err)
	}

	// Existing kingdom
	exists, kingdom, err := ValidateKingdomExists("valdorine", factions)
	if !exists || err != nil {
		t.Error("Valdorine should exist")
	}
	if kingdom == nil {
		t.Error("Should return kingdom for valdorine")
	}

	// Non-existing kingdom
	exists, _, err = ValidateKingdomExists("nonexistent", factions)
	if exists {
		t.Error("NonExistent kingdom should not exist")
	}
	if err == nil {
		t.Error("Should return error for non-existent kingdom")
	}
}
