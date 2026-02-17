package adventure

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadCampaignPlan_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &Adventure{basePath: tmpDir}

	plan, err := adv.LoadCampaignPlan()
	if err != nil {
		t.Fatalf("expected no error when file doesn't exist, got: %v", err)
	}
	if plan != nil {
		t.Fatalf("expected nil plan, got: %v", plan)
	}
}

func TestSaveAndLoadCampaignPlan(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &Adventure{basePath: tmpDir}

	// Create a campaign plan
	plan := &CampaignPlan{
		Version: "1.0.0",
		Metadata: CampaignMetadata{
			CampaignTitle: "Test Campaign",
			Theme:         "Epic Adventure",
			CreatedAt:     time.Now(),
			GeneratedBy:   "test",
		},
		NarrativeStructure: NarrativeStructure{
			Objective: "Save the world",
			Hook:      "A mysterious stranger arrives",
			Acts: []Act{
				{
					Number:      1,
					Title:       "Act I",
					Description: "The Beginning",
					Status:      "in_progress",
					TargetSessions: []int{1, 2, 3},
				},
			},
		},
		Progression: Progression{
			CurrentAct:     1,
			CurrentSession: 1,
		},
		Foreshadows: ForeshadowsContainer{
			Active:   []ForeshadowLinked{},
			Resolved: []ForeshadowLinked{},
			NextID:   1,
		},
	}

	// Save
	if err := adv.SaveCampaignPlan(plan); err != nil {
		t.Fatalf("SaveCampaignPlan failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, "campaign-plan.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("campaign-plan.json was not created")
	}

	// Load
	loaded, err := adv.LoadCampaignPlan()
	if err != nil {
		t.Fatalf("LoadCampaignPlan failed: %v", err)
	}
	if loaded == nil {
		t.Fatalf("expected non-nil plan")
	}

	// Verify data
	if loaded.Metadata.CampaignTitle != "Test Campaign" {
		t.Errorf("expected CampaignTitle 'Test Campaign', got '%s'", loaded.Metadata.CampaignTitle)
	}
	if loaded.NarrativeStructure.Objective != "Save the world" {
		t.Errorf("expected Objective 'Save the world', got '%s'", loaded.NarrativeStructure.Objective)
	}
	if len(loaded.NarrativeStructure.Acts) != 1 {
		t.Errorf("expected 1 act, got %d", len(loaded.NarrativeStructure.Acts))
	}
}

func TestGetCurrentAct(t *testing.T) {
	plan := &CampaignPlan{
		NarrativeStructure: NarrativeStructure{
			Acts: []Act{
				{Number: 1, Title: "Act I", Status: "completed"},
				{Number: 2, Title: "Act II", Status: "in_progress"},
				{Number: 3, Title: "Act III", Status: "pending"},
			},
		},
		Progression: Progression{
			CurrentAct: 2,
		},
	}

	currentAct := plan.GetCurrentAct()
	if currentAct == nil {
		t.Fatalf("expected non-nil act")
	}
	if currentAct.Number != 2 {
		t.Errorf("expected act 2, got %d", currentAct.Number)
	}
	if currentAct.Title != "Act II" {
		t.Errorf("expected 'Act II', got '%s'", currentAct.Title)
	}
}

func TestGetCriticalForeshadows(t *testing.T) {
	plan := &CampaignPlan{
		Progression: Progression{
			CurrentSession: 10,
		},
		Foreshadows: ForeshadowsContainer{
			Active: []ForeshadowLinked{
				{
					ID:             "fsh_001",
					Description:    "Ancient prophecy",
					PlantedSession: 2,
					Importance:     ImportanceCritical,
				},
				{
					ID:             "fsh_002",
					Description:    "Minor clue",
					PlantedSession: 2,
					Importance:     ImportanceMinor,
				},
				{
					ID:             "fsh_003",
					Description:    "Recent major event",
					PlantedSession: 9,
					Importance:     ImportanceMajor,
				},
				{
					ID:             "fsh_004",
					Description:    "Old major event",
					PlantedSession: 5,
					Importance:     ImportanceMajor,
				},
			},
		},
	}

	critical := plan.GetCriticalForeshadows()

	// Should return fsh_001 (critical, age=8) and fsh_004 (major, age=5)
	// Should NOT return fsh_002 (minor) or fsh_003 (age=1)
	if len(critical) != 2 {
		t.Errorf("expected 2 critical foreshadows, got %d", len(critical))
	}

	foundCritical := false
	foundMajor := false
	for _, f := range critical {
		if f.ID == "fsh_001" {
			foundCritical = true
		}
		if f.ID == "fsh_004" {
			foundMajor = true
		}
	}

	if !foundCritical {
		t.Errorf("expected fsh_001 (critical) in results")
	}
	if !foundMajor {
		t.Errorf("expected fsh_004 (major, old) in results")
	}
}

func TestAdvanceAct(t *testing.T) {
	plan := &CampaignPlan{
		NarrativeStructure: NarrativeStructure{
			Acts: []Act{
				{Number: 1, Title: "Act I", Status: "in_progress"},
				{Number: 2, Title: "Act II", Status: "pending"},
				{Number: 3, Title: "Act III", Status: "pending"},
			},
		},
		Progression: Progression{
			CurrentAct: 1,
		},
	}

	// Advance to Act 2
	err := plan.AdvanceAct(2)
	if err != nil {
		t.Fatalf("AdvanceAct failed: %v", err)
	}

	// Verify Act 1 is completed
	if plan.NarrativeStructure.Acts[0].Status != "completed" {
		t.Errorf("expected Act 1 status 'completed', got '%s'", plan.NarrativeStructure.Acts[0].Status)
	}
	if plan.NarrativeStructure.Acts[0].CompletionCriteria.CompletedAt == nil {
		t.Errorf("expected CompletedAt to be set")
	}

	// Verify Act 2 is in progress
	if plan.NarrativeStructure.Acts[1].Status != "in_progress" {
		t.Errorf("expected Act 2 status 'in_progress', got '%s'", plan.NarrativeStructure.Acts[1].Status)
	}

	// Verify current act is updated
	if plan.Progression.CurrentAct != 2 {
		t.Errorf("expected CurrentAct 2, got %d", plan.Progression.CurrentAct)
	}
}

func TestCompletePlotPoint(t *testing.T) {
	plan := &CampaignPlan{
		Progression: Progression{
			CompletedPlotPoints: []string{"plot_001"},
		},
	}

	// Complete a new plot point
	err := plan.CompletePlotPoint("plot_002")
	if err != nil {
		t.Fatalf("CompletePlotPoint failed: %v", err)
	}

	if len(plan.Progression.CompletedPlotPoints) != 2 {
		t.Errorf("expected 2 completed plot points, got %d", len(plan.Progression.CompletedPlotPoints))
	}

	// Try to complete the same plot point again
	err = plan.CompletePlotPoint("plot_002")
	if err == nil {
		t.Errorf("expected error when completing duplicate plot point")
	}
}

func TestAddRemoveActiveThread(t *testing.T) {
	plan := &CampaignPlan{
		Progression: Progression{
			ActiveThreads: []string{"thread_001"},
		},
	}

	// Add new thread
	err := plan.AddActiveThread("thread_002")
	if err != nil {
		t.Fatalf("AddActiveThread failed: %v", err)
	}

	if len(plan.Progression.ActiveThreads) != 2 {
		t.Errorf("expected 2 active threads, got %d", len(plan.Progression.ActiveThreads))
	}

	// Try to add duplicate
	err = plan.AddActiveThread("thread_002")
	if err == nil {
		t.Errorf("expected error when adding duplicate thread")
	}

	// Remove thread
	err = plan.RemoveActiveThread("thread_001")
	if err != nil {
		t.Fatalf("RemoveActiveThread failed: %v", err)
	}

	if len(plan.Progression.ActiveThreads) != 1 {
		t.Errorf("expected 1 active thread, got %d", len(plan.Progression.ActiveThreads))
	}

	// Try to remove non-existent thread
	err = plan.RemoveActiveThread("thread_999")
	if err == nil {
		t.Errorf("expected error when removing non-existent thread")
	}
}

func TestPlantForeshadowLinked(t *testing.T) {
	plan := &CampaignPlan{
		Progression: Progression{
			CurrentSession:     5,
			PendingResolutions: []string{},
		},
		Foreshadows: ForeshadowsContainer{
			Active: []ForeshadowLinked{},
			NextID: 1,
		},
	}

	foreshadow := &ForeshadowLinked{
		Description:         "A dark prophecy",
		Importance:          ImportanceCritical,
		Category:            CategoryProphecy,
		LinkedToAct:         2,
		TargetPayoffSession: 10,
		PayoffType:          "revelation",
	}

	err := plan.PlantForeshadowLinked(foreshadow)
	if err != nil {
		t.Fatalf("PlantForeshadowLinked failed: %v", err)
	}

	// Verify ID was generated
	if foreshadow.ID != "fsh_001" {
		t.Errorf("expected ID 'fsh_001', got '%s'", foreshadow.ID)
	}

	// Verify NextID incremented
	if plan.Foreshadows.NextID != 2 {
		t.Errorf("expected NextID 2, got %d", plan.Foreshadows.NextID)
	}

	// Verify added to active list
	if len(plan.Foreshadows.Active) != 1 {
		t.Errorf("expected 1 active foreshadow, got %d", len(plan.Foreshadows.Active))
	}

	// Verify added to pending resolutions
	if len(plan.Progression.PendingResolutions) != 1 {
		t.Errorf("expected 1 pending resolution, got %d", len(plan.Progression.PendingResolutions))
	}

	// Verify session was set
	if plan.Foreshadows.Active[0].PlantedSession != 5 {
		t.Errorf("expected PlantedSession 5, got %d", plan.Foreshadows.Active[0].PlantedSession)
	}
}

func TestResolveForeshadowLinked(t *testing.T) {
	plan := &CampaignPlan{
		Foreshadows: ForeshadowsContainer{
			Active: []ForeshadowLinked{
				{
					ID:          "fsh_001",
					Description: "Test foreshadow",
				},
			},
			Resolved: []ForeshadowLinked{},
		},
		Progression: Progression{
			PendingResolutions: []string{"fsh_001"},
		},
	}

	err := plan.ResolveForeshadowLinked("fsh_001", "The prophecy came true")
	if err != nil {
		t.Fatalf("ResolveForeshadowLinked failed: %v", err)
	}

	// Verify moved from active to resolved
	if len(plan.Foreshadows.Active) != 0 {
		t.Errorf("expected 0 active foreshadows, got %d", len(plan.Foreshadows.Active))
	}
	if len(plan.Foreshadows.Resolved) != 1 {
		t.Errorf("expected 1 resolved foreshadow, got %d", len(plan.Foreshadows.Resolved))
	}

	// Verify resolution notes
	if plan.Foreshadows.Resolved[0].ResolutionNotes != "The prophecy came true" {
		t.Errorf("expected resolution notes, got '%s'", plan.Foreshadows.Resolved[0].ResolutionNotes)
	}

	// Verify removed from pending resolutions
	if len(plan.Progression.PendingResolutions) != 0 {
		t.Errorf("expected 0 pending resolutions, got %d", len(plan.Progression.PendingResolutions))
	}

	// Verify resolved time set
	if plan.Foreshadows.Resolved[0].ResolvedAt == nil {
		t.Errorf("expected ResolvedAt to be set")
	}
}

func TestAbandonForeshadowLinked(t *testing.T) {
	plan := &CampaignPlan{
		Foreshadows: ForeshadowsContainer{
			Active: []ForeshadowLinked{
				{
					ID:          "fsh_001",
					Description: "Test foreshadow",
				},
			},
			Abandoned: []ForeshadowLinked{},
		},
		Progression: Progression{
			PendingResolutions: []string{"fsh_001"},
		},
	}

	err := plan.AbandonForeshadowLinked("fsh_001", "No longer relevant")
	if err != nil {
		t.Fatalf("AbandonForeshadowLinked failed: %v", err)
	}

	// Verify moved from active to abandoned
	if len(plan.Foreshadows.Active) != 0 {
		t.Errorf("expected 0 active foreshadows, got %d", len(plan.Foreshadows.Active))
	}
	if len(plan.Foreshadows.Abandoned) != 1 {
		t.Errorf("expected 1 abandoned foreshadow, got %d", len(plan.Foreshadows.Abandoned))
	}

	// Verify reason in notes
	if plan.Foreshadows.Abandoned[0].ResolutionNotes != "Abandoned: No longer relevant" {
		t.Errorf("unexpected resolution notes: '%s'", plan.Foreshadows.Abandoned[0].ResolutionNotes)
	}
}

func TestGetForeshadow(t *testing.T) {
	plan := &CampaignPlan{
		Foreshadows: ForeshadowsContainer{
			Active: []ForeshadowLinked{
				{ID: "fsh_001", Description: "Active"},
			},
			Resolved: []ForeshadowLinked{
				{ID: "fsh_002", Description: "Resolved"},
			},
			Abandoned: []ForeshadowLinked{
				{ID: "fsh_003", Description: "Abandoned"},
			},
		},
	}

	// Test finding active
	f := plan.GetForeshadow("fsh_001")
	if f == nil || f.Description != "Active" {
		t.Errorf("failed to find active foreshadow")
	}

	// Test finding resolved
	f = plan.GetForeshadow("fsh_002")
	if f == nil || f.Description != "Resolved" {
		t.Errorf("failed to find resolved foreshadow")
	}

	// Test finding abandoned
	f = plan.GetForeshadow("fsh_003")
	if f == nil || f.Description != "Abandoned" {
		t.Errorf("failed to find abandoned foreshadow")
	}

	// Test not found
	f = plan.GetForeshadow("fsh_999")
	if f != nil {
		t.Errorf("expected nil for non-existent foreshadow")
	}
}

func TestUpdatePacing(t *testing.T) {
	plan := &CampaignPlan{
		Metadata: CampaignMetadata{
			TargetDuration: TargetDuration{
				Sessions:        12,
				HoursPerSession: 3,
			},
		},
		NarrativeStructure: NarrativeStructure{
			Acts: []Act{
				{
					Number:         1,
					Status:         "completed",
					TargetSessions: []int{1, 2, 3, 4},
				},
				{
					Number:         2,
					Status:         "in_progress",
					TargetSessions: []int{5, 6, 7, 8},
				},
				{
					Number:         3,
					Status:         "pending",
					TargetSessions: []int{9, 10, 11, 12},
				},
			},
		},
		Progression: Progression{
			CurrentAct:     2,
			CurrentSession: 6,
		},
		Pacing: Pacing{},
	}

	plan.UpdatePacing()

	// Verify sessions played
	if plan.Pacing.SessionsPlayed != 6 {
		t.Errorf("expected SessionsPlayed 6, got %d", plan.Pacing.SessionsPlayed)
	}

	// Verify sessions remaining
	if plan.Pacing.SessionsRemainingEstimate != 6 {
		t.Errorf("expected SessionsRemainingEstimate 6, got %d", plan.Pacing.SessionsRemainingEstimate)
	}

	// Verify act breakdown
	if plan.Pacing.ActBreakdown == nil {
		t.Fatalf("expected ActBreakdown to be initialized")
	}

	act1 := plan.Pacing.ActBreakdown["act_1"]
	if act1.Planned != 4 {
		t.Errorf("expected Act 1 planned 4, got %d", act1.Planned)
	}
	if act1.Actual != 4 {
		t.Errorf("expected Act 1 actual 4, got %d", act1.Actual)
	}

	act2 := plan.Pacing.ActBreakdown["act_2"]
	if act2.Planned != 4 {
		t.Errorf("expected Act 2 planned 4, got %d", act2.Planned)
	}
	if act2.Actual != 2 {
		t.Errorf("expected Act 2 actual 2, got %d", act2.Actual)
	}
}

func TestAddMemorableMoment(t *testing.T) {
	plan := &CampaignPlan{
		DMNotes: DMNotes{
			MemorableMoments: []string{},
		},
	}

	plan.AddMemorableMoment("Epic dragon battle")
	plan.AddMemorableMoment("Betrayal by NPC")

	if len(plan.DMNotes.MemorableMoments) != 2 {
		t.Errorf("expected 2 memorable moments, got %d", len(plan.DMNotes.MemorableMoments))
	}

	if plan.DMNotes.MemorableMoments[0] != "Epic dragon battle" {
		t.Errorf("unexpected first moment: '%s'", plan.DMNotes.MemorableMoments[0])
	}
}
