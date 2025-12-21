package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"dungeons/internal/monster"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	bestiary, err := monster.NewBestiary(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "show", "get":
		err = cmdShow(bestiary, args)
	case "search", "find":
		err = cmdSearch(bestiary, args)
	case "list":
		err = cmdList(bestiary, args)
	case "encounter", "enc":
		err = cmdEncounter(bestiary, args)
	case "roll":
		err = cmdRoll(bestiary, args)
	case "types":
		err = cmdTypes(bestiary)
	case "tables":
		err = cmdTables(bestiary)
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
	fmt.Println(`SkillsWeaver - Monster Manual - Bestiaire BFRPG

UTILISATION:
  sw-monster <commande> [arguments]

COMMANDES:
  show <id>                    Afficher la fiche complète d'un monstre
  search <terme>               Rechercher des monstres par nom ou type
  list [--type=<type>]         Lister tous les monstres (optionnel: par type)
  encounter <table>            Générer une rencontre aléatoire
  encounter --level=<N>        Générer une rencontre pour un niveau de groupe
  roll <id> [--count=N]        Créer N instances avec PV aléatoires
  types                        Lister les types de monstres
  tables                       Lister les tables de rencontres
  help                         Afficher cette aide

OPTIONS:
  --format=<md|json|short>     Format de sortie (défaut: md)
  --type=<type>                Filtrer par type (undead, humanoid, etc.)
  --level=<N>                  Niveau du groupe pour les rencontres
  --count=<N>                  Nombre de monstres à générer

TYPES DE MONSTRES:
  animal, dragon, giant, humanoid, monstrosity, ooze, undead, vermin

TABLES DE RENCONTRES:
  dungeon_level_1    Niveau 1 (faible)
  dungeon_level_2    Niveau 2 (modéré)
  dungeon_level_3    Niveau 3 (élevé)
  dungeon_level_4    Niveau 4+ (très élevé)
  forest             Forêt
  undead_crypt       Crypte/Cimetière

EXEMPLES:
  sw-monster show goblin                      # Fiche du gobelin
  sw-monster search dragon                    # Tous les dragons
  sw-monster list --type=undead               # Tous les morts-vivants
  sw-monster encounter dungeon_level_1        # Rencontre niveau 1
  sw-monster encounter --level=3              # Rencontre pour groupe niveau 3
  sw-monster roll orc --count=4               # 4 orcs avec PV aléatoires`)
}

func cmdShow(b *monster.Bestiary, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("ID du monstre requis")
	}

	id := args[0]
	opts := parseOptions(args[1:])

	m, err := b.GetMonster(id)
	if err != nil {
		return err
	}

	format := opts["format"]
	if format == "" {
		format = "md"
	}

	switch format {
	case "json":
		jsonStr, err := m.ToJSON()
		if err != nil {
			return err
		}
		fmt.Println(jsonStr)
	case "short":
		fmt.Println(m.ToShortDescription())
	default:
		fmt.Print(m.ToMarkdown())
	}

	return nil
}

func cmdSearch(b *monster.Bestiary, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("terme de recherche requis")
	}

	query := args[0]
	opts := parseOptions(args[1:])

	results := b.SearchMonsters(query)
	if len(results) == 0 {
		fmt.Println("Aucun monstre trouvé.")
		return nil
	}

	format := opts["format"]
	if format == "" {
		format = "short"
	}

	fmt.Printf("## %d monstre(s) trouvé(s)\n\n", len(results))
	for _, m := range results {
		switch format {
		case "md":
			fmt.Print(m.ToMarkdown())
			fmt.Println("---")
		default:
			fmt.Println(m.ToShortDescription())
		}
	}

	return nil
}

func cmdList(b *monster.Bestiary, args []string) error {
	opts := parseOptions(args)

	var monsters []*monster.Monster
	if opts["type"] != "" {
		monsters = b.ListByType(opts["type"])
		if len(monsters) == 0 {
			return fmt.Errorf("type inconnu: %s", opts["type"])
		}
		fmt.Printf("## Monstres de type '%s'\n\n", opts["type"])
	} else {
		monsters = b.ListAll()
		fmt.Print("## Tous les Monstres\n\n")
	}

	format := opts["format"]
	if format == "" {
		format = "short"
	}

	for _, m := range monsters {
		switch format {
		case "md":
			fmt.Print(m.ToMarkdown())
			fmt.Println("---")
		default:
			fmt.Println(m.ToShortDescription())
		}
	}

	return nil
}

func cmdEncounter(b *monster.Bestiary, args []string) error {
	opts := parseOptions(args)

	var encounter *monster.EncounterResult
	var err error

	if opts["level"] != "" {
		level, parseErr := strconv.Atoi(opts["level"])
		if parseErr != nil {
			return fmt.Errorf("niveau invalide: %s", opts["level"])
		}
		encounter, err = b.GenerateEncounterByLevel(level)
	} else if len(args) > 0 && !strings.HasPrefix(args[0], "--") {
		tableName := args[0]
		encounter, err = b.GenerateEncounter(tableName)
	} else {
		// Default to dungeon level 1
		encounter, err = b.GenerateEncounter("dungeon_level_1")
	}

	if err != nil {
		return err
	}

	format := opts["format"]
	if format == "" {
		format = "md"
	}

	switch format {
	case "json":
		jsonStr, err := encounter.ToJSON()
		if err != nil {
			return err
		}
		fmt.Println(jsonStr)
	default:
		fmt.Print(encounter.ToMarkdown())
	}

	return nil
}

func cmdRoll(b *monster.Bestiary, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("ID du monstre requis")
	}

	id := args[0]
	opts := parseOptions(args[1:])

	m, err := b.GetMonster(id)
	if err != nil {
		return err
	}

	count := 1
	if opts["count"] != "" {
		count, err = strconv.Atoi(opts["count"])
		if err != nil || count < 1 {
			count = 1
		}
	}

	fmt.Printf("## %s x%d\n\n", m.NameFR, count)
	fmt.Printf("CA %d | XP %d chacun\n\n", m.ArmorClass, m.XP)

	totalXP := 0
	fmt.Println("| # | PV | Statut |")
	fmt.Println("|---|-----|--------|")
	for i := 1; i <= count; i++ {
		inst := b.CreateInstance(m)
		fmt.Printf("| %d | %d | En vie |\n", i, inst.HitPoints)
		totalXP += m.XP
	}

	fmt.Printf("\n**XP Total** : %d\n", totalXP)

	// Show attacks
	fmt.Println("\n### Attaques")
	for _, atk := range m.Attacks {
		fmt.Printf("- **%s** : +%d, %s", atk.NameFR, atk.Bonus, atk.Damage)
		if atk.Special != "" {
			fmt.Printf(" (%s)", atk.Special)
		}
		fmt.Println()
	}

	// Show special abilities
	if len(m.Special) > 0 {
		fmt.Println("\n### Capacités Spéciales")
		for _, s := range m.Special {
			fmt.Printf("- %s\n", s)
		}
	}

	return nil
}

func cmdTypes(b *monster.Bestiary) error {
	types := b.GetTypes()

	fmt.Print("## Types de Monstres\n\n")
	for _, t := range types {
		monsters := b.ListByType(t)
		fmt.Printf("- **%s** (%d monstres)\n", t, len(monsters))
	}

	return nil
}

func cmdTables(b *monster.Bestiary) error {
	tables := b.GetEncounterTables()

	fmt.Print("## Tables de Rencontres\n\n")
	for _, t := range tables {
		fmt.Printf("- %s\n", t)
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
