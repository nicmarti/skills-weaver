package agent

import (
	"encoding/json"
	"fmt"

	"dungeons/internal/llm"
)

// SerializableMessage retains the pre-migration JSON field names. Image bytes
// are never written to agent-states.json; resource references are lightweight.
type SerializableMessage struct {
	Role          string                   `json:"role"`
	TextContent   string                   `json:"text_content,omitempty"`
	ToolUses      []SerializableToolUse    `json:"tool_uses,omitempty"`
	ToolResults   []SerializableToolResult `json:"tool_results,omitempty"`
	ImageResource string                   `json:"image_resource,omitempty"`
	TokenEstimate int                      `json:"token_estimate"`
}

type SerializableToolUse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type SerializableToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// SerializeMessage saves a neutral message without storing image payloads.
func SerializeMessage(msg llm.Message) (*SerializableMessage, error) {
	serialized := &SerializableMessage{Role: string(msg.Role), TextContent: msg.Text}
	for _, call := range msg.ToolCalls {
		if !json.Valid(call.Arguments) {
			return nil, fmt.Errorf("invalid tool arguments for %q", call.Name)
		}
		serialized.ToolUses = append(serialized.ToolUses, SerializableToolUse{
			ID: call.ID, Name: call.Name, Input: append(json.RawMessage(nil), call.Arguments...),
		})
	}
	for _, result := range msg.ToolResults {
		serialized.ToolResults = append(serialized.ToolResults, SerializableToolResult{
			ToolUseID: result.ToolCallID, Content: result.Content, IsError: result.IsError,
		})
	}
	if len(msg.Images) > 0 {
		serialized.ImageResource = msg.Images[0].ResourceRef
		if serialized.TextContent == "" {
			serialized.TextContent = "[image non persistée]"
		}
	}
	serialized.TokenEstimate = estimateMessage(msg)
	return serialized, nil
}

// DeserializeMessage accepts old user-role tool results as well as new tool-role
// messages. The conversation loader validates whole exchanges before calling it.
func DeserializeMessage(msg *SerializableMessage) (llm.Message, error) {
	if msg == nil {
		return llm.Message{}, fmt.Errorf("nil serialized message")
	}
	role := llm.Role(msg.Role)
	if role != llm.RoleUser && role != llm.RoleAssistant && role != llm.RoleTool {
		return llm.Message{}, fmt.Errorf("unknown role: %s", msg.Role)
	}
	if role == llm.RoleUser && len(msg.ToolResults) > 0 && msg.TextContent == "" {
		role = llm.RoleTool
	}
	message := llm.Message{Role: role, Text: msg.TextContent}
	for _, use := range msg.ToolUses {
		if !json.Valid(use.Input) {
			return llm.Message{}, fmt.Errorf("invalid tool arguments for %q", use.Name)
		}
		message.ToolCalls = append(message.ToolCalls, llm.ToolCall{
			ID: use.ID, Name: use.Name, Arguments: append(json.RawMessage(nil), use.Input...),
		})
	}
	for _, result := range msg.ToolResults {
		message.ToolResults = append(message.ToolResults, llm.ToolResult{
			ToolCallID: result.ToolUseID, Content: result.Content, IsError: result.IsError,
		})
	}
	// A persisted resource needs to be restored by its owner (World Keeper).
	if msg.ImageResource != "" {
		message.Images = []llm.ImagePart{{ResourceRef: msg.ImageResource}}
	}
	if message.Text == "" && len(message.ToolCalls) == 0 && len(message.ToolResults) == 0 && len(message.Images) == 0 {
		return llm.Message{}, fmt.Errorf("empty content for role %s", role)
	}
	return message, nil
}

