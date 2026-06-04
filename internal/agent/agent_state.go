// Package agent implements the Dungeon Master agent loop using Anthropic API.
package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AgentStatesFile represents the JSON structure for persisting agent states.
type AgentStatesFile struct {
	SessionID   int                         `json:"session_id"`
	LastUpdated string                      `json:"last_updated"`
	Agents      map[string]*SerializedAgent `json:"agents"`
}

// SerializedAgent represents a serialized nested agent state.
type SerializedAgent struct {
	InvocationCount     int                   `json:"invocation_count"`
	LastInvoked         string                `json:"last_invoked"`
	ConversationHistory []SerializableMessage `json:"conversation_history"`
	TokenEstimate       int                   `json:"token_estimate"`
	MaxTokens           int                   `json:"max_tokens"`
	Metrics             *SerializedMetrics    `json:"metrics"`
}

// SerializedMetrics represents serialized agent metrics.
type SerializedMetrics struct {
	TotalTokensUsed       int64  `json:"total_tokens_used"`
	TotalInputTokens      int64  `json:"total_input_tokens"`
	TotalOutputTokens     int64  `json:"total_output_tokens"`
	TotalResponseTimeMS   int64  `json:"total_response_time_ms"`
	AverageTokensPerCall  int64  `json:"average_tokens_per_call"`
	AverageResponseTimeMS int64  `json:"average_response_time_ms"`
	ModelUsed             string `json:"model_used"`
	LastCallTokens        int64  `json:"last_call_tokens"`
	LastCallDurationMS    int64  `json:"last_call_duration_ms"`
	// Advisor tool metrics (Opus-billed, tracked separately from executor tokens).
	AdvisorCalls               int64  `json:"advisor_calls,omitempty"`
	AdvisorInputTokens         int64  `json:"advisor_input_tokens,omitempty"`
	AdvisorOutputTokens        int64  `json:"advisor_output_tokens,omitempty"`
	AdvisorCacheCreationTokens int64  `json:"advisor_cache_creation_tokens,omitempty"`
	AdvisorCacheReadTokens     int64  `json:"advisor_cache_read_tokens,omitempty"`
	AdvisorModelUsed           string `json:"advisor_model_used,omitempty"`
}

