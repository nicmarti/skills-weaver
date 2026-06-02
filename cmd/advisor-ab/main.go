// Command advisor-ab runs an A/B comparison of a nested agent's briefing with
// and without the Advisor tool, on the SAME engine version (this build), and
// dumps anonymized briefings ready for blind quality judging plus objective
// cost/latency metrics per arm.
//
// It toggles only one variable — SW_ADVISOR_ENABLED — across two arms:
//   - control:  advisor OFF (standard path)
//   - treatment: advisor ON (beta path with the Advisor tool)
//
// Each trial uses a FRESH AgentManager so metrics reflect that single
// invocation (agent-states.json accumulates otherwise). The real adventure is
// only read, never mutated (the briefing path is read-only and we never call
// SaveAgentStates).
//
// Usage:
//
//	go run ./cmd/advisor-ab -adventure le-voyageur-de-tuncmor -trials 3 -out advisor-ab-out
//
// Requires ANTHROPIC_API_KEY. The target agent's persona must declare an
// advisor (e.g. world-keeper) for the treatment arm to actually consult it.
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"dungeons/internal/agent"
)

const defaultPrompt = "Brief stratégique pour la prochaine session de jeu. " +
	"En tant que world-keeper, indique au Maître du Jeu : comment ouvrir la session, " +
	"comment faire avancer les threads narratifs actifs, et comment payer les foreshadows " +
	"critiques en attente sans les forcer. Tiens compte de l'état du groupe et du campaign-plan."

type runMetrics struct {
	Arm                 string `json:"arm"`
	Trial               int    `json:"trial"`
	AnonID              string `json:"anon_id"`
	LatencyMS           int64  `json:"latency_ms"`
	ExecInputTokens     int64  `json:"exec_input_tokens"`
	ExecOutputTokens    int64  `json:"exec_output_tokens"`
	AdvisorCalls        int64  `json:"advisor_calls"`
	AdvisorInputTokens  int64  `json:"advisor_input_tokens"`
	AdvisorOutputTokens int64  `json:"advisor_output_tokens"`
	AdvisorCacheCreate  int64  `json:"advisor_cache_creation_tokens"`
	AdvisorCacheRead    int64  `json:"advisor_cache_read_tokens"`
	AdvisorModel        string `json:"advisor_model"`
	BriefingChars       int    `json:"briefing_chars"`
	briefing            string
}

// prices in $ per 1M tokens (0 = skip $ cost). Plug in current rates.
type prices struct {
	execIn, execOut, advIn, advOut float64
}

func (p prices) any() bool { return p.execIn+p.execOut+p.advIn+p.advOut > 0 }

// dollarCost computes $ for a run. Cache creation is billed ~1.25x input,
// cache reads ~0.10x input (standard Anthropic ephemeral cache multipliers).
func (p prices) dollarCost(m runMetrics) float64 {
	exec := float64(m.ExecInputTokens)*p.execIn + float64(m.ExecOutputTokens)*p.execOut
	adv := float64(m.AdvisorInputTokens)*p.advIn +
		float64(m.AdvisorCacheCreate)*p.advIn*1.25 +
		float64(m.AdvisorCacheRead)*p.advIn*0.10 +
		float64(m.AdvisorOutputTokens)*p.advOut
	return (exec + adv) / 1_000_000
}

