package dmtools

import (
	"fmt"

	"dungeons/internal/adventure"
)

// NewUpdateTimeTool creates a tool to update the in-game time.
func NewUpdateTimeTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "update_time",
		description: "Update the in-game time. Use this when time passes in the narrative (rest, travel, waiting). The time is tracked as day/hour/minute.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"day": map[string]interface{}{
					"type":        "integer",
					"description": "Day number (1 = first day of adventure)",
				},
				"hour": map[string]interface{}{
					"type":        "integer",
					"description": "Hour (0-23)",
					"minimum":     0,
					"maximum":     23,
				},
				"minute": map[string]interface{}{
					"type":        "integer",
					"description": "Minute (0-59)",
					"minimum":     0,
					"maximum":     59,
				},
			},
			"required": []string{"day", "hour"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			day := int(params["day"].(float64))
			hour := int(params["hour"].(float64))
			minute := 0
			if m, ok := params["minute"].(float64); ok {
				minute = int(m)
			}

			// Load current state
			state, err := adv.LoadState()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load state: %v", err),
				}, nil
			}

			oldTime := fmt.Sprintf("Jour %d, %02d:%02d", state.Time.Day, state.Time.Hour, state.Time.Minute)
			state.Time.Day = day
			state.Time.Hour = hour
			state.Time.Minute = minute
			newTime := fmt.Sprintf("Jour %d, %02d:%02d", day, hour, minute)

			// Save updated state
			if err := adv.SaveState(state); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to save state: %v", err),
				}, nil
			}

			return map[string]interface{}{
				"success":  true,
				"old_time": oldTime,
				"new_time": newTime,
				"display":  fmt.Sprintf("✓ Temps mis à jour: %s → %s", oldTime, newTime),
			}, nil
		},
	}
}

// NewSetFlagTool creates a tool to set or unset narrative flags.
func NewSetFlagTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "set_flag",
		description: "Set or unset a narrative flag. Flags track important story events (e.g., 'defeated_boss', 'found_secret_door', 'allied_with_faction'). Use snake_case for flag names.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"flag": map[string]interface{}{
					"type":        "string",
					"description": "Flag name in snake_case (e.g., 'defeated_possessed_creature', 'explored_crypt')",
				},
				"value": map[string]interface{}{
					"type":        "boolean",
					"description": "True to set the flag, false to unset it",
					"default":     true,
				},
			},
			"required": []string{"flag"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			flag := params["flag"].(string)
			value := true
			if v, ok := params["value"].(bool); ok {
				value = v
			}

			// Load current state
			state, err := adv.LoadState()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load state: %v", err),
				}, nil
			}

			// Initialize flags map if nil
			if state.Flags == nil {
				state.Flags = make(map[string]bool)
			}

			oldValue, existed := state.Flags[flag]
			state.Flags[flag] = value

			// Save updated state
			if err := adv.SaveState(state); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to save state: %v", err),
				}, nil
			}

			var display string
			if value {
				if existed {
					display = fmt.Sprintf("✓ Flag '%s' confirmé (était déjà %v)", flag, oldValue)
				} else {
					display = fmt.Sprintf("✓ Flag '%s' activé", flag)
				}
			} else {
				display = fmt.Sprintf("✓ Flag '%s' désactivé", flag)
			}

			return map[string]interface{}{
				"success":   true,
				"flag":      flag,
				"value":     value,
				"old_value": oldValue,
				"existed":   existed,
				"display":   display,
			}, nil
		},
	}
}

// NewAddQuestTool creates a tool to add a new quest.
func NewAddQuestTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "add_quest",
		description: "Add a new quest or objective to track. Quests help maintain narrative focus and can be marked as completed later.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Short quest name (e.g., 'Trouver Thomas Brenner')",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Detailed quest description",
				},
			},
			"required": []string{"name", "description"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			name := params["name"].(string)
			description := params["description"].(string)

			// Load current state
			state, err := adv.LoadState()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load state: %v", err),
				}, nil
			}

			// Check if quest already exists
			for _, q := range state.Quests {
				if q.Name == name {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("Quest '%s' already exists", name),
					}, nil
				}
			}

			// Add new quest
			newQuest := adventure.Quest{
				Name:        name,
				Description: description,
				Status:      "active",
			}
			state.Quests = append(state.Quests, newQuest)

			// Save updated state
			if err := adv.SaveState(state); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to save state: %v", err),
				}, nil
			}

			return map[string]interface{}{
				"success": true,
				"quest":   newQuest,
				"display": fmt.Sprintf("✓ Nouvelle quête ajoutée: %s", name),
			}, nil
		},
	}
}

