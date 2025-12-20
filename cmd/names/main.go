package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"dungeons/internal/names"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	gen, err := names.NewGenerator(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "generate", "gen":
		err = cmdGenerate(gen, args)
	case "npc":
		err = cmdNPC(gen, args)
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
	fmt.Println(`Name Generator - Générateur de Noms Fantasy BFRPG

UTILISATION:
  names <commande> [arguments]

COMMANDES:
  generate <race> [options]    Générer un nom pour une race
  npc <type>                   Générer un nom de PNJ
  list [races|npc]             Lister les options disponibles
  help                         Afficher cette aide

OPTIONS GENERATE:
  --gender=<m|f>               Sexe (m=masculin, f=féminin, omis=aléatoire)
  --count=<N>                  Nombre de noms à générer (défaut: 1)
  --first-only                 Générer uniquement le prénom

RACES DISPONIBLES:
  dwarf (nain), elf (elfe), halfling (halfelin), human (humain)

TYPES DE PNJ:
  innkeeper (tavernier), merchant (marchand), guard (garde),
  noble, wizard (mage), villain (méchant)

EXEMPLES:
  names generate dwarf                    # Nom de nain aléatoire
  names generate elf --gender=f           # Nom d'elfe féminin
  names generate human --count=5          # 5 noms humains
  names generate halfling --first-only    # Prénom de halfelin
  names npc innkeeper                     # Nom de tavernier
  names npc villain                       # Nom de méchant`)
}

func cmdGenerate(gen *names.Generator, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: names generate <race> [--gender=m|f] [--count=N] [--first-only]")
	}

	race := args[0]
	gender := ""
	count := 1
	firstOnly := false

	// Parse options
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "--gender=") {
			gender = strings.TrimPrefix(arg, "--gender=")
		} else if strings.HasPrefix(arg, "--count=") {
			n, err := strconv.Atoi(strings.TrimPrefix(arg, "--count="))
			if err == nil && n > 0 {
				count = n
			}
		} else if arg == "--first-only" {
			firstOnly = true
		}
	}

	if count == 1 {
		var name string
		var err error

		if firstOnly {
			name, err = gen.GenerateFirstName(race, gender)
		} else {
			name, err = gen.GenerateName(race, gender)
		}

		if err != nil {
			return err
		}

		fmt.Println(name)
	} else {
		if firstOnly {
			for i := 0; i < count; i++ {
				name, err := gen.GenerateFirstName(race, gender)
				if err != nil {
					return err
				}
				fmt.Println(name)
			}
		} else {
			namesList, err := gen.GenerateMultiple(race, gender, count)
			if err != nil {
				return err
			}

			for _, name := range namesList {
				fmt.Println(name)
			}
		}
	}

	return nil
}

func cmdNPC(gen *names.Generator, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: names npc <type>")
	}

	npcType := args[0]
	count := 1

	// Parse options
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "--count=") {
			n, err := strconv.Atoi(strings.TrimPrefix(arg, "--count="))
			if err == nil && n > 0 {
				count = n
			}
		}
	}

	for i := 0; i < count; i++ {
		name, err := gen.GenerateNPCName(npcType)
		if err != nil {
			return err
		}
		fmt.Println(name)
	}

	return nil
}

func cmdList(gen *names.Generator, args []string) error {
	listType := "all"
	if len(args) > 0 {
		listType = strings.ToLower(args[0])
	}

	switch listType {
	case "races", "race":
		fmt.Println("## Races Disponibles")
		fmt.Println()
		for _, race := range gen.GetAvailableRaces() {
			fmt.Printf("- %s\n", race)
		}
	case "npc", "npcs":
		fmt.Println("## Types de PNJ Disponibles")
		fmt.Println()
		for _, npcType := range gen.GetAvailableNPCTypes() {
			fmt.Printf("- %s\n", npcType)
		}
	default:
		fmt.Println("## Races Disponibles")
		fmt.Println()
		for _, race := range gen.GetAvailableRaces() {
			fmt.Printf("- %s\n", race)
		}
		fmt.Println()
		fmt.Println("## Types de PNJ Disponibles")
		fmt.Println()
		for _, npcType := range gen.GetAvailableNPCTypes() {
			fmt.Printf("- %s\n", npcType)
		}
	}

	return nil
}