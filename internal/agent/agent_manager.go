// Package agent implements the Dungeon Master agent loop using Anthropic API.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClientFactory is a function type that creates an Anthropic client.
// This allows for dependency injection of mock clients in tests.
type ClientFactory func(apiKey string) anthropicClient

// anthropicClient is an interface that matches the Anthropic client's Messages service.
type anthropicClient interface {
	GetMessages() messagesService
}

// messagesService is an interface for the Messages.New method.
type messagesService interface {
	New(ctx context.Context, params anthropic.MessageNewParams, opts ...option.RequestOption) (*anthropic.Message, error)
}

// realAnthropicClient wraps the real Anthropic SDK client.
type realAnthropicClient struct {
	client anthropic.Client
}

func (r *realAnthropicClient) GetMessages() messagesService {
	return &r.client.Messages
}

// AgentManager manages multiple nested agent instances with stateful conversation contexts.
type AgentManager struct {
	nestedAgents     map[string]*NestedAgentState
	anthropicKey     string
	adventureCtx     *AdventureContext
	logger           *Logger
	outputHandler    OutputHandler
	personaLoader    *PersonaLoader
	mainToolRegistry *ToolRegistry      // Main agent's tool registry, used to create filtered registries
	maxDepth         int                // Maximum nesting depth (always 1 for now)
	clientFactory    ClientFactory      // Factory for creating Anthropic clients (allows mocking)
	worldResources   *WorldResources    // World map description + image for world-keeper
}

// NestedAgentState represents a nested agent with its own conversation context.
type NestedAgentState struct {
	agentName       string
	personaPath     string
	personaContent  string
	personaMetadata *PersonaMetadata
	conversationCtx *ConversationContext
	lastInvoked     time.Time
	invocationCount int
	client          anthropicClient  // Changed to interface for testability
	tokenLimit      int
	metrics         *AgentMetrics
	model           anthropic.Model  // Model to use for this agent (from persona)
	toolRegistry    *ToolRegistry    // Filtered tool registry for this agent
	toolPolicy      *ToolAccessPolicy // Tool access policy for this agent
}

