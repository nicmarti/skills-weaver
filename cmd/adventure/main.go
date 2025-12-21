package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"dungeons/internal/adventure"
)

const (
	adventuresDir = "data/adventures"
	charactersDir = "data/characters"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Ensure directories exist
	os.MkdirAll(adventuresDir, 0755)

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "create":
		err = cmdCreate(args)
	case "list":
		err = cmdList()
	case "show":
		err = cmdShow(args)
	case "delete":
		err = cmdDelete(args)
	case "add-character":
		err = cmdAddCharacter(args)
	case "remove-character":
		err = cmdRemoveCharacter(args)
	case "party":
		err = cmdShowParty(args)
	case "inventory":
		err = cmdInventory(args)
	case "add-gold":
		err = cmdAddGold(args)
	case "add-item":
		err = cmdAddItem(args)
	case "remove-item":
		err = cmdRemoveItem(args)
	case "start-session":
		err = cmdStartSession(args)
	case "end-session":
		err = cmdEndSession(args)
	case "sessions":
		err = cmdListSessions(args)
	case "log":
		err = cmdLog(args)
	case "journal":
		err = cmdJournal(args)
	case "status":
		err = cmdStatus(args)
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
	fmt.Println(`SkillsWeaver - Adventure Manager - Gestionnaire d'Aventures BFRPG

UTILISATION:
  sw-adventure <commande> [arguments]

COMMANDES AVENTURE:
  create <nom> [description]    Créer une nouvelle aventure
  list                          Lister toutes les aventures
  show <nom>                    Afficher les détails d'une aventure
  delete <nom>                  Supprimer une aventure
  status <nom>                  Afficher le statut complet

COMMANDES GROUPE:
  add-character <aventure> <personnage>    Ajouter un personnage
  remove-character <aventure> <personnage> Retirer un personnage
  party <aventure>                         Afficher le groupe

COMMANDES INVENTAIRE:
  inventory <aventure>                     Afficher l'inventaire partagé
  add-gold <aventure> <montant> [source]   Ajouter de l'or
  add-item <aventure> <nom> [quantité]     Ajouter un objet
  remove-item <aventure> <nom> [quantité]  Retirer un objet

COMMANDES SESSION:
  start-session <aventure>                 Démarrer une session
  end-session <aventure> [résumé]          Terminer une session
  sessions <aventure>                      Lister les sessions

COMMANDES JOURNAL:
  log <aventure> <type> <message>          Ajouter une entrée
  journal <aventure> [--session=N]         Afficher le journal

TYPES DE JOURNAL:
  combat, loot, story, note, quest, npc, location, rest

EXEMPLES:
  sw-adventure create "La Mine Perdue" "Une aventure dans les montagnes"
  sw-adventure add-character "La Mine Perdue" "Aldric"
  sw-adventure start-session "La Mine Perdue"
  sw-adventure log "La Mine Perdue" combat "Le groupe affronte 3 gobelins"
  sw-adventure add-gold "La Mine Perdue" 50 "Trésor gobelin"
  sw-adventure end-session "La Mine Perdue" "Le groupe explore le premier niveau"`)
}

func cmdCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure create <nom> [description]")
	}

	name := args[0]
	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ")
	}

	adv := adventure.New(name, description)
	if err := adv.Save(adventuresDir); err != nil {
		return err
	}

	fmt.Printf("## Aventure créée : %s\n\n", name)
	fmt.Printf("- **ID** : %s\n", adv.ID)
	fmt.Printf("- **Slug** : %s\n", adv.Slug)
	fmt.Printf("- **Répertoire** : %s/%s/\n", adventuresDir, adv.Slug)

	if description != "" {
		fmt.Printf("- **Description** : %s\n", description)
	}

	fmt.Println("\nCommandes suivantes :")
	fmt.Printf("  sw-adventure add-character \"%s\" <personnage>\n", name)
	fmt.Printf("  sw-adventure start-session \"%s\"\n", name)

	return nil
}

