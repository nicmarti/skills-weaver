package adventure

import (
	"strings"
	"testing"
	"time"
)

func TestLoadJournalNonExistent(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	journal, err := adv.LoadJournal()
	if err != nil {
		t.Fatalf("LoadJournal() error = %v, want nil", err)
	}
	if len(journal.Entries) != 0 {
		t.Errorf("LoadJournal() entries = %v, want []", journal.Entries)
	}
	if journal.NextID != 1 {
		t.Errorf("LoadJournal() NextID = %d, want 1", journal.NextID)
	}
	if len(journal.Categories) == 0 {
		t.Errorf("LoadJournal() categories is empty, want default categories")
	}
}

func TestLogEvent(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	err := adv.LogEvent("combat", "Defeated 3 goblins")
	if err != nil {
		t.Fatalf("LogEvent() error = %v", err)
	}

	entries, _ := adv.GetJournalEntries()
	if len(entries) != 1 {
		t.Errorf("LogEvent() resulted in %d entries, want 1", len(entries))
	}

	entry := entries[0]
	if entry.Type != "combat" {
		t.Errorf("Entry type = %q, want 'combat'", entry.Type)
	}
	if entry.Content != "Defeated 3 goblins" {
		t.Errorf("Entry content = %q, want 'Defeated 3 goblins'", entry.Content)
	}
	if entry.ID != 1 {
		t.Errorf("Entry ID = %d, want 1", entry.ID)
	}
}

func TestLogEventWithDescriptions(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	err := adv.LogEventWithDescriptions(
		"combat",
		"Boss battle",
		"The party faces the dragon in epic combat",
		"Le groupe affronte le dragon dans un combat épique",
	)
	if err != nil {
		t.Fatalf("LogEventWithDescriptions() error = %v", err)
	}

	entries, _ := adv.GetJournalEntries()
	entry := entries[0]
	if entry.Description != "The party faces the dragon in epic combat" {
		t.Errorf("Description mismatch")
	}
	if entry.DescriptionFr != "Le groupe affronte le dragon dans un combat épique" {
		t.Errorf("DescriptionFr mismatch")
	}
}

func TestLogImportantEvent(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	tags := []string{"dragon", "boss", "victory"}
	err := adv.LogImportantEvent("combat", "Defeated the dragon lord", tags)
	if err != nil {
		t.Fatalf("LogImportantEvent() error = %v", err)
	}

	entries, _ := adv.GetJournalEntries()
	entry := entries[0]
	if !entry.Important {
		t.Errorf("Entry.Important = false, want true")
	}
	if len(entry.Tags) != len(tags) {
		t.Errorf("Entry tags count = %d, want %d", len(entry.Tags), len(tags))
	}
}

func TestGetEntriesByType(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Log different entry types
	adv.LogEvent("combat", "Fight 1")
	adv.LogEvent("loot", "Found gold")
	adv.LogEvent("combat", "Fight 2")
	adv.LogEvent("note", "Important note")
	adv.LogEvent("combat", "Fight 3")

	// Get combat entries
	entries, err := adv.GetEntriesByType("combat")
	if err != nil {
		t.Fatalf("GetEntriesByType() error = %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Combat entries count = %d, want 3", len(entries))
	}

	for _, e := range entries {
		if e.Type != "combat" {
			t.Errorf("Non-combat entry in combat filter: %q", e.Type)
		}
	}
}

func TestGetEntriesBySession(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Create session 1
	adv.StartSession()
	adv.LogEvent("combat", "Fight in session 1")
	adv.LogEvent("loot", "Loot in session 1")
	adv.EndSession("Session 1 summary")

	// Create session 2
	adv.StartSession()
	adv.LogEvent("combat", "Fight in session 2")
	adv.EndSession("Session 2 summary")

	entries, err := adv.GetEntriesBySession(1)
	if err != nil {
		t.Fatalf("GetEntriesBySession() error = %v", err)
	}

	// Should get 2 events (combat, loot) + 1 session marker
	if len(entries) != 3 {
		t.Errorf("Session 1 entries = %d, want 3", len(entries))
	}

	for _, e := range entries {
		if e.SessionID != 1 {
			t.Errorf("Non-session-1 entry: SessionID=%d", e.SessionID)
		}
	}
}

func TestGetImportantEntries(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.LogEvent("combat", "Normal fight")
	adv.LogImportantEvent("discovery", "Found ancient artifact", []string{"treasure"})
	adv.LogEvent("loot", "Regular loot")
	adv.LogImportantEvent("death", "Character died", []string{"tragedy"})

	entries, err := adv.GetImportantEntries()
	if err != nil {
		t.Fatalf("GetImportantEntries() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Important entries count = %d, want 2", len(entries))
	}

	for _, e := range entries {
		if !e.Important {
			t.Errorf("Non-important entry in important filter")
		}
	}
}

