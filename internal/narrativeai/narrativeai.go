// Package narrativeai orchestrates the AI narrative judgment of an adventure.
//
// It is the bridge between the deterministic coherence engine (which produces a
// NarrativeBrief evidence dossier) and the project's nested specialist agents.
// Three analytical lenses each judge the dossier from a distinct angle, then a
// light synthesis ties them together:
//
//   - "world"    -> world-keeper:    consistency with the world/lore, NPC continuity
//   - "rules"    -> rules-keeper:     D&D 5e mechanics, skill usage, encounter balance
//   - "scenario" -> scenario-critic:  pacing, encounter repetition, stagnation, payoffs
//
// The orchestrator depends only on an Invoker (satisfied by *agent.AgentManager),
// so it is decoupled from the heavy agent runtime and unit-testable with a fake.
package narrativeai

import (
	"fmt"
	"strings"
	"time"

	"dungeons/internal/coherence"
)

// Invoker abstracts a single specialist-agent call. *agent.AgentManager satisfies
// it via InvokeAgentSilent (one API call, no tool loop).
type Invoker interface {
	InvokeAgentSilent(agentName, question string, depth int) (string, error)
}

// nestedDepth is the recursion depth for these consultant calls (1 level deep).
const nestedDepth = 1

// lens pairs a human label with the backing persona and its prompt builder.
type lens struct {
	label  string
	agent  string
	prompt func(*coherence.NarrativeBrief) string
}

var lenses = []lens{
	{label: "Monde & histoire", agent: "world-keeper", prompt: worldPrompt},
	{label: "Règles & compétences", agent: "rules-keeper", prompt: rulesPrompt},
	{label: "Scénario & rebondissements", agent: "scenario-critic", prompt: scenarioPrompt},
}

// Judge runs the three lenses over the brief plus a light synthesis and returns a
// NarrativeJudgment. Each lens and the synthesis is a single InvokeAgentSilent call.
//
// Optional progress callbacks are invoked with a human-readable message at each
// step (lens start/done, synthesis, completion) so callers can log server-side
// progress. They are best-effort and may be nil.
func Judge(brief *coherence.NarrativeBrief, inv Invoker, progress ...func(string)) (*coherence.NarrativeJudgment, error) {
	if brief == nil {
		return nil, fmt.Errorf("nil narrative brief")
	}
	report := func(msg string) {
		for _, p := range progress {
			if p != nil {
				p(msg)
			}
		}
	}

	report(fmt.Sprintf("Démarrage du jugement narratif — %d session(s), %d perspectives + synthèse", brief.PlayedSessions, len(lenses)))

	j := &coherence.NarrativeJudgment{
		GeneratedAt:    time.Now(),
		PlayedSessions: brief.PlayedSessions,
	}

	for i, l := range lenses {
		report(fmt.Sprintf("(%d/%d) Consultation de %s — %s…", i+1, len(lenses), l.agent, l.label))
		resp, err := inv.InvokeAgentSilent(l.agent, l.prompt(brief), nestedDepth)
		if err != nil {
			return nil, fmt.Errorf("lens %q (%s): %w", l.label, l.agent, err)
		}
		report(fmt.Sprintf("(%d/%d) %s a répondu (%d caractères)", i+1, len(lenses), l.agent, len(resp)))
		j.Lenses = append(j.Lenses, coherence.LensVerdict{
			Lens:       l.label,
			Agent:      l.agent,
			Assessment: strings.TrimSpace(resp),
		})
	}

	// Light cross-cutting synthesis, delegated to the scenario-critic (which owns
	// the overall adventure view), given the three verdicts.
	report("Synthèse transversale (scenario-critic)…")
	syn, err := inv.InvokeAgentSilent("scenario-critic", synthesisPrompt(j.Lenses), nestedDepth)
	if err != nil {
		return nil, fmt.Errorf("synthesis: %w", err)
	}
	j.Synthesis = strings.TrimSpace(syn)

	report("Jugement narratif terminé.")
	return j, nil
}

