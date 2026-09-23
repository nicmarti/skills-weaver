// Package agent implements the Dungeon Master agent loop on the provider-neutral llm boundary.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"dungeons/internal/llm"
)

// Agent orchestrates the Dungeon Master agent loop.
type Agent struct {
	client llm.Client
	// model is the resolved OpenRouter model ID (Sonnet 5 default). Atomic:
	// written by the web model selector (HTTP goroutine) while the agent
	// loop reads it mid-turn.
	model           atomic.Pointer[string]
	toolRegistry    *ToolRegistry
	conversationCtx *ConversationContext
	adventureCtx    *AdventureContext
	outputHandler   OutputHandler
	logger          *Logger
	personaLoader   *PersonaLoader
	agentManager    *AgentManager
	personaMetadata *PersonaMetadata
	systemGuidance  string // Hidden campaign/session briefing injected into system context
	// turnUsage aggregates provider usage/cost across the tool loop of one turn.
	lastTurnUsage llm.Usage
	// sessionID is the stable per-adventure routing hint sent to OpenRouter.
	sessionID string
}

// mainAgentContextTokenLimit is the local estimated-history budget for the main
// DM agent (Sonnet 5 default). This is a local character-based estimate, not
// provider tokens; TruncateIfNeeded drops the oldest turns to stay under it.
const mainAgentContextTokenLimit = 900000

// mainAgentMaxOutputTokens is the per-request output budget for the main DM.
const mainAgentMaxOutputTokens = 32000

// mainAgentMaxToolIterations bounds the streamed tool loop per user turn so a
// model that keeps requesting tools cannot loop forever.
const mainAgentMaxToolIterations = 40

// New creates a new agent from the shared OpenRouter configuration.
// The DM model comes from cfg.ModelDM (Sonnet 5 default; OPENROUTER_MODEL_DM
// or a future CLI flag overrides it). The persona's legacy `model:` field is
// metadata only on the migrated runtime and no longer selects the model.
func New(cfg llm.Config, adventureCtx *AdventureContext, outputHandler OutputHandler) (*Agent, error) {
	if adventureCtx == nil {
		return nil, fmt.Errorf("adventure context is required")
	}
	if outputHandler == nil {
		return nil, fmt.Errorf("output handler is required")
	}

	client, err := llm.NewOpenRouterClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRouter client: %w", err)
	}
	return newWithClient(client, cfg, adventureCtx, outputHandler)
}

// NewWithClient creates an agent backed by an existing llm.Client (tests inject
// a fake here). cfg provides the resolved DM model.
func NewWithClient(client llm.Client, cfg llm.Config, adventureCtx *AdventureContext, outputHandler OutputHandler) (*Agent, error) {
	if client == nil {
		return nil, fmt.Errorf("llm client is required")
	}
	return newWithClient(client, cfg, adventureCtx, outputHandler)
}