func TestGetRecentEntries(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Log 10 entries
	for i := 1; i <= 10; i++ {
		adv.LogEvent("note", "Entry "+string(rune(48+i)))
	}

	recent, err := adv.GetRecentEntries(3)
	if err != nil {
		t.Fatalf("GetRecentEntries() error = %v", err)
	}

	if len(recent) != 3 {
		t.Errorf("GetRecentEntries(3) = %d entries, want 3", len(recent))
	}

	// Last 3 entries should have highest IDs
	expectedIDs := []int{8, 9, 10}
	for i, e := range recent {
		if e.ID != expectedIDs[i] {
			t.Errorf("Recent entry[%d].ID = %d, want %d", i, e.ID, expectedIDs[i])
		}
	}
}

func TestGetRecentEntriesMoreThanAvailable(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.LogEvent("note", "Entry 1")
	adv.LogEvent("note", "Entry 2")

	recent, _ := adv.GetRecentEntries(10)
	if len(recent) != 2 {
		t.Errorf("GetRecentEntries(10) with 2 entries = %d, want 2", len(recent))
	}
}

func TestSearchEntries(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.LogEvent("combat", "Defeated the goblin king")
	adv.LogEvent("loot", "Found goblin gold")
	adv.LogEvent("note", "Goblins are weak to fire")
	adv.LogEvent("story", "Met the dragon")

	results, err := adv.SearchEntries("goblin")
	if err != nil {
		t.Fatalf("SearchEntries() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("SearchEntries('goblin') = %d results, want 3", len(results))
	}

	// Test case-insensitive
	results, _ = adv.SearchEntries("GOBLIN")
	if len(results) != 3 {
		t.Errorf("SearchEntries('GOBLIN') case-insensitive = %d, want 3", len(results))
	}
}

func TestSearchEntriesNoResults(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.LogEvent("note", "Some event")

	results, _ := adv.SearchEntries("dragon")
	if len(results) != 0 {
		t.Errorf("SearchEntries() with no matches = %d, want 0", len(results))
	}
}

func TestMarkEntryImportant(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.LogEvent("loot", "Found something")

	// Mark as important
	err := adv.MarkEntryImportant(1, true)
	if err != nil {
		t.Fatalf("MarkEntryImportant(true) error = %v", err)
	}

	entries, _ := adv.GetJournalEntries()
	if !entries[0].Important {
		t.Errorf("Entry not marked as important")
	}

	// Mark as not important
	adv.MarkEntryImportant(1, false)
	entries, _ = adv.GetJournalEntries()
	if entries[0].Important {
		t.Errorf("Entry should not be important after unmarking")
	}
}

func TestAddTagToEntry(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.LogEvent("combat", "Fight")

	err := adv.AddTagToEntry(1, "boss")
	if err != nil {
		t.Fatalf("AddTagToEntry() error = %v", err)
	}

	entries, _ := adv.GetJournalEntries()
	if len(entries[0].Tags) != 1 || entries[0].Tags[0] != "boss" {
		t.Errorf("Tag not added properly: %v", entries[0].Tags)
	}

	// Add another tag
	adv.AddTagToEntry(1, "important")
	entries, _ = adv.GetJournalEntries()
	if len(entries[0].Tags) != 2 {
		t.Errorf("Second tag not added: %v", entries[0].Tags)
	}
}

func TestAddTagDuplicate(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.LogEvent("note", "Entry")
	adv.AddTagToEntry(1, "tag1")
	adv.AddTagToEntry(1, "tag1") // Try to add same tag

	entries, _ := adv.GetJournalEntries()
	if len(entries[0].Tags) != 1 {
		t.Errorf("Duplicate tag was added: %v", entries[0].Tags)
	}
}

func TestDeleteEntry(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.LogEvent("note", "Entry 1")
	adv.LogEvent("note", "Entry 2")
	adv.LogEvent("note", "Entry 3")

	err := adv.DeleteEntry(2)
	if err != nil {
		t.Fatalf("DeleteEntry() error = %v", err)
	}

	entries, _ := adv.GetJournalEntries()
	if len(entries) != 2 {
		t.Errorf("DeleteEntry() resulted in %d entries, want 2", len(entries))
	}

	// Verify correct entries remain
	for _, e := range entries {
		if e.ID == 2 {
			t.Errorf("Deleted entry still exists")
		}
	}
}

func TestDeleteEntryNotFound(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	err := adv.DeleteEntry(999)
	if err == nil {
		t.Errorf("DeleteEntry() should fail for non-existent entry")
	}
}

func TestJournalToMarkdown(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Add entries
	adv.LogEvent("combat", "Fought goblins")
	adv.LogEvent("loot", "Found treasure")

	md, err := adv.JournalToMarkdown()
	if err != nil {
		t.Fatalf("JournalToMarkdown() error = %v", err)
	}

	if !strings.Contains(md, adv.Name) {
		t.Errorf("Markdown doesn't contain adventure name")
	}
	if !strings.Contains(md, "Fought goblins") {
		t.Errorf("Markdown doesn't contain entry content")
	}
	if !strings.Contains(md, "⚔️") { // Combat emoji
		t.Errorf("Markdown doesn't contain combat icon")
	}
}

func TestJournalToMarkdownEmpty(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	md, err := adv.JournalToMarkdown()
	if err != nil {
		t.Fatalf("JournalToMarkdown() error = %v", err)
	}

	if !strings.Contains(md, "Aucune entrée") {
		t.Errorf("Empty journal markdown should mention no entries")
	}
}

func TestJournalToMarkdownWithSessions(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Create session and entries
	adv.StartSession()
	adv.LogEvent("combat", "Fight in session")
	adv.EndSession("Done")

	// Add entry outside session
	adv.LogEvent("note", "Out-of-session note")

	md, _ := adv.JournalToMarkdown()

	if !strings.Contains(md, "Session 1") {
		t.Errorf("Markdown should contain session heading")
	}
	if !strings.Contains(md, "Hors session") {
		t.Errorf("Markdown should contain out-of-session section")
	}
}

func TestSessionSummaryMarkdown(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.StartSession()
	adv.UpdateSessionLocation(1, "Dragon's Lair")
	adv.LogEvent("combat", "Fought dragon")
	adv.AwardXP(1, 500)
	adv.RecordGoldFound(1, 1000)
	adv.EndSession("Victory!")

	md, err := adv.SessionSummaryMarkdown(1)
	if err != nil {
		t.Fatalf("SessionSummaryMarkdown() error = %v", err)
	}

	if !strings.Contains(md, "Session 1") {
		t.Errorf("Markdown should contain session number")
	}
	if !strings.Contains(md, "Dragon's Lair") {
		t.Errorf("Markdown should contain location")
	}
	if !strings.Contains(md, "500") {
		t.Errorf("Markdown should contain XP")
	}
	if !strings.Contains(md, "1000") {
		t.Errorf("Markdown should contain gold")
	}
}

func TestGetTypeIcon(t *testing.T) {
	tests := []struct {
		entryType string
		expected  string
	}{
		{"combat", "⚔️"},
		{"loot", "💰"},
		{"story", "📖"},
		{"note", "📝"},
		{"unknown", "•"},
	}

	for _, tt := range tests {
		icon := getTypeIcon(tt.entryType)
		if icon != tt.expected {
			t.Errorf("getTypeIcon(%q) = %q, want %q", tt.entryType, icon, tt.expected)
		}
	}
}

func TestJournalEntrySerialization(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test", "Test")
	adv.SetBasePath(baseDir)

	// Log entry with all fields
	adv.LogEventWithDescriptions(
		"combat",
		"Test content",
		"Test description",
		"Description FR",
	)
	adv.MarkEntryImportant(1, true)
	adv.AddTagToEntry(1, "boss")

	// Reload and verify
	entries, _ := adv.GetJournalEntries()
	e := entries[0]

	if e.Content != "Test content" {
		t.Errorf("Content mismatch after serialization")
	}
	if e.Description != "Test description" {
		t.Errorf("Description mismatch after serialization")
	}
	if e.DescriptionFr != "Description FR" {
		t.Errorf("DescriptionFr mismatch after serialization")
	}
	if !e.Important {
		t.Errorf("Important flag not preserved")
	}
	if len(e.Tags) != 1 {
		t.Errorf("Tags not preserved")
	}
}

func TestJournalNextID(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test", "Test")
	adv.SetBasePath(baseDir)

	for i := 1; i <= 5; i++ {
		adv.LogEvent("note", "Entry")
	}

	entries, _ := adv.GetJournalEntries()
	if entries[0].ID != 1 || entries[4].ID != 5 {
		t.Errorf("Journal IDs not sequential")
	}

	// Verify NextID for new entry
	journal, _ := adv.LoadJournal()
	if journal.NextID != 6 {
		t.Errorf("NextID = %d, want 6", journal.NextID)
	}
}

func TestSessionEntryTimestamp(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test", "Test")
	adv.SetBasePath(baseDir)

	before := time.Now()
	adv.LogEvent("note", "Test")
	after := time.Now()

	entries, _ := adv.GetJournalEntries()
	ts := entries[0].Timestamp

	if ts.Before(before) || ts.After(after) {
		t.Errorf("Entry timestamp not within expected range")
	}
}
