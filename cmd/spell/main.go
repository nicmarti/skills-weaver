package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"dungeons/internal/character"
	"dungeons/internal/data"
	"dungeons/internal/spell"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	manager, err := spell.NewManagerFromDataDir(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "list", "liste":
		err = cmdList(manager, args)
	case "show", "afficher":
		err = cmdShow(manager, args)
	case "search", "chercher":
		err = cmdSearch(manager, args)
	case "cantrips":
		err = cmdCantrips(manager, args)
	case "schools", "ecoles":
		err = cmdSchools(manager, args)
	case "concentration":
		err = cmdConcentration(manager, args)
	case "rituals", "rituels":
		err = cmdRituals(manager, args)
	case "slots":
		err = cmdSlots(args)
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
	fmt.Println(`SkillsWeaver - Spell Reference - Grimoire des Sorts D&D 5e

UTILISATION:
  sw-spell <commande> [arguments]

COMMANDES:
  list                             Lister tous les sorts
  list --class=<classe>            Sorts d'une classe
  list --class=<classe> --level=N  Sorts d'une classe et niveau
  list --level=N                   Sorts d'un niveau
  list --school=<école>            Sorts d'une école de magie
  show <id>                        Afficher un sort en détail
  search <terme>                   Rechercher par nom (FR ou EN)
  cantrips <classe>                Lister les cantrips d'une classe
  schools                          Lister les écoles de magie
  concentration                    Lister les sorts de concentration
  rituals                          Lister les sorts rituels
  slots <classe> --level=N         Afficher les slots de sorts d'une classe
  help                             Afficher cette aide

OPTIONS:
  --format=<md|json|short>         Format de sortie (défaut: short pour list, md pour show)
  --class=<classe>                 Filtrer par classe
  --level=<0-9>                    Filtrer par niveau (0 = cantrips)
  --school=<école>                 Filtrer par école de magie

CLASSES:
  wizard, magicien                 Magicien (full caster)
  sorcerer, ensorceleur            Ensorceleur (full caster)
  cleric, clerc                    Clerc (full caster)
  druid, druide                    Druide (full caster)
  bard, barde                      Barde (full caster)
  warlock, occultiste              Occultiste (pact caster)
  paladin                          Paladin (half caster)
  ranger, rôdeur                   Rôdeur (half caster)
  fighter, guerrier                Guerrier (1/3 caster - Eldritch Knight)
  rogue, roublard                  Roublard (1/3 caster - Arcane Trickster)

ÉCOLES DE MAGIE:
  abjuration                       Abjuration (protection)
  conjuration                      Invocation (création/téléportation)
  divination                       Divination (connaissance)
  enchantment                      Enchantement (contrôle mental)
  evocation                        Évocation (énergie/dégâts)
  illusion                         Illusion (tromperie)
  necromancy                       Nécromancie (mort/non-mort)
  transmutation                    Transmutation (transformation)

EXEMPLES:
  sw-spell list                                    # Tous les sorts
  sw-spell list --class=wizard                     # Sorts de magicien
  sw-spell list --class=wizard --level=1           # Magicien niveau 1
  sw-spell list --school=evocation                 # Sorts d'évocation
  sw-spell cantrips wizard                         # Cantrips de magicien
  sw-spell show projectile_magique                 # Détails Projectile magique
  sw-spell search "feu"                            # Recherche "feu"
  sw-spell concentration                           # Sorts de concentration
  sw-spell rituals                                 # Sorts rituels
  sw-spell slots wizard --level=5                  # Slots magicien niveau 5`)
}