func newWithClient(client llm.Client, cfg llm.Config, adventureCtx *AdventureContext, outputHandler OutputHandler) (*Agent, error) {
	model := cfg.ModelDM
	if model == "" {
		model = llm.DefaultModelDM
	}

	// Initialize persona loader
	personaLoader := NewPersonaLoader()

	// Load persona metadata for version tracking
	personaMetadata, _, err := personaLoader.LoadWithMetadata("dungeon-master")
	if err != nil {
		return nil, fmt.Errorf("failed to load dungeon-master persona: %w", err)
	}

	// Initialize logger
	logger, err := NewLogger(adventureCtx.BasePath())
	if err != nil {
		// Non-fatal: continue without logging
		fmt.Printf("Warning: Could not create logger: %v\n", err)
	}

	// Initialize agent manager: nested agents share the main agent's neutral
	// client and per-agent model resolution from the same config.
	agentManager := NewAgentManager(client, cfg, adventureCtx, logger, outputHandler, personaLoader)

	// Initialize tool registry with adventure context
	toolRegistry := NewToolRegistry(adventureCtx)

	// Register all tools - pass Adventure object, agentManager, and outputHandler for real persistence
	if err := registerAllTools(toolRegistry, "data", adventureCtx.Adventure, agentManager, outputHandler, client, cfg); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	// Set the main tool registry in the agent manager for nested agent tool filtering
	agentManager.SetMainToolRegistry(toolRegistry)

	conversationCtx := NewConversationContextWithLimit(mainAgentContextTokenLimit)

	agent := &Agent{
		client:          client,
		toolRegistry:    toolRegistry,
		conversationCtx: conversationCtx,
		adventureCtx:    adventureCtx,
		outputHandler:   outputHandler,
		logger:          logger,
		personaLoader:   personaLoader,
		agentManager:    agentManager,
		personaMetadata: personaMetadata,
		sessionID:       fmt.Sprintf("sw-dm-%s", adventureCtx.Adventure.Slug),
	}
	agent.model.Store(&model)

	// Load agent states from previous sessions
	statesPath := fmt.Sprintf("%s/agent-states.json", adventureCtx.BasePath())
	if err := agentManager.LoadAgentStates(statesPath); err != nil {
		// Non-fatal: log warning and continue with fresh state
		fmt.Printf("Warning: Could not load agent states: %v\n", err)
	}

	return agent, nil
}

// ProcessUserMessage processes a user message and returns the agent's response.
func (a *Agent) ProcessUserMessage(message string) error {
	return a.ProcessUserMessageContext(context.Background(), message)
}

// ProcessUserMessageContext processes a user message with cancellation support.
// ctx bounds the whole turn: generation and every tool-loop iteration.
func (a *Agent) ProcessUserMessageContext(ctx context.Context, message string) error {
	// Log user message
	if a.logger != nil {
		a.logger.LogUserMessage(message)
	}

	// === SESSION GUARD-RAIL ===
	// A player message must never be processed outside an active session, or its
	// log_event entries leak into journal-session-0 (see AGENTS.md). start_session
	// is otherwise purely LLM-discretionary, and the DM may skip it on resume — so
	// we enforce it deterministically here, at the common chokepoint for sw-web and
	// sw-dm. This reuses the exact start_session tool logic (StartSession + coherence
	// gate + world-keeper briefing) so the resumed turn still gets its hidden brief.
	a.ensureSessionStarted()

	// Add user message to conversation history
	a.conversationCtx.AddUserMessage(message)

	// Build system prompt with DM persona and adventure context
	systemPrompt, err := a.buildSystemPrompt()
	if err != nil {
		return fmt.Errorf("failed to build system prompt: %w", err)
	}

	definitions, err := a.toolRegistry.ToolDefinitions()
	if err != nil {
		return fmt.Errorf("prepare tool definitions: %w", err)
	}

	// Bounded streamed tool loop.
	for iteration := 0; iteration < mainAgentMaxToolIterations; iteration++ {
		resp, err := a.callLLM(ctx, systemPrompt, definitions)
		if err != nil {
			if a.logger != nil {
				a.logger.LogError("API call", err)
			}
			// Do not commit partial assistant/tool-call history and never retry
			// after observable output: side effects of executed tools are kept,
			// but this turn is not replayed automatically.
			return fmt.Errorf("API call failed: %w", err)
		}
		a.aggregateUsage(resp)

		assistantContent := resp.Message.Text
		toolUses := toolUsesFromCalls(resp.Message.ToolCalls)

		// Log assistant content if present
		if a.logger != nil && assistantContent != "" {
			a.logger.LogAssistantResponse(assistantContent)
		}

		// Handle every finish reason explicitly. Incomplete finishes end the
		// turn without committing partial assistant history.
		switch resp.FinishReason {
		case llm.FinishLength, llm.FinishContentFilter, llm.FinishError:
			err := fmt.Errorf("generation ended with finish reason %q", resp.FinishReason)
			if a.logger != nil {
				a.logger.LogError("API call", err)
			}
			a.outputHandler.OnError(err)
			a.outputHandler.OnComplete()
			a.saveAgentStates()
			return nil
		case llm.FinishStop, llm.FinishToolCalls, llm.FinishNone:
			// Normal completion paths, handled below.
		default:
			err := fmt.Errorf("generation ended with unexpected finish reason %q", resp.FinishReason)
			if a.logger != nil {
				a.logger.LogError("API call", err)
			}
			a.outputHandler.OnError(err)
			a.outputHandler.OnComplete()
			a.saveAgentStates()
			return nil
		}

		// No tool calls: the turn is complete.
		if len(toolUses) == 0 {
			a.conversationCtx.AddAssistantMessage(assistantContent)
			a.outputHandler.OnComplete()

			// Save agent states after processing message
			a.saveAgentStates()

			return nil
		}

		// Add assistant message with tool uses to history
		a.conversationCtx.AddAssistantMessageWithToolUses(assistantContent, toolUses)

		// Execute tools
		toolResults := a.executeTools(toolUses)

		// Add tool results to conversation
		a.conversationCtx.AddToolResults(toolResults)

		// Continue loop to get final response with tool results
	}

	err = fmt.Errorf("tool loop exceeded %d iterations", mainAgentMaxToolIterations)
	if a.logger != nil {
		a.logger.LogError("API call", err)
	}
	a.outputHandler.OnError(err)
	a.outputHandler.OnComplete()
	return nil
}

