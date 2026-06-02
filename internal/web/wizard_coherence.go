package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"dungeons/internal/adventure"
	"dungeons/internal/npcmanager"
)

// WizardCoherence captures lightweight context from recent adventures so a newly
// generated adventure stays loosely consistent with the living world WITHOUT
// continuing old plots or reusing established NPC names.
type WizardCoherence struct {
	RecentAdventures []RecentAdventure
	NPCNamesToAvoid  []string
}

// RecentAdventure is a trimmed view of a recently played adventure.
type RecentAdventure struct {
	Name        string
	Description string
}

// buildWizardCoherence reads the 1-2 most recently played active adventures and
// gathers their names/descriptions plus the union of NPC names to avoid (their
// generated NPCs + the world's recurring NPCs). All operations are local file
// reads, so this is cheap to call at creation time.
func (s *Server) buildWizardCoherence() WizardCoherence {
	var ctx WizardCoherence

	advs, err := adventure.ListAdventures(adventuresDir)
	if err != nil {
		return ctx
	}

	active := make([]*adventure.Adventure, 0, len(advs))
	for _, a := range advs {
		if a.Status == adventure.StatusActive {
			active = append(active, a)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].LastPlayed.After(active[j].LastPlayed)
	})
	if len(active) > 2 {
		active = active[:2]
	}

	seen := map[string]bool{}
	addName := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		key := strings.ToLower(name)
		if !seen[key] {
			seen[key] = true
			ctx.NPCNamesToAvoid = append(ctx.NPCNamesToAvoid, name)
		}
	}

	for _, a := range active {
		ctx.RecentAdventures = append(ctx.RecentAdventures, RecentAdventure{
			Name:        a.Name,
			Description: a.Description,
		})

		mgr := npcmanager.NewManager(filepath.Join(adventuresDir, a.Slug))
		db, err := mgr.Load()
		if err != nil {
			continue
		}
		for _, records := range db.Sessions {
			for _, rec := range records {
				if rec.NPC != nil {
					addName(rec.NPC.Name)
				}
			}
		}
	}

	for _, name := range loadWorldNPCNames() {
		addName(name)
	}

	return ctx
}

// loadWorldNPCNames reads the recurring NPC names from data/world/npcs.json.
// It returns nil on any error (coherence is best-effort).
func loadWorldNPCNames() []string {
	data, err := os.ReadFile(filepath.Join("data", "world", "npcs.json"))
	if err != nil {
		return nil
	}
	var doc struct {
		RecurringNPCs []struct {
			Name string `json:"name"`
		} `json:"recurring_npcs"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil
	}
	names := make([]string, 0, len(doc.RecurringNPCs))
	for _, n := range doc.RecurringNPCs {
		names = append(names, n.Name)
	}
	return names
}

// HasContext reports whether there is any coherence information to inject.
func (c WizardCoherence) HasContext() bool {
	return len(c.RecentAdventures) > 0 || len(c.NPCNamesToAvoid) > 0
}

// PromptSection renders the coherence guidance for the campaign-plan prompt.
func (c WizardCoherence) PromptSection() string {
	if !c.HasContext() {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Cohérence avec le monde\n")
	if len(c.RecentAdventures) > 0 {
		sb.WriteString("Histoires récentes dans ce même monde (reste vaguement cohérent en ton et en géographie, mais NE PAS poursuivre ni réutiliser leurs intrigues) :\n")
		for _, a := range c.RecentAdventures {
			desc := a.Description
			// Rune-aware truncation: descriptions are French (accented multi-byte
			// runes), so slicing by bytes could split a rune and emit invalid UTF-8.
			if r := []rune(desc); len(r) > 240 {
				desc = string(r[:240]) + "…"
			}
			fmt.Fprintf(&sb, "- « %s » : %s\n", a.Name, desc)
		}
	}
	if len(c.NPCNamesToAvoid) > 0 {
		fmt.Fprintf(&sb, "NE PAS réutiliser ces noms de PNJ existants ; invente de nouveaux personnages : %s\n",
			strings.Join(c.NPCNamesToAvoid, ", "))
	}
	return sb.String()
}
