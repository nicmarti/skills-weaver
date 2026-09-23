// Package agent implements the Dungeon Master agent and nested-agent orchestration.
package agent

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"dungeons/internal/llm"
)

// AgentManager manages multiple nested agent instances with stateful conversation contexts.
// All nested agents run on the shared provider-neutral llm.Client (OpenRouter
// Chat); per-agent models resolve from llm.Config.ModelForAgent.
type AgentManager struct {
	nestedAgents     map[string]*NestedAgentState
	client           llm.Client
	llmCfg           llm.Config
	adventureCtx     *AdventureContext
	logger           *Logger
	outputHandler    OutputHandler
	personaLoader    *PersonaLoader
	mainToolRegistry *ToolRegistry   // Main agent's tool registry, used to create filtered registries
	maxDepth         int             // Maximum nesting depth (always 1 for now)
	worldResources   *WorldResources // World map description + image for world-keeper
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
	client          llm.Client
	tokenLimit      int
	metrics         *AgentMetrics
	model           string           // Resolved OpenRouter model ID for this agent
	toolRegistry    *ToolRegistry    // Filtered tool registry for this agent
	toolPolicy      *ToolAccessPolicy // Tool access policy for this agent
}

// AgentMetrics tracks performance metrics for an agent.
type AgentMetrics struct {
	TotalTokensUsed      int64         `json:"total_tokens_used"`
	TotalInputTokens     int64         `json:"total_input_tokens"`
	TotalOutputTokens    int64         `json:"total_output_tokens"`
	TotalResponseTime    time.Duration `json:"total_response_time"`
	AverageTokensPerCall int64         `json:"average_tokens_per_call"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	ModelUsed            string        `json:"model_used"`
	LastCallTokens       int64         `json:"last_call_tokens"`
	LastCallDuration     time.Duration `json:"last_call_duration"`
	// Advisor tool metrics from the pre-migration Anthropic beta runtime
	// (historical values stay readable from old agent-states.json files).
	AdvisorCalls               int64  `json:"advisor_calls,omitempty"`
	AdvisorInputTokens         int64  `json:"advisor_input_tokens,omitempty"`
	AdvisorOutputTokens        int64  `json:"advisor_output_tokens,omitempty"`
	AdvisorCacheCreationTokens int64  `json:"advisor_cache_creation_tokens,omitempty"`
	AdvisorCacheReadTokens     int64  `json:"advisor_cache_read_tokens,omitempty"`
	AdvisorModelUsed           string `json:"advisor_model_used,omitempty"`
	// OpenRouter aggregate metrics (new in Phase 5). These are aggregate
	// request-scoped values; provider sub-calls such as server tools are not
	// separable here, unlike the historical Advisor fields above.
	RequestedModel   string  `json:"requested_model,omitempty"`
	RoutedModel      string  `json:"routed_model,omitempty"`
	TotalCost        float64 `json:"total_cost,omitempty"`
	CachedTokens     int64   `json:"cached_tokens,omitempty"`
	CacheWriteTokens int64   `json:"cache_write_tokens,omitempty"`
	ReasoningTokens  int64   `json:"reasoning_tokens,omitempty"`
}

// NewAgentManager creates a new AgentManager for managing nested agents on the
// given shared neutral client. cfg provides per-agent model resolution.
func NewAgentManager(
	client llm.Client,
	cfg llm.Config,
	adventureCtx *AdventureContext,
	logger *Logger,
	outputHandler OutputHandler,
	personaLoader *PersonaLoader,
) *AgentManager {
	return &AgentManager{
		nestedAgents:     make(map[string]*NestedAgentState),
		client:           client,
		llmCfg:           cfg,
		adventureCtx:     adventureCtx,
		logger:           logger,
		outputHandler:    outputHandler,
		personaLoader:    personaLoader,
		mainToolRegistry: nil, // Will be set via SetMainToolRegistry
		maxDepth:         1,   // Nested agents cannot invoke other agents
		worldResources:   LoadWorldResources(),
	}
}

// SetMainToolRegistry sets the main tool registry from which nested agent registries are derived.
// This should be called after the main agent's tool registry is set up.
func (am *AgentManager) SetMainToolRegistry(registry *ToolRegistry) {
	am.mainToolRegistry = registry
}

// NewAgentManagerWithTools builds an AgentManager with the full tool registry
// wired, so nested agents get their policy-filtered read-only tools (the same
// setup the main DM loop uses). Useful for tooling that invokes a nested agent
// outside the main loop (e.g. the advisor A/B harness).
func NewAgentManagerWithTools(cfg llm.Config, adventureCtx *AdventureContext, logger *Logger, outputHandler OutputHandler) (*AgentManager, error) {
	client, err := llm.NewOpenRouterClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRouter client: %w", err)
	}
	personaLoader := NewPersonaLoader()
	am := NewAgentManager(client, cfg, adventureCtx, logger, outputHandler, personaLoader)
	registry := NewToolRegistry(adventureCtx)
	if err := registerAllTools(registry, "data", adventureCtx.Adventure, am, outputHandler); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}
	am.SetMainToolRegistry(registry)
	return am, nil
}

// nestedAgentTokenLimit is the live estimated-history budget per nested agent.
const nestedAgentTokenLimit = 20000

// nestedAgentMaxOutputTokens bounds each nested provider call's output.
const nestedAgentMaxOutputTokens = 4096

// nestedSupportsImageInput reports whether the resolved model is known to
// accept image input. The check is deliberately conservative: Claude-family
// OpenRouter models are known-good; any other provider/model is assumed
// text-only so World Keeper degrades to its text map description rather than
// sending an invalid image payload. The model is never silently swapped.
func nestedSupportsImageInput(model string) bool {
	return strings.HasPrefix(strings.ToLower(model), "anthropic/") ||
		strings.EqualFold(model, "openrouter/auto")
}

// stripImages returns the messages without base64 image payloads, preserving
// text (the text-only-map fallback for image-incapable models).
func stripImages(messages []llm.Message) []llm.Message {
	stripped := make([]llm.Message, 0, len(messages))
	changed := false
	for _, msg := range messages {
		if len(msg.Images) == 0 {
			stripped = append(stripped, msg)
			continue
		}
		if msg.Text == "" {
			// Nothing textual remains; skip the image-only message entirely.
			changed = true
			continue
		}
		clone := msg
		clone.Images = nil
		stripped = append(stripped, clone)
		changed = true
	}
	if !changed {
		return messages
	}
	return stripped
}

// InvokeAgent invokes a specialized agent with a question and optional context.
// The agent runs its own bounded tool loop (up to MaxIterations from policy) on
// the shared neutral client, with its model resolved from llm.Config.
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

	// Prepare filtered read-only tools
	var definitions []llm.ToolDefinition
	hasTools := nestedAgent.toolRegistry != nil && nestedAgent.toolRegistry.Count() > 0
	if hasTools {
		definitions, err = nestedAgent.toolRegistry.ToolDefinitions()
		if err != nil {
			return "", fmt.Errorf("prepare %s tool definitions: %w", agentName, err)
		}
	}

	// Create API call context with timeout (preserved from the pre-migration loop)
	const invocationTimeout = 120 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), invocationTimeout)
	defer cancel()

	var finalResponseText string
	var aggregate llm.Usage
	aggregate.Present = true

	// Bounded agent loop with tool execution
	for iteration := 0; iteration < maxIterations; iteration++ {
		req := llm.Request{
			Model:               nestedAgent.model,
			System:              systemPrompt,
			Messages:            nestedAgent.conversationCtx.NeutralMessages(),
			MaxCompletionTokens: nestedAgentMaxOutputTokens,
		}
		if hasTools {
			req.Tools = definitions
			// Tools must not be silently dropped by endpoints that don't support them.
			req.RequireParameters = true
		}
		req.Messages = am.adaptMessagesForModel(nestedAgent, req.Messages)

		resp, callErr := am.client.Complete(ctx, req)
		if callErr != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return "", &ErrAgentTimeout{AgentName: agentName, Timeout: invocationTimeout}
			}
			return "", &AgentError{AgentName: agentName, Operation: "API call", Err: callErr}
		}
		am.recordCallMetrics(nestedAgent, resp, &aggregate)

		// Incomplete finishes abort the invocation without committing partial history.
		switch resp.FinishReason {
		case llm.FinishStop, llm.FinishToolCalls, llm.FinishNone:
			// Normal completion paths.
		default:
			return "", &AgentError{
				AgentName: agentName,
				Operation: fmt.Sprintf("API call (finish reason %q)", resp.FinishReason),
				Err:       fmt.Errorf("generation ended with finish reason %q", resp.FinishReason),
			}
		}

		toolUses := toolUsesFromCalls(resp.Message.ToolCalls)
		textContent := resp.Message.Text

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

	duration := time.Since(startTime)
	am.finalizeInvocationMetrics(nestedAgent, aggregate, duration)

	// Log the invocation
	if am.logger != nil {
		invocationID := fmt.Sprintf("agent_%d", nestedAgent.invocationCount)
		am.logger.LogAgentInvocation(agentName, invocationID, question, contextInfo, finalResponseText, duration, int(aggregate.PromptTokens+aggregate.CompletionTokens))
	}

	// Notify output handler completion
	if am.outputHandler != nil {
		am.outputHandler.OnAgentInvocationComplete(agentName, duration)
	}

	return finalResponseText, nil
}

// adaptMessagesForModel applies the text-only-map fallback for image-incapable
// models (World Keeper keeps its authoritative textual map description).
func (am *AgentManager) adaptMessagesForModel(nestedAgent *NestedAgentState, messages []llm.Message) []llm.Message {
	hasImages := false
	for _, msg := range messages {
		for _, img := range msg.Images {
			if img.Base64 != "" {
				hasImages = true
				break
			}
		}
	}
	if !hasImages || nestedSupportsImageInput(nestedAgent.model) {
		return messages
	}
	if am.logger != nil {
		am.logger.LogInfo(fmt.Sprintf("[%s] model %q is not known to support images; falling back to the textual map description only",
			nestedAgent.agentName, nestedAgent.model))
	}
	return stripImages(messages)
}

// recordCallMetrics folds one provider response's usage into the aggregate and
// the persisted OpenRouter metric fields.
func (am *AgentManager) recordCallMetrics(nestedAgent *NestedAgentState, resp llm.Response, aggregate *llm.Usage) {
	if resp.Model != "" {
		nestedAgent.metrics.RoutedModel = resp.Model
	}
	if !resp.Usage.Present {
		return
	}
	aggregate.PromptTokens += resp.Usage.PromptTokens
	aggregate.CompletionTokens += resp.Usage.CompletionTokens
	aggregate.TotalTokens += resp.Usage.TotalTokens
	aggregate.CachedTokens += resp.Usage.CachedTokens
	aggregate.CacheWriteTokens += resp.Usage.CacheWriteTokens
	aggregate.ReasoningTokens += resp.Usage.ReasoningTokens
	aggregate.Cost += resp.Usage.Cost
	aggregate.Present = true

	nestedAgent.metrics.CachedTokens += resp.Usage.CachedTokens
	nestedAgent.metrics.CacheWriteTokens += resp.Usage.CacheWriteTokens
	nestedAgent.metrics.ReasoningTokens += resp.Usage.ReasoningTokens
	nestedAgent.metrics.TotalCost += resp.Usage.Cost
}

// finalizeInvocationMetrics folds an invocation's aggregate usage into the
// persisted metrics and refreshes averages.
func (am *AgentManager) finalizeInvocationMetrics(nestedAgent *NestedAgentState, aggregate llm.Usage, duration time.Duration) {
	nestedAgent.lastInvoked = time.Now()
	nestedAgent.invocationCount++

	if aggregate.Present {
		totalTokens := aggregate.PromptTokens + aggregate.CompletionTokens
		if aggregate.TotalTokens > 0 {
			totalTokens = aggregate.TotalTokens
		}
		nestedAgent.metrics.TotalTokensUsed += totalTokens
		nestedAgent.metrics.TotalInputTokens += aggregate.PromptTokens
		nestedAgent.metrics.TotalOutputTokens += aggregate.CompletionTokens
		nestedAgent.metrics.LastCallTokens = totalTokens
	}
	nestedAgent.metrics.TotalResponseTime += duration
	nestedAgent.metrics.LastCallDuration = duration

	if nestedAgent.invocationCount > 0 {
		nestedAgent.metrics.AverageTokensPerCall = nestedAgent.metrics.TotalTokensUsed / int64(nestedAgent.invocationCount)
		nestedAgent.metrics.AverageResponseTime = nestedAgent.metrics.TotalResponseTime / time.Duration(nestedAgent.invocationCount)
	}
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
				Content: formatToolResult(map[string]interface{}{
					"success": false, "error": fmt.Sprintf("Tool not found: %s", use.Name),
				}),
				IsError: true,
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
				Content: formatToolResult(map[string]interface{}{
					"success": false, "error": err.Error(),
				}),
				IsError: true,
			})
			continue
		}

		isError := false
		if resultMap, ok := result.(map[string]interface{}); ok {
			if success, ok := resultMap["success"].(bool); ok && !success {
				isError = true
			}
		}
		// Use the same JSON encoding as the main agent, including failures.
		resultJSON, encodingFailed := encodeToolResult(result)
		if encodingFailed {
			if am.logger != nil {
				am.logger.LogInfo(fmt.Sprintf("[%s] Tool %s result not serializable", agent.agentName, use.Name))
			}
			isError = true
		}

		results = append(results, ToolResultMessage{
			ToolUseID: use.ID,
			Content:   resultJSON,
			IsError:   isError,
		})
	}

	return results
}

// InvokeAgentSilent invokes a specialized agent and returns response without extensive logging.
// This is used for pre-session briefings where the full response should not be visible to players.
// The response is intended to be injected into system context only.
// Unlike InvokeAgent, this version uses a single API call without tools for faster responses.
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

	// Single neutral call with NO TOOLS (faster, simpler). The native
	// OpenRouter advisor server tool is Phase 6 work and stays off here.
	resp, callErr := am.client.Complete(ctx, llm.Request{
		Model:               nestedAgent.model,
		System:              systemPrompt,
		Messages:            am.adaptMessagesForModel(nestedAgent, nestedAgent.conversationCtx.NeutralMessages()),
		MaxCompletionTokens:  nestedAgentMaxOutputTokens,
		RequireParameters:   false,
	})
	if callErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", &ErrAgentTimeout{AgentName: agentName, Timeout: invocationTimeout}
		}
		return "", &AgentError{AgentName: agentName, Operation: "API call (silent)", Err: callErr}
	}

	responseText := resp.Message.Text
	if responseText == "" {
		return "", fmt.Errorf("agent %s returned empty response", agentName)
	}

	// Incomplete finishes still produce text; keep it but record the reason.
	if resp.FinishReason != llm.FinishStop && resp.FinishReason != llm.FinishNone && am.logger != nil {
		am.logger.LogInfo(fmt.Sprintf("[%s] silent invocation ended with finish reason %q", agentName, resp.FinishReason))
	}

	// Add assistant response to conversation context
	nestedAgent.conversationCtx.AddAssistantMessage(responseText)

	var aggregate llm.Usage
	aggregate.Present = true
	am.recordCallMetrics(nestedAgent, resp, &aggregate)
	duration := time.Since(startTime)
	am.finalizeInvocationMetrics(nestedAgent, aggregate, duration)

	// Log the invocation (minimal logging for silent mode)
	if am.logger != nil {
		invocationID := fmt.Sprintf("agent_%d_silent", nestedAgent.invocationCount)
		am.logger.LogAgentInvocation(agentName, invocationID, question, "(silent mode)", "[response hidden]", duration, int(nestedAgent.metrics.LastCallTokens))
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

	// Resolve the model from the shared config: per-agent environment/CLI
	// override > nested default > Sonnet 5. The persona's `model:` field is
	// metadata only — it must never silently override the user's selection.
	model, err := am.llmCfg.ModelForAgent(agentName)
	if err != nil {
		return nil, fmt.Errorf("resolve model for %s: %w", agentName, err)
	}

	// Create conversation context with reduced token limit for nested agents
	conversationCtx := NewConversationContextWithLimit(nestedAgentTokenLimit)

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
		client:           am.client,
		tokenLimit:       nestedAgentTokenLimit,
		model:            model,
		toolRegistry:     filteredRegistry,
		toolPolicy:       policy,
		metrics: &AgentMetrics{
			ModelUsed:       model,
			RequestedModel:  model,
		},
	}

	// Inject world map image for world-keeper as initial conversation context
	if agentName == "world-keeper" && am.worldResources != nil && am.worldResources.MapImageBase64 != "" {
		agent.conversationCtx.AddUserMessageWithImageResource(
			"Voici la carte du monde des Quatre Royaumes. Utilise-la comme référence géographique pour toutes tes validations.",
			am.worldResources.MapImageBase64,
			am.worldResources.MapImageMediaType,
			"world-map",
		)
		agent.conversationCtx.AddAssistantMessage(
			"J'ai bien reçu la carte du monde des Quatre Royaumes. Je l'utiliserai comme référence pour assurer la cohérence géographique de l'aventure.",
		)
		// Non-vision models degrade to the textual map description only.
		if !nestedSupportsImageInput(model) && am.logger != nil {
			am.logger.LogInfo(fmt.Sprintf("[%s] model %q is not known to support images; the textual map description will be used instead",
				agentName, model))
		}
	}

	// Store in map
	am.nestedAgents[agentName] = agent

	return agent, nil
}

// LogInfo writes a diagnostic line to the adventure log when a logger is
// available. It lets best-effort callers (e.g. the session-start briefing)
// surface failures without exposing confidential content to players.
func (am *AgentManager) LogInfo(message string) {
	if am.logger != nil {
		am.logger.LogInfo(message)
	}
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

// Metrics returns the agent's performance/cost metrics (read-only accessor).
func (s *NestedAgentState) Metrics() *AgentMetrics {
	return s.metrics
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
			"invocation_count":         agent.invocationCount,
			"last_invoked":             agent.lastInvoked.Format(time.RFC3339),
			"message_count":            len(agent.conversationCtx.NeutralMessages()),
			"token_estimate":           agent.conversationCtx.tokenEstimate,
			"total_tokens_used":        agent.metrics.TotalTokensUsed,
			"total_input_tokens":       agent.metrics.TotalInputTokens,
			"total_output_tokens":      agent.metrics.TotalOutputTokens,
			"average_tokens_per_call":  agent.metrics.AverageTokensPerCall,
			"total_response_time_ms":   agent.metrics.TotalResponseTime.Milliseconds(),
			"average_response_time_ms": agent.metrics.AverageResponseTime.Milliseconds(),
			"model_used":               agent.metrics.ModelUsed,
			"requested_model":          agent.metrics.RequestedModel,
			"routed_model":             agent.metrics.RoutedModel,
			"total_cost":               agent.metrics.TotalCost,
			"last_call_tokens":         agent.metrics.LastCallTokens,
			"last_call_duration_ms":    agent.metrics.LastCallDuration.Milliseconds(),
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