// NewCompleteQuestTool creates a tool to mark a quest as completed.
func NewCompleteQuestTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "complete_quest",
		description: "Mark a quest as completed. Use when players achieve a quest objective.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"quest_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the quest to complete",
				},
			},
			"required": []string{"quest_name"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			questName := params["quest_name"].(string)

			// Load current state
			state, err := adv.LoadState()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load state: %v", err),
				}, nil
			}

			// Find and update quest
			found := false
			for i, q := range state.Quests {
				if q.Name == questName {
					if q.Status == "completed" {
						return map[string]interface{}{
							"success": false,
							"error":   fmt.Sprintf("Quest '%s' is already completed", questName),
						}, nil
					}
					state.Quests[i].Status = "completed"
					found = true
					break
				}
			}

			if !found {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Quest '%s' not found", questName),
				}, nil
			}

			// Save updated state
			if err := adv.SaveState(state); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to save state: %v", err),
				}, nil
			}

			return map[string]interface{}{
				"success": true,
				"display": fmt.Sprintf("✓ Quête complétée: %s", questName),
			}, nil
		},
	}
}

// NewSetVariableTool creates a tool to set narrative variables.
func NewSetVariableTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "set_variable",
		description: "Set a narrative variable for tracking story elements (e.g., 'current_inn', 'allied_faction', 'villain_name'). Variables store string values.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Variable name in snake_case",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Variable value (use empty string to delete)",
				},
			},
			"required": []string{"key", "value"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			key := params["key"].(string)
			value := params["value"].(string)

			// Load current state
			state, err := adv.LoadState()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load state: %v", err),
				}, nil
			}

			// Initialize variables map if nil
			if state.Variables == nil {
				state.Variables = make(map[string]string)
			}

			oldValue, existed := state.Variables[key]

			if value == "" {
				// Delete variable
				delete(state.Variables, key)
			} else {
				state.Variables[key] = value
			}

			// Save updated state
			if err := adv.SaveState(state); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to save state: %v", err),
				}, nil
			}

			var display string
			if value == "" {
				display = fmt.Sprintf("✓ Variable '%s' supprimée", key)
			} else if existed {
				display = fmt.Sprintf("✓ Variable '%s' mise à jour: %s → %s", key, oldValue, value)
			} else {
				display = fmt.Sprintf("✓ Variable '%s' définie: %s", key, value)
			}

			return map[string]interface{}{
				"success":   true,
				"key":       key,
				"value":     value,
				"old_value": oldValue,
				"display":   display,
			}, nil
		},
	}
}

// NewGetStateTool creates a tool to get the current game state.
func NewGetStateTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "get_state",
		description: "Get the current game state including location, time, quests, flags, and variables. Useful for checking current narrative status.",
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			state, err := adv.LoadState()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load state: %v", err),
				}, nil
			}

			// Build display
			display := fmt.Sprintf("## État du Jeu\n\n")
			display += fmt.Sprintf("**Lieu**: %s\n", state.CurrentLocation)
			display += fmt.Sprintf("**Temps**: Jour %d, %02d:%02d\n\n", state.Time.Day, state.Time.Hour, state.Time.Minute)

			if len(state.Quests) > 0 {
				display += "**Quêtes**:\n"
				for _, q := range state.Quests {
					status := "🔵"
					if q.Status == "completed" {
						status = "✅"
					}
					display += fmt.Sprintf("  %s %s\n", status, q.Name)
				}
				display += "\n"
			}

			if len(state.Flags) > 0 {
				display += "**Flags actifs**:\n"
				for flag, value := range state.Flags {
					if value {
						display += fmt.Sprintf("  • %s\n", flag)
					}
				}
				display += "\n"
			}

			if len(state.Variables) > 0 {
				display += "**Variables**:\n"
				for key, value := range state.Variables {
					display += fmt.Sprintf("  • %s = %s\n", key, value)
				}
			}

			return map[string]interface{}{
				"success":  true,
				"state":    state,
				"display":  display,
				"location": state.CurrentLocation,
				"time": map[string]int{
					"day":    state.Time.Day,
					"hour":   state.Time.Hour,
					"minute": state.Time.Minute,
				},
				"quests":    state.Quests,
				"flags":     state.Flags,
				"variables": state.Variables,
			}, nil
		},
	}
}
