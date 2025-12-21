package main

import (
	"fmt"
	"os"
	"strings"

	"dungeons/internal/npc"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	gen, err := npc.NewGenerator(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "generate", "gen":
		err = cmdGenerate(gen, args)
	case "quick":
		err = cmdQuick(gen, args)
	case "list":
		err = cmdList(gen, args)
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
	fmt.Println(`SkillsWeaver - NPC Generator - Générateur de PNJ pour BFRPG

UTILISATION:
  sw-npc <commande> [arguments]

COMMANDES:
  generate [options]           Générer un PNJ complet avec description
  quick [options]              Générer un PNJ avec description courte
  list [races|occupations|attitudes]  Lister les options disponibles
  help                         Afficher cette aide

OPTIONS:
  --race=<race>               Race (human, dwarf, elf, halfling)
  --gender=<m|f>              Sexe (m=masculin, f=féminin)
  --occupation=<type>         Type d'occupation (commoner, skilled, authority, etc.)
  --attitude=<type>           Attitude (positive, neutral, negative)
  --format=<md|json|short>    Format de sortie (défaut: md)

TYPES D'OCCUPATION:
  commoner    - Gens du peuple (fermier, boulanger, serveur...)
  skilled     - Artisans qualifiés (marchand, apothicaire, musicien...)
  authority   - Figures d'autorité (garde, noble, magistrat...)
  underworld  - Monde criminel (voleur, espion, contrebandier...)
  religious   - Religieux (prêtre, moine, pèlerin...)
  adventurer  - Aventuriers (chasseur de primes, explorateur...)

EXEMPLES:
  sw-npc generate                              # PNJ aléatoire complet
  sw-npc generate --race=dwarf --gender=m      # Nain masculin
  sw-npc generate --occupation=authority       # Figure d'autorité
  sw-npc generate --attitude=hostile           # PNJ hostile
  sw-npc quick --race=elf                      # Description courte d'elfe
  sw-npc generate --format=json                # Sortie JSON`)
}

func cmdGenerate(gen *npc.Generator, args []string) error {
	opts := parseOptions(args)

	var genOpts []npc.Option
	if opts["race"] != "" {
		genOpts = append(genOpts, npc.WithRace(opts["race"]))
	}
	if opts["gender"] != "" {
		genOpts = append(genOpts, npc.WithGender(opts["gender"]))
	}
	if opts["occupation"] != "" {
		genOpts = append(genOpts, npc.WithOccupationType(opts["occupation"]))
	}
	if opts["attitude"] != "" {
		genOpts = append(genOpts, npc.WithAttitude(opts["attitude"]))
	}

	n, err := gen.Generate(genOpts...)
	if err != nil {
		return err
	}

	format := opts["format"]
	if format == "" {
		format = "md"
	}

	switch format {
	case "json":
		jsonStr, err := n.ToJSON()
		if err != nil {
			return err
		}
		fmt.Println(jsonStr)
	case "short":
		fmt.Println(n.ToShortDescription())
	default:
		fmt.Print(n.ToMarkdown())
	}

	return nil
}

func cmdQuick(gen *npc.Generator, args []string) error {
	opts := parseOptions(args)

	var genOpts []npc.Option
	if opts["race"] != "" {
		genOpts = append(genOpts, npc.WithRace(opts["race"]))
	}
	if opts["gender"] != "" {
		genOpts = append(genOpts, npc.WithGender(opts["gender"]))
	}
	if opts["occupation"] != "" {
		genOpts = append(genOpts, npc.WithOccupationType(opts["occupation"]))
	}
	if opts["attitude"] != "" {
		genOpts = append(genOpts, npc.WithAttitude(opts["attitude"]))
	}

	count := 1
	if opts["count"] != "" {
		fmt.Sscanf(opts["count"], "%d", &count)
	}

	for i := 0; i < count; i++ {
		n, err := gen.Generate(genOpts...)
		if err != nil {
			return err
		}
		fmt.Println(n.ToShortDescription())
	}

	return nil
}

func cmdList(gen *npc.Generator, args []string) error {
	listType := "all"
	if len(args) > 0 {
		listType = strings.ToLower(args[0])
	}

	switch listType {
	case "races", "race":
		fmt.Println("## Races Disponibles")
		fmt.Println()
		racesFr := map[string]string{
			"human":    "Humain",
			"dwarf":    "Nain",
			"elf":      "Elfe",
			"halfling": "Halfelin",
		}
		for _, race := range gen.GetAvailableRaces() {
			fmt.Printf("- %s (%s)\n", race, racesFr[race])
		}

	case "occupations", "occupation", "occ":
		fmt.Println("## Types d'Occupation Disponibles")
		fmt.Println()
		occFr := map[string]string{
			"commoner":   "Gens du peuple",
			"skilled":    "Artisans qualifiés",
			"authority":  "Figures d'autorité",
			"underworld": "Monde criminel",
			"religious":  "Religieux",
			"adventurer": "Aventuriers",
		}
		for _, occ := range gen.GetAvailableOccupationTypes() {
			fmt.Printf("- %s (%s)\n", occ, occFr[occ])
		}

	case "attitudes", "attitude", "att":
		fmt.Println("## Attitudes Disponibles")
		fmt.Println()
		attFr := map[string]string{
			"positive": "Amical",
			"neutral":  "Neutre",
			"negative": "Hostile",
		}
		for _, att := range gen.GetAvailableAttitudes() {
			fmt.Printf("- %s (%s)\n", att, attFr[att])
		}

	default:
		fmt.Println("## Races Disponibles")
		fmt.Println()
		racesFr := map[string]string{
			"human":    "Humain",
			"dwarf":    "Nain",
			"elf":      "Elfe",
			"halfling": "Halfelin",
		}
		for _, race := range gen.GetAvailableRaces() {
			fmt.Printf("- %s (%s)\n", race, racesFr[race])
		}

		fmt.Println()
		fmt.Println("## Types d'Occupation Disponibles")
		fmt.Println()
		occFr := map[string]string{
			"commoner":   "Gens du peuple",
			"skilled":    "Artisans qualifiés",
			"authority":  "Figures d'autorité",
			"underworld": "Monde criminel",
			"religious":  "Religieux",
			"adventurer": "Aventuriers",
		}
		for _, occ := range gen.GetAvailableOccupationTypes() {
			fmt.Printf("- %s (%s)\n", occ, occFr[occ])
		}

		fmt.Println()
		fmt.Println("## Attitudes Disponibles")
		fmt.Println()
		attFr := map[string]string{
			"positive": "Amical",
			"neutral":  "Neutre",
			"negative": "Hostile",
		}
		for _, att := range gen.GetAvailableAttitudes() {
			fmt.Printf("- %s (%s)\n", att, attFr[att])
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
