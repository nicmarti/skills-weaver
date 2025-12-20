// Command dice provides a CLI for rolling dice using RPG notation.
//
// Usage:
//
//	dice roll 2d6+3           Roll 2d6 and add 3
//	dice roll 4d6kh3          Roll 4d6, keep highest 3 (stat generation)
//	dice roll d20 --advantage Roll d20 with advantage
//	dice stats                Generate a full set of 6 stats (4d6kh3 method)
//	dice stats --classic      Generate stats using 3d6 method
package main

import (
	"fmt"
	"os"
	"strings"

	"dungeons/internal/dice"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	roller := dice.New()

	switch command {
	case "roll":
		handleRoll(roller, os.Args[2:])
	case "stats":
		handleStats(roller, os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		// Try to parse as a direct roll expression
		result, err := roller.Roll(command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			printUsage()
			os.Exit(1)
		}
		printResult(result)
	}
}

func handleRoll(roller *dice.Roller, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing dice expression")
		fmt.Fprintln(os.Stderr, "Usage: dice roll <expression> [--advantage|--disadvantage]")
		os.Exit(1)
	}

	expression := args[0]
	var result *dice.Result

	// Check for advantage/disadvantage flags
	hasAdvantage := containsFlag(args, "--advantage", "-a")
	hasDisadvantage := containsFlag(args, "--disadvantage", "-d")

	if hasAdvantage && hasDisadvantage {
		fmt.Fprintln(os.Stderr, "Error: cannot use both advantage and disadvantage")
		os.Exit(1)
	}

	if hasAdvantage {
		result = roller.RollAdvantage()
	} else if hasDisadvantage {
		result = roller.RollDisadvantage()
	} else {
		var err error
		result, err = roller.Roll(expression)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	printResult(result)
}

func handleStats(roller *dice.Roller, args []string) {
	classic := containsFlag(args, "--classic", "-c")

	var results []dice.Result
	var method string

	if classic {
		results = roller.RollStatsClassic()
		method = "3d6 (classic)"
	} else {
		results = roller.RollStats()
		method = "4d6kh3 (standard)"
	}

	statNames := []string{"Force", "Intelligence", "Sagesse", "Dextérité", "Constitution", "Charisme"}

	fmt.Printf("## Génération de caractéristiques (%s)\n\n", method)
	fmt.Println("| Caractéristique | Jets | Total |")
	fmt.Println("|-----------------|------|-------|")

	total := 0
	for i, result := range results {
		total += result.Total
		rolls := formatRolls(result.Rolls, result.KeptIndices)
		fmt.Printf("| %-15s | %s | **%2d** |\n", statNames[i], rolls, result.Total)
	}

	fmt.Println()
	fmt.Printf("Total: %d | Moyenne: %.1f\n", total, float64(total)/6.0)
}

func printResult(result *dice.Result) {
	fmt.Println(result.String())
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

func containsFlag(args []string, flags ...string) bool {
	for _, arg := range args {
		for _, flag := range flags {
			if arg == flag {
				return true
			}
		}
	}
	return false
}

func printUsage() {
	fmt.Println(`Dice Roller - Outil de lancer de dés pour JdR

USAGE:
    dice <command> [arguments]

COMMANDES:
    roll <expression> [options]   Lance des dés selon la notation standard
    stats [--classic]             Génère 6 caractéristiques pour un personnage
    help                          Affiche cette aide

EXPRESSIONS SUPPORTÉES:
    d20         Lance un d20
    2d6         Lance 2d6
    2d6+3       Lance 2d6 et ajoute 3
    4d6kh3      Lance 4d6, garde les 3 plus hauts
    2d20kl1     Lance 2d20, garde le plus bas (désavantage)

OPTIONS POUR ROLL:
    --advantage, -a       Lance avec avantage (2d20kh1)
    --disadvantage, -d    Lance avec désavantage (2d20kl1)

OPTIONS POUR STATS:
    --classic, -c         Utilise la méthode 3d6 au lieu de 4d6kh3

EXEMPLES:
    dice roll d20
    dice roll 2d6+3
    dice roll 4d6kh3
    dice roll d20 --advantage
    dice stats
    dice stats --classic
    dice 3d8+2                  # Raccourci pour "dice roll 3d8+2"`)
}
