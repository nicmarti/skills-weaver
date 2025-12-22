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

// JournalMetadata tracks global journal state across all sessions.
type JournalMetadata struct {
	NextID     int       `json:"next_id"`      // Global ID counter
	Categories []string  `json:"categories"`   // Available entry types
	LastUpdate time.Time `json:"last_update"` // Last modification time
}

// SessionJournal holds entries for a specific session.
type SessionJournal struct {
	SessionID int            `json:"session_id"` // 0 for out-of-session
	Entries   []JournalEntry `json:"entries"`
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

// LoadJournal loads the adventure journal by aggregating all session journals.
func (a *Adventure) LoadJournal() (*Journal, error) {
	// Try new format first (session-based journals)
	meta, err := a.loadJournalMetadata()
	if err != nil {
		return nil, err
	}

	// Find all journal-session-*.json files
	pattern := filepath.Join(a.basePath, "journal-session-*.json")
	sessionFiles, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("finding session journals: %w", err)
	}

	// If no session files exist, check for legacy journal.json
	if len(sessionFiles) == 0 {
		legacyPath := filepath.Join(a.basePath, "journal.json")
		if _, err := os.Stat(legacyPath); err == nil {
			// Legacy journal.json exists, load it
			return a.loadLegacyJournal()
		}
		// No journals at all, return empty
		return &Journal{
			Entries:    []JournalEntry{},
			NextID:     meta.NextID,
			Categories: meta.Categories,
		}, nil
	}

	// Load and aggregate entries from all session journals
	var allEntries []JournalEntry
	for _, file := range sessionFiles {
		// Extract session ID from filename
		var sessionID int
		_, err := fmt.Sscanf(filepath.Base(file), "journal-session-%d.json", &sessionID)
		if err != nil {
			continue // Skip malformed filenames
		}

		sessionJournal, err := a.loadSessionJournal(sessionID)
		if err != nil {
			return nil, fmt.Errorf("loading session %d: %w", sessionID, err)
		}
		allEntries = append(allEntries, sessionJournal.Entries...)
	}

	// Sort by timestamp (oldest first)
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Timestamp.Before(allEntries[j].Timestamp)
	})

	return &Journal{
		Entries:    allEntries,
		NextID:     meta.NextID,
		Categories: meta.Categories,
	}, nil
}

