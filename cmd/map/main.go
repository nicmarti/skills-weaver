package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dungeons/internal/image"
	mapgen "dungeons/internal/map"
	"dungeons/internal/world"
)

const version = "1.0.0"

// MapPrompt holds the generated map prompt and metadata.
type MapPrompt struct {
	Prompt       string   `json:"prompt"`
	MapType      string   `json:"map_type"`
	LocationName string   `json:"location_name"`
	Kingdom      string   `json:"kingdom"`
	Features     []string `json:"features"`
	Scale        string   `json:"scale"`
	Style        string   `json:"style"`
	GeneratedAt  string   `json:"generated_at"`
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "generate", "gen":
		err = cmdGenerate(args)
	case "validate", "val":
		err = cmdValidate(args)
	case "list", "ls":
		err = cmdList(args)
	case "types":
		err = cmdTypes()
	case "help", "--help", "-h":
		showHelp()
	case "version", "--version", "-v":
		fmt.Printf("sw-map version %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	help := `sw-map - Générateur de prompts pour cartes 2D fantasy

USAGE:
    sw-map <command> [options]

COMMANDS:
    generate <type> <name>  Générer un prompt de carte détaillé
    validate <name>         Valider l'existence d'un lieu
    list [category]         Lister les ressources disponibles
    types                   Afficher les types de cartes disponibles
    help                    Afficher cette aide
    version                 Afficher la version

GENERATE OPTIONS:
    --kingdom=<id>          Royaume pour validation (valdorine, karvath, lumenciel, astrene)
    --style=<style>         Style visuel (illustrated, dark_fantasy) [défaut: illustrated]
    --scale=<scale>         Échelle (small, medium, large) [défaut: medium]
    --features=<list>       POIs additionnels (séparés par des virgules)
    --terrain=<type>        Type de terrain (forêt, montagne, plaine, etc.)
    --level=<n>             Niveau de donjon (pour type dungeon)
    --scene=<desc>          Description de scène (pour type tactical)
    --output=<file>         Sauvegarder dans un fichier JSON
    --generate-image        Générer aussi l'image via fal.ai flux-2
    --image-size=<size>     Taille d'image (square_hd, landscape_16_9, etc.)

VALIDATE OPTIONS:
    --kingdom=<id>          Royaume attendu
    --suggest               Afficher des suggestions de noms

LIST CATEGORIES:
    types                   Types de cartes disponibles
    kingdoms                Royaumes disponibles
    locations               Tous les lieux documentés
    cities                  Toutes les cités
    cities --kingdom=<id>   Cités d'un royaume spécifique

MAP TYPES:
    city       Carte détaillée de ville (vue aérienne)
    region     Carte régionale (bird's eye view)
    dungeon    Plan de donjon (top-down avec grille)
    tactical   Carte tactique de combat (grille 1.5m)

EXAMPLES:
    # Carte de ville
    sw-map generate city Cordova

    # Avec POIs additionnels
    sw-map generate city Cordova --features="Taverne du Voile Écarlate,Docks"

    # Carte régionale
    sw-map generate region "Côte Occidentale" --scale=large

    # Plan de donjon
    sw-map generate dungeon "La Crypte des Ombres" --level=1

    # Carte tactique
    sw-map generate tactical "Embuscade" --terrain=forêt --scene="Combat en forêt"

    # Avec génération d'image
    sw-map generate city Cordova --generate-image --image-size=landscape_16_9

    # Validation de lieu
    sw-map validate "Port-Nouveau" --kingdom=valdorine --suggest

    # Lister les ressources
    sw-map list kingdoms
    sw-map list cities --kingdom=valdorine
    sw-map types

ENVIRONMENT:
    FAL_KEY                 Requis pour génération d'images (--generate-image)

DATA LOCATION:
    Prompts cache: data/maps/<name>_<type>_<scale>_prompt.json
    Images:        data/maps/<name>_<type>_<scale>.png
`
	fmt.Print(help)
}

func cmdGenerate(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: sw-map generate <type> <name> [options]")
	}

	mapType := args[0]
	name := args[1]

	// Parse options
	var (
		kingdom       string
		style         = "illustrated"
		scale         = "medium"
		features      []string
		terrain       string
		level         int
		scene         string
		outputFile    string
		generateImage bool
		imageSize     = "landscape_16_9"
	)

	for i := 2; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--kingdom=") {
			kingdom = strings.TrimPrefix(arg, "--kingdom=")
		} else if strings.HasPrefix(arg, "--style=") {
			style = strings.TrimPrefix(arg, "--style=")
		} else if strings.HasPrefix(arg, "--scale=") {
			scale = strings.TrimPrefix(arg, "--scale=")
		} else if strings.HasPrefix(arg, "--features=") {
			featStr := strings.TrimPrefix(arg, "--features=")
			features = strings.Split(featStr, ",")
			// Trim spaces
			for i, f := range features {
				features[i] = strings.TrimSpace(f)
			}
		} else if strings.HasPrefix(arg, "--terrain=") {
			terrain = strings.TrimPrefix(arg, "--terrain=")
		} else if strings.HasPrefix(arg, "--level=") {
			fmt.Sscanf(strings.TrimPrefix(arg, "--level="), "%d", &level)
		} else if strings.HasPrefix(arg, "--scene=") {
			scene = strings.TrimPrefix(arg, "--scene=")
		} else if strings.HasPrefix(arg, "--output=") {
			outputFile = strings.TrimPrefix(arg, "--output=")
		} else if arg == "--generate-image" {
			generateImage = true
		} else if strings.HasPrefix(arg, "--image-size=") {
			imageSize = strings.TrimPrefix(arg, "--image-size=")
		}
	}

	// Validate map type
	validTypes := map[string]bool{"city": true, "region": true, "dungeon": true, "tactical": true}
	if !validTypes[mapType] {
		return fmt.Errorf("invalid map type: %s (valid: city, region, dungeon, tactical)", mapType)
	}

	// Load world data
	dataDir := "data"
	geo, err := world.LoadGeography(dataDir)
	if err != nil {
		return fmt.Errorf("loading geography: %w", err)
	}

	factions, err := world.LoadFactions(dataDir)
	if err != nil {
		return fmt.Errorf("loading factions: %w", err)
	}

	// Validate and get location data (for city/region types)
	var location *world.Location
	var kingdomData *world.Kingdom

	if mapType == "city" || mapType == "region" {
		exists, loc, _, err := world.ValidateLocationExists(name, geo)
		if !exists {
			// Provide suggestions
			suggestions := world.GetSuggestions(name, geo, 5)
			if len(suggestions) > 0 {
				fmt.Fprintf(os.Stderr, "✗ Lieu \"%s\" non trouvé dans geography.json\n\n", name)
				fmt.Fprintf(os.Stderr, "Vouliez-vous dire ?\n")
				for _, s := range suggestions {
					fmt.Fprintf(os.Stderr, "  - %s (%s)\n", s.Location.Name, s.Location.Kingdom)
				}
				fmt.Fprintf(os.Stderr, "\nAstuce: Utilisez 'sw-map list cities' pour voir toutes les villes.\n")
			}
			return err
		}
		location = loc

		// Get kingdom data
		if location.Kingdom != "" {
			_, kd, err := world.ValidateKingdomExists(strings.ToLower(location.Kingdom), factions)
			if err == nil {
				kingdomData = kd
			}
		}

		// Validate kingdom if specified
		if kingdom != "" && location.Kingdom != "" {
			if strings.ToLower(location.Kingdom) != strings.ToLower(kingdom) {
				return fmt.Errorf("location %s belongs to %s, not %s", name, location.Kingdom, kingdom)
			}
		}
	} else {
		// For dungeon/tactical, kingdom is optional
		if kingdom != "" {
			_, kd, err := world.ValidateKingdomExists(strings.ToLower(kingdom), factions)
			if err != nil {
				return fmt.Errorf("invalid kingdom: %w", err)
			}
			kingdomData = kd
		}
	}

	// Build prompt context
	ctx := &mapgen.MapContext{
		Location: location,
		Kingdom:  kingdomData,
		Scale:    scale,
	}
	opts := mapgen.PromptOptions{
		Features: features,
		Terrain:  terrain,
		Style:    style,
	}

	// Generate prompt based on type
	var prompt string
	switch mapType {
	case "city":
		prompt = mapgen.BuildCityMapPrompt(ctx, opts)
	case "region":
		prompt = mapgen.BuildRegionalMapPrompt(ctx, opts)
	case "dungeon":
		prompt = mapgen.BuildDungeonMapPrompt(name, level, opts)
	case "tactical":
		prompt = mapgen.BuildTacticalMapPrompt(terrain, scene, opts)
	}

	// Create result structure
	result := &MapPrompt{
		Prompt:       prompt,
		MapType:      mapType,
		LocationName: name,
		Scale:        scale,
		Style:        style,
		Features:     features,
		GeneratedAt:  time.Now().Format(time.RFC3339),
	}
	if kingdomData != nil {
		result.Kingdom = kingdomData.ID
	}

	// Display result
	fmt.Println("✓ Prompt généré")
	fmt.Printf("\nPROMPT (%s):\n%s\n", result.MapType, result.Prompt)
	fmt.Printf("\nMétadonnées:\n")
	if result.LocationName != "" {
		fmt.Printf("  Location: %s\n", result.LocationName)
	}
	if result.Kingdom != "" {
		fmt.Printf("  Kingdom: %s\n", result.Kingdom)
	}
	if len(result.Features) > 0 {
		fmt.Printf("  Features: %s\n", strings.Join(result.Features, ", "))
	}
	fmt.Printf("  Scale: %s\n", result.Scale)
	fmt.Printf("  Style: %s\n", result.Style)
	fmt.Printf("  Generated: %s\n", result.GeneratedAt)

	// Save to debug file
	debugFile := getDebugFile(name, mapType, scale)
	if err := saveCachedPrompt(debugFile, result); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save debug file: %v\n", err)
	} else {
		fmt.Printf("\n✓ Prompt sauvegardé dans: %s\n", debugFile)
	}

	// Save to output file if specified
	if outputFile != "" {
		if err := saveCachedPrompt(outputFile, result); err != nil {
			return fmt.Errorf("saving to output file: %w", err)
		}
		fmt.Printf("✓ Prompt sauvegardé dans: %s\n", outputFile)
	}

	// Generate image if requested
	if generateImage {
		return generateMapImage(result.Prompt, name, mapType, scale, imageSize)
	}

	return nil
}

func getDebugFile(name, mapType, scale string) string {
	// Sanitize name for filename
	safeName := strings.ReplaceAll(name, " ", "_")
	safeName = strings.ToLower(safeName)
	return filepath.Join("data", "maps", fmt.Sprintf("%s_%s_%s_prompt.json", safeName, mapType, scale))
}

func saveCachedPrompt(cacheFile string, result *MapPrompt) error {
	// Ensure directory exists
	dir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

func generateMapImage(prompt, name, mapType, scale, imageSize string) error {
	fmt.Println("\n⏳ Génération de l'image...")

	// Create image generator
	outputDir := filepath.Join("data", "maps")
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return fmt.Errorf("creating image generator: %w\nAstuce: Définissez FAL_KEY dans votre environnement", err)
	}

	// Generate unique filename
	safeName := strings.ReplaceAll(name, " ", "_")
	safeName = strings.ToLower(safeName)
	filename := fmt.Sprintf("%s_%s_%s", safeName, mapType, scale)

	opts := []image.Option{
		image.WithModelInstance(image.ModelFluxPro11),
		image.WithImageSize(imageSize),
		image.WithNumImages(1),
		image.WithFilenamePrefix(filename),
	}

	result, err := gen.Generate(prompt, opts...)
	if err != nil {
		return fmt.Errorf("generating image: %w", err)
	}

	fmt.Printf("✓ Image générée: %s\n", result.LocalPath)
	fmt.Printf("  URL: %s\n", result.URL)
	fmt.Printf("  Dimensions: %dx%d\n", result.Width, result.Height)

	return nil
}

func cmdValidate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sw-map validate <name> [options]")
	}

	name := args[0]
	var kingdom string
	var suggest bool

	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--kingdom=") {
			kingdom = strings.TrimPrefix(arg, "--kingdom=")
		} else if arg == "--suggest" {
			suggest = true
		}
	}

	// Load geography
	geo, err := world.LoadGeography("data")
	if err != nil {
		return fmt.Errorf("loading geography: %w", err)
	}

	// Validate location
	exists, loc, region, err := world.ValidateLocationExists(name, geo)

	if !exists {
		fmt.Printf("✗ Lieu \"%s\" non trouvé\n\n", name)

		if suggest {
			suggestions := world.GetSuggestions(name, geo, 10)
			if len(suggestions) > 0 {
				fmt.Println("Suggestions:")
				for _, s := range suggestions {
					fmt.Printf("  - %s (%s, %s) - score: %d\n",
						s.Location.Name, s.Location.Type, s.Location.Kingdom, s.Score)
				}
			} else {
				fmt.Println("Aucune suggestion trouvée.")
			}
		}

		return err
	}

	// Location exists
	fmt.Printf("✓ Lieu validé: %s\n\n", loc.Name)
	fmt.Printf("Type: %s\n", loc.Type)
	fmt.Printf("Royaume: %s\n", loc.Kingdom)
	if region != nil {
		fmt.Printf("Région: %s\n", region.Name)
	}
	if loc.Population != "" {
		fmt.Printf("Population: %s\n", loc.Population)
	}
	if loc.Description != "" {
		fmt.Printf("Description: %s\n", loc.Description)
	}
	if len(loc.KeyLocations) > 0 {
		fmt.Printf("Points d'intérêt: %s\n", strings.Join(loc.KeyLocations, ", "))
	}

	// Validate kingdom if specified
	if kingdom != "" {
		if strings.ToLower(loc.Kingdom) != strings.ToLower(kingdom) {
			return fmt.Errorf("\n✗ Le lieu appartient au royaume %s, pas %s", loc.Kingdom, kingdom)
		}
		fmt.Printf("\n✓ Royaume validé: %s\n", loc.Kingdom)
	}

	return nil
}

func cmdList(args []string) error {
	category := "types"
	if len(args) > 0 {
		category = args[0]
	}

	var kingdom string
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--kingdom=") {
			kingdom = strings.TrimPrefix(arg, "--kingdom=")
		}
	}

	switch category {
	case "types":
		fmt.Println("Types de cartes disponibles:")
		fmt.Println("  city      - Carte détaillée de ville (vue aérienne)")
		fmt.Println("              Districts, POIs, infrastructure")
		fmt.Println("              Nécessite: location existante")
		fmt.Println()
		fmt.Println("  region    - Carte régionale (bird's eye view)")
		fmt.Println("              Multiple settlements, routes, terrain")
		fmt.Println("              Nécessite: location existante")
		fmt.Println()
		fmt.Println("  dungeon   - Plan de donjon (top-down)")
		fmt.Println("              Salles, couloirs, pièges, grille 1.5m")
		fmt.Println("              Nécessite: nom, niveau optionnel")
		fmt.Println()
		fmt.Println("  tactical  - Carte tactique de combat (grille)")
		fmt.Println("              Terrain, couverture, obstacles, élévation")
		fmt.Println("              Nécessite: terrain, scène optionnelle")

	case "kingdoms":
		factions, err := world.LoadFactions("data")
		if err != nil {
			return fmt.Errorf("loading factions: %w", err)
		}

		fmt.Println("Royaumes disponibles:")
		for _, k := range factions.Kingdoms {
			fmt.Printf("  %s - %s\n", k.ID, k.Name)
			fmt.Printf("    Capitale: %s\n", k.Capital)
			fmt.Printf("    Couleurs: %s\n", strings.Join(k.Colors, ", "))
			fmt.Printf("    Devise: %s\n", k.Motto)
			fmt.Println()
		}

	case "locations":
		geo, err := world.LoadGeography("data")
		if err != nil {
			return fmt.Errorf("loading geography: %w", err)
		}

		fmt.Println("Lieux documentés:")
		for _, continent := range geo.Continents {
			for _, region := range continent.Regions {
				for _, city := range region.Cities {
					if kingdom != "" && strings.ToLower(city.Kingdom) != strings.ToLower(kingdom) {
						continue
					}
					fmt.Printf("  %s (%s, %s)\n", city.Name, city.Type, city.Kingdom)
				}
			}
		}

	case "cities":
		geo, err := world.LoadGeography("data")
		if err != nil {
			return fmt.Errorf("loading geography: %w", err)
		}

		if kingdom != "" {
			fmt.Printf("Cités du royaume %s:\n\n", kingdom)
		} else {
			fmt.Println("Toutes les cités:")
			fmt.Println()
		}

		for _, continent := range geo.Continents {
			for _, region := range continent.Regions {
				for _, city := range region.Cities {
					if kingdom != "" && strings.ToLower(city.Kingdom) != strings.ToLower(kingdom) {
						continue
					}
					fmt.Printf("  %s - %s\n", city.Name, city.Type)

				}
			}
		}

	default:
		return fmt.Errorf("unknown category: %s (valid: types, kingdoms, locations, cities)", category)
	}

	return nil
}

func cmdTypes() error {
	fmt.Println("Types de Cartes - Structures et Spécifications")
	fmt.Println()

	fmt.Println("=== CITY (Carte de Ville) ===")
	fmt.Println("Perspective: Vue aérienne")
	fmt.Println("Structure: [Kingdom Style] + [Geography] + [Districts] + [POIs] + [Infrastructure]")
	fmt.Println("Éléments clés:")
	fmt.Println("  - Style architectural du royaume")
	fmt.Println("  - Géographie (côtes, rivières, relief)")
	fmt.Println("  - Quartiers organiques (pas de grille)")
	fmt.Println("  - Points d'intérêt nommés")
	fmt.Println("  - Infrastructure (ports, murailles, routes)")
	fmt.Println("Usage: sw-map generate city <nom>")
	fmt.Println()

	fmt.Println("=== REGION (Carte Régionale) ===")
	fmt.Println("Perspective: Bird's eye view")
	fmt.Println("Structure: [Territory] + [Settlements] + [Roads] + [Terrain] + [Borders]")
	fmt.Println("Éléments clés:")
	fmt.Println("  - Multiple settlements (villes, villages)")
	fmt.Println("  - Routes commerciales")
	fmt.Println("  - Terrain varié (montagnes, forêts, plaines)")
	fmt.Println("  - Frontières du royaume")
	fmt.Println("  - Style cartographique médiéval")
	fmt.Println("Usage: sw-map generate region <nom> --scale=large")
	fmt.Println()

	fmt.Println("=== DUNGEON (Plan de Donjon) ===")
	fmt.Println("Perspective: Top-down")
	fmt.Println("Structure: [Level] + [Rooms] + [Corridors] + [Hazards] + [Grid]")
	fmt.Println("Éléments clés:")
	fmt.Println("  - Salles de différentes tailles")
	fmt.Println("  - Couloirs et passages")
	fmt.Println("  - Pièges marqués (X, !, △)")
	fmt.Println("  - Portes secrètes (lignes pointillées)")
	fmt.Println("  - Grille 1.5m pour figurines")
	fmt.Println("Usage: sw-map generate dungeon <nom> --level=1")
	fmt.Println()

	fmt.Println("=== TACTICAL (Carte Tactique) ===")
	fmt.Println("Perspective: Top-down avec grille")
	fmt.Println("Structure: [Terrain] + [Cover] + [Obstacles] + [Elevation] + [Grid]")
	fmt.Println("Éléments clés:")
	fmt.Println("  - Type de terrain (forêt, montagne, etc.)")
	fmt.Println("  - Couverture (totale, partielle)")
	fmt.Println("  - Obstacles et terrain difficile")
	fmt.Println("  - Variations d'élévation")
	fmt.Println("  - Grille 1.5m (format carré)")
	fmt.Println("Usage: sw-map generate tactical <nom> --terrain=forêt")
	fmt.Println()

	fmt.Println("LONGUEURS DE PROMPTS:")
	fmt.Println("  Minimum: 80 mots")
	fmt.Println("  Cible: 100-200 mots (idéal: 150)")
	fmt.Println("  Maximum: 250 mots")

	return nil
}
