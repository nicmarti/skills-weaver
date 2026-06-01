package coherence

import (
	"fmt"
	"sort"
	"strings"

	"dungeons/internal/adventure"
	"dungeons/internal/npcmanager"
)

// AnalyzeDrift runs the drift layer: does the played flow match the planned
// skeleton? It is a deterministic diff between the campaign plan and the actual
// state (sessions played, journal contents, generated NPCs). It deliberately
// avoids judging prose quality — that is the narrative layer's job.
//
// When no campaign plan exists there is nothing to drift from, so the layer is
// empty (score 100).
func AnalyzeDrift(adv *adventure.Adventure) (LayerReport, error) {
	plan, err := adv.LoadCampaignPlan()
	if err != nil {
		return LayerReport{}, fmt.Errorf("loading campaign plan: %w", err)
	}
	if plan == nil {
		return newLayerReport(LayerDrift, nil), nil
	}

	history, err := adv.LoadSessions()
	if err != nil {
		return LayerReport{}, fmt.Errorf("loading sessions: %w", err)
	}
	played := maxPlayedSession(history)

	journal, err := adv.LoadJournal()
	if err != nil {
		return LayerReport{}, fmt.Errorf("loading journal: %w", err)
	}
	journalText := journalTextBlob(journal.Entries)

	npcNames, err := generatedNPCNames(adv)
	if err != nil {
		return LayerReport{}, fmt.Errorf("loading generated NPCs: %w", err)
	}

	var findings []Finding
	findings = append(findings, checkForeshadowPayoffs(plan, played)...)
	findings = append(findings, checkActWindows(plan, played)...)
	findings = append(findings, checkCurrentAct(plan)...)
	findings = append(findings, checkAntagonistPresence(plan, played, journalText)...)
	findings = append(findings, checkPlannedNPCsUsed(plan, played, npcNames)...)

	return newLayerReport(LayerDrift, findings), nil
}

// maxPlayedSession returns the highest session id recorded in sessions.json,
// i.e. how far the campaign has actually progressed. 0 if none.
func maxPlayedSession(history *adventure.SessionHistory) int {
	played := 0
	for _, s := range history.Sessions {
		if s.ID > played {
			played = s.ID
		}
	}
	return played
}

// journalTextBlob returns a single lowercased string of all journal text, used
// for presence checks of proper nouns (antagonist, etc.).
func journalTextBlob(entries []adventure.JournalEntry) string {
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(e.Content)
		b.WriteByte(' ')
		b.WriteString(e.Description)
		b.WriteByte(' ')
		b.WriteString(e.DescriptionFr)
		b.WriteByte(' ')
	}
	return strings.ToLower(b.String())
}

// generatedNPCNames returns the lowercased names of every NPC actually generated
// in the adventure (from npcs-generated.json).
func generatedNPCNames(adv *adventure.Adventure) (map[string]bool, error) {
	m := npcmanager.NewManager(adv.BasePath())
	db, err := m.Load()
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool)
	for _, records := range db.Sessions {
		for _, r := range records {
			if r.NPC != nil && r.NPC.Name != "" {
				names[strings.ToLower(r.NPC.Name)] = true
			}
		}
	}
	return names, nil
}