func cmdList() error {
	adventures, err := adventure.ListAdventures(adventuresDir)
	if err != nil {
		return err
	}

	if len(adventures) == 0 {
		fmt.Println("Aucune aventure trouvée.")
		fmt.Println("\nCréez-en une avec : adventure create \"Nom\" \"Description\"")
		return nil
	}

	fmt.Println("## Aventures")
	fmt.Println()
	fmt.Println("| Nom | Statut | Sessions | Dernière partie |")
	fmt.Println("|-----|--------|----------|-----------------|")

	for _, a := range adventures {
		fmt.Printf("| %s | %s | %d | %s |\n",
			a.Name,
			a.Status,
			a.SessionCount,
			a.LastPlayed.Format("02/01/2006"),
		)
	}

	return nil
}

func cmdShow(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure show <nom>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	fmt.Print(adv.ToMarkdown())
	return nil
}

func cmdDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure delete <nom>")
	}

	if err := adventure.Delete(adventuresDir, args[0]); err != nil {
		return err
	}

	fmt.Printf("Aventure \"%s\" supprimée.\n", args[0])
	return nil
}

func cmdAddCharacter(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: adventure add-character <aventure> <personnage>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	if err := adv.AddCharacter(charactersDir, args[1]); err != nil {
		return err
	}

	fmt.Printf("%s a rejoint l'aventure \"%s\".\n", args[1], adv.Name)
	return nil
}

func cmdRemoveCharacter(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: adventure remove-character <aventure> <personnage>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	if err := adv.RemoveCharacter(args[1]); err != nil {
		return err
	}

	fmt.Printf("%s a quitté l'aventure \"%s\".\n", args[1], adv.Name)
	return nil
}

func cmdShowParty(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure party <aventure>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	party, err := adv.LoadParty()
	if err != nil {
		return err
	}

	characters, err := adv.GetCharacters()
	if err != nil {
		return err
	}

	fmt.Printf("## Groupe - %s\n\n", adv.Name)

	if len(characters) == 0 {
		fmt.Println("*Aucun personnage dans le groupe.*")
		fmt.Printf("\nAjoutez des personnages avec : adventure add-character \"%s\" <nom>\n", adv.Name)
		return nil
	}

	fmt.Printf("**Formation** : %s\n\n", party.Formation)

	fmt.Println("### Membres")
	fmt.Println("| Nom | Race | Classe | Niveau | PV |")
	fmt.Println("|-----|------|--------|--------|-----|")

	for _, c := range characters {
		fmt.Printf("| %s | %s | %s | %d | %d/%d |\n",
			c.Name, c.Race, c.Class, c.Level, c.HitPoints, c.MaxHitPoints)
	}

	if len(party.MarchingOrder) > 0 {
		fmt.Printf("\n**Ordre de marche** : %s\n", strings.Join(party.MarchingOrder, " → "))
	}

	return nil
}

func cmdInventory(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure inventory <aventure>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	inv, err := adv.LoadInventory()
	if err != nil {
		return err
	}

	fmt.Printf("## Inventaire Partagé - %s\n\n", adv.Name)
	fmt.Printf("**Or** : %d po\n\n", inv.Gold)

	if len(inv.Items) == 0 {
		fmt.Println("*Aucun objet dans l'inventaire.*")
		return nil
	}

	fmt.Println("### Objets")
	fmt.Println("| Objet | Quantité | Description |")
	fmt.Println("|-------|----------|-------------|")

	for _, item := range inv.Items {
		desc := item.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fmt.Printf("| %s | %d | %s |\n", item.Name, item.Quantity, desc)
	}

	return nil
}

