// Package mapgen provides map prompt building functionality for sw-map CLI.
package mapgen

import (
	"dungeons/internal/world"
	"fmt"
	"strings"
)

// MapContext provides contextual data for map generation.
type MapContext struct {
	Location *world.Location
	Region   *world.Region
	Kingdom  *world.Kingdom
	Scale    string // "small", "medium", "large"
}

// PromptOptions configures prompt generation.
type PromptOptions struct {
	Features []string // Specific POIs to include
	Terrain  string   // Terrain type override
	Style    string   // "illustrated" or "dark_fantasy"
}

// BuildCityMapPrompt generates a prompt for a city map.
// Returns a base prompt that can be enriched by AI.
func BuildCityMapPrompt(ctx *MapContext, opts PromptOptions) string {
	loc := ctx.Location
	kingdom := ctx.Kingdom

	parts := []string{}

	// 1. Perspective and map type
	parts = append(parts, "Cette carte montre la ville")

	// 2. Add city name and type
	cityType := strings.ToLower(loc.Type)
	if strings.Contains(cityType, "port") {
		parts = append(parts, fmt.Sprintf("portuaire de %s", loc.Name))
	} else if strings.Contains(cityType, "forteresse") {
		parts = append(parts, fmt.Sprintf("fortifiée de %s", loc.Name))
	} else {
		parts = append(parts, fmt.Sprintf("de %s", loc.Name))
	}

	// 3. Add perspective
	parts = append(parts, "en vue aérienne")

	// 4. Add kingdom architectural style
	if kingdom != nil {
		styleDesc := getKingdomArchitecturalStyle(kingdom.ID)
		parts = append(parts, fmt.Sprintf("Style architectural %s", styleDesc))

		// Add colors
		if len(kingdom.Colors) > 0 {
			colors := strings.Join(kingdom.Colors, " et ")
			parts = append(parts, fmt.Sprintf("avec bâtiments aux couleurs %s", colors))
		}
	}

	// 5. Add geographic context from location description
	if loc.Description != "" {
		parts = append(parts, loc.Description)
	}

	// 6. Add terrain/geographic features
	if opts.Terrain != "" {
		parts = append(parts, fmt.Sprintf("Terrain : %s", opts.Terrain))
	}

	// 7. Add districts/quarters (general description)
	if strings.Contains(cityType, "port") {
		parts = append(parts, "Quartiers : port commercial, district marchand, quartier résidentiel")
	} else if strings.Contains(cityType, "forteresse") {
		parts = append(parts, "Quartiers : casernes militaires, quartier administratif, quartier civil")
	} else {
		parts = append(parts, "Quartiers : centre-ville, quartier marchand, quartiers résidentiels")
	}

	// 8. Add POIs
	pois := collectPOIs(loc, opts.Features)
	if len(pois) > 0 {
		parts = append(parts, fmt.Sprintf("Points d'intérêt : %s", strings.Join(pois, ", ")))
	}

	// 9. Add infrastructure based on city type
	infra := getInfrastructure(cityType)
	if infra != "" {
		parts = append(parts, infra)
	}

	// 10. Add realism instruction
	parts = append(parts, "Ajoute différents éléments réalistes d'une carte")
	parts = append(parts, "La ville doit avoir une forme naturelle et non monotone")
	parts = append(parts, "Niveau détaillé")
	parts = append(parts, "Chemins cohérents")

	return strings.Join(parts, ". ") + "."
}

