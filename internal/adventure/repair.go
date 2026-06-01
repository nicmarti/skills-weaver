package adventure

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"
)

// RepairOptions controls how RepairJournal behaves.
type RepairOptions struct {
	// DryRun reports what would change without writing anything to disk.
	DryRun bool
}

// RepairAction is a safe, idempotent fix that repair applied (or would apply).
type RepairAction struct {
	Kind   string `json:"kind"`   // "categories", "next_id", "chrono_sort", "rehome"
	Target string `json:"target"` // affected file or field
	Detail string `json:"detail"` // human-readable description
}

// RepairAnomaly is a problem repair detected but did NOT touch (needs human judgment).
type RepairAnomaly struct {
	Kind   string `json:"kind"`   // "duplicate_id", "empty_content"
	Detail string `json:"detail"` // human-readable description
}

// RepairReport summarizes a repair pass over an adventure's journal.
type RepairReport struct {
	AdventureName string          `json:"adventure_name"`
	DryRun        bool            `json:"dry_run"`
	TotalEntries  int             `json:"total_entries"`
	Actions       []RepairAction  `json:"actions"`
	Anomalies     []RepairAnomaly `json:"anomalies"`
	BackupDir     string          `json:"backup_dir,omitempty"` // where originals were copied (empty in dry-run)
}

// Changed reports whether the repair pass found anything to fix.
func (r *RepairReport) Changed() bool {
	return len(r.Actions) > 0
}

// effectiveSessionID returns the session an entry belongs to.
// SessionID is omitempty in JSON, so a missing field unmarshals to 0 (out-of-session).
func effectiveSessionID(e JournalEntry) int {
	return e.SessionID
}

