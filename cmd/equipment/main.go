package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"dungeons/internal/equipment"
)

const dataDir = "data"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	catalog, err := equipment.NewCatalog(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "weapons", "armes":
		err = cmdWeapons(catalog, args)
	case "armor", "armures":
		err = cmdArmor(catalog, args)
	case "gear", "equipement":
		err = cmdGear(catalog, args)
	case "ammo", "munitions":
		err = cmdAmmunition(catalog, args)
	case "show", "get":
		err = cmdShow(catalog, args)
	case "search", "chercher":
		err = cmdSearch(catalog, args)
	case "starting", "depart":
		err = cmdStarting(catalog, args)
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
	fmt.Println(`SkillsWeaver - Equipment Browser - Catalogue d'Équipement BFRPG

UTILISATION:
  sw-equipment <commande> [arguments]

COMMANDES:
  weapons [--type=<type>]        Lister toutes les armes
  armor [--type=<type>]          Lister toutes les armures
  gear                           Lister l'équipement d'aventure
  ammo                           Lister les munitions
  show <id>                      Afficher un item en détail
  search <terme>                 Rechercher par nom (FR ou EN)
  starting <class>               Équipement de départ par classe
  help                           Afficher cette aide

OPTIONS:
  --format=<md|json|short>       Format de sortie (défaut: md)
  --type=<type>                  Filtrer par type

TYPES D'ARMES:
  melee, ranged

TYPES D'ARMURES:
  light, medium, heavy, shield

CLASSES (pour équipement de départ):
  fighter, cleric, magic-user, thief

EXEMPLES:
  sw-equipment weapons                      # Liste les armes
  sw-equipment weapons --type=ranged        # Armes à distance
  sw-equipment armor                        # Liste les armures
  sw-equipment show longsword               # Détails de l'épée longue
  sw-equipment search épée                  # Recherche "épée"
  sw-equipment starting fighter             # Équipement guerrier`)
}

