// sw-bench-campaign benchmarks one OpenRouter model on the wizard's
// campaign-plan generation.
//
// It builds exactly the same prompt as the web wizard (internal/campaign),
// streams campaign plans (see --runs), and reports wall-clock time,
// time-to-first delta, token usage, cost and the plan's validation score so
// models can be compared on speed, price and output quality.
//
// Nothing is written under data/adventures: generated plans go to --out
// (default /tmp/sw-bench) for manual quality review.
//
// Usage:
//
//	./sw-bench-campaign anthropic/claude-sonnet-5 --runs 2
//	./sw-bench-campaign z-ai/glm-5.3
//	./sw-bench-campaign openai/gpt-6-luna-pro --theme "Un siège en hiver"
//
// The model is a full OpenRouter ID (provider/model); the convenience
// aliases (sonnet, haiku, opus) are deliberately NOT accepted here — a
// benchmark run must name the exact model under test.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/campaign"
	"dungeons/internal/llm"
)

const (
	defaultName = "Aventure de benchmark"
	// The fixed default theme keeps runs comparable across models; override
	// with --theme for specific scenarios.
	defaultTheme    = "Un artefact ancien réapparaît aux marges des Quatre Royaumes et attire les convoitises"
	defaultDuration = "campaign"
	defaultOutDir   = "/tmp/sw-bench"
)

type runResult struct {
	Model    string
	Run      int
	Err      error
	Elapsed  time.Duration
	TTFB     time.Duration
	Usage    llm.Usage
	Routed   string
	Score    int
	Errs     int
	Warns    int
	Title    string
	PlanPath string
}

func main() {
	var (
		theme     = flag.String("theme", defaultTheme, "thème de la campagne (identique pour tous les runs comparés)")
		name      = flag.String("name", defaultName, "nom de l'aventure factice")
		desc      = flag.String("description", "", "description de l'aventure factice")
		duration  = flag.String("duration", defaultDuration, "durée: oneshot, short ou campaign")
		advType   = flag.String("type", "", "type d'aventure (escort, mystery, heist...) — vide = choix libre")
		runs      = flag.Int("runs", 1, "nombre de générations")
		outDir    = flag.String("out", defaultOutDir, "répertoire de sortie des plans générés")
		timeout   = flag.Duration("timeout", campaign.DefaultTimeout, "budget temps par génération")
		noImage   = flag.Bool("no-image", false, "ne pas joindre la carte du monde (modèles sans support d'image; la description géographique reste dans le prompt)")
		reasoning = flag.String("reasoning", campaign.DefaultReasoningEffort, "effort de raisonnement: none|minimal|low|medium|high|xhigh|max")
	)
	flagArgs, positional := splitFlagsAndPositionals(os.Args[1:])
	_ = flag.CommandLine.Parse(flagArgs)

	if len(positional) != 1 {
		fmt.Fprintln(os.Stderr, "Erreur: précise exactement un modèle OpenRouter (format provider/model).")
		fmt.Fprintln(os.Stderr, "Exemples: ./sw-bench-campaign anthropic/claude-sonnet-5")
		fmt.Fprintln(os.Stderr, "          ./sw-bench-campaign z-ai/glm-5.3 --runs 2")
		os.Exit(2)
	}

	// A benchmark run must name the exact model under test: reject the
	// sonnet/haiku/opus aliases by requiring the full provider/model form
	// on the raw input (ParseModelChoice would resolve aliases silently).
	if !strings.Contains(positional[0], "/") {
		fmt.Fprintf(os.Stderr, "Erreur: modèle %q invalide — le format OpenRouter complet est requis (ex: anthropic/claude-sonnet-5, z-ai/glm-5.3).\n", positional[0])
		os.Exit(2)
	}
	model, err := llm.ParseModelChoice(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: modèle %q invalide — le format OpenRouter complet est requis (ex: anthropic/claude-sonnet-5, z-ai/glm-5.3).\n", positional[0])
		os.Exit(2)
	}

	cfg, err := llm.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}
	client, err := llm.NewOpenRouterClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: création du client OpenRouter: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: création du répertoire de sortie: %v\n", err)
		os.Exit(1)
	}

	// One shared prompt for every model and run: same inputs, same persona,
	// same world resources — differences in output come from the model only.
	adv := adventure.New(*name, strings.TrimSpace(*desc))
	req, err := campaign.BuildRequest(adv, strings.TrimSpace(*theme), *duration, strings.TrimSpace(*advType), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: %v\n", err)
		os.Exit(1)
	}

	// Persist the exact prompt for traceability of the benchmark inputs.
	if err := os.WriteFile(filepath.Join(*outDir, "prompt.txt"), []byte(req.Prompt), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: écriture du prompt: %v\n", err)
		os.Exit(1)
	}

	// The wizard attaches the world map as a base64 image; models without
	// vision endpoints reject the request outright. --no-image drops the
	// attachment while keeping the textual geography section of the prompt.
	if *noImage && req.WorldResources != nil {
		req.WorldResources.MapImageBase64 = ""
		req.WorldResources.MapImageMediaType = ""
	}
	req.ReasoningEffort = *reasoning

	fmt.Printf("Benchmark de génération de plan de campagne\n")
	fmt.Printf("  modèle        : %s\n", model)
	fmt.Printf("  durée         : %s (%d actes, %d-%d sessions)\n", *duration, req.Duration.Acts, req.Duration.MinSessions, req.Duration.MaxSessions)
	fmt.Printf("  runs          : %d, timeout %s\n", *runs, *timeout)
	fmt.Printf("  raisonnement  : %s\n", req.ReasoningEffort)
	fmt.Printf("  sortie        : %s\n\n", *outDir)

	results := make([]runResult, 0, *runs)
	for run := 1; run <= *runs; run++ {
		fmt.Printf("→ %s (run %d/%d)...\n", model, run, *runs)
		r := runOnce(client, req, model, run, *outDir, *timeout)
		if r.Err != nil {
			fmt.Printf("  ✗ %v\n", r.Err)
			if strings.Contains(r.Err.Error(), "support image input") && !*noImage {
				fmt.Printf("  ℹ ce modèle n'accepte pas d'image en entrée — réessaie avec --no-image\n")
			}
		} else {
			fmt.Printf("  ✓ %s — %s | %.1fs (TTFB %.1fs) | %d→%d tok | $%.4f | score %d/100\n",
				r.Routed, r.Title,
				r.Elapsed.Seconds(), r.TTFB.Seconds(),
				r.Usage.PromptTokens, r.Usage.CompletionTokens, r.Usage.Cost, r.Score)
		}
		results = append(results, r)
	}

	printTable(results)

	if allFailed(results) {
		os.Exit(1)
	}
}