func cmdAddGold(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: adventure add-gold <aventure> <montant> [source]")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("montant invalide: %s", args[1])
	}

	source := "inconnu"
	if len(args) > 2 {
		source = strings.Join(args[2:], " ")
	}

	if err := adv.AddGold(amount, source); err != nil {
		return err
	}

	inv, _ := adv.LoadInventory()
	fmt.Printf("Or ajouté : %d po (%s)\n", amount, source)
	fmt.Printf("Total : %d po\n", inv.Gold)

	return nil
}

func cmdAddItem(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: adventure add-item <aventure> <nom> [quantité]")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	name := args[1]
	quantity := 1
	if len(args) > 2 {
		q, err := strconv.Atoi(args[2])
		if err == nil {
			quantity = q
		}
	}

	if err := adv.AddItem(name, quantity, "", "groupe"); err != nil {
		return err
	}

	fmt.Printf("Ajouté : %d× %s\n", quantity, name)
	return nil
}

func cmdRemoveItem(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: adventure remove-item <aventure> <nom> [quantité]")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	name := args[1]
	quantity := 1
	if len(args) > 2 {
		q, err := strconv.Atoi(args[2])
		if err == nil {
			quantity = q
		}
	}

	if err := adv.RemoveItem(name, quantity); err != nil {
		return err
	}

	fmt.Printf("Retiré : %d× %s\n", quantity, name)
	return nil
}

func cmdStartSession(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure start-session <aventure>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	session, err := adv.StartSession()
	if err != nil {
		return err
	}

	fmt.Printf("## Session %d démarrée\n\n", session.ID)
	fmt.Printf("**Aventure** : %s\n", adv.Name)
	fmt.Printf("**Début** : %s\n", session.StartedAt.Format("02/01/2006 15:04"))
	fmt.Println("\nBonne partie !")

	return nil
}

func cmdEndSession(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure end-session <aventure> [résumé]")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	summary := ""
	if len(args) > 1 {
		summary = strings.Join(args[1:], " ")
	}

	session, err := adv.EndSession(summary)
	if err != nil {
		return err
	}

	fmt.Printf("## Session %d terminée\n\n", session.ID)
	fmt.Printf("**Durée** : %s\n", session.Duration)

	if summary != "" {
		fmt.Printf("**Résumé** : %s\n", summary)
	}

	return nil
}

func cmdListSessions(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure sessions <aventure>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	sessions, err := adv.GetAllSessions()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("Aucune session enregistrée.")
		fmt.Printf("\nDémarrez une session avec : adventure start-session \"%s\"\n", adv.Name)
		return nil
	}

	fmt.Printf("## Sessions - %s\n\n", adv.Name)
	fmt.Println("| # | Date | Durée | Statut | XP | Or |")
	fmt.Println("|---|------|-------|--------|-----|-----|")

	for _, s := range sessions {
		duration := "-"
		if s.Status == "completed" {
			duration = s.Duration
		}
		fmt.Printf("| %d | %s | %s | %s | %d | %d |\n",
			s.ID,
			s.StartedAt.Format("02/01/2006"),
			duration,
			s.Status,
			s.XPAwarded,
			s.GoldFound,
		)
	}

	// Stats
	totalXP, _ := adv.GetTotalXPAwarded()
	totalGold, _ := adv.GetTotalGoldFound()
	totalTime, _ := adv.GetTotalPlayTime()

	fmt.Printf("\n**Total** : %d sessions, %s de jeu, %d XP, %d po\n",
		len(sessions),
		formatDuration(totalTime),
		totalXP,
		totalGold,
	)

	return nil
}

func cmdLog(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: adventure log <aventure> <type> <message>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	entryType := args[1]
	message := strings.Join(args[2:], " ")

	if err := adv.LogEvent(entryType, message); err != nil {
		return err
	}

	fmt.Printf("Entrée ajoutée au journal : [%s] %s\n", entryType, message)
	return nil
}