// foreshadowSeverity maps a foreshadow's importance to a finding severity.
func foreshadowSeverity(imp adventure.Importance) Severity {
	switch imp {
	case adventure.ImportanceCritical:
		return SeverityError
	case adventure.ImportanceMajor:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// checkForeshadowPayoffs flags active foreshadows whose planned payoff session
// has already passed (overdue), and foreshadows planted in a future session.
func checkForeshadowPayoffs(plan *adventure.CampaignPlan, played int) []Finding {
	var findings []Finding
	for _, fs := range plan.Foreshadows.Active {
		if fs.TargetPayoffSession > 0 && fs.TargetPayoffSession < played {
			findings = append(findings, Finding{
				Layer:    LayerDrift,
				Severity: foreshadowSeverity(fs.Importance),
				Rule:     "foreshadow.overdue_payoff",
				Message: fmt.Sprintf("Le foreshadow '%s' (%s) devait être résolu en session %d, on en est à la session %d et il est toujours actif",
					fs.ID, fs.Importance, fs.TargetPayoffSession, played),
				Context: map[string]any{"id": fs.ID, "importance": string(fs.Importance), "target_payoff": fs.TargetPayoffSession, "played": played},
			})
		}
		if fs.PlantedSession > played {
			findings = append(findings, Finding{
				Layer:    LayerDrift,
				Severity: SeverityInfo,
				Rule:     "foreshadow.planted_in_future",
				Message: fmt.Sprintf("Le foreshadow '%s' est marqué planté en session %d, au-delà de la session jouée %d",
					fs.ID, fs.PlantedSession, played),
				Context: map[string]any{"id": fs.ID, "planted": fs.PlantedSession, "played": played},
			})
		}
	}
	return findings
}

// checkActWindows flags acts whose planned session window has elapsed but that
// are not marked completed.
func checkActWindows(plan *adventure.CampaignPlan, played int) []Finding {
	var findings []Finding
	for _, act := range plan.NarrativeStructure.Acts {
		if len(act.TargetSessions) == 0 || act.Status == "completed" {
			continue
		}
		maxTarget := 0
		for _, s := range act.TargetSessions {
			if s > maxTarget {
				maxTarget = s
			}
		}
		if maxTarget > 0 && maxTarget < played {
			findings = append(findings, Finding{
				Layer:    LayerDrift,
				Severity: SeverityWarning,
				Rule:     "act.window_elapsed",
				Message: fmt.Sprintf("L'acte %d (%s) était planifié jusqu'à la session %d mais est toujours '%s' en session %d",
					act.Number, act.Title, maxTarget, act.Status, played),
				Context: map[string]any{"act": act.Number, "max_target": maxTarget, "status": act.Status, "played": played},
			})
		}
	}
	return findings
}

// checkCurrentAct flags a progression.current_act that is out of range.
func checkCurrentAct(plan *adventure.CampaignPlan) []Finding {
	n := len(plan.NarrativeStructure.Acts)
	if n == 0 {
		return nil
	}
	ca := plan.Progression.CurrentAct
	if ca < 1 || ca > n {
		return []Finding{{
			Layer:    LayerDrift,
			Severity: SeverityWarning,
			Rule:     "progression.current_act_out_of_range",
			Message:  fmt.Sprintf("progression.current_act=%d hors de la plage des actes (1..%d)", ca, n),
			Context:  map[string]any{"current_act": ca, "act_count": n},
		}}
	}
	return nil
}

// checkAntagonistPresence flags an antagonist whose introduction session has
// passed but whose name never appears in the journal.
func checkAntagonistPresence(plan *adventure.CampaignPlan, played int, journalText string) []Finding {
	a := plan.PlotElements.Antagonist
	if a.Name == "" || a.IntroductionSession <= 0 || a.IntroductionSession > played {
		return nil
	}
	if nameMentioned(a.Name, journalText) {
		return nil
	}
	return []Finding{{
		Layer:    LayerDrift,
		Severity: SeverityWarning,
		Rule:     "antagonist.absent_from_journal",
		Message: fmt.Sprintf("L'antagoniste '%s' devait être introduit en session %d mais son nom n'apparaît pas dans le journal",
			a.Name, a.IntroductionSession),
		Context: map[string]any{"name": a.Name, "introduction_session": a.IntroductionSession, "played": played},
	}}
}

// checkPlannedNPCsUsed flags plan NPCs whose introduction session has passed but
// that were never generated in play.
func checkPlannedNPCsUsed(plan *adventure.CampaignPlan, played int, generated map[string]bool) []Finding {
	var findings []Finding
	for _, def := range plan.PlotElements.NPCs {
		if def.Name == "" || def.NarrativeIntegration == nil {
			continue // no scheduled introduction to check against
		}
		intro := def.NarrativeIntegration.IntroductionSession
		if intro <= 0 || intro > played {
			continue // not due yet
		}
		if npcGenerated(def.Name, generated) {
			continue
		}
		findings = append(findings, Finding{
			Layer:    LayerDrift,
			Severity: SeverityWarning,
			Rule:     "npc.planned_never_used",
			Message: fmt.Sprintf("Le PNJ planifié '%s' devait apparaître en session %d mais n'a jamais été généré",
				def.Name, intro),
			Context: map[string]any{"name": def.Name, "introduction_session": intro, "played": played},
		})
	}
	sort.SliceStable(findings, func(i, j int) bool {
		return findings[i].Message < findings[j].Message
	})
	return findings
}

// nameMentioned reports whether a proper name appears in the journal text blob,
// matching the full name or its last token (e.g. "Vastren" for "Colonel Vastren").
func nameMentioned(name, journalText string) bool {
	lname := strings.ToLower(strings.TrimSpace(name))
	if lname == "" {
		return true
	}
	if strings.Contains(journalText, lname) {
		return true
	}
	fields := strings.Fields(lname)
	if len(fields) > 1 {
		last := fields[len(fields)-1]
		if len(last) > 3 && strings.Contains(journalText, last) {
			return true
		}
	}
	return false
}

// npcGenerated reports whether a plan NPC name matches a generated NPC name,
// tolerating minor differences by substring match in either direction.
func npcGenerated(planName string, generated map[string]bool) bool {
	p := strings.ToLower(strings.TrimSpace(planName))
	if generated[p] {
		return true
	}
	for g := range generated {
		if strings.Contains(g, p) || strings.Contains(p, g) {
			return true
		}
	}
	return false
}
