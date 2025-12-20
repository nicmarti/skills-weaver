package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"dungeons/internal/character"
	"dungeons/internal/image"
	"dungeons/internal/npc"
)

const (
	outputDir = "data/images"
	dataDir   = "data"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "character", "char":
		err = cmdCharacter(args)
	case "npc":
		err = cmdNPC(args)
	case "scene":
		err = cmdScene(args)
	case "monster":
		err = cmdMonster(args)
	case "item":
		err = cmdItem(args)
	case "location":
		err = cmdLocation(args)
	case "custom":
		err = cmdCustom(args)
	case "list":
		err = cmdList(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Commande inconnue: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Image Generator - Générateur d'images Heroic Fantasy

UTILISATION:
  image <commande> [arguments]

COMMANDES:
  character <nom>              Générer le portrait d'un personnage sauvegardé
  npc [options]                Générer le portrait d'un PNJ
  scene <description>          Générer une scène d'aventure
  monster <type>               Générer une illustration de monstre
  item <type> [description]    Générer une illustration d'objet magique
  location <type> [nom]        Générer une vue de lieu
  custom <prompt>              Générer avec un prompt personnalisé
  list [styles|scenes|monsters|items|locations]  Lister les options
  help                         Afficher cette aide

OPTIONS COMMUNES:
  --style=<style>              Style artistique (realistic, painted, illustrated, dark_fantasy, epic)
  --size=<size>                Taille d'image (square_hd, portrait_4_3, landscape_16_9, etc.)
  --format=<format>            Format de sortie (png, jpeg, webp)

OPTIONS NPC:
  --race=<race>                Race du PNJ (human, dwarf, elf, halfling)
  --gender=<m|f>               Sexe du PNJ
  --occupation=<type>          Occupation du PNJ

OPTIONS SCENE:
  --type=<type>                Type de scène (tavern, dungeon, forest, castle, etc.)

EXEMPLES:
  image character "Aldric"                     # Portrait du personnage Aldric
  image npc --race=dwarf --occupation=skilled  # Portrait d'un artisan nain
  image scene "bataille contre un dragon" --type=battle
  image monster dragon --style=epic
  image item weapon "épée flamboyante"
  image location dungeon "Les Mines de Moria"
  image custom "Un elfe archer dans une forêt enchantée"

NOTES:
  - Nécessite la variable d'environnement FAL_KEY
  - Les images sont sauvegardées dans data/images/`)
}

func cmdCharacter(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("nom du personnage requis")
	}

	opts := parseOptions(args[1:])
	charName := args[0]

	// Build character file path
	charPath := fmt.Sprintf("%s/characters/%s.json", dataDir, strings.ToLower(strings.ReplaceAll(charName, " ", "_")))

	// Load character
	char, err := character.Load(charPath)
	if err != nil {
		return fmt.Errorf("chargement du personnage '%s': %w", charName, err)
	}

	// Create generator
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleIllustrated
	}
	prompt := image.BuildCharacterPrompt(char, style)

	fmt.Printf("Génération du portrait de %s...\n", charName)
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Generate image
	imgOpts := buildImageOptions(opts)
	result, err := gen.Generate(prompt, imgOpts...)
	if err != nil {
		return err
	}

	fmt.Printf("Image générée: %s\n", result.LocalPath)
	fmt.Printf("Dimensions: %dx%d\n", result.Width, result.Height)

	return nil
}

func cmdNPC(args []string) error {
	opts := parseOptions(args)

	// Create NPC generator
	npcGen, err := npc.NewGenerator(dataDir)
	if err != nil {
		return fmt.Errorf("création du générateur NPC: %w", err)
	}

	// Generate NPC with options
	var npcOpts []npc.Option
	if opts["race"] != "" {
		npcOpts = append(npcOpts, npc.WithRace(opts["race"]))
	}
	if opts["gender"] != "" {
		npcOpts = append(npcOpts, npc.WithGender(opts["gender"]))
	}
	if opts["occupation"] != "" {
		npcOpts = append(npcOpts, npc.WithOccupationType(opts["occupation"]))
	}

	n, err := npcGen.Generate(npcOpts...)
	if err != nil {
		return fmt.Errorf("génération du PNJ: %w", err)
	}

	fmt.Printf("PNJ généré: %s (%s %s, %s)\n\n", n.Name, n.Race, n.Gender, n.Occupation)

	// Create image generator
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleIllustrated
	}
	prompt := image.BuildNPCPrompt(n, style)

	fmt.Printf("Génération du portrait...\n")
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Generate image
	imgOpts := buildImageOptions(opts)
	result, err := gen.Generate(prompt, imgOpts...)
	if err != nil {
		return err
	}

	fmt.Printf("Image générée: %s\n", result.LocalPath)
	fmt.Printf("Dimensions: %dx%d\n", result.Width, result.Height)

	return nil
}

func cmdScene(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("description de la scène requise")
	}

	// Find description (first non-option argument)
	var description string
	var optArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			optArgs = append(optArgs, arg)
		} else if description == "" {
			description = arg
		}
	}

	opts := parseOptions(optArgs)

	// Create generator
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleEpic
	}
	sceneType := opts["type"]
	prompt := image.BuildScenePrompt(description, sceneType, style)

	fmt.Printf("Génération de la scène...\n")
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Generate image
	imgOpts := buildImageOptions(opts)
	if opts["size"] == "" {
		imgOpts = append(imgOpts, image.WithImageSize("landscape_16_9"))
	}
	result, err := gen.Generate(prompt, imgOpts...)
	if err != nil {
		return err
	}

	fmt.Printf("Image générée: %s\n", result.LocalPath)
	fmt.Printf("Dimensions: %dx%d\n", result.Width, result.Height)

	return nil
}

func cmdMonster(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("type de monstre requis")
	}

	monsterType := args[0]
	opts := parseOptions(args[1:])

	// Create generator
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleDarkFantasy
	}
	prompt := image.BuildMonsterPrompt(monsterType, style)

	fmt.Printf("Génération du monstre: %s...\n", monsterType)
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Generate image
	imgOpts := buildImageOptions(opts)
	result, err := gen.Generate(prompt, imgOpts...)
	if err != nil {
		return err
	}

	fmt.Printf("Image générée: %s\n", result.LocalPath)
	fmt.Printf("Dimensions: %dx%d\n", result.Width, result.Height)

	return nil
}

func cmdItem(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("type d'objet requis")
	}

	itemType := args[0]
	var description string
	var optArgs []string

	for i, arg := range args[1:] {
		if strings.HasPrefix(arg, "--") {
			optArgs = append(optArgs, arg)
		} else if i == 0 {
			description = arg
		}
	}

	opts := parseOptions(optArgs)

	// Create generator
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StylePainted
	}
	prompt := image.BuildItemPrompt(itemType, description, style)

	fmt.Printf("Génération de l'objet: %s...\n", itemType)
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Generate image
	imgOpts := buildImageOptions(opts)
	if opts["size"] == "" {
		imgOpts = append(imgOpts, image.WithImageSize("square_hd"))
	}
	result, err := gen.Generate(prompt, imgOpts...)
	if err != nil {
		return err
	}

	fmt.Printf("Image générée: %s\n", result.LocalPath)
	fmt.Printf("Dimensions: %dx%d\n", result.Width, result.Height)

	return nil
}

func cmdLocation(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("type de lieu requis")
	}

	locationType := args[0]
	var name string
	var optArgs []string

	for i, arg := range args[1:] {
		if strings.HasPrefix(arg, "--") {
			optArgs = append(optArgs, arg)
		} else if i == 0 {
			name = arg
		}
	}

	opts := parseOptions(optArgs)

	// Create generator
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StylePainted
	}
	prompt := image.BuildLocationPrompt(locationType, name, style)

	fmt.Printf("Génération du lieu: %s...\n", locationType)
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Generate image
	imgOpts := buildImageOptions(opts)
	if opts["size"] == "" {
		imgOpts = append(imgOpts, image.WithImageSize("landscape_16_9"))
	}
	result, err := gen.Generate(prompt, imgOpts...)
	if err != nil {
		return err
	}

	fmt.Printf("Image générée: %s\n", result.LocalPath)
	fmt.Printf("Dimensions: %dx%d\n", result.Width, result.Height)

	return nil
}

func cmdCustom(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("prompt personnalisé requis")
	}

	// Join non-option arguments as prompt
	var promptParts []string
	var optArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			optArgs = append(optArgs, arg)
		} else {
			promptParts = append(promptParts, arg)
		}
	}

	prompt := strings.Join(promptParts, " ")
	opts := parseOptions(optArgs)

	// Add base suffix for consistency
	prompt = prompt + ", " + image.BasePromptSuffix

	// Create generator
	gen, err := image.NewGenerator(outputDir)
	if err != nil {
		return err
	}

	fmt.Printf("Génération avec prompt personnalisé...\n")
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Generate image
	imgOpts := buildImageOptions(opts)
	result, err := gen.Generate(prompt, imgOpts...)
	if err != nil {
		return err
	}

	fmt.Printf("Image générée: %s\n", result.LocalPath)
	fmt.Printf("Dimensions: %dx%d\n", result.Width, result.Height)

	return nil
}

func cmdList(args []string) error {
	listType := "all"
	if len(args) > 0 {
		listType = strings.ToLower(args[0])
	}

	switch listType {
	case "styles", "style":
		fmt.Println("## Styles Disponibles\n")
		for _, style := range image.GetAvailableStyles() {
			fmt.Printf("- %s\n", style)
		}

	case "scenes", "scene":
		fmt.Println("## Types de Scènes Disponibles\n")
		for _, scene := range image.GetAvailableSceneTypes() {
			fmt.Printf("- %s\n", scene)
		}

	case "monsters", "monster":
		fmt.Println("## Types de Monstres Disponibles\n")
		for _, monster := range image.GetAvailableMonsterTypes() {
			fmt.Printf("- %s\n", monster)
		}

	case "items", "item":
		fmt.Println("## Types d'Objets Disponibles\n")
		for _, item := range image.GetAvailableItemTypes() {
			fmt.Printf("- %s\n", item)
		}

	case "locations", "location":
		fmt.Println("## Types de Lieux Disponibles\n")
		for _, loc := range image.GetAvailableLocationTypes() {
			fmt.Printf("- %s\n", loc)
		}

	case "sizes", "size":
		fmt.Println("## Tailles d'Image Disponibles\n")
		for _, size := range image.GetAvailableImageSizes() {
			fmt.Printf("- %s\n", size)
		}

	default:
		fmt.Println("## Styles Disponibles\n")
		for _, style := range image.GetAvailableStyles() {
			fmt.Printf("- %s\n", style)
		}

		fmt.Println("\n## Types de Scènes\n")
		for _, scene := range image.GetAvailableSceneTypes() {
			fmt.Printf("- %s\n", scene)
		}

		fmt.Println("\n## Types de Monstres\n")
		for _, monster := range image.GetAvailableMonsterTypes() {
			fmt.Printf("- %s\n", monster)
		}

		fmt.Println("\n## Types d'Objets\n")
		for _, item := range image.GetAvailableItemTypes() {
			fmt.Printf("- %s\n", item)
		}

		fmt.Println("\n## Types de Lieux\n")
		for _, loc := range image.GetAvailableLocationTypes() {
			fmt.Printf("- %s\n", loc)
		}

		fmt.Println("\n## Tailles d'Image\n")
		for _, size := range image.GetAvailableImageSizes() {
			fmt.Printf("- %s\n", size)
		}
	}

	return nil
}

func parseOptions(args []string) map[string]string {
	opts := make(map[string]string)

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			parts := strings.SplitN(strings.TrimPrefix(arg, "--"), "=", 2)
			if len(parts) == 2 {
				opts[parts[0]] = parts[1]
			} else {
				opts[parts[0]] = "true"
			}
		}
	}

	return opts
}

func buildImageOptions(opts map[string]string) []image.Option {
	var imgOpts []image.Option

	if opts["size"] != "" {
		imgOpts = append(imgOpts, image.WithImageSize(opts["size"]))
	}
	if opts["format"] != "" {
		imgOpts = append(imgOpts, image.WithOutputFormat(opts["format"]))
	}

	return imgOpts
}

// Helper to output JSON
func outputJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