func main() {
	var (
		adventure = flag.String("adventure", "le-voyageur-de-tuncmor", "adventure slug under data/adventures")
		agentName = flag.String("agent", "world-keeper", "nested agent to brief")
		promptStr = flag.String("prompt", defaultPrompt, "briefing prompt (identical for both arms)")
		promptFl  = flag.String("prompt-file", "", "read prompt from file (overrides -prompt)")
		trials    = flag.Int("trials", 3, "trials per arm")
		outDir    = flag.String("out", "advisor-ab-out", "output directory")
		seed      = flag.Int64("seed", 42, "shuffle seed for anonymization")
		pExecIn   = flag.Float64("price-exec-in", 0, "$/1M executor input tokens")
		pExecOut  = flag.Float64("price-exec-out", 0, "$/1M executor output tokens")
		pAdvIn    = flag.Float64("price-adv-in", 0, "$/1M advisor input tokens")
		pAdvOut   = flag.Float64("price-adv-out", 0, "$/1M advisor output tokens")
	)
	flag.Parse()

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "ANTHROPIC_API_KEY not set")
		os.Exit(1)
	}

	prompt := *promptStr
	if *promptFl != "" {
		b, err := os.ReadFile(*promptFl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read prompt-file: %v\n", err)
			os.Exit(1)
		}
		prompt = string(b)
	}

	pr := prices{*pExecIn, *pExecOut, *pAdvIn, *pAdvOut}

	briefDir := filepath.Join(*outDir, "briefings")
	logDir := filepath.Join(*outDir, "logs")
	for _, d := range []string{*outDir, briefDir, logDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", d, err)
			os.Exit(1)
		}
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	var runs []runMetrics

	arms := []struct{ name, flag string }{
		{"control", "0"},   // advisor OFF
		{"treatment", "1"}, // advisor ON
	}

	for _, arm := range arms {
		os.Setenv("SW_ADVISOR_ENABLED", arm.flag)
		for trial := 1; trial <= *trials; trial++ {
			fmt.Printf("[%s] trial %d/%d ...\n", arm.name, trial, *trials)

			m, err := runOnce(apiKey, *adventure, *agentName, prompt, logDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[%s trial %d] error: %v\n", arm.name, trial, err)
				continue
			}
			m.Arm = arm.name
			m.Trial = trial
			runs = append(runs, m)
			fmt.Printf("    %.1fs | exec in=%d out=%d | advisor calls=%d in=%d out=%d cacheW=%d cacheR=%d\n",
				float64(m.LatencyMS)/1000, m.ExecInputTokens, m.ExecOutputTokens,
				m.AdvisorCalls, m.AdvisorInputTokens, m.AdvisorOutputTokens,
				m.AdvisorCacheCreate, m.AdvisorCacheRead)
		}
	}

	if len(runs) == 0 {
		fmt.Fprintln(os.Stderr, "no successful runs")
		os.Exit(1)
	}

	// Anonymize: shuffle a stable id assignment so the judge can't infer the arm.
	rng := rand.New(rand.NewSource(*seed))
	order := rng.Perm(len(runs))
	key := map[string]map[string]any{}
	for anonIdx, runIdx := range order {
		id := fmt.Sprintf("brief-%02d", anonIdx+1)
		runs[runIdx].AnonID = id
		key[id] = map[string]any{"arm": runs[runIdx].Arm, "trial": runs[runIdx].Trial}
	}

	writeBriefings(briefDir, runs)
	writeKey(filepath.Join(*outDir, "_key.json"), key)
	writeMetrics(*outDir, runs)
	writeSummary(filepath.Join(*outDir, "summary.md"), runs, pr, *adventure, *agentName, *trials)
	writeJudging(filepath.Join(*outDir, "JUDGING.md"), runs)

	fmt.Printf("\nDone. %d runs. See %s/ (briefings/, summary.md, JUDGING.md, metrics.csv).\n", len(runs), *outDir)
	fmt.Println("Judge the anonymized briefings in JUDGING.md, then hand the scores back.")
}

