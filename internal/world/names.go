package world

import (
	"fmt"
	"strings"
)

// ValidateLocationName checks if a location name follows kingdom naming conventions.
// Returns true if valid, along with any warnings or suggestions.
func ValidateLocationName(name, locationType, kingdomID string, names *LocationNames) (bool, []string) {
	var warnings []string

	// Get kingdom naming patterns
	naming, ok := names.Kingdoms[strings.ToLower(kingdomID)]
	if !ok {
		return false, []string{fmt.Sprintf("Kingdom '%s' not found in naming conventions", kingdomID)}
	}

	// Get pattern based on location type
	var pattern NamingPattern
	switch strings.ToLower(locationType) {
	case "city", "ville", "cité", "capitale", "port", "port majeur", "port industriel", "port financier":
		pattern = naming.Cities
	case "town", "bourg", "forteresse", "forteresse capitale", "forteresse frontalière":
		pattern = naming.Towns
	case "village":
		pattern = naming.Villages
	default:
		// Default to cities for unknown types
		pattern = naming.Cities
		warnings = append(warnings, fmt.Sprintf("Unknown location type '%s', using city patterns", locationType))
	}

	// Check if name matches conventions
	matches := MatchesNamingConvention(name, pattern)
	if !matches {
		warnings = append(warnings, fmt.Sprintf(
			"Name '%s' doesn't follow %s naming conventions for %s",
			name, kingdomID, locationType))
		warnings = append(warnings, fmt.Sprintf(
			"Expected patterns: %s + %s",
			strings.Join(pattern.Prefixes, "/"),
			strings.Join(pattern.Suffixes, "/")))
	}

	return matches, warnings
}

// MatchesNamingConvention checks if a name follows a naming pattern.
func MatchesNamingConvention(name string, pattern NamingPattern) bool {
	nameLower := strings.ToLower(name)

	// Check prefixes
	matchesPrefix := false
	for _, prefix := range pattern.Prefixes {
		if strings.HasPrefix(nameLower, strings.ToLower(prefix)) {
			matchesPrefix = true
			break
		}
	}

	// Check suffixes
	matchesSuffix := false
	for _, suffix := range pattern.Suffixes {
		if strings.HasSuffix(nameLower, strings.ToLower(suffix)) {
			matchesSuffix = true
			break
		}
	}

	// Check roots (if present in pattern)
	matchesRoot := len(pattern.Roots) == 0 // If no roots defined, consider it a match
	for _, root := range pattern.Roots {
		if strings.Contains(nameLower, strings.ToLower(root)) {
			matchesRoot = true
			break
		}
	}

	// Name should match prefix AND suffix (AND root if applicable)
	return matchesPrefix && matchesSuffix && matchesRoot
}

// GenerateNameSuggestions generates location name suggestions based on kingdom conventions.
func GenerateNameSuggestions(locationType, kingdomID string, names *LocationNames, count int) []string {
	if count <= 0 {
		count = 5
	}

	naming, ok := names.Kingdoms[strings.ToLower(kingdomID)]
	if !ok {
		return []string{}
	}

	// Get pattern based on location type
	var pattern NamingPattern
	switch strings.ToLower(locationType) {
	case "city", "ville", "cité", "capitale", "port":
		pattern = naming.Cities
	case "town", "bourg", "forteresse":
		pattern = naming.Towns
	case "village":
		pattern = naming.Villages
	default:
		pattern = naming.Cities
	}

	// Generate combinations
	var suggestions []string
	generated := 0

	// Simple generation: prefix + root + suffix
	for _, prefix := range pattern.Prefixes {
		if generated >= count {
			break
		}

		for _, suffix := range pattern.Suffixes {
			if generated >= count {
				break
			}

			// If roots exist, include them
			if len(pattern.Roots) > 0 {
				for _, root := range pattern.Roots {
					if generated >= count {
						break
					}
					suggestions = append(suggestions, fmt.Sprintf("%s%s%s", prefix, root, suffix))
					generated++
				}
			} else {
				// No roots, just prefix + suffix
				suggestions = append(suggestions, fmt.Sprintf("%s%s", prefix, suffix))
				generated++
			}
		}
	}

	return suggestions[:min2(len(suggestions), count)]
}

// GetKingdomNamingStyle returns a descriptive string of a kingdom's naming style.
func GetKingdomNamingStyle(kingdomID string) string {
	styles := map[string]string{
		"valdorine":  "Maritime (Italian-inspired with water-themed names)",
		"karvath":    "Militaristic (German-inspired with martial names)",
		"lumenciel":  "Religious (Latin-inspired with sacred names)",
		"astrene":    "Melancholic (Nordic-inspired with somber names)",
	}

	if style, ok := styles[strings.ToLower(kingdomID)]; ok {
		return style
	}

	return "Unknown style"
}

// GetKingdomNameExamples returns example location names for a kingdom.
func GetKingdomNameExamples(kingdomID string) []string {
	examples := map[string][]string{
		"valdorine": {
			"Cordova", "Port-de-Lune", "Havre-d'Argent",
			"Belvento", "Calmonte", "Torrosso",
		},
		"karvath": {
			"Fer-de-Lance", "Porte-de-Fer", "Forge-Noire",
			"Hautfort", "Rocstein", "Starkwald",
		},
		"lumenciel": {
			"Aurore-Sainte", "Refuge-des-Purs", "Vallon-de-Prière",
			"Altarium", "Primanus", "Novensus",
		},
		"astrene": {
			"Étoile-d'Automne", "Valombre", "Port-des-Souvenirs",
			"Gamborg", "Vindlund", "Koldsø",
		},
	}

	if exs, ok := examples[strings.ToLower(kingdomID)]; ok {
		return exs
	}

	return []string{}
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
