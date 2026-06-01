package coherence

import (
	"fmt"
	"path/filepath"
	"sort"

	"dungeons/internal/adventure"
)

// AnalyzeIntegrity runs the data-integrity layer: every file well-formed, and
// cross-file references valid. It is read-only and deterministic.
//
// It detects (and never fixes) the same classes of problem sw-adventure repair
// addresses, plus cross-file ones repair does not touch. The two intentionally
// share no code: repair lives in the adventure package and mutates; this is a
// pure reporter that imports adventure one-way.
func AnalyzeIntegrity(adv *adventure.Adventure) (LayerReport, error) {
	var findings []Finding

	journal, err := adv.LoadJournal()
	if err != nil {
		return LayerReport{}, fmt.Errorf("loading journal: %w", err)
	}

	perFile, err := loadPerFileEntries(adv)
	if err != nil {
		return LayerReport{}, fmt.Errorf("loading per-file journal: %w", err)
	}

	findings = append(findings, checkJournalIDs(journal.Entries)...)
	findings = append(findings, checkNextID(journal)...)
	findings = append(findings, checkCategories(journal)...)
	findings = append(findings, checkMisfiledEntries(perFile)...)
	findings = append(findings, checkChronoOrder(perFile)...)
	findings = append(findings, checkEmptyContent(journal.Entries)...)

	sessions, err := adv.LoadSessions()
	if err != nil {
		return LayerReport{}, fmt.Errorf("loading sessions: %w", err)
	}
	findings = append(findings, checkSessions(journal.Entries, sessions, adv)...)

	plan, err := adv.LoadCampaignPlan()
	if err != nil {
		return LayerReport{}, fmt.Errorf("loading campaign plan: %w", err)
	}
	findings = append(findings, checkCampaignPlan(plan)...)

	return newLayerReport(LayerIntegrity, findings), nil
}

