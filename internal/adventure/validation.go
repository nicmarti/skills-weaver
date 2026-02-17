package adventure

import (
	"fmt"
	"strings"
)

// ValidationResult holds the results of a campaign plan validation.
type ValidationResult struct {
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Score    int      `json:"score"` // 0-100
}

// cultKeywords are terms that suggest cosmic/cult elements that should be avoided.
var cultKeywords = []string{
	"culte", "cult",
	"entite cosmique", "cosmic entity",
	"eldritch", "lovecraft",
	"rituel ancien", "ancient ritual",
	"fin du monde", "end of the world", "apocalypse",
	"dieu ancien", "old god", "ancien dieu",
	"dimension", "portail dimensionnel",
	"secte", "sect",
}

// ValidateCampaignPlan checks a campaign plan for structural errors and warnings.
func ValidateCampaignPlan(plan *CampaignPlan) *ValidationResult {
	result := &ValidationResult{
		Score: 100,
	}

	if plan == nil {
		result.Errors = append(result.Errors, "Campaign plan is nil")
		result.Score = 0
		return result
	}

	// === ERRORS (blocking) ===

	// Check acts have events and goals
	for _, act := range plan.NarrativeStructure.Acts {
		if len(act.KeyEvents) == 0 {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Act %d (%s) has no key events", act.Number, act.Title))
		}
		if len(act.Goals) == 0 {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Act %d (%s) has no goals", act.Number, act.Title))
		}
	}

	// Check antagonist has name and motivation
	if plan.PlotElements.Antagonist.Name == "" {
		result.Errors = append(result.Errors, "Antagonist has no name")
	}
	if plan.PlotElements.Antagonist.Motivation == "" {
		result.Errors = append(result.Errors, "Antagonist has no motivation")
	}

	// Check session targets are contiguous
	if len(plan.NarrativeStructure.Acts) > 1 {
		allSessions := map[int]bool{}
		for _, act := range plan.NarrativeStructure.Acts {
			for _, s := range act.TargetSessions {
				if allSessions[s] {
					result.Errors = append(result.Errors,
						fmt.Sprintf("Session %d is assigned to multiple acts", s))
				}
				allSessions[s] = true
			}
		}
	}

	// Check antagonist is introduced before final confrontation
	antag := plan.PlotElements.Antagonist
	if antag.IntroductionSession > 0 && antag.FinalConfrontationSession > 0 {
		if antag.IntroductionSession >= antag.FinalConfrontationSession {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Antagonist '%s' is introduced (session %d) at or after final confrontation (session %d)",
					antag.Name, antag.IntroductionSession, antag.FinalConfrontationSession))
		}
	}

	// === WARNINGS ===

	// Check for locations referenced in acts but absent from key_locations
	locationNames := map[string]bool{}
	for _, loc := range plan.PlotElements.KeyLocations {
		locationNames[strings.ToLower(loc.Name)] = true
	}

	// Check for acts with 0 sessions
	for _, act := range plan.NarrativeStructure.Acts {
		if len(act.TargetSessions) == 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Act %d (%s) has no target sessions assigned", act.Number, act.Title))
		}
	}

	// Check foreshadows point to valid acts
	maxAct := len(plan.NarrativeStructure.Acts)
	for _, fs := range plan.Foreshadows.Active {
		if fs.LinkedToAct > maxAct {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Foreshadow '%s' links to act %d but only %d act(s) exist",
					fs.ID, fs.LinkedToAct, maxAct))
		}
	}

	// Check for NPCs in foreshadows but absent from NPCs list
	npcNames := map[string]bool{}
	for _, n := range plan.PlotElements.NPCs {
		npcNames[strings.ToLower(n.Name)] = true
	}
	// Also add antagonist and supporting characters
	if plan.PlotElements.Antagonist.Name != "" {
		npcNames[strings.ToLower(plan.PlotElements.Antagonist.Name)] = true
	}
	for _, sc := range plan.PlotElements.SupportingCharacters {
		npcNames[strings.ToLower(sc.Name)] = true
	}

	for _, fs := range plan.Foreshadows.Active {
		for _, npcRef := range fs.RelatedNPCs {
			if !npcNames[strings.ToLower(npcRef)] {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Foreshadow '%s' references NPC '%s' not found in NPCs or characters",
						fs.ID, npcRef))
			}
		}
	}

	// Detect cult/cosmic keywords in text fields
	textsToCheck := []string{
		plan.NarrativeStructure.Objective,
		plan.NarrativeStructure.Hook,
		plan.PlotElements.Antagonist.Motivation,
		plan.PlotElements.Antagonist.Arc,
		plan.NarrativeStructure.Climax.Description,
		plan.NarrativeStructure.Climax.Stakes,
	}
	for _, act := range plan.NarrativeStructure.Acts {
		textsToCheck = append(textsToCheck, act.Description)
		textsToCheck = append(textsToCheck, act.KeyEvents...)
	}

	for _, text := range textsToCheck {
		lower := strings.ToLower(text)
		for _, keyword := range cultKeywords {
			if strings.Contains(lower, keyword) {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Cult/cosmic keyword '%s' detected in plan text: '%.50s...'",
						keyword, text))
				break // One warning per text block is enough
			}
		}
	}

	// === SCORE CALCULATION ===
	// Start at 100, deduct points
	result.Score -= len(result.Errors) * 15
	result.Score -= len(result.Warnings) * 5

	if result.Score < 0 {
		result.Score = 0
	}

	return result
}
