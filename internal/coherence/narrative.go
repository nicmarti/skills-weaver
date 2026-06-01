package coherence

import (
	"fmt"
	"sort"
	"strings"

	"dungeons/internal/adventure"
)

// NarrativeBrief is the evidence dossier the narrative layer assembles for an
// LLM to judge (repetition, stagnation, what did not work). It contains only
// deterministic, reliable facts — no quality verdict. The judgment step
// (delegated to Claude) is wired on top of this in a later iteration.
type NarrativeBrief struct {
	PlayedSessions     int                `json:"played_sessions"`
	TypeTotals         map[string]int     `json:"type_totals"` // entry-type histogram across the journal
	Sessions           []SessionDigest    `json:"sessions"`
	OpenThreads        []string           `json:"open_threads"`        // plan active narrative threads
	PendingForeshadows []ForeshadowDigest `json:"pending_foreshadows"` // active (unresolved) foreshadows
	Behavior           *BehaviorProfile   `json:"behavior,omitempty"`  // log-derived DM behavior profile
	Notes              []string           `json:"notes"`               // data caveats for the LLM interpreter
}

// SessionDigest summarizes one played session for the dossier.
type SessionDigest struct {
	ID                 int            `json:"id"`
	EntryCount         int            `json:"entry_count"`
	TypeCounts         map[string]int `json:"type_counts"`
	CombatLogLines     int            `json:"combat_log_lines"`    // honest: blow-by-blow lines, NOT encounter count
	ProgressionMarkers int            `json:"progression_markers"` // raw count; sparsely logged, do not over-read
	Summary            string         `json:"summary"`             // DM recap from sessions.json
	EngineVersion      string         `json:"engine_version"`      // engine build that played this session ("v0" = legacy)
}

// ForeshadowDigest is a compact view of an unresolved foreshadow.
type ForeshadowDigest struct {
	ID                  string `json:"id"`
	Description         string `json:"description"`
	Importance          string `json:"importance"`
	TargetPayoffSession int    `json:"target_payoff_session,omitempty"`
}

// narrativeDataNotes are the caveats every brief carries so the downstream LLM
// interprets the numbers correctly.
var narrativeDataNotes = []string{
	"Les combats sont loggés coup par coup : combat_log_lines ne reflète PAS le nombre de rencontres.",
	"Les marqueurs de progression sont rarement loggés : progression_markers=0 n'implique PAS l'absence de progrès.",
	"sessions.json location/xp/gold ne sont pas peuplés — non inclus dans ce dossier.",
}

// AnalyzeNarrative runs the narrative layer. Per the layered design, it emits
// only conservative, reliable observations (never errors): genuine narrative
// judgment (repetition, circling, what failed) is delegated to an LLM that
// consumes BuildNarrativeBrief's dossier. This keeps the deterministic engine
// free of false-positive narrative verdicts.
func AnalyzeNarrative(adv *adventure.Adventure) (LayerReport, error) {
	brief, err := BuildNarrativeBrief(adv)
	if err != nil {
		return LayerReport{}, err
	}

	var findings []Finding

	// Thin sessions: a played session with almost no journal activity.
	for _, sd := range brief.Sessions {
		if sd.EntryCount <= 2 {
			findings = append(findings, Finding{
				Layer:    LayerNarrative,
				Severity: SeverityInfo,
				Rule:     "narrative.thin_session",
				Message:  fmt.Sprintf("Session %d : seulement %d entrée(s) de journal — session très courte ou avortée", sd.ID, sd.EntryCount),
				Context:  map[string]any{"session": sd.ID, "entry_count": sd.EntryCount},
			})
		}
	}

	// Open narrative loops still pending at the latest session — a pointer for
	// the designer / LLM, not a verdict.
	open := len(brief.OpenThreads) + len(brief.PendingForeshadows)
	if open > 0 {
		findings = append(findings, Finding{
			Layer:    LayerNarrative,
			Severity: SeverityInfo,
			Rule:     "narrative.open_loops",
			Message: fmt.Sprintf("%d fil(s) narratif(s) actif(s) et %d foreshadow(s) non résolu(s) à la session %d",
				len(brief.OpenThreads), len(brief.PendingForeshadows), brief.PlayedSessions),
			Context: map[string]any{"open_threads": len(brief.OpenThreads), "pending_foreshadows": len(brief.PendingForeshadows)},
		})
	}

	// Behavior signals — conservative, evidence-only (never errors). Only fire on a
	// meaningful sample so they are reliable rather than noisy.
	if b := brief.Behavior; b != nil && b.TotalRolls.Total >= 20 {
		r := b.TotalRolls
		if r.Social == 0 {
			findings = append(findings, Finding{
				Layer:    LayerNarrative,
				Severity: SeverityInfo,
				Rule:     "behavior.no_social_rolls",
				Message:  fmt.Sprintf("Aucun jet social sur %d jets — interactions sociales peu mises à l'épreuve mécaniquement", r.Total),
				Context:  map[string]any{"total": r.Total, "social": r.Social},
			})
		}
		if r.Combat > 2*(r.Skill+r.Social) {
			findings = append(findings, Finding{
				Layer:    LayerNarrative,
				Severity: SeverityInfo,
				Rule:     "behavior.combat_skewed",
				Message:  fmt.Sprintf("Jeu orienté combat : %d jets de combat vs %d compétence + %d social", r.Combat, r.Skill, r.Social),
				Context:  map[string]any{"combat": r.Combat, "skill": r.Skill, "social": r.Social},
			})
		}
	}

	return newLayerReport(LayerNarrative, findings), nil
}

