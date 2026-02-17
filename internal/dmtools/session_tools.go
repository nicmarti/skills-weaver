package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/adventure"
)

// NewStartSessionTool creates a tool to start a new game session.
// If agentManager is provided, it will automatically consult world-keeper for campaign briefing.
func NewStartSessionTool(adv *adventure.Adventure, agentManager AgentManager) *SimpleTool {
	return &SimpleTool{
		name:        "start_session",
		description: "Start a new game session. This MUST be called at the beginning of each play session to properly track events and journal entries.",
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Start the session
			session, err := adv.StartSession()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			display := fmt.Sprintf("✓ Session %d démarrée", session.ID)

			// === NEW: LOAD CAMPAIGN PLAN AND GENERATE BRIEFING ===
			var systemBrief string

			campaignPlan, err := adv.LoadCampaignPlan()
			if err == nil && campaignPlan != nil && agentManager != nil {
				currentAct := campaignPlan.GetCurrentAct()
				criticalForeshadows := campaignPlan.GetCriticalForeshadows()

				// Load current state for location context
				state, _ := adv.LoadState()
				currentLocation := "Unknown"
				if state != nil {
					currentLocation = state.CurrentLocation
				}

				// Build campaign context for world-keeper
				campaignContext := buildCampaignContext(campaignPlan, currentAct, criticalForeshadows, session.ID, currentLocation)

				// === SILENT WORLD-KEEPER INVOCATION ===
				worldKeeperResponse, err := agentManager.InvokeAgentSilent("world-keeper", campaignContext, 1)

				if err == nil && worldKeeperResponse != "" {
					// Format system brief (hidden from player, for DM only)
					systemBrief = formatSystemBrief(campaignPlan, currentAct, criticalForeshadows, worldKeeperResponse)
				}
			}

			// Check for stale foreshadows (legacy compatibility - still show for adventures without campaign plan)
			staleForeshadows, err := adv.GetStaleForeshadows(3)
			if err == nil && len(staleForeshadows) > 0 {
				display += fmt.Sprintf("\n\n⚠️  RAPPEL: %d foreshadow(s) en attente depuis plus de 3 sessions:", len(staleForeshadows))
				for i, f := range staleForeshadows {
					age := session.ID - f.PlantedSession
					display += fmt.Sprintf("\n  %d. [%s] %s (%d sessions ago, %s)", i+1, f.ID, f.Description, age, f.Importance)
				}
				display += "\n\n💡 Utilisez list_foreshadows ou get_stale_foreshadows pour plus de détails."
			}

			result := map[string]interface{}{
				"success":    true,
				"session_id": session.ID,
				"started_at": session.StartedAt.Format("2006-01-02 15:04:05"),
				"display":    display,
			}

			// Add system brief if available (hidden from player, injected into agent context)
			if systemBrief != "" {
				result["system_brief"] = systemBrief
			}

			return result, nil
		},
	}
}