// loadLegacyJournal loads the old monolithic journal.json format.
// This provides backward compatibility during migration.
func (a *Adventure) loadLegacyJournal() (*Journal, error) {
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

// loadJournalMetadata loads the journal metadata file.
func (a *Adventure) loadJournalMetadata() (*JournalMetadata, error) {
	path := filepath.Join(a.basePath, "journal-meta.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Initialize default metadata
		return &JournalMetadata{
			NextID:     1,
			Categories: defaultCategories,
			LastUpdate: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading journal-meta.json: %w", err)
	}

	var meta JournalMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parsing journal-meta.json: %w", err)
	}

	return &meta, nil
}

// SaveJournalMetadata saves the journal metadata file.
func (a *Adventure) SaveJournalMetadata(meta *JournalMetadata) error {
	meta.LastUpdate = time.Now()
	path := filepath.Join(a.basePath, "journal-meta.json")
	return a.saveJSON(path, meta)
}

// loadSessionJournal loads the journal for a specific session.
func (a *Adventure) loadSessionJournal(sessionID int) (*SessionJournal, error) {
	path := filepath.Join(a.basePath, fmt.Sprintf("journal-session-%d.json", sessionID))

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Create new session journal
		return &SessionJournal{
			SessionID: sessionID,
			Entries:   []JournalEntry{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading journal-session-%d.json: %w", sessionID, err)
	}

	var sj SessionJournal
	if err := json.Unmarshal(data, &sj); err != nil {
		return nil, fmt.Errorf("parsing journal-session-%d.json: %w", sessionID, err)
	}

	return &sj, nil
}

// SaveSessionJournal saves the journal for a specific session.
func (a *Adventure) SaveSessionJournal(sj *SessionJournal) error {
	path := filepath.Join(a.basePath, fmt.Sprintf("journal-session-%d.json", sj.SessionID))
	return a.saveJSON(path, sj)
}

// SaveJournalEntry saves a single journal entry to the appropriate session file.
func (a *Adventure) SaveJournalEntry(entry JournalEntry) error {
	// Load session journal for this entry's session
	sessionJournal, err := a.loadSessionJournal(entry.SessionID)
	if err != nil {
		return fmt.Errorf("loading session journal: %w", err)
	}

	// Check if entry already exists (update) or needs to be added (insert)
	found := false
	for i := range sessionJournal.Entries {
		if sessionJournal.Entries[i].ID == entry.ID {
			// Update existing entry
			sessionJournal.Entries[i] = entry
			found = true
			break
		}
	}

	if !found {
		// Add new entry
		sessionJournal.Entries = append(sessionJournal.Entries, entry)
	}

	// Sort entries by timestamp
	sort.Slice(sessionJournal.Entries, func(i, j int) bool {
		return sessionJournal.Entries[i].Timestamp.Before(sessionJournal.Entries[j].Timestamp)
	})

	// Save session journal
	return a.SaveSessionJournal(sessionJournal)
}

// LogEvent adds an entry to the journal.
func (a *Adventure) LogEvent(entryType, content string) error {
	// Load metadata for NextID
	meta, err := a.loadJournalMetadata()
	if err != nil {
		return err
	}

	// Get current session ID if any
	sessionID := 0
	if session, _ := a.GetCurrentSession(); session != nil {
		sessionID = session.ID
	}

	// Create entry
	entry := JournalEntry{
		ID:        meta.NextID,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Type:      entryType,
		Content:   content,
	}

	// Increment NextID and save metadata
	meta.NextID++
	if err := a.SaveJournalMetadata(meta); err != nil {
		return fmt.Errorf("saving metadata: %w", err)
	}

	// Save entry to session-specific file
	return a.SaveJournalEntry(entry)
}

// LogEventWithDescriptions adds an entry with bilingual descriptions to the journal.
func (a *Adventure) LogEventWithDescriptions(eventType, content, description, descriptionFr string) error {
	// Load metadata for NextID
	meta, err := a.loadJournalMetadata()
	if err != nil {
		return err
	}

	// Get current session ID if any
	sessionID := 0
	if session, _ := a.GetCurrentSession(); session != nil {
		sessionID = session.ID
	}

	// Create entry
	entry := JournalEntry{
		ID:            meta.NextID,
		Timestamp:     time.Now(),
		SessionID:     sessionID,
		Type:          eventType,
		Content:       content,
		Description:   description,
		DescriptionFr: descriptionFr,
	}

	// Increment NextID and save metadata
	meta.NextID++
	if err := a.SaveJournalMetadata(meta); err != nil {
		return fmt.Errorf("saving metadata: %w", err)
	}

	// Save entry to session-specific file
	return a.SaveJournalEntry(entry)
}

// LogImportantEvent adds an important entry to the journal.
func (a *Adventure) LogImportantEvent(entryType, content string, tags []string) error {
	// Load metadata for NextID
	meta, err := a.loadJournalMetadata()
	if err != nil {
		return err
	}

	// Get current session ID if any
	sessionID := 0
	if session, _ := a.GetCurrentSession(); session != nil {
		sessionID = session.ID
	}

	// Create entry
	entry := JournalEntry{
		ID:        meta.NextID,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Type:      entryType,
		Content:   content,
		Tags:      tags,
		Important: true,
	}

	// Increment NextID and save metadata
	meta.NextID++
	if err := a.SaveJournalMetadata(meta); err != nil {
		return fmt.Errorf("saving metadata: %w", err)
	}

	// Save entry to session-specific file
	return a.SaveJournalEntry(entry)
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
// Optimized to load only the specific session journal file.
func (a *Adventure) GetEntriesBySession(sessionID int) ([]JournalEntry, error) {
	// Load only the specific session journal
	sessionJournal, err := a.loadSessionJournal(sessionID)
	if err != nil {
		return nil, err
	}

	// Entries are already sorted by timestamp when saved
	return sessionJournal.Entries, nil
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
			return a.SaveJournalEntry(journal.Entries[i])
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
			return a.SaveJournalEntry(journal.Entries[i])
		}
	}

	return fmt.Errorf("entry %d not found", entryID)
}

// DeleteEntry removes an entry from the journal.
func (a *Adventure) DeleteEntry(entryID int) error {
	// Load all journals to find the entry and determine its session
	journal, err := a.LoadJournal()
	if err != nil {
		return err
	}

	// Find the entry and determine which session it belongs to
	var sessionID int
	var found bool
	for i := range journal.Entries {
		if journal.Entries[i].ID == entryID {
			sessionID = journal.Entries[i].SessionID
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("entry %d not found", entryID)
	}

	// Load the session journal
	sessionJournal, err := a.loadSessionJournal(sessionID)
	if err != nil {
		return err
	}

	// Remove the entry from the session
	for i := range sessionJournal.Entries {
		if sessionJournal.Entries[i].ID == entryID {
			sessionJournal.Entries = append(sessionJournal.Entries[:i], sessionJournal.Entries[i+1:]...)
			break
		}
	}

	// Save the session journal
	return a.SaveSessionJournal(sessionJournal)
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