// toolUsesFromCalls converts neutral streamed tool calls to executable ToolUses.
func toolUsesFromCalls(calls []llm.ToolCall) []ToolUse {
	if len(calls) == 0 {
		return nil
	}
	toolUses := make([]ToolUse, 0, len(calls))
	for _, call := range calls {
		var input map[string]interface{}
		if len(call.Arguments) > 0 {
			if err := json.Unmarshal(call.Arguments, &input); err != nil {
				input = map[string]interface{}{}
			}
		}
		toolUses = append(toolUses, ToolUse{ID: call.ID, Name: call.Name, Input: input})
	}
	return toolUses
}

// aggregateUsage accumulates provider usage/cost across the tool loop of a turn.
func (a *Agent) aggregateUsage(resp llm.Response) {
	u := resp.Usage
	if !u.Present {
		return
	}
	a.lastTurnUsage.PromptTokens += u.PromptTokens
	a.lastTurnUsage.CompletionTokens += u.CompletionTokens
	a.lastTurnUsage.TotalTokens += u.TotalTokens
	a.lastTurnUsage.CachedTokens += u.CachedTokens
	a.lastTurnUsage.CacheWriteTokens += u.CacheWriteTokens
	a.lastTurnUsage.ReasoningTokens += u.ReasoningTokens
	a.lastTurnUsage.Cost += u.Cost
	a.lastTurnUsage.Present = true
	if a.logger != nil {
		a.logger.LogInfo(fmt.Sprintf(
			"turn usage so far: prompt=%d completion=%d cached=%d cost=%.6f (routed model=%s, generation=%s)",
			a.lastTurnUsage.PromptTokens, a.lastTurnUsage.CompletionTokens, a.lastTurnUsage.CachedTokens,
			a.lastTurnUsage.Cost, resp.Model, resp.ID))
	}
}

// LastTurnUsage returns aggregated usage for the most recent turn.
func (a *Agent) LastTurnUsage() llm.Usage {
	return a.lastTurnUsage
}