// briefDigest renders the evidence dossier as readable text for an agent prompt.
func briefDigest(b *coherence.NarrativeBrief) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Sessions jouées: %d\n\n", b.PlayedSessions)

	sb.WriteString("Profil par session (combat-log = lignes de log, PAS nb de rencontres):\n")
	for _, s := range b.Sessions {
		fmt.Fprintf(&sb, "- Session %d: %d entrées, combat-log=%d, marqueurs-progression=%d\n",
			s.ID, s.EntryCount, s.CombatLogLines, s.ProgressionMarkers)
		if s.Summary != "" {
			fmt.Fprintf(&sb, "  résumé MJ: %s\n", strings.TrimSpace(s.Summary))
		}
	}

	if len(b.TypeTotals) > 0 {
		sb.WriteString("\nTotaux par type d'entrée: ")
		first := true
		for typ, n := range b.TypeTotals {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%s:%d", typ, n)
			first = false
		}
		sb.WriteString("\n")
	}

	if len(b.OpenThreads) > 0 {
		fmt.Fprintf(&sb, "\nFils narratifs actifs (%d): %s\n", len(b.OpenThreads), strings.Join(b.OpenThreads, "; "))
	}
	if len(b.PendingForeshadows) > 0 {
		fmt.Fprintf(&sb, "\nForeshadows non résolus (%d):\n", len(b.PendingForeshadows))
		for _, f := range b.PendingForeshadows {
			payoff := "non planifié"
			if f.TargetPayoffSession > 0 {
				payoff = fmt.Sprintf("session %d", f.TargetPayoffSession)
			}
			fmt.Fprintf(&sb, "- [%s] %s (importance %s, payoff prévu: %s)\n", f.ID, f.Description, f.Importance, payoff)
		}
	}

	if bp := b.Behavior; bp != nil {
		r := bp.TotalRolls
		fmt.Fprintf(&sb, "\nProfil de jeu (depuis les logs, %d session(s)) :\n", bp.SessionsAnalyzed)
		fmt.Fprintf(&sb, "  Jets de dés: total=%d (combat=%d, compétence=%d, social=%d, sauvegarde=%d, autre=%d)\n",
			r.Total, r.Combat, r.Skill, r.Social, r.Save, r.Other)
		if len(bp.TotalAgentConsults) > 0 {
			sb.WriteString("  Consultations d'agents: ")
			first := true
			for ag, n := range bp.TotalAgentConsults {
				if !first {
					sb.WriteString(", ")
				}
				fmt.Fprintf(&sb, "%s:%d", ag, n)
				first = false
			}
			sb.WriteString("\n")
		}
	}

	if len(b.Notes) > 0 {
		sb.WriteString("\nLimites des données (à respecter dans ton analyse):\n")
		for _, n := range b.Notes {
			fmt.Fprintf(&sb, "- %s\n", n)
		}
	}

	return sb.String()
}

func worldPrompt(b *coherence.NarrativeBrief) string {
	return "Analyse ce déroulé d'aventure du point de vue de la COHÉRENCE DU MONDE et de l'HISTOIRE. " +
		"Repère les contradictions avec le canon du monde, les incohérences de PNJ (continuité, statut, présence), " +
		"de géographie ou de chronologie. Ne juge ni les règles ni le rythme — concentre-toi sur le monde.\n\n" +
		"=== DOSSIER ===\n" + briefDigest(b)
}

func rulesPrompt(b *coherence.NarrativeBrief) string {
	return "Analyse ce déroulé d'aventure du point de vue du MOTEUR DE JEU D&D 5e et de l'USAGE DES COMPÉTENCES. " +
		"Repère les mauvais usages de règles, le manque de variété dans les jets/compétences mobilisés, " +
		"la répétition MÉCANIQUE des rencontres et les problèmes d'équilibrage. Ne juge ni le monde ni l'intrigue.\n\n" +
		"=== DOSSIER ===\n" + briefDigest(b)
}

func scenarioPrompt(b *coherence.NarrativeBrief) string {
	return "Analyse ce déroulé d'aventure selon ta mission de critique de scénario : répétition des types de rencontres, " +
		"stagnation (le groupe qui tourne en rond), pacing, paiement des foreshadows, et ce qui n'a pas fonctionné.\n\n" +
		"=== DOSSIER ===\n" + briefDigest(b)
}

func synthesisPrompt(verdicts []coherence.LensVerdict) string {
	var sb strings.Builder
	sb.WriteString("Voici trois analyses indépendantes du même déroulé d'aventure, sous trois angles différents. " +
		"Produis une SYNTHÈSE LÉGÈRE (max 120 mots) : dégage LE point n°1 qui n'a pas fonctionné à travers les trois angles, " +
		"puis 1 à 3 priorités d'amélioration. Ne répète pas tout, croise les analyses.\n\n")
	for _, v := range verdicts {
		fmt.Fprintf(&sb, "=== %s (%s) ===\n%s\n\n", v.Lens, v.Agent, v.Assessment)
	}
	return sb.String()
}
