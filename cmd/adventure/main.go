package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/ai"
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
	case "enrich":
		err = cmdEnrich(args)
	case "status":
		err = cmdStatus(args)
	case "migrate-journal":
		err = cmdMigrateJournal(args)
	case "validate-journal":
		err = cmdValidateJournal(args)
	case "sync-characters":
		err = cmdSyncCharacters(args)
	case "archive":
		err = cmdArchive(args)
	case "unarchive":
		err = cmdUnarchive(args)
	case "list-archived":
		err = cmdListArchived()
	case "purge-maps":
		err = cmdPurgeMaps()
	case "clean-session":
		err = cmdCleanSession(args)
	case "inspect-sessions":
		err = cmdInspectSessions(args)
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
	fmt.Println(`SkillsWeaver - Adventure Manager - Gestionnaire d'Aventures D&D 5e

UTILISATION:
  sw-adventure <commande> [arguments]

COMMANDES AVENTURE:
  create <nom> [description]    Créer une nouvelle aventure
  list                          Lister toutes les aventures
  show <nom>                    Afficher les détails d'une aventure
  delete <nom>                  Supprimer une aventure (avec confirmation)
  sync-characters <nom>         Sauvegarder la progression des personnages
  archive <nom>                 Archiver une aventure (réversible)
  unarchive <nom>               Restaurer une aventure archivée
  list-archived                 Lister les aventures archivées
  purge-maps                    Supprimer toutes les maps globales
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
      [--description="English description"]
      [--description-fr="Description française"]
  journal <aventure> [--session=N]         Afficher le journal
  enrich <aventure> [options]              Enrichir le journal avec IA
      [--session=N]           Filtrer par session
      [--recent=N]            Dernières N entrées
      [--from=ID]             À partir de l'ID
      [--to=ID]               Jusqu'à l'ID
      [--batch=N]             Taille des lots (défaut: 10)
      [--force]               Re-générer les descriptions existantes
      [--dry-run]             Prévisualiser sans générer

TYPES DE JOURNAL:
  combat, loot, story, note, quest, npc, location, rest

COMMANDES MAINTENANCE:
  migrate-journal <aventure>    Diviser journal.json en fichiers par session
  validate-journal <aventure>   Valider l'intégrité des journaux
  inspect-sessions <aventure>   Analyser les sessions pour détecter les problèmes
  clean-session <aventure> <session_id>  Supprimer une session invalide

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

	fmt.Printf("Supprimer définitivement l'aventure \"%s\" ? (oui/non): ", args[0])
	var response string
	fmt.Scanln(&response)
	if response != "oui" {
		fmt.Println("Annulé.")
		return nil
	}

	// Sync characters to global before deleting
	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err == nil {
		if synced, syncErr := adv.SyncCharactersToGlobal(charactersDir); syncErr != nil {
			fmt.Printf("Avertissement : impossible de sauvegarder les personnages : %v\n", syncErr)
		} else if len(synced) > 0 {
			fmt.Printf("Progression sauvegardée pour : %s\n", strings.Join(synced, ", "))
		}
	}

	if err := adventure.Delete(adventuresDir, args[0]); err != nil {
		return err
	}

	fmt.Printf("Aventure \"%s\" supprimée.\n", args[0])
	return nil
}

func cmdSyncCharacters(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure sync-characters <nom>")
	}

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	synced, err := adv.SyncCharactersToGlobal(charactersDir)
	if err != nil {
		return err
	}

	if len(synced) == 0 {
		fmt.Println("Aucun personnage à synchroniser.")
		return nil
	}

	fmt.Printf("Progression sauvegardée pour %d personnage(s) :\n", len(synced))
	for _, name := range synced {
		fmt.Printf("  - %s → %s/%s.json\n", name, charactersDir, strings.ToLower(strings.ReplaceAll(name, " ", "-")))
	}
	return nil
}

func cmdArchive(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure archive <nom>")
	}

	if err := adventure.Archive(adventuresDir, args[0]); err != nil {
		return err
	}

	fmt.Printf("Aventure \"%s\" archivée.\n", args[0])
	fmt.Println("Restaurez-la avec : sw-adventure unarchive \"" + args[0] + "\"")
	return nil
}

func cmdUnarchive(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure unarchive <nom>")
	}

	if err := adventure.Unarchive(adventuresDir, args[0]); err != nil {
		return err
	}

	fmt.Printf("Aventure \"%s\" restaurée.\n", args[0])
	return nil
}

func cmdListArchived() error {
	adventures, err := adventure.ListArchived(adventuresDir)
	if err != nil {
		return err
	}

	if len(adventures) == 0 {
		fmt.Println("Aucune aventure archivée.")
		return nil
	}

	fmt.Println("## Aventures archivées")
	fmt.Println()
	fmt.Println("| Nom | Sessions | Dernière partie |")
	fmt.Println("|-----|----------|-----------------|")

	for _, a := range adventures {
		fmt.Printf("| %s | %d | %s |\n",
			a.Name,
			a.SessionCount,
			a.LastPlayed.Format("02/01/2006"),
		)
	}

	return nil
}

func cmdPurgeMaps() error {
	fmt.Print("Supprimer TOUTES les maps globales dans data/maps/ ? (oui/non): ")
	var response string
	fmt.Scanln(&response)
	if response != "oui" {
		fmt.Println("Annulé.")
		return nil
	}

	count, err := adventure.PurgeMaps("data")
	if err != nil {
		return err
	}

	fmt.Printf("%d fichier(s) supprimé(s) dans data/maps/.\n", count)
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
			c.Name, c.Species, c.Class, c.Level, c.HitPoints, c.MaxHitPoints)
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
		return fmt.Errorf("usage: adventure log <aventure> <type> <message> [--description=\"...\"] [--description-fr=\"...\"]")
	}

	advName := args[0]
	entryType := args[1]

	// Parse message and optional description flags
	var messageParts []string
	description := ""
	descriptionFr := ""

	for _, arg := range args[2:] {
		if strings.HasPrefix(arg, "--description=") {
			description = strings.TrimPrefix(arg, "--description=")
		} else if strings.HasPrefix(arg, "--description-fr=") {
			descriptionFr = strings.TrimPrefix(arg, "--description-fr=")
		} else {
			messageParts = append(messageParts, arg)
		}
	}

	message := strings.Join(messageParts, " ")

	adv, err := adventure.LoadByName(adventuresDir, advName)
	if err != nil {
		return err
	}

	// Use appropriate method based on whether descriptions are provided
	if description != "" || descriptionFr != "" {
		err = adv.LogEventWithDescriptions(entryType, message, description, descriptionFr)
	} else {
		err = adv.LogEvent(entryType, message) // Legacy
	}
	if err != nil {
		return err
	}

	fmt.Printf("Entrée ajoutée : [%s] %s\n", entryType, message)
	if description != "" {
		fmt.Printf("  📝 EN: %s\n", description)
	}
	if descriptionFr != "" {
		fmt.Printf("  📝 FR: %s\n", descriptionFr)
	}
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
				c.Name, c.Species, c.Class, c.Level, c.HitPoints, c.MaxHitPoints)
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

func cmdEnrich(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adventure enrich <aventure> [options]\n\nOptions:\n  --session=N    Enrich entries from session N\n  --recent=N     Enrich last N entries\n  --from=ID      Start from entry ID\n  --to=ID        End at entry ID\n  --batch=N      Batch size (default 10)\n  --force        Re-enrich entries with existing descriptions\n  --dry-run      Preview entries without enriching")
	}

	opts := parseEnrichOptions(args[1:])

	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return err
	}

	// Get entries to enrich
	entries, err := adv.GetEntriesToEnrich(opts)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("✓ No entries to enrich")
		return nil
	}

	fmt.Printf("Found %d entries to enrich\n\n", len(entries))

	if opts.DryRun {
		// Preview mode
		for _, e := range entries {
			ctx, _ := adv.GetEnrichmentContext(e)
			fmt.Printf("\n[%d] %s: %s\n", e.ID, e.Type, e.Content)
			if len(ctx.PartyMembers) > 0 {
				fmt.Printf("  Party: %s\n", strings.Join(ctx.PartyMembers, ", "))
			}
			if len(ctx.RecentEntries) > 0 {
				fmt.Printf("  Context: %s\n", strings.Join(ctx.RecentEntries, " → "))
			}
			if ctx.SessionInfo != "" {
				fmt.Printf("  %s\n", ctx.SessionInfo)
			}
		}
		fmt.Printf("\n%d entries ready for enrichment\n", len(entries))
		fmt.Println("Run without --dry-run to enrich with AI")
		return nil
	}

	// Create AI enricher
	enricher, err := ai.NewEnricher()
	if err != nil {
		fmt.Println("✗ AI enrichment requires ANTHROPIC_API_KEY")
		fmt.Printf("  Error: %v\n", err)
		fmt.Println("\nSet your API key:")
		fmt.Println("  export ANTHROPIC_API_KEY=\"your-key-here\"")
		fmt.Println("\nOr use --dry-run to preview entries without enriching")
		return err
	}

	fmt.Printf("Enriching %d entries with Claude...\n\n", len(entries))

	// Process in batches with interactive confirmation
	successCount := 0
	for i := 0; i < len(entries); i += opts.BatchSize {
		end := i + opts.BatchSize
		if end > len(entries) {
			end = len(entries)
		}
		batch := entries[i:end]

		// Enrich batch
		results := make(map[int]*ai.EnrichmentResult)
		for _, entry := range batch {
			ctx, _ := adv.GetEnrichmentContext(entry)
			result, err := enricher.EnrichEntry(entry, ctx)
			if err != nil {
				fmt.Printf("  ✗ Entry %d: %v\n", entry.ID, err)
				continue
			}
			results[entry.ID] = result
		}

		if len(results) == 0 {
			fmt.Println("  ✗ No entries enriched in this batch")
			continue
		}

		// Display batch results
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Batch %d-%d (%d enriched)\n", i+1, end, len(results))
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		for _, entry := range batch {
			result, ok := results[entry.ID]
			if !ok {
				continue
			}

			fmt.Printf("[%d] %s: \"%s\"\n\n", entry.ID, entry.Type, entry.Content)
			fmt.Printf("📝 EN (%d words):\n%s\n\n", len(strings.Fields(result.Description)), result.Description)
			fmt.Printf("📝 FR (%d words):\n%s\n\n", len(strings.Fields(result.DescriptionFr)), result.DescriptionFr)
		}

		// Interactive confirmation
		fmt.Print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Print("[A]ccept  [S]kip  [Q]uit: ")
		var choice string
		fmt.Scanln(&choice)

		switch strings.ToLower(choice) {
		case "a", "accept":
			for entryID, result := range results {
				if err := adv.UpdateEntryDescriptions(entryID, result.Description, result.DescriptionFr); err != nil {
					fmt.Printf("✗ Error saving entry %d: %v\n", entryID, err)
				} else {
					successCount++
				}
			}
			fmt.Printf("✓ Saved %d entries\n\n", len(results))
		case "s", "skip":
			fmt.Print("⊘ Skipped batch\n")
		case "q", "quit":
			fmt.Printf("\n✓ Enrichment stopped. %d entries enriched.\n", successCount)
			return nil
		default:
			fmt.Print("⊘ Invalid choice, skipping batch\n")
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("✓ Enrichment complete: %d entries enriched\n", successCount)
	return nil
}

func parseEnrichOptions(args []string) adventure.EnrichOptions {
	opts := adventure.EnrichOptions{
		BatchSize: 10, // Default batch size
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "--session=") {
			fmt.Sscanf(arg, "--session=%d", &opts.SessionID)
		} else if strings.HasPrefix(arg, "--recent=") {
			fmt.Sscanf(arg, "--recent=%d", &opts.RecentN)
		} else if strings.HasPrefix(arg, "--from=") {
			fmt.Sscanf(arg, "--from=%d", &opts.FromID)
		} else if strings.HasPrefix(arg, "--to=") {
			fmt.Sscanf(arg, "--to=%d", &opts.ToID)
		} else if strings.HasPrefix(arg, "--batch=") {
			fmt.Sscanf(arg, "--batch=%d", &opts.BatchSize)
		} else if arg == "--force" {
			opts.Force = true
		} else if arg == "--dry-run" {
			opts.DryRun = true
		}
	}

	return opts
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

// cmdMigrateJournal migrates a monolithic journal.json to per-session files.
func cmdMigrateJournal(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: migrate-journal <aventure>")
	}

	// Load adventure
	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return fmt.Errorf("chargement aventure: %w", err)
	}

	fmt.Printf("🔄 Migration du journal: %s\n\n", adv.Name)

	// Check if legacy journal.json exists
	legacyPath := adv.BasePath() + "/journal.json"
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		fmt.Println("✅ Aucun journal.json trouvé - aventure déjà migrée ou pas de journal")
		return nil
	}

	// Check if already migrated (has session files)
	metaPath := adv.BasePath() + "/journal-meta.json"
	if _, err := os.Stat(metaPath); err == nil {
		fmt.Println("⚠️  L'aventure semble déjà migrée (journal-meta.json existe)")
		fmt.Println("   Voulez-vous continuer? Le journal.json sera sauvegardé.")
		// For now, continue anyway
	}

	// Load legacy journal
	fmt.Println("📖 Chargement du journal monolithique...")
	journal, err := adv.LoadJournal()
	if err != nil {
		return fmt.Errorf("chargement journal: %w", err)
	}

	if len(journal.Entries) == 0 {
		fmt.Println("✅ Journal vide - aucune migration nécessaire")
		return nil
	}

	fmt.Printf("   Trouvé: %d entrées (NextID: %d)\n\n", len(journal.Entries), journal.NextID)

	// Group entries by session
	fmt.Println("📊 Analyse des sessions...")
	sessionGroups := make(map[int][]adventure.JournalEntry)
	for _, entry := range journal.Entries {
		sessionGroups[entry.SessionID] = append(sessionGroups[entry.SessionID], entry)
	}

	var sessionIDs []int
	for id := range sessionGroups {
		sessionIDs = append(sessionIDs, id)
	}

	fmt.Printf("   Sessions trouvées: %v\n", sessionIDs)
	for _, id := range sessionIDs {
		sessionName := fmt.Sprintf("session-%d", id)
		if id == 0 {
			sessionName = "hors session"
		}
		fmt.Printf("   - %s: %d entrées\n", sessionName, len(sessionGroups[id]))
	}
	fmt.Println()

	// Create backup
	fmt.Println("💾 Sauvegarde du journal.json...")
	backupPath := legacyPath + ".backup"
	if err := copyFile(legacyPath, backupPath); err != nil {
		return fmt.Errorf("création backup: %w", err)
	}
	fmt.Printf("   ✅ Backup créé: %s\n\n", backupPath)

	// Create journal-meta.json
	fmt.Println("📝 Création de journal-meta.json...")
	meta := &adventure.JournalMetadata{
		NextID:     journal.NextID,
		Categories: journal.Categories,
		LastUpdate: time.Now(),
	}
	if err := adv.SaveJournalMetadata(meta); err != nil {
		return fmt.Errorf("sauvegarde metadata: %w", err)
	}
	fmt.Print("   ✅ Métadonnées créées\n")

	// Create session journal files
	fmt.Println("📁 Création des fichiers de session...")
	for sessionID, entries := range sessionGroups {
		sessionJournal := &adventure.SessionJournal{
			SessionID: sessionID,
			Entries:   entries,
		}

		// Sort entries by timestamp
		sort.Slice(sessionJournal.Entries, func(i, j int) bool {
			return sessionJournal.Entries[i].Timestamp.Before(sessionJournal.Entries[j].Timestamp)
		})

		if err := adv.SaveSessionJournal(sessionJournal); err != nil {
			return fmt.Errorf("sauvegarde session %d: %w", sessionID, err)
		}

		sessionName := fmt.Sprintf("journal-session-%d.json", sessionID)
		fmt.Printf("   ✅ %s (%d entrées)\n", sessionName, len(entries))
	}
	fmt.Println()

	// Migrate images
	fmt.Println("🖼️  Migration des images...")
	imagesDir := adv.BasePath() + "/images"
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		fmt.Print("   ℹ️  Aucun répertoire images/ trouvé\n")
	} else {
		migratedCount, err := migrateImages(adv, imagesDir, sessionGroups)
		if err != nil {
			fmt.Printf("   ⚠️  Erreur migration images: %v\n", err)
		} else {
			fmt.Printf("   ✅ %d images migrées vers images/session-N/\n\n", migratedCount)
		}
	}

	// Validate migration
	fmt.Println("✔️  Validation de la migration...")
	if err := validateMigration(adv, journal); err != nil {
		return fmt.Errorf("validation échouée: %w", err)
	}
	fmt.Print("   ✅ Validation réussie\n")

	// Archive journal.json
	fmt.Println("📦 Archivage de journal.json...")
	archivePath := legacyPath + ".archive"
	if err := os.Rename(legacyPath, archivePath); err != nil {
		fmt.Printf("   ⚠️  Impossible de renommer: %v\n", err)
		fmt.Println("   Vous pouvez supprimer journal.json manuellement")
	} else {
		fmt.Printf("   ✅ Renommé en: journal.json.archive\n")
	}

	fmt.Println("\n🎉 Migration terminée avec succès!")
	fmt.Println("   Le journal est maintenant divisé en fichiers par session")
	fmt.Printf("   Backup disponible: %s\n", backupPath)

	return nil
}

// migrateImages moves images to session-specific directories.
func migrateImages(adv *adventure.Adventure, imagesDir string, sessionGroups map[int][]adventure.JournalEntry) (int, error) {
	// Create entry ID to session mapping
	entryToSession := make(map[int]int)
	for sessionID, entries := range sessionGroups {
		for _, entry := range entries {
			entryToSession[entry.ID] = sessionID
		}
	}

	// Read images directory
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return 0, err
	}

	migratedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Parse entry ID from filename: journal_NNN_type_model.png
		var entryID int
		if n, _ := fmt.Sscanf(filename, "journal_%d_", &entryID); n != 1 {
			continue // Skip non-journal images
		}

		// Find session for this entry
		sessionID, ok := entryToSession[entryID]
		if !ok {
			fmt.Printf("   ⚠️  Image %s: entrée %d non trouvée\n", filename, entryID)
			continue
		}

		// Create session directory
		sessionDir := fmt.Sprintf("%s/session-%d", imagesDir, sessionID)
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			return migratedCount, err
		}

		// Move image
		oldPath := fmt.Sprintf("%s/%s", imagesDir, filename)
		newPath := fmt.Sprintf("%s/%s", sessionDir, filename)
		if err := os.Rename(oldPath, newPath); err != nil {
			fmt.Printf("   ⚠️  Erreur déplacement %s: %v\n", filename, err)
			continue
		}

		migratedCount++
	}

	return migratedCount, nil
}

// validateMigration checks that the migration was successful.
func validateMigration(adv *adventure.Adventure, originalJournal *adventure.Journal) error {
	// Load migrated journal (should aggregate all sessions)
	migratedJournal, err := adv.LoadJournal()
	if err != nil {
		return fmt.Errorf("chargement journal migré: %w", err)
	}

	// Check entry count
	if len(migratedJournal.Entries) != len(originalJournal.Entries) {
		return fmt.Errorf("nombre d'entrées: %d != %d", len(migratedJournal.Entries), len(originalJournal.Entries))
	}

	// Check NextID
	if migratedJournal.NextID != originalJournal.NextID {
		return fmt.Errorf("NextID: %d != %d", migratedJournal.NextID, originalJournal.NextID)
	}

	// Check all entry IDs are unique
	seenIDs := make(map[int]bool)
	for _, entry := range migratedJournal.Entries {
		if seenIDs[entry.ID] {
			return fmt.Errorf("ID dupliqué trouvé: %d", entry.ID)
		}
		seenIDs[entry.ID] = true
	}

	return nil
}

// cmdValidateJournal validates the integrity of split journal files.
func cmdValidateJournal(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: validate-journal <aventure>")
	}

	// Load adventure
	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return fmt.Errorf("chargement aventure: %w", err)
	}

	fmt.Printf("🔍 Validation du journal: %s\n\n", adv.Name)

	// Check for metadata file
	metaPath := adv.BasePath() + "/journal-meta.json"
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return fmt.Errorf("journal-meta.json non trouvé - aventure pas encore migrée?")
	}

	// Load all journals
	journal, err := adv.LoadJournal()
	if err != nil {
		return fmt.Errorf("chargement journal: %w", err)
	}

	fmt.Printf("📊 Statistiques:\n")
	fmt.Printf("   Entrées totales: %d\n", len(journal.Entries))
	fmt.Printf("   NextID: %d\n", journal.NextID)
	fmt.Printf("   Catégories: %d\n\n", len(journal.Categories))

	// Validate: All entry IDs unique
	fmt.Println("✔️  Vérification des IDs uniques...")
	seenIDs := make(map[int]bool)
	duplicates := []int{}
	for _, entry := range journal.Entries {
		if seenIDs[entry.ID] {
			duplicates = append(duplicates, entry.ID)
		}
		seenIDs[entry.ID] = true
	}
	if len(duplicates) > 0 {
		return fmt.Errorf("IDs dupliqués trouvés: %v", duplicates)
	}
	fmt.Print("   ✅ Tous les IDs sont uniques\n")

	// Validate: Chronological order within sessions
	fmt.Println("✔️  Vérification de l'ordre chronologique par session...")
	sessions, _ := adv.GetAllSessions()
	for _, session := range sessions {
		entries, _ := adv.GetEntriesBySession(session.ID)
		for i := 1; i < len(entries); i++ {
			if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
				return fmt.Errorf("session %d: ordre chronologique incorrect", session.ID)
			}
		}
	}
	// Check session 0 (out-of-session)
	entries, _ := adv.GetEntriesBySession(0)
	for i := 1; i < len(entries); i++ {
		if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
			return fmt.Errorf("hors session: ordre chronologique incorrect")
		}
	}
	fmt.Print("   ✅ Ordre chronologique correct\n")

	// Validate: NextID is correct
	fmt.Println("✔️  Vérification du NextID...")
	maxID := 0
	for _, entry := range journal.Entries {
		if entry.ID > maxID {
			maxID = entry.ID
		}
	}
	if journal.NextID != maxID+1 {
		return fmt.Errorf("NextID incorrect: %d (devrait être %d)", journal.NextID, maxID+1)
	}
	fmt.Print("   ✅ NextID correct\n")

	// Report session distribution
	fmt.Println("📁 Distribution par session:")
	sessionCounts := make(map[int]int)
	for _, entry := range journal.Entries {
		sessionCounts[entry.SessionID]++
	}

	var sessionIDs []int
	for id := range sessionCounts {
		sessionIDs = append(sessionIDs, id)
	}
	sort.Ints(sessionIDs)

	for _, id := range sessionIDs {
		sessionName := fmt.Sprintf("Session %d", id)
		if id == 0 {
			sessionName = "Hors session"
		}
		fmt.Printf("   %s: %d entrées\n", sessionName, sessionCounts[id])
	}

	fmt.Println("\n✅ Validation réussie - journal intègre!")
	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// cmdCleanSession removes an invalid session and its journal entries.
func cmdCleanSession(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: clean-session <aventure> <session_id>")
	}

	adventureName := args[0]
	sessionID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("session_id invalide: %s", args[1])
	}

	if sessionID == 0 {
		return fmt.Errorf("impossible de supprimer la session 0 (hors session)")
	}

	// Load adventure
	adv, err := adventure.LoadByName(adventuresDir, adventureName)
	if err != nil {
		return fmt.Errorf("chargement aventure: %w", err)
	}

	fmt.Printf("🗑️  Suppression de la session %d de l'aventure '%s'\n\n", sessionID, adv.Name)

	// Check if session journal file exists
	sessionJournalPath := fmt.Sprintf("%s/journal-session-%d.json", adv.BasePath(), sessionID)
	if _, err := os.Stat(sessionJournalPath); os.IsNotExist(err) {
		return fmt.Errorf("journal de la session %d non trouvé: %s", sessionID, sessionJournalPath)
	}

	// Load session journal to show what will be deleted
	sessionEntries, err := adv.GetEntriesBySession(sessionID)
	if err != nil {
		return fmt.Errorf("chargement des entrées: %w", err)
	}

	fmt.Printf("📊 Entrées à supprimer: %d\n", len(sessionEntries))
	if len(sessionEntries) > 0 {
		fmt.Println("\nAperçu des entrées:")
		for i, entry := range sessionEntries {
			if i >= 5 {
				fmt.Printf("   ... et %d autres entrées\n", len(sessionEntries)-5)
				break
			}
			fmt.Printf("   - ID %d [%s]: %s\n", entry.ID, entry.Type, truncate(entry.Content, 60))
		}
	}

	// Confirmation prompt
	fmt.Printf("\n⚠️  Cette action va supprimer:\n")
	fmt.Printf("   - Le fichier journal-session-%d.json\n", sessionID)
	fmt.Printf("   - %d entrées du journal\n", len(sessionEntries))
	fmt.Printf("\nCette action est IRRÉVERSIBLE. Continuer? (oui/non): ")

	var response string
	fmt.Scanln(&response)
	if response != "oui" {
		fmt.Println("❌ Annulé")
		return nil
	}

	// Delete session journal file
	if err := os.Remove(sessionJournalPath); err != nil {
		return fmt.Errorf("suppression du fichier journal: %w", err)
	}
	fmt.Printf("✅ Fichier supprimé: journal-session-%d.json\n", sessionID)

	// Remove session from sessions.json
	sessionHistory, err := adv.LoadSessions()
	if err == nil {
		newSessions := []adventure.Session{}
		for _, s := range sessionHistory.Sessions {
			if s.ID != sessionID {
				newSessions = append(newSessions, s)
			}
		}
		sessionHistory.Sessions = newSessions
		if err := adv.SaveSessions(sessionHistory); err != nil {
			fmt.Printf("⚠️  Avertissement: impossible de mettre à jour sessions.json: %v\n", err)
		} else {
			fmt.Printf("✅ Session %d retirée de sessions.json\n", sessionID)
		}
	}

	// Validate remaining journal integrity
	fmt.Println("\n🔍 Validation de l'intégrité du journal restant...")
	journal, err := adv.LoadJournal()
	if err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	// Check for missing IDs
	if len(journal.Entries) > 0 {
		allIDs := make([]int, 0)
		for _, entry := range journal.Entries {
			allIDs = append(allIDs, entry.ID)
		}
		sort.Ints(allIDs)

		missingIDs := []int{}
		for i := allIDs[0]; i < allIDs[len(allIDs)-1]; i++ {
			found := false
			for _, id := range allIDs {
				if id == i {
					found = true
					break
				}
			}
			if !found {
				missingIDs = append(missingIDs, i)
			}
		}

		if len(missingIDs) > 0 {
			fmt.Printf("⚠️  IDs manquants détectés: %v\n", missingIDs)
			fmt.Println("   (Cela peut être normal si des entrées ont été supprimées)")
		}
	}

	fmt.Printf("\n✅ Session %d supprimée avec succès!\n", sessionID)
	fmt.Println("\n💡 Conseil: Exécutez 'sw-adventure validate-journal' pour vérifier l'intégrité complète")

	return nil
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// cmdInspectSessions analyzes all sessions to detect problems.
func cmdInspectSessions(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: inspect-sessions <aventure>")
	}

	// Load adventure
	adv, err := adventure.LoadByName(adventuresDir, args[0])
	if err != nil {
		return fmt.Errorf("chargement aventure: %w", err)
	}

	fmt.Printf("🔍 Inspection des sessions: %s\n\n", adv.Name)

	// Load all journal data
	journal, err := adv.LoadJournal()
	if err != nil {
		return fmt.Errorf("chargement journal: %w", err)
	}

	// Get all sessions
	sessions, err := adv.GetAllSessions()
	if err != nil {
		fmt.Printf("⚠️  Impossible de charger sessions.json: %v\n", err)
		sessions = []adventure.Session{}
	}

	// Build session statistics
	sessionStats := make(map[int]*SessionStats)

	// Initialize with session 0 (out-of-session)
	sessionStats[0] = &SessionStats{
		ID: 0,
		Name: "Hors session",
		EntryCount: 0,
		Types: make(map[string]int),
	}

	// Initialize from sessions.json
	for _, s := range sessions {
		sessionStats[s.ID] = &SessionStats{
			ID: s.ID,
			Name: fmt.Sprintf("Session %d", s.ID),
			StartedAt: s.StartedAt,
			EndedAt: s.EndedAt,
			Summary: s.Summary,
			EntryCount: 0,
			Types: make(map[string]int),
		}
	}

	// Count entries per session
	for _, entry := range journal.Entries {
		if stats, exists := sessionStats[entry.SessionID]; exists {
			stats.EntryCount++
			stats.Types[entry.Type]++

			// Track first and last entry
			if stats.FirstEntry.IsZero() || entry.Timestamp.Before(stats.FirstEntry) {
				stats.FirstEntry = entry.Timestamp
			}
			if entry.Timestamp.After(stats.LastEntry) {
				stats.LastEntry = entry.Timestamp
			}
		} else {
			// Session exists in journal but not in sessions.json
			sessionStats[entry.SessionID] = &SessionStats{
				ID: entry.SessionID,
				Name: fmt.Sprintf("Session %d", entry.SessionID),
				EntryCount: 1,
				Types: map[string]int{entry.Type: 1},
				FirstEntry: entry.Timestamp,
				LastEntry: entry.Timestamp,
				Orphaned: true,
			}
		}
	}

	// Check for missing IDs
	allIDs := make([]int, 0)
	for _, entry := range journal.Entries {
		allIDs = append(allIDs, entry.ID)
	}
	sort.Ints(allIDs)

	missingIDs := []int{}
	if len(allIDs) > 0 {
		for i := allIDs[0]; i < allIDs[len(allIDs)-1]; i++ {
			found := false
			for _, id := range allIDs {
				if id == i {
					found = true
					break
				}
			}
			if !found {
				missingIDs = append(missingIDs, i)
			}
		}
	}

	// Print report
	fmt.Println("\n📊 Rapport par session:")

	// Sort session IDs
	sessionIDs := make([]int, 0)
	for id := range sessionStats {
		sessionIDs = append(sessionIDs, id)
	}
	sort.Ints(sessionIDs)

	problemSessions := []int{}
	for _, id := range sessionIDs {
		stats := sessionStats[id]
		status := "✅ OK"
		issues := []string{}

		// Detect problems
		if stats.EntryCount == 0 {
			status = "⚠️  VIDE"
			issues = append(issues, "Aucune entrée")
			problemSessions = append(problemSessions, id)
		} else if stats.EntryCount == 1 && stats.Types["session"] == 1 {
			status = "⚠️  INCOMPLÈTE"
			issues = append(issues, "Seulement 'Session N démarrée'")
			problemSessions = append(problemSessions, id)
		} else if stats.EntryCount <= 3 {
			status = "⚠️  SUSPECTE"
			issues = append(issues, fmt.Sprintf("Seulement %d entrées", stats.EntryCount))
		}

		if stats.Orphaned {
			status = "❌ ORPHELINE"
			issues = append(issues, "Pas dans sessions.json")
			problemSessions = append(problemSessions, id)
		}

		// Check for test data
		testTypes := []string{"Potion de test", "Test verification", "Test foreshadow"}
		hasTestData := false
		for _, entry := range journal.Entries {
			if entry.SessionID == id {
				for _, testStr := range testTypes {
					if strings.Contains(entry.Content, testStr) {
						hasTestData = true
						break
					}
				}
			}
		}
		if hasTestData {
			status = "⚠️  TEST DATA"
			issues = append(issues, "Contient des données de test")
			if id != 0 { // Don't mark session 0 as problem for test data
				problemSessions = append(problemSessions, id)
			}
		}

		if !stats.EndedAt.IsZero() && stats.Summary == "" {
			issues = append(issues, "Pas de résumé")
		}

		fmt.Printf("%s %s:\n", status, stats.Name)
		fmt.Printf("   Entrées: %d", stats.EntryCount)
		if stats.EntryCount > 0 {
			typeList := ""
			for typ, count := range stats.Types {
				if typeList != "" {
					typeList += ", "
				}
				typeList += fmt.Sprintf("%s:%d", typ, count)
			}
			fmt.Printf(" (%s)", typeList)
		}
		fmt.Println()

		if !stats.StartedAt.IsZero() {
			fmt.Printf("   Démarrée: %s\n", stats.StartedAt.Format("2006-01-02 15:04"))
		}
		if !stats.EndedAt.IsZero() {
			fmt.Printf("   Terminée: %s\n", stats.EndedAt.Format("2006-01-02 15:04"))
		}
		if stats.Summary != "" {
			fmt.Printf("   Résumé: %s\n", truncate(stats.Summary, 60))
		}

		if len(issues) > 0 {
			fmt.Printf("   ⚠️  Problèmes: %s\n", strings.Join(issues, ", "))
		}
		fmt.Println()
	}

	// Report missing IDs
	if len(missingIDs) > 0 {
		fmt.Printf("⚠️  IDs manquants: %v\n\n", missingIDs)
	}

	// Summary
	fmt.Println("📋 Résumé:")
	fmt.Printf("   Sessions totales: %d\n", len(sessionStats))
	fmt.Printf("   Entrées totales: %d\n", len(journal.Entries))
	if len(problemSessions) > 0 {
		fmt.Printf("   ⚠️  Sessions problématiques: %v\n", problemSessions)
		fmt.Println("\n💡 Utilisez 'sw-adventure clean-session' pour supprimer une session invalide")
	} else {
		fmt.Println("   ✅ Aucun problème détecté")
	}

	return nil
}

// SessionStats holds statistics about a session.
type SessionStats struct {
	ID          int
	Name        string
	StartedAt   time.Time
	EndedAt     time.Time
	Summary     string
	EntryCount  int
	Types       map[string]int
	FirstEntry  time.Time
	LastEntry   time.Time
	Orphaned    bool // Session in journal but not in sessions.json
}
