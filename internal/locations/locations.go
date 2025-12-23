package locations

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CityParts holds components for city names.
type CityParts struct {
	Prefixes []string `json:"prefixes"`
	Roots    []string `json:"roots"`
	Suffixes []string `json:"suffixes"`
}

// TownParts holds components for town names.
type TownParts struct {
	Prefixes []string `json:"prefixes"`
	Suffixes []string `json:"suffixes"`
}

// VillageParts holds components for village names.
type VillageParts struct {
	Prefixes []string `json:"prefixes"`
	Names    []string `json:"names"`
	Suffixes []string `json:"suffixes"`
}

// RegionTemplates holds templates for region names.
type RegionTemplates struct {
	Templates         []string `json:"templates"`
	Names             []string `json:"names"`
	Adjectives        []string `json:"adjectives"`
	AdjectivesPlural  []string `json:"adjectives_plural"`
	Nouns             []string `json:"nouns"`
}

// KingdomNames holds all naming data for a kingdom.
type KingdomNames struct {
	Cities  CityParts       `json:"cities"`
	Towns   TownParts       `json:"towns"`
	Villages VillageParts   `json:"villages"`
	Regions RegionTemplates `json:"regions"`
}

// RuinParts holds components for ruins.
type RuinParts struct {
	Prefixes []string `json:"prefixes"`
	Bases    []string `json:"bases"`
	Suffixes []string `json:"suffixes"`
}

// GenericNames holds neutral/generic location names.
type GenericNames struct {
	Geographical []string `json:"geographical"`
	Adjectives   []string `json:"adjectives"`
}

// SpecialNames holds special location patterns.
type SpecialNames struct {
	Templates   []string `json:"templates"`
	Descriptors []string `json:"descriptors"`
}

// NeutralNames holds neutral naming data.
type NeutralNames struct {
	Ruins   RuinParts    `json:"ruins"`
	Generic GenericNames `json:"generic"`
	Special SpecialNames `json:"special"`
}

// LocationData holds all location naming data.
type LocationData struct {
	Valdorine KingdomNames `json:"valdorine"`
	Karvath   KingdomNames `json:"karvath"`
	Lumenciel KingdomNames `json:"lumenciel"`
	Astrene   KingdomNames `json:"astrene"`
	Neutral   NeutralNames `json:"neutral"`
}

// Generator generates random location names.
type Generator struct {
	data *LocationData
	rng  *rand.Rand
}

