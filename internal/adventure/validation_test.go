package adventure

import (
	"testing"
)

func TestValidateCampaignPlan_Nil(t *testing.T) {
	result := ValidateCampaignPlan(nil)
	if result.Score != 0 {
		t.Errorf("expected score 0 for nil plan, got %d", result.Score)
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error for nil plan, got %d", len(result.Errors))
	}
}

func TestValidateCampaignPlan_ValidOneAct(t *testing.T) {
	plan := &CampaignPlan{
		NarrativeStructure: NarrativeStructure{
			Objective: "Find the lost treasure",
			Hook:      "A merchant asks for help",
			Acts: []Act{
				{
					Number:         1,
					Title:          "The Hunt",
					TargetSessions: []int{1, 2, 3},
					KeyEvents:      []string{"Find the map"},
					Goals:          []string{"Recover the treasure"},
				},
			},
			Climax: Climax{Description: "Face the guardian", TargetSession: 3},
		},
		PlotElements: PlotElements{
			Antagonist: Character{
				Name:                      "Bandit Chief Gorlag",
				Motivation:                "Greed and power",
				IntroductionSession:       1,
				FinalConfrontationSession: 3,
			},
		},
		Foreshadows: ForeshadowsContainer{
			Active: []ForeshadowLinked{},
		},
	}

	result := ValidateCampaignPlan(plan)
	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(result.Errors), result.Errors)
	}
	if result.Score < 80 {
		t.Errorf("expected score >= 80 for valid plan, got %d", result.Score)
	}
}

func TestValidateCampaignPlan_MissingAntagonist(t *testing.T) {
	plan := &CampaignPlan{
		NarrativeStructure: NarrativeStructure{
			Acts: []Act{
				{
					Number:         1,
					Title:          "Act 1",
					TargetSessions: []int{1, 2},
					KeyEvents:      []string{"Event"},
					Goals:          []string{"Goal"},
				},
			},
		},
		PlotElements: PlotElements{
			Antagonist: Character{}, // Missing name and motivation
		},
		Foreshadows: ForeshadowsContainer{},
	}

	result := ValidateCampaignPlan(plan)
	if len(result.Errors) < 2 {
		t.Errorf("expected at least 2 errors (no name, no motivation), got %d: %v",
			len(result.Errors), result.Errors)
	}
}

func TestValidateCampaignPlan_DuplicateSessions(t *testing.T) {
	plan := &CampaignPlan{
		NarrativeStructure: NarrativeStructure{
			Acts: []Act{
				{Number: 1, Title: "Act 1", TargetSessions: []int{1, 2, 3}, KeyEvents: []string{"E"}, Goals: []string{"G"}},
				{Number: 2, Title: "Act 2", TargetSessions: []int{3, 4, 5}, KeyEvents: []string{"E"}, Goals: []string{"G"}},
			},
		},
		PlotElements: PlotElements{
			Antagonist: Character{Name: "Villain", Motivation: "Evil", IntroductionSession: 1, FinalConfrontationSession: 5},
		},
		Foreshadows: ForeshadowsContainer{},
	}

	result := ValidateCampaignPlan(plan)
	found := false
	for _, e := range result.Errors {
		if e == "Session 3 is assigned to multiple acts" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate session error, got: %v", result.Errors)
	}
}

func TestValidateCampaignPlan_AntagonistIntroAfterConfrontation(t *testing.T) {
	plan := &CampaignPlan{
		NarrativeStructure: NarrativeStructure{
			Acts: []Act{
				{Number: 1, Title: "Act 1", TargetSessions: []int{1, 2}, KeyEvents: []string{"E"}, Goals: []string{"G"}},
			},
		},
		PlotElements: PlotElements{
			Antagonist: Character{
				Name:                      "Villain",
				Motivation:                "Revenge",
				IntroductionSession:       5,
				FinalConfrontationSession: 3,
			},
		},
		Foreshadows: ForeshadowsContainer{},
	}

	result := ValidateCampaignPlan(plan)
	found := false
	for _, e := range result.Errors {
		if len(e) > 0 && e[:10] == "Antagonist" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected antagonist intro/confrontation error, got: %v", result.Errors)
	}
}

func TestValidateCampaignPlan_CultKeywordWarning(t *testing.T) {
	plan := &CampaignPlan{
		NarrativeStructure: NarrativeStructure{
			Objective: "Stop the ancient cult from summoning the entity",
			Acts: []Act{
				{Number: 1, Title: "Act 1", TargetSessions: []int{1, 2}, KeyEvents: []string{"E"}, Goals: []string{"G"}},
			},
		},
		PlotElements: PlotElements{
			Antagonist: Character{Name: "Cult Leader", Motivation: "Summon the old god"},
		},
		Foreshadows: ForeshadowsContainer{},
	}

	result := ValidateCampaignPlan(plan)
	if len(result.Warnings) == 0 {
		t.Errorf("expected cult keyword warnings, got none")
	}
}

func TestValidateCampaignPlan_ForeshadowInvalidAct(t *testing.T) {
	plan := &CampaignPlan{
		NarrativeStructure: NarrativeStructure{
			Acts: []Act{
				{Number: 1, Title: "Act 1", TargetSessions: []int{1, 2}, KeyEvents: []string{"E"}, Goals: []string{"G"}},
			},
		},
		PlotElements: PlotElements{
			Antagonist: Character{Name: "Villain", Motivation: "Greed"},
		},
		Foreshadows: ForeshadowsContainer{
			Active: []ForeshadowLinked{
				{ID: "fsh_001", LinkedToAct: 5}, // Act 5 doesn't exist
			},
		},
	}

	result := ValidateCampaignPlan(plan)
	found := false
	for _, w := range result.Warnings {
		if len(w) > 0 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected foreshadow warning for invalid act, got: %v", result.Warnings)
	}
}
