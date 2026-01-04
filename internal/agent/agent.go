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

	// Initialize tool registry with adventure context
	toolRegistry := NewToolRegistry(adventureCtx)

	// Register all tools - pass Adventure object for real persistence
	if err := registerAllTools(toolRegistry, "data", adventureCtx.Adventure); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	conversationCtx := NewConversationContext()

	// Initialize logger
	logger, err := NewLogger(adventureCtx.BasePath())
	if err != nil {
		// Non-fatal: continue without logging
		fmt.Printf("Warning: Could not create logger: %v\n", err)
	}

	return &Agent{
		client:          client,
		model:           anthropic.ModelClaudeHaiku4_5,
		toolRegistry:    toolRegistry,
		conversationCtx: conversationCtx,
		adventureCtx:    adventureCtx,
		outputHandler:   outputHandler,
		logger:          logger,
	}, nil
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
	systemPrompt := a.buildSystemPrompt()

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
func (a *Agent) buildSystemPrompt() string {
	// Load DM persona from file
	dmPersona, err := os.ReadFile(".claude/agents/dungeon-master.md")
	if err != nil {
		dmPersona = []byte("Tu es le Maître du Donjon pour des parties de Basic Fantasy RPG.")
	}

	// Build adventure context
	adventureInfo := fmt.Sprintf(`
## Contexte de l'Aventure Actuelle

**Aventure** : %s
%s

**Groupe de PJ (contrôlés par le joueur)** : %s

**RAPPEL CRITIQUE** :
- "Que faites-vous ?" s'adresse au joueur pour ses PJ uniquement
- Les PNJ (tous les autres personnages) sont contrôlés par TOI
- Ne propose JAMAIS d'options numérotées (1, 2, 3...)
- Ne demande JAMAIS ce que fait un PNJ au joueur

**Or** : %d po
**Lieu actuel** : %s

**Journal récent** (5 dernières entrées) :
%s

## Tools Disponibles

Tu as accès aux tools suivants pour gérer la partie :

- **roll_dice** : Lance des dés avec notation RPG (d20, 2d6+3, 4d6kh3)
- **get_monster** : Consulte les stats d'un monstre par son ID
- **log_event** : Enregistre un événement dans le journal de l'aventure
- **add_gold** : Modifie l'or du groupe
- **get_inventory** : Consulte l'inventaire partagé
- **generate_treasure** : Génère un trésor selon les tables BFRPG
- **generate_npc** : Crée un PNJ complet avec traits et personnalité
- **generate_image** : Génère une image de style fantasy à partir d'un prompt détaillé

## Instructions Importantes

- Utilise **log_event** pour enregistrer les événements significatifs
- Mets à jour l'or et l'inventaire quand approprié
- Lance les dés pour tous les jets de hasard
- Utilise **generate_image** pour créer des illustrations de scènes importantes
- Reste dans le rôle du Maître du Jeu
- Narre au présent, en français
- Sois concis mais immersif
- Termine TOUJOURS par "Que faites-vous ?" sans proposer d'options
`,
		a.adventureCtx.Adventure.Name,
		a.adventureCtx.Adventure.Description,
		formatParty(a.adventureCtx),
		a.adventureCtx.Inventory.Gold,
		a.adventureCtx.State.CurrentLocation,
		formatRecentJournal(a.adventureCtx),
	)

	return string(dmPersona) + "\n\n" + adventureInfo
}

// callAnthropicAPI calls the Anthropic API with streaming and returns tool uses if any.
func (a *Agent) callAnthropicAPI(systemPrompt string, messages []anthropic.MessageParam, tools []anthropic.ToolUnionParam) ([]ToolUse, string, error) {
	// Create streaming message
	stream := a.client.Messages.NewStreaming(context.Background(), anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: 8192,
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
				parts = append(parts, fmt.Sprintf("%s (%s %s)", char.Name, char.Race, char.Class))
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
	// Take last 5 entries
	start := 0
	if len(ctx.RecentJournal) > 5 {
		start = len(ctx.RecentJournal) - 5
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
