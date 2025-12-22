package adventure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// JournalEntry represents a single entry in the adventure journal.
type JournalEntry struct {
	ID            int       `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	SessionID     int       `json:"session_id,omitempty"` // 0 if outside session
	Type          string    `json:"type"`                 // combat, loot, story, note, quest, etc.
	Content       string    `json:"content"`
	Description   string    `json:"description,omitempty"`    // Detailed English description for image generation
	DescriptionFr string    `json:"description_fr,omitempty"` // Detailed French description for reading
	Tags          []string  `json:"tags,omitempty"`
	Important     bool      `json:"important,omitempty"`
}

// Journal holds all journal entries for an adventure.
type Journal struct {
	Entries    []JournalEntry `json:"entries"`
	NextID     int            `json:"next_id"`
	Categories []string       `json:"categories"` // Available entry types
}

// Default journal categories
var defaultCategories = []string{
	"combat",     // Combat encounters
	"loot",       // Treasure and items found
	"story",      // Story progression
	"note",       // General notes
	"quest",      // Quest updates
	"npc",        // NPC interactions
	"location",   // Location discoveries
	"rest",       // Resting and recovery
	"death",      // Character death
	"levelup",    // Level advancement
	"session",    // Session markers
	"party",      // Party changes
	"xp",         // XP awards
	"expense",    // Gold spent
	"use",        // Item usage
}

// LoadJournal loads the adventure journal.
func (a *Adventure) LoadJournal() (*Journal, error) {
	path := filepath.Join(a.basePath, "journal.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Journal{
			Entries:    []JournalEntry{},
			NextID:     1,
			Categories: defaultCategories,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading journal.json: %w", err)
	}

	var journal Journal
	if err := json.Unmarshal(data, &journal); err != nil {
		return nil, fmt.Errorf("parsing journal.json: %w", err)
	}

	return &journal, nil
}

// SaveJournal saves the adventure journal.
func (a *Adventure) SaveJournal(journal *Journal) error {
	path := filepath.Join(a.basePath, "journal.json")
	return a.saveJSON(path, journal)
}

// LogEvent adds an entry to the journal.
func (a *Adventure) LogEvent(entryType, content string) error {
	journal, err := a.LoadJournal()
	if err != nil {
		return err
	}

	// Get current session ID if any
	sessionID := 0
	if session, _ := a.GetCurrentSession(); session != nil {
		sessionID = session.ID
	}

	entry := JournalEntry{
		ID:        journal.NextID,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Type:      entryType,
		Content:   content,
	}

	journal.Entries = append(journal.Entries, entry)
	journal.NextID++

	return a.SaveJournal(journal)
}

// LogEventWithDescriptions adds an entry with bilingual descriptions to the journal.
func (a *Adventure) LogEventWithDescriptions(eventType, content, description, descriptionFr string) error {
	journal, err := a.LoadJournal()
	if err != nil {
		return err
	}

	// Get current session ID if any
	sessionID := 0
	if session, _ := a.GetCurrentSession(); session != nil {
		sessionID = session.ID
	}

	entry := JournalEntry{
		ID:            journal.NextID,
		Timestamp:     time.Now(),
		SessionID:     sessionID,
		Type:          eventType,
		Content:       content,
		Description:   description,
		DescriptionFr: descriptionFr,
	}

	journal.Entries = append(journal.Entries, entry)
	journal.NextID++

	return a.SaveJournal(journal)
}

// LogImportantEvent adds an important entry to the journal.
func (a *Adventure) LogImportantEvent(entryType, content string, tags []string) error {
	journal, err := a.LoadJournal()
	if err != nil {
		return err
	}

	sessionID := 0
	if session, _ := a.GetCurrentSession(); session != nil {
		sessionID = session.ID
	}

	entry := JournalEntry{
		ID:        journal.NextID,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Type:      entryType,
		Content:   content,
		Tags:      tags,
		Important: true,
	}

	journal.Entries = append(journal.Entries, entry)
	journal.NextID++

	return a.SaveJournal(journal)
}

// GetJournalEntries returns all journal entries.
func (a *Adventure) GetJournalEntries() ([]JournalEntry, error) {
	journal, err := a.LoadJournal()
	if err != nil {
		return nil, err
	}

	// Sort by timestamp (oldest first)
	sort.Slice(journal.Entries, func(i, j int) bool {
		return journal.Entries[i].Timestamp.Before(journal.Entries[j].Timestamp)
	})

	return journal.Entries, nil
}

// GetEntriesByType returns entries filtered by type.
func (a *Adventure) GetEntriesByType(entryType string) ([]JournalEntry, error) {
	entries, err := a.GetJournalEntries()
	if err != nil {
		return nil, err
	}

	var filtered []JournalEntry
	for _, e := range entries {
		if e.Type == entryType {
			filtered = append(filtered, e)
		}
	}

	return filtered, nil
}

// GetEntriesBySession returns entries for a specific session.
func (a *Adventure) GetEntriesBySession(sessionID int) ([]JournalEntry, error) {
	entries, err := a.GetJournalEntries()
	if err != nil {
		return nil, err
	}

	var filtered []JournalEntry
	for _, e := range entries {
		if e.SessionID == sessionID {
			filtered = append(filtered, e)
		}
	}

	return filtered, nil
}

// GetImportantEntries returns only important entries.
func (a *Adventure) GetImportantEntries() ([]JournalEntry, error) {
	entries, err := a.GetJournalEntries()
	if err != nil {
		return nil, err
	}

	var filtered []JournalEntry
	for _, e := range entries {
		if e.Important {
			filtered = append(filtered, e)
		}
	}

	return filtered, nil
}

// GetRecentEntries returns the last N entries.
func (a *Adventure) GetRecentEntries(n int) ([]JournalEntry, error) {
	entries, err := a.GetJournalEntries()
	if err != nil {
		return nil, err
	}

	if len(entries) <= n {
		return entries, nil
	}

	return entries[len(entries)-n:], nil
}

// SearchEntries searches entries by content.
func (a *Adventure) SearchEntries(query string) ([]JournalEntry, error) {
	entries, err := a.GetJournalEntries()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []JournalEntry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Content), query) {
			results = append(results, e)
		}
	}

	return results, nil
}

// MarkEntryImportant marks an entry as important.
func (a *Adventure) MarkEntryImportant(entryID int, important bool) error {
	journal, err := a.LoadJournal()
	if err != nil {
		return err
	}

	for i := range journal.Entries {
		if journal.Entries[i].ID == entryID {
			journal.Entries[i].Important = important
			return a.SaveJournal(journal)
		}
	}

	return fmt.Errorf("entry %d not found", entryID)
}

// AddTagToEntry adds a tag to an entry.
func (a *Adventure) AddTagToEntry(entryID int, tag string) error {
	journal, err := a.LoadJournal()
	if err != nil {
		return err
	}

	for i := range journal.Entries {
		if journal.Entries[i].ID == entryID {
			// Check if tag already exists
			for _, t := range journal.Entries[i].Tags {
				if t == tag {
					return nil // Already has tag
				}
			}
			journal.Entries[i].Tags = append(journal.Entries[i].Tags, tag)
			return a.SaveJournal(journal)
		}
	}

	return fmt.Errorf("entry %d not found", entryID)
}

// DeleteEntry removes an entry from the journal.
func (a *Adventure) DeleteEntry(entryID int) error {
	journal, err := a.LoadJournal()
	if err != nil {
		return err
	}

	for i := range journal.Entries {
		if journal.Entries[i].ID == entryID {
			journal.Entries = append(journal.Entries[:i], journal.Entries[i+1:]...)
			return a.SaveJournal(journal)
		}
	}

	return fmt.Errorf("entry %d not found", entryID)
}

// JournalToMarkdown generates a markdown summary of the journal.
func (a *Adventure) JournalToMarkdown() (string, error) {
	entries, err := a.GetJournalEntries()
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "# Journal de l'Aventure\n\n*Aucune entrée pour le moment.*\n", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Journal de l'Aventure : %s\n\n", a.Name))

	// Group by session
	sessions := make(map[int][]JournalEntry)
	var noSession []JournalEntry

	for _, e := range entries {
		if e.SessionID > 0 {
			sessions[e.SessionID] = append(sessions[e.SessionID], e)
		} else {
			noSession = append(noSession, e)
		}
	}

	// Write entries outside sessions first
	if len(noSession) > 0 {
		sb.WriteString("## Hors session\n\n")
		for _, e := range noSession {
			writeEntry(&sb, e)
		}
		sb.WriteString("\n")
	}

	// Get sorted session IDs
	var sessionIDs []int
	for id := range sessions {
		sessionIDs = append(sessionIDs, id)
	}
	sort.Ints(sessionIDs)

	// Write each session
	for _, sessionID := range sessionIDs {
		sb.WriteString(fmt.Sprintf("## Session %d\n\n", sessionID))
		for _, e := range sessions[sessionID] {
			writeEntry(&sb, e)
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// SessionSummaryMarkdown generates a summary for a specific session.
func (a *Adventure) SessionSummaryMarkdown(sessionID int) (string, error) {
	session, err := a.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	entries, err := a.GetEntriesBySession(sessionID)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Session %d\n\n", sessionID))
	sb.WriteString(fmt.Sprintf("**Date** : %s\n", session.StartedAt.Format("02/01/2006 15:04")))

	if session.Status == "completed" {
		sb.WriteString(fmt.Sprintf("**Durée** : %s\n", session.Duration))
	} else {
		sb.WriteString("**Statut** : En cours\n")
	}

	if session.Location != "" {
		sb.WriteString(fmt.Sprintf("**Lieu** : %s\n", session.Location))
	}

	sb.WriteString("\n## Événements\n\n")

	if len(entries) == 0 {
		sb.WriteString("*Aucun événement enregistré.*\n")
	} else {
		for _, e := range entries {
			writeEntry(&sb, e)
		}
	}

	if session.Summary != "" {
		sb.WriteString(fmt.Sprintf("\n## Résumé\n\n%s\n", session.Summary))
	}

	if session.XPAwarded > 0 {
		sb.WriteString(fmt.Sprintf("\n**XP attribués** : %d\n", session.XPAwarded))
	}

	if session.GoldFound > 0 {
		sb.WriteString(fmt.Sprintf("**Or trouvé** : %d po\n", session.GoldFound))
	}

	return sb.String(), nil
}

// Helper to write a single entry
func writeEntry(sb *strings.Builder, e JournalEntry) {
	icon := getTypeIcon(e.Type)
	timestamp := e.Timestamp.Format("15:04")

	marker := ""
	if e.Important {
		marker = " ⭐"
	}

	sb.WriteString(fmt.Sprintf("- `%s` %s %s%s\n", timestamp, icon, e.Content, marker))
}

// getTypeIcon returns an icon for the entry type
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
