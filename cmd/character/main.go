// Command character provides a CLI for creating and managing BFRPG characters.
//
// Usage:
//
//	sw-character create "Name" --race=human --class=fighter
//	sw-character create "Name" --race=elf --class=magic-user --method=classic
//	sw-character list
//	sw-character show "Name"
//	sw-character delete "Name"
//	sw-character export "Name" --format=json
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dungeons/internal/character"
	"dungeons/internal/data"
)

const (
	defaultDataDir      = "data"
	defaultCharacterDir = "data/characters"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		handleCreate(os.Args[2:])
	case "list":
		handleList()
	case "show":
		handleShow(os.Args[2:])
	case "delete":
		handleDelete(os.Args[2:])
	case "export":
		handleExport(os.Args[2:])
	case "appearance":
		handleAppearance(os.Args[2:])
	case "set-reference":
		handleSetReference(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCreate(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: character name required")
		fmt.Fprintln(os.Stderr, "Usage: character create \"Name\" --race=human --class=fighter")
		os.Exit(1)
	}

	name := args[0]
	race := getFlag(args, "--race", "human")
	class := getFlag(args, "--class", "fighter")
	method := getFlag(args, "--method", "standard")
	maxHP := hasFlag(args, "--max-hp") // Use max HP at level 1 (variant rule)

	// Load game data
	gd, err := data.Load(defaultDataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading game data: %v\n", err)
		os.Exit(1)
	}

	// Create character
	c := character.New(name, race, class)

	// Validate race/class combination
	if err := c.Validate(gd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Generate abilities
	var genMethod character.GenerationMethod
	if method == "classic" {
		genMethod = character.MethodClassic
	} else {
		genMethod = character.MethodStandard
	}

	fmt.Printf("## Création de %s\n\n", name)

	// Roll abilities
	fmt.Println("### Génération des caractéristiques")
	fmt.Println()
	results := c.GenerateAbilities(genMethod)

	statNames := []string{"Force", "Intelligence", "Sagesse", "Dextérité", "Constitution", "Charisme"}
	fmt.Println("| Caractéristique | Jets | Total |")
	fmt.Println("|-----------------|------|-------|")
	for i, result := range results {
		rolls := formatRolls(result.Rolls, result.KeptIndices)
		fmt.Printf("| %-15s | %s | **%2d** |\n", statNames[i], rolls, result.Total)
	}
	fmt.Println()

	// Apply racial modifiers
	if err := c.ApplyRacialModifiers(gd); err != nil {
		fmt.Fprintf(os.Stderr, "Error applying racial modifiers: %v\n", err)
		os.Exit(1)
	}

	raceData, _ := gd.GetRace(race)
	if len(raceData.AbilityModifiers) > 0 {
		fmt.Printf("### Modificateurs raciaux (%s)\n\n", raceData.Name)
		for ability, mod := range raceData.AbilityModifiers {
			if mod > 0 {
				fmt.Printf("- %s: +%d\n", ability, mod)
			} else {
				fmt.Printf("- %s: %d\n", ability, mod)
			}
		}
		fmt.Println()
	}

	// Calculate modifiers
	c.CalculateModifiers()

	// Roll hit points
	if err := c.RollHitPoints(gd, maxHP); err != nil {
		fmt.Fprintf(os.Stderr, "Error rolling hit points: %v\n", err)
		os.Exit(1)
	}

	classData, _ := gd.GetClass(class)
	if maxHP {
		fmt.Printf("### Points de vie (niveau 1, %s max)\n\n", classData.HitDie)
		fmt.Printf("PV = %d (dé max) + %d (CON) = **%d**\n\n", classData.HitDieSides, c.Modifiers.Constitution, c.HitPoints)
	} else {
		dieRoll := c.HitPoints - c.Modifiers.Constitution
		if dieRoll < 1 {
			dieRoll = 1 // Clamp for display when HP is minimum 1
		}
		fmt.Printf("### Points de vie (niveau 1, %s lancé)\n\n", classData.HitDie)
		fmt.Printf("PV = %d (%s) + %d (CON) = **%d**\n\n", dieRoll, classData.HitDie, c.Modifiers.Constitution, c.HitPoints)
	}

	// Roll starting gold
	if err := c.RollStartingGold(gd); err != nil {
		fmt.Fprintf(os.Stderr, "Error rolling starting gold: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("### Or de départ\n\n")
	fmt.Printf("**%d po**\n\n", c.Gold)

	// Calculate AC (base only, no armor yet)
	c.CalculateArmorClass(gd)

	// Save character
	if err := c.Save(defaultCharacterDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving character: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("---")
	fmt.Println()
	fmt.Printf("Personnage sauvegardé dans `%s/%s.json`\n\n", defaultCharacterDir, strings.ToLower(strings.ReplaceAll(name, " ", "_")))

	// Print full character sheet
	fmt.Println("---")
	fmt.Println()
	fmt.Println(c.ToMarkdown(gd))
}

func handleList() {
	characters, err := character.ListCharacters(defaultCharacterDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing characters: %v\n", err)
		os.Exit(1)
	}

	if len(characters) == 0 {
		fmt.Println("Aucun personnage trouvé.")
		return
	}

	// Load game data for names
	gd, _ := data.Load(defaultDataDir)

	fmt.Println("## Personnages")
	fmt.Println()
	fmt.Println("| Nom | Race | Classe | Niveau | PV |")
	fmt.Println("|-----|------|--------|--------|-----|")

	for _, c := range characters {
		raceName := c.Race
		className := c.Class

		if gd != nil {
			if race, ok := gd.GetRace(c.Race); ok {
				raceName = race.Name
			}
			if class, ok := gd.GetClass(c.Class); ok {
				className = class.Name
			}
		}

		fmt.Printf("| %s | %s | %s | %d | %d/%d |\n",
			c.Name, raceName, className, c.Level, c.HitPoints, c.MaxHitPoints)
	}
}

func handleShow(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: character name required")
		fmt.Fprintln(os.Stderr, "Usage: character show \"Name\"")
		os.Exit(1)
	}

	name := args[0]
	filename := strings.ToLower(strings.ReplaceAll(name, " ", "_")) + ".json"
	path := filepath.Join(defaultCharacterDir, filename)

	c, err := character.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading character: %v\n", err)
		os.Exit(1)
	}

	gd, _ := data.Load(defaultDataDir)
	fmt.Println(c.ToMarkdown(gd))
}

func handleDelete(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: character name required")
		fmt.Fprintln(os.Stderr, "Usage: character delete \"Name\"")
		os.Exit(1)
	}

	name := args[0]

	if err := character.Delete(defaultCharacterDir, name); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting character: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Personnage '%s' supprimé.\n", name)
}

func handleExport(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: character name required")
		fmt.Fprintln(os.Stderr, "Usage: character export \"Name\" [--format=json|md]")
		os.Exit(1)
	}

	name := args[0]
	format := getFlag(args, "--format", "md")

	filename := strings.ToLower(strings.ReplaceAll(name, " ", "_")) + ".json"
	path := filepath.Join(defaultCharacterDir, filename)

	c, err := character.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading character: %v\n", err)
		os.Exit(1)
	}

	switch format {
	case "json":
		json, err := c.ToJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting to JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(json)
	case "md", "markdown":
		gd, _ := data.Load(defaultDataDir)
		fmt.Println(c.ToMarkdown(gd))
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s (use json or md)\n", format)
		os.Exit(1)
	}
}

