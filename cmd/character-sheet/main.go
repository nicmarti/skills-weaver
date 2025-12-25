package main

import (
	"dungeons/internal/charactersheet"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		if err := cmdGenerate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
			os.Exit(1)
		}
	case "regenerate":
		if err := cmdRegenerate(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
			os.Exit(1)
		}
	case "bio":
		if err := cmdBio(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
			os.Exit(1)
		}
	case "templates":
		cmdTemplates()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Commande inconnue: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func cmdGenerate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("nom de personnage requis")
	}

	characterName := args[0]
	opts := parseOptions(args[1:])

	// Create sheet generator
	gen, err := charactersheet.NewSheetGenerator(dataDir)
	if err != nil {
		return err
	}

	// Build sheet options
	sheetOpts := charactersheet.SheetOptions{
		CharacterName:    characterName,
		Adventure:        opts["adventure"],
		IncludeBiography: opts["with-biography"] == "true",
		RefreshBio:       opts["refresh-bio"] == "true",
		IncludePortrait:  opts["include-portrait"] != "false", // Default true
		GenerateBanner:   opts["generate-banner"] == "true",
		GenerateIcons:    opts["generate-icons"] == "true",
	}

	// Set output path
	if opts["output"] != "" {
		sheetOpts.OutputPath = opts["output"]
	} else {
		filename := strings.ToLower(strings.ReplaceAll(characterName, " ", "-")) + ".html"
		sheetOpts.OutputPath = filepath.Join(dataDir, "characters", filename)
	}

	// Generate sheet
	fmt.Printf("Génération de la fiche pour %s...\n", characterName)
	sheet, err := gen.Generate(sheetOpts)
	if err != nil {
		return err
	}

	// Render HTML
	html, err := gen.RenderHTML(sheet)
	if err != nil {
		return err
	}

	// Save HTML
	if err := gen.Save(html, sheetOpts.OutputPath); err != nil {
		return err
	}

	fmt.Printf("✓ Fiche générée: %s\n", sheetOpts.OutputPath)

	// Open in browser if requested
	if opts["open"] == "true" {
		if err := openInBrowser(sheetOpts.OutputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open in browser: %v\n", err)
		} else {
			fmt.Println("✓ Ouvert dans le navigateur")
		}
	}

	return nil
}

func cmdRegenerate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("nom de personnage requis")
	}

	characterName := args[0]
	opts := parseOptions(args[1:])

	// Regenerate is same as generate, but we keep existing options
	fmt.Printf("Régénération de la fiche pour %s...\n", characterName)

	// Create sheet generator
	gen, err := charactersheet.NewSheetGenerator(dataDir)
	if err != nil {
		return err
	}

	// Build sheet options (use cached bio if exists)
	sheetOpts := charactersheet.SheetOptions{
		CharacterName:    characterName,
		Adventure:        opts["adventure"],
		IncludeBiography: true, // Always include bio on regenerate
		RefreshBio:       opts["refresh-bio"] == "true",
		IncludePortrait:  true,
	}

	// Set output path
	filename := strings.ToLower(strings.ReplaceAll(characterName, " ", "-")) + ".html"
	sheetOpts.OutputPath = filepath.Join(dataDir, "characters", filename)

	// Generate sheet
	sheet, err := gen.Generate(sheetOpts)
	if err != nil {
		return err
	}

	// Render HTML
	html, err := gen.RenderHTML(sheet)
	if err != nil {
		return err
	}

	// Save HTML
	if err := gen.Save(html, sheetOpts.OutputPath); err != nil {
		return err
	}

	fmt.Printf("✓ Fiche régénérée: %s\n", sheetOpts.OutputPath)

	return nil
}

