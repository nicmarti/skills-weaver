package world

import (
	"fmt"
	"sort"
	"strings"
)

// ValidateLocationExists checks if a location exists in geography data.
// Returns true if found, along with the location and its region.
func ValidateLocationExists(name string, geo *Geography) (bool, *Location, *Region, error) {
	loc, region, err := geo.GetLocationByName(name)
	if err != nil {
		return false, nil, nil, err
	}
	return true, loc, region, nil
}

// ValidateKingdomExists checks if a kingdom exists in factions data.
func ValidateKingdomExists(kingdomID string, factions *Factions) (bool, *Kingdom, error) {
	kingdom, err := factions.GetKingdomByID(kingdomID)
	if err != nil {
		return false, nil, err
	}
	return true, kingdom, nil
}

// ValidateLocationKingdom checks if a location belongs to the specified kingdom.
func ValidateLocationKingdom(location *Location, kingdomID string) error {
	if strings.ToLower(location.Kingdom) != strings.ToLower(kingdomID) {
		return fmt.Errorf("location '%s' belongs to %s, not %s",
			location.Name, location.Kingdom, kingdomID)
	}
	return nil
}

// GetSuggestions returns location names similar to the partial name using fuzzy matching.
// Uses Levenshtein-like distance for similarity scoring.
func GetSuggestions(partial string, geo *Geography, maxResults int) []LocationSuggestion {
	if maxResults <= 0 {
		maxResults = 5
	}

	partialLower := strings.ToLower(partial)
	var suggestions []LocationSuggestion

	for _, loc := range geo.GetAllLocations() {
		nameLower := strings.ToLower(loc.Name)

		// Calculate similarity score
		score := calculateSimilarity(partialLower, nameLower)

		if score > 0 {
			suggestions = append(suggestions, LocationSuggestion{
				Location: loc,
				Score:    score,
			})
		}
	}

	// Sort by score (descending)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	// Return top N results
	if len(suggestions) > maxResults {
		suggestions = suggestions[:maxResults]
	}

	return suggestions
}

// LocationSuggestion represents a suggested location with similarity score.
type LocationSuggestion struct {
	Location Location
	Score    int
}

// calculateSimilarity returns a similarity score between two strings.
// Higher score means more similar.
func calculateSimilarity(partial, name string) int {
	score := 0

	// Exact match (case-insensitive)
	if partial == name {
		return 1000
	}

	// Starts with partial
	if strings.HasPrefix(name, partial) {
		score += 500
	}

	// Contains partial
	if strings.Contains(name, partial) {
		score += 250
	}

	// Count matching characters (simple approach)
	for i, char := range partial {
		if i < len(name) && name[i] == byte(char) {
			score += 10
		}
	}

	// Levenshtein distance (penalize difference)
	distance := levenshteinDistance(partial, name)
	score -= distance * 5

	return score
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// GetLocationTypeString returns a human-readable location type description.
func GetLocationTypeString(locationType string) string {
	typeMap := map[string]string{
		"port majeur":          "Port Majeur",
		"port industriel":      "Port Industriel",
		"port financier":       "Port Financier",
		"village":              "Village",
		"forteresse capitale":  "Forteresse Capitale",
		"forteresse frontalière": "Forteresse Frontalière",
		"cité industrielle":    "Cité Industrielle",
		"ville sainte":         "Ville Sainte",
		"capitale":             "Capitale",
		"cité universitaire":   "Cité Universitaire",
	}

	if readable, ok := typeMap[strings.ToLower(locationType)]; ok {
		return readable
	}

	// Capitalize first letter of each word
	words := strings.Fields(locationType)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