// ensureSessionStarted is the deterministic session guard-rail. If no game session
// is active, it auto-invokes the registered start_session tool before the player's
// message is processed, then applies the tool's hidden campaign briefing the same way
// executeTools does. Best-effort: any failure is logged and never blocks the turn
// (the DM can still call start_session itself — it becomes a harmless no-op).
func (a *Agent) ensureSessionStarted() {
	if a.adventureCtx == nil || a.adventureCtx.Adventure == nil {
		return
	}
	// Disk-fresh check: GetCurrentSession reloads sessions.json, so this reflects the
	// real state even across process restarts. (nil, nil) means no active session.
	cur, err := a.adventureCtx.Adventure.GetCurrentSession()
	if err != nil || cur != nil {
		return
	}

	tool, ok := a.toolRegistry.Get("start_session")
	if !ok {
		return
	}
	result, err := tool.Execute(map[string]interface{}{})
	if err != nil {
		if a.logger != nil {
			a.logger.LogError("session guard-rail: start_session", err)
		}
		return
	}
	if a.logger != nil {
		a.logger.LogInfo("session guard-rail: no active session — auto-started before processing player message")
	}
	// Apply the hidden campaign briefing, mirroring executeTools' handling of a
	// start_session result so the resumed turn is guided like a normal session start.
	if resultMap, ok := result.(map[string]interface{}); ok {
		if systemBrief, ok := resultMap["system_brief"].(string); ok && systemBrief != "" {
			a.AddSystemGuidance(systemBrief)
		}
	}
}

// buildSystemPrompt constructs the system prompt with DM persona and adventure context.
func (a *Agent) buildSystemPrompt() (string, error) {
	// Load DM persona using PersonaLoader (searches core_agents/agents/, then .claude/agents/)
	dmPersona, err := a.personaLoader.Load("dungeon-master")
	if err != nil {
		return "", fmt.Errorf("failed to load dungeon-master persona: %w", err)
	}

	// Build adventure context
	adventureInfo := fmt.Sprintf(`
## Contexte de l'Aventure Actuelle

**Aventure** : %s
%s

**Groupe de PJ (contrôlés par le joueur)** : %s
**Or** : %d po
**Lieu actuel** : %s

**Journal récent** (jusqu'à 20 dernières entrées) :
%s
`,
		a.adventureCtx.Adventure.Name,
		a.adventureCtx.Adventure.Description,
		formatParty(a.adventureCtx),
		a.adventureCtx.Inventory.Gold,
		a.adventureCtx.State.CurrentLocation,
		formatRecentJournal(a.adventureCtx),
	)

	// Post-journal reminder to counter recency bias
	postJournalReminder := `
=== RAPPEL CRITIQUE APRÈS LECTURE DU JOURNAL ===

Le journal ci-dessus montre des événements PASSÉS de cette aventure.

**TYPES D'ENTRÉES AUTOMATIQUES** (générées par tools) :
  • [xp] : Créé automatiquement par add_xp
  • [loot] : Créé automatiquement par generate_treasure
  • [combat] : Certains créés automatiquement par update_hp

**TYPES D'ENTRÉES MANUELLES** (TU DOIS appeler log_event) :
  • [story] : Événements narratifs (dialogues, décisions, découvertes)
  • [npc] : Rencontres de PNJ clés, alliances, trahisons
  • [discovery] : Révélations importantes, indices critiques
  • [quest] : Nouveaux objectifs, changements de plan

⚠️ SANS log_event régulier pour événements narratifs, le contexte sera PERDU au rechargement.

**APPELER log_event MAINTENANT si le joueur vient de** :
  • Recevoir information critique d'un PNJ
  • Prendre décision stratégique
  • Découvrir indice ou lieu important
  • Faire alliance ou trahison
  • Terminer combat (même si update_hp a créé entrée automatique)

========================
`

	// Load campaign plan directive (narrative guardrails, always present)
	campaignDirective := a.buildCampaignDirective()
	if campaignDirective != "" {
		adventureInfo += "\n" + campaignDirective
	}

	systemPrompt := dmPersona + "\n\n" + adventureInfo + "\n" + postJournalReminder

	// Add system guidance if available (campaign briefing, hidden from player)
	if a.systemGuidance != "" {
		systemPrompt += "\n\n" + a.systemGuidance
	}

	return systemPrompt, nil
}