// buildCampaignContext constructs the briefing request for world-keeper.
func buildCampaignContext(plan *adventure.CampaignPlan, currentAct *adventure.Act, criticalForeshadows []adventure.ForeshadowLinked, sessionID int, currentLocation string) string {
	var sb strings.Builder

	sb.WriteString("Campaign Briefing Request for Session Start\n\n")
	sb.WriteString(fmt.Sprintf("**Campaign**: %s\n", plan.Metadata.CampaignTitle))
	sb.WriteString(fmt.Sprintf("**Session**: %d\n", sessionID))
	sb.WriteString(fmt.Sprintf("**Current Act**: %d - %s\n", currentAct.Number, currentAct.Title))
	sb.WriteString(fmt.Sprintf("**Act Description**: %s\n\n", currentAct.Description))
	sb.WriteString(fmt.Sprintf("**Campaign Objective**: %s\n\n", plan.NarrativeStructure.Objective))

	// Active narrative threads
	if len(plan.Progression.ActiveThreads) > 0 {
		sb.WriteString("**Active Narrative Threads**:\n")
		for _, thread := range plan.Progression.ActiveThreads {
			sb.WriteString(fmt.Sprintf("  • %s\n", thread))
		}
		sb.WriteString("\n")
	}

	// Critical foreshadows needing payoff
	if len(criticalForeshadows) > 0 {
		sb.WriteString(fmt.Sprintf("**Critical Foreshadows** (%d pending resolution):\n", len(criticalForeshadows)))
		for _, f := range criticalForeshadows {
			age := sessionID - f.PlantedSession
			sb.WriteString(fmt.Sprintf("  • [%s] %s (planted %d sessions ago, importance: %s)\n",
				f.ID, f.Description, age, f.Importance))
		}
		sb.WriteString("\n")
	}

	// Current location
	sb.WriteString(fmt.Sprintf("**Current Location**: %s\n\n", currentLocation))

	// Act goals
	if len(currentAct.Goals) > 0 {
		sb.WriteString("**Act Goals**:\n")
		for _, goal := range currentAct.Goals {
			sb.WriteString(fmt.Sprintf("  • %s\n", goal))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("**Request**: Provide brief pre-session guidance for the Dungeon Master. ")
	sb.WriteString("Suggest narrative focus, potential developments, and how to advance active threads. ")
	sb.WriteString("Keep response concise (3-5 paragraphs max).")

	return sb.String()
}

// formatSystemBrief formats the confidential briefing for the DM.
func formatSystemBrief(plan *adventure.CampaignPlan, currentAct *adventure.Act, criticalForeshadows []adventure.ForeshadowLinked, worldKeeperResponse string) string {
	var sb strings.Builder

	sb.WriteString("=== CAMPAIGN CONTEXT (CONFIDENTIAL - DO NOT QUOTE DIRECTLY) ===\n\n")
	sb.WriteString(fmt.Sprintf("**Act %d**: %s\n", currentAct.Number, currentAct.Title))
	sb.WriteString(fmt.Sprintf("%s\n\n", currentAct.Description))
	sb.WriteString(fmt.Sprintf("**Campaign Objective**: %s\n\n", plan.NarrativeStructure.Objective))

	// Active threads
	if len(plan.Progression.ActiveThreads) > 0 {
		sb.WriteString("**Active Threads**:\n")
		for _, thread := range plan.Progression.ActiveThreads {
			sb.WriteString(fmt.Sprintf("  • %s\n", thread))
		}
		sb.WriteString("\n")
	}

	// Critical foreshadows
	if len(criticalForeshadows) > 0 {
		sb.WriteString(fmt.Sprintf("**Critical Foreshadows** (%d):\n", len(criticalForeshadows)))
		for _, f := range criticalForeshadows {
			sb.WriteString(fmt.Sprintf("  • [%s] %s (linked to Act %d, %s)\n",
				f.ID, f.Description, f.LinkedToAct, f.Importance))
		}
		sb.WriteString("\n")
	}

	// World-keeper guidance
	sb.WriteString("**World-Keeper Briefing**:\n")
	sb.WriteString(worldKeeperResponse)
	sb.WriteString("\n\n")

	// Instructions for DM
	sb.WriteString("=== INSTRUCTIONS ===\n")
	sb.WriteString("• Use this context to guide your narration naturally\n")
	sb.WriteString("• DO NOT quote world-keeper directly to players\n")
	sb.WriteString("• DO NOT say \"The world-keeper informs me that...\"\n")
	sb.WriteString("• Integrate information organically into the story\n")
	sb.WriteString("• Show, don't tell: use NPC dialogue, environmental clues, rumors\n")
	sb.WriteString("===\n")

	return sb.String()
}

// NewEndSessionTool creates a tool to end the current game session.
func NewEndSessionTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "end_session",
		description: "End the current game session with a summary. This MUST be called when players finish playing to properly close the session and organize the journal.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "Summary of what happened during this session (in French)",
				},
			},
			"required": []string{"summary"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			summary := params["summary"].(string)

			// End the session
			session, err := adv.EndSession(summary)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			return map[string]interface{}{
				"success":    true,
				"session_id": session.ID,
				"duration":   session.Duration,
				"summary":    session.Summary,
				"display":    fmt.Sprintf("Session %d terminée - Durée: %s", session.ID, session.Duration),
			}, nil
		},
	}
}

// NewGetSessionInfoTool creates a tool to get information about the current session.
func NewGetSessionInfoTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "get_session_info",
		description: "Get information about the current active session (if any)",
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			session, err := adv.GetCurrentSession()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			if session == nil {
				return map[string]interface{}{
					"success":        true,
					"session_active": false,
					"display":        "Aucune session active",
				}, nil
			}

			return map[string]interface{}{
				"success":        true,
				"session_active": true,
				"session_id":     session.ID,
				"started_at":     session.StartedAt.Format("2006-01-02 15:04:05"),
				"display":        fmt.Sprintf("Session %d en cours depuis %s", session.ID, session.StartedAt.Format("15:04")),
			}, nil
		},
	}
}
