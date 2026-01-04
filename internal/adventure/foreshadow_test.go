package adventure

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPlantForeshadow(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	// Ensure directory exists
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Plant a foreshadow
	foreshadow, err := adv.PlantForeshadow(
		"Mysterious artifact mentioned",
		"Dark wizard spoke of an ancient relic",
		ImportanceMajor,
		CategoryArtifact,
		[]string{"artifact", "mystery"},
		[]string{"Dark Wizard"},
		[]string{"Ancient Tower"},
		nil,
	)

	if err != nil {
		t.Fatalf("PlantForeshadow failed: %v", err)
	}

	// Verify foreshadow was created
	if foreshadow.ID != "fsh_001" {
		t.Errorf("Expected ID fsh_001, got %s", foreshadow.ID)
	}
	if foreshadow.Description != "Mysterious artifact mentioned" {
		t.Errorf("Description mismatch")
	}
	if foreshadow.Importance != ImportanceMajor {
		t.Errorf("Importance mismatch")
	}
	if foreshadow.Category != CategoryArtifact {
		t.Errorf("Category mismatch")
	}
	if foreshadow.Status != ForeshadowActive {
		t.Errorf("Expected active status")
	}
	if len(foreshadow.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(foreshadow.Tags))
	}

	// Verify file was created
	path := filepath.Join(tmpDir, "foreshadows.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("foreshadows.json was not created")
	}

	// Load and verify
	history, err := adv.LoadForeshadows()
	if err != nil {
		t.Fatalf("LoadForeshadows failed: %v", err)
	}
	if len(history.Foreshadows) != 1 {
		t.Errorf("Expected 1 foreshadow, got %d", len(history.Foreshadows))
	}
	if history.NextID != 2 {
		t.Errorf("Expected NextID=2, got %d", history.NextID)
	}
}

func TestResolveForeshadow(t *testing.T) {
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Plant a foreshadow
	foreshadow, err := adv.PlantForeshadow(
		"Dark prophecy mentioned",
		"Seer spoke of doom",
		ImportanceCritical,
		CategoryProphecy,
		[]string{"prophecy"},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("PlantForeshadow failed: %v", err)
	}

	// Resolve it
	resolved, err := adv.ResolveForeshadow(foreshadow.ID, "Prophecy came true - dragon defeated")
	if err != nil {
		t.Fatalf("ResolveForeshadow failed: %v", err)
	}

	// Verify resolution
	if resolved.Status != ForeshadowResolved {
		t.Errorf("Expected resolved status, got %s", resolved.Status)
	}
	if resolved.ResolvedAt == nil {
		t.Errorf("ResolvedAt should not be nil")
	}
	if resolved.ResolutionNotes != "Prophecy came true - dragon defeated" {
		t.Errorf("Resolution notes mismatch")
	}

	// Verify persistence
	loaded, err := adv.GetForeshadow(foreshadow.ID)
	if err != nil {
		t.Fatalf("GetForeshadow failed: %v", err)
	}
	if loaded.Status != ForeshadowResolved {
		t.Errorf("Loaded foreshadow should be resolved")
	}
}

func TestAbandonForeshadow(t *testing.T) {
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Plant a foreshadow
	foreshadow, err := adv.PlantForeshadow(
		"Red herring clue",
		"False lead",
		ImportanceMinor,
		CategoryMystery,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("PlantForeshadow failed: %v", err)
	}

	// Abandon it
	abandoned, err := adv.AbandonForeshadow(foreshadow.ID, "Players took different path")
	if err != nil {
		t.Fatalf("AbandonForeshadow failed: %v", err)
	}

	if abandoned.Status != ForeshadowAbandoned {
		t.Errorf("Expected abandoned status")
	}
	if abandoned.ResolutionNotes != "Abandoned: Players took different path" {
		t.Errorf("Resolution notes mismatch: %s", abandoned.ResolutionNotes)
	}
}

func TestGetActiveForeshadows(t *testing.T) {
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Plant multiple foreshadows
	adv.PlantForeshadow("Foreshadow 1", "Context 1", ImportanceMajor, CategoryVillain, nil, nil, nil, nil)
	adv.PlantForeshadow("Foreshadow 2", "Context 2", ImportanceMinor, CategoryMystery, nil, nil, nil, nil)
	f3, _ := adv.PlantForeshadow("Foreshadow 3", "Context 3", ImportanceMajor, CategoryArtifact, nil, nil, nil, nil)

	// Resolve one
	adv.ResolveForeshadow(f3.ID, "Resolved")

	// Get active
	active, err := adv.GetActiveForeshadows()
	if err != nil {
		t.Fatalf("GetActiveForeshadows failed: %v", err)
	}

	if len(active) != 2 {
		t.Errorf("Expected 2 active foreshadows, got %d", len(active))
	}

	// Verify none are resolved
	for _, f := range active {
		if f.Status != ForeshadowActive {
			t.Errorf("Found non-active foreshadow in active list: %s", f.ID)
		}
	}
}

