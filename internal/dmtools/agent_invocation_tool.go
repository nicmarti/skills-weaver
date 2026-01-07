package dmtools

import (
	"fmt"
)

// AgentManager interface defines the methods needed to invoke nested agents.
// This allows dmtools to depend on the interface rather than the concrete implementation.
type AgentManager interface {
	InvokeAgent(agentName, question, contextInfo string, depth int) (string, error)
}

// NewInvokeAgentTool creates a tool to invoke specialized nested agents.
// This tool allows the main DM agent to consult character-creator, rules-keeper, or world-keeper.
func NewInvokeAgentTool(agentManager AgentManager) *SimpleTool {
	return &SimpleTool{
		name: "invoke_agent",
		description: `Invoke a specialized nested agent to answer a question or provide guidance.
Available agents:
- character-creator: Guide character creation, explain races/classes, help with build decisions
- rules-keeper: Answer D&D 5e rules questions, resolve combat/magic/ability disputes
- world-keeper: Validate world consistency, review locations/NPCs, ensure canon coherence

Use this when you need specialized expertise beyond your immediate knowledge.
The nested agent maintains conversation history within the session.`,
		schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_name": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"character-creator", "rules-keeper", "world-keeper"},
					"description": "Name of the specialized agent to invoke",
				},
				"question": map[string]interface{}{
					"type":        "string",
					"description": "The question or task for the nested agent",
				},
				"context": map[string]interface{}{
					"type":        "string",
					"description": "Optional additional context for the agent (e.g., current situation, relevant stats)",
				},
			},
			"required": []string{"agent_name", "question"},
		},
		execute: func(params map[string]interface{}) (interface{}, error) {
			agentName := params["agent_name"].(string)
			question := params["question"].(string)

			contextInfo := ""
			if ctx, ok := params["context"].(string); ok {
				contextInfo = ctx
			}

			// Invoke the nested agent with depth=1 (nested agents cannot invoke other agents)
			response, err := agentManager.InvokeAgent(agentName, question, contextInfo, 1)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
					"display": fmt.Sprintf("Failed to consult %s: %v", agentName, err),
				}, nil
			}

			// Return the agent's response
			// The console output is handled by OutputHandler.OnAgentInvocationStart/Complete
			return map[string]interface{}{
				"success":    true,
				"agent_name": agentName,
				"response":   response,
				"display":    fmt.Sprintf("%s provided guidance", agentName),
			}, nil
		},
	}
}
