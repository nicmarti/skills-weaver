package dmtools

import (
	"fmt"
	"path/filepath"
	"strings"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"dungeons/internal/data"
)

// NewCreateCharacterTool creates a tool to create a new character and add them to the party.
// The character is saved to both the adventure directory and the global data/characters directory,
// and automatically added to party.json.
func NewCreateCharacterTool(adv *adventure.Adventure, gd *data.GameData) *SimpleTool {
	return &SimpleTool{
		name: "create_character",
		description: `Créer un nouveau personnage joueur complet et l'ajouter au groupe.
Le personnage est sauvegardé dans l'aventure ET dans data/characters/, puis ajouté à party.json.
Si abilities n'est pas fourni, les scores sont générés via 4d6kh3 (méthode standard).
Si hit_points n'est pas fourni, le maximum du dé de vie + CON est utilisé.
Si armor_class n'est pas fourni, il est calculé (10 + DEX).
Si gold n'est pas fourni, l'or de départ est lancé selon la classe.`,
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Nom complet du personnage",
				},
				"species": map[string]interface{}{
					"type":        "string",
					"description": "Espèce (human, elf, dwarf, halfling, gnome, dragonborn, tiefling, orc, goliath)",
				},
				"class": map[string]interface{}{
					"type":        "string",
					"description": "Classe (fighter, wizard, cleric, rogue, ranger, paladin, barbarian, bard, druid, monk, sorcerer, warlock)",
				},
				"level": map[string]interface{}{
					"type":        "integer",
					"description": "Niveau du personnage (défaut: 1)",
				},
				"abilities": map[string]interface{}{
					"type":        "object",
					"description": "Scores de caractéristiques. Si omis, générés via 4d6kh3.",
					"properties": map[string]interface{}{
						"strength":     map[string]interface{}{"type": "integer"},
						"dexterity":    map[string]interface{}{"type": "integer"},
						"constitution": map[string]interface{}{"type": "integer"},
						"intelligence": map[string]interface{}{"type": "integer"},
						"wisdom":       map[string]interface{}{"type": "integer"},
						"charisma":     map[string]interface{}{"type": "integer"},
					},
				},
				"hit_points": map[string]interface{}{
					"type":        "integer",
					"description": "Points de vie max. Si omis, calculé: max dé de vie + CON mod.",
				},
				"armor_class": map[string]interface{}{
					"type":        "integer",
					"description": "Classe d'armure. Si omis, calculé: 10 + DEX mod.",
				},
				"gold": map[string]interface{}{
					"type":        "integer",
					"description": "Or de départ. Si omis, lancé selon la classe.",
				},
				"equipment": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Liste d'équipement (noms libres)",
				},
				"background": map[string]interface{}{
					"type":        "string",
					"description": "Background du personnage (ex: 'Sage', 'Soldat', 'Criminel')",
				},
				"skills": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Compétences maîtrisées (ex: ['perception', 'stealth'])",
				},
				"saving_throw_profs": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Jets de sauvegarde maîtrisés (ex: ['strength', 'constitution'])",
				},
				"class_features": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Capacités de classe (ex: ['Second Wind', 'Fighting Style: Defense'])",
				},
				"known_spells": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Sorts connus (IDs ou noms)",
				},
				"prepared_spells": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Sorts préparés (IDs ou noms)",
				},
				"appearance": map[string]interface{}{
					"type":        "object",
					"description": "Description visuelle pour génération d'images",
					"properties": map[string]interface{}{
						"age":                map[string]interface{}{"type": "integer"},
						"gender":             map[string]interface{}{"type": "string"},
						"build":              map[string]interface{}{"type": "string"},
						"height":             map[string]interface{}{"type": "string"},
						"hair_color":         map[string]interface{}{"type": "string"},
						"hair_style":         map[string]interface{}{"type": "string"},
						"eye_color":          map[string]interface{}{"type": "string"},
						"skin_tone":          map[string]interface{}{"type": "string"},
						"facial_feature":     map[string]interface{}{"type": "string"},
						"distinctive_feature": map[string]interface{}{"type": "string"},
						"armor_description":  map[string]interface{}{"type": "string"},
						"weapon_description": map[string]interface{}{"type": "string"},
						"accessories":        map[string]interface{}{"type": "string"},
					},
				},
			},
			"required": []string{"name", "species", "class"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			// Extract required parameters
			name, ok := params["name"].(string)
			if !ok || name == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "name est requis",
				}, nil
			}

			species, ok := params["species"].(string)
			if !ok || species == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "species est requis",
				}, nil
			}

			class, ok := params["class"].(string)
			if !ok || class == "" {
				return map[string]interface{}{
					"success": false,
					"error":   "class est requis",
				}, nil
			}

			// Check for duplicate name in party
			existingChars, _ := adv.GetCharacters()
			nameLower := strings.ToLower(name)
			for _, c := range existingChars {
				if strings.ToLower(c.Name) == nameLower {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("Un personnage nommé '%s' existe déjà dans le groupe", c.Name),
					}, nil
				}
			}

			// Create character
			char := character.New(name, species, class)

			// Validate species/class
			if err := char.Validate(gd); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Validation échouée: %v", err),
				}, nil
			}

			// Set level
			level := 1
			if l, ok := params["level"].(float64); ok && l >= 1 {
				level = int(l)
			}
			char.Level = level
			char.CalculateProficiencyBonus()

			// Set abilities
			if abilities, ok := params["abilities"].(map[string]interface{}); ok {
				if v, ok := abilities["strength"].(float64); ok {
					char.Abilities.Strength = int(v)
				}
				if v, ok := abilities["dexterity"].(float64); ok {
					char.Abilities.Dexterity = int(v)
				}
				if v, ok := abilities["constitution"].(float64); ok {
					char.Abilities.Constitution = int(v)
				}
				if v, ok := abilities["intelligence"].(float64); ok {
					char.Abilities.Intelligence = int(v)
				}
				if v, ok := abilities["wisdom"].(float64); ok {
					char.Abilities.Wisdom = int(v)
				}
				if v, ok := abilities["charisma"].(float64); ok {
					char.Abilities.Charisma = int(v)
				}
			} else {
				char.GenerateAbilities(character.MethodStandard)
			}
			char.CalculateModifiers()

			// Set hit points
			if hp, ok := params["hit_points"].(float64); ok && hp > 0 {
				char.HitPoints = int(hp)
				char.MaxHitPoints = int(hp)
			} else {
				if err := char.RollHitPoints(gd, true); err != nil {
					return map[string]interface{}{
						"success": false,
						"error":   fmt.Sprintf("Erreur calcul PV: %v", err),
					}, nil
				}
			}

			// Set armor class
			if ac, ok := params["armor_class"].(float64); ok && ac > 0 {
				char.ArmorClass = int(ac)
			} else {
				char.CalculateArmorClass(gd)
			}

			// Set gold
			if gold, ok := params["gold"].(float64); ok && gold >= 0 {
				char.Gold = int(gold)
			} else {
				if err := char.RollStartingGold(gd); err != nil {
					// Non-fatal: default to 0
					char.Gold = 0
				}
			}

			// Set equipment
			if equip, ok := params["equipment"].([]interface{}); ok {
				for _, item := range equip {
					if s, ok := item.(string); ok && s != "" {
						char.Equipment = append(char.Equipment, s)
					}
				}
			}

			// Set background
			if bg, ok := params["background"].(string); ok && bg != "" {
				char.Background = bg
			}

			// Set skills
			if skillsList, ok := params["skills"].([]interface{}); ok && len(skillsList) > 0 {
				char.Skills = make(map[string]bool)
				for _, s := range skillsList {
					if skillName, ok := s.(string); ok && skillName != "" {
						char.Skills[skillName] = true
					}
				}
			}

			// Set saving throw proficiencies
			if stProfs, ok := params["saving_throw_profs"].([]interface{}); ok && len(stProfs) > 0 {
				char.SavingThrowProfs = make(map[string]bool)
				for _, s := range stProfs {
					if stName, ok := s.(string); ok && stName != "" {
						char.SavingThrowProfs[stName] = true
					}
				}
			}

			// Set class features
			if features, ok := params["class_features"].([]interface{}); ok {
				for _, f := range features {
					if s, ok := f.(string); ok && s != "" {
						char.ClassFeatures = append(char.ClassFeatures, s)
					}
				}
			}

			// Set spells
			if spells, ok := params["known_spells"].([]interface{}); ok {
				for _, s := range spells {
					if spellName, ok := s.(string); ok && spellName != "" {
						char.KnownSpells = append(char.KnownSpells, spellName)
					}
				}
			}
			if spells, ok := params["prepared_spells"].([]interface{}); ok {
				for _, s := range spells {
					if spellName, ok := s.(string); ok && spellName != "" {
						char.PreparedSpells = append(char.PreparedSpells, spellName)
					}
				}
			}

			// Initialize spell slots for spellcasters
			char.InitializeSpellSlots(gd)

			// Set hit dice
			char.HitDice = level
			char.MaxHitDice = level

			// Set appearance
			if appearanceData, ok := params["appearance"].(map[string]interface{}); ok {
				appearance := character.CharacterAppearance{}
				if v, ok := appearanceData["age"].(float64); ok {
					appearance.Age = int(v)
				}
				if v, ok := appearanceData["gender"].(string); ok {
					appearance.Gender = v
				}
				if v, ok := appearanceData["build"].(string); ok {
					appearance.Build = v
				}
				if v, ok := appearanceData["height"].(string); ok {
					appearance.Height = v
				}
				if v, ok := appearanceData["hair_color"].(string); ok {
					appearance.HairColor = v
				}
				if v, ok := appearanceData["hair_style"].(string); ok {
					appearance.HairStyle = v
				}
				if v, ok := appearanceData["eye_color"].(string); ok {
					appearance.EyeColor = v
				}
				if v, ok := appearanceData["skin_tone"].(string); ok {
					appearance.SkinTone = v
				}
				if v, ok := appearanceData["facial_feature"].(string); ok {
					appearance.FacialFeature = v
				}
				if v, ok := appearanceData["distinctive_feature"].(string); ok {
					appearance.DistinctiveFeature = v
				}
				if v, ok := appearanceData["armor_description"].(string); ok {
					appearance.ArmorDescription = v
				}
				if v, ok := appearanceData["weapon_description"].(string); ok {
					appearance.WeaponDescription = v
				}
				if v, ok := appearanceData["accessories"].(string); ok {
					appearance.Accessories = v
				}
				char.Appearance = &appearance
			}

			// Save to adventure characters directory
			advCharDir := filepath.Join(adv.BasePath(), "characters")
			if err := char.Save(advCharDir); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur sauvegarde aventure: %v", err),
				}, nil
			}

			// Save to global characters directory
			if err := char.Save("data/characters"); err != nil {
				// Non-fatal: log but continue
				fmt.Printf("Warning: could not save to global characters: %v\n", err)
			}

			// Add to party.json
			party, err := adv.LoadParty()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur chargement party: %v", err),
				}, nil
			}
			party.Characters = append(party.Characters, char.Name)
			party.MarchingOrder = append(party.MarchingOrder, char.Name)
			if err := adv.SaveParty(party); err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   fmt.Sprintf("Erreur sauvegarde party: %v", err),
				}, nil
			}

			// Log journal event
			adv.LogEvent("story", fmt.Sprintf("Nouveau personnage créé: %s (%s %s, niveau %d)", char.Name, capitalize(char.Species), capitalize(char.Class), char.Level))

			// Build display
			display := fmt.Sprintf("✓ Personnage créé: %s\n", char.Name)
			display += fmt.Sprintf("  %s %s, Niveau %d\n", capitalize(char.Species), capitalize(char.Class), char.Level)
			display += fmt.Sprintf("  FOR %d (%s) DEX %d (%s) CON %d (%s)\n",
				char.Abilities.Strength, formatMod(char.Modifiers.Strength),
				char.Abilities.Dexterity, formatMod(char.Modifiers.Dexterity),
				char.Abilities.Constitution, formatMod(char.Modifiers.Constitution))
			display += fmt.Sprintf("  INT %d (%s) SAG %d (%s) CHA %d (%s)\n",
				char.Abilities.Intelligence, formatMod(char.Modifiers.Intelligence),
				char.Abilities.Wisdom, formatMod(char.Modifiers.Wisdom),
				char.Abilities.Charisma, formatMod(char.Modifiers.Charisma))
			display += fmt.Sprintf("  PV: %d/%d | CA: %d | Or: %d po\n", char.HitPoints, char.MaxHitPoints, char.ArmorClass, char.Gold)
			if len(char.Equipment) > 0 {
				display += fmt.Sprintf("  Équipement: %s\n", strings.Join(char.Equipment, ", "))
			}
			display += "  Ajouté au groupe et sauvegardé."

			return map[string]interface{}{
				"success": true,
				"character": map[string]interface{}{
					"name":             char.Name,
					"species":          char.Species,
					"class":            char.Class,
					"level":            char.Level,
					"hp":               char.HitPoints,
					"max_hp":           char.MaxHitPoints,
					"ac":               char.ArmorClass,
					"gold":             char.Gold,
					"proficiency_bonus": char.ProficiencyBonus,
					"abilities": map[string]int{
						"strength":     char.Abilities.Strength,
						"dexterity":    char.Abilities.Dexterity,
						"constitution": char.Abilities.Constitution,
						"intelligence": char.Abilities.Intelligence,
						"wisdom":       char.Abilities.Wisdom,
						"charisma":     char.Abilities.Charisma,
					},
					"modifiers": map[string]int{
						"strength":     char.Modifiers.Strength,
						"dexterity":    char.Modifiers.Dexterity,
						"constitution": char.Modifiers.Constitution,
						"intelligence": char.Modifiers.Intelligence,
						"wisdom":       char.Modifiers.Wisdom,
						"charisma":     char.Modifiers.Charisma,
					},
				},
				"display": display,
			}, nil
		},
	}
}
