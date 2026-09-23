package agent

import "time"

// OutputHandler handles output from the agent (text, tool calls, errors).
type OutputHandler interface {
	OnTextChunk(text string)
	OnToolStart(toolName, toolID string)
	OnToolComplete(toolName string, result interface{})
	OnAgentInvocationStart(agentName string)
	OnAgentInvocationComplete(agentName string, duration time.Duration)
	OnError(err error)
	OnComplete()
}