func TestGetStaleForeshadows(t *testing.T) {
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create session history
	history := &SessionHistory{
		Sessions: []Session{
			{ID: 1, StartedAt: time.Now(), Status: "completed"},
			{ID: 2, StartedAt: time.Now(), Status: "completed"},
			{ID: 3, StartedAt: time.Now(), Status: "completed"},
			{ID: 4, StartedAt: time.Now(), Status: "completed"},
			{ID: 5, StartedAt: time.Now(), Status: "active"},
		},
		CurrentSession: intPtr(5),
	}
	if err := adv.SaveSessions(history); err != nil {
		t.Fatalf("Failed to save sessions: %v", err)
	}

	// Plant foreshadows at different sessions
	fHistory := &ForeshadowHistory{
		Foreshadows: []Foreshadow{
			{
				ID:             "fsh_001",
				Description:    "Old foreshadow",
				PlantedSession: 1,
				Status:         ForeshadowActive,
				Importance:     ImportanceMajor,
				Category:       CategoryVillain,
			},
			{
				ID:             "fsh_002",
				Description:    "Recent foreshadow",
				PlantedSession: 4,
				Status:         ForeshadowActive,
				Importance:     ImportanceMinor,
				Category:       CategoryMystery,
			},
			{
				ID:             "fsh_003",
				Description:    "Very old foreshadow",
				PlantedSession: 1,
				Status:         ForeshadowActive,
				Importance:     ImportanceCritical,
				Category:       CategoryProphecy,
			},
		},
		NextID: 4,
	}
	if err := adv.SaveForeshadows(fHistory); err != nil {
		t.Fatalf("Failed to save foreshadows: %v", err)
	}

	// Get stale foreshadows (older than 3 sessions)
	stale, err := adv.GetStaleForeshadows(3)
	if err != nil {
		t.Fatalf("GetStaleForeshadows failed: %v", err)
	}

	// Current session is 5, so foreshadows from session 1 are 4 sessions old (>= 3)
	// foreshadow from session 4 is only 1 session old (< 3)
	if len(stale) != 2 {
		t.Errorf("Expected 2 stale foreshadows, got %d", len(stale))
		for _, f := range stale {
			t.Logf("Stale: %s (session %d)", f.ID, f.PlantedSession)
		}
	}
}

func TestGetForeshadowsByCategory(t *testing.T) {
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Plant various categories
	adv.PlantForeshadow("Villain clue", "Context", ImportanceMajor, CategoryVillain, nil, nil, nil, nil)
	adv.PlantForeshadow("Artifact hint", "Context", ImportanceMajor, CategoryArtifact, nil, nil, nil, nil)
	adv.PlantForeshadow("Another villain", "Context", ImportanceMinor, CategoryVillain, nil, nil, nil, nil)

	// Get villains
	villains, err := adv.GetForeshadowsByCategory(CategoryVillain)
	if err != nil {
		t.Fatalf("GetForeshadowsByCategory failed: %v", err)
	}

	if len(villains) != 2 {
		t.Errorf("Expected 2 villain foreshadows, got %d", len(villains))
	}

	// Verify all are villains
	for _, f := range villains {
		if f.Category != CategoryVillain {
			t.Errorf("Found non-villain foreshadow: %s", f.Category)
		}
	}
}

func TestGetForeshadowsByImportance(t *testing.T) {
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Plant various importance levels
	adv.PlantForeshadow("Critical 1", "Context", ImportanceCritical, CategoryVillain, nil, nil, nil, nil)
	adv.PlantForeshadow("Minor 1", "Context", ImportanceMinor, CategoryMystery, nil, nil, nil, nil)
	adv.PlantForeshadow("Critical 2", "Context", ImportanceCritical, CategoryArtifact, nil, nil, nil, nil)

	// Get critical
	critical, err := adv.GetForeshadowsByImportance(ImportanceCritical)
	if err != nil {
		t.Fatalf("GetForeshadowsByImportance failed: %v", err)
	}

	if len(critical) != 2 {
		t.Errorf("Expected 2 critical foreshadows, got %d", len(critical))
	}

	for _, f := range critical {
		if f.Importance != ImportanceCritical {
			t.Errorf("Found non-critical foreshadow: %s", f.Importance)
		}
	}
}

func TestUpdateForeshadow(t *testing.T) {
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Plant a foreshadow
	foreshadow, err := adv.PlantForeshadow(
		"Original description",
		"Original context",
		ImportanceMinor,
		CategoryMystery,
		[]string{"tag1"},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("PlantForeshadow failed: %v", err)
	}

	// Update it
	updates := map[string]interface{}{
		"description": "Updated description",
		"importance":  string(ImportanceMajor),
		"tags":        []string{"tag1", "tag2", "tag3"},
	}

	updated, err := adv.UpdateForeshadow(foreshadow.ID, updates)
	if err != nil {
		t.Fatalf("UpdateForeshadow failed: %v", err)
	}

	// Verify updates
	if updated.Description != "Updated description" {
		t.Errorf("Description not updated")
	}
	if updated.Importance != ImportanceMajor {
		t.Errorf("Importance not updated")
	}
	if len(updated.Tags) != 3 {
		t.Errorf("Tags not updated, expected 3 got %d", len(updated.Tags))
	}

	// Verify persistence
	loaded, err := adv.GetForeshadow(foreshadow.ID)
	if err != nil {
		t.Fatalf("GetForeshadow failed: %v", err)
	}
	if loaded.Description != "Updated description" {
		t.Errorf("Updates not persisted")
	}
}

func TestLoadForeshadowsEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	adv := New("Test Adventure", "Test description")
	adv.SetBasePath(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Load when file doesn't exist
	history, err := adv.LoadForeshadows()
	if err != nil {
		t.Fatalf("LoadForeshadows failed: %v", err)
	}

	if len(history.Foreshadows) != 0 {
		t.Errorf("Expected empty foreshadows, got %d", len(history.Foreshadows))
	}
	if history.NextID != 1 {
		t.Errorf("Expected NextID=1, got %d", history.NextID)
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}