// runOnce executes one benchmark run and persists the generated plan.
func runOnce(client llm.Client, req *campaign.Request, model string, run int, outDir string, timeout time.Duration) runResult {
	r := runResult{Model: model, Run: run}

	benchReq := *req
	benchReq.Model = model

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := campaign.Generate(ctx, client, &benchReq, nil)
	if err != nil {
		r.Err = err
		// Keep the raw output of parse failures for offline analysis.
		var parseErr *campaign.ParseError
		if errors.As(err, &parseErr) {
			rawPath := filepath.Join(outDir, fmt.Sprintf("%s-run%d.raw.txt", sanitize(model), run))
			if werr := os.WriteFile(rawPath, []byte(parseErr.Raw), 0644); werr == nil {
				fmt.Printf("  (raw output saved to %s)\n", rawPath)
			}
		}
		return r
	}

	r.Elapsed = result.Elapsed
	r.TTFB = result.TimeToFirstDelta
	r.Usage = result.Response.Usage
	r.Routed = result.Response.Model
	if r.Routed == "" {
		r.Routed = model
	}
	r.Title = result.Plan.Metadata.CampaignTitle

	validation := adventure.ValidateCampaignPlan(result.Plan)
	r.Score = validation.Score
	r.Errs = len(validation.Errors)
	r.Warns = len(validation.Warnings)

	planPath := filepath.Join(outDir, fmt.Sprintf("%s-run%d.json", sanitize(model), run))
	if data, err := json.MarshalIndent(result.Plan, "", "  "); err == nil {
		if err := os.WriteFile(planPath, data, 0644); err == nil {
			r.PlanPath = planPath
		}
	}

	return r
}

func printTable(results []runResult) {
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "MODEL\tRUN\tTOTAL\tTTFB\tPROMPT\tCOMPL\tREASON\tCACHED\tCOST\tSCORE\tERR/WARN\tPLAN")
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "%s\t%d\tERROR\t%s\n", r.Model, r.Run, shortErr(r.Err))
			continue
		}
		fmt.Fprintf(w, "%s\t%d\t%.1fs\t%.1fs\t%d\t%d\t%d\t%d\t$%.4f\t%d\t%d/%d\t%s\n",
			r.Model, r.Run,
			r.Elapsed.Seconds(), r.TTFB.Seconds(),
			r.Usage.PromptTokens, r.Usage.CompletionTokens,
			r.Usage.ReasoningTokens, r.Usage.CachedTokens,
			r.Usage.Cost,
			r.Score, r.Errs, r.Warns,
			r.Title)
	}
	_ = w.Flush()
}

func allFailed(results []runResult) bool {
	for _, r := range results {
		if r.Err == nil {
			return false
		}
	}
	return true
}

// shortErr keeps the table readable: the llm.Error format already packs the
// essential diagnosis.
func shortErr(err error) string {
	msg := err.Error()
	if len(msg) > 80 {
		msg = msg[:77] + "..."
	}
	return msg
}

// sanitize makes a model ID safe as a filename (anthropic/claude-sonnet-5 ->
// anthropic__claude-sonnet-5).
func sanitize(model string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', ':':
			return '_'
		default:
			return r
		}
	}, model)
}

// splitFlagsAndPositionals allows the natural usage
// "sw-bench-campaign z-ai/glm-5.3 --runs 2" even though the stdlib flag
// package stops parsing at the first positional argument. Value-taking flags
// (either --flag=value or --flag value) consume their following token;
// boolean flags don't.
func splitFlagsAndPositionals(args []string) (flags, positional []string) {
	// Boolean flags never consume the next token.
	boolFlags := map[string]bool{"no-image": true, "help": true, "h": true}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 1 && strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			name := strings.TrimLeft(arg, "-")
			isBool := boolFlags[name] && !strings.Contains(arg, "=")
			if !isBool && !strings.Contains(arg, "=") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positional = append(positional, arg)
	}
	return flags, positional
}