// translateToEnglish translates common French appearance terms to English
func translateToEnglish(value string) string {
	// Start with the original value
	result := value

	// Common French to English translations for character appearance
	translations := map[string]string{
		// Build / Corpulence
		"mince":      "slender",
		"maigre":     "thin",
		"svelte":     "slender",
		"musclé":     "muscular",
		"musculeux":  "muscular",
		"trapu":      "stocky",
		"robuste":    "stocky",
		"moyen":      "average",
		"moyenne":    "average",
		"enrobé":     "heavy",
		"corpulent":  "heavy",

		// Height / Taille (specific forms to avoid partial matches)
		"très grand":  "very tall",
		"très grande": "very tall",
		"très petit":  "very short",
		"très petite": "very short",
		"grande":      "tall",
		"petite":      "short",

		// Hair color / Couleur cheveux
		"noir":       "black",
		"noire":      "black",
		"noirs":      "black",
		"brun":       "brown",
		"brune":      "brown",
		"bruns":      "brown",
		"châtain":    "brown",
		"blond":      "blond",
		"blonde":     "blond",
		"argenté":    "silver",
		"argentée":   "silver",
		"roux":       "red",
		"rousse":     "red",
		"gris":       "gray",
		"grise":      "gray",
		"blanc":      "white",
		"blanche":    "white",

		// Hair style / Style cheveux (multi-word first)
		"longs et ondulés":   "long and wavy",
		"longs et bouclés":   "long and curly",
		"courts et bouclés":  "short and curly",
		"rasés":              "shaved",
		"rasée":              "shaved",
		"chauve":             "bald",
		"courts":             "short",
		"longs":              "long",
		"ondulés":            "wavy",
		"bouclé":             "curly",
		"bouclés":            "curly",
		"frisé":              "curly",
		"ondulé":             "wavy",
		"tressé":             "braided",
		"attaché":            "tied back",
		"hirsute":            "unkempt",
		"presque":            "nearly",
		" et ":               " and ",

		// Eye color / Couleur yeux (multi-word phrases first to handle French word order)
		"bleu clair":     "light blue",
		"bleu foncé":     "dark blue",
		"vert clair":     "light green",
		"vert foncé":     "dark green",
		"marron clair":   "light brown",
		"marron foncé":   "dark brown",
		"brun clair":     "light brown",
		"brun foncé":     "dark brown",
		"gris clair":     "light gray",
		"gris foncé":     "dark gray",
		"marron":         "brown",
		"marrons":        "brown",
		"bleu":           "blue",
		"bleue":          "blue",
		"bleus":          "blue",
		"vert":           "green",
		"verte":          "green",
		"verts":          "green",
		"noisette":       "hazel",

		// Skin tone / Teint
		"pâle":       "pale",
		"bronzé":     "tanned",
		"bronzée":    "tanned",
		"basané":     "olive",
		"basanée":    "olive",
		"sombre":     "dark",
		"légèrement": "lightly",

		// Facial features / Traits faciaux
		"barbu":      "bearded",
		"imberbe":    "clean-shaven",
		"rasé":       "clean-shaven",
		"moustache":  "mustached",
		"cicatrice":  "scarred",
		"cicatrisé":  "scarred",
		"ridé":       "wrinkled",

		// Distinctive features / Traits distinctifs
		"oreilles pointues":  "pointed ears",
		"cicatrice à l'œil":  "eye scar",
		"tatouage":           "tattoo",
		"tatoué":             "tattooed",
		"borgne":             "one-eyed",
		"cache-œil":          "eye patch",
		"balafre":            "scar",
		"balafrée":           "scarred",
		"percé":              "pierced",
		"percée":             "pierced",

		// Armor / Armure
		"armure de plates":   "plate armor",
		"cotte de mailles":   "chainmail",
		"armure de cuir":     "leather armor",
		"cuir clouté":        "studded leather",
		"armure légère":      "light armor",
		"armure lourde":      "heavy armor",
		"robe":               "robe",

		// Weapons / Armes
		"épée longue":        "longsword",
		"épée courte":        "shortsword",
		"épée à deux mains":  "greatsword",
		"dague":              "dagger",
		"hache":              "axe",
		"masse":              "mace",
		"bâton":              "staff",
		"arc":                "bow",
		"arbalète":           "crossbow",
		"lance":              "spear",

		// Accessories / Accessoires
		"bouclier":           "shield",
		"symbole sacré":      "holy symbol",
		"livre de sorts":     "spell book",
		"grimoire":           "spell book",
		"sacoche":            "pouch",
		"besace":             "satchel",
		"cape":               "cloak",
		"capuche":            "hood",
	}

	// Sort translations by length (longest first) to handle multi-word phrases before single words
	type translation struct {
		fr string
		en string
	}
	var sortedTranslations []translation
	for fr, en := range translations {
		sortedTranslations = append(sortedTranslations, translation{fr, en})
	}
	// Sort by length descending
	for i := 0; i < len(sortedTranslations); i++ {
		for j := i + 1; j < len(sortedTranslations); j++ {
			if len(sortedTranslations[j].fr) > len(sortedTranslations[i].fr) {
				sortedTranslations[i], sortedTranslations[j] = sortedTranslations[j], sortedTranslations[i]
			}
		}
	}

	// Apply translations in order (longest phrases first)
	for _, t := range sortedTranslations {
		result = strings.ReplaceAll(result, t.fr, t.en)
	}

	return result
}

