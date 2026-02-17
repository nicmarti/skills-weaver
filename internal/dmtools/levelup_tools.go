package dmtools

import (
	"fmt"
	"path/filepath"
	"strings"

	"dungeons/internal/adventure"
	"dungeons/internal/data"
)

// NewUpdateCharacterStatTool creates a tool to update character stats during level-up or ASI.
// Supports updating ability scores (auto-recalculates modifiers), max_hp, armor_class,
// spell_save_dc, and spell_attack_bonus.
func NewUpdateCharacterStatTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name: "update_character_stat",
		description: `Modifier une statistique d'un personnage (montée de niveau, ASI, effets magiques).
Statistiques modifiables :
- ability scores: "strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma" (recalcule automatiquement le modificateur)
- "max_hp" : points de vie maximum (met aussi à jour les PV courants si heal_to_max est true)
- "armor_class" : classe d'armure
- "spell_save_dc" : DD des sorts
- "spell_attack_bonus" : bonus d'attaque des sorts
Utiliser ce tool pour les montées de niveau (ASI) et les effets permanents.`,
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"character_name": map[string]interface{}{
					"type":        "string",
					"description": "Nom du personnage (insensible à la casse)",
				},
				"stat": map[string]interface{}{
					"type":        "string",
					"description": "Statistique à modifier",
					"enum":        []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma", "max_hp", "armor_class", "spell_save_dc", "spell_attack_bonus"},
				},
				"value": map[string]interface{}{
					"type":        "integer",
					"description": "Nouvelle valeur absolue de la statistique (ex: 17 pour FOR, 31 pour max_hp)",
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Raison du changement (ex: 'ASI niveau 4', 'Tome of Clear Thought')",
				},
				"heal_to_max": map[string]interface{}{
					"type":        "boolean",
					"description": "Si true et stat=max_hp, met aussi les PV courants au nouveau maximum (utile après repos long + level up)",
				},
			},
			"required": []string{"character_name", "stat", "value"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			name, ok := params["character_name"].(string)
			if !ok || name == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "character_name est requis",
				}, nil
			}

			stat, ok := params["stat"].(string)
			if !ok || stat == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "stat est requis",
				}, nil
			}

			valueFloat, ok := params["value"].(float64)
			if !ok {
				return map[string]interface{}{
					"success": false,
					"error":   "value doit être un nombre entier",
				}, nil
			}
			value := int(valueFloat)

			reason := ""
			if r, ok := params["reason"].(string); ok {
				reason = r
			}

			healToMax := false
			if h, ok := params["heal_to_max"].(bool); ok {
				healToMax = h
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
			var oldValue int
			var displayStat string

			switch stat {
			case "strength":
				oldValue = char.Abilities.Strength
				char.Abilities.Strength = value
				char.Modifiers.Strength = data.AbilityModifier(value)
				displayStat = "FOR"
			case "dexterity":
				oldValue = char.Abilities.Dexterity
				char.Abilities.Dexterity = value
				char.Modifiers.Dexterity = data.AbilityModifier(value)
				displayStat = "DEX"
			case "constitution":
				oldValue = char.Abilities.Constitution
				char.Abilities.Constitution = value
				char.Modifiers.Constitution = data.AbilityModifier(value)
				displayStat = "CON"
			case "intelligence":
				oldValue = char.Abilities.Intelligence
				char.Abilities.Intelligence = value
				char.Modifiers.Intelligence = data.AbilityModifier(value)
				displayStat = "INT"
			case "wisdom":
				oldValue = char.Abilities.Wisdom
				char.Abilities.Wisdom = value
				char.Modifiers.Wisdom = data.AbilityModifier(value)
				displayStat = "SAG"
			case "charisma":
				oldValue = char.Abilities.Charisma
				char.Abilities.Charisma = value
				char.Modifiers.Charisma = data.AbilityModifier(value)
				displayStat = "CHA"
			case "max_hp":
				oldValue = char.MaxHitPoints
				char.MaxHitPoints = value
				if healToMax {
					char.HitPoints = value
				}
				displayStat = "PV Max"
			case "armor_class":
				oldValue = char.ArmorClass
				char.ArmorClass = value
				displayStat = "CA"
			case "spell_save_dc":
				oldValue = char.SpellSaveDC
				char.SpellSaveDC = value
				displayStat = "DD Sorts"
			case "spell_attack_bonus":
				oldValue = char.SpellAttackBonus
				char.SpellAttackBonus = value
				displayStat = "Attaque Sorts"
			default:
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Statistique inconnue: %s", stat),
				}, nil
			}

			// Save the character
			charDir := filepath.Join(adv.BasePath(), "characters")
			if err := char.Save(charDir); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur de sauvegarde: %v", err),
				}, nil
			}

			// Build display
			newModifier := ""
			if isAbilityScore(stat) {
				mod := data.AbilityModifier(value)
				newModifier = fmt.Sprintf(" (mod %+d)", mod)
			}

			display := fmt.Sprintf("✓ %s: %s %d → %d%s", char.Name, displayStat, oldValue, value, newModifier)
			if stat == "max_hp" && healToMax {
				display += fmt.Sprintf(" (PV: %d/%d)", char.HitPoints, char.MaxHitPoints)
			}
			if reason != "" {
				display += fmt.Sprintf(" [%s]", reason)
			}

			// Log event
			logContent := fmt.Sprintf("%s: %s %d → %d", char.Name, displayStat, oldValue, value)
			if reason != "" {
				logContent += fmt.Sprintf(" (%s)", reason)
			}
			adv.LogEvent("story", logContent)

			result := map[string]interface{}{
				"success":        true,
				"character_name": char.Name,
				"stat":           stat,
				"old_value":      oldValue,
				"new_value":      value,
				"display":        display,
			}

			if isAbilityScore(stat) {
				result["new_modifier"] = data.AbilityModifier(value)
			}

			return result, nil
		},
	}
}