// BuildRegionalMapPrompt generates a prompt for a regional map.
func BuildRegionalMapPrompt(ctx *MapContext, opts PromptOptions) string {
	region := ctx.Region
	kingdom := ctx.Kingdom

	parts := []string{}

	// 1. Perspective
	parts = append(parts, "Cette carte montre la région")
	if region != nil {
		parts = append(parts, fmt.Sprintf("de %s", region.Name))
	}
	parts = append(parts, "vue du ciel, style carte géographique")

	// 2. Kingdom context
	if kingdom != nil {
		parts = append(parts, fmt.Sprintf("Territoire du royaume %s", kingdom.Name))
	}

	// 3. Geographic description
	if region != nil && region.Description != "" {
		parts = append(parts, region.Description)
	}

	// 4. Major settlements
	if region != nil && len(region.Cities) > 0 {
		cityNames := []string{}
		for _, city := range region.Cities {
			cityNames = append(cityNames, city.Name)
		}
		if len(cityNames) <= 5 {
			parts = append(parts, fmt.Sprintf("Villes principales : %s", strings.Join(cityNames, ", ")))
		} else {
			parts = append(parts, fmt.Sprintf("Villes principales : %s et d'autres", strings.Join(cityNames[:5], ", ")))
		}
	}

	// 5. Terrain features
	if opts.Terrain != "" {
		parts = append(parts, fmt.Sprintf("Terrain : %s", opts.Terrain))
	} else {
		// Infer from region description
		parts = append(parts, "Terrain : montagnes, forêts, plaines et côtes")
	}

	// 6. Roads and trade routes
	parts = append(parts, "Routes commerciales reliant les villes")
	parts = append(parts, "Chemins et sentiers secondaires")

	// 7. Borders
	if kingdom != nil {
		parts = append(parts, fmt.Sprintf("Frontières du royaume %s clairement marquées", kingdom.Name))
	}

	// 8. Scale indicators
	scaleDesc := "moyenne"
	if ctx.Scale == "large" {
		scaleDesc = "large, montrant plusieurs régions"
	} else if ctx.Scale == "small" {
		scaleDesc = "détaillée, focus sur une zone spécifique"
	}
	parts = append(parts, fmt.Sprintf("Échelle %s", scaleDesc))

	// 9. Cartographic elements
	parts = append(parts, "Avec légende cartographique")
	parts = append(parts, "Symboles géographiques (montagnes, rivières, forêts)"  )
	parts = append(parts, "Style carte médiévale fantasy")

	return strings.Join(parts, ". ") + "."
}

// BuildDungeonMapPrompt generates a prompt for a dungeon floor plan.
func BuildDungeonMapPrompt(name string, dungeonLevel int, opts PromptOptions) string {
	parts := []string{}

	// 1. Map type and perspective
	parts = append(parts, fmt.Sprintf("Plan de donjon en vue du dessus : %s", name))

	// 2. Level specification
	if dungeonLevel > 0 {
		parts = append(parts, fmt.Sprintf("Niveau %d", dungeonLevel))
	}

	// 3. Layout description
	parts = append(parts, "Salles de différentes tailles et formes")
	parts = append(parts, "Couloirs étroits et passages")
	parts = append(parts, "Portes et entrées clairement indiquées")

	// 4. Hazards and features
	parts = append(parts, "Pièges marqués par des symboles")
	parts = append(parts, "Portes secrètes (lignes pointillées)")
	parts = append(parts, "Escaliers vers autres niveaux")

	// 5. Specific features from options
	if len(opts.Features) > 0 {
		parts = append(parts, fmt.Sprintf("Éléments spéciaux : %s", strings.Join(opts.Features, ", ")))
	}

	// 6. Architecture style
	parts = append(parts, "Architecture de pierre médiévale")
	parts = append(parts, "Torches fixées aux murs")
	parts = append(parts, "Piliers dans grandes salles")

	// 7. Grid and scale
	parts = append(parts, "Grille au sol (carrés de 1.5m)")
	parts = append(parts, "Échelle indiquée en mètres")

	// 8. Style
	parts = append(parts, "Style plan de D&D classique")
	parts = append(parts, "Noir et blanc avec ombrage")

	return strings.Join(parts, ". ") + "."
}

// BuildTacticalMapPrompt generates a prompt for a tactical battle map.
func BuildTacticalMapPrompt(terrain string, sceneDescription string, opts PromptOptions) string {
	parts := []string{}

	// 1. Map type
	parts = append(parts, "Carte tactique de combat en vue du dessus")

	// 2. Scene context
	if sceneDescription != "" {
		parts = append(parts, sceneDescription)
	}

	// 3. Terrain type
	if terrain == "" {
		terrain = "terrain varié"
	}
	parts = append(parts, fmt.Sprintf("Terrain : %s", terrain))

	// 4. Terrain features based on type
	terrainFeatures := getTacticalTerrainFeatures(terrain)
	if len(terrainFeatures) > 0 {
		parts = append(parts, strings.Join(terrainFeatures, ", "))
	}

	// 5. Cover and obstacles
	parts = append(parts, "Éléments de couverture (rochers, arbres, murs)")
	parts = append(parts, "Obstacles variés")

	// 6. Elevation
	parts = append(parts, "Variations d'élévation marquées")
	parts = append(parts, "Zones de hauteur différente")

	// 7. Grid
	parts = append(parts, "Grille de combat (carrés de 1.5m)")
	parts = append(parts, "Format carré pour alignement des figurines")

	// 8. Specific features
	if len(opts.Features) > 0 {
		parts = append(parts, fmt.Sprintf("Éléments spéciaux : %s", strings.Join(opts.Features, ", ")))
	}

	// 9. Style
	parts = append(parts, "Style carte de combat D&D")
	parts = append(parts, "Couleurs distinctes pour zones différentes")
	parts = append(parts, "Lisible et pratique pour le jeu")

	return strings.Join(parts, ". ") + "."
}

