package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/adventure"
)

// NewPlantForeshadowTool creates a tool to plant narrative seeds.
func NewPlantForeshadowTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "plant_foreshadow",
		description: "Plant a narrative seed (foreshadow) for future payoff. Use this to track hints, clues, prophecies, villain mentions, or any story element that should be resolved later. The tool automatically associates the foreshadow with the current session.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Brief description of the foreshadow (e.g., 'Mysterious Seigneur Noir mentioned by Grimbold')",
				},
				"context": map[string]interface{}{
					"type":        "string",
					"description": "Additional context about how it was introduced (e.g., 'Tavern dialogue, cryptic warning about eastern threat')",
				},
				"importance": map[string]interface{}{
					"type":        "string",
					"enum":        []interface{}{"minor", "moderate", "major", "critical"},
					"description": "Importance level: minor (background detail), moderate (notable hint), major (key plot point), critical (central to campaign)",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"enum":        []interface{}{"villain", "artifact", "prophecy", "mystery", "faction", "location", "character"},
					"description": "Category to help organize narrative threads",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"description": "Optional tags for searchability (e.g., ['seigneur-noir', 'antagoniste', 'prophecy'])",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"related_npcs": map[string]interface{}{
					"type":        "array",
					"description": "Optional list of related NPC names",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"related_locations": map[string]interface{}{
					"type":        "array",
					"description": "Optional list of related location names",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"payoff_session": map[string]interface{}{
					"type":        "number",
					"description": "Optional: planned session number for payoff",
				},
			},
			"required": []interface{}{"description", "importance", "category"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			description := params["description"].(string)
			importance := adventure.Importance(params["importance"].(string))
			category := adventure.ForeshadowCategory(params["category"].(string))

			context := ""
			if c, ok := params["context"].(string); ok {
				context = c
			}

			var tags []string
			if tagsRaw, ok := params["tags"].([]interface{}); ok {
				for _, t := range tagsRaw {
					if tStr, ok := t.(string); ok {
						tags = append(tags, tStr)
					}
				}
			}

			var relatedNPCs []string
			if npcsRaw, ok := params["related_npcs"].([]interface{}); ok {
				for _, n := range npcsRaw {
					if nStr, ok := n.(string); ok {
						relatedNPCs = append(relatedNPCs, nStr)
					}
				}
			}

			var relatedLocations []string
			if locsRaw, ok := params["related_locations"].([]interface{}); ok {
				for _, l := range locsRaw {
					if lStr, ok := l.(string); ok {
						relatedLocations = append(relatedLocations, lStr)
					}
				}
			}

			var payoffSession *int
			if payoff, ok := params["payoff_session"].(float64); ok {
				p := int(payoff)
				payoffSession = &p
			}

			foreshadow, err := adv.PlantForeshadow(
				description,
				context,
				importance,
				category,
				tags,
				relatedNPCs,
				relatedLocations,
				payoffSession,
			)

			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to plant foreshadow: %v", err),
				}, nil
			}

			display := fmt.Sprintf("✓ Foreshadow planté: %s\n", foreshadow.ID)
			display += fmt.Sprintf("  Description: %s\n", foreshadow.Description)
			display += fmt.Sprintf("  Importance: %s | Category: %s\n", foreshadow.Importance, foreshadow.Category)
			display += fmt.Sprintf("  Session: %d", foreshadow.PlantedSession)

			if len(relatedNPCs) > 0 {
				display += fmt.Sprintf(" | NPCs: %s", strings.Join(relatedNPCs, ", "))
			}

			return map[string]interface{}{
				"success":      true,
				"foreshadow_id": foreshadow.ID,
				"display":      display,
			}, nil
		},
	}
}

