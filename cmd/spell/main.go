package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"dungeons/internal/spell"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	spellBook, err := spell.NewSpellBook(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "list", "liste":
		err = cmdList(spellBook, args)
	case "show", "afficher":
		err = cmdShow(spellBook, args)
	case "search", "chercher":
		err = cmdSearch(spellBook, args)
	case "reversible":
		err = cmdReversible(spellBook, args)
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
	fmt.Println(`SkillsWeaver - Spell Reference - Grimoire des Sorts BFRPG

UTILISATION:
  sw-spell <commande> [arguments]

COMMANDES:
  list                             Lister tous les sorts
  list --class=<classe>            Sorts d'une classe
  list --class=<classe> --level=N  Sorts d'une classe et niveau
  list --level=N                   Sorts d'un niveau
  show <id>                        Afficher un sort en détail
  search <terme>                   Rechercher par nom (FR ou EN)
  reversible                       Lister les sorts réversibles
  help                             Afficher cette aide

OPTIONS:
  --format=<md|json|short>         Format de sortie (défaut: short pour list, md pour show)
  --class=<cleric|magic-user>      Filtrer par classe
  --level=<1|2>                    Filtrer par niveau

CLASSES:
  cleric, clerc                    Clerc (sorts divins)
  magic-user, magicien, mage       Magicien (sorts arcaniques)

EXEMPLES:
  sw-spell list                          # Tous les sorts
  sw-spell list --class=cleric           # Sorts de clerc
  sw-spell list --class=magic-user --level=1  # Magicien niveau 1
  sw-spell show magic_missile            # Détails Projectile magique
  sw-spell search lumière                # Recherche "lumière"
  sw-spell reversible                    # Sorts avec forme inversée`)
}

func cmdList(sb *spell.SpellBook, args []string) error {
	opts := parseOptions(args)
	class := opts["class"]
	levelStr := opts["level"]
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	var spells []*spell.Spell
	var title string

	// Determine which spells to list
	if class != "" && levelStr != "" {
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			return fmt.Errorf("niveau invalide: %s", levelStr)
		}
		spells = sb.ListByClassAndLevel(class, level)
		title = fmt.Sprintf("## Sorts de %s Niveau %d", spell.GetClassLabel(class), level)
	} else if class != "" {
		spells = sb.ListByClass(class)
		title = fmt.Sprintf("## Sorts de %s", spell.GetClassLabel(class))
	} else if levelStr != "" {
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			return fmt.Errorf("niveau invalide: %s", levelStr)
		}
		spells = sb.ListByLevel(level)
		title = fmt.Sprintf("## Sorts de Niveau %d", level)
	} else {
		spells = sb.ListAllSpells()
		title = "## Tous les Sorts"
	}

	if len(spells) == 0 {
		fmt.Println("Aucun sort trouvé.")
		return nil
	}

	fmt.Printf("%s (%d sorts)\n\n", title, len(spells))

	switch format {
	case "json":
		data, err := json.MarshalIndent(spells, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "md":
		for _, s := range spells {
			fmt.Print(s.ToMarkdown())
			fmt.Println("---")
		}
	default:
		// Group by level for better readability
		currentLevel := -1
		for _, s := range spells {
			if s.Level != currentLevel {
				if currentLevel != -1 {
					fmt.Println()
				}
				fmt.Printf("### Niveau %d\n\n", s.Level)
				currentLevel = s.Level
			}
			fmt.Println(s.ToShortDescription())
		}
	}

	return nil
}

func cmdShow(sb *spell.SpellBook, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("ID du sort requis")
	}

	id := args[0]
	opts := parseOptions(args[1:])
	format := opts["format"]
	if format == "" {
		format = "md"
	}

	s, err := sb.GetSpell(id)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "short":
		fmt.Println(s.ToShortDescription())
	default:
		fmt.Print(s.ToMarkdown())
	}

	return nil
}

func cmdSearch(sb *spell.SpellBook, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("terme de recherche requis")
	}

	query := args[0]
	opts := parseOptions(args[1:])
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	results := sb.SearchSpells(query)
	if len(results) == 0 {
		fmt.Println("Aucun sort trouvé.")
		return nil
	}

	fmt.Printf("## %d sort(s) trouvé(s) pour '%s'\n\n", len(results), query)

	switch format {
	case "md":
		for _, s := range results {
			fmt.Print(s.ToMarkdown())
			fmt.Println("---")
		}
	case "json":
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		for _, s := range results {
			fmt.Println(s.ToShortDescription())
		}
	}

	return nil
}

func cmdReversible(sb *spell.SpellBook, args []string) error {
	opts := parseOptions(args)
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	spells := sb.ListReversible()
	if len(spells) == 0 {
		fmt.Println("Aucun sort réversible trouvé.")
		return nil
	}

	fmt.Printf("## Sorts Réversibles (%d sorts)\n\n", len(spells))

	switch format {
	case "json":
		data, err := json.MarshalIndent(spells, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "md":
		for _, s := range spells {
			fmt.Print(s.ToMarkdown())
			fmt.Println("---")
		}
	default:
		fmt.Println("| Sort | Forme inversée | Niveau | Type |")
		fmt.Println("|------|----------------|--------|------|")
		for _, s := range spells {
			fmt.Printf("| %s | %s | %d | %s |\n",
				s.NameFR, s.ReverseNameFR, s.Level, spell.GetTypeLabel(s.Type))
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
