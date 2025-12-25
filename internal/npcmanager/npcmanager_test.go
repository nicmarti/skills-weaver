package npcmanager

import (
	"os"
	"path/filepath"
	"testing"

	"dungeons/internal/npc"
)

func TestManager_AddNPC(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "npcmanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// Create test NPC
	testNPC := &npc.NPC{
		Name:       "Test NPC",
		Race:       "human",
		Gender:     "male",
		Occupation: "merchant",
	}

	// Add NPC
	record, err := mgr.AddNPC(1, testNPC, "Tavern encounter", "Validated by world-keeper")
	if err != nil {
		t.Fatalf("AddNPC failed: %v", err)
	}

	if record.ID != "npc_001" {
		t.Errorf("Expected ID npc_001, got %s", record.ID)
	}

	if record.Importance != ImportanceMentioned {
		t.Errorf("Expected importance %s, got %s", ImportanceMentioned, record.Importance)
	}

	if record.Appearances != 1 {
		t.Errorf("Expected 1 appearance, got %d", record.Appearances)
	}

	// Verify file was created
	dbPath := filepath.Join(tmpDir, "npcs-generated.json")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("npcs-generated.json was not created")
	}

	// Load and verify
	db, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	session1 := db.Sessions["session_1"]
	if len(session1) != 1 {
		t.Errorf("Expected 1 NPC in session_1, got %d", len(session1))
	}

	if session1[0].NPC.Name != "Test NPC" {
		t.Errorf("Expected NPC name 'Test NPC', got %s", session1[0].NPC.Name)
	}
}

func TestManager_UpdateImportance(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "npcmanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// Add test NPC
	testNPC := &npc.NPC{
		Name:       "Test NPC",
		Race:       "human",
		Gender:     "male",
		Occupation: "merchant",
	}
	mgr.AddNPC(1, testNPC, "Initial encounter", "")

	// Update importance
	err = mgr.UpdateImportance("Test NPC", ImportanceInteracted, "Had dialogue about quest")
	if err != nil {
		t.Fatalf("UpdateImportance failed: %v", err)
	}

	// Verify update
	db, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	record := db.Sessions["session_1"][0]
	if record.Importance != ImportanceInteracted {
		t.Errorf("Expected importance %s, got %s", ImportanceInteracted, record.Importance)
	}

	if len(record.Notes) != 1 {
		t.Errorf("Expected 1 note, got %d", len(record.Notes))
	}

	if record.Notes[0] != "Had dialogue about quest" {
		t.Errorf("Expected note 'Had dialogue about quest', got %s", record.Notes[0])
	}

	if record.Appearances != 2 {
		t.Errorf("Expected 2 appearances, got %d", record.Appearances)
	}
}

func TestManager_GetNPCHistory(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "npcmanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// Add test NPC
	testNPC := &npc.NPC{
		Name:       "Test NPC",
		Race:       "human",
		Gender:     "male",
		Occupation: "merchant",
	}
	mgr.AddNPC(1, testNPC, "Initial encounter", "World-keeper note")

	// Get history
	record, err := mgr.GetNPCHistory("Test NPC")
	if err != nil {
		t.Fatalf("GetNPCHistory failed: %v", err)
	}

	if record.NPC.Name != "Test NPC" {
		t.Errorf("Expected NPC name 'Test NPC', got %s", record.NPC.Name)
	}

	if record.Context != "Initial encounter" {
		t.Errorf("Expected context 'Initial encounter', got %s", record.Context)
	}

	if record.WorldKeeperNotes != "World-keeper note" {
		t.Errorf("Expected world-keeper note 'World-keeper note', got %s", record.WorldKeeperNotes)
	}

	// Test non-existent NPC
	_, err = mgr.GetNPCHistory("Non-existent NPC")
	if err == nil {
		t.Error("Expected error for non-existent NPC, got nil")
	}
}

func TestManager_ListNPCsForReview(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "npcmanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// Add NPCs with different importance levels
	npc1 := &npc.NPC{Name: "NPC1", Race: "human", Gender: "male", Occupation: "merchant"}
	npc2 := &npc.NPC{Name: "NPC2", Race: "dwarf", Gender: "male", Occupation: "blacksmith"}
	npc3 := &npc.NPC{Name: "NPC3", Race: "elf", Gender: "female", Occupation: "mage"}

	mgr.AddNPC(1, npc1, "Context1", "")
	mgr.AddNPC(1, npc2, "Context2", "")
	mgr.AddNPC(1, npc3, "Context3", "")

	// Update importance
	mgr.UpdateImportance("NPC1", ImportanceMentioned, "")   // Should NOT appear in review
	mgr.UpdateImportance("NPC2", ImportanceInteracted, "") // Should appear
	mgr.UpdateImportance("NPC3", ImportanceRecurring, "")  // Should appear

	// Get review list
	candidates, err := mgr.ListNPCsForReview()
	if err != nil {
		t.Fatalf("ListNPCsForReview failed: %v", err)
	}

	if len(candidates) != 2 {
		t.Errorf("Expected 2 candidates, got %d", len(candidates))
	}

	// Verify NPCs are correct
	names := make(map[string]bool)
	for _, c := range candidates {
		names[c.NPC.Name] = true
	}

	if !names["NPC2"] || !names["NPC3"] {
		t.Error("Expected NPC2 and NPC3 in candidates")
	}

	if names["NPC1"] {
		t.Error("NPC1 should not be in candidates")
	}
}

func TestManager_MarkAsPromoted(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "npcmanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// Add test NPC
	testNPC := &npc.NPC{Name: "Test NPC", Race: "human", Gender: "male", Occupation: "merchant"}
	mgr.AddNPC(1, testNPC, "Context", "")

	// Mark as promoted
	err = mgr.MarkAsPromoted("Test NPC")
	if err != nil {
		t.Fatalf("MarkAsPromoted failed: %v", err)
	}

	// Verify
	db, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	record := db.Sessions["session_1"][0]
	if !record.PromotedToWorld {
		t.Error("Expected PromotedToWorld to be true")
	}

	// Verify it doesn't appear in review list anymore
	candidates, err := mgr.ListNPCsForReview()
	if err != nil {
		t.Fatalf("ListNPCsForReview failed: %v", err)
	}

	if len(candidates) != 0 {
		t.Errorf("Expected 0 candidates after promotion, got %d", len(candidates))
	}
}