// buildCampaignDirective loads the campaign plan and builds a compact narrative directive
// that is always included in the system prompt. This ensures the DM never operates without
// knowing the planned storyline, even if start_session is not called.
func (a *Agent) buildCampaignDirective() string {
	plan, err := a.adventureCtx.Adventure.LoadCampaignPlan()
	if err != nil || plan == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("=== PLAN NARRATIF (OBLIGATOIRE - NE PAS DÉVIER) ===\n")

	// Campaign objective
	if plan.NarrativeStructure.Objective != "" {
		b.WriteString(fmt.Sprintf("Objectif: %s\n", plan.NarrativeStructure.Objective))
	}

	// Current act
	if act := plan.GetCurrentAct(); act != nil {
		b.WriteString(fmt.Sprintf("Acte %d: %s — %s\n", act.Number, act.Title, act.Description))
		if len(act.Goals) > 0 {
			b.WriteString("Objectifs:\n")
			for _, goal := range act.Goals {
				b.WriteString(fmt.Sprintf("  • %s\n", goal))
			}
		}
	}

	// Antagonist
	antag := plan.PlotElements.Antagonist
	if antag.Name != "" {
		b.WriteString(fmt.Sprintf("Antagoniste: %s (%s, %s)\n", antag.Name, antag.Role, antag.Motivation))
	}

	// Key locations
	if len(plan.PlotElements.KeyLocations) > 0 {
		names := make([]string, 0, len(plan.PlotElements.KeyLocations))
		for _, loc := range plan.PlotElements.KeyLocations {
			names = append(names, loc.Name)
		}
		b.WriteString(fmt.Sprintf("Lieux clés: %s\n", strings.Join(names, ", ")))
	}

	b.WriteString("\n⚠️ Tu DOIS suivre ce plan narratif. N'invente PAS de nouveaux antagonistes, cultes ou intrigues absents de ce plan.\n")
	b.WriteString("===")

	return b.String()
}

// AddSystemGuidance injects hidden campaign/session briefing into system context.
// This is used for pre-session briefings from world-keeper that should guide DM narration
// without being directly visible to players.
func (a *Agent) AddSystemGuidance(guidance string) {
	a.systemGuidance = guidance
}

// ClearSystemGuidance removes the current system guidance.
// Useful when guidance is only relevant for current session.
func (a *Agent) ClearSystemGuidance() {
	a.systemGuidance = ""
}

// dmStreamObserver forwards streamed text deltas to the output handler.
type dmStreamObserver struct {
	handler OutputHandler
}

func (o dmStreamObserver) OnTextDelta(text string) {
	o.handler.OnTextChunk(text)
}

// callLLM performs one streamed provider turn over the neutral boundary.
func (a *Agent) callLLM(ctx context.Context, systemPrompt string, definitions []llm.ToolDefinition) (llm.Response, error) {
	// Log system prompt to file (only on first call)
	if len(a.conversationCtx.NeutralMessages()) == 1 {
		if err := os.WriteFile("system-prompt.log", []byte(systemPrompt), 0644); err != nil {
			// Non-fatal: just log to stderr
			fmt.Fprintf(os.Stderr, "Warning: Could not write system prompt to log: %v\n", err)
		}
	}

	// Log model being used for this API call
	model := a.currentModel()
	if a.logger != nil {
		a.logger.LogInfo(fmt.Sprintf("API call using model: %s", llm.DisplayName(model)))
	}

	req := llm.Request{
		Model:               model,
		System:              systemPrompt,
		Messages:            a.conversationCtx.NeutralMessages(),
		Tools:               definitions,
		MaxCompletionTokens: mainAgentMaxOutputTokens,
		SessionID:           a.sessionID,
		// Tools must not be silently dropped by endpoints that do not support
		// them; fail routing loudly instead.
		RequireParameters: len(definitions) > 0,
	}

	return a.client.Stream(ctx, req, dmStreamObserver{handler: a.outputHandler})
}