// BuildNarrativeBrief assembles the deterministic evidence dossier.
func BuildNarrativeBrief(adv *adventure.Adventure) (*NarrativeBrief, error) {
	journal, err := adv.LoadJournal()
	if err != nil {
		return nil, fmt.Errorf("loading journal: %w", err)
	}
	perFile, err := loadPerFileEntries(adv)
	if err != nil {
		return nil, fmt.Errorf("loading per-file journal: %w", err)
	}
	history, err := adv.LoadSessions()
	if err != nil {
		return nil, fmt.Errorf("loading sessions: %w", err)
	}

	summaries := make(map[int]string)
	versions := make(map[int]string)
	for _, s := range history.Sessions {
		summaries[s.ID] = s.Summary
		versions[s.ID] = engineVersionOrV0(s.EngineVersion)
	}

	brief := &NarrativeBrief{
		PlayedSessions: maxPlayedSession(history),
		TypeTotals:     map[string]int{},
		Notes:          narrativeDataNotes,
	}

	for _, e := range journal.Entries {
		brief.TypeTotals[e.Type]++
	}

	// One digest per played session file (id >= 1), sorted.
	sids := make([]int, 0, len(perFile))
	for sid := range perFile {
		if sid >= 1 {
			sids = append(sids, sid)
		}
	}
	sort.Ints(sids)
	for _, sid := range sids {
		entries := perFile[sid]
		digest := SessionDigest{
			ID:            sid,
			EntryCount:    len(entries),
			TypeCounts:    map[string]int{},
			Summary:       summaries[sid],
			EngineVersion: engineVersionOrV0(versions[sid]),
		}
		for _, e := range entries {
			digest.TypeCounts[e.Type]++
			if e.Type == "combat" {
				digest.CombatLogLines++
			}
			if isProgressionMarker(e) {
				digest.ProgressionMarkers++
			}
		}
		brief.Sessions = append(brief.Sessions, digest)
	}

	// Log-derived DM behavior profile (best-effort; nil when no logs).
	if bp, bErr := BuildBehaviorProfile(adv); bErr == nil && bp.SessionsAnalyzed > 0 {
		brief.Behavior = bp
	}

	// Open narrative loops from the campaign plan (if any).
	plan, err := adv.LoadCampaignPlan()
	if err != nil {
		return nil, fmt.Errorf("loading campaign plan: %w", err)
	}
	if plan != nil {
		brief.OpenThreads = append(brief.OpenThreads, plan.Progression.ActiveThreads...)
		for _, fs := range plan.Foreshadows.Active {
			brief.PendingForeshadows = append(brief.PendingForeshadows, ForeshadowDigest{
				ID:                  fs.ID,
				Description:         fs.Description,
				Importance:          string(fs.Importance),
				TargetPayoffSession: fs.TargetPayoffSession,
			})
		}
	}

	return brief, nil
}

// engineVersionOrV0 labels legacy (pre-versioning) sessions as "v0".
func engineVersionOrV0(v string) string {
	if v == "" {
		return "v0"
	}
	return v
}

// isProgressionMarker reports whether an entry signals story progression.
// This is a sparse, best-effort signal — see narrativeDataNotes.
func isProgressionMarker(e adventure.JournalEntry) bool {
	if e.Type == "quest" || e.Type == "levelup" {
		return true
	}
	c := strings.ToLower(e.Content)
	for _, kw := range []string{"plot point completed", "narrative thread", "quête", "quest"} {
		if strings.Contains(c, kw) {
			return true
		}
	}
	return false
}