func cmdWeapons(c *equipment.Catalog, args []string) error {
	opts := parseOptions(args)
	weaponType := opts["type"]
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	weapons := c.ListWeapons(weaponType)
	if len(weapons) == 0 {
		fmt.Println("Aucune arme trouvée.")
		return nil
	}

	if weaponType != "" {
		fmt.Printf("## Armes de type '%s'\n\n", weaponType)
	} else {
		fmt.Print("## Toutes les Armes\n\n")
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(weapons, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "md":
		for _, w := range weapons {
			fmt.Print(w.ToMarkdown())
			fmt.Println("---")
		}
	default:
		fmt.Println("| Arme | Dégâts | Type | Coût | Propriétés |")
		fmt.Println("|------|--------|------|------|------------|")
		for _, w := range weapons {
			props := "-"
			if len(w.Properties) > 0 {
				props = strings.Join(w.Properties, ", ")
			}
			fmt.Printf("| %s | %s | %s | %.0f po | %s |\n",
				w.Name, w.Damage, w.Type, w.Cost, props)
		}
	}

	return nil
}

func cmdArmor(c *equipment.Catalog, args []string) error {
	opts := parseOptions(args)
	armorType := opts["type"]
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	armors := c.ListArmor(armorType)
	if len(armors) == 0 {
		fmt.Println("Aucune armure trouvée.")
		return nil
	}

	if armorType != "" {
		fmt.Printf("## Armures de type '%s'\n\n", armorType)
	} else {
		fmt.Print("## Toutes les Armures\n\n")
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(armors, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "md":
		for _, a := range armors {
			fmt.Print(a.ToMarkdown())
			fmt.Println("---")
		}
	default:
		fmt.Println("| Armure | Bonus CA | Type | Coût | Poids |")
		fmt.Println("|--------|----------|------|------|-------|")
		for _, a := range armors {
			fmt.Printf("| %s | +%d | %s | %.0f po | %.1f po |\n",
				a.Name, a.ACBonus, a.Type, a.Cost, a.Weight)
		}
	}

	return nil
}

func cmdGear(c *equipment.Catalog, args []string) error {
	opts := parseOptions(args)
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	gear := c.ListGear()

	fmt.Print("## Équipement d'Aventure\n\n")

	switch format {
	case "json":
		data, err := json.MarshalIndent(gear, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "md":
		for _, g := range gear {
			fmt.Print(g.ToMarkdown())
			fmt.Println("---")
		}
	default:
		fmt.Println("| Objet | Coût | Poids |")
		fmt.Println("|-------|------|-------|")
		for _, g := range gear {
			fmt.Printf("| %s | %.2f po | %.1f po |\n",
				g.Name, g.Cost, g.Weight)
		}
	}

	return nil
}

func cmdAmmunition(c *equipment.Catalog, args []string) error {
	opts := parseOptions(args)
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	ammo := c.ListAmmunition()

	fmt.Print("## Munitions\n\n")

	switch format {
	case "json":
		data, err := json.MarshalIndent(ammo, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		fmt.Println("| Munitions | Coût | Poids |")
		fmt.Println("|-----------|------|-------|")
		for _, a := range ammo {
			fmt.Printf("| %s | %.0f po | %.0f po |\n",
				a.Name, a.Cost, a.Weight)
		}
	}

	return nil
}

func cmdShow(c *equipment.Catalog, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("ID de l'item requis")
	}

	id := args[0]
	opts := parseOptions(args[1:])
	format := opts["format"]
	if format == "" {
		format = "md"
	}

	item, itemType, err := c.GetItem(id)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "short":
		fmt.Printf("[%s] %s\n", itemType, equipment.ItemToShortDescription(item))
	default:
		fmt.Print(equipment.ItemToMarkdown(item))
	}

	return nil
}

func cmdSearch(c *equipment.Catalog, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("terme de recherche requis")
	}

	query := args[0]
	opts := parseOptions(args[1:])
	format := opts["format"]
	if format == "" {
		format = "short"
	}

	results := c.SearchItems(query)
	if len(results) == 0 {
		fmt.Println("Aucun item trouvé.")
		return nil
	}

	fmt.Printf("## %d item(s) trouvé(s) pour '%s'\n\n", len(results), query)

	switch format {
	case "md":
		for _, item := range results {
			fmt.Print(equipment.ItemToMarkdown(item))
			fmt.Println("---")
		}
	default:
		for _, item := range results {
			fmt.Println(equipment.ItemToShortDescription(item))
		}
	}

	return nil
}

func cmdStarting(c *equipment.Catalog, args []string) error {
	if len(args) < 1 {
		// List all classes
		classes := c.GetClasses()
		fmt.Print("## Classes Disponibles\n\n")
		for _, class := range classes {
			fmt.Printf("- %s\n", class)
		}
		fmt.Println("\nUtilisez: sw-equipment starting <class>")
		return nil
	}

	class := args[0]
	se, err := c.GetStartingEquipment(class)
	if err != nil {
		return err
	}

	fmt.Printf("## Équipement de Départ - %s\n\n", strings.Title(class))

	// Required items
	fmt.Print("### Équipement Obligatoire\n\n")
	for _, itemID := range se.Required {
		item, _, err := c.GetItem(itemID)
		if err != nil {
			fmt.Printf("- %s (non trouvé)\n", itemID)
		} else {
			fmt.Printf("- %s\n", equipment.ItemToShortDescription(item))
		}
	}

	// Weapon choices
	fmt.Print("\n### Choix d'Armes (choisir un ensemble)\n\n")
	for i, choice := range se.WeaponChoices {
		fmt.Printf("**Option %d** :\n", i+1)
		for _, itemID := range choice {
			item, _, err := c.GetItem(itemID)
			if err != nil {
				fmt.Printf("  - %s (non trouvé)\n", itemID)
			} else {
				fmt.Printf("  - %s\n", equipment.ItemToShortDescription(item))
			}
		}
	}

	// Armor choices
	if len(se.ArmorChoices) > 0 {
		fmt.Print("\n### Choix d'Armure (choisir une)\n\n")
		for _, itemID := range se.ArmorChoices {
			item, _, err := c.GetItem(itemID)
			if err != nil {
				fmt.Printf("- %s (non trouvé)\n", itemID)
			} else {
				fmt.Printf("- %s\n", equipment.ItemToShortDescription(item))
			}
		}
	} else {
		fmt.Print("\n### Armure\n\nAucune armure disponible pour cette classe.\n")
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