// SerializeConversationContextWithOptimization retains newest complete tool
// exchanges when the disk-only budget is exceeded, never individual results.
func SerializeConversationContextWithOptimization(ctx *ConversationContext, maxTokens int) ([]SerializableMessage, error) {
	messages := ctx.NeutralMessages()
	serialized := make([]SerializableMessage, 0, len(messages))
	for _, msg := range messages {
		entry, err := SerializeMessage(msg)
		if err != nil {
			return nil, err
		}
		serialized = append(serialized, *entry)
	}
	serialized = cleanOrphanedToolResults(serialized)
	if maxTokens <= 0 {
		return serialized, nil
	}
	// Build atomic ranges over the cleaned messages. Newest unit is always
	// retained even when it alone exceeds the budget.
	neutral := make([]llm.Message, len(serialized))
	for i := range serialized {
		msg, err := DeserializeMessage(&serialized[i])
		if err != nil {
			return nil, err
		}
		neutral[i] = msg
	}
	starts := []int{0}
	for i := 0; i < len(neutral); {
		i = exchangeEnd(neutral, i)
		starts = append(starts, i)
	}
	start, tokens := len(serialized), 0
	for i := len(starts) - 2; i >= 0; i-- {
		unitTokens := 0
		for _, msg := range serialized[starts[i]:starts[i+1]] {
			unitTokens += msg.TokenEstimate
		}
		if start < len(serialized) && tokens+unitTokens > maxTokens {
			break
		}
		tokens += unitTokens
		start = starts[i]
	}
	if start > 0 {
		fmt.Printf("[agent-state] saved history trimmed to ~%d/%d-token budget (older exchanges dropped from disk only, live context unaffected)\n", tokens, maxTokens)
	}
	return serialized[start:], nil
}

// cleanOrphanedToolResults removes malformed or incomplete tool exchanges as
// a unit; it also converts legacy user-role results into tool-role messages.
func cleanOrphanedToolResults(messages []SerializableMessage) []SerializableMessage {
	cleaned := make([]SerializableMessage, 0, len(messages))
	for i := 0; i < len(messages); {
		msg := messages[i]
		if msg.Role == "assistant" && len(msg.ToolUses) > 0 {
			want := make(map[string]bool, len(msg.ToolUses))
			for _, use := range msg.ToolUses {
				want[use.ID] = true
			}
			seen := make(map[string]bool, len(want))
			j := i + 1
			var results []SerializableMessage
			var userTexts []SerializableMessage
			valid := len(want) == len(msg.ToolUses)
			for _, use := range msg.ToolUses {
				if use.ID == "" || use.Name == "" {
					valid = false
				}
			}
			for j < len(messages) && (messages[j].Role == "tool" || len(messages[j].ToolResults) > 0) {
				row := messages[j]
				for _, result := range row.ToolResults {
					if !want[result.ToolUseID] || seen[result.ToolUseID] {
						valid = false
					}
					seen[result.ToolUseID] = true
					results = append(results, SerializableMessage{
						Role: "tool", ToolResults: []SerializableToolResult{result}, TokenEstimate: len(result.Content)/4 + 50,
					})
				}
				if row.Role == "user" && row.TextContent != "" {
					userTexts = append(userTexts, SerializableMessage{Role: "user", TextContent: row.TextContent, TokenEstimate: len(row.TextContent) / 4})
				}
				j++
			}
			if valid && len(seen) == len(want) {
				cleaned = append(cleaned, msg)
				cleaned = append(cleaned, results...)
				cleaned = append(cleaned, userTexts...)
			} else {
				// Preserve standalone user text even if its attached results were invalid.
				cleaned = append(cleaned, userTexts...)
			}
			i = j
			continue
		}
		if msg.Role == "tool" || len(msg.ToolResults) > 0 {
			if msg.Role == "user" && msg.TextContent != "" {
				msg.ToolResults = nil
				cleaned = append(cleaned, msg)
			}
			i++
			continue
		}
		if msg.TextContent != "" || msg.ImageResource != "" || len(msg.ToolUses) > 0 {
			cleaned = append(cleaned, msg)
		}
		i++
	}
	return cleaned
}

// DeserializeConversationContextFromMessages restores a validated neutral
// conversation, treating missing schema_version as legacy format.
func DeserializeConversationContextFromMessages(messages []SerializableMessage, tokenLimit int) (*ConversationContext, error) {
	ctx := NewConversationContextWithLimit(tokenLimit)
	for _, row := range cleanOrphanedToolResults(messages) {
		msg, err := DeserializeMessage(&row)
		if err != nil {
			return nil, err
		}
		ctx.messages = append(ctx.messages, msg)
		ctx.tokenEstimate += estimateMessage(msg)
	}
	return ctx, nil
}
