package dmtools

import (
	"fmt"
	"strings"

	"dungeons/internal/locations"
	"dungeons/internal/names"
)

// NewGenerateNameTool creates a tool to generate character names.
func NewGenerateNameTool(generator *names.Generator) *SimpleTool {
	return &SimpleTool{
		name:        "generate_name",
		description: "Generate random fantasy character names. Use this for quick NPC naming without creating a full NPC profile.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"race": map[string]interface{}{
					"type":        "string",
					"description": "Race for the name (human, elf, dwarf, halfling).",
					"enum":        []string{"human", "elf", "dwarf", "halfling"},
				},
				"gender": map[string]interface{}{
					"type":        "string",
					"description": "Gender for the name (m, f, or omit for random).",
					"enum":        []string{"m", "f"},
				},
				"npc_type": map[string]interface{}{
					"type":        "string",
					"description": "Alternative: Generate name for NPC occupation type.",
					"enum":        []string{"innkeeper", "merchant", "guard", "noble", "wizard", "villain"},
				},
				"count": map[string]interface{}{
					"type":        "integer",
					"description": "Number of names to generate (default: 1, max: 10).",
					"minimum":     1,
					"maximum":     10,
				},
			},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			race, hasRace := params["race"].(string)
			gender, _ := params["gender"].(string)
			npcType, hasNPCType := params["npc_type"].(string)

			count := 1
			if c, ok := params["count"].(float64); ok && c > 0 {
				count = int(c)
				if count > 10 {
					count = 10
				}
			}

			// Generate NPC type names
			if hasNPCType && npcType != "" {
				generatedNames := make([]string, 0, count)
				for i := 0; i < count; i++ {
					name, err := generator.GenerateNPCName(npcType)
					if err != nil {
						return map[string]interface{}{
							"success":        false,
							"error":          err.Error(),
							"available_types": generator.GetAvailableNPCTypes(),
						}, nil
					}
					generatedNames = append(generatedNames, name)
				}

				return map[string]interface{}{
					"success":  true,
					"npc_type": npcType,
					"count":    len(generatedNames),
					"names":    generatedNames,
					"display":  formatNameList(generatedNames, fmt.Sprintf("Noms de %s", npcType)),
				}, nil
			}

			// Generate race-based names
			if hasRace && race != "" {
				if count == 1 {
					name, err := generator.GenerateName(race, gender)
					if err != nil {
						return map[string]interface{}{
							"success":         false,
							"error":           err.Error(),
							"available_races": generator.GetAvailableRaces(),
						}, nil
					}

					return map[string]interface{}{
						"success": true,
						"race":    race,
						"gender":  gender,
						"name":    name,
					}, nil
				}

				generatedNames, err := generator.GenerateMultiple(race, gender, count)
				if err != nil {
					return map[string]interface{}{
						"success":         false,
						"error":           err.Error(),
						"available_races": generator.GetAvailableRaces(),
					}, nil
				}

				genderLabel := "aléatoire"
				if gender == "m" {
					genderLabel = "masculin"
				} else if gender == "f" {
					genderLabel = "féminin"
				}

				return map[string]interface{}{
					"success": true,
					"race":    race,
					"gender":  genderLabel,
					"count":   len(generatedNames),
					"names":   generatedNames,
					"display": formatNameList(generatedNames, fmt.Sprintf("Noms %s (%s)", race, genderLabel)),
				}, nil
			}

			// No parameters - return usage info
			return map[string]interface{}{
				"success":         false,
				"error":           "Provide 'race' or 'npc_type' parameter",
				"available_races": generator.GetAvailableRaces(),
				"available_types": generator.GetAvailableNPCTypes(),
			}, nil
		},
	}
}

// NewGenerateLocationNameTool creates a tool to generate location names.
func NewGenerateLocationNameTool(generator *locations.Generator) *SimpleTool {
	return &SimpleTool{
		name:        "generate_location_name",
		description: "Generate location names consistent with the four kingdoms' naming styles. Use this when improvising new locations during gameplay.",
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"kingdom": map[string]interface{}{
					"type":        "string",
					"description": "Kingdom for naming style (valdorine=maritime, karvath=military, lumenciel=religious, astrene=scholarly).",
					"enum":        []string{"valdorine", "karvath", "lumenciel", "astrene"},
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Type of location (city, town, village, region).",
					"enum":        []string{"city", "town", "village", "region"},
				},
				"count": map[string]interface{}{
					"type":        "integer",
					"description": "Number of names to generate (default: 1, max: 10).",
					"minimum":     1,
					"maximum":     10,
				},
			},
			"required": []string{"kingdom", "type"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			kingdom, ok := params["kingdom"].(string)
			if !ok || kingdom == "" {
				return map[string]interface{}{
					"success":           false,
					"error":             "Kingdom is required",
					"available_kingdoms": generator.GetAvailableKingdoms(),
				}, nil
			}

			locationType, ok := params["type"].(string)
			if !ok || locationType == "" {
				return map[string]interface{}{
					"success":         false,
					"error":           "Location type is required",
					"available_types": []string{"city", "town", "village", "region"},
				}, nil
			}

			count := 1
			if c, ok := params["count"].(float64); ok && c > 0 {
				count = int(c)
				if count > 10 {
					count = 10
				}
			}

			generatedNames, err := generator.GenerateMultiple(kingdom, locationType, count)
			if err != nil {
				return map[string]interface{}{
					"success":           false,
					"error":             err.Error(),
					"available_kingdoms": generator.GetAvailableKingdoms(),
					"available_types":   []string{"city", "town", "village", "region"},
				}, nil
			}

			return map[string]interface{}{
				"success": true,
				"kingdom": kingdom,
				"type":    locationType,
				"count":   len(generatedNames),
				"names":   generatedNames,
				"display": formatNameList(generatedNames, fmt.Sprintf("Noms de %s en %s", locationType, capitalizeKingdom(kingdom))),
			}, nil
		},
	}
}

// formatNameList formats a list of names for display.
func formatNameList(generatedNames []string, title string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s\n\n", title))
	for _, name := range generatedNames {
		sb.WriteString(fmt.Sprintf("- %s\n", name))
	}
	return sb.String()
}

// capitalizeKingdom returns the kingdom name with proper capitalization.
func capitalizeKingdom(kingdom string) string {
	switch strings.ToLower(kingdom) {
	case "valdorine":
		return "Valdorine"
	case "karvath":
		return "Karvath"
	case "lumenciel":
		return "Lumenciel"
	case "astrene":
		return "Astrène"
	default:
		return kingdom
	}
}