// NewResolveForeshadowTool creates a tool to mark foreshadows as resolved.
func NewResolveForeshadowTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "resolve_foreshadow",
		description: "Mark a foreshadow as resolved when its payoff is delivered. Records when and how the narrative thread was concluded.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"foreshadow_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the foreshadow to resolve (e.g., 'fsh_001')",
				},
				"resolution": map[string]interface{}{
					"type":        "string",
					"description": "How the foreshadow was resolved (e.g., 'Seigneur Noir revealed as corrupted archmage, defeated in final battle')",
				},
			},
			"required": []interface{}{"foreshadow_id", "resolution"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			id := params["foreshadow_id"].(string)
			resolution := params["resolution"].(string)

			foreshadow, err := adv.ResolveForeshadow(id, resolution)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to resolve foreshadow: %v", err),
				}, nil
			}

			display := fmt.Sprintf("✓ Foreshadow résolu: %s\n", foreshadow.ID)
			display += fmt.Sprintf("  Description: %s\n", foreshadow.Description)
			display += fmt.Sprintf("  Resolution: %s\n", resolution)

			plantedSession := foreshadow.PlantedSession
			var currentSession int
			if session, err := adv.GetCurrentSession(); err == nil && session != nil {
				currentSession = session.ID
			}
			sessionSpan := currentSession - plantedSession
			if sessionSpan > 0 {
				display += fmt.Sprintf("  (Planted session %d, resolved session %d - %d sessions span)", plantedSession, currentSession, sessionSpan)
			}

			return map[string]interface{}{
				"success": true,
				"display": display,
			}, nil
		},
	}
}

// NewListForeshadowsTool creates a tool to list foreshadows with filters.
func NewListForeshadowsTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "list_foreshadows",
		description: "List foreshadows with optional filters. Shows active unresolved narrative threads by default, or all foreshadows with their status.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type":        "string",
					"enum":        []interface{}{"active", "resolved", "abandoned", "all"},
					"description": "Filter by status. Default: 'active' (shows only unresolved foreshadows)",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"enum":        []interface{}{"villain", "artifact", "prophecy", "mystery", "faction", "location", "character"},
					"description": "Optional: filter by category",
				},
				"importance": map[string]interface{}{
					"type":        "string",
					"enum":        []interface{}{"minor", "moderate", "major", "critical"},
					"description": "Optional: filter by importance level",
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			status := "active"
			if s, ok := params["status"].(string); ok {
				status = s
			}

			var foreshadows []adventure.Foreshadow
			var err error

			switch status {
			case "active":
				foreshadows, err = adv.GetActiveForeshadows()
			case "all":
				foreshadows, err = adv.GetAllForeshadows()
			case "resolved", "abandoned":
				// Get all and filter
				all, err := adv.GetAllForeshadows()
				if err != nil {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("Failed to load foreshadows: %v", err),
					}, nil
				}
				for _, f := range all {
					if string(f.Status) == status {
						foreshadows = append(foreshadows, f)
					}
				}
			}

			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to list foreshadows: %v", err),
				}, nil
			}

			// Apply optional filters
			if category, ok := params["category"].(string); ok {
				var filtered []adventure.Foreshadow
				for _, f := range foreshadows {
					if string(f.Category) == category {
						filtered = append(filtered, f)
					}
				}
				foreshadows = filtered
			}

			if importance, ok := params["importance"].(string); ok {
				var filtered []adventure.Foreshadow
				for _, f := range foreshadows {
					if string(f.Importance) == importance {
						filtered = append(filtered, f)
					}
				}
				foreshadows = filtered
			}

			if len(foreshadows) == 0 {
				return map[string]interface{}{
					"success": true,
					"count":   0,
					"display": "Aucun foreshadow trouvé avec les filtres spécifiés.",
				}, nil
			}

			// Format output
			display := fmt.Sprintf("=== Foreshadows (%d) ===\n\n", len(foreshadows))

			// Get current session for age calculation
			var currentSession int
			if session, err := adv.GetCurrentSession(); err == nil && session != nil {
				currentSession = session.ID
			}

			for i, f := range foreshadows {
				display += fmt.Sprintf("%d. [%s] %s\n", i+1, f.ID, f.Description)
				display += fmt.Sprintf("   Status: %s | Importance: %s | Category: %s\n", f.Status, f.Importance, f.Category)
				display += fmt.Sprintf("   Planted: Session %d", f.PlantedSession)

				if f.Status == adventure.ForeshadowActive && currentSession > 0 {
					age := currentSession - f.PlantedSession
					display += fmt.Sprintf(" (%d sessions ago)", age)
				}

				if f.Status == adventure.ForeshadowResolved && f.ResolutionNotes != "" {
					display += fmt.Sprintf("\n   Resolution: %s", f.ResolutionNotes)
				}

				if f.Context != "" {
					display += fmt.Sprintf("\n   Context: %s", f.Context)
				}

				if len(f.RelatedNPCs) > 0 {
					display += fmt.Sprintf("\n   NPCs: %s", strings.Join(f.RelatedNPCs, ", "))
				}

				if len(f.Tags) > 0 {
					display += fmt.Sprintf("\n   Tags: %s", strings.Join(f.Tags, ", "))
				}

				display += "\n\n"
			}

			return map[string]interface{}{
				"success":     true,
				"count":       len(foreshadows),
				"foreshadows": formatForeshadowsSummary(foreshadows),
				"display":     display,
			}, nil
		},
	}
}

