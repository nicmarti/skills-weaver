package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
)

// NewGetPartyInfoTool creates a tool to get an overview of the party with combat-relevant stats.
func NewGetPartyInfoTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "get_party_info",
		description: "Get an overview of the party with combat-relevant stats (HP, AC, level, proficiency bonus, speed, primary abilities, skills). Use this to quickly check the group's status during combat or when planning encounters.",
		schema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Load party
			party, err := adv.LoadParty()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load party: %v", err),
				}, nil
			}

			// Load characters
			characters, err := adv.GetCharacters()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load characters: %v", err),
				}, nil
			}

			// Build party info
			members := []map[string]interface{}{}
			for _, charName := range party.Characters {
				for _, char := range characters {
					if char.Name == charName {
						// Find primary stat (highest modifier)
						primaryStat := findPrimaryStat(char)

						// Calculate proficiency bonus (D&D 5e)
						profBonus := 2 + ((char.Level - 1) / 4)

						member := map[string]interface{}{
							"name":              char.Name,
							"species":           char.Species,
							"class":             char.Class,
							"level":             char.Level,
							"hp":                char.HitPoints,
							"max_hp":            char.MaxHitPoints,
							"ac":                char.ArmorClass,
							"proficiency_bonus": profBonus,
							"speed":             30, // Default 30 feet for most species
							"primary_stat":      primaryStat,
						}
						members = append(members, member)
						break
					}
				}
			}

			// Build display string
			display := formatPartyDisplay(party, members)

			return map[string]interface{}{
				"success": true,
				"party": map[string]interface{}{
					"formation":      party.Formation,
					"marching_order": party.MarchingOrder,
					"members":        members,
				},
				"display": display,
			}, nil
		},
	}
}

// NewGetCharacterInfoTool creates a tool to get detailed info about a specific character.
func NewGetCharacterInfoTool(adv *adventure.Adventure) *SimpleTool {
	return &SimpleTool{
		name:        "get_character_info",
		description: "Get detailed information about a specific character including abilities, modifiers, proficiency bonus, skills, spell save DC (if caster), equipment, and appearance. Use this for skill checks, combat bonuses, spell casting, or roleplay details.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "The character's name (case-insensitive)",
				},
			},
			"required": []string{"name"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			name, ok := params["name"].(string)
			if !ok || name == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "Character name is required",
				}, nil
			}

			// Load characters
			characters, err := adv.GetCharacters()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Failed to load characters: %v", err),
				}, nil
			}

			// Find character (case-insensitive)
			var found *character.Character
			nameLower := strings.ToLower(name)
			for _, char := range characters {
				if strings.ToLower(char.Name) == nameLower {
					found = char
					break
				}
			}

			if found == nil {
				// List available characters for helpful error
				available := []string{}
				for _, char := range characters {
					available = append(available, char.Name)
				}
				return map[string]interface{}{
					"success":   false,
					"error":     fmt.Sprintf("Character '%s' not found in party", name),
					"available": available,
				}, nil
			}

			// Build character info
			charInfo := buildCharacterInfo(found)
			display := formatCharacterDisplay(found)

			return map[string]interface{}{
				"success":   true,
				"character": charInfo,
				"display":   display,
			}, nil
		},
	}
}

// findPrimaryStat returns the highest ability and its modifier.
func findPrimaryStat(char *character.Character) map[string]interface{} {
	abilities := map[string]int{
		"Strength":     char.Abilities.Strength,
		"Intelligence": char.Abilities.Intelligence,
		"Wisdom":       char.Abilities.Wisdom,
		"Dexterity":    char.Abilities.Dexterity,
		"Constitution": char.Abilities.Constitution,
		"Charisma":     char.Abilities.Charisma,
	}
	modifiers := map[string]int{
		"Strength":     char.Modifiers.Strength,
		"Intelligence": char.Modifiers.Intelligence,
		"Wisdom":       char.Modifiers.Wisdom,
		"Dexterity":    char.Modifiers.Dexterity,
		"Constitution": char.Modifiers.Constitution,
		"Charisma":     char.Modifiers.Charisma,
	}

	// Find highest value
	maxName := "Strength"
	maxVal := abilities["Strength"]
	for name, val := range abilities {
		if val > maxVal {
			maxVal = val
			maxName = name
		}
	}

	return map[string]interface{}{
		"name":     maxName,
		"value":    maxVal,
		"modifier": modifiers[maxName],
	}
}

