package dmtools

import (
	"dungeons/internal/adventure"
	"fmt"
)

// NewGetCampaignPlanTool creates a tool to query the current campaign plan.
func NewGetCampaignPlanTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name: "get_campaign_plan",
		description: `Get the current campaign plan including narrative structure, progression, and foreshadows.
Returns the 3-act structure, current objectives, active narrative threads, and pending plot resolutions.
Use this to understand the overall campaign arc and guide your narration.`,
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"section": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"all", "current_act", "progression", "foreshadows", "pacing"},
					"description": "Which section of the campaign plan to retrieve (default: all)",
					"default":     "all",
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			section := "all"
			if s, ok := params["section"].(string); ok {
				section = s
			}

			plan, err := adv.LoadCampaignPlan()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			if plan == nil {
				return map[string]interface{}{
					"success": false,
					"message": "No campaign plan exists for this adventure",
					"display": "⚠️  No campaign plan found. This adventure was created before campaign planning was introduced.",
				}, nil
			}

			result := map[string]interface{}{
				"success": true,
			}

			switch section {
			case "all":
				result["campaign_plan"] = plan
				result["display"] = formatCampaignPlanSummary(plan)

			case "current_act":
				currentAct := plan.GetCurrentAct()
				if currentAct == nil {
					return map[string]interface{}{
						"success": false,
						"error":   "No current act found",
					}, nil
				}
				result["current_act"] = currentAct
				result["display"] = formatActSummary(currentAct, plan.Progression.CurrentSession)

			case "progression":
				result["progression"] = plan.Progression
				result["display"] = formatProgressionSummary(&plan.Progression, plan.GetCurrentAct())

			case "foreshadows":
				result["active_foreshadows"] = plan.Foreshadows.Active
				result["resolved_foreshadows"] = plan.Foreshadows.Resolved
				result["critical_foreshadows"] = plan.GetCriticalForeshadows()
				result["display"] = formatCampaignForeshadowsSummary(&plan.Foreshadows, plan.GetCriticalForeshadows())

			case "pacing":
				result["pacing"] = plan.Pacing
				result["display"] = formatPacingSummary(&plan.Pacing, plan.Progression.CurrentSession)

			default:
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("unknown section: %s", section),
				}, nil
			}

			return result, nil
		},
	}
}

// NewUpdateCampaignProgressTool creates a tool to update campaign progress.
func NewUpdateCampaignProgressTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name: "update_campaign_progress",
		description: `Update campaign progress by marking plot points complete or advancing to the next act.
Use this when major story milestones are achieved (e.g., completing an act's objective).`,
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"complete_plot_point", "advance_act"},
					"description": "Type of progress update",
				},
				"plot_point_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of plot point to mark complete (for complete_plot_point action)",
				},
				"act_number": map[string]interface{}{
					"type":        "integer",
					"description": "Act number to advance to (for advance_act action)",
					"minimum":     1,
					"maximum":     3,
				},
			},
			"required": []string{"action"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			action := params["action"].(string)

			plan, err := adv.LoadCampaignPlan()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			if plan == nil {
				return map[string]interface{}{
					"success": false,
					"error":   "No campaign plan exists",
				}, nil
			}

			switch action {
			case "complete_plot_point":
				plotPointID, ok := params["plot_point_id"].(string)
				if !ok {
					return map[string]interface{}{
						"success": false,
						"error":   "plot_point_id required for complete_plot_point action",
					}, nil
				}

				if err := plan.CompletePlotPoint(plotPointID); err != nil {
					return map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					}, nil
				}

				if err := adv.SaveCampaignPlan(plan); err != nil {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("failed to save: %v", err),
					}, nil
				}

				adv.LogEvent("story", fmt.Sprintf("Plot point completed: %s", plotPointID))

				return map[string]interface{}{
					"success": true,
					"display": fmt.Sprintf("✓ Plot point '%s' marked as completed", plotPointID),
				}, nil

			case "advance_act":
				actNumber, ok := params["act_number"].(float64) // JSON numbers are float64
				if !ok {
					return map[string]interface{}{
						"success": false,
						"error":   "act_number required for advance_act action",
					}, nil
				}

				actInt := int(actNumber)
				if err := plan.AdvanceAct(actInt); err != nil {
					return map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					}, nil
				}

				// Update pacing
				plan.UpdatePacing()

				if err := adv.SaveCampaignPlan(plan); err != nil {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("failed to save: %v", err),
					}, nil
				}

				currentAct := plan.GetCurrentAct()
				adv.LogEvent("session", fmt.Sprintf("Advanced to Act %d: %s", actInt, currentAct.Title))

				return map[string]interface{}{
					"success": true,
					"display": fmt.Sprintf("✓ Advanced to Act %d: %s", actInt, currentAct.Title),
				}, nil

			default:
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("unknown action: %s", action),
				}, nil
			}
		},
	}
}

// NewAddNarrativeThreadTool creates a tool to add a new active narrative thread.
func NewAddNarrativeThreadTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name: "add_narrative_thread",
		description: `Add a new active narrative thread to track.
Use this when introducing a new subplot or mystery that will span multiple sessions.`,
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"thread_name": map[string]interface{}{
					"type":        "string",
					"description": "Name/ID of the narrative thread (e.g., 'vaskir_ritual_countdown', 'alliance_betrayal')",
				},
			},
			"required": []string{"thread_name"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			threadName := params["thread_name"].(string)

			plan, err := adv.LoadCampaignPlan()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			if plan == nil {
				return map[string]interface{}{
					"success": false,
					"error":   "No campaign plan exists",
				}, nil
			}

			if err := plan.AddActiveThread(threadName); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			if err := adv.SaveCampaignPlan(plan); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("failed to save: %v", err),
				}, nil
			}

			adv.LogEvent("story", fmt.Sprintf("New narrative thread: %s", threadName))

			return map[string]interface{}{
				"success": true,
				"display": fmt.Sprintf("✓ Added narrative thread: %s", threadName),
			}, nil
		},
	}
}