func handleAppearance(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: character name required")
		fmt.Fprintln(os.Stderr, "Usage: character appearance \"Name\" [--field=value ...]")
		os.Exit(1)
	}

	name := args[0]
	filename := strings.ToLower(strings.ReplaceAll(name, " ", "_")) + ".json"
	path := filepath.Join(defaultCharacterDir, filename)

	// Load character
	c, err := character.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading character: %v\n", err)
		os.Exit(1)
	}

	// Initialize appearance if needed
	if c.Appearance == nil {
		c.Appearance = &character.CharacterAppearance{}
	}

	// Parse and apply appearance fields
	updated := false
	for _, arg := range args[1:] {
		if !strings.HasPrefix(arg, "--") {
			continue
		}

		parts := strings.SplitN(strings.TrimPrefix(arg, "--"), "=", 2)
		if len(parts) != 2 {
			continue
		}

		field, value := parts[0], parts[1]
		updated = true

		// Translate French to English for all appearance fields
		translatedValue := translateToEnglish(value)

		switch field {
		case "age":
			fmt.Sscanf(value, "%d", &c.Appearance.Age)
		case "gender":
			c.Appearance.Gender = value // Keep as-is: "male", "female", "non-binary"
		case "build":
			c.Appearance.Build = translatedValue
		case "height":
			c.Appearance.Height = translatedValue
		case "hair-color":
			c.Appearance.HairColor = translatedValue
		case "hair-style":
			c.Appearance.HairStyle = translatedValue
		case "eye-color":
			c.Appearance.EyeColor = translatedValue
		case "skin-tone":
			c.Appearance.SkinTone = translatedValue
		case "facial-feature":
			c.Appearance.FacialFeature = translatedValue
		case "distinctive":
			c.Appearance.DistinctiveFeature = translatedValue
		case "armor":
			c.Appearance.ArmorDescription = translatedValue
		case "weapon":
			c.Appearance.WeaponDescription = translatedValue
		case "accessories":
			c.Appearance.Accessories = translatedValue
		default:
			fmt.Fprintf(os.Stderr, "Unknown field: %s\n", field)
		}
	}

	if !updated {
		// No fields specified, just display current appearance
		fmt.Printf("## Apparence de %s\n\n", c.Name)
		if c.Appearance == nil || (c.Appearance.Age == 0 && c.Appearance.Build == "" && c.Appearance.Height == "") {
			fmt.Println("Aucune apparence définie.")
		} else {
			fmt.Println(c.GetVisualDescription())
			fmt.Println()
			if c.Appearance.ReferenceImage != "" {
				fmt.Printf("Image de référence : %s\n", c.Appearance.ReferenceImage)
			}
		}
		return
	}

	// Save character
	if err := c.Save(defaultCharacterDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving character: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Apparence mise à jour pour %s\n\n", c.Name)
	fmt.Println(c.GetVisualDescription())
	fmt.Println()
	fmt.Printf("Pour la génération d'images : %s\n", c.GetImagePromptSnippet())
}