// runOnce loads a fresh adventure context + AgentManager and invokes the agent's
// silent briefing path once, returning parsed metrics.
func runOnce(apiKey, adventureSlug, agentName, prompt, logDir string) (runMetrics, error) {
	var m runMetrics

	adventureCtx, err := agent.LoadAdventureContext("data/adventures", adventureSlug)
	if err != nil {
		return m, fmt.Errorf("load adventure: %w", err)
	}
	logger, err := agent.NewLogger(logDir)
	if err != nil {
		return m, fmt.Errorf("logger: %w", err)
	}
	// Wire the full tool registry so world-keeper gets its read-only tools
	// (get_session_info, get_campaign_plan, list_foreshadows, ...) in BOTH arms.
	// The non-silent InvokeAgent path lets the executor actually call those
	// tools (and, in the treatment arm, the advisor) before writing the briefing.
	am, err := agent.NewAgentManagerWithTools(apiKey, adventureCtx, logger, nil)
	if err != nil {
		return m, fmt.Errorf("agent manager: %w", err)
	}

	start := time.Now()
	briefing, err := am.InvokeAgent(agentName, prompt, "", 1)
	m.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		return m, fmt.Errorf("invoke: %w", err)
	}

	state, ok := am.GetNestedAgentState(agentName)
	if !ok {
		return m, fmt.Errorf("agent %s not tracked", agentName)
	}
	mt := state.Metrics()
	m.briefing = briefing
	m.BriefingChars = len(briefing)
	m.ExecInputTokens = mt.TotalInputTokens
	m.ExecOutputTokens = mt.TotalOutputTokens
	m.AdvisorCalls = mt.AdvisorCalls
	m.AdvisorInputTokens = mt.AdvisorInputTokens
	m.AdvisorOutputTokens = mt.AdvisorOutputTokens
	m.AdvisorCacheCreate = mt.AdvisorCacheCreationTokens
	m.AdvisorCacheRead = mt.AdvisorCacheReadTokens
	m.AdvisorModel = mt.AdvisorModelUsed
	return m, nil
}

func writeBriefings(dir string, runs []runMetrics) {
	for _, r := range runs {
		path := filepath.Join(dir, r.AnonID+".md")
		content := fmt.Sprintf("# %s\n\n_(blind sample — do not infer the arm)_\n\n---\n\n%s\n", r.AnonID, r.briefing)
		_ = os.WriteFile(path, []byte(content), 0644)
	}
}

func writeKey(path string, key map[string]map[string]any) {
	b, _ := json.MarshalIndent(key, "", "  ")
	_ = os.WriteFile(path, b, 0644)
}

