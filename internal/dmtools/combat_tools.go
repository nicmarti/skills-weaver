package dmtools

import (
	"fmt"
	"path/filepath"
	"strings"

	"dungeons/internal/adventure"
)

// NewUpdateHPTool creates a tool to modify a character's HP during combat.
// Use negative values for damage, positive for healing.
func NewUpdateHPTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "update_hp",
		description: "Modifier les PV d'un personnage (dégâts ou soins). Utilise un nombre négatif pour les dégâts, positif pour les soins. Gère automatiquement les limites (0 minimum, max_hp maximum).",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"character_name": map[string]interface{}{
					"type":        "string",
					"description": "Nom du personnage (insensible à la casse)",
				},
				"amount": map[string]interface{}{
					"type":        "integer",
					"description": "Modification des PV: négatif pour dégâts (ex: -8), positif pour soins (ex: +5)",
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Raison du changement (ex: 'Attaque de gobelin', 'Sort de soins', 'Poison')",
				},
			},
			"required": []string{"character_name", "amount"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			name, ok := params["character_name"].(string)
			if !ok || name == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "character_name est requis",
				}, nil
			}

			amountFloat, ok := params["amount"].(float64)
			if !ok {
				return map[string]interface{}{
					"success": false,
					"error":   "amount doit être un nombre entier",
				}, nil
			}
			amount := int(amountFloat)

			reason := ""
			if r, ok := params["reason"].(string); ok {
				reason = r
			}

			// Load characters
			characters, err := adv.GetCharacters()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur de chargement des personnages: %v", err),
				}, nil
			}

			// Find character (case-insensitive)
			nameLower := strings.ToLower(name)
			var foundIndex int = -1
			for i, char := range characters {
				if strings.ToLower(char.Name) == nameLower {
					foundIndex = i
					break
				}
			}

			if foundIndex == -1 {
				available := []string{}
				for _, char := range characters {
					available = append(available, char.Name)
				}
				return map[string]interface{}{
					"success":   false,
					"error":     fmt.Sprintf("Personnage '%s' non trouvé", name),
					"available": available,
				}, nil
			}

			char := characters[foundIndex]
			oldHP := char.HitPoints

			// Apply modification
			char.HitPoints += amount

			// Clamp to valid range
			if char.HitPoints < 0 {
				char.HitPoints = 0
			}
			if char.HitPoints > char.MaxHitPoints {
				char.HitPoints = char.MaxHitPoints
			}

			newHP := char.HitPoints
			actualChange := newHP - oldHP

			// Save the character
			charDir := filepath.Join(adv.BasePath(), "characters")
			if err := char.Save(charDir); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur de sauvegarde: %v", err),
				}, nil
			}

			// Determine status and display
			var status string
			var display string

			if newHP == 0 {
				status = "inconscient"
			} else if float64(newHP) <= float64(char.MaxHitPoints)*0.25 {
				status = "gravement blessé"
			} else if float64(newHP) <= float64(char.MaxHitPoints)*0.5 {
				status = "blessé"
			} else {
				status = "en forme"
			}

			if actualChange < 0 {
				display = fmt.Sprintf("✓ %s subit %d dégâts", char.Name, -actualChange)
			} else if actualChange > 0 {
				display = fmt.Sprintf("✓ %s récupère %d PV", char.Name, actualChange)
			} else {
				display = fmt.Sprintf("✓ %s: PV inchangés", char.Name)
			}

			display += fmt.Sprintf(" (PV: %d/%d - %s)", newHP, char.MaxHitPoints, status)

			if reason != "" {
				display += fmt.Sprintf(" [%s]", reason)
			}

			// Log event
			var logType string
			var logContent string
			if actualChange < 0 {
				logType = "combat"
				logContent = fmt.Sprintf("%s: %d dégâts", char.Name, -actualChange)
				if reason != "" {
					logContent += fmt.Sprintf(" (%s)", reason)
				}
				logContent += fmt.Sprintf(" - PV: %d/%d", newHP, char.MaxHitPoints)
			} else if actualChange > 0 {
				logType = "story"
				logContent = fmt.Sprintf("%s récupère %d PV", char.Name, actualChange)
				if reason != "" {
					logContent += fmt.Sprintf(" (%s)", reason)
				}
				logContent += fmt.Sprintf(" - PV: %d/%d", newHP, char.MaxHitPoints)
			}

			if logContent != "" {
				adv.LogEvent(logType, logContent)
			}

			return map[string]interface{}{
				"success":        true,
				"character_name": char.Name,
				"old_hp":         oldHP,
				"new_hp":         newHP,
				"max_hp":         char.MaxHitPoints,
				"actual_change":  actualChange,
				"status":         status,
				"display":        display,
			}, nil
		},
	}
}

