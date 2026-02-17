package main

import (
	"fmt"
	"os"
	"strings"

	"dungeons/internal/treasure"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	gen, err := treasure.NewGenerator(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "generate", "gen", "roll":
		err = cmdGenerate(gen, args)
	case "types":
		err = cmdTypes(gen)
	case "info":
		err = cmdInfo(gen, args)
	case "items":
		err = cmdItems(gen, args)
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
	fmt.Println(`SkillsWeaver - Treasure Generator - Générateur de Trésors D&D 5e

UTILISATION:
  sw-treasure <commande> [arguments]

COMMANDES:
  generate <type>              Générer un trésor selon un type (A-U)
  generate <type> --count=N    Générer N trésors du même type
  types                        Lister tous les types de trésors
  info <type>                  Afficher les probabilités d'un type
  items [category]             Lister les objets magiques
  help                         Afficher cette aide

OPTIONS:
  --format=<md|json>           Format de sortie (défaut: md)
  --count=<N>                  Nombre de trésors à générer

TYPES DE TRESORS:
  A-H    Trésors de repaire (lairs) - pour groupes de monstres
  I-O    Trésors individuels
  P-U    Trésors individuels mineurs

EXEMPLES:
  sw-treasure generate R              # Trésor type R (gobelin)
  sw-treasure generate A              # Trésor type A (dragon)
  sw-treasure generate B --count=3    # 3 trésors type B
  sw-treasure types                   # Liste tous les types
  sw-treasure info A                  # Détails du type A
  sw-treasure items potions           # Liste toutes les potions
  sw-treasure items weapons           # Liste les armes magiques

CATEGORIES D'OBJETS:
  potions, scrolls, rings, weapons, armor, wands, misc`)
}

func cmdGenerate(gen *treasure.Generator, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("type de trésor requis (A-U)")
	}

	treasureType := strings.ToUpper(args[0])
	opts := parseOptions(args[1:])

	count := 1
	if opts["count"] != "" {
		fmt.Sscanf(opts["count"], "%d", &count)
		if count < 1 {
			count = 1
		}
	}

	format := opts["format"]
	if format == "" {
		format = "md"
	}

	for i := 0; i < count; i++ {
		if count > 1 {
			fmt.Printf("### Trésor %d/%d\n\n", i+1, count)
		}

		t, err := gen.GenerateTreasure(treasureType)
		if err != nil {
			return err
		}

		switch format {
		case "json":
			jsonStr, err := t.ToJSON()
			if err != nil {
				return err
			}
			fmt.Println(jsonStr)
		default:
			fmt.Print(t.ToMarkdown())
		}

		if count > 1 && i < count-1 {
			fmt.Print("---\n\n")
		}
	}

	return nil
}

func cmdTypes(gen *treasure.Generator) error {
	types := gen.GetTreasureTypes()

	fmt.Print("## Types de Trésors D&D 5e\n\n")

	fmt.Println("### Trésors de Repaire (A-H)")
	fmt.Print("*Pour les repaires de groupes de monstres*\n\n")
	for _, t := range types {
		if t >= "A" && t <= "H" {
			tt, _ := gen.GetTreasureType(t)
			fmt.Printf("- **%s** : %s\n", t, tt.Description)
		}
	}

	fmt.Println("\n### Trésors Individuels (I-O)")
	fmt.Print("*Portés par des créatures individuelles*\n\n")
	for _, t := range types {
		if t >= "I" && t <= "O" {
			tt, _ := gen.GetTreasureType(t)
			fmt.Printf("- **%s** : %s\n", t, tt.Description)
		}
	}

	fmt.Println("\n### Trésors Mineurs (P-U)")
	fmt.Print("*Petits trésors et cas particuliers*\n\n")
	for _, t := range types {
		if t >= "P" && t <= "U" {
			tt, _ := gen.GetTreasureType(t)
			fmt.Printf("- **%s** : %s\n", t, tt.Description)
		}
	}

	return nil
}