// loadPerFileEntries returns, for each existing journal-session-<sid>.json file,
// the entries as currently stored (unsorted), keyed by the file's session id.
func loadPerFileEntries(adv *adventure.Adventure) (map[int][]adventure.JournalEntry, error) {
	pattern := filepath.Join(adv.BasePath(), "journal-session-*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	out := make(map[int][]adventure.JournalEntry)
	for _, f := range files {
		var sid int
		if _, err := fmt.Sscanf(filepath.Base(f), "journal-session-%d.json", &sid); err != nil {
			continue // skip malformed filenames
		}
		entries, err := adv.GetEntriesBySession(sid)
		if err != nil {
			return nil, err
		}
		out[sid] = entries
	}
	return out, nil
}

// checkJournalIDs flags duplicate ids (error) and gaps in the global id sequence
// (info — a gap can be the legitimate result of a deleted entry).
func checkJournalIDs(entries []adventure.JournalEntry) []Finding {
	var findings []Finding
	if len(entries) == 0 {
		return findings
	}

	seen := make(map[int]int)
	for _, e := range entries {
		seen[e.ID]++
	}
	dups := []int{}
	for id, n := range seen {
		if n > 1 {
			dups = append(dups, id)
		}
	}
	sort.Ints(dups)
	for _, id := range dups {
		findings = append(findings, Finding{
			Layer:    LayerIntegrity,
			Severity: SeverityError,
			Rule:     "journal.duplicate_id",
			Message:  fmt.Sprintf("L'identifiant d'entrée %d apparaît %d fois", id, seen[id]),
			Context:  map[string]any{"id": id, "count": seen[id]},
		})
	}

	ids := make([]int, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	min, max := ids[0], ids[len(ids)-1]
	missing := []int{}
	for i := min; i <= max; i++ {
		if seen[i] == 0 {
			missing = append(missing, i)
		}
	}
	if len(missing) > 0 {
		findings = append(findings, Finding{
			Layer:    LayerIntegrity,
			Severity: SeverityInfo,
			Rule:     "journal.id_gap",
			Message:  fmt.Sprintf("%d identifiant(s) manquant(s) dans la séquence %d..%d (possible suppression)", len(missing), min, max),
			Context:  map[string]any{"missing": missing, "min": min, "max": max},
		})
	}
	return findings
}

// checkNextID flags journal-meta next_id desynced from the real max id.
func checkNextID(journal *adventure.Journal) []Finding {
	maxID := 0
	for _, e := range journal.Entries {
		if e.ID > maxID {
			maxID = e.ID
		}
	}
	want := maxID + 1
	if journal.NextID != want {
		return []Finding{{
			Layer:    LayerIntegrity,
			Severity: SeverityWarning,
			Rule:     "journal.next_id_desync",
			Message:  fmt.Sprintf("next_id=%d alors que l'id max est %d (devrait être %d)", journal.NextID, maxID, want),
			Context:  map[string]any{"next_id": journal.NextID, "max_id": maxID, "want": want},
		}}
	}
	return nil
}

// checkCategories flags an empty categories[] in journal-meta.
func checkCategories(journal *adventure.Journal) []Finding {
	if len(journal.Categories) == 0 {
		return []Finding{{
			Layer:    LayerIntegrity,
			Severity: SeverityWarning,
			Rule:     "journal.categories_empty",
			Message:  "Le champ categories[] de journal-meta.json est vide",
			Context:  map[string]any{},
		}}
	}
	return nil
}

// checkMisfiledEntries flags entries whose session_id does not match the session
// file they are stored in.
func checkMisfiledEntries(perFile map[int][]adventure.JournalEntry) []Finding {
	var findings []Finding
	for fileSid, entries := range perFile {
		for _, e := range entries {
			if e.SessionID != fileSid {
				findings = append(findings, Finding{
					Layer:    LayerIntegrity,
					Severity: SeverityError,
					Rule:     "journal.misfiled_entry",
					Message:  fmt.Sprintf("L'entrée id=%d (session_id=%d) est stockée dans journal-session-%d.json", e.ID, e.SessionID, fileSid),
					Context:  map[string]any{"id": e.ID, "session_id": e.SessionID, "file_session": fileSid},
				})
			}
		}
	}
	sortFindingsByRuleAndID(findings)
	return findings
}

// checkChronoOrder flags session files whose entries are not stored in
// timestamp order.
func checkChronoOrder(perFile map[int][]adventure.JournalEntry) []Finding {
	var findings []Finding
	sids := make([]int, 0, len(perFile))
	for sid := range perFile {
		sids = append(sids, sid)
	}
	sort.Ints(sids)
	for _, sid := range sids {
		entries := perFile[sid]
		for i := 1; i < len(entries); i++ {
			if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
				findings = append(findings, Finding{
					Layer:    LayerIntegrity,
					Severity: SeverityWarning,
					Rule:     "journal.chrono_order",
					Message:  fmt.Sprintf("journal-session-%d.json : entrées non triées par timestamp (id=%d avant id=%d)", sid, entries[i-1].ID, entries[i].ID),
					Context:  map[string]any{"session": sid, "before_id": entries[i-1].ID, "after_id": entries[i].ID},
				})
				break // one finding per file is enough
			}
		}
	}
	return findings
}

// checkEmptyContent flags entries with no content and no descriptions.
func checkEmptyContent(entries []adventure.JournalEntry) []Finding {
	var findings []Finding
	for _, e := range entries {
		if e.Content == "" && e.Description == "" && e.DescriptionFr == "" {
			findings = append(findings, Finding{
				Layer:    LayerIntegrity,
				Severity: SeverityInfo,
				Rule:     "journal.empty_content",
				Message:  fmt.Sprintf("Entrée id=%d (type=%s, session=%d) sans contenu", e.ID, e.Type, e.SessionID),
				Context:  map[string]any{"id": e.ID, "type": e.Type, "session": e.SessionID},
			})
		}
	}
	return findings
}

// checkSessions cross-checks journal sessions against sessions.json.
func checkSessions(entries []adventure.JournalEntry, history *adventure.SessionHistory, adv *adventure.Adventure) []Finding {
	var findings []Finding

	declared := make(map[int]bool)
	for _, s := range history.Sessions {
		declared[s.ID] = true
	}

	// Sessions referenced by journal entries but absent from sessions.json.
	journalSessions := make(map[int]bool)
	for _, e := range entries {
		if e.SessionID > 0 {
			journalSessions[e.SessionID] = true
		}
	}
	orphans := []int{}
	for sid := range journalSessions {
		if !declared[sid] {
			orphans = append(orphans, sid)
		}
	}
	sort.Ints(orphans)
	for _, sid := range orphans {
		findings = append(findings, Finding{
			Layer:    LayerIntegrity,
			Severity: SeverityWarning,
			Rule:     "session.orphaned",
			Message:  fmt.Sprintf("La session %d a des entrées de journal mais n'est pas déclarée dans sessions.json", sid),
			Context:  map[string]any{"session": sid},
		})
	}

	// adventure.json session_count vs the number of declared sessions.
	if adv.SessionCount != len(history.Sessions) {
		findings = append(findings, Finding{
			Layer:    LayerIntegrity,
			Severity: SeverityInfo,
			Rule:     "session.count_mismatch",
			Message:  fmt.Sprintf("adventure.json session_count=%d mais sessions.json déclare %d session(s)", adv.SessionCount, len(history.Sessions)),
			Context:  map[string]any{"session_count": adv.SessionCount, "declared": len(history.Sessions)},
		})
	}

	return findings
}

// checkCampaignPlan folds ONLY the structural errors of the existing campaign-plan
// validation into the integrity layer (a malformed/self-contradictory plan file is
// an integrity problem). Its warnings are deliberately NOT included here: they are
// either content policy (e.g. anti-cult keywords) or skeleton-vs-cast drift (e.g. a
// foreshadow referencing an unknown NPC), which belong to the drift/policy layers.
// Folding them in would let a thematic concern flood and mask the integrity score.
// Skipped when no plan exists.
func checkCampaignPlan(plan *adventure.CampaignPlan) []Finding {
	if plan == nil {
		return nil
	}
	result := adventure.ValidateCampaignPlan(plan)
	var findings []Finding
	for _, e := range result.Errors {
		findings = append(findings, Finding{
			Layer:    LayerIntegrity,
			Severity: SeverityError,
			Rule:     "campaign.plan_invalid",
			Message:  e,
			Context:  map[string]any{},
		})
	}
	return findings
}

// sortFindingsByRuleAndID gives misfiled findings a stable order for output.
func sortFindingsByRuleAndID(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		ai, _ := findings[i].Context["id"].(int)
		aj, _ := findings[j].Context["id"].(int)
		return ai < aj
	})
}