// NewGenerator creates a new location name generator.
func NewGenerator(dataDir string) (*Generator, error) {
	path := filepath.Join(dataDir, "location-names.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading location-names.json: %w", err)
	}

	var locations LocationData
	if err := json.Unmarshal(data, &locations); err != nil {
		return nil, fmt.Errorf("parsing location-names.json: %w", err)
	}

	return &Generator{
		data: &locations,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// GenerateCity generates a city name for a kingdom.
func (g *Generator) GenerateCity(kingdom string) (string, error) {
	kingdom = strings.ToLower(kingdom)

	var parts CityParts
	switch kingdom {
	case "valdorine":
		parts = g.data.Valdorine.Cities
	case "karvath":
		parts = g.data.Karvath.Cities
	case "lumenciel":
		parts = g.data.Lumenciel.Cities
	case "astrene", "astrène":
		parts = g.data.Astrene.Cities
	default:
		return "", fmt.Errorf("unknown kingdom: %s", kingdom)
	}

	if len(parts.Prefixes) == 0 || len(parts.Roots) == 0 || len(parts.Suffixes) == 0 {
		return "", fmt.Errorf("insufficient city parts for kingdom: %s", kingdom)
	}

	prefix := parts.Prefixes[g.rng.Intn(len(parts.Prefixes))]
	root := parts.Roots[g.rng.Intn(len(parts.Roots))]
	suffix := parts.Suffixes[g.rng.Intn(len(parts.Suffixes))]

	return fmt.Sprintf("%s%s%s", prefix, root, suffix), nil
}

// GenerateTown generates a town name for a kingdom.
func (g *Generator) GenerateTown(kingdom string) (string, error) {
	kingdom = strings.ToLower(kingdom)

	var parts TownParts
	switch kingdom {
	case "valdorine":
		parts = g.data.Valdorine.Towns
	case "karvath":
		parts = g.data.Karvath.Towns
	case "lumenciel":
		parts = g.data.Lumenciel.Towns
	case "astrene", "astrène":
		parts = g.data.Astrene.Towns
	default:
		return "", fmt.Errorf("unknown kingdom: %s", kingdom)
	}

	if len(parts.Prefixes) == 0 || len(parts.Suffixes) == 0 {
		return "", fmt.Errorf("insufficient town parts for kingdom: %s", kingdom)
	}

	prefix := parts.Prefixes[g.rng.Intn(len(parts.Prefixes))]
	suffix := parts.Suffixes[g.rng.Intn(len(parts.Suffixes))]

	return fmt.Sprintf("%s%s", prefix, suffix), nil
}

// GenerateVillage generates a village name for a kingdom.
func (g *Generator) GenerateVillage(kingdom string) (string, error) {
	kingdom = strings.ToLower(kingdom)

	var parts VillageParts
	switch kingdom {
	case "valdorine":
		parts = g.data.Valdorine.Villages
	case "karvath":
		parts = g.data.Karvath.Villages
	case "lumenciel":
		parts = g.data.Lumenciel.Villages
	case "astrene", "astrène":
		parts = g.data.Astrene.Villages
	default:
		return "", fmt.Errorf("unknown kingdom: %s", kingdom)
	}

	// All kingdoms use "Prefix + Suffix" pattern
	if len(parts.Prefixes) == 0 || len(parts.Suffixes) == 0 {
		return "", fmt.Errorf("insufficient village parts for kingdom: %s", kingdom)
	}

	prefix := parts.Prefixes[g.rng.Intn(len(parts.Prefixes))]
	suffix := parts.Suffixes[g.rng.Intn(len(parts.Suffixes))]

	return fmt.Sprintf("%s%s", prefix, suffix), nil
}

// GenerateRegion generates a region name for a kingdom.
func (g *Generator) GenerateRegion(kingdom string) (string, error) {
	kingdom = strings.ToLower(kingdom)

	var templates RegionTemplates
	switch kingdom {
	case "valdorine":
		templates = g.data.Valdorine.Regions
	case "karvath":
		templates = g.data.Karvath.Regions
	case "lumenciel":
		templates = g.data.Lumenciel.Regions
	case "astrene", "astrène":
		templates = g.data.Astrene.Regions
	default:
		return "", fmt.Errorf("unknown kingdom: %s", kingdom)
	}

	if len(templates.Templates) == 0 {
		return "", fmt.Errorf("no region templates for kingdom: %s", kingdom)
	}

	template := templates.Templates[g.rng.Intn(len(templates.Templates))]

	// Replace placeholders
	result := template
	if strings.Contains(template, "{name}") {
		if len(templates.Names) == 0 {
			return "", fmt.Errorf("no names for kingdom: %s", kingdom)
		}
		name := templates.Names[g.rng.Intn(len(templates.Names))]
		result = strings.Replace(result, "{name}", name, 1)
	}
	if strings.Contains(template, "{adjective-pl}") {
		if len(templates.AdjectivesPlural) == 0 {
			return "", fmt.Errorf("no plural adjectives for kingdom: %s", kingdom)
		}
		adj := templates.AdjectivesPlural[g.rng.Intn(len(templates.AdjectivesPlural))]
		result = strings.Replace(result, "{adjective-pl}", adj, 1)
	}
	if strings.Contains(template, "{adjective}") {
		if len(templates.Adjectives) == 0 {
			return "", fmt.Errorf("no adjectives for kingdom: %s", kingdom)
		}
		adj := templates.Adjectives[g.rng.Intn(len(templates.Adjectives))]
		result = strings.Replace(result, "{adjective}", adj, 1)
	}
	if strings.Contains(template, "{noun-gen}") {
		if len(templates.Nouns) == 0 {
			return "", fmt.Errorf("no nouns for kingdom: %s", kingdom)
		}
		noun := templates.Nouns[g.rng.Intn(len(templates.Nouns))]
		result = strings.Replace(result, "{noun-gen}", "des "+noun, 1)
	}
	if strings.Contains(template, "{noun}") {
		if len(templates.Nouns) == 0 {
			return "", fmt.Errorf("no nouns for kingdom: %s", kingdom)
		}
		noun := templates.Nouns[g.rng.Intn(len(templates.Nouns))]
		result = strings.Replace(result, "{noun}", noun, 1)
	}

	return result, nil
}

// GenerateRuin generates a ruin name.
func (g *Generator) GenerateRuin(includeType bool) string {
	parts := g.data.Neutral.Ruins

	prefix := parts.Prefixes[g.rng.Intn(len(parts.Prefixes))]
	base := parts.Bases[g.rng.Intn(len(parts.Bases))]

	if includeType && g.rng.Intn(2) == 0 {
		suffix := parts.Suffixes[g.rng.Intn(len(parts.Suffixes))]
		return fmt.Sprintf("%s %s (%s)", prefix, base, suffix)
	}

	return fmt.Sprintf("%s %s", prefix, base)
}

// GenerateGeneric generates a generic geographical name.
func (g *Generator) GenerateGeneric() string {
	generic := g.data.Neutral.Generic

	geo := generic.Geographical[g.rng.Intn(len(generic.Geographical))]

	// 50% chance to add an adjective
	if g.rng.Intn(2) == 0 {
		adj := generic.Adjectives[g.rng.Intn(len(generic.Adjectives))]
		return fmt.Sprintf("%s %s", geo, adj)
	}

	return geo
}

// GenerateSpecial generates a special neutral location name.
func (g *Generator) GenerateSpecial() string {
	special := g.data.Neutral.Special

	// 50% chance to use predefined descriptors
	if len(special.Descriptors) > 0 && g.rng.Intn(10) < 5 {
		return special.Descriptors[g.rng.Intn(len(special.Descriptors))]
	}

	// Otherwise use templates if available
	if len(special.Templates) > 0 {
		template := special.Templates[g.rng.Intn(len(special.Templates))]
		result := template

		// Replace {geographical}
		if strings.Contains(template, "{geographical}") {
			geo := g.data.Neutral.Generic.Geographical[g.rng.Intn(len(g.data.Neutral.Generic.Geographical))]
			result = strings.Replace(result, "{geographical}", geo, 1)
		}

		// Replace {adjective}
		if strings.Contains(template, "{adjective}") {
			adj := g.data.Neutral.Generic.Adjectives[g.rng.Intn(len(g.data.Neutral.Generic.Adjectives))]
			result = strings.Replace(result, "{adjective}", adj, 1)
		}

		// Replace {name} - generate a random region-like name
		if strings.Contains(template, "{name}") {
			// Use a mix of geographical and proper nouns
			name := g.data.Neutral.Generic.Geographical[g.rng.Intn(len(g.data.Neutral.Generic.Geographical))]
			result = strings.Replace(result, "{name}", name, 1)
		}

		return result
	}

	// Fallback to descriptor if no templates
	if len(special.Descriptors) > 0 {
		return special.Descriptors[g.rng.Intn(len(special.Descriptors))]
	}

	return "Terres Inconnues"
}

// GenerateMultiple generates multiple unique names.
func (g *Generator) GenerateMultiple(kingdom, locationType string, count int) ([]string, error) {
	names := make([]string, 0, count)
	seen := make(map[string]bool)

	for len(names) < count {
		var name string
		var err error

		switch strings.ToLower(locationType) {
		case "city", "ville":
			name, err = g.GenerateCity(kingdom)
		case "town", "bourg":
			name, err = g.GenerateTown(kingdom)
		case "village":
			name, err = g.GenerateVillage(kingdom)
		case "region", "région":
			name, err = g.GenerateRegion(kingdom)
		case "ruin", "ruine":
			name = g.GenerateRuin(true)
		case "generic", "générique":
			name = g.GenerateGeneric()
		case "special", "spécial":
			name = g.GenerateSpecial()
		default:
			return nil, fmt.Errorf("unknown location type: %s", locationType)
		}

		if err != nil {
			return nil, err
		}

		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}

		// Prevent infinite loop
		if len(seen) > count*20 {
			break
		}
	}

	return names, nil
}

// GetAvailableKingdoms returns the list of available kingdoms.
func (g *Generator) GetAvailableKingdoms() []string {
	return []string{"valdorine", "karvath", "lumenciel", "astrene"}
}

// GetAvailableTypes returns the list of available location types.
func (g *Generator) GetAvailableTypes() []string {
	return []string{"city", "town", "village", "region", "ruin", "generic", "special"}
}
