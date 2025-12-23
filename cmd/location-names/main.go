package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"dungeons/internal/locations"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	gen, err := locations.NewGenerator(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "city", "ville":
		err = cmdGenerate(gen, "city", args)
	case "town", "bourg":
		err = cmdGenerate(gen, "town", args)
	case "village":
		err = cmdGenerate(gen, "village", args)
	case "region", "région":
		err = cmdGenerate(gen, "region", args)
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
	fmt.Println(`Location Name Generator - Générateur de Noms de Lieux BFRPG

UTILISATION:
  sw-location-names <commande> [options]

COMMANDES:
  city <royaume>                Générer un nom de cité
  town <royaume>                Générer un nom de bourg
  village <royaume>             Générer un nom de village
  region <royaume>              Générer un nom de région
  list [kingdoms|types]         Lister les options disponibles
  help                          Afficher cette aide

OPTIONS:
  --kingdom=<royaume>           Royaume (pour city/town/village/region)
  --count=<N>                   Nombre de noms à générer (défaut: 1)

ROYAUMES DISPONIBLES:
  valdorine     Royaume maritime et marchand
  karvath       Empire militariste et défensif
  lumenciel     Théocratie hypocrite
  astrene       Empire décadent et érudit

EXEMPLES:
  sw-location-names city --kingdom=valdorine
  sw-location-names town --kingdom=karvath --count=5
  sw-location-names village --kingdom=lumenciel
  sw-location-names region --kingdom=astrene`)
}

func cmdGenerate(gen *locations.Generator, locationType string, args []string) error {
	kingdom := ""
	count := 1

	// Parse options
	for _, arg := range args {
		if strings.HasPrefix(arg, "--kingdom=") {
			kingdom = strings.TrimPrefix(arg, "--kingdom=")
		} else if strings.HasPrefix(arg, "--count=") {
			n, err := strconv.Atoi(strings.TrimPrefix(arg, "--count="))
			if err == nil && n > 0 {
				count = n
			}
		} else if !strings.HasPrefix(arg, "--") {
			// First non-flag argument is kingdom
			if kingdom == "" {
				kingdom = arg
			}
		}
	}

	if kingdom == "" {
		return fmt.Errorf("usage: sw-location-names %s <royaume> [--count=N]\n\nRoyaumes disponibles: valdorine, karvath, lumenciel, astrene", locationType)
	}

	if count == 1 {
		var name string
		var err error

		switch locationType {
		case "city":
			name, err = gen.GenerateCity(kingdom)
		case "town":
			name, err = gen.GenerateTown(kingdom)
		case "village":
			name, err = gen.GenerateVillage(kingdom)
		case "region":
			name, err = gen.GenerateRegion(kingdom)
		}

		if err != nil {
			return err
		}

		fmt.Println(name)
	} else {
		names, err := gen.GenerateMultiple(kingdom, locationType, count)
		if err != nil {
			return err
		}

		for _, name := range names {
			fmt.Println(name)
		}
	}

	return nil
}

func cmdList(gen *locations.Generator, args []string) error {
	listType := "all"
	if len(args) > 0 {
		listType = strings.ToLower(args[0])
	}

	switch listType {
	case "kingdoms", "kingdom", "royaumes", "royaume":
		fmt.Println("## Royaumes Disponibles")
		fmt.Println()
		fmt.Println("- valdorine    (Royaume maritime et marchand)")
		fmt.Println("- karvath      (Empire militariste et défensif)")
		fmt.Println("- lumenciel    (Théocratie hypocrite)")
		fmt.Println("- astrene      (Empire décadent et érudit)")

	case "types", "type":
		fmt.Println("## Types de Lieux Disponibles")
		fmt.Println()
		fmt.Println("- city         (Cité majeure)")
		fmt.Println("- town         (Bourg)")
		fmt.Println("- village      (Village)")
		fmt.Println("- region       (Région géographique)")

	default:
		fmt.Println("## Royaumes Disponibles")
		fmt.Println()
		fmt.Println("- valdorine    (Royaume maritime et marchand)")
		fmt.Println("- karvath      (Empire militariste et défensif)")
		fmt.Println("- lumenciel    (Théocratie hypocrite)")
		fmt.Println("- astrene      (Empire décadent et érudit)")
		fmt.Println()
		fmt.Println("## Types de Lieux Disponibles")
		fmt.Println()
		fmt.Println("- city         (Cité majeure)")
		fmt.Println("- town         (Bourg)")
		fmt.Println("- village      (Village)")
		fmt.Println("- region       (Région géographique)")
	}

	return nil
}