func cmdList(m *spell.Manager, args []string) error {
	opts := parseOptions(args)
	class := opts["class"]
	levelStr := opts["level"]
	school := opts["school"]
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	var spells []*data.Spell5e
	var title string

	// Determine which spells to list
	if school != "" {
		spells = m.ListBySchool(school)
		title = fmt.Sprintf("## Sorts de l'école %s", spell.GetSchoolLabel(school))
	} else if class != "" && levelStr != "" {
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			return fmt.Errorf("niveau invalide: %s", levelStr)
		}
		spells = m.ListByClassAndLevel(class, level)
		levelName := fmt.Sprintf("Niveau %d", level)
		if level == 0 {
			levelName = "Cantrips"
		}
		title = fmt.Sprintf("## Sorts de %s - %s", spell.GetClassLabel(class), levelName)
	} else if class != "" {
		spells = m.ListByClass(class)
		title = fmt.Sprintf("## Sorts de %s", spell.GetClassLabel(class))
	} else if levelStr != "" {
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			return fmt.Errorf("niveau invalide: %s", levelStr)
		}
		spells = m.ListByLevel(level)
		levelName := fmt.Sprintf("Niveau %d", level)
		if level == 0 {
			levelName = "Cantrips"
		}
		title = fmt.Sprintf("## Sorts de %s", levelName)
	} else {
		spells = m.ListAllSpells()
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
			fmt.Print(spell.ToMarkdown(s))
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
				levelName := fmt.Sprintf("Niveau %d", s.Level)
				if s.Level == 0 {
					levelName = "Cantrips"
				}
				fmt.Printf("### %s\n\n", levelName)
				currentLevel = s.Level
			}
			fmt.Println(spell.ToShortDescription(s))
		}
	}

	return nil
}

func cmdShow(m *spell.Manager, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("ID du sort requis")
	}

	id := args[0]
	opts := parseOptions(args[1:])
	format := opts["format"]
	if format == "" {
		format = "md"
	}

	s, err := m.GetSpell(id)
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
		fmt.Println(spell.ToShortDescription(s))
	default:
		fmt.Print(spell.ToMarkdown(s))
	}

	return nil
}

func cmdSearch(m *spell.Manager, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("terme de recherche requis")
	}

	query := args[0]
	opts := parseOptions(args[1:])
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	results := m.SearchSpells(query)
	if len(results) == 0 {
		fmt.Println("Aucun sort trouvé.")
		return nil
	}

	fmt.Printf("## %d sort(s) trouvé(s) pour '%s'\n\n", len(results), query)

	switch format {
	case "md":
		for _, s := range results {
			fmt.Print(spell.ToMarkdown(s))
			fmt.Println("---")
		}
	case "json":
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		currentLevel := -1
		for _, s := range results {
			if s.Level != currentLevel {
				if currentLevel != -1 {
					fmt.Println()
				}
				levelName := fmt.Sprintf("Niveau %d", s.Level)
				if s.Level == 0 {
					levelName = "Cantrips"
				}
				fmt.Printf("### %s\n\n", levelName)
				currentLevel = s.Level
			}
			fmt.Println(spell.ToShortDescription(s))
		}
	}

	return nil
}

func cmdCantrips(m *spell.Manager, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("classe requise")
	}

	class := args[0]
	opts := parseOptions(args[1:])
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	spells := m.ListCantrips(class)
	if len(spells) == 0 {
		fmt.Printf("Aucun cantrip trouvé pour %s.\n", spell.GetClassLabel(class))
		return nil
	}

	fmt.Printf("## Cantrips de %s (%d sorts)\n\n", spell.GetClassLabel(class), len(spells))

	switch format {
	case "json":
		data, err := json.MarshalIndent(spells, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "md":
		for _, s := range spells {
			fmt.Print(spell.ToMarkdown(s))
			fmt.Println("---")
		}
	default:
		for _, s := range spells {
			fmt.Println(spell.ToShortDescription(s))
		}
	}

	return nil
}

func cmdSchools(m *spell.Manager, args []string) error {
	opts := parseOptions(args)
	format := opts["format"]

	schools := m.GetAvailableSchools()

	fmt.Println("## Écoles de Magie D&D 5e\n")

	for _, school := range schools {
		spells := m.ListBySchool(school)
		fmt.Printf("**%s** : %d sorts\n", spell.GetSchoolLabel(school), len(spells))

		if format == "detail" {
			// Show first 3 examples
			count := 3
			if len(spells) < count {
				count = len(spells)
			}
			for i := 0; i < count; i++ {
				fmt.Printf("  - %s\n", spells[i].Name)
			}
			if len(spells) > 3 {
				fmt.Println("  ...")
			}
		}
	}

	return nil
}