// executeTools executes all tool uses and returns the results.
func (a *Agent) executeTools(toolUses []ToolUse) []ToolResultMessage {
	results := []ToolResultMessage{}
	stateModified := false

	for _, use := range toolUses {
		// Log tool call
		if a.logger != nil {
			a.logger.LogToolCall(use.Name, use.ID, use.Input)
			// Log equivalent CLI command if available
			if cliCmd := ToolToCLICommand(use.Name, use.Input); cliCmd != "" {
				a.logger.LogCLICommand(cliCmd)
			}
		}

		// Notify output handler
		a.outputHandler.OnToolStart(use.Name, use.ID)

		// Execute tool
		tool, exists := a.toolRegistry.Get(use.Name)
		if !exists {
			a.outputHandler.OnError(fmt.Errorf("tool not found: %s", use.Name))
			errorResult := map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Tool not found: %s", use.Name),
			}
			if a.logger != nil {
				a.logger.LogToolResult(use.Name, use.ID, errorResult)
			}
			results = append(results, ToolResultMessage{
				ToolUseID: use.ID,
				Content:   formatToolResult(errorResult),
				IsError:   true,
			})
			continue
		}

		result, err := tool.Execute(use.Input)
		if err != nil {
			a.outputHandler.OnError(err)
			errorResult := map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
			if a.logger != nil {
				a.logger.LogToolResult(use.Name, use.ID, errorResult)
			}
			results = append(results, ToolResultMessage{
				ToolUseID: use.ID,
				Content: formatToolResult(map[string]interface{}{
					"success": false, "error": err.Error(),
				}),
				IsError: true,
			})
			continue
		}

		// Log tool result
		if a.logger != nil {
			a.logger.LogToolResult(use.Name, use.ID, result)
		}

		isError := false
		// Check for system_brief in tool result (hidden campaign guidance)
		if resultMap, ok := result.(map[string]interface{}); ok {
			if success, ok := resultMap["success"].(bool); ok && !success {
				isError = true
			}
			if systemBrief, ok := resultMap["system_brief"].(string); ok && systemBrief != "" {
				// Inject system briefing into agent context (hidden from player)
				a.AddSystemGuidance(systemBrief)

				// Log that guidance was injected
				if a.logger != nil {
					a.logger.LogInfo(fmt.Sprintf("[SYSTEM] Injected campaign briefing into agent context (%d chars)", len(systemBrief)))
				}
			}
		}

		// Convert result to JSON string
		resultJSON, encodingFailed := encodeToolResult(result)
		results = append(results, ToolResultMessage{
			ToolUseID: use.ID,
			Content:   resultJSON,
			IsError:   isError || encodingFailed,
		})

		// Notify completion
		a.outputHandler.OnToolComplete(use.Name, result)

		// Check if state was modified
		if isStateModifyingTool(use.Name) {
			stateModified = true
		}
	}

	// Reload adventure context if state was modified
	if stateModified {
		if err := a.adventureCtx.Reload(); err != nil {
			a.outputHandler.OnError(fmt.Errorf("failed to reload adventure context: %w", err))
		}
	}

	return results
}

// formatParty formats the party members for display.
func formatParty(ctx *AdventureContext) string {
	if len(ctx.Party.Characters) == 0 {
		return "Aucun personnage"
	}

	parts := []string{}
	for _, charName := range ctx.Party.Characters {
		// Find character in loaded characters
		for _, char := range ctx.Characters {
			if char.Name == charName {
				parts = append(parts, fmt.Sprintf("%s (%s %s)", char.Name, char.Species, char.Class))
				break
			}
		}
	}

	return strings.Join(parts, ", ")
}

// formatRecentJournal formats recent journal entries for display.
func formatRecentJournal(ctx *AdventureContext) string {
	if len(ctx.RecentJournal) == 0 {
		return "Aucune entrée récente"
	}

	parts := []string{}
	// Take last 20 entries (or all if less than 20)
	start := 0
	if len(ctx.RecentJournal) > 20 {
		start = len(ctx.RecentJournal) - 20
	}

	for i := start; i < len(ctx.RecentJournal); i++ {
		entry := ctx.RecentJournal[i]
		parts = append(parts, fmt.Sprintf("- [%s] %s", entry.Type, entry.Content))
	}

	return strings.Join(parts, "\n")
}

