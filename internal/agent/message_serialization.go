package agent

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// SerializableMessage represents a message that can be serialized to JSON.
type SerializableMessage struct {
	Role         string                   `json:"role"`
	TextContent  string                   `json:"text_content,omitempty"`
	ToolUses     []SerializableToolUse    `json:"tool_uses,omitempty"`
	ToolResults  []SerializableToolResult `json:"tool_results,omitempty"`
	TokenEstimate int                     `json:"token_estimate"`
}

// SerializableToolUse represents a tool use that can be serialized.
type SerializableToolUse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// SerializableToolResult represents a tool result that can be serialized.
type SerializableToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// SerializeMessage converts an anthropic.MessageParam to a serializable format.
func SerializeMessage(msg anthropic.MessageParam) (*SerializableMessage, error) {
	serialized := &SerializableMessage{
		Role:          string(msg.Role),
		ToolUses:      []SerializableToolUse{},
		ToolResults:   []SerializableToolResult{},
		TokenEstimate: 0,
	}

	// Extract content blocks from the message
	// msg.Content is already a slice of ContentBlockParamUnion
	for _, block := range msg.Content {
		if err := extractContentBlock(block, serialized); err != nil {
			return nil, err
		}
	}

	// Estimate tokens
	serialized.TokenEstimate = len(serialized.TextContent)/4 +
		len(serialized.ToolUses)*100 +
		len(serialized.ToolResults)*50

	return serialized, nil
}

// extractContentBlock extracts content from a content block union.
// Since AsAny() is private, we use JSON marshaling to extract the data.
func extractContentBlock(block anthropic.ContentBlockParamUnion, msg *SerializableMessage) error {
	// Marshal to JSON to inspect the block type
	blockJSON, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// Try to unmarshal as different block types
	var blockData map[string]interface{}
	if err := json.Unmarshal(blockJSON, &blockData); err != nil {
		return fmt.Errorf("failed to unmarshal block data: %w", err)
	}

	// Determine block type by checking for type field
	blockType, _ := blockData["type"].(string)

	switch blockType {
	case "text":
		// Text block
		if text, ok := blockData["text"].(string); ok {
			if msg.TextContent != "" {
				msg.TextContent += "\n"
			}
			msg.TextContent += text
		}

	case "tool_use":
		// Tool use block
		toolUse := SerializableToolUse{}
		if id, ok := blockData["id"].(string); ok {
			toolUse.ID = id
		}
		if name, ok := blockData["name"].(string); ok {
			toolUse.Name = name
		}
		if input, ok := blockData["input"].(map[string]interface{}); ok {
			toolUse.Input = input
		}
		msg.ToolUses = append(msg.ToolUses, toolUse)

	case "tool_result":
		// Tool result block
		toolResult := SerializableToolResult{}
		if toolUseID, ok := blockData["tool_use_id"].(string); ok {
			toolResult.ToolUseID = toolUseID
		}
		if isError, ok := blockData["is_error"].(bool); ok {
			toolResult.IsError = isError
		}

		// Extract content (can be string or array of blocks)
		if content, ok := blockData["content"].(string); ok {
			toolResult.Content = content
		} else if contentBlocks, ok := blockData["content"].([]interface{}); ok {
			// Extract text from content blocks
			for _, cb := range contentBlocks {
				if cbMap, ok := cb.(map[string]interface{}); ok {
					if cbType, _ := cbMap["type"].(string); cbType == "text" {
						if text, ok := cbMap["text"].(string); ok {
							toolResult.Content += text
						}
					}
				}
			}
		}

		msg.ToolResults = append(msg.ToolResults, toolResult)

	default:
		// Unknown or unsupported block type - log but don't fail
		fmt.Printf("Warning: Unknown content block type: %s\n", blockType)
	}

	return nil
}

// DeserializeMessage converts a serializable message back to anthropic.MessageParam.
func DeserializeMessage(msg *SerializableMessage) (anthropic.MessageParam, error) {
	role := anthropic.MessageParamRole(msg.Role)

	// Build content blocks
	var contentBlocks []anthropic.ContentBlockParamUnion

	// Add text content if present
	if msg.TextContent != "" {
		contentBlocks = append(contentBlocks, anthropic.NewTextBlock(msg.TextContent))
	}

	// Add tool uses if present
	for _, toolUse := range msg.ToolUses {
		contentBlocks = append(contentBlocks, anthropic.NewToolUseBlock(
			toolUse.ID,
			toolUse.Input,
			toolUse.Name,
		))
	}

	// Add tool results if present
	for _, toolResult := range msg.ToolResults {
		contentBlocks = append(contentBlocks, anthropic.NewToolResultBlock(
			toolResult.ToolUseID,
			toolResult.Content,
			toolResult.IsError,
		))
	}

	// Create message param based on role
	switch role {
	case anthropic.MessageParamRoleUser:
		return anthropic.NewUserMessage(contentBlocks...), nil
	case anthropic.MessageParamRoleAssistant:
		return anthropic.NewAssistantMessage(contentBlocks...), nil
	default:
		return anthropic.MessageParam{}, fmt.Errorf("unknown role: %s", role)
	}
}

// SerializeConversationContextWithOptimization serializes conversation with token optimization.
func SerializeConversationContextWithOptimization(ctx *ConversationContext, maxTokens int) ([]SerializableMessage, error) {
	messages := ctx.GetMessages()
	serialized := make([]SerializableMessage, 0, len(messages))
	totalTokens := 0

	// Serialize messages in reverse order (newest first)
	for i := len(messages) - 1; i >= 0; i-- {
		msg, err := SerializeMessage(messages[i])
		if err != nil {
			// Skip messages that can't be serialized
			fmt.Printf("Warning: Failed to serialize message %d: %v\n", i, err)
			continue
		}

		// Check if we've exceeded token limit
		if maxTokens > 0 && totalTokens+msg.TokenEstimate > maxTokens {
			// Stop adding older messages
			fmt.Printf("Token limit reached: %d tokens (limit: %d). Truncating %d older messages.\n",
				totalTokens, maxTokens, i+1)
			break
		}

		serialized = append(serialized, *msg)
		totalTokens += msg.TokenEstimate
	}

	// Reverse back to original order (oldest first)
	for i, j := 0, len(serialized)-1; i < j; i, j = i+1, j-1 {
		serialized[i], serialized[j] = serialized[j], serialized[i]
	}

	return serialized, nil
}

// DeserializeConversationContextFromMessages creates a conversation context from serialized messages.
func DeserializeConversationContextFromMessages(messages []SerializableMessage, tokenLimit int) (*ConversationContext, error) {
	ctx := NewConversationContextWithLimit(tokenLimit)

	for i, msg := range messages {
		anthropicMsg, err := DeserializeMessage(&msg)
		if err != nil {
			fmt.Printf("Warning: Failed to deserialize message %d: %v\n", i, err)
			continue
		}

		// Add message directly to context
		ctx.messages = append(ctx.messages, anthropicMsg)
		ctx.tokenEstimate += msg.TokenEstimate
	}

	return ctx, nil
}
