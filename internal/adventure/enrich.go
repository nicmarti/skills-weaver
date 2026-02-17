package adventure

import (
	"fmt"
)

// EnrichOptions configures enrichment behavior.
type EnrichOptions struct {
	SessionID int  // Filter by session
	RecentN   int  // Last N entries
	FromID    int  // Start entry ID
	ToID      int  // End entry ID
	Force     bool // Re-enrich existing descriptions
	DryRun    bool // Preview only
	BatchSize int  // Interactive batch size
}

// GetEntriesToEnrich returns entries needing enrichment based on options.
func (a *Adventure) GetEntriesToEnrich(opts EnrichOptions) ([]JournalEntry, error) {
	journal, err := a.LoadJournal()
	if err != nil {
		return nil, err
	}

	var entries []JournalEntry
	for _, entry := range journal.Entries {
		// Apply session filter
		if opts.SessionID > 0 && entry.SessionID != opts.SessionID {
			continue
		}

		// Apply ID range filters
		if opts.FromID > 0 && entry.ID < opts.FromID {
			continue
		}
		if opts.ToID > 0 && entry.ID > opts.ToID {
			continue
		}

		// Apply recent entries filter
		if opts.RecentN > 0 && entry.ID < (journal.NextID-opts.RecentN) {
			continue
		}

		// Skip if already has descriptions (unless --force)
		if !opts.Force && (entry.Description != "" || entry.DescriptionFr != "") {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// UpdateEntryDescriptions updates description fields for a single entry.
func (a *Adventure) UpdateEntryDescriptions(entryID int, description, descriptionFr string) error {
	// Load all journals to find the entry
	journal, err := a.LoadJournal()
	if err != nil {
		return err
	}

	// Find the entry and update it
	for i := range journal.Entries {
		if journal.Entries[i].ID == entryID {
			journal.Entries[i].Description = description
			journal.Entries[i].DescriptionFr = descriptionFr
			// Save to correct session file
			return a.SaveJournalEntry(journal.Entries[i])
		}
	}

	return fmt.Errorf("entry %d not found", entryID)
}

// EnrichmentContext provides context for AI enrichment.
type EnrichmentContext struct {
	AdventureName string
	PartyMembers  []string // ["Aldric (Human Fighter)", ...]
	RecentEntries []string // Last 5 entries
	SessionInfo   string
}

// GetEnrichmentContext builds context for an entry.
func (a *Adventure) GetEnrichmentContext(entry JournalEntry) (*EnrichmentContext, error) {
	ctx := &EnrichmentContext{
		AdventureName: a.Name,
	}

	// Get party composition
	characters, _ := a.GetCharacters()
	for _, c := range characters {
		ctx.PartyMembers = append(ctx.PartyMembers,
			fmt.Sprintf("%s (%s %s)", c.Name, c.Species, c.Class))
	}

	// Get recent entries (last 5 before this one)
	journal, _ := a.LoadJournal()
	count := 0
	for i := len(journal.Entries) - 1; i >= 0 && count < 5; i-- {
		e := journal.Entries[i]
		if e.ID < entry.ID {
			ctx.RecentEntries = append([]string{
				fmt.Sprintf("[%s] %s", e.Type, e.Content),
			}, ctx.RecentEntries...)
			count++
		}
	}

	// Get session info
	if entry.SessionID > 0 {
		session, _ := a.GetSession(entry.SessionID)
		if session != nil {
			ctx.SessionInfo = fmt.Sprintf("Session %d, started %s",
				session.ID, session.StartedAt.Format("2006-01-02"))
		}
	}

	return ctx, nil
}