func handleSetReference(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: character name and image path required")
		fmt.Fprintln(os.Stderr, "Usage: character set-reference \"Name\" <image-path>")
		os.Exit(1)
	}

	name := args[0]
	imagePath := args[1]

	filename := strings.ToLower(strings.ReplaceAll(name, " ", "_")) + ".json"
	charPath := filepath.Join(defaultCharacterDir, filename)

	// Verify image exists
	if _, err := os.Stat(imagePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: image not found: %v\n", err)
		os.Exit(1)
	}

	// Load character
	c, err := character.Load(charPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading character: %v\n", err)
		os.Exit(1)
	}

	// Initialize appearance if needed
	if c.Appearance == nil {
		c.Appearance = &character.CharacterAppearance{}
	}

	// Copy image to characters directory
	charSlug := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	destFilename := charSlug + "_reference" + filepath.Ext(imagePath)
	destPath := filepath.Join(defaultCharacterDir, destFilename)

	input, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading image: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(destPath, input, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving reference image: %v\n", err)
		os.Exit(1)
	}

	// Update character with reference image filename (relative path)
	c.Appearance.ReferenceImage = destFilename

	// Save character
	if err := c.Save(defaultCharacterDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving character: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Image de référence définie pour %s\n", c.Name)
	fmt.Printf("  Source : %s\n", imagePath)
	fmt.Printf("  Copié vers : %s\n", destPath)
	fmt.Println()
	fmt.Println("Cette image sera utilisée avec FLUX PuLID pour la cohérence des personnages.")
}