func cmdBio(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("nom de personnage requis")
	}

	characterName := args[0]
	opts := parseOptions(args[1:])

	// Load character
	charPath := filepath.Join(dataDir, "characters", strings.ToLower(strings.ReplaceAll(characterName, " ", "-"))+".json")
	c, err := charactersheet.LoadCharacterForBio(charPath)
	if err != nil {
		return err
	}

	// Generate biography
	bioGen := charactersheet.NewBiographyGenerator()
	bio, err := bioGen.Generate(c, opts["adventure"])
	if err != nil {
		return err
	}

	// Display biography
	fmt.Printf("\n=== Biographie de %s ===\n\n", characterName)
	fmt.Printf("Origine: %s\n\n", bio.Origin)
	fmt.Printf("Passé: %s\n\n", bio.Background)
	fmt.Printf("Motivation: %s\n\n", bio.Motivation)
	fmt.Printf("Personnalité: %s\n\n", bio.Personality)

	if len(bio.Bonds) > 0 {
		fmt.Println("Relations:")
		for _, bond := range bio.Bonds {
			fmt.Printf("  - %s (%s): %s\n", bond.Name, bond.Type, bond.Description)
		}
		fmt.Println()
	}

	if len(bio.Secrets) > 0 {
		fmt.Println("Secrets:")
		for _, secret := range bio.Secrets {
			fmt.Printf("  - %s\n", secret)
		}
		fmt.Println()
	}

	// Ask if user wants to save
	fmt.Print("Sauvegarder cette biographie? (o/n): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "o" || strings.ToLower(response) == "oui" {
		if err := bio.Save(filepath.Join(dataDir, "characters")); err != nil {
			return fmt.Errorf("failed to save biography: %w", err)
		}
		fmt.Println("✓ Biographie sauvegardée")
	}

	return nil
}

func cmdTemplates() {
	fmt.Println("Templates disponibles:")
	fmt.Println()
	fmt.Println("  dark-fantasy  - Style Baldur's Gate avec fond sombre et accents dorés (par défaut)")
	fmt.Println("                  Optimisé pour affichage écran, effets visuels riches")
	fmt.Println()
	fmt.Println("Utilisation:")
	fmt.Println("  sw-character-sheet generate \"Nom\" --template=dark-fantasy")
	fmt.Println()
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

func openInBrowser(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absPath)
	case "linux":
		cmd = exec.Command("xdg-open", absPath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", absPath)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func printUsage() {
	fmt.Println(`sw-character-sheet - Générateur de Fiches de Personnage HTML

UTILISATION:
  sw-character-sheet <commande> [arguments] [options]

COMMANDES:
  generate <nom>        Générer une nouvelle fiche de personnage
  regenerate <nom>      Régénérer une fiche existante
  bio <nom>             Générer et prévisualiser la biographie uniquement
  templates             Lister les templates disponibles
  help                  Afficher cette aide

OPTIONS:
  --adventure=<nom>         Inclure données d'aventure (inventaire, or partagé)
  --with-biography          Générer la biographie du personnage
  --refresh-bio             Forcer la régénération de la biographie
  --include-portrait        Inclure l'image de référence (défaut: oui)
  --generate-banner         Générer bannière décorative de classe
  --generate-icons          Générer icônes d'équipement
  --output=<chemin>         Chemin de sortie personnalisé
  --open                    Ouvrir dans le navigateur après génération

EXEMPLES:
  # Fiche simple
  sw-character-sheet generate "Aldric"

  # Fiche complète avec biographie et données d'aventure
  sw-character-sheet generate "Aldric" --adventure="la-crypte-des-ombres" --with-biography

  # Régénérer après montée de niveau
  sw-character-sheet regenerate "Aldric" --adventure="la-crypte-des-ombres"

  # Prévisualiser la biographie
  sw-character-sheet bio "Lyra" --adventure="la-crypte-des-ombres"

  # Générer et ouvrir dans le navigateur
  sw-character-sheet generate "Thorin" --with-biography --open

  # Forcer régénération de la biographie
  sw-character-sheet regenerate "Gareth" --refresh-bio

SORTIE:
  Les fiches HTML sont sauvegardées dans data/characters/<nom>.html
  Les biographies sont cachées dans data/characters/<nom>_bio.json

STYLE:
  Dark Fantasy (Baldur's Gate) - Fond sombre, accents dorés, effets de glow
  `)
}