// AgentMetrics tracks performance metrics for an agent.
type AgentMetrics struct {
	TotalTokensUsed    int64         `json:"total_tokens_used"`
	TotalInputTokens   int64         `json:"total_input_tokens"`
	TotalOutputTokens  int64         `json:"total_output_tokens"`
	TotalResponseTime  time.Duration `json:"total_response_time"`
	AverageTokensPerCall int64       `json:"average_tokens_per_call"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	ModelUsed          string        `json:"model_used"`
	LastCallTokens     int64         `json:"last_call_tokens"`
	LastCallDuration   time.Duration `json:"last_call_duration"`
}

// defaultClientFactory creates a real Anthropic client.
func defaultClientFactory(apiKey string) anthropicClient {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &realAnthropicClient{client: client}
}

// NewAgentManager creates a new AgentManager for managing nested agents.
func NewAgentManager(
	apiKey string,
	adventureCtx *AdventureContext,
	logger *Logger,
	outputHandler OutputHandler,
	personaLoader *PersonaLoader,
) *AgentManager {
	return &AgentManager{
		nestedAgents:     make(map[string]*NestedAgentState),
		anthropicKey:     apiKey,
		adventureCtx:     adventureCtx,
		logger:           logger,
		outputHandler:    outputHandler,
		personaLoader:    personaLoader,
		mainToolRegistry: nil, // Will be set via SetMainToolRegistry
		maxDepth:         1,   // Nested agents cannot invoke other agents
		clientFactory:    defaultClientFactory, // Use real client by default
		worldResources:   LoadWorldResources(),
	}
}

// SetMainToolRegistry sets the main tool registry from which nested agent registries are derived.
// This should be called after the main agent's tool registry is set up.
func (am *AgentManager) SetMainToolRegistry(registry *ToolRegistry) {
	am.mainToolRegistry = registry
}

// NewAgentManagerWithClientFactory creates an AgentManager with a custom client factory.
// This is primarily used for testing with mock clients.
func NewAgentManagerWithClientFactory(
	apiKey string,
	adventureCtx *AdventureContext,
	logger *Logger,
	outputHandler OutputHandler,
	personaLoader *PersonaLoader,
	clientFactory ClientFactory,
) *AgentManager {
	return &AgentManager{
		nestedAgents:     make(map[string]*NestedAgentState),
		anthropicKey:     apiKey,
		adventureCtx:     adventureCtx,
		logger:           logger,
		outputHandler:    outputHandler,
		personaLoader:    personaLoader,
		mainToolRegistry: nil, // Will be set via SetMainToolRegistry
		maxDepth:         1,
		clientFactory:    clientFactory,
		worldResources:   LoadWorldResources(),
	}
}

// InvokeAgent invokes a specialized agent with a question and optional context.
// The agent runs with its own tool loop (up to MaxIterations from policy) and
// uses the model specified in its persona.
// Returns the agent's response or an error.
func (am *AgentManager) InvokeAgent(agentName, question, contextInfo string, depth int) (string, error) {
	startTime := time.Now()

	// Validate recursion depth
	if depth > am.maxDepth {
		return "", &ErrRecursionLimit{
			AgentName:    agentName,
			CurrentDepth: depth,
			MaxDepth:     am.maxDepth,
			CallChain:    []string{"dungeon-master", agentName},
		}
	}

	// Validate agent name
	validAgents := []string{"character-creator", "rules-keeper", "world-keeper"}
	if !slices.Contains(validAgents, agentName) {
		return "", &ErrAgentNotFound{
			AgentName:       agentName,
			AvailableAgents: validAgents,
		}
	}

	// Notify output handler (shows "[Consulting <agent>...]" message)
	if am.outputHandler != nil {
		am.outputHandler.OnAgentInvocationStart(agentName)
	}

	// Get or create nested agent
	nestedAgent, err := am.getOrCreateNestedAgent(agentName)
	if err != nil {
		return "", fmt.Errorf("failed to get/create agent %s: %w", agentName, err)
	}

	// Build user message
	var messageContent string
	if contextInfo != "" {
		messageContent = fmt.Sprintf("%s\n\nContext: %s", question, contextInfo)
	} else {
		messageContent = question
	}

	// Add user message to conversation context
	nestedAgent.conversationCtx.AddUserMessage(messageContent)

	// Build system prompt with agent persona + adventure context
	systemPrompt := am.buildNestedAgentSystemPrompt(nestedAgent)

	// Get max iterations from policy (default to 5)
	maxIterations := 5
	if nestedAgent.toolPolicy != nil {
		maxIterations = nestedAgent.toolPolicy.MaxIterations
	}

	// Prepare tools (may be nil if agent has no tools)
	var toolsParam []anthropic.ToolUnionParam
	hasTools := nestedAgent.toolRegistry != nil && nestedAgent.toolRegistry.Count() > 0
	if hasTools {
		toolsParam = nestedAgent.toolRegistry.ToAnthropicToolsParam()
	}

	// Create API call context with timeout
	// Use 80 seconds (1m20s) to give nested agents more time for complex queries
	const invocationTimeout = 120 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), invocationTimeout)
	defer cancel()

	var finalResponseText string
	var totalInputTokens, totalOutputTokens int64

	// Agent loop with tool execution
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Build API request
		params := anthropic.MessageNewParams{
			Model:     nestedAgent.model,
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{
					Type: "text",
					Text: systemPrompt,
				},
			},
			Messages: nestedAgent.conversationCtx.GetMessages(),
		}

		// Add tools if available
		if hasTools {
			params.Tools = toolsParam
		}

		// Call Anthropic API
		response, err := nestedAgent.client.GetMessages().New(ctx, params)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return "", &ErrAgentTimeout{
					AgentName: agentName,
					Timeout:   invocationTimeout,
				}
			}
			return "", &AgentError{
				AgentName: agentName,
				Operation: "API call",
				Err:       err,
			}
		}

		// Track tokens
		totalInputTokens += int64(response.Usage.InputTokens)
		totalOutputTokens += int64(response.Usage.OutputTokens)

		// Process response content
		var textContent string
		var toolUses []ToolUse

		for _, block := range response.Content {
			switch contentBlock := block.AsAny().(type) {
			case anthropic.TextBlock:
				textContent += contentBlock.Text
			case anthropic.ToolUseBlock:
				// Parse tool input
				var input map[string]interface{}
				if err := json.Unmarshal(contentBlock.Input, &input); err != nil {
					input = make(map[string]interface{})
				}
				toolUses = append(toolUses, ToolUse{
					ID:    contentBlock.ID,
					Name:  contentBlock.Name,
					Input: input,
				})
			}
		}

		// If no tool uses, we're done
		if len(toolUses) == 0 {
			finalResponseText = textContent
			nestedAgent.conversationCtx.AddAssistantMessage(textContent)
			break
		}

		// Add assistant message with tool uses to conversation
		nestedAgent.conversationCtx.AddAssistantMessageWithToolUses(textContent, toolUses)

		// Execute tools
		toolResults := am.executeNestedAgentTools(nestedAgent, toolUses)

		// Add tool results to conversation
		nestedAgent.conversationCtx.AddToolResults(toolResults)

		// Log tool calls
		if am.logger != nil {
			for _, use := range toolUses {
				am.logger.LogInfo(fmt.Sprintf("[%s] Tool call: %s", agentName, use.Name))
			}
		}

		// Continue loop to get response with tool results
	}

	if finalResponseText == "" {
		return "", fmt.Errorf("agent %s returned empty response after %d iterations", agentName, maxIterations)
	}

	// Calculate metrics
	duration := time.Since(startTime)
	totalTokens := totalInputTokens + totalOutputTokens

	// Update agent state
	nestedAgent.lastInvoked = time.Now()
	nestedAgent.invocationCount++

	// Update metrics
	nestedAgent.metrics.TotalTokensUsed += totalTokens
	nestedAgent.metrics.TotalInputTokens += totalInputTokens
	nestedAgent.metrics.TotalOutputTokens += totalOutputTokens
	nestedAgent.metrics.TotalResponseTime += duration
	nestedAgent.metrics.LastCallTokens = totalTokens
	nestedAgent.metrics.LastCallDuration = duration

	// Calculate averages
	if nestedAgent.invocationCount > 0 {
		nestedAgent.metrics.AverageTokensPerCall = nestedAgent.metrics.TotalTokensUsed / int64(nestedAgent.invocationCount)
		nestedAgent.metrics.AverageResponseTime = nestedAgent.metrics.TotalResponseTime / time.Duration(nestedAgent.invocationCount)
	}

	// Log the invocation
	if am.logger != nil {
		invocationID := fmt.Sprintf("agent_%d", nestedAgent.invocationCount)
		am.logger.LogAgentInvocation(agentName, invocationID, question, contextInfo, finalResponseText, duration, int(totalTokens))
	}

	// Notify output handler completion
	if am.outputHandler != nil {
		am.outputHandler.OnAgentInvocationComplete(agentName, duration)
	}

	return finalResponseText, nil
}

// executeNestedAgentTools executes tools for a nested agent and returns results.
func (am *AgentManager) executeNestedAgentTools(agent *NestedAgentState, toolUses []ToolUse) []ToolResultMessage {
	results := make([]ToolResultMessage, 0, len(toolUses))

	for _, use := range toolUses {
		// Get tool from agent's filtered registry
		tool, exists := agent.toolRegistry.Get(use.Name)
		if !exists {
			if am.logger != nil {
				am.logger.LogInfo(fmt.Sprintf("[%s] Tool not found: %s", agent.agentName, use.Name))
			}
			results = append(results, ToolResultMessage{
				ToolUseID: use.ID,
				Content:   fmt.Sprintf(`{"success": false, "error": "Tool not found: %s"}`, use.Name),
				IsError:   true,
			})
			continue
		}

		// Execute tool
		result, err := tool.Execute(use.Input)
		if err != nil {
			if am.logger != nil {
				am.logger.LogInfo(fmt.Sprintf("[%s] Tool %s error: %v", agent.agentName, use.Name, err))
			}
			results = append(results, ToolResultMessage{
				ToolUseID: use.ID,
				Content:   fmt.Sprintf(`{"success": false, "error": "%s"}`, err.Error()),
				IsError:   true,
			})
			continue
		}

		// Convert result to JSON string
		resultJSON, err := json.Marshal(result)
		if err != nil {
			// Log serialization failure and return a warning so the agent knows the result couldn't be serialized
			if am.logger != nil {
				am.logger.LogInfo(fmt.Sprintf("[%s] Tool %s result not serializable: %v", agent.agentName, use.Name, err))
			}
			resultJSON = []byte(fmt.Sprintf(`{"success": true, "warning": "result not serializable: %s"}`, err.Error()))
		}

		results = append(results, ToolResultMessage{
			ToolUseID: use.ID,
			Content:   string(resultJSON),
			IsError:   false,
		})
	}

	return results
}

// InvokeAgentSilent invokes a specialized agent and returns response without extensive logging.
// This is used for pre-session briefings where the full response should not be visible to players.
// The response is intended to be injected into system context only.
// Unlike InvokeAgent, this version uses a single API call without tool loop for faster responses.
func (am *AgentManager) InvokeAgentSilent(agentName, question string, depth int) (string, error) {
	startTime := time.Now()

	// Validate recursion depth
	if depth > am.maxDepth {
		return "", &ErrRecursionLimit{
			AgentName:    agentName,
			CurrentDepth: depth,
			MaxDepth:     am.maxDepth,
			CallChain:    []string{"dungeon-master", agentName},
		}
	}

	// Validate agent name
	validAgents := []string{"character-creator", "rules-keeper", "world-keeper"}
	if !slices.Contains(validAgents, agentName) {
		return "", &ErrAgentNotFound{
			AgentName:       agentName,
			AvailableAgents: validAgents,
		}
	}

	// Notify output handler with brief message only
	if am.outputHandler != nil {
		am.outputHandler.OnAgentInvocationStart(agentName)
	}

	// Get or create nested agent
	nestedAgent, err := am.getOrCreateNestedAgent(agentName)
	if err != nil {
		return "", fmt.Errorf("failed to get/create agent %s: %w", agentName, err)
	}

	// Add user message to conversation context
	nestedAgent.conversationCtx.AddUserMessage(question)

	// Build system prompt with agent persona + adventure context
	systemPrompt := am.buildNestedAgentSystemPrompt(nestedAgent)

	// Create API call context with timeout
	const invocationTimeout = 120 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), invocationTimeout)
	defer cancel()

	// Call Anthropic API with NO TOOLS for silent mode (faster, simpler response)
	// Silent mode is used for briefings where we don't need tool execution
	response, err := nestedAgent.client.GetMessages().New(ctx, anthropic.MessageNewParams{
		Model:     nestedAgent.model, // Use model from persona
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: systemPrompt,
			},
		},
		Messages: nestedAgent.conversationCtx.GetMessages(),
		// Tools parameter intentionally omitted for silent mode
	})

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", &ErrAgentTimeout{
				AgentName: agentName,
				Timeout:   invocationTimeout,
			}
		}
		return "", &AgentError{
			AgentName: agentName,
			Operation: "API call (silent)",
			Err:       err,
		}
	}

	// Extract text content from response
	var responseText string
	for _, block := range response.Content {
		switch contentBlock := block.AsAny().(type) {
		case anthropic.TextBlock:
			responseText += contentBlock.Text
		}
	}

	if responseText == "" {
		return "", fmt.Errorf("agent %s returned empty response", agentName)
	}

	// Add assistant response to conversation context
	nestedAgent.conversationCtx.AddAssistantMessage(responseText)

	// Calculate metrics
	duration := time.Since(startTime)
	inputTokens := int64(response.Usage.InputTokens)
	outputTokens := int64(response.Usage.OutputTokens)
	totalTokens := inputTokens + outputTokens

	// Update agent state
	nestedAgent.lastInvoked = time.Now()
	nestedAgent.invocationCount++

	// Update metrics
	nestedAgent.metrics.TotalTokensUsed += totalTokens
	nestedAgent.metrics.TotalInputTokens += inputTokens
	nestedAgent.metrics.TotalOutputTokens += outputTokens
	nestedAgent.metrics.TotalResponseTime += duration
	nestedAgent.metrics.LastCallTokens = totalTokens
	nestedAgent.metrics.LastCallDuration = duration

	// Calculate averages
	if nestedAgent.invocationCount > 0 {
		nestedAgent.metrics.AverageTokensPerCall = nestedAgent.metrics.TotalTokensUsed / int64(nestedAgent.invocationCount)
		nestedAgent.metrics.AverageResponseTime = nestedAgent.metrics.TotalResponseTime / time.Duration(nestedAgent.invocationCount)
	}

	// Log the invocation (minimal logging for silent mode)
	if am.logger != nil {
		invocationID := fmt.Sprintf("agent_%d_silent", nestedAgent.invocationCount)
		am.logger.LogAgentInvocation(agentName, invocationID, question, "(silent mode)", "[response hidden]", duration, int(totalTokens))
	}

	// Notify output handler completion
	if am.outputHandler != nil {
		am.outputHandler.OnAgentInvocationComplete(agentName, duration)
	}

	return responseText, nil
}

// getOrCreateNestedAgent gets an existing nested agent or creates a new one.
func (am *AgentManager) getOrCreateNestedAgent(agentName string) (*NestedAgentState, error) {
	// Check if agent already exists
	if agent, exists := am.nestedAgents[agentName]; exists {
		return agent, nil
	}

	// Load agent persona
	metadata, personaBody, err := am.personaLoader.LoadWithMetadata(agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to load persona: %w", err)
	}

	// Create Anthropic client using factory (allows mocking in tests)
	client := am.clientFactory(am.anthropicKey)

	// Create conversation context with reduced token limit for nested agents
	const nestedAgentTokenLimit = 20000 // Lower than main agent's 50K
	conversationCtx := NewConversationContextWithLimit(nestedAgentTokenLimit)

	// Map persona model to Anthropic model
	// If persona doesn't specify a model, defaults to Sonnet
	model := MapPersonaModelToAnthropic(metadata.Model)
	modelDisplayName := GetModelDisplayName(model)

	// Get tool access policy for this agent
	policy := GetPolicyForAgent(agentName)

	// Create filtered tool registry for this agent
	var filteredRegistry *ToolRegistry
	if am.mainToolRegistry != nil && policy != nil {
		filteredRegistry = am.mainToolRegistry.CreateFilteredRegistry(
			policy.GetAllowedToolNames(),
			policy.ForbiddenTools,
		)
		// Log tool configuration
		if am.logger != nil && filteredRegistry.Count() > 0 {
			am.logger.LogInfo(fmt.Sprintf("[%s] Initialized with %d tools: %v",
				agentName, filteredRegistry.Count(), filteredRegistry.Names()))
		}
	}

	// Create nested agent state
	agent := &NestedAgentState{
		agentName:       agentName,
		personaPath:     fmt.Sprintf("core_agents/agents/%s.md", agentName),
		personaContent:  personaBody, // Store body without frontmatter
		personaMetadata: metadata,
		conversationCtx: conversationCtx,
		lastInvoked:     time.Now(),
		invocationCount: 0,
		client:          client,
		tokenLimit:      nestedAgentTokenLimit,
		model:           model,
		toolRegistry:    filteredRegistry,
		toolPolicy:      policy,
		metrics: &AgentMetrics{
			ModelUsed: modelDisplayName, // Use actual model from persona
		},
	}

	// Inject world map image for world-keeper as initial conversation context
	if agentName == "world-keeper" && am.worldResources != nil && am.worldResources.MapImageBase64 != "" {
		agent.conversationCtx.AddUserMessageWithImage(
			"Voici la carte du monde des Quatre Royaumes. Utilise-la comme référence géographique pour toutes tes validations.",
			am.worldResources.MapImageBase64,
			am.worldResources.MapImageMediaType,
		)
		agent.conversationCtx.AddAssistantMessage(
			"J'ai bien reçu la carte du monde des Quatre Royaumes. Je l'utiliserai comme référence pour assurer la cohérence géographique de l'aventure.",
		)
	}

	// Store in map
	am.nestedAgents[agentName] = agent

	return agent, nil
}

// buildNestedAgentSystemPrompt builds the system prompt for a nested agent.
// Combines the agent's persona with relevant adventure context.
func (am *AgentManager) buildNestedAgentSystemPrompt(agent *NestedAgentState) string {
	var sb strings.Builder

	// Agent persona
	sb.WriteString(agent.personaContent)
	sb.WriteString("\n\n")

	// Add adventure context (read-only information)
	sb.WriteString("## Current Adventure Context\n\n")
	sb.WriteString(fmt.Sprintf("**Adventure**: %s\n", am.adventureCtx.Adventure.Name))
	sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", am.adventureCtx.Adventure.Description))
	sb.WriteString(fmt.Sprintf("**Party**: %s\n", formatParty(am.adventureCtx)))
	sb.WriteString(fmt.Sprintf("**Gold**: %d gp\n", am.adventureCtx.Inventory.Gold))
	sb.WriteString(fmt.Sprintf("**Current Location**: %s\n\n", am.adventureCtx.State.CurrentLocation))

	// Inject world map description for world-keeper
	if agent.agentName == "world-keeper" && am.worldResources != nil && am.worldResources.MapDescription != "" {
		sb.WriteString("\n## World Map Reference\n\n")
		sb.WriteString("Use this detailed geographical description of the Four Kingdoms as your authoritative reference ")
		sb.WriteString("for all geography, distances, trade routes, and location validation:\n\n")
		sb.WriteString(am.worldResources.MapDescription)
		sb.WriteString("\n\n")
	}

	// Add constraint for nested agents
	sb.WriteString("**Important**: You are a specialized consultant agent. ")
	sb.WriteString("You cannot modify game state or invoke other agents. ")
	sb.WriteString("Provide clear, expert guidance based on your specialization.\n")

	return sb.String()
}

// GetNestedAgentState returns the state of a nested agent if it exists.
func (am *AgentManager) GetNestedAgentState(agentName string) (*NestedAgentState, bool) {
	agent, exists := am.nestedAgents[agentName]
	return agent, exists
}

// ListNestedAgents returns a list of all active nested agent names.
func (am *AgentManager) ListNestedAgents() []string {
	agents := make([]string, 0, len(am.nestedAgents))
	for name := range am.nestedAgents {
		agents = append(agents, name)
	}
	return agents
}

// ClearNestedAgent removes a nested agent and its conversation history.
// Useful for resetting an agent's memory.
func (am *AgentManager) ClearNestedAgent(agentName string) {
	delete(am.nestedAgents, agentName)
}

// ClearAllNestedAgents removes all nested agents.
func (am *AgentManager) ClearAllNestedAgents() {
	am.nestedAgents = make(map[string]*NestedAgentState)
}

// GetStatistics returns statistics about nested agent usage.
func (am *AgentManager) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_agents"] = len(am.nestedAgents)

	agentStats := make(map[string]map[string]interface{})
	for name, agent := range am.nestedAgents {
		agentStats[name] = map[string]interface{}{
			"invocation_count":        agent.invocationCount,
			"last_invoked":           agent.lastInvoked.Format(time.RFC3339),
			"message_count":          len(agent.conversationCtx.GetMessages()),
			"token_estimate":         agent.conversationCtx.tokenEstimate,
			"total_tokens_used":      agent.metrics.TotalTokensUsed,
			"total_input_tokens":     agent.metrics.TotalInputTokens,
			"total_output_tokens":    agent.metrics.TotalOutputTokens,
			"average_tokens_per_call": agent.metrics.AverageTokensPerCall,
			"total_response_time_ms": agent.metrics.TotalResponseTime.Milliseconds(),
			"average_response_time_ms": agent.metrics.AverageResponseTime.Milliseconds(),
			"model_used":             agent.metrics.ModelUsed,
			"last_call_tokens":       agent.metrics.LastCallTokens,
			"last_call_duration_ms":  agent.metrics.LastCallDuration.Milliseconds(),
		}
	}
	stats["agents"] = agentStats

	return stats
}

// GetAgentMetrics returns detailed metrics for a specific agent.
func (am *AgentManager) GetAgentMetrics(agentName string) (*AgentMetrics, bool) {
	agent, exists := am.nestedAgents[agentName]
	if !exists {
		return nil, false
	}
	return agent.metrics, true
}