// NewRemoveNarrativeThreadTool creates a tool to remove a resolved narrative thread.
func NewRemoveNarrativeThreadTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name: "remove_narrative_thread",
		description: `Remove a resolved narrative thread.
Use this when a subplot or mystery has been fully resolved.`,
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"thread_name": map[string]interface{}{
					"type":        "string",
					"description": "Name/ID of the narrative thread to remove",
				},
			},
			"required": []string{"thread_name"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			threadName := params["thread_name"].(string)

			plan, err := adv.LoadCampaignPlan()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			if plan == nil {
				return map[string]interface{}{
					"success": false,
					"error":   "No campaign plan exists",
				}, nil
			}

			if err := plan.RemoveActiveThread(threadName); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			if err := adv.SaveCampaignPlan(plan); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("failed to save: %v", err),
				}, nil
			}

			adv.LogEvent("story", fmt.Sprintf("Resolved narrative thread: %s", threadName))

			return map[string]interface{}{
				"success": true,
				"display": fmt.Sprintf("✓ Removed narrative thread: %s", threadName),
			}, nil
		},
	}
}

// Helper formatting functions

func formatCampaignPlanSummary(plan *adventure.CampaignPlan) string {
	currentAct := plan.GetCurrentAct()
	criticalForeshadows := plan.GetCriticalForeshadows()

	return fmt.Sprintf(`📖 Campaign: %s

🎯 Objective: %s
🎬 Hook: %s

📍 Current State:
  • Act %d/%d: %s
  • Session: %d
  • Active Threads: %d
  • Critical Foreshadows: %d
  • Pending Resolutions: %d

⏱️  Pacing:
  • Sessions played: %d
  • Estimated remaining: %d
`,
		plan.Metadata.CampaignTitle,
		plan.NarrativeStructure.Objective,
		plan.NarrativeStructure.Hook,
		currentAct.Number, len(plan.NarrativeStructure.Acts), currentAct.Title,
		plan.Progression.CurrentSession,
		len(plan.Progression.ActiveThreads),
		len(criticalForeshadows),
		len(plan.Progression.PendingResolutions),
		plan.Pacing.SessionsPlayed,
		plan.Pacing.SessionsRemainingEstimate,
	)
}

func formatActSummary(act *adventure.Act, currentSession int) string {
	return fmt.Sprintf(`🎭 Act %d: %s (%s)

📝 %s

🎯 Goals:
%s

🔑 Key Events:
%s

✅ Completion: %s
`,
		act.Number,
		act.Title,
		act.Status,
		act.Description,
		formatListItems(act.Goals),
		formatListItems(act.KeyEvents),
		act.CompletionCriteria.Milestone,
	)
}

func formatProgressionSummary(prog *adventure.Progression, currentAct *adventure.Act) string {
	return fmt.Sprintf(`📊 Campaign Progression

Current Act: Act %d - %s
Current Session: %d

✅ Completed Plot Points (%d):
%s

🔄 Active Threads (%d):
%s

⏳ Pending Resolutions (%d):
%s
`,
		prog.CurrentAct,
		currentAct.Title,
		prog.CurrentSession,
		len(prog.CompletedPlotPoints),
		formatListItems(prog.CompletedPlotPoints),
		len(prog.ActiveThreads),
		formatListItems(prog.ActiveThreads),
		len(prog.PendingResolutions),
		formatListItems(prog.PendingResolutions),
	)
}

func formatCampaignForeshadowsSummary(container *adventure.ForeshadowsContainer, critical []adventure.ForeshadowLinked) string {
	result := fmt.Sprintf(`🔮 Foreshadows

📌 Active: %d
✅ Resolved: %d
⚠️  Critical (>= 3 sessions old): %d
`,
		len(container.Active),
		len(container.Resolved),
		len(critical),
	)

	if len(critical) > 0 {
		result += "\nCritical Foreshadows needing payoff:\n"
		for _, f := range critical {
			result += fmt.Sprintf("  • [%s] %s (planted session %d, importance: %s)\n",
				f.ID, f.Description, f.PlantedSession, f.Importance)
		}
	}

	return result
}

func formatPacingSummary(pacing *adventure.Pacing, currentSession int) string {
	result := fmt.Sprintf(`⏱️  Campaign Pacing

Sessions Played: %d
Estimated Remaining: %d

Act Breakdown:
`,
		pacing.SessionsPlayed,
		pacing.SessionsRemainingEstimate,
	)

	for actKey, breakdown := range pacing.ActBreakdown {
		variance := ""
		if breakdown.Variance > 0 {
			variance = fmt.Sprintf(" (+%d over)", breakdown.Variance)
		} else if breakdown.Variance < 0 {
			variance = fmt.Sprintf(" (%d under)", -breakdown.Variance)
		}

		result += fmt.Sprintf("  • %s: %d/%d sessions%s\n",
			actKey, breakdown.Actual, breakdown.Planned, variance)
	}

	return result
}

func formatListItems(items []string) string {
	if len(items) == 0 {
		return "  (none)"
	}

	result := ""
	for _, item := range items {
		result += fmt.Sprintf("  • %s\n", item)
	}
	return result
}
