package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"dungeons/internal/image"
	"dungeons/internal/npc"
	"dungeons/internal/tarot"
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
	case "journal":
		err = cmdJournal(args)
	case "tarot-deck":
		err = cmdTarotDeck(args)
	case "voyante-assets":
		err = cmdVoyanteAssets(args)
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
  sw-image <commande> [arguments]

COMMANDES:
  character <nom>              Générer le portrait d'un personnage sauvegardé
  npc [options]                Générer le portrait d'un PNJ
  scene <description>          Générer une scène d'aventure
  monster <type>               Générer une illustration de monstre
  item <type> [description]    Générer une illustration d'objet magique
  location <type> [nom]        Générer une vue de lieu
  custom <prompt>              Générer avec un prompt personnalisé
  journal <aventure>           Illustrer le journal d'une aventure (parallèle)
  tarot-deck                   Générer le jeu de cartes de la Voyante (wizard de création)
  voyante-assets               Générer les visuels d'interface de la scène de la Voyante
  list [styles|scenes|...]     Lister les options
  help                         Afficher cette aide

OPTIONS COMMUNES:
  --style=<style>              Style artistique (realistic, painted, illustrated, dark_fantasy, epic)
  --size=<size>                Taille d'image (square_hd, portrait_4_3, landscape_16_9, etc.)
  --format=<format>            Format de sortie (png, jpeg, webp)
  --output=<dir>               Répertoire de sortie (défaut: data/images)

OPTIONS JOURNAL:
  --types=<types>              Types à illustrer (combat,exploration,story,discovery,loot,session)
  --start-id=<n>               ID de départ pour reprendre depuis une entrée (optionnel)
  --max=<n>                    Nombre maximum d'images à générer
  --parallel=<n>               Nombre de générations en parallèle (défaut: 4)
  --model=<model>              Modèle fal.ai (seedream, zimage) défaut: seedream
  --dry-run                    Afficher les prompts sans générer

MODÈLES JOURNAL:
  seedream                     Seedream v4 - Haute qualité (~8s), ~$0.01/megapixel (DÉFAUT)
  zimage                       Z-Image Turbo - Rapide (~2s), ~$0.005/megapixel

