package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
			fmt.Println("⊘ Skipped batch\n")
		case "q", "quit":
			fmt.Printf("\n✓ Enrichment stopped. %d entries enriched.\n", successCount)
			return nil
		default:
			fmt.Println("⊘ Invalid choice, skipping batch\n")
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
