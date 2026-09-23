package agent

import (
	"encoding/json"
	"fmt"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"dungeons/internal/llm"

	"github.com/anthropics/anthropic-sdk-go"
)

// ConversationContext stores provider-neutral conversation history.
type ConversationContext struct {
	messages      []llm.Message
	tokenEstimate int
	maxTokens     int
}

// NewConversationContext creates a new conversation context with default token limit (50K).
func NewConversationContext() *ConversationContext {
	return NewConversationContextWithLimit(50000)
}

// NewConversationContextWithLimit creates a new conversation context with a custom token limit.
// This is useful for nested agents which have lower token limits (e.g., 20K).
func NewConversationContextWithLimit(maxTokens int) *ConversationContext {
	return &ConversationContext{
		messages:      []llm.Message{},
		tokenEstimate: 0,
		maxTokens:     maxTokens,
	}
}

// AddUserMessage adds a user message to the conversation.
func (ctx *ConversationContext) AddUserMessage(content string) {
	ctx.appendMessage(llm.UserMessage(content))
}

// AddUserMessageWithImageResource includes a reference to an image reloaded on
// restore, without persisting its base64 payload.
func (ctx *ConversationContext) AddUserMessageWithImageResource(content, imageBase64, mediaType, resourceRef string) {
	ctx.appendMessage(llm.UserMessageWithImage(content, llm.ImagePart{
		MediaType: mediaType, Base64: imageBase64, ResourceRef: resourceRef,
	}))
}

// AddUserMessageWithImage adds a user message containing text and a base64-encoded image.
func (ctx *ConversationContext) AddUserMessageWithImage(content string, imageBase64 string, mediaType string) {
	ctx.AddUserMessageWithImageResource(content, imageBase64, mediaType, "")
}

func (ctx *ConversationContext) appendMessage(message llm.Message) {
	ctx.messages = append(ctx.messages, message)
	ctx.tokenEstimate += estimateMessage(message)
	ctx.TruncateIfNeeded()
}

// AddAssistantMessage adds an assistant message to the conversation.
func (ctx *ConversationContext) AddAssistantMessage(content string) {
	// Don't add message if content is empty - API rejects empty text blocks
	if len(content) == 0 {
		return
	}

	ctx.appendMessage(llm.AssistantMessage(content))
}

// AddAssistantMessageWithTools adds an assistant message with tool uses.
func (ctx *ConversationContext) AddAssistantMessageWithTools(content string, toolUses []ToolUse) {
	calls := make([]llm.ToolCall, 0, len(toolUses))
	for _, use := range toolUses {
		args, err := json.Marshal(use.Input)
		if err != nil {
			args = []byte("{}") // Legacy tool input is always a JSON-compatible map.
		}
		calls = append(calls, llm.ToolCall{ID: use.ID, Name: use.Name, Arguments: args})
	}
	ctx.appendMessage(llm.AssistantToolCallMessage(content, calls))
}

// AddAssistantMessageWithToolUses is an alias for AddAssistantMessageWithTools.
func (ctx *ConversationContext) AddAssistantMessageWithToolUses(content string, toolUses []ToolUse) {
	ctx.AddAssistantMessageWithTools(content, toolUses)
}

// AddToolResultMessage adds a tool result message.
func (ctx *ConversationContext) AddToolResultMessage(result ToolResultMessage) {
	ctx.appendMessage(llm.ToolResultMessage(llm.ToolResult{
		ToolCallID: result.ToolUseID, Content: result.Content, IsError: result.IsError,
	}))
}

// AddToolResults adds multiple tool result messages.
func (ctx *ConversationContext) AddToolResults(results []ToolResultMessage) {
	for _, result := range results {
		ctx.AddToolResultMessage(result)
	}
}

// NeutralMessages returns the provider-independent conversation.
func (ctx *ConversationContext) NeutralMessages() []llm.Message {
	return ctx.messages
}

// GetMessages is a temporary Anthropic wire adapter for the still-unmigrated
// main and nested-agent loops. Phases 4/5 replace these callers with llm.Client.
func (ctx *ConversationContext) GetMessages() []anthropic.MessageParam {
	messages := make([]anthropic.MessageParam, 0, len(ctx.messages))
	for _, msg := range ctx.messages {
		switch msg.Role {
		case llm.RoleUser:
			blocks := []anthropic.ContentBlockParamUnion{}
			if msg.Text != "" {
				blocks = append(blocks, anthropic.NewTextBlock(msg.Text))
			}
			for _, img := range msg.Images {
				if img.Base64 != "" {
					blocks = append(blocks, anthropic.NewImageBlockBase64(img.MediaType, img.Base64))
				}
			}
			messages = append(messages, anthropic.NewUserMessage(blocks...))
		case llm.RoleAssistant:
			blocks := []anthropic.ContentBlockParamUnion{}
			if msg.Text != "" {
				blocks = append(blocks, anthropic.NewTextBlock(msg.Text))
			}
			for _, call := range msg.ToolCalls {
				var args map[string]interface{}
				if err := json.Unmarshal(call.Arguments, &args); err != nil {
					args = map[string]interface{}{}
				}
				blocks = append(blocks, anthropic.NewToolUseBlock(call.ID, args, call.Name))
			}
			messages = append(messages, anthropic.NewAssistantMessage(blocks...))
		case llm.RoleTool:
			blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.ToolResults))
			for _, result := range msg.ToolResults {
				blocks = append(blocks, anthropic.NewToolResultBlock(result.ToolCallID, result.Content, result.IsError))
			}
			messages = append(messages, anthropic.NewUserMessage(blocks...))
		}
	}
	return messages
}