// isStateModifyingTool returns true if the tool modifies adventure state.
func isStateModifyingTool(toolName string) bool {
	modifyingTools := map[string]bool{
		"log_event":             true,
		"add_gold":              true,
		"add_item":              true,
		"remove_item":           true,
		"update_time":           true,
		"update_location":       true,
		"set_flag":              true,
		"add_quest":             true,
		"complete_quest":        true,
		"set_variable":          true,
		"update_hp":             true,
		"use_spell_slot":        true,
		"add_xp":                true,
		"generate_npc":          true,
		"update_npc_importance": true,
		"update_character_stat": true,
		"long_rest":             true,
	}
	return modifyingTools[toolName]
}

// formatToolResult always produces valid JSON, including for a result that
// cannot itself be marshaled. Never interpolate unescaped tool/error text.
func formatToolResult(result interface{}) string {
	encoded, _ := encodeToolResult(result)
	return encoded
}

func encodeToolResult(result interface{}) (string, bool) {
	if result == nil {
		result = map[string]interface{}{"success": true}
	}
	data, err := json.Marshal(result)
	if err != nil {
		data, _ = json.Marshal(map[string]interface{}{
			"success": false, "error": fmt.Sprintf("tool result is not JSON serializable: %v", err),
		})
		return string(data), true
	}
	return string(data), false
}

// AgentManager exposes the nested-agent manager so callers (e.g. the coherence
// narrative judgment) can invoke specialist agents via InvokeAgentSilent.
func (a *Agent) AgentManager() *AgentManager {
	return a.agentManager
}

// saveAgentStates saves the current agent states to disk.
// This is called after each user message to persist nested agent conversation history.
func (a *Agent) saveAgentStates() {
	if a.agentManager == nil {
		return
	}

	statesPath := fmt.Sprintf("%s/agent-states.json", a.adventureCtx.BasePath())
	if err := a.agentManager.SaveAgentStates(statesPath); err != nil {
		// Non-fatal: log error but don't crash
		fmt.Printf("Warning: Could not save agent states: %v\n", err)
		if a.logger != nil {
			a.logger.LogError("save_agent_states", err)
		}
	}
}

// SetModel changes the model used by the agent for API calls. The value must
// be an OpenRouter model ID or a sonnet/haiku/opus alias. Safe to call while
// a turn is streaming: the web model selector writes it from an HTTP handler
// goroutine while the agent loop reads it.
func (a *Agent) SetModel(model string) error {
	resolved, err := llm.ParseModelChoice(model)
	if err != nil {
		return fmt.Errorf("invalid model selection: %w", err)
	}
	oldModel := a.currentModel()
	a.model.Store(&resolved)
	if a.logger != nil {
		a.logger.LogInfo(fmt.Sprintf("Model changed: %s -> %s", llm.DisplayName(oldModel), llm.DisplayName(resolved)))
	}
	return nil
}

// GetModel returns the current model used by the agent.
func (a *Agent) GetModel() string {
	return a.currentModel()
}

// currentModel loads the resolved model ID atomically; Agent embeds an
// atomic value and must not be copied.
func (a *Agent) currentModel() string {
	if m := a.model.Load(); m != nil {
		return *m
	}
	return llm.DefaultModelDM
}

// GetPersonaVersion returns the version of the loaded persona.
func (a *Agent) GetPersonaVersion() string {
	if a.personaMetadata == nil {
		return "unknown"
	}
	if a.personaMetadata.Version == "" {
		return "unversioned"
	}
	return a.personaMetadata.Version
}

// GetPersonaName returns the name of the loaded persona.
func (a *Agent) GetPersonaName() string {
	if a.personaMetadata == nil {
		return "unknown"
	}
	return a.personaMetadata.Name
}