// buildCharacterInfo builds a complete character info map.
func buildCharacterInfo(char *character.Character) map[string]interface{} {
	// Calculate proficiency bonus (D&D 5e)
	profBonus := 2 + ((char.Level - 1) / 4)

	info := map[string]interface{}{
		"name":              char.Name,
		"species":           char.Species,
		"class":             char.Class,
		"level":             char.Level,
		"xp":                char.XP,
		"hp":                char.HitPoints,
		"max_hp":            char.MaxHitPoints,
		"ac":                char.ArmorClass,
		"gold":              char.Gold,
		"proficiency_bonus": profBonus,
		"speed":             30, // Default 30 feet
		"abilities": map[string]int{
			"strength":     char.Abilities.Strength,
			"intelligence": char.Abilities.Intelligence,
			"wisdom":       char.Abilities.Wisdom,
			"dexterity":    char.Abilities.Dexterity,
			"constitution": char.Abilities.Constitution,
			"charisma":     char.Abilities.Charisma,
		},
		"modifiers": map[string]int{
			"strength":     char.Modifiers.Strength,
			"intelligence": char.Modifiers.Intelligence,
			"wisdom":       char.Modifiers.Wisdom,
			"dexterity":    char.Modifiers.Dexterity,
			"constitution": char.Modifiers.Constitution,
			"charisma":     char.Modifiers.Charisma,
		},
		"equipment": char.Equipment,
	}

	// Add D&D 5e specific fields
	if char.Background != "" {
		info["background"] = char.Background
	}
	if char.Skills != nil && len(char.Skills) > 0 {
		// Extract proficient skills
		proficientSkills := []string{}
		for skill, proficient := range char.Skills {
			if proficient {
				proficientSkills = append(proficientSkills, skill)
			}
		}
		if len(proficientSkills) > 0 {
			info["skills"] = proficientSkills
		}
	}

	// Add spell save DC and attack bonus if character has spell slots (is a caster)
	if char.SpellSlots != nil && len(char.SpellSlots) > 0 {
		// Determine spellcasting ability based on class
		// Cleric, Druid, Ranger: WIS
		// Wizard: INT
		// Sorcerer, Bard, Paladin, Warlock: CHA
		spellMod := 0
		switch char.Class {
		case "cleric", "druide", "ranger", "rôdeur":
			spellMod = char.Modifiers.Wisdom
		case "wizard", "magicien":
			spellMod = char.Modifiers.Intelligence
		case "sorcerer", "ensorceleur", "bard", "barde", "paladin", "warlock", "occultiste":
			spellMod = char.Modifiers.Charisma
		}
		info["spell_save_dc"] = 8 + profBonus + spellMod
		info["spell_attack_bonus"] = profBonus + spellMod
	}

	// Add spells if applicable
	if len(char.KnownSpells) > 0 {
		info["known_spells"] = char.KnownSpells
	}
	if len(char.PreparedSpells) > 0 {
		info["prepared_spells"] = char.PreparedSpells
	}
	if char.SpellSlots != nil && len(char.SpellSlots) > 0 {
		info["spell_slots"] = char.SpellSlots
		info["spell_slots_used"] = char.SpellSlotsUsed
	}

	// Add appearance if available
	if char.Appearance != nil {
		appearance := map[string]interface{}{}
		if char.Appearance.Age > 0 {
			appearance["age"] = char.Appearance.Age
		}
		if char.Appearance.Gender != "" {
			appearance["gender"] = char.Appearance.Gender
		}
		if char.Appearance.Build != "" {
			appearance["build"] = char.Appearance.Build
		}
		if char.Appearance.Height != "" {
			appearance["height"] = char.Appearance.Height
		}
		if char.Appearance.HairColor != "" {
			appearance["hair_color"] = char.Appearance.HairColor
		}
		if char.Appearance.EyeColor != "" {
			appearance["eye_color"] = char.Appearance.EyeColor
		}
		if char.Appearance.DistinctiveFeature != "" {
			appearance["distinctive_feature"] = char.Appearance.DistinctiveFeature
		}
		if char.Appearance.ArmorDescription != "" {
			appearance["armor_description"] = char.Appearance.ArmorDescription
		}
		if char.Appearance.WeaponDescription != "" {
			appearance["weapon_description"] = char.Appearance.WeaponDescription
		}
		if len(appearance) > 0 {
			info["appearance"] = appearance
		}
	}

	return info
}

// formatPartyDisplay formats party info for display.
func formatPartyDisplay(party *adventure.Party, members []map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("## Groupe\n\n")
	sb.WriteString(fmt.Sprintf("**Formation**: %s\n", party.Formation))
	sb.WriteString(fmt.Sprintf("**Ordre de marche**: %s\n\n", strings.Join(party.MarchingOrder, " → ")))

	// Table header
	sb.WriteString("| Nom | Race/Classe | Niv | PV | CA | Stat Principale |\n")
	sb.WriteString("|-----|-------------|-----|----|----|----------------|\n")

	for _, m := range members {
		primaryStat := m["primary_stat"].(map[string]interface{})
		statName := primaryStat["name"].(string)[:3] // First 3 letters
		statMod := primaryStat["modifier"].(int)
		modStr := fmt.Sprintf("%+d", statMod)

		sb.WriteString(fmt.Sprintf("| %s | %s %s | %d | %d/%d | %d | %s %s |\n",
			m["name"],
			m["species"],
			m["class"],
			m["level"],
			m["hp"],
			m["max_hp"],
			m["ac"],
			statName,
			modStr,
		))
	}

	return sb.String()
}