func cmdInfo(gen *treasure.Generator, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("type de trésor requis")
	}

	treasureType := strings.ToUpper(args[0])
	tt, err := gen.GetTreasureType(treasureType)
	if err != nil {
		return err
	}

	fmt.Printf("## Trésor Type %s\n\n", treasureType)
	fmt.Printf("*%s*\n\n", tt.Description)

	fmt.Print("### Probabilités\n\n")

	if len(tt.Coins) > 0 {
		fmt.Println("**Pièces :**")
		for _, c := range tt.Coins {
			fmt.Printf("- %s : %d%% de %s\n", strings.ToUpper(c.Type), c.Chance, c.Amount)
		}
		fmt.Println()
	}

	if tt.Gems != nil {
		fmt.Printf("**Gemmes :** %d%% de %s\n\n", tt.Gems.Chance, tt.Gems.Amount)
	}

	if tt.Jewelry != nil {
		fmt.Printf("**Bijoux :** %d%% de %s\n\n", tt.Jewelry.Chance, tt.Jewelry.Amount)
	}

	if tt.Potions != nil {
		fmt.Printf("**Potions :** %d%% de %s\n\n", tt.Potions.Chance, tt.Potions.Amount)
	}

	if tt.Scrolls != nil {
		fmt.Printf("**Parchemins :** %d%% de %s\n\n", tt.Scrolls.Chance, tt.Scrolls.Amount)
	}

	if tt.Magic != nil {
		restriction := ""
		if tt.Magic.NoWeapons {
			restriction = " (sans armes)"
		}
		fmt.Printf("**Objets Magiques :** %d%% de %d objet(s)%s\n\n", tt.Magic.Chance, tt.Magic.Items, restriction)
	}

	return nil
}

func cmdItems(gen *treasure.Generator, args []string) error {
	category := ""
	if len(args) > 0 {
		category = strings.ToLower(args[0])
	}

	switch category {
	case "potions", "potion":
		fmt.Print("## Potions\n\n")
		for _, p := range gen.GetPotions() {
			fmt.Printf("- **%s** : %s (%d po)\n", p.NameFR, p.Effect, p.Value)
		}

	case "scrolls", "scroll", "parchemins":
		fmt.Print("## Parchemins\n\n")
		for _, s := range gen.GetScrolls() {
			fmt.Printf("- **%s** (%d po)\n", s.NameFR, s.Value)
		}

	case "rings", "ring", "anneaux":
		fmt.Print("## Anneaux Magiques\n\n")
		for _, r := range gen.GetRings() {
			fmt.Printf("- **%s** : %s (%d po)\n", r.NameFR, r.Effect, r.Value)
		}

	case "weapons", "weapon", "armes":
		fmt.Print("## Armes Magiques\n\n")
		for _, w := range gen.GetWeapons() {
			desc := fmt.Sprintf("+%d", w.Bonus)
			if w.Special != "" {
				desc += ", " + w.Special
			}
			fmt.Printf("- **%s** : %s (%d po)\n", w.NameFR, desc, w.Value)
		}

	case "armor", "armures":
		fmt.Print("## Armures Magiques\n\n")
		for _, a := range gen.GetArmor() {
			fmt.Printf("- **%s** : +%d (%d po)\n", a.NameFR, a.Bonus, a.Value)
		}

	case "wands", "wand", "baguettes":
		fmt.Print("## Baguettes Magiques\n\n")
		for _, w := range gen.GetWands() {
			fmt.Printf("- **%s** : %s charges (%d po)\n", w.NameFR, w.Charges, w.Value)
		}

	case "misc", "divers":
		fmt.Print("## Objets Magiques Divers\n\n")
		for _, m := range gen.GetMiscItems() {
			fmt.Printf("- **%s** : %s (%d po)\n", m.NameFR, m.Effect, m.Value)
		}

	default:
		fmt.Print("## Catégories d'Objets Magiques\n\n")
		fmt.Printf("- **potions** : %d potions\n", len(gen.GetPotions()))
		fmt.Printf("- **scrolls** : %d parchemins\n", len(gen.GetScrolls()))
		fmt.Printf("- **rings** : %d anneaux\n", len(gen.GetRings()))
		fmt.Printf("- **weapons** : %d armes\n", len(gen.GetWeapons()))
		fmt.Printf("- **armor** : %d armures/boucliers\n", len(gen.GetArmor()))
		fmt.Printf("- **wands** : %d baguettes\n", len(gen.GetWands()))
		fmt.Printf("- **misc** : %d objets divers\n", len(gen.GetMiscItems()))
		fmt.Println("\nUtilisez `treasure items <category>` pour voir les détails.")
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
