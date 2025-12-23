package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"dungeons/internal/adventure"
)

// JournalMetadata mirrors the internal type for reading
type JournalMetadata struct {
	NextID     int       `json:"next_id"`
	Categories []string  `json:"categories"`
	LastUpdate time.Time `json:"last_update"`
}

func main() {
	// Load adventure
	adv, err := adventure.LoadByName("data/adventures", "la-crypte-des-ombres")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading adventure: %v\n", err)
		os.Exit(1)
	}

	// Read metadata directly to get next ID
	metaPath := filepath.Join(adv.BasePath(), "journal-meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading journal metadata: %v\n", err)
		os.Exit(1)
	}

	var meta JournalMetadata
	if err := json.Unmarshal(metaData, &meta); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing journal metadata: %v\n", err)
		os.Exit(1)
	}

	startID := meta.NextID
	fmt.Printf("Starting reconstruction from ID %d\n\n", startID)

	// Define all 16 lost events from sw-dm.log
	events := []struct {
		timestamp string
		typ       string
		content   string
	}{
		{"2025-12-23T16:11:19+01:00", "session", "SESSION 5 RÉSUMÉE - Grande Chambre Funéraire découverte. Dalle centrale avec glyphes, 3 passages (nord: flux énergie, est: inscriptions, ouest: scellé). Présence mystérieuse scellée détectée."},
		{"2025-12-23T16:11:19+01:00", "quest", "OBJECTIF EN COURS : Déterminer la nature de la créature scellée dans la dalle centrale de la Grande Chambre Funéraire."},
		{"2025-12-23T16:11:19+01:00", "quest", "SOUS-QUÊTES : (1) Déchiffrer les runes naines sur la dalle, (2) Explorer les 3 passages (nord/est/ouest), (3) Décider d'établir ou non un contact avec la créature."},
		{"2025-12-23T16:11:19+01:00", "note", "ÉTAT DU GROUPE fin session 5 : Aldric (6/8 HP), Lyra (5/5 HP, sorts utilisés), Thorin (6/7 HP, 1 sort niveau 1 restant), Adanel (6/6 HP), Gareth (5/8 HP)."},
		{"2025-12-23T16:11:19+01:00", "location", "POSITION ACTUELLE : Grande Chambre Funéraire Souterraine (Niveau 2 des catacombes de Sombregarde)."},
		{"2025-12-23T16:11:19+01:00", "note", "HOOKS POUR SESSION 6 : Kess mentionnée à Cordova (port). Groupe devra acheter équipement/sorts avant retour crypte."},
		{"2025-12-23T16:16:58+01:00", "npc", "RENCONTRE : Sirène, ancienne voleuse à la Taverne du Voile Écarlate. Connaît Kess depuis 10 ans (Guilde de l'Ombre). Attitude amicale mais prudente."},
		{"2025-12-23T16:21:12+01:00", "discovery", "INFORMATION : Kess aperçue récemment au port de Cordova, surveillant discrètement des cargaisons de contrebande."},
		{"2025-12-23T16:21:12+01:00", "npc", "RÉVÉLATION : Sirène et Kess ont travaillé ensemble dans la Guilde de l'Ombre. Kess a mystérieusement quitté Pierrebrune il y a quelques semaines."},
		{"2025-12-23T16:26:17+01:00", "story", "BACKSTORY : Sirène et Kess se connaissent depuis 10 ans. Ont fait un job ensemble il y a 7 ans qui a mal tourné. Kess a disparu après, Sirène a raccroché."},
		{"2025-12-23T16:26:17+01:00", "npc", "INFORMATION : Marta (aubergiste Voile Écarlate) a vu Kess recevoir un visiteur nocturne mystérieux il y a 3 semaines. Homme encapuchonné, conversation tendue."},
		{"2025-12-23T16:29:54+01:00", "discovery", "RÉVÉLATION MAJEURE : Kess s'est embarquée volontairement sur le navire 'Les Corbeaux des Mers' (capitaine Meren le Noir) en direction de Shasseth."},
		{"2025-12-23T16:36:36+01:00", "quest", "INFORMATION DE MEREN : Kess est partie de son plein gré à Shasseth. A payé 500 po pour la traversée. Semblait poursuivre quelque chose d'important."},
		{"2025-12-23T16:36:36+01:00", "note", "NÉGOCIATION VOYAGE : Meren demande 2000 po pour transporter le groupe à Shasseth. Groupe dispose actuellement de 1293 po. Besoin de 707 po supplémentaires."},
		{"2025-12-23T16:38:42+01:00", "quest", "NOUVEAU CONTRAT : Valorian le Doré (riche marchand) offre 600 po pour sauver sa fille Elara, enlevée par des bandits et retenue dans une grotte près de Cordova."},
		{"2025-12-23T16:42:55+01:00", "quest", "NÉGOCIATION RÉUSSIE : Valorian accepte 800 po total (700 po à la livraison + 100 po d'acompte immédiat). Le groupe récupère l'acompte de 100 po."},
	}

	// Reconstruct events
	for i, e := range events {
		ts, err := time.Parse(time.RFC3339, e.timestamp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing timestamp %s: %v\n", e.timestamp, err)
			continue
		}

		entry := adventure.JournalEntry{
			ID:        startID + i,
			Timestamp: ts,
			SessionID: 0, // Out of session
			Type:      e.typ,
			Content:   e.content,
		}

		if err := adv.SaveJournalEntry(entry); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error saving entry %d: %v\n", entry.ID, err)
		} else {
			fmt.Printf("✓ Saved entry %d [%s]: %s\n", entry.ID, e.typ, truncate(e.content, 60))
		}
	}

	// Update metadata NextID manually
	meta.NextID = startID + len(events)
	meta.LastUpdate = time.Now()

	updatedData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling metadata: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(metaPath, updatedData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Reconstructed %d events (IDs %d-%d)\n", len(events), startID, meta.NextID-1)

	// Fix inventory: add missing 100 po from Valorian's advance
	inv, err := adv.LoadInventory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading inventory: %v\n", err)
		os.Exit(1)
	}

	oldGold := inv.Gold
	inv.Gold += 100

	if err := adv.SaveInventory(inv); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving inventory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Fixed inventory: +100 po (was %d po, now %d po)\n", oldGold, inv.Gold)
	fmt.Println("\n✓ Journal reconstruction complete!")
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