// formatCharacterDisplay formats detailed character info for display.
func formatCharacterDisplay(char *character.Character) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# %s\n", char.Name))
	sb.WriteString(fmt.Sprintf("**%s %s, Niveau %d** (XP: %d)\n\n", capitalize(char.Species), capitalize(char.Class), char.Level, char.XP))

	// Combat stats
	sb.WriteString("## Combat\n")
	sb.WriteString(fmt.Sprintf("- **PV**: %d/%d\n", char.HitPoints, char.MaxHitPoints))
	sb.WriteString(fmt.Sprintf("- **CA**: %d\n", char.ArmorClass))
	sb.WriteString(fmt.Sprintf("- **Or**: %d po\n\n", char.Gold))

	// Abilities table
	sb.WriteString("## Caractéristiques\n\n")
	sb.WriteString("| FOR | INT | SAG | DEX | CON | CHA |\n")
	sb.WriteString("|-----|-----|-----|-----|-----|-----|\n")
	sb.WriteString(fmt.Sprintf("| %d (%s) | %d (%s) | %d (%s) | %d (%s) | %d (%s) | %d (%s) |\n\n",
		char.Abilities.Strength, formatMod(char.Modifiers.Strength),
		char.Abilities.Intelligence, formatMod(char.Modifiers.Intelligence),
		char.Abilities.Wisdom, formatMod(char.Modifiers.Wisdom),
		char.Abilities.Dexterity, formatMod(char.Modifiers.Dexterity),
		char.Abilities.Constitution, formatMod(char.Modifiers.Constitution),
		char.Abilities.Charisma, formatMod(char.Modifiers.Charisma),
	))

	// Equipment
	if len(char.Equipment) > 0 {
		sb.WriteString("## Équipement\n")
		for _, item := range char.Equipment {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	// Spells
	if len(char.KnownSpells) > 0 || len(char.PreparedSpells) > 0 {
		sb.WriteString("## Magie\n")
		if char.SpellSlots != nil {
			sb.WriteString("**Emplacements**: ")
			slots := []string{}
			for lvl := 1; lvl <= 6; lvl++ {
				if s, ok := char.SpellSlots[lvl]; ok && s > 0 {
					used := 0
					if char.SpellSlotsUsed != nil {
						used = char.SpellSlotsUsed[lvl]
					}
					slots = append(slots, fmt.Sprintf("N%d: %d/%d", lvl, s-used, s))
				}
			}
			sb.WriteString(strings.Join(slots, ", "))
			sb.WriteString("\n")
		}
		if len(char.PreparedSpells) > 0 {
			sb.WriteString(fmt.Sprintf("**Préparés**: %s\n", strings.Join(char.PreparedSpells, ", ")))
		}
		sb.WriteString("\n")
	}

	// Appearance
	if char.Appearance != nil {
		sb.WriteString("## Apparence\n")
		details := []string{}
		if char.Appearance.Age > 0 {
			details = append(details, fmt.Sprintf("%d ans", char.Appearance.Age))
		}
		if char.Appearance.Gender != "" {
			details = append(details, char.Appearance.Gender)
		}
		if char.Appearance.Build != "" {
			details = append(details, char.Appearance.Build)
		}
		if char.Appearance.Height != "" {
			details = append(details, char.Appearance.Height)
		}
		if len(details) > 0 {
			sb.WriteString(strings.Join(details, ", "))
			sb.WriteString("\n")
		}
		if char.Appearance.DistinctiveFeature != "" {
			sb.WriteString(fmt.Sprintf("**Trait distinctif**: %s\n", char.Appearance.DistinctiveFeature))
		}
		if char.Appearance.ArmorDescription != "" {
			sb.WriteString(fmt.Sprintf("**Armure**: %s\n", char.Appearance.ArmorDescription))
		}
		if char.Appearance.WeaponDescription != "" {
			sb.WriteString(fmt.Sprintf("**Arme**: %s\n", char.Appearance.WeaponDescription))
		}
	}

	return sb.String()
}

// formatMod formats a modifier with sign.
func formatMod(mod int) string {
	if mod >= 0 {
		return fmt.Sprintf("+%d", mod)
	}
	return fmt.Sprintf("%d", mod)
}

// capitalize capitalizes the first letter of a string.
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