OPTIONS TAROT-DECK:
  --provider=<auto|fal|google> Fournisseur d'images (défaut: auto = GEMINI puis FAL)
  --model=<model>              Modèle (honoré par fal.ai ; ignoré par Google Imagen)
  --output=<dir>               Dossier de sortie (défaut: web/static/cards)
  --deck=<path>                Fichier deck (défaut: data/tarot-deck.json)
  --force                      Régénérer même les cartes déjà présentes
  --dry-run                    Afficher les prompts sans générer
  (par défaut, les cartes déjà générées sont ignorées lors d'une ré-exécution)

EXEMPLES:
  sw-image character "Aldric"
  sw-image journal "la-crypte-des-ombres"
  sw-image journal "la-crypte-des-ombres" --start-id=60
  sw-image journal "la-crypte-des-ombres" --model=seedream --max=5
  sw-image journal "la-crypte-des-ombres" --model=zimage --start-id=60 --dry-run
  sw-image tarot-deck --dry-run
  sw-image tarot-deck --provider=fal --model=flux-2-pro
  sw-image tarot-deck --provider=google
  sw-image voyante-assets --provider=fal
  sw-image voyante-assets --dry-run

NOTES:
  - Nécessite GEMINI_API_KEY (Google Imagen, prioritaire) ou FAL_KEY (fal.ai, fallback)
  - Les images sont sauvegardées dans data/images/ ou data/adventures/<nom>/images/`)
}

func cmdCharacter(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("nom du personnage requis")
	}

	opts := parseOptions(args[1:])
	charName := args[0]

	// Set defaults for character portraits
	if opts["style"] == "" {
		opts["style"] = "painted"
	}
	if opts["size"] == "" {
		opts["size"] = "portrait_4_3"
	}
	if opts["model"] == "" {
		opts["model"] = "banana"
	}

	// Output directory (default: data/images, can be overridden with --output)
	outDir := outputDir
	if opts["output"] != "" {
		outDir = opts["output"]
	}

	// Build character file path
	charPath := fmt.Sprintf("%s/characters/%s.json", dataDir, character.SanitizeFilename(charName))

	// Load character
	char, err := character.Load(charPath)
	if err != nil {
		return fmt.Errorf("chargement du personnage '%s': %w", charName, err)
	}

	// Create generator with specified output directory
	gen, err := image.NewGeneratorAuto(outDir)
	if err != nil {
		return err
	}

	// Build prompt with specified style
	style := image.PromptStyle(opts["style"])
	if style == "" {
		opts["style"] = string(image.StyleIllustrated)
		style = image.StyleIllustrated
	}
	prompt := image.BuildCharacterPrompt(char, style)

	// For banana model (default for character portraits), add neutral background directive
	model := opts["model"]
	if model == "" {
		model = "banana" // Default for character portraits
	}
	if model == "banana" {
		prompt = prompt + ", background is a simple neutral color like black or white, character focus"
	}

	fmt.Printf("Génération du portrait de %s...\n", charName)
	printGenerationInfo(opts, "square_hd")
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
	gen, err := image.NewGeneratorAuto(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleIllustrated
		opts["style"] = string(style)
	}
	prompt := image.BuildNPCPrompt(n, style)

	// For banana model (default), add neutral background directive for portraits
	model := opts["model"]
	if model == "" || model == "banana" {
		prompt = prompt + ", neutral background, character focus"
	}

	fmt.Printf("Génération du portrait...\n")
	printGenerationInfo(opts, "square_hd")
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
	gen, err := image.NewGeneratorAuto(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleIllustrated
		opts["style"] = string(style)
	}
	sceneType := opts["type"]
	prompt := image.BuildScenePrompt(description, sceneType, style)

	// Set default model for scenes
	if opts["model"] == "" {
		opts["model"] = "banana"
	}

	fmt.Printf("Génération de la scène...\n")
	printGenerationInfo(opts, "landscape_16_9")
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
	gen, err := image.NewGeneratorAuto(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleDarkFantasy
		opts["style"] = string(style)
	}
	prompt := image.BuildMonsterPrompt(monsterType, style)

	fmt.Printf("Génération du monstre: %s...\n", monsterType)
	printGenerationInfo(opts, "square_hd")
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
	gen, err := image.NewGeneratorAuto(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleIllustrated
		opts["style"] = string(style)
	}
	prompt := image.BuildItemPrompt(itemType, description, style)

	fmt.Printf("Génération de l'objet: %s...\n", itemType)
	printGenerationInfo(opts, "square_hd")
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
	gen, err := image.NewGeneratorAuto(outputDir)
	if err != nil {
		return err
	}

	// Build prompt
	style := image.PromptStyle(opts["style"])
	if style == "" {
		style = image.StyleIllustrated
		opts["style"] = string(style)
	}
	prompt := image.BuildLocationPrompt(locationType, name, style)

	// Set default model for locations (used for maps)
	if opts["model"] == "" {
		opts["model"] = "flux-pro-11"
	}

	fmt.Printf("Génération du lieu: %s...\n", locationType)
	printGenerationInfo(opts, "landscape_16_9")
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
	gen, err := image.NewGeneratorAuto(outputDir)
	if err != nil {
		return err
	}

	fmt.Printf("Génération avec prompt personnalisé...\n")
	printGenerationInfo(opts, "square_hd")
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

// newImageProvider selects an image generator. fal.ai honors --model; Google
// Imagen ignores it. "auto"/"" keeps the GEMINI > FAL priority.
func newImageProvider(provider, outDir string) (image.GeneratorI, error) {
	switch provider {
	case "fal":
		return image.NewGenerator(outDir)
	case "google", "gemini":
		return image.NewGoogleGenerator(outDir)
	case "auto", "":
		return image.NewGeneratorAuto(outDir)
	default:
		return nil, fmt.Errorf("fournisseur inconnu %q (attendu: auto, fal, google)", provider)
	}
}

// buildTarotPrompt builds the art prompt for a tarot card from the wizard deck.
func buildTarotPrompt(card tarot.Card) string {
	return fmt.Sprintf(
		"Tarot card illustration titled '%s', ornate art-nouveau border with gold filigree, "+
			"mystical symbolism: %s, vertical card composition, single central figure, "+
			"painted fantasy style, high detail, %s",
		card.Name, card.Flavor, image.BasePromptSuffix)
}

// stableSeed derives a deterministic seed from a card id so re-runs of the deck
// reproduce the same art.
func stableSeed(id string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return int(h.Sum32() % 1_000_000)
}

// cmdTarotDeck generates the static fortune-teller card deck into web/static/cards/.
// Re-runs skip cards whose art already exists (use --force to regenerate them).
// Flags: --output=<dir> --model=<name> --deck=<path> --provider=<auto|fal|google> --force --dry-run
func cmdTarotDeck(args []string) error {
	outDir := filepath.Join("web", "static", "cards")
	modelName := "flux-2-pro"
	deckPath := tarot.DefaultDeckPath
	provider := "auto"
	dryRun := false
	force := false

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--output="):
			outDir = strings.TrimPrefix(arg, "--output=")
		case strings.HasPrefix(arg, "--model="):
			modelName = strings.TrimPrefix(arg, "--model=")
		case strings.HasPrefix(arg, "--deck="):
			deckPath = strings.TrimPrefix(arg, "--deck=")
		case strings.HasPrefix(arg, "--provider="):
			provider = strings.ToLower(strings.TrimPrefix(arg, "--provider="))
		case arg == "--force":
			force = true
		case arg == "--dry-run":
			dryRun = true
		default:
			return fmt.Errorf("argument inconnu: %s", arg)
		}
	}

	deck, err := tarot.LoadDeck(deckPath)
	if err != nil {
		return err
	}
	model := image.GetModel(modelName)

	fmt.Printf("Jeu de la Voyante: %d cartes, modèle=%s, fournisseur=%s, sortie=%s\n",
		len(deck.Cards), model.Short, provider, outDir)

	if dryRun {
		for _, card := range deck.Cards {
			fmt.Printf("\n[%s] %s -> %s\n  %s\n", card.ID, card.Name, card.Art, buildTarotPrompt(card))
		}
		fmt.Println("\n(dry-run: aucune image générée)")
		return nil
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("création du dossier de sortie: %w", err)
	}

	// A card's art already exists on disk?
	cardExists := func(card tarot.Card) bool {
		_, statErr := os.Stat(filepath.Join(outDir, card.Art))
		return statErr == nil
	}

	// Pre-scan: skip already-generated cards unless --force. If nothing is left
	// to do we return before touching any provider (no API key needed).
	pending := 0
	for _, card := range deck.Cards {
		if force || !cardExists(card) {
			pending++
		}
	}
	if pending == 0 {
		fmt.Printf("✓ Les %d cartes sont déjà présentes dans %s — rien à générer (--force pour régénérer).\n",
			len(deck.Cards), outDir)
		return nil
	}
	fmt.Printf("%d carte(s) à générer (%d déjà présente(s)).\n", pending, len(deck.Cards)-pending)

	// Select the image provider. fal.ai honors --model (flux-2-pro, schnell, ...);
	// Google Imagen ignores the model. "auto" keeps the GEMINI > FAL priority.
	gen, err := newImageProvider(provider, outDir)
	if err != nil {
		return err
	}

	var failures, skipped, generated int
	for i, card := range deck.Cards {
		dest := filepath.Join(outDir, card.Art)
		if !force && cardExists(card) {
			fmt.Printf("[%d/%d] %s (%s) — déjà présent, ignoré\n", i+1, len(deck.Cards), card.Name, card.ID)
			skipped++
			continue
		}
		fmt.Printf("[%d/%d] %s (%s)...\n", i+1, len(deck.Cards), card.Name, card.ID)
		result, err := gen.Generate(buildTarotPrompt(card),
			image.WithModelInstance(model),
			image.WithImageSize("portrait_4_3"),
			image.WithFilenamePrefix(card.ID+"_raw"),
			image.WithSeed(stableSeed(card.ID)),
			image.WithOutputFormat("png"),
		)
		if err != nil {
			fmt.Printf("  ⚠ échec: %v\n", err)
			failures++
			continue
		}
		// Normalize to the bare art filename referenced by the deck (e.g. le_soleil.png).
		if result.LocalPath != dest {
			if err := os.Rename(result.LocalPath, dest); err != nil {
				fmt.Printf("  ⚠ renommage échoué: %v\n", err)
				failures++
				continue
			}
		}
		generated++
		fmt.Printf("  ✓ %s\n", dest)
	}

	fmt.Printf("Bilan: %d générée(s), %d ignorée(s), %d échec(s).\n", generated, skipped, failures)
	if failures > 0 {
		return fmt.Errorf("%d carte(s) en échec", failures)
	}
	return nil
}

// voyanteAsset describes one UI visual for the fortune-teller wizard scene.
type voyanteAsset struct {
	ID     string
	File   string
	Size   string
	Prompt string
}

// voyanteAssetList is the set of "La Tente de la Voyante" UI visuals (warm
// candle-lit tent, burgundy + gold). Prompts ask for no text/no people where
// HTML overlays the content.
func voyanteAssetList() []voyanteAsset {
	s := image.BasePromptSuffix
	return []voyanteAsset{
		{
			ID: "tent_background", File: "tent_background.png", Size: "landscape_16_9",
			Prompt: "Interior of a mystical fortune-teller's tent at night, warm candlelight, deep burgundy velvet drapes, hanging brass lanterns, drifting incense smoke, ornate patterned rugs, golden tassels, dark atmospheric vignette in the corners suitable for overlaying text, no people, no text, cinematic depth, painted, " + s,
		},
		{
			ID: "header_banner", File: "header_banner.png", Size: "landscape_16_9",
			Prompt: "Ornate carved dark-wood and antique-gold signboard, art-nouveau filigree and scrollwork, blank empty smooth center panel, fortune-teller aesthetic, deep burgundy and gold, isolated decorative horizontal header element on plain dark background, no text, no letters, painted, " + s,
		},
		{
			ID: "voyante_portrait", File: "voyante_portrait.png", Size: "portrait_4_3",
			Prompt: "Portrait of a mysterious female fortune teller, hooded embroidered shawl, golden hoop earrings and jewelry, candlelit face half in shadow, faint glowing crystal ball, deep burgundy and gold tones, tarot cards on the table before her, enigmatic knowing expression, painted fantasy character portrait, " + s,
		},
		{
			ID: "velvet_table", File: "velvet_table.png", Size: "landscape_16_9",
			Prompt: "Top-down view of a dark burgundy velvet tablecloth with gold-embroidered border and hanging tassels, soft warm candlelight glow, empty cleared surface, rich fabric folds, seamless horizontal runner, painted texture, no people, no text, " + s,
		},
	}
}

// cmdVoyanteAssets generates the wizard scene UI visuals into web/static/voyante/.
// Re-runs skip already-present assets (use --force to regenerate).
// Flags: --output=<dir> --model=<name> --provider=<auto|fal|google> --force --dry-run
func cmdVoyanteAssets(args []string) error {
	outDir := filepath.Join("web", "static", "voyante")
	modelName := "flux-2-pro"
	provider := "auto"
	dryRun := false
	force := false

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--output="):
			outDir = strings.TrimPrefix(arg, "--output=")
		case strings.HasPrefix(arg, "--model="):
			modelName = strings.TrimPrefix(arg, "--model=")
		case strings.HasPrefix(arg, "--provider="):
			provider = strings.ToLower(strings.TrimPrefix(arg, "--provider="))
		case arg == "--force":
			force = true
		case arg == "--dry-run":
			dryRun = true
		default:
			return fmt.Errorf("argument inconnu: %s", arg)
		}
	}

	assets := voyanteAssetList()
	model := image.GetModel(modelName)

	fmt.Printf("Visuels de la Voyante: %d éléments, modèle=%s, fournisseur=%s, sortie=%s\n",
		len(assets), model.Short, provider, outDir)

	if dryRun {
		for _, a := range assets {
			fmt.Printf("\n[%s] -> %s (%s)\n  %s\n", a.ID, a.File, a.Size, a.Prompt)
		}
		fmt.Println("\n(dry-run: aucune image générée)")
		return nil
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("création du dossier de sortie: %w", err)
	}

	assetExists := func(a voyanteAsset) bool {
		_, statErr := os.Stat(filepath.Join(outDir, a.File))
		return statErr == nil
	}

	pending := 0
	for _, a := range assets {
		if force || !assetExists(a) {
			pending++
		}
	}
	if pending == 0 {
		fmt.Printf("✓ Les %d visuels sont déjà présents dans %s — rien à générer (--force pour régénérer).\n",
			len(assets), outDir)
		return nil
	}
	fmt.Printf("%d visuel(s) à générer (%d déjà présent(s)).\n", pending, len(assets)-pending)

	gen, err := newImageProvider(provider, outDir)
	if err != nil {
		return err
	}

	var failures, skipped, generated int
	for i, a := range assets {
		dest := filepath.Join(outDir, a.File)
		if !force && assetExists(a) {
			fmt.Printf("[%d/%d] %s — déjà présent, ignoré\n", i+1, len(assets), a.ID)
			skipped++
			continue
		}
		fmt.Printf("[%d/%d] %s...\n", i+1, len(assets), a.ID)
		result, err := gen.Generate(a.Prompt,
			image.WithModelInstance(model),
			image.WithImageSize(a.Size),
			image.WithFilenamePrefix(a.ID+"_raw"),
			image.WithSeed(stableSeed(a.ID)),
			image.WithOutputFormat("png"),
		)
		if err != nil {
			fmt.Printf("  ⚠ échec: %v\n", err)
			failures++
			continue
		}
		if result.LocalPath != dest {
			if err := os.Rename(result.LocalPath, dest); err != nil {
				fmt.Printf("  ⚠ renommage échoué: %v\n", err)
				failures++
				continue
			}
		}
		generated++
		fmt.Printf("  ✓ %s\n", dest)
	}

	fmt.Printf("Bilan: %d généré(s), %d ignoré(s), %d échec(s).\n", generated, skipped, failures)
	if failures > 0 {
		return fmt.Errorf("%d visuel(s) en échec", failures)
	}
	return nil
}

func cmdList(args []string) error {
	listType := "all"
	if len(args) > 0 {
		listType = strings.ToLower(args[0])
	}

	switch listType {
	case "styles", "style":
		fmt.Print("## Styles Disponibles\n\n")
		for _, style := range image.GetAvailableStyles() {
			fmt.Printf("- %s\n", style)
		}

	case "scenes", "scene":
		fmt.Print("## Types de Scènes Disponibles\n\n")
		for _, scene := range image.GetAvailableSceneTypes() {
			fmt.Printf("- %s\n", scene)
		}

	case "monsters", "monster":
		fmt.Print("## Types de Monstres Disponibles\n\n")
		for _, monster := range image.GetAvailableMonsterTypes() {
			fmt.Printf("- %s\n", monster)
		}

	case "items", "item":
		fmt.Print("## Types d'Objets Disponibles\n\n")
		for _, item := range image.GetAvailableItemTypes() {
			fmt.Printf("- %s\n", item)
		}

	case "locations", "location":
		fmt.Print("## Types de Lieux Disponibles\n\n")
		for _, loc := range image.GetAvailableLocationTypes() {
			fmt.Printf("- %s\n", loc)
		}

	case "sizes", "size":
		fmt.Print("## Tailles d'Image Disponibles\n\n")
		for _, size := range image.GetAvailableImageSizes() {
			fmt.Printf("- %s\n", size)
		}

	default:
		fmt.Print("## Styles Disponibles\n\n")
		for _, style := range image.GetAvailableStyles() {
			fmt.Printf("- %s\n", style)
		}

		fmt.Print("\n## Types de Scènes\n\n")
		for _, scene := range image.GetAvailableSceneTypes() {
			fmt.Printf("- %s\n", scene)
		}

		fmt.Print("\n## Types de Monstres\n\n")
		for _, monster := range image.GetAvailableMonsterTypes() {
			fmt.Printf("- %s\n", monster)
		}

		fmt.Print("\n## Types d'Objets\n\n")
		for _, item := range image.GetAvailableItemTypes() {
			fmt.Printf("- %s\n", item)
		}

		fmt.Print("\n## Types de Lieux\n\n")
		for _, loc := range image.GetAvailableLocationTypes() {
			fmt.Printf("- %s\n", loc)
		}

		fmt.Print("\n## Tailles d'Image\n\n")
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
	if opts["model"] != "" {
		imgOpts = append(imgOpts, image.WithModel(opts["model"]))
	}

	return imgOpts
}

// printGenerationInfo displays generation parameters (model, style, size).
func printGenerationInfo(opts map[string]string, defaultSize string) {
	model := opts["model"]
	if model == "" {
		model = "schnell"
	}

	style := opts["style"]
	if style == "" {
		style = "(défaut)"
	}

	size := opts["size"]
	if size == "" {
		size = defaultSize
	}

	fmt.Printf("→ Modèle: %s | Style: %s | Taille: %s\n", model, style, size)
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

// filterCharactersByMention returns only characters whose names are mentioned in the journal entry.
// It checks both the description (English) and description_fr (French) fields.
// If "the party" is mentioned but no specific characters, returns all characters in random order.
func filterCharactersByMention(entry adventure.JournalEntry, characters []*character.Character) []*character.Character {
	if len(characters) == 0 {
		return nil
	}

	// Get the text to search (prefer English description, fallback to French, then content)
	searchText := strings.ToLower(entry.Description)
	if searchText == "" {
		searchText = strings.ToLower(entry.DescriptionFr)
	}
	if searchText == "" {
		searchText = strings.ToLower(entry.Content)
	}

	// Split text into words once for all checks
	words := strings.FieldsFunc(searchText, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	})

	var mentioned []*character.Character
	for _, char := range characters {
		// Check if character name appears as a whole word (case-insensitive)
		// This prevents "Aldric" from matching "Valdric"
		charName := strings.ToLower(char.Name)
		for _, word := range words {
			if strings.ToLower(word) == charName {
				mentioned = append(mentioned, char)
				break
			}
		}
	}

	// If no specific characters found but "party"/"groupe" is mentioned, include all in random order
	// Check if words "party" or "groupe" appear (handles "the party", "the victorious party", etc.)
	if len(mentioned) == 0 {
		for _, word := range words {
			if word == "party" || word == "groupe" {
				mentioned = make([]*character.Character, len(characters))
				copy(mentioned, characters)
				// Shuffle to randomize order so Aldric isn't always first
				rand.Shuffle(len(mentioned), func(i, j int) {
					mentioned[i], mentioned[j] = mentioned[j], mentioned[i]
				})
				break
			}
		}
	}

	return mentioned
}

// cmdJournal generates images for journal entries in parallel.
func cmdJournal(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("nom de l'aventure requis")
	}

	// Parse adventure name and options
	var advName string
	var optArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			optArgs = append(optArgs, arg)
		} else if advName == "" {
			advName = arg
		}
	}

	opts := parseOptions(optArgs)

	// Build adventure path
	advPath := filepath.Join(dataDir, "adventures", strings.ToLower(strings.ReplaceAll(advName, " ", "-")))

	// Load adventure
	adv, err := adventure.Load(advPath)
	if err != nil {
		return fmt.Errorf("chargement de l'aventure '%s': %w", advName, err)
	}

	// Load adventure characters for context
	characters, err := adv.GetCharacters()
	if err != nil {
		fmt.Printf("⚠️  Attention : Impossible de charger les personnages : %v\n", err)
		characters = nil
	}

	// Display party info
	if len(characters) > 0 {
		fmt.Printf("Groupe : ")
		for i, c := range characters {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(c.Name)
			if c.Appearance != nil && c.Appearance.ReferenceImage != "" {
				fmt.Print(" ✓") // Has reference image
			}
		}
		fmt.Println()
	}

	// Load journal
	journal, err := adv.LoadJournal()
	if err != nil {
		return fmt.Errorf("chargement du journal: %w", err)
	}

	// Determine which types to illustrate
	typesToIllustrate := image.IllustratableTypes()
	if opts["types"] != "" {
		typesToIllustrate = strings.Split(opts["types"], ",")
	}

	// Parse start-id if provided
	startID := 0
	if opts["start-id"] != "" {
		fmt.Sscanf(opts["start-id"], "%d", &startID)
	}

	// Filter entries
	var entriesToIllustrate []adventure.JournalEntry
	for _, entry := range journal.Entries {
		// Skip entries before start-id
		if startID > 0 && entry.ID < startID {
			continue
		}

		// Skip "Session démarrée" entries (only keep session end summaries)
		if entry.Type == "session" && strings.Contains(entry.Content, "démarrée") {
			continue
		}

		// Check if type is in the list
		for _, t := range typesToIllustrate {
			if entry.Type == t {
				entriesToIllustrate = append(entriesToIllustrate, entry)
				break
			}
		}
	}

	// Apply max limit
	maxImages := len(entriesToIllustrate)
	if opts["max"] != "" {
		var m int
		fmt.Sscanf(opts["max"], "%d", &m)
		if m > 0 && m < maxImages {
			maxImages = m
			entriesToIllustrate = entriesToIllustrate[:maxImages]
		}
	}

	if len(entriesToIllustrate) == 0 {
		fmt.Println("Aucune entrée à illustrer trouvée dans le journal.")
		return nil
	}

	fmt.Printf("## Illustration du journal : %s\n\n", adv.Name)
	fmt.Printf("Entrées à illustrer : %d\n", len(entriesToIllustrate))

	// Build prompts for all entries
	type promptJob struct {
		entry  adventure.JournalEntry
		prompt *image.JournalEntryPrompt
	}

	var jobs []promptJob
	for _, entry := range entriesToIllustrate {
		// Filter characters to only those mentioned in the entry description
		mentionedChars := filterCharactersByMention(entry, characters)
		prompt := image.BuildJournalEntryPromptWithCharacters(entry, mentionedChars)
		if prompt != nil {
			jobs = append(jobs, promptJob{entry: entry, prompt: prompt})
		}
	}

	// Select model based on active provider
	var model image.Model
	if os.Getenv("GEMINI_API_KEY") != "" {
		// Google Imagen active — skip FAL model validation
		model = image.Model{Short: "imagen-4.0", ID: "imagen-4.0-generate-001"}
		fmt.Printf("→ Provider: Google Imagen (imagen-4.0-generate-001)\n")
	} else {
		// FAL.ai model selection
		journalModels := image.JournalModels()
		modelName := opts["model"]
		if modelName == "" {
			modelName = "flux-2-pro" // Default: FLUX.2 Pro for high quality journal illustrations
		}

		var ok bool
		model, ok = journalModels[modelName]
		if !ok {
			availableModels := []string{}
			for name := range journalModels {
				availableModels = append(availableModels, name)
			}
			return fmt.Errorf("modèle '%s' non disponible pour le journal. Modèles disponibles: %v", modelName, availableModels)
		}
	}

	// Dry run mode - just print prompts
	if opts["dry-run"] == "true" {
		fmt.Printf("\n### Mode dry-run : %d prompts générés (modèle: %s)\n\n", len(jobs), model.Short)
		for i, job := range jobs {
			filename := fmt.Sprintf("journal_%03d_%s_%s.png", job.entry.ID, job.entry.Type, model.Short)
			fmt.Printf("**[%d] %s (ID: %d)**\n", i+1, job.entry.Type, job.entry.ID)
			fmt.Printf("  Fichier : %s\n", filename)
			fmt.Printf("  Contenu : %s\n", job.entry.Content)
			fmt.Printf("  Style : %s\n", job.prompt.Style)
			fmt.Printf("  Taille : %s\n", job.prompt.ImageSize)
			fmt.Printf("  Prompt : %s\n\n", job.prompt.Prompt)
		}
		return nil
	}

	// Note: Images will be saved in session-specific directories
	// images/session-N/ for each entry based on its SessionID
	advImagesDir := filepath.Join(advPath, "images")

	// Determine parallelism level
	parallelism := 4
	if opts["parallel"] != "" {
		fmt.Sscanf(opts["parallel"], "%d", &parallelism)
	}
	if parallelism < 1 {
		parallelism = 1
	}
	if parallelism > 8 {
		parallelism = 8
	}

	fmt.Printf("→ Modèle: %s | Parallélisme: %d\n", model.Short, parallelism)
	fmt.Printf("Génération de %d images...\n", len(jobs))

	// Display first prompt as example
	if len(jobs) > 0 {
		firstJob := jobs[0]
		fmt.Printf("\n### Premier prompt (exemple)\n")
		fmt.Printf("  Entrée ID: %d\n", firstJob.prompt.EntryID)
		fmt.Printf("  Type: %s\n", firstJob.entry.Type)
		fmt.Printf("  Style: %s\n", firstJob.prompt.Style)
		fmt.Printf("  Taille: %s\n", firstJob.prompt.ImageSize)
		fmt.Printf("  Prompt: %s\n\n", firstJob.prompt.Prompt)
	}

	// Channel for job distribution
	jobsChan := make(chan promptJob, len(jobs))
	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)

	// Results
	type result struct {
		entryID int
		path    string
		err     error
	}
	resultsChan := make(chan result, len(jobs))

	// Worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func(workerID int, m image.Model) {
			defer wg.Done()
			for job := range jobsChan {
				// Determine session directory for this entry
				sessionDir := filepath.Join(advImagesDir, fmt.Sprintf("session-%d", job.entry.SessionID))

				// Create session directory if needed
				if err := os.MkdirAll(sessionDir, 0755); err != nil {
					resultsChan <- result{entryID: job.prompt.EntryID, err: fmt.Errorf("création répertoire session: %w", err)}
					continue
				}

				// Create generator for this session directory
				gen, err := image.NewGeneratorAuto(sessionDir)
				if err != nil {
					resultsChan <- result{entryID: job.prompt.EntryID, err: fmt.Errorf("création générateur: %w", err)}
					continue
				}

				// Build filename prefix: journal_XXX_type (e.g., journal_008_combat)
				// Model name is appended automatically by the generator
				filenamePrefix := fmt.Sprintf("journal_%03d_%s", job.prompt.EntryID, job.entry.Type)

				var img *image.GeneratedImage

				// Generate image with standard model (nano-banana by default)
				imgOpts := []image.Option{
					image.WithImageSize(job.prompt.ImageSize),
					image.WithFilenamePrefix(filenamePrefix),
					image.WithModel(m.Short),
				}

				// Set deterministic seed for seedream model
				if m.Short == "seedream" {
					imgOpts = append(imgOpts, image.WithSeed(1024))
				}

				img, err = gen.Generate(job.prompt.Prompt, imgOpts...)

				if err != nil {
					resultsChan <- result{entryID: job.prompt.EntryID, err: err}
					continue
				}

				resultsChan <- result{entryID: job.prompt.EntryID, path: img.LocalPath}
			}
		}(i, model)
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var successCount, errorCount int
	for res := range resultsChan {
		if res.err != nil {
			fmt.Printf("  [ERREUR] Entrée %d : %v\n", res.entryID, res.err)
			errorCount++
		} else {
			fmt.Printf("  [OK] Entrée %d : %s\n", res.entryID, filepath.Base(res.path))
			successCount++
		}
	}

	fmt.Printf("\n## Résumé\n")
	fmt.Printf("  Images générées : %d\n", successCount)
	fmt.Printf("  Erreurs : %d\n", errorCount)
	fmt.Printf("  Répertoire de base : %s\n", advImagesDir)
	fmt.Printf("  Images organisées par session : images/session-N/\n")

	return nil
}