func writeMetrics(outDir string, runs []runMetrics) {
	// JSON
	b, _ := json.MarshalIndent(runs, "", "  ")
	_ = os.WriteFile(filepath.Join(outDir, "metrics.json"), b, 0644)

	// CSV
	f, err := os.Create(filepath.Join(outDir, "metrics.csv"))
	if err != nil {
		return
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	_ = w.Write([]string{"anon_id", "arm", "trial", "latency_ms", "exec_in", "exec_out",
		"advisor_calls", "advisor_in", "advisor_out", "advisor_cache_create", "advisor_cache_read", "advisor_model", "briefing_chars"})
	for _, r := range runs {
		_ = w.Write([]string{r.AnonID, r.Arm, strconv.Itoa(r.Trial), i64(r.LatencyMS),
			i64(r.ExecInputTokens), i64(r.ExecOutputTokens), i64(r.AdvisorCalls),
			i64(r.AdvisorInputTokens), i64(r.AdvisorOutputTokens), i64(r.AdvisorCacheCreate),
			i64(r.AdvisorCacheRead), r.AdvisorModel, strconv.Itoa(r.BriefingChars)})
	}
}

func writeSummary(path string, runs []runMetrics, pr prices, adventure, agentName string, trials int) {
	var sb []byte
	add := func(s string) { sb = append(sb, s...) }

	add(fmt.Sprintf("# A/B Advisor — %s (%s)\n\n", agentName, adventure))
	add(fmt.Sprintf("Trials per arm: %d. Same engine build, only SW_ADVISOR_ENABLED differs.\n\n", trials))
	add("Objective cost/latency per arm (median across trials):\n\n")
	add("| arm | n | latency (s) | exec in | exec out | adv calls | adv in | adv out | adv cacheW | adv cacheR |\n")
	add("|---|---|---|---|---|---|---|---|---|---|\n")

	for _, arm := range []string{"control", "treatment"} {
		g := filterArm(runs, arm)
		if len(g) == 0 {
			continue
		}
		add(fmt.Sprintf("| %s | %d | %.1f | %d | %d | %d | %d | %d | %d | %d |\n",
			arm, len(g),
			float64(medInt(field(g, func(r runMetrics) int64 { return r.LatencyMS })))/1000,
			medInt(field(g, func(r runMetrics) int64 { return r.ExecInputTokens })),
			medInt(field(g, func(r runMetrics) int64 { return r.ExecOutputTokens })),
			medInt(field(g, func(r runMetrics) int64 { return r.AdvisorCalls })),
			medInt(field(g, func(r runMetrics) int64 { return r.AdvisorInputTokens })),
			medInt(field(g, func(r runMetrics) int64 { return r.AdvisorOutputTokens })),
			medInt(field(g, func(r runMetrics) int64 { return r.AdvisorCacheCreate })),
			medInt(field(g, func(r runMetrics) int64 { return r.AdvisorCacheRead })),
		))
	}
	add("\n")

	if pr.any() {
		add("Median $ cost per briefing (cache creation ×1.25, cache read ×0.10):\n\n")
		add("| arm | median $/briefing |\n|---|---|\n")
		for _, arm := range []string{"control", "treatment"} {
			g := filterArm(runs, arm)
			if len(g) == 0 {
				continue
			}
			costs := make([]int64, len(g))
			for i, r := range g {
				costs[i] = int64(pr.dollarCost(r) * 1e6) // micro-dollars for median
			}
			add(fmt.Sprintf("| %s | $%.5f |\n", arm, float64(medInt(costs))/1e6))
		}
		add("\n")
	} else {
		add("> $ cost not computed. Re-run with -price-exec-in/-price-exec-out/-price-adv-in/-price-adv-out (current $/1M rates) to add a $ column.\n")
		add("> Cost reminder: with caching on, advisor input shifts to cacheW (×1.25) then cacheR (×0.10) — sum all advisor input columns.\n\n")
	}

	add("Quality is judged separately and blind — see JUDGING.md. Hand the scores back to combine with this cost table.\n")
	_ = os.WriteFile(path, sb, 0644)
}

func writeJudging(path string, runs []runMetrics) {
	var sb []byte
	add := func(s string) { sb = append(sb, s...) }

	add("# Blind judging — Advisor A/B\n\n")
	add("The briefings in `briefings/` are anonymized (`brief-NN.md`) and shuffled across both arms.\n")
	add("Do NOT open `_key.json` — that's the de-anonymization key, used only after you score.\n\n")
	add("## Rubric (score each 1–5)\n\n")
	add("- **strategy**: respecte l'acte/threads/objectifs du campaign-plan\n")
	add("- **actionable**: ouverture concrète et utilisable par le DM (pas du vague)\n")
	add("- **foreshadow**: payoff naturel des foreshadows critiques (pas de railroad)\n")
	add("- **canon**: cohérence géographique/lore, n'invente rien qui contredise le plan\n")
	add("- **overall**: impression générale\n\n")
	add("## Score table (fill the numbers, then send back)\n\n")
	add("| anon_id | strategy | actionable | foreshadow | canon | overall | notes |\n")
	add("|---|---|---|---|---|---|---|\n")
	// list anon ids in sorted order
	ids := make([]string, 0, len(runs))
	for _, r := range runs {
		ids = append(ids, r.AnonID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		add(fmt.Sprintf("| %s |  |  |  |  |  |  |\n", id))
	}
	add("\nWhen you send these back, I'll de-anonymize via _key.json and combine with the cost table in summary.md.\n")
	_ = os.WriteFile(path, sb, 0644)
}

// helpers

func i64(v int64) string { return strconv.FormatInt(v, 10) }

func filterArm(runs []runMetrics, arm string) []runMetrics {
	var out []runMetrics
	for _, r := range runs {
		if r.Arm == arm {
			out = append(out, r)
		}
	}
	return out
}

func field(runs []runMetrics, f func(runMetrics) int64) []int64 {
	out := make([]int64, len(runs))
	for i, r := range runs {
		out[i] = f(r)
	}
	return out
}

func medInt(vals []int64) int64 {
	if len(vals) == 0 {
		return 0
	}
	cp := append([]int64(nil), vals...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	n := len(cp)
	if n%2 == 1 {
		return cp[n/2]
	}
	return (cp[n/2-1] + cp[n/2]) / 2
}
