package dmtools

import (
	"fmt"

	"dungeons/internal/adventure"
	"dungeons/internal/npcmanager"
)

// NewUpdateNPCImportanceTool creates a tool to update an NPC's importance level.
func NewUpdateNPCImportanceTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "update_npc_importance",
		description: "Update an NPC's importance level and add notes about their role in the story. Use this when an NPC becomes more significant.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"npc_name": map[string]interface{}{
					"type":        "string",
					"description": "The exact name of the NPC to update",
				},
				"importance": map[string]interface{}{
					"type": "string",
					"enum": []string{"mentioned", "interacted", "recurring", "key"},
					"description": "New importance level. mentioned < interacted < recurring < key. Only increases, never decreases.",
				},
				"note": map[string]interface{}{
					"type":        "string",
					"description": "Note about why this NPC is important. Example: 'Revealed key information about Vaskir', 'Agreed to help party', 'Became recurring ally'",
				},
			},
			"required": []string{"npc_name", "importance"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			npcName := params["npc_name"].(string)
			importanceStr := params["importance"].(string)
			note := ""
			if n, ok := params["note"].(string); ok {
				note = n
			}

			// Convert string to ImportanceLevel
			var importance npcmanager.ImportanceLevel
			switch importanceStr {
			case "mentioned":
				importance = npcmanager.ImportanceMentioned
			case "interacted":
				importance = npcmanager.ImportanceInteracted
			case "recurring":
				importance = npcmanager.ImportanceRecurring
			case "key":
				importance = npcmanager.ImportanceKey
			default:
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Invalid importance level: %s", importanceStr),
				}, nil
			}

			// Update importance
			mgr := npcmanager.NewManager(adv.BasePath())
			err := mgr.UpdateImportance(npcName, importance, note)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			// Get updated record
			record, err := mgr.GetNPCHistory(npcName)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			display := fmt.Sprintf("✓ Updated %s: importance=%s, appearances=%d",
				npcName, record.Importance, record.Appearances)
			if note != "" {
				display += fmt.Sprintf("\n  Note: %s", note)
			}

			return map[string]interface{}{
				"success": true,
				"display": display,
				"record":  record,
			}, nil
		},
	}
}

// NewGetNPCHistoryTool creates a tool to retrieve an NPC's full history.
func NewGetNPCHistoryTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "get_npc_history",
		description: "Retrieve the complete history of an NPC including all notes, appearances, and context from previous sessions.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"npc_name": map[string]interface{}{
					"type":        "string",
					"description": "The exact name of the NPC to query",
				},
			},
			"required": []string{"npc_name"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			npcName := params["npc_name"].(string)

			// Get NPC history
			mgr := npcmanager.NewManager(adv.BasePath())
			record, err := mgr.GetNPCHistory(npcName)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}, nil
			}

			// Format display
			display := fmt.Sprintf("**%s** (ID: %s)\n", record.NPC.Name, record.ID)
			display += fmt.Sprintf("- Race: %s, Gender: %s\n", record.NPC.Race, record.NPC.Gender)
			display += fmt.Sprintf("- Occupation: %s\n", record.NPC.Occupation)
			display += fmt.Sprintf("- Importance: %s (appeared %d times)\n", record.Importance, record.Appearances)
			display += fmt.Sprintf("- First seen: Session %d (%s)\n", record.SessionNumber, record.GeneratedAt.Format("2006-01-02"))

			if record.Context != "" {
				display += fmt.Sprintf("- Context: %s\n", record.Context)
			}

			if len(record.Notes) > 0 {
				display += "\n**Notes:**\n"
				for i, note := range record.Notes {
					display += fmt.Sprintf("%d. %s\n", i+1, note)
				}
			}

			if record.WorldKeeperNotes != "" {
				display += fmt.Sprintf("\n**World-Keeper:** %s\n", record.WorldKeeperNotes)
			}

			if record.PromotedToWorld {
				display += "\n✓ Promoted to world/npcs.json"
			}

			// Include full NPC data for detailed reference
			display += fmt.Sprintf("\n**Appearance:** %s, %s, %s eyes, %s hair\n",
				record.NPC.Appearance.Build,
				record.NPC.Appearance.Height,
				record.NPC.Appearance.EyeColor,
				record.NPC.Appearance.HairColor,
			)
			display += fmt.Sprintf("**Personality:** %s, %s\n",
				record.NPC.Personality.TraitPrincipal,
				record.NPC.Personality.TraitSecondaire,
			)
			display += fmt.Sprintf("**Attitude:** %s\n", record.NPC.Attitude)

			return map[string]interface{}{
				"success": true,
				"display": display,
				"record":  record,
			}, nil
		},
	}
}
