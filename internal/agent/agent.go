// Package agent implements the Dungeon Master agent loop using Anthropic API.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Agent orchestrates the Dungeon Master agent loop.
type Agent struct {
	client          anthropic.Client
	model           anthropic.Model
	toolRegistry    *ToolRegistry
	conversationCtx *ConversationContext
	adventureCtx    *AdventureContext
	outputHandler   OutputHandler
	logger          *Logger
	personaLoader   *PersonaLoader
	agentManager    *AgentManager
	personaMetadata *PersonaMetadata
	systemGuidance  string // Hidden campaign/session briefing injected into system context
}

// New creates a new agent with the given configuration.
func New(apiKey string, adventureCtx *AdventureContext, outputHandler OutputHandler) (*Agent, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if adventureCtx == nil {
		return nil, fmt.Errorf("adventure context is required")
	}
	if outputHandler == nil {
		return nil, fmt.Errorf("output handler is required")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

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

	// Initialize agent manager
	agentManager := NewAgentManager(apiKey, adventureCtx, logger, outputHandler, personaLoader)

	// Initialize tool registry with adventure context
	toolRegistry := NewToolRegistry(adventureCtx)

	// Register all tools - pass Adventure object, agentManager, and outputHandler for real persistence
	if err := registerAllTools(toolRegistry, "data", adventureCtx.Adventure, agentManager, outputHandler); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	// Set the main tool registry in the agent manager for nested agent tool filtering
	agentManager.SetMainToolRegistry(toolRegistry)

	conversationCtx := NewConversationContext()

	agent := &Agent{
		client:          client,
		model:           anthropic.ModelClaudeHaiku4_5,
		toolRegistry:    toolRegistry,
		conversationCtx: conversationCtx,
		adventureCtx:    adventureCtx,
		outputHandler:   outputHandler,
		logger:          logger,
		personaLoader:   personaLoader,
		agentManager:    agentManager,
		personaMetadata: personaMetadata,
	}

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
	// Log user message
	if a.logger != nil {
		a.logger.LogUserMessage(message)
	}

	// Add user message to conversation history
	a.conversationCtx.AddUserMessage(message)

	// Build system prompt with DM persona and adventure context
	systemPrompt, err := a.buildSystemPrompt()
	if err != nil {
		return fmt.Errorf("failed to build system prompt: %w", err)
	}

	// Prepare messages for API call
	messages := a.conversationCtx.GetMessages()

	// Convert tools to Anthropic format
	toolsParam := a.toolRegistry.ToAnthropicToolsParam()

	// Call API with streaming and tools in a loop
	for {
		toolUses, assistantContent, err := a.callAnthropicAPI(systemPrompt, messages, toolsParam)
		if err != nil {
			if a.logger != nil {
				a.logger.LogError("API call", err)
			}
			return fmt.Errorf("API call failed: %w", err)
		}

		// Log assistant content if present
		if a.logger != nil && assistantContent != "" {
			a.logger.LogAssistantResponse(assistantContent)
		}

		// If no tool uses, we're done
		if len(toolUses) == 0 {
			// Add assistant response to conversation history
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

		// Update messages for next iteration
		messages = a.conversationCtx.GetMessages()

		// Continue loop to get final response with tool results
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

	systemPrompt := dmPersona + "\n\n" + adventureInfo + "\n" + postJournalReminder

	// Add system guidance if available (campaign briefing, hidden from player)
	if a.systemGuidance != "" {
		systemPrompt += "\n\n" + a.systemGuidance
	}

	return systemPrompt, nil
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

// callAnthropicAPI calls the Anthropic API with streaming and returns tool uses if any.
func (a *Agent) callAnthropicAPI(systemPrompt string, messages []anthropic.MessageParam, tools []anthropic.ToolUnionParam) ([]ToolUse, string, error) {
	// Log system prompt to file (only on first call)
	if len(a.conversationCtx.GetMessages()) == 1 {
		if err := os.WriteFile("system-prompt.log", []byte(systemPrompt), 0644); err != nil {
			// Non-fatal: just log to stderr
			fmt.Fprintf(os.Stderr, "Warning: Could not write system prompt to log: %v\n", err)
		}
	}

	// Create streaming message
	stream := a.client.Messages.NewStreaming(context.Background(), anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: 16384, // Haiku 4.5
		System: []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: systemPrompt,
			},
		},
		Messages: messages,
		Tools:    tools,
	})

	// Process streaming events
	streamHandler := NewStreamHandler(a.outputHandler)
	toolUses, assistantContent, err := streamHandler.ProcessStream(stream)
	if err != nil {
		return nil, "", fmt.Errorf("stream processing failed: %w", err)
	}

	return toolUses, assistantContent, nil
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
				Content:   fmt.Sprintf(`{"success": false, "error": "Tool not found: %s"}`, use.Name),
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
				Content:   fmt.Sprintf(`{"success": false, "error": "%s"}`, err.Error()),
				IsError:   true,
			})
			continue
		}

		// Log tool result
		if a.logger != nil {
			a.logger.LogToolResult(use.Name, use.ID, result)
		}

		// Check for system_brief in tool result (hidden campaign guidance)
		if resultMap, ok := result.(map[string]interface{}); ok {
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
		resultJSON := formatToolResult(result)
		results = append(results, ToolResultMessage{
			ToolUseID: use.ID,
			Content:   resultJSON,
			IsError:   false,
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
		"log_event": true,
		"add_gold":  true,
		"add_item":  true,
	}
	return modifyingTools[toolName]
}

// formatToolResult converts a tool result to JSON string.
func formatToolResult(result interface{}) string {
	// If result is already a map with success field, use it directly
	if m, ok := result.(map[string]interface{}); ok {
		// Try to marshal to JSON
		b, err := json.Marshal(m)
		if err == nil {
			return string(b)
		}

		// Fallback: simple serialization
		if success, ok := m["success"].(bool); ok {
			if !success {
				if errMsg, ok := m["error"].(string); ok {
					return fmt.Sprintf(`{"success": false, "error": "%s"}`, errMsg)
				}
			}
			// For successful results, include display field if present
			if display, ok := m["display"].(string); ok {
				// Escape quotes in display
				display = strings.ReplaceAll(display, `"`, `\"`)
				return fmt.Sprintf(`{"success": true, "display": "%s"}`, display)
			}
		}
	}

	// Fallback: return generic success
	return `{"success": true}`
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
