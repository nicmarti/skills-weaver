package dmtools

import (
	"dungeons/internal/adventure"
	"testing"
	"time"
)

func TestGetCampaignPlanTool_NoPlan(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &adventure.Adventure{}
	adv.SetBasePath(tmpDir)

	tool := NewGetCampaignPlanTool(adv)

	result, err := tool.Execute(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if success, ok := resultMap["success"].(bool); ok && success {
		t.Errorf("expected success=false when no campaign plan exists")
	}
}

func TestGetCampaignPlanTool_Success(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &adventure.Adventure{}
	adv.SetBasePath(tmpDir)

	// Create a campaign plan
	plan := &adventure.CampaignPlan{
		Version: "1.0.0",
		Metadata: adventure.CampaignMetadata{
			CampaignTitle: "Test Campaign",
			Theme:         "Epic Quest",
			CreatedAt:     time.Now(),
		},
		NarrativeStructure: adventure.NarrativeStructure{
			Objective: "Save the world",
			Acts: []adventure.Act{
				{
					Number: 1,
					Title:  "Beginning",
					Status: "in_progress",
				},
			},
		},
		Progression: adventure.Progression{
			CurrentAct:     1,
			CurrentSession: 1,
		},
		Foreshadows: adventure.ForeshadowsContainer{
			Active: []adventure.ForeshadowLinked{},
		},
	}

	if err := adv.SaveCampaignPlan(plan); err != nil {
		t.Fatalf("Failed to save campaign plan: %v", err)
	}

	tool := NewGetCampaignPlanTool(adv)

	result, err := tool.Execute(map[string]interface{}{
		"section": "all",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Errorf("expected success=true")
	}

	if _, ok := resultMap["campaign_plan"]; !ok {
		t.Errorf("expected campaign_plan in result")
	}
}

func TestUpdateCampaignProgressTool_CompletePlotPoint(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &adventure.Adventure{}
	adv.SetBasePath(tmpDir)

	// Create a campaign plan
	plan := &adventure.CampaignPlan{
		Version: "1.0.0",
		Metadata: adventure.CampaignMetadata{
			CampaignTitle: "Test Campaign",
			CreatedAt:     time.Now(),
		},
		NarrativeStructure: adventure.NarrativeStructure{
			Acts: []adventure.Act{
				{Number: 1, Title: "Act 1", Status: "in_progress"},
			},
		},
		Progression: adventure.Progression{
			CurrentAct:          1,
			CompletedPlotPoints: []string{},
		},
	}

	if err := adv.SaveCampaignPlan(plan); err != nil {
		t.Fatalf("Failed to save campaign plan: %v", err)
	}

	tool := NewUpdateCampaignProgressTool(adv)

	result, err := tool.Execute(map[string]interface{}{
		"action":        "complete_plot_point",
		"plot_point_id": "first_encounter",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Errorf("expected success=true")
	}

	// Verify plot point was added
	loadedPlan, err := adv.LoadCampaignPlan()
	if err != nil {
		t.Fatalf("Failed to load campaign plan: %v", err)
	}

	if len(loadedPlan.Progression.CompletedPlotPoints) != 1 {
		t.Errorf("expected 1 completed plot point, got %d", len(loadedPlan.Progression.CompletedPlotPoints))
	}
}

func TestAddNarrativeThreadTool(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &adventure.Adventure{}
	adv.SetBasePath(tmpDir)

	// Create a campaign plan
	plan := &adventure.CampaignPlan{
		Version: "1.0.0",
		Metadata: adventure.CampaignMetadata{
			CampaignTitle: "Test Campaign",
			CreatedAt:     time.Now(),
		},
		NarrativeStructure: adventure.NarrativeStructure{
			Acts: []adventure.Act{
				{Number: 1, Title: "Act 1"},
			},
		},
		Progression: adventure.Progression{
			CurrentAct:    1,
			ActiveThreads: []string{},
		},
	}

	if err := adv.SaveCampaignPlan(plan); err != nil {
		t.Fatalf("Failed to save campaign plan: %v", err)
	}

	tool := NewAddNarrativeThreadTool(adv)

	result, err := tool.Execute(map[string]interface{}{
		"thread_name": "mysterious_stranger",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Errorf("expected success=true")
	}

	// Verify thread was added
	loadedPlan, err := adv.LoadCampaignPlan()
	if err != nil {
		t.Fatalf("Failed to load campaign plan: %v", err)
	}

	if len(loadedPlan.Progression.ActiveThreads) != 1 {
		t.Errorf("expected 1 active thread, got %d", len(loadedPlan.Progression.ActiveThreads))
	}

	if loadedPlan.Progression.ActiveThreads[0] != "mysterious_stranger" {
		t.Errorf("expected thread 'mysterious_stranger', got '%s'", loadedPlan.Progression.ActiveThreads[0])
	}
}

func TestRemoveNarrativeThreadTool(t *testing.T) {
	tmpDir := t.TempDir()
	adv := &adventure.Adventure{}
	adv.SetBasePath(tmpDir)

	// Create a campaign plan with an active thread
	plan := &adventure.CampaignPlan{
		Version: "1.0.0",
		Metadata: adventure.CampaignMetadata{
			CampaignTitle: "Test Campaign",
			CreatedAt:     time.Now(),
		},
		NarrativeStructure: adventure.NarrativeStructure{
			Acts: []adventure.Act{
				{Number: 1, Title: "Act 1"},
			},
		},
		Progression: adventure.Progression{
			CurrentAct:    1,
			ActiveThreads: []string{"old_thread"},
		},
	}

	if err := adv.SaveCampaignPlan(plan); err != nil {
		t.Fatalf("Failed to save campaign plan: %v", err)
	}

	tool := NewRemoveNarrativeThreadTool(adv)

	result, err := tool.Execute(map[string]interface{}{
		"thread_name": "old_thread",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Errorf("expected success=true")
	}

	// Verify thread was removed
	loadedPlan, err := adv.LoadCampaignPlan()
	if err != nil {
		t.Fatalf("Failed to load campaign plan: %v", err)
	}

	if len(loadedPlan.Progression.ActiveThreads) != 0 {
		t.Errorf("expected 0 active threads, got %d", len(loadedPlan.Progression.ActiveThreads))
	}
}