// SaveAgentStates saves all nested agent states to a JSON file.
func (am *AgentManager) SaveAgentStates(filePath string) error {
	// Build agent states structure
	agents := make(map[string]*SerializedAgent)

	for name, state := range am.nestedAgents {
		// Persist up to the agent's own live token limit so the restored history
		// matches what the agent actually used during the session. Previously a
		// hardcoded 15K trimmed nested agents (20K live), silently dropping ~5K of
		// context on restore. Fall back to 15K only if the limit is unset (0).
		saveBudget := state.tokenLimit
		if saveBudget <= 0 {
			saveBudget = 15000
		}
		conversationHistory, err := SerializeConversationContextWithOptimization(
			state.conversationCtx,
			saveBudget,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to serialize conversation for %s: %v\n", name, err)
			conversationHistory = []SerializableMessage{}
		}

		// Serialize metrics
		serializedMetrics := &SerializedMetrics{
			TotalTokensUsed:            state.metrics.TotalTokensUsed,
			TotalInputTokens:           state.metrics.TotalInputTokens,
			TotalOutputTokens:          state.metrics.TotalOutputTokens,
			TotalResponseTimeMS:        state.metrics.TotalResponseTime.Milliseconds(),
			AverageTokensPerCall:       state.metrics.AverageTokensPerCall,
			AverageResponseTimeMS:      state.metrics.AverageResponseTime.Milliseconds(),
			ModelUsed:                  state.metrics.ModelUsed,
			LastCallTokens:             state.metrics.LastCallTokens,
			LastCallDurationMS:         state.metrics.LastCallDuration.Milliseconds(),
			AdvisorCalls:               state.metrics.AdvisorCalls,
			AdvisorInputTokens:         state.metrics.AdvisorInputTokens,
			AdvisorOutputTokens:        state.metrics.AdvisorOutputTokens,
			AdvisorCacheCreationTokens: state.metrics.AdvisorCacheCreationTokens,
			AdvisorCacheReadTokens:     state.metrics.AdvisorCacheReadTokens,
			AdvisorModelUsed:           state.metrics.AdvisorModelUsed,
		}

		serialized := &SerializedAgent{
			InvocationCount:     state.invocationCount,
			LastInvoked:         state.lastInvoked.Format(time.RFC3339),
			ConversationHistory: conversationHistory,
			TokenEstimate:       state.conversationCtx.tokenEstimate,
			MaxTokens:           state.tokenLimit,
			Metrics:             serializedMetrics,
		}
		agents[name] = serialized
	}

	// Get current session ID from adventure context
	sessionID := 0
	if am.adventureCtx != nil && am.adventureCtx.Adventure != nil {
		sessionsData, err := am.adventureCtx.Adventure.LoadSessions()
		if err == nil && len(sessionsData.Sessions) > 0 {
			sessionID = sessionsData.Sessions[len(sessionsData.Sessions)-1].ID
		}
	}

	// Create file structure
	stateFile := AgentStatesFile{
		SessionID:   sessionID,
		LastUpdated: time.Now().Format(time.RFC3339),
		Agents:      agents,
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(stateFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent states: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Backup existing file if it exists
	if _, err := os.Stat(filePath); err == nil {
		backupPath := filePath + ".backup"
		if err := os.Rename(filePath, backupPath); err != nil {
			// Log warning but continue
			fmt.Printf("Warning: failed to backup agent states: %v\n", err)
		}
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write agent states: %w", err)
	}

	return nil
}

// LoadAgentStates loads nested agent states from a JSON file.
func (am *AgentManager) LoadAgentStates(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist - this is not an error, just means no saved states
		return nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read agent states file: %w", err)
	}

	// Unmarshal JSON
	var stateFile AgentStatesFile
	if err := json.Unmarshal(data, &stateFile); err != nil {
		// Corrupted file - log error and continue with empty state
		fmt.Printf("Warning: agent states file corrupted, starting fresh: %v\n", err)
		return nil
	}

	// Restore agent states
	for name, serialized := range stateFile.Agents {
		// Get or create agent (this loads the persona)
		agent, err := am.getOrCreateNestedAgent(name)
		if err != nil {
			fmt.Printf("Warning: failed to restore agent %s: %v\n", name, err)
			continue
		}

		// Restore conversation history
		restoredCtx, err := DeserializeConversationContextFromMessages(
			serialized.ConversationHistory,
			serialized.MaxTokens,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to deserialize conversation for %s: %v\n", name, err)
			restoredCtx = NewConversationContextWithLimit(agent.tokenLimit)
		}
		agent.conversationCtx = restoredCtx

		// Restore metadata
		agent.invocationCount = serialized.InvocationCount
		lastInvoked, err := time.Parse(time.RFC3339, serialized.LastInvoked)
		if err == nil {
			agent.lastInvoked = lastInvoked
		}

		// Restore metrics
		if serialized.Metrics != nil {
			agent.metrics = &AgentMetrics{
				TotalTokensUsed:            serialized.Metrics.TotalTokensUsed,
				TotalInputTokens:           serialized.Metrics.TotalInputTokens,
				TotalOutputTokens:          serialized.Metrics.TotalOutputTokens,
				TotalResponseTime:          time.Duration(serialized.Metrics.TotalResponseTimeMS) * time.Millisecond,
				AverageTokensPerCall:       serialized.Metrics.AverageTokensPerCall,
				AverageResponseTime:        time.Duration(serialized.Metrics.AverageResponseTimeMS) * time.Millisecond,
				ModelUsed:                  serialized.Metrics.ModelUsed,
				LastCallTokens:             serialized.Metrics.LastCallTokens,
				LastCallDuration:           time.Duration(serialized.Metrics.LastCallDurationMS) * time.Millisecond,
				AdvisorCalls:               serialized.Metrics.AdvisorCalls,
				AdvisorInputTokens:         serialized.Metrics.AdvisorInputTokens,
				AdvisorOutputTokens:        serialized.Metrics.AdvisorOutputTokens,
				AdvisorCacheCreationTokens: serialized.Metrics.AdvisorCacheCreationTokens,
				AdvisorCacheReadTokens:     serialized.Metrics.AdvisorCacheReadTokens,
				AdvisorModelUsed:           serialized.Metrics.AdvisorModelUsed,
			}
		} else {
			// Initialize empty metrics if not present (backward compatibility)
			agent.metrics = &AgentMetrics{
				ModelUsed: "claude-haiku-4-5",
			}
		}
	}

	return nil
}