func estimateMessage(msg llm.Message) int {
	count := len(msg.Text)/4 + len(msg.ToolCalls)*100 + len(msg.ToolResults)*50
	for _, result := range msg.ToolResults {
		count += len(result.Content) / 4
	}
	count += len(msg.Images) * 1600
	return count
}

// exchangeEnd returns the end of the first atomic tool exchange (or one message).
func exchangeEnd(messages []llm.Message, start int) int {
	end := start + 1
	if messages[start].Role != llm.RoleAssistant || len(messages[start].ToolCalls) == 0 {
		return end
	}
	ids := make(map[string]bool, len(messages[start].ToolCalls))
	for _, call := range messages[start].ToolCalls {
		ids[call.ID] = true
	}
	for end < len(messages) && messages[end].Role == llm.RoleTool {
		valid := true
		for _, result := range messages[end].ToolResults {
			if !ids[result.ToolCallID] {
				valid = false
			}
		}
		if !valid {
			break
		}
		end++
	}
	return end
}

// TruncateIfNeeded removes oldest complete exchanges without splitting tools.
func (ctx *ConversationContext) TruncateIfNeeded() {
	for ctx.maxTokens > 0 && ctx.tokenEstimate > ctx.maxTokens && len(ctx.messages) > 1 {
		end := exchangeEnd(ctx.messages, 0)
		if end >= len(ctx.messages) {
			break // Keep an oversized latest exchange intact.
		}
		for _, msg := range ctx.messages[:end] {
			ctx.tokenEstimate -= estimateMessage(msg)
		}
		ctx.messages = ctx.messages[end:]
	}
}

// AdventureContext holds the current adventure state.
type AdventureContext struct {
	basePath       string
	Adventure      *adventure.Adventure
	Party          *adventure.Party
	Characters     []*character.Character
	Inventory      *adventure.SharedInventory
	CurrentSession *adventure.Session
	RecentJournal  []adventure.JournalEntry
	State          *adventure.GameState
}

// BasePath returns the adventure's base directory path.
func (ctx *AdventureContext) BasePath() string {
	return ctx.basePath
}

// LoadAdventureContext loads an adventure and all its associated data.
func LoadAdventureContext(baseDir, adventureName string) (*AdventureContext, error) {
	// Load adventure
	adv, err := adventure.LoadByName(baseDir, adventureName)
	if err != nil {
		return nil, fmt.Errorf("failed to load adventure: %w", err)
	}

	ctx := &AdventureContext{
		basePath:  adv.BasePath(),
		Adventure: adv,
	}

	// Load party
	party, err := adv.LoadParty()
	if err != nil {
		// If party doesn't exist, create an empty one
		party = &adventure.Party{
			Characters: []string{},
		}
	}
	ctx.Party = party

	// Load characters
	characters, err := adv.GetCharacters()
	if err != nil {
		characters = []*character.Character{}
	}
	ctx.Characters = characters

	// Load inventory
	inventory, err := adv.LoadInventory()
	if err != nil {
		// If inventory doesn't exist, create an empty one
		inventory = &adventure.SharedInventory{
			Gold:  0,
			Items: []adventure.InventoryItem{},
		}
	}
	ctx.Inventory = inventory

	// Load sessions
	sessionHistory, err := adv.LoadSessions()
	if err == nil && len(sessionHistory.Sessions) > 0 {
		// Get most recent session
		ctx.CurrentSession = &sessionHistory.Sessions[len(sessionHistory.Sessions)-1]
	}

	// Load recent journal entries
	journal, err := adv.LoadJournal()
	if err == nil {
		// Get last 30 entries (provides context from previous sessions)
		start := 0
		if len(journal.Entries) > 30 {
			start = len(journal.Entries) - 30
		}
		ctx.RecentJournal = journal.Entries[start:]
	} else {
		ctx.RecentJournal = []adventure.JournalEntry{}
	}

	// Load game state
	state, err := adv.LoadState()
	if err != nil {
		// If state doesn't exist, create a default one
		state = &adventure.GameState{
			CurrentLocation: "Point de départ",
			Quests:          []adventure.Quest{},
			Flags:           map[string]bool{},
			Variables:       map[string]string{},
		}
	}
	ctx.State = state

	return ctx, nil
}

// Reload reloads the adventure context after modifications.
func (ctx *AdventureContext) Reload() error {
	// Reload characters (to reflect level ups, HP changes, etc.)
	characters, err := ctx.Adventure.GetCharacters()
	if err != nil {
		return fmt.Errorf("failed to reload characters: %w", err)
	}
	ctx.Characters = characters

	// Reload inventory
	inventory, err := ctx.Adventure.LoadInventory()
	if err != nil {
		return fmt.Errorf("failed to reload inventory: %w", err)
	}
	ctx.Inventory = inventory

	// Reload recent journal entries
	journal, err := ctx.Adventure.LoadJournal()
	if err != nil {
		return fmt.Errorf("failed to reload journal: %w", err)
	}

	// Get last 30 entries (provides context from previous sessions)
	start := 0
	if len(journal.Entries) > 30 {
		start = len(journal.Entries) - 30
	}
	ctx.RecentJournal = journal.Entries[start:]

	// Reload game state
	state, err := ctx.Adventure.LoadState()
	if err != nil {
		return fmt.Errorf("failed to reload game state: %w", err)
	}
	ctx.State = state

	return nil
}