func cmdJournal(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure journal <aventure> [--session=N] [--recent=N]")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	// Parse options
	sessionID := 0
	recentN := 0

	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "--session=") {
			id, _ := strconv.Atoi(strings.TrimPrefix(arg, "--session="))
			sessionID = id
		}
		if strings.HasPrefix(arg, "--recent=") {
			n, _ := strconv.Atoi(strings.TrimPrefix(arg, "--recent="))
			recentN = n
		}
	}

	if sessionID > 0 {
		md, err := adv.SessionSummaryMarkdown(sessionID)
		if err != nil {
			return err
		}
		fmt.Print(md)
		return nil
	}

	if recentN > 0 {
		entries, err := adv.GetRecentEntries(recentN)
		if err != nil {
			return err
		}

		fmt.Printf("## %d dernières entrées\n\n", recentN)
		for _, e := range entries {
			icon := getTypeIcon(e.Type)
			timestamp := e.Timestamp.Format("02/01 15:04")
			fmt.Printf("- `%s` %s %s\n", timestamp, icon, e.Content)
		}
		return nil
	}

	md, err := adv.JournalToMarkdown()
	if err != nil {
		return err
	}
	fmt.Print(md)

	return nil
}

func cmdStatus(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure status <aventure>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	fmt.Printf("# %s\n\n", adv.Name)

	if adv.Description != "" {
		fmt.Printf("*%s*\n\n", adv.Description)
	}

	// Adventure info
	fmt.Println("## Informations")
	fmt.Printf("- **Statut** : %s\n", adv.Status)
	fmt.Printf("- **Sessions** : %d\n", adv.SessionCount)
	fmt.Printf("- **Dernière partie** : %s\n\n", adv.LastPlayed.Format("02/01/2006 15:04"))

	// Current session?
	session, _ := adv.GetCurrentSession()
	if session != nil {
		fmt.Println("## Session en cours")
		fmt.Printf("- **Session #%d** démarrée le %s\n\n", session.ID, session.StartedAt.Format("02/01/2006 15:04"))
	}

	// Party
	party, _ := adv.LoadParty()
	characters, _ := adv.GetCharacters()

	fmt.Println("## Groupe")
	if len(characters) == 0 {
		fmt.Println("*Aucun membre*")
		fmt.Println()
	} else {
		fmt.Printf("**Formation** : %s\n", party.Formation)
		for _, c := range characters {
			fmt.Printf("- %s (%s %s N%d) - PV: %d/%d\n",
				c.Name, c.Race, c.Class, c.Level, c.HitPoints, c.MaxHitPoints)
		}
		fmt.Println()
	}

	// Inventory
	inv, _ := adv.LoadInventory()
	fmt.Println("## Inventaire")
	fmt.Printf("**Or** : %d po\n", inv.Gold)
	if len(inv.Items) > 0 {
		fmt.Printf("**Objets** : %d\n", len(inv.Items))
	}
	fmt.Println()

	// Recent journal entries
	entries, _ := adv.GetRecentEntries(5)
	if len(entries) > 0 {
		fmt.Println("## Derniers événements")
		for _, e := range entries {
			icon := getTypeIcon(e.Type)
			timestamp := e.Timestamp.Format("02/01 15:04")
			fmt.Printf("- `%s` %s %s\n", timestamp, icon, e.Content)
		}
	}

	return nil
}

// Helper functions
func formatDuration(d interface{}) string {
	switch v := d.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", d)
	}
}

func getTypeIcon(entryType string) string {
	icons := map[string]string{
		"combat":   "⚔️",
		"loot":     "💰",
		"story":    "📖",
		"note":     "📝",
		"quest":    "🎯",
		"npc":      "👤",
		"location": "📍",
		"rest":     "🏕️",
		"death":    "💀",
		"levelup":  "⬆️",
		"session":  "🎲",
		"party":    "👥",
		"xp":       "✨",
		"expense":  "💸",
		"use":      "🔧",
	}

	if icon, ok := icons[entryType]; ok {
		return icon
	}
	return "•"
}