// NewGetStaleForeshadowsTool creates a tool to alert about old unresolved foreshadows.
func NewGetStaleForeshadowsTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "get_stale_foreshadows",
		description: "Get foreshadows that have been unresolved for many sessions. Use this at session start to remind yourself of old narrative threads that need attention. Default threshold: 3 sessions.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"max_age": map[string]interface{}{
					"type":        "number",
					"description": "Maximum age in sessions before considering stale. Default: 3",
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			maxAge := 3
			if age, ok := params["max_age"].(float64); ok {
				maxAge = int(age)
			}

			staleForeshadows, err := adv.GetStaleForeshadows(maxAge)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to get stale foreshadows: %v", err),
				}, nil
			}

			if len(staleForeshadows) == 0 {
				return map[string]interface{}{
					"success": true,
					"count":   0,
					"display": fmt.Sprintf("✓ Aucun foreshadow en attente depuis plus de %d sessions.", maxAge),
				}, nil
			}

			// Get current session
			var currentSession int
			if session, err := adv.GetCurrentSession(); err == nil && session != nil {
				currentSession = session.ID
			}

			display := fmt.Sprintf("⚠️  ALERTE: %d foreshadow(s) en attente depuis plus de %d sessions:\n\n", len(staleForeshadows), maxAge)

			for i, f := range staleForeshadows {
				age := currentSession - f.PlantedSession
				display += fmt.Sprintf("%d. [%s] %s\n", i+1, f.ID, f.Description)
				display += fmt.Sprintf("   Importance: %s | Category: %s\n", f.Importance, f.Category)
				display += fmt.Sprintf("   Planté session %d (%d sessions ago)\n", f.PlantedSession, age)

				if f.Context != "" {
					display += fmt.Sprintf("   Context: %s\n", f.Context)
				}

				display += "\n"
			}

			display += "💡 Suggestion: Considère intégrer ces éléments dans la narration de cette session."

			return map[string]interface{}{
				"success":     true,
				"count":       len(staleForeshadows),
				"foreshadows": formatForeshadowsSummary(staleForeshadows),
				"display":     display,
			}, nil
		},
	}
}

// Helper to format foreshadows summary for tool response
func formatForeshadowsSummary(foreshadows []adventure.Foreshadow) []map[string]interface{} {
	var result []map[string]interface{}
	for _, f := range foreshadows {
		result = append(result, map[string]interface{}{
			"id":             f.ID,
			"description":    f.Description,
			"importance":     string(f.Importance),
			"category":       string(f.Category),
			"status":         string(f.Status),
			"planted_session": f.PlantedSession,
		})
	}
	return result
}