// Helper functions

func getKingdomArchitecturalStyle(kingdomID string) string {
	styles := map[string]string{
		"valdorine":  "valdorin maritime avec influences italiennes",
		"karvath":    "karvath militaire avec fortifications germaniques",
		"lumenciel":  "lumenciel religieux avec architecture sacrée latine",
		"astrene":    "astrène mélancolique avec influences nordiques",
	}

	if style, ok := styles[strings.ToLower(kingdomID)]; ok {
		return style
	}

	return "médiéval fantasy générique"
}

func collectPOIs(loc *world.Location, extraFeatures []string) []string {
	pois := []string{}

	// Add POIs from location data
	if loc != nil {
		pois = append(pois, loc.KeyLocations...)
	}

	// Add extra features from options
	pois = append(pois, extraFeatures...)

	return pois
}

func getInfrastructure(cityType string) string {
	infra := map[string]string{
		"port majeur":          "Infrastructure : docks, entrepôts, chantier naval, phare",
		"port industriel":      "Infrastructure : chantiers navals, forges, entrepôts industriels",
		"port financier":       "Infrastructure : banques, coffres royaux, maisons de commerce",
		"forteresse capitale":  "Infrastructure : murailles épaisses, tours de garde, casernes, arsenal",
		"forteresse frontalière": "Infrastructure : remparts, tours de guet, camp militaire",
		"capitale":             "Infrastructure : palais royal, cathédrale, murailles, places publiques",
		"ville sainte":         "Infrastructure : cathédrale majeure, monastères, hospices",
	}

	cityType = strings.ToLower(cityType)
	for key, value := range infra {
		if strings.Contains(cityType, key) {
			return value
		}
	}

	return "Infrastructure : murailles, places publiques, marché central"
}

func getTacticalTerrainFeatures(terrain string) []string {
	features := map[string][]string{
		"forest": {
			"Arbres denses",
			"Sous-bois épais",
			"Clairières",
			"Souches et rondins",
		},
		"forêt": {
			"Arbres denses",
			"Sous-bois épais",
			"Clairières",
			"Souches et rondins",
		},
		"mountain": {
			"Rochers et pierres",
			"Pentes raides",
			"Passages étroits",
			"Précipices",
		},
		"montagne": {
			"Rochers et pierres",
			"Pentes raides",
			"Passages étroits",
			"Précipices",
		},
		"plains": {
			"Herbe haute",
			"Quelques arbres isolés",
			"Collines douces",
			"Ruisseau",
		},
		"plaine": {
			"Herbe haute",
			"Quelques arbres isolés",
			"Collines douces",
			"Ruisseau",
		},
		"swamp": {
			"Eau stagnante",
			"Arbres morts",
			"Zones boueuses",
			"Plantes aquatiques",
		},
		"marais": {
			"Eau stagnante",
			"Arbres morts",
			"Zones boueuses",
			"Plantes aquatiques",
		},
		"dungeon": {
			"Murs de pierre",
			"Piliers",
			"Portes et couloirs",
			"Torches murales",
		},
		"cave": {
			"Stalactites et stalagmites",
			"Passages étroits",
			"Eau souterraine",
			"Formations rocheuses",
		},
		"grotte": {
			"Stalactites et stalagmites",
			"Passages étroits",
			"Eau souterraine",
			"Formations rocheuses",
		},
	}

	terrainLower := strings.ToLower(terrain)
	for key, feats := range features {
		if strings.Contains(terrainLower, key) {
			return feats
		}
	}

	// Default generic features
	return []string{
		"Terrain varié",
		"Obstacles naturels",
		"Zones de couverture",
	}
}