func cmdConcentration(m *spell.Manager, args []string) error {
	opts := parseOptions(args)
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	spells := m.ListConcentration()
	if len(spells) == 0 {
		fmt.Println("Aucun sort de concentration trouvé.")
		return nil
	}

	fmt.Printf("## Sorts de Concentration (%d sorts)\n\n", len(spells))
	fmt.Println("Note: Seul 1 sort de concentration peut être actif à la fois.\n")

	switch format {
	case "json":
		data, err := json.MarshalIndent(spells, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "md":
		for _, s := range spells {
			fmt.Print(spell.ToMarkdown(s))
			fmt.Println("---")
		}
	default:
		currentLevel := -1
		for _, s := range spells {
			if s.Level != currentLevel {
				if currentLevel != -1 {
					fmt.Println()
				}
				levelName := fmt.Sprintf("Niveau %d", s.Level)
				if s.Level == 0 {
					levelName = "Cantrips"
				}
				fmt.Printf("### %s\n\n", levelName)
				currentLevel = s.Level
			}
			fmt.Println(spell.ToShortDescription(s))
		}
	}

	return nil
}

func cmdRituals(m *spell.Manager, args []string) error {
	opts := parseOptions(args)
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	spells := m.ListRituals()
	if len(spells) == 0 {
		fmt.Println("Aucun sort rituel trouvé.")
		return nil
	}

	fmt.Printf("## Sorts Rituels (%d sorts)\n\n", len(spells))
	fmt.Println("Note: Sorts rituels prennent +10 minutes mais ne consomment pas de slot.\n")

	switch format {
	case "json":
		data, err := json.MarshalIndent(spells, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "md":
		for _, s := range spells {
			fmt.Print(spell.ToMarkdown(s))
			fmt.Println("---")
		}
	default:
		currentLevel := -1
		for _, s := range spells {
			if s.Level != currentLevel {
				if currentLevel != -1 {
					fmt.Println()
				}
				levelName := fmt.Sprintf("Niveau %d", s.Level)
				if s.Level == 0 {
					levelName = "Cantrips"
				}
				fmt.Printf("### %s\n\n", levelName)
				currentLevel = s.Level
			}
			fmt.Println(spell.ToShortDescription(s))
		}
	}

	return nil
}

func cmdSlots(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("classe requise")
	}

	class := args[0]
	opts := parseOptions(args[1:])
	levelStr := opts["level"]

	if levelStr == "" {
		return fmt.Errorf("niveau requis (--level=N)")
	}

	level, err := strconv.Atoi(levelStr)
	if err != nil {
		return fmt.Errorf("niveau invalide: %s", levelStr)
	}

	if level < 1 || level > 20 {
		return fmt.Errorf("niveau doit être entre 1 et 20")
	}

	// Normalize class ID
	classID := strings.ToLower(class)
	aliases := map[string]string{
		"clerc":       "cleric",
		"magicien":    "wizard",
		"mage":        "wizard",
		"ensorceleur": "sorcerer",
		"occultiste":  "warlock",
		"rôdeur":      "ranger",
		"rodeur":      "ranger",
		"guerrier":    "fighter",
		"roublard":    "rogue",
		"druide":      "druid",
		"barde":       "bard",
	}
	if normalized, ok := aliases[classID]; ok {
		classID = normalized
	}

	slots := character.GetSpellSlots(classID, level)
	if slots == nil || len(slots) == 0 {
		fmt.Printf("%s niveau %d n'a pas encore de slots de sorts.\n",
			spell.GetClassLabel(classID), level)
		return nil
	}

	fmt.Printf("## Slots de Sorts - %s Niveau %d\n\n", spell.GetClassLabel(classID), level)

	fmt.Println("| Niveau de Sort | Nombre de Slots |")
	fmt.Println("|----------------|-----------------|")
	for lvl := 1; lvl <= 9; lvl++ {
		if count, ok := slots[lvl]; ok && count > 0 {
			fmt.Printf("| %d              | %d               |\n", lvl, count)
		}
	}

	cantrips := character.GetCantripsKnown(classID, level)
	if cantrips > 0 {
		fmt.Printf("\n**Cantrips connus** : %d (illimités)\n", cantrips)
	}

	// Special note for Warlock
	if classID == "warlock" {
		fmt.Println("\n**Note Occultiste** : Tous les slots sont du même niveau (Pact Magic).")
		fmt.Println("Les slots se restaurent au repos court.")
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