// NewLongRestTool creates a tool to apply a long rest to the party or a character.
// Restores spell slots, heals to max HP, and restores hit dice.
func NewLongRestTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "long_rest",
		description: "Appliquer un repos long (8h) à tout le groupe ou un personnage. Restaure les PV au maximum, remet tous les emplacements de sorts à 0 utilisé, et restaure les dés de vie (la moitié du niveau, minimum 1). Appeler ce tool quand le groupe se repose 8 heures.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"character_name": map[string]interface{}{
					"type":        "string",
					"description": "Nom du personnage (optionnel - si omis, applique à tout le groupe)",
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Load characters
			characters, err := adv.GetCharacters()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur de chargement des personnages: %v", err),
				}, nil
			}

			// Determine which characters to rest
			targetName := ""
			if n, ok := params["character_name"].(string); ok && n != "" {
				targetName = strings.ToLower(n)
			}

			var results []map[string]interface{}
			var displayParts []string
			charDir := filepath.Join(adv.BasePath(), "characters")

			for _, char := range characters {
				// Filter if a specific character is requested
				if targetName != "" && strings.ToLower(char.Name) != targetName {
					continue
				}

				oldHP := char.HitPoints
				changes := []string{}

				// Restore HP to max
				if char.HitPoints < char.MaxHitPoints {
					char.HitPoints = char.MaxHitPoints
					changes = append(changes, fmt.Sprintf("PV %d → %d/%d", oldHP, char.HitPoints, char.MaxHitPoints))
				}

				// Restore spell slots
				slotsRestored := false
				if char.SpellSlotsUsed != nil {
					for level, used := range char.SpellSlotsUsed {
						if used > 0 {
							slotsRestored = true
							break
						}
						_ = level
					}
					if slotsRestored {
						char.RestoreSpellSlots()
						changes = append(changes, "emplacements de sorts restaurés")
					}
				}

				// Restore hit dice (D&D 5e: regain half level, minimum 1)
				if char.HitDice < char.MaxHitDice {
					regain := char.MaxHitDice / 2
					if regain < 1 {
						regain = 1
					}
					oldDice := char.HitDice
					char.HitDice += regain
					if char.HitDice > char.MaxHitDice {
						char.HitDice = char.MaxHitDice
					}
					if char.HitDice != oldDice {
						changes = append(changes, fmt.Sprintf("dés de vie %d → %d/%d", oldDice, char.HitDice, char.MaxHitDice))
					}
				}

				// Save character
				if err := char.Save(charDir); err != nil {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("Erreur de sauvegarde %s: %v", char.Name, err),
					}, nil
				}

				charResult := map[string]interface{}{
					"name":   char.Name,
					"hp":     fmt.Sprintf("%d/%d", char.HitPoints, char.MaxHitPoints),
					"healed": char.HitPoints != oldHP,
				}

				if char.SpellSlots != nil && len(char.SpellSlots) > 0 {
					charResult["spell_slots"] = formatAvailableSlotsFromMaps(char.SpellSlots, char.SpellSlotsUsed)
				}

				results = append(results, charResult)

				if len(changes) > 0 {
					displayParts = append(displayParts, fmt.Sprintf("  • %s: %s", char.Name, strings.Join(changes, ", ")))
				} else {
					displayParts = append(displayParts, fmt.Sprintf("  • %s: déjà en forme", char.Name))
				}
			}

			if len(results) == 0 && targetName != "" {
				available := []string{}
				for _, char := range characters {
					available = append(available, char.Name)
				}
				return map[string]interface{}{
					"success":   false,
					"error":     fmt.Sprintf("Personnage '%s' non trouvé", targetName),
					"available": available,
				}, nil
			}

			display := "✓ Repos long terminé (8h)\n" + strings.Join(displayParts, "\n")

			// Log event
			adv.LogEvent("story", "Repos long (8h) - Groupe entièrement restauré")

			return map[string]interface{}{
				"success":    true,
				"characters": results,
				"display":    display,
			}, nil
		},
	}
}

// isAbilityScore returns true if the stat name is a D&D ability score.
func isAbilityScore(stat string) bool {
	switch stat {
	case "strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma":
		return true
	default:
		return false
	}
}