// NewUseSpellSlotTool creates a tool to consume a spell slot.
func NewUseSpellSlotTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "use_spell_slot",
		description: "Consommer un emplacement de sort. Le personnage doit avoir un emplacement disponible au niveau indiqué. Appeler ce tool AVANT de résoudre l'effet du sort.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"character_name": map[string]interface{}{
					"type":        "string",
					"description": "Nom du lanceur de sorts (insensible à la casse)",
				},
				"spell_level": map[string]interface{}{
					"type":        "integer",
					"description": "Niveau de l'emplacement à utiliser (1-9). Utiliser le niveau du sort ou plus haut pour upcast.",
					"minimum":     1,
					"maximum":     9,
				},
				"spell_name": map[string]interface{}{
					"type":        "string",
					"description": "Nom du sort lancé (optionnel, pour le journal)",
				},
			},
			"required": []string{"character_name", "spell_level"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			name, ok := params["character_name"].(string)
			if !ok || name == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "character_name est requis",
				}, nil
			}

			levelFloat, ok := params["spell_level"].(float64)
			if !ok {
				return map[string]interface{}{
					"success": false,
					"error":   "spell_level doit être un nombre entier entre 1 et 9",
				}, nil
			}
			level := int(levelFloat)

			if level < 1 || level > 9 {
				return map[string]interface{}{
					"success": false,
					"error":   "spell_level doit être entre 1 et 9",
				}, nil
			}

			spellName := ""
			if s, ok := params["spell_name"].(string); ok {
				spellName = s
			}

			// Load characters
			characters, err := adv.GetCharacters()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur de chargement des personnages: %v", err),
				}, nil
			}

			// Find character (case-insensitive)
			nameLower := strings.ToLower(name)
			var foundIndex int = -1
			for i, char := range characters {
				if strings.ToLower(char.Name) == nameLower {
					foundIndex = i
					break
				}
			}

			if foundIndex == -1 {
				available := []string{}
				for _, char := range characters {
					available = append(available, char.Name)
				}
				return map[string]interface{}{
					"success":   false,
					"error":     fmt.Sprintf("Personnage '%s' non trouvé", name),
					"available": available,
				}, nil
			}

			char := characters[foundIndex]

			// Check if character has spell slots at all
			if char.SpellSlots == nil || len(char.SpellSlots) == 0 {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("%s n'est pas un lanceur de sorts", char.Name),
				}, nil
			}

			// Get available slots before using
			availableBefore := char.GetAvailableSlots(level)
			maxSlots := char.SpellSlots[level]

			// Use the spell slot (this checks availability internally)
			if err := char.UseSpellSlot(level); err != nil {
				return map[string]interface{}{
					"success":          false,
					"error":            fmt.Sprintf("Impossible d'utiliser l'emplacement: %v", err),
					"available_slots":  formatAvailableSlotsFromMaps(char.SpellSlots, char.SpellSlotsUsed),
				}, nil
			}

			availableAfter := char.GetAvailableSlots(level)

			// Save the character
			charDir := filepath.Join(adv.BasePath(), "characters")
			if err := char.Save(charDir); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur de sauvegarde: %v", err),
				}, nil
			}

			// Build display
			display := fmt.Sprintf("✓ %s utilise un emplacement de niveau %d", char.Name, level)
			if spellName != "" {
				display += fmt.Sprintf(" pour lancer %s", spellName)
			}
			display += fmt.Sprintf(" (reste: %d/%d)", availableAfter, maxSlots)

			// Log event
			logContent := fmt.Sprintf("%s utilise emplacement N%d", char.Name, level)
			if spellName != "" {
				logContent += fmt.Sprintf(" (%s)", spellName)
			}
			adv.LogEvent("combat", logContent)

			return map[string]interface{}{
				"success":          true,
				"character_name":   char.Name,
				"spell_level":      level,
				"spell_name":       spellName,
				"slots_before":     availableBefore,
				"slots_after":      availableAfter,
				"max_slots":        maxSlots,
				"all_slots":        formatAvailableSlotsFromMaps(char.SpellSlots, char.SpellSlotsUsed),
				"display":          display,
			}, nil
		},
	}
}

// formatAvailableSlotsFromMaps returns a formatted string of all spell slots.
func formatAvailableSlotsFromMaps(spellSlots, spellSlotsUsed map[int]int) string {
	if spellSlots == nil {
		return "Aucun emplacement"
	}

	slots := []string{}
	for lvl := 1; lvl <= 9; lvl++ {
		max, exists := spellSlots[lvl]
		if !exists || max == 0 {
			continue
		}
		used := 0
		if spellSlotsUsed != nil {
			used = spellSlotsUsed[lvl]
		}
		available := max - used
		if available < 0 {
			available = 0
		}
		slots = append(slots, fmt.Sprintf("N%d: %d/%d", lvl, available, max))
	}

	if len(slots) == 0 {
		return "Aucun emplacement"
	}
	return strings.Join(slots, ", ")
}
