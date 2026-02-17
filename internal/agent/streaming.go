package agent

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

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

// StreamHandler processes streaming events from Anthropic API.
type StreamHandler struct {
	outputHandler OutputHandler
}

// NewStreamHandler creates a new stream handler.
func NewStreamHandler(outputHandler OutputHandler) *StreamHandler {
	return &StreamHandler{
		outputHandler: outputHandler,
	}
}

// ProcessStream processes streaming events and returns tool uses and accumulated text.
func (sh *StreamHandler) ProcessStream(stream *ssestream.Stream[anthropic.MessageStreamEventUnion]) ([]ToolUse, string, error) {
	defer stream.Close()

	message := anthropic.Message{}
	toolUses := []ToolUse{}

	for stream.Next() {
		event := stream.Current()

		// Accumulate the event into the message
		err := message.Accumulate(event)
		if err != nil {
			return nil, "", fmt.Errorf("failed to accumulate event: %w", err)
		}

		// Process event for display
		switch eventVariant := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			// Handle text deltas
			switch deltaVariant := eventVariant.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				// Display text immediately
				sh.outputHandler.OnTextChunk(deltaVariant.Text)
			}
		}
	}

	// Check for stream errors
	if stream.Err() != nil {
		return nil, "", fmt.Errorf("stream error: %w", stream.Err())
	}

	// Extract text content
	textContent := ""
	for _, block := range message.Content {
		switch contentBlock := block.AsAny().(type) {
		case anthropic.TextBlock:
			textContent += contentBlock.Text
		case anthropic.ToolUseBlock:
			// Extract tool use
			var input map[string]interface{}
			if err := json.Unmarshal([]byte(contentBlock.JSON.Input.Raw()), &input); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal tool input: %w", err)
			}

			toolUses = append(toolUses, ToolUse{
				ID:    contentBlock.ID,
				Name:  contentBlock.Name,
				Input: input,
			})
		}
	}

	return toolUses, textContent, nil
}
