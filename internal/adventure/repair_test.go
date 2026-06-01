package adventure

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeRawEmptyCategoriesMeta writes a journal-meta.json with empty categories
// directly, bypassing SaveJournalMetadata's normalization, to simulate a legacy
// corrupted file on disk.
func writeRawEmptyCategoriesMeta(t *testing.T, adv *Adventure, nextID int) {
	t.Helper()
	raw := fmt.Sprintf(`{"next_id":%d,"categories":[],"last_update":"2026-01-01T00:00:00Z"}`, nextID)
	if err := os.WriteFile(filepath.Join(adv.BasePath(), "journal-meta.json"), []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
}

// newRepairAdv builds an adventure rooted in a temp dir for repair tests.
func newRepairAdv(t *testing.T) *Adventure {
	t.Helper()
	adv := New("Repair Test", "Test")
	adv.SetBasePath(t.TempDir())
	return adv
}

func ts(min int) time.Time {
	return time.Date(2026, 2, 14, 18, 0, 0, 0, time.UTC).Add(time.Duration(min) * time.Minute)
}

// hasAction reports whether the report contains an action of the given kind.
func hasAction(r *RepairReport, kind string) bool {
	for _, a := range r.Actions {
		if a.Kind == kind {
			return true
		}
	}
	return false
}

func hasAnomaly(r *RepairReport, kind string) bool {
	for _, a := range r.Anomalies {
		if a.Kind == kind {
			return true
		}
	}
	return false
}

func TestRepairHealthyJournalIsIdempotent(t *testing.T) {
	adv := newRepairAdv(t)
	if err := adv.SaveSessionJournal(&SessionJournal{SessionID: 1, Entries: []JournalEntry{
		{ID: 1, SessionID: 1, Type: "combat", Content: "a", Timestamp: ts(0)},
		{ID: 2, SessionID: 1, Type: "loot", Content: "b", Timestamp: ts(5)},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := adv.SaveJournalMetadata(&JournalMetadata{NextID: 3, Categories: defaultCategories}); err != nil {
		t.Fatal(err)
	}

	report, err := adv.RepairJournal(RepairOptions{DryRun: true})
	if err != nil {
		t.Fatalf("RepairJournal() error = %v", err)
	}
	if report.Changed() {
		t.Errorf("healthy journal should need no fixes, got actions: %+v", report.Actions)
	}
	if report.TotalEntries != 2 {
		t.Errorf("TotalEntries = %d, want 2", report.TotalEntries)
	}
}

func TestRepairRepopulatesEmptyCategories(t *testing.T) {
	adv := newRepairAdv(t)
	_ = adv.SaveSessionJournal(&SessionJournal{SessionID: 1, Entries: []JournalEntry{
		{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)},
	}})
	// Empty categories, correct next_id.
	writeRawEmptyCategoriesMeta(t, adv, 2)

	report, err := adv.RepairJournal(RepairOptions{DryRun: false})
	if err != nil {
		t.Fatalf("RepairJournal() error = %v", err)
	}
	if !hasAction(report, "categories") {
		t.Fatalf("expected categories action, got %+v", report.Actions)
	}
	meta, _ := adv.loadJournalMetadata()
	if len(meta.Categories) != len(defaultCategories) {
		t.Errorf("categories not repopulated: got %d, want %d", len(meta.Categories), len(defaultCategories))
	}
}

func TestRepairResyncsNextID(t *testing.T) {
	adv := newRepairAdv(t)
	_ = adv.SaveSessionJournal(&SessionJournal{SessionID: 1, Entries: []JournalEntry{
		{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)},
		{ID: 7, SessionID: 1, Type: "story", Content: "b", Timestamp: ts(5)},
	}})
	// next_id desynced (says 2, but max id is 7 -> want 8).
	_ = adv.SaveJournalMetadata(&JournalMetadata{NextID: 2, Categories: defaultCategories})

	report, err := adv.RepairJournal(RepairOptions{DryRun: false})
	if err != nil {
		t.Fatalf("RepairJournal() error = %v", err)
	}
	if !hasAction(report, "next_id") {
		t.Fatalf("expected next_id action, got %+v", report.Actions)
	}
	meta, _ := adv.loadJournalMetadata()
	if meta.NextID != 8 {
		t.Errorf("NextID = %d, want 8", meta.NextID)
	}
}

func TestRepairReHomesMisfiledEntry(t *testing.T) {
	adv := newRepairAdv(t)
	// An entry whose session_id=2 is wrongly stored in journal-session-1.json.
	_ = adv.SaveSessionJournal(&SessionJournal{SessionID: 1, Entries: []JournalEntry{
		{ID: 1, SessionID: 1, Type: "story", Content: "stays", Timestamp: ts(0)},
		{ID: 2, SessionID: 2, Type: "story", Content: "misfiled", Timestamp: ts(5)},
	}})
	_ = adv.SaveJournalMetadata(&JournalMetadata{NextID: 3, Categories: defaultCategories})

	report, err := adv.RepairJournal(RepairOptions{DryRun: false})
	if err != nil {
		t.Fatalf("RepairJournal() error = %v", err)
	}
	if !hasAction(report, "rehome") {
		t.Fatalf("expected rehome action, got %+v", report.Actions)
	}

	s1, _ := adv.GetEntriesBySession(1)
	if len(s1) != 1 || s1[0].ID != 1 {
		t.Errorf("session 1 should keep only id=1, got %+v", s1)
	}
	s2, _ := adv.GetEntriesBySession(2)
	if len(s2) != 1 || s2[0].ID != 2 {
		t.Errorf("session 2 should receive id=2, got %+v", s2)
	}
}

func TestRepairSortsChronologically(t *testing.T) {
	adv := newRepairAdv(t)
	// Entries stored out of timestamp order within the file.
	_ = adv.SaveSessionJournal(&SessionJournal{SessionID: 1, Entries: []JournalEntry{
		{ID: 2, SessionID: 1, Type: "story", Content: "later", Timestamp: ts(10)},
		{ID: 1, SessionID: 1, Type: "story", Content: "earlier", Timestamp: ts(0)},
	}})
	_ = adv.SaveJournalMetadata(&JournalMetadata{NextID: 3, Categories: defaultCategories})

	report, err := adv.RepairJournal(RepairOptions{DryRun: false})
	if err != nil {
		t.Fatalf("RepairJournal() error = %v", err)
	}
	if !hasAction(report, "chrono_sort") {
		t.Fatalf("expected chrono_sort action, got %+v", report.Actions)
	}
	entries, _ := adv.GetEntriesBySession(1)
	if len(entries) != 2 || entries[0].ID != 1 || entries[1].ID != 2 {
		t.Errorf("entries not sorted by timestamp: %+v", entries)
	}
}

func TestRepairReportsAnomaliesWithoutTouching(t *testing.T) {
	adv := newRepairAdv(t)
	_ = adv.SaveSessionJournal(&SessionJournal{SessionID: 1, Entries: []JournalEntry{
		{ID: 1, SessionID: 1, Type: "note", Content: "", Timestamp: ts(0)},     // empty content
		{ID: 1, SessionID: 1, Type: "story", Content: "dup", Timestamp: ts(5)}, // duplicate id
	}})
	_ = adv.SaveJournalMetadata(&JournalMetadata{NextID: 2, Categories: defaultCategories})

	report, err := adv.RepairJournal(RepairOptions{DryRun: true})
	if err != nil {
		t.Fatalf("RepairJournal() error = %v", err)
	}
	if !hasAnomaly(report, "duplicate_id") {
		t.Errorf("expected duplicate_id anomaly, got %+v", report.Anomalies)
	}
	if !hasAnomaly(report, "empty_content") {
		t.Errorf("expected empty_content anomaly, got %+v", report.Anomalies)
	}
}

func TestRepairDryRunWritesNothing(t *testing.T) {
	adv := newRepairAdv(t)
	_ = adv.SaveSessionJournal(&SessionJournal{SessionID: 1, Entries: []JournalEntry{
		{ID: 5, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)},
	}})
	writeRawEmptyCategoriesMeta(t, adv, 1)

	report, err := adv.RepairJournal(RepairOptions{DryRun: true})
	if err != nil {
		t.Fatalf("RepairJournal() error = %v", err)
	}
	if !report.Changed() {
		t.Fatal("expected proposed actions in dry-run")
	}
	if report.BackupDir != "" {
		t.Errorf("dry-run must not create a backup dir, got %q", report.BackupDir)
	}
	// Metadata must be untouched on disk.
	meta, _ := adv.loadJournalMetadata()
	if meta.NextID != 1 || len(meta.Categories) != 0 {
		t.Errorf("dry-run modified metadata on disk: NextID=%d categories=%d", meta.NextID, len(meta.Categories))
	}
	// No backup directory should exist.
	matches, _ := filepath.Glob(filepath.Join(adv.BasePath(), "repair-backup-*"))
	if len(matches) != 0 {
		t.Errorf("dry-run created backup dirs: %v", matches)
	}
}

func TestRepairApplyCreatesBackup(t *testing.T) {
	adv := newRepairAdv(t)
	_ = adv.SaveSessionJournal(&SessionJournal{SessionID: 1, Entries: []JournalEntry{
		{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)},
	}})
	writeRawEmptyCategoriesMeta(t, adv, 1)

	report, err := adv.RepairJournal(RepairOptions{DryRun: false})
	if err != nil {
		t.Fatalf("RepairJournal() error = %v", err)
	}
	if report.BackupDir == "" {
		t.Fatal("apply must create a backup dir")
	}
	if _, err := os.Stat(filepath.Join(report.BackupDir, "journal-meta.json")); err != nil {
		t.Errorf("backup should contain journal-meta.json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(report.BackupDir, "journal-session-1.json")); err != nil {
		t.Errorf("backup should contain journal-session-1.json: %v", err)
	}
}