// RepairJournal canonicalizes an adventure's journal to the invariants the rest of
// the engine assumes, applying only safe, idempotent fixes. It never deletes content.
//
// Fixes applied (recorded as Actions):
//   - categories: repopulate journal-meta.json categories[] if empty/missing
//   - next_id:    resync journal-meta.json next_id to max(id)+1
//   - chrono_sort: re-sort entries within each session file by timestamp
//   - rehome:     move an entry whose session_id != its file's session to the right file
//
// Problems detected but left untouched (recorded as Anomalies):
//   - duplicate_id:   the same id appears on more than one entry
//   - empty_content:  an entry has no content and no descriptions
//
// In DryRun mode nothing is written; the returned report still lists what would change.
// Otherwise, every file that will be modified is first copied into a timestamped
// repair-backup-<ts>/ directory under the adventure folder.
func (a *Adventure) RepairJournal(opts RepairOptions) (*RepairReport, error) {
	report := &RepairReport{
		AdventureName: a.Name,
		DryRun:        opts.DryRun,
	}

	// --- Load every session journal file --------------------------------------
	pattern := filepath.Join(a.basePath, "journal-session-*.json")
	sessionFiles, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("finding session journals: %w", err)
	}

	// original[sid] = entries as currently stored in journal-session-<sid>.json
	original := make(map[int][]JournalEntry)
	for _, file := range sessionFiles {
		var sid int
		if _, err := fmt.Sscanf(filepath.Base(file), "journal-session-%d.json", &sid); err != nil {
			continue // skip malformed filenames
		}
		sj, err := a.loadSessionJournal(sid)
		if err != nil {
			return nil, fmt.Errorf("loading session %d: %w", sid, err)
		}
		original[sid] = sj.Entries
	}

	// --- Re-home entries to the file matching their session_id ----------------
	// grouped[targetSid] = entries that belong in journal-session-<targetSid>.json
	grouped := make(map[int][]JournalEntry)
	var allEntries []JournalEntry
	for fileSid, entries := range original {
		for _, e := range entries {
			target := effectiveSessionID(e)
			if target != fileSid {
				report.Actions = append(report.Actions, RepairAction{
					Kind:   "rehome",
					Target: fmt.Sprintf("journal-session-%d.json", fileSid),
					Detail: fmt.Sprintf("entry id=%d moved to journal-session-%d.json (session_id=%d)", e.ID, target, target),
				})
			}
			grouped[target] = append(grouped[target], e)
			allEntries = append(allEntries, e)
		}
	}
	report.TotalEntries = len(allEntries)

	// --- Re-sort each target group by timestamp -------------------------------
	for sid := range grouped {
		entries := grouped[sid]
		sorted := make([]JournalEntry, len(entries))
		copy(sorted, entries)
		sort.SliceStable(sorted, func(i, j int) bool {
			return sorted[i].Timestamp.Before(sorted[j].Timestamp)
		})
		grouped[sid] = sorted

		// Only report a chrono_sort action when this file already existed and the
		// order actually changed (avoid noise from pure re-homing).
		if orig, existed := original[sid]; existed && !reflect.DeepEqual(orig, sorted) {
			// Distinguish a real re-ordering from a re-homing-only change: if the set
			// of ids differs, the change is already captured by rehome actions.
			if sameIDSet(orig, sorted) {
				report.Actions = append(report.Actions, RepairAction{
					Kind:   "chrono_sort",
					Target: fmt.Sprintf("journal-session-%d.json", sid),
					Detail: "entries re-sorted by timestamp",
				})
			}
		}
	}

	// --- Detect anomalies (report only) ---------------------------------------
	seen := make(map[int]int)
	for _, e := range allEntries {
		seen[e.ID]++
	}
	dupIDs := []int{}
	for id, n := range seen {
		if n > 1 {
			dupIDs = append(dupIDs, id)
		}
	}
	sort.Ints(dupIDs)
	for _, id := range dupIDs {
		report.Anomalies = append(report.Anomalies, RepairAnomaly{
			Kind:   "duplicate_id",
			Detail: fmt.Sprintf("id %d appears %d times", id, seen[id]),
		})
	}

	for _, e := range allEntries {
		if e.Content == "" && e.Description == "" && e.DescriptionFr == "" {
			report.Anomalies = append(report.Anomalies, RepairAnomaly{
				Kind:   "empty_content",
				Detail: fmt.Sprintf("entry id=%d (type=%s, session=%d) has no content", e.ID, e.Type, e.SessionID),
			})
		}
	}

	// --- Metadata: categories + next_id ---------------------------------------
	meta, err := a.loadJournalMetadata()
	if err != nil {
		return nil, fmt.Errorf("loading journal metadata: %w", err)
	}
	newMeta := *meta // copy

	if len(newMeta.Categories) == 0 {
		newMeta.Categories = append([]string(nil), defaultCategories...)
		report.Actions = append(report.Actions, RepairAction{
			Kind:   "categories",
			Target: "journal-meta.json",
			Detail: fmt.Sprintf("empty categories[] repopulated with %d defaults", len(defaultCategories)),
		})
	}

	maxID := 0
	for _, e := range allEntries {
		if e.ID > maxID {
			maxID = e.ID
		}
	}
	wantNextID := maxID + 1
	if newMeta.NextID != wantNextID {
		report.Actions = append(report.Actions, RepairAction{
			Kind:   "next_id",
			Target: "journal-meta.json",
			Detail: fmt.Sprintf("next_id %d -> %d (max id is %d)", newMeta.NextID, wantNextID, maxID),
		})
		newMeta.NextID = wantNextID
	}

	// --- Nothing to write in dry-run, or nothing changed ----------------------
	if opts.DryRun || !report.Changed() {
		return report, nil
	}

	// --- Backup originals before writing --------------------------------------
	backupDir := filepath.Join(a.basePath, "repair-backup-"+time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("creating backup dir: %w", err)
	}
	report.BackupDir = backupDir

	backup := func(name string) error {
		src := filepath.Join(a.basePath, name)
		if _, statErr := os.Stat(src); statErr != nil {
			return nil // nothing to back up
		}
		data, readErr := os.ReadFile(src)
		if readErr != nil {
			return readErr
		}
		return os.WriteFile(filepath.Join(backupDir, name), data, 0644)
	}

	// Back up meta and every existing session file.
	if err := backup("journal-meta.json"); err != nil {
		return nil, fmt.Errorf("backing up metadata: %w", err)
	}
	for sid := range original {
		if err := backup(fmt.Sprintf("journal-session-%d.json", sid)); err != nil {
			return nil, fmt.Errorf("backing up session %d: %w", sid, err)
		}
	}

	// --- Write changed session files ------------------------------------------
	// A file is rewritten if its new entries differ from the original ones, or if
	// re-homing created a new target file.
	for sid, entries := range grouped {
		if orig, existed := original[sid]; existed && reflect.DeepEqual(orig, entries) {
			continue // unchanged
		}
		if err := a.SaveSessionJournal(&SessionJournal{SessionID: sid, Entries: entries}); err != nil {
			return nil, fmt.Errorf("writing session %d: %w", sid, err)
		}
	}
	// A source file fully emptied by re-homing must be written back as empty so it
	// no longer holds the moved entries.
	for sid := range original {
		if _, stillHasEntries := grouped[sid]; !stillHasEntries {
			if err := a.SaveSessionJournal(&SessionJournal{SessionID: sid, Entries: []JournalEntry{}}); err != nil {
				return nil, fmt.Errorf("emptying session %d: %w", sid, err)
			}
		}
	}

	// --- Write metadata --------------------------------------------------------
	if err := a.SaveJournalMetadata(&newMeta); err != nil {
		return nil, fmt.Errorf("writing journal metadata: %w", err)
	}

	return report, nil
}

// sameIDSet reports whether two entry slices contain exactly the same set of ids.
func sameIDSet(a, b []JournalEntry) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[int]int, len(a))
	for _, e := range a {
		set[e.ID]++
	}
	for _, e := range b {
		set[e.ID]--
		if set[e.ID] < 0 {
			return false
		}
	}
	return true
}
