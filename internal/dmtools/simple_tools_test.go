package dmtools

import (
	"testing"

	"dungeons/internal/adventure"
)

func TestAddGoldToolUpdatesSessionStats(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &adventure.Adventure{}
	adv.SetBasePath(tmpDir)

	// Start a session
	session, err := adv.StartSession()
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}
	sessionID := session.ID

	// Execute the add_gold tool with a positive amount
	tool := NewAddGoldTool(adv)
	result, err := tool.Execute(map[string]interface{}{
		"amount": float64(750),
		"reason": "Trésor trouvé",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	resultMap := result.(map[string]interface{})
	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true, got: %v", resultMap)
	}

	// End session so it's findable via GetSession
	adv.EndSession("test")

	// Verify sessions.json was updated with gold
	s, err := adv.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if s.GoldFound != 750 {
		t.Errorf("GoldFound = %d, want 750", s.GoldFound)
	}
}

func TestAddGoldToolNegativeAmountNotTracked(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &adventure.Adventure{}
	adv.SetBasePath(tmpDir)

	session, err := adv.StartSession()
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}
	sessionID := session.ID

	// First add gold, then remove some
	tool := NewAddGoldTool(adv)
	tool.Execute(map[string]interface{}{"amount": float64(1000)}) //nolint:errcheck
	tool.Execute(map[string]interface{}{"amount": float64(-200)}) //nolint:errcheck

	adv.EndSession("test")

	// Only the positive amount (1000) should be in GoldFound; the negative (-200) is a spend
	s, err := adv.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if s.GoldFound != 1000 {
		t.Errorf("GoldFound = %d, want 1000 (negative amounts should not be tracked)", s.GoldFound)
	}
}

func TestAddGoldToolNoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &adventure.Adventure{}
	adv.SetBasePath(tmpDir)

	// No session started — tool should still succeed (stats silently skipped)
	tool := NewAddGoldTool(adv)
	result, err := tool.Execute(map[string]interface{}{
		"amount": float64(500),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	resultMap := result.(map[string]interface{})
	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Errorf("expected success=true even without active session, got: %v", resultMap)
	}
}