func getFlag(args []string, flag, defaultValue string) string {
	prefix := flag + "="
	for _, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return defaultValue
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func formatRolls(rolls []int, keptIndices []int) string {
	parts := make([]string, len(rolls))
	for i, roll := range rolls {
		isKept := false
		for _, idx := range keptIndices {
			if idx == i {
				isKept = true
				break
			}
		}
		if len(keptIndices) < len(rolls) && !isKept {
			parts[i] = fmt.Sprintf("~~%d~~", roll)
		} else {
			parts[i] = fmt.Sprintf("%d", roll)
		}
	}
	return strings.Join(parts, ", ")
}

func printUsage() {
	fmt.Println(`SkillsWeaver - Character Generator - Générateur de personnages BFRPG

USAGE:
    sw-character <command> [arguments]

COMMANDES:
    create "Nom" [options]       Crée un nouveau personnage
    list                         Liste tous les personnages
    show "Nom"                   Affiche la fiche d'un personnage
    delete "Nom"                 Supprime un personnage
    export "Nom" [--format]      Exporte un personnage
    appearance "Nom" [options]   Définit/affiche l'apparence du personnage
    set-reference "Nom" <image>  Définit l'image de référence pour PuLID
    help                         Affiche cette aide

OPTIONS POUR CREATE:
    --race=<race>       Race du personnage (human, elf, dwarf, halfling)
    --class=<class>     Classe du personnage (fighter, cleric, magic-user, thief)
    --method=<method>   Méthode de génération (standard=4d6kh3, classic=3d6)
    --max-hp            PV max au niveau 1 (variante pour survie)

OPTIONS POUR EXPORT:
    --format=<format>   Format d'export (json, md)

OPTIONS POUR APPEARANCE:
    --age=<n>               Âge du personnage
    --gender=<value>        Genre (male, female, non-binary)
    --build=<value>         Corpulence (slender, stocky, muscular, average)
    --height=<value>        Taille (tall, average, short)
    --hair-color=<value>    Couleur des cheveux
    --hair-style=<value>    Style de coiffure
    --eye-color=<value>     Couleur des yeux
    --skin-tone=<value>     Teint de peau
    --facial-feature=<val>  Trait facial (bearded, clean-shaven, scarred)
    --distinctive=<value>   Trait distinctif (battle scar, tattoo, eye patch)
    --armor=<value>         Description de l'armure (plate armor, leather vest)
    --weapon=<value>        Description de l'arme (longsword, staff with crystal)
    --accessories=<value>   Accessoires (shield, holy symbol, spell book)

RACES DISPONIBLES:
    human     - Humain (toutes classes, niveau illimité)
    elf       - Elfe (+1 DEX, -1 CON) : Guerrier 6, Magicien 9, Voleur
    dwarf     - Nain (+1 CON, -1 CHA) : Guerrier 7, Clerc 6, Voleur
    halfling  - Halfelin (+1 DEX, -1 FOR) : Guerrier 4, Voleur

CLASSES DISPONIBLES:
    fighter     - Guerrier (d8 PV, toutes armes/armures)
    cleric      - Clerc (d6 PV, sorts divins, armes contondantes)
    magic-user  - Magicien (d4 PV, sorts arcaniques)
    thief       - Voleur (d4 PV, compétences spéciales)

EXEMPLES:
    sw-character create "Aldric" --race=human --class=fighter
    sw-character create "Lyra" --race=elf --class=magic-user --method=classic
    sw-character create "Gorim" --race=dwarf --class=fighter --max-hp
    sw-character list
    sw-character show "Aldric"
    sw-character export "Aldric" --format=json
    sw-character appearance "Aldric" --age=34 --build=muscular --armor="plate armor"
    sw-character set-reference "Aldric" path/to/portrait.png

NOTES SUR LES PV:
    Par défaut, les PV au niveau 1 sont lancés aléatoirement (règle BFRPG standard).
    Avec --max-hp, le personnage reçoit le maximum du dé de vie (variante populaire
    pour améliorer la survie des personnages de bas niveau).`)
}
