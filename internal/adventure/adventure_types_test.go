package adventure

import (
	"testing"
)

func TestGetAdventureTypes(t *testing.T) {
	types := GetAdventureTypes()
	if len(types) != 9 {
		t.Errorf("expected 9 adventure types, got %d", len(types))
	}

	// Check all IDs are unique
	ids := map[string]bool{}
	for _, at := range types {
		if ids[at.ID] {
			t.Errorf("duplicate adventure type ID: %s", at.ID)
		}
		ids[at.ID] = true

		if at.NameFR == "" {
			t.Errorf("adventure type %s has empty NameFR", at.ID)
		}
		if at.PromptGuide == "" {
			t.Errorf("adventure type %s has empty PromptGuide", at.ID)
		}
	}
}

func TestGetAdventureType(t *testing.T) {
	at := GetAdventureType("escort")
	if at == nil {
		t.Fatal("expected escort type, got nil")
	}
	if at.NameFR != "Mission d'escorte" {
		t.Errorf("expected 'Mission d'escorte', got '%s'", at.NameFR)
	}

	// Unknown type
	at = GetAdventureType("unknown")
	if at != nil {
		t.Errorf("expected nil for unknown type, got %v", at)
	}
}

func TestGetAdventureDurations(t *testing.T) {
	durations := GetAdventureDurations()
	if len(durations) != 3 {
		t.Errorf("expected 3 durations, got %d", len(durations))
	}

	// Verify oneshot
	oneshot := GetDuration("oneshot")
	if oneshot.Acts != 1 || oneshot.MaxSessions != 2 {
		t.Errorf("oneshot: expected 1 act, 2 max sessions, got %d acts, %d max",
			oneshot.Acts, oneshot.MaxSessions)
	}

	// Verify short (default)
	short := GetDuration("short")
	if short.Acts != 1 || short.MaxSessions != 5 {
		t.Errorf("short: expected 1 act, 5 max sessions, got %d acts, %d max",
			short.Acts, short.MaxSessions)
	}

	// Verify campaign
	campaign := GetDuration("campaign")
	if campaign.Acts != 3 || campaign.MaxSessions != 12 {
		t.Errorf("campaign: expected 3 acts, 12 max sessions, got %d acts, %d max",
			campaign.Acts, campaign.MaxSessions)
	}

	// Unknown defaults to short
	unknown := GetDuration("unknown")
	if unknown.ID != "short" {
		t.Errorf("expected default to 'short', got '%s'", unknown.ID)
	}
}

func TestGetAntiCultConstraints(t *testing.T) {
	constraints := GetAntiCultConstraints()
	if constraints == "" {
		t.Error("expected non-empty anti-cult constraints")
	}
}
