package agent

import (
	"fmt"

	"dungeons/internal/adventure"
	"dungeons/internal/character"

	"github.com/anthropics/anthropic-sdk-go"
)

// ConversationContext manages the conversation history with Claude.
type ConversationContext struct {
	messages      []anthropic.MessageParam
	tokenEstimate int
	maxTokens     int
}

// NewConversationContext creates a new conversation context.
func NewConversationContext() *ConversationContext {
	return &ConversationContext{
		messages:      []anthropic.MessageParam{},
		tokenEstimate: 0,
		maxTokens:     50000, // Keep history under 50K tokens
	}
}

// AddUserMessage adds a user message to the conversation.
func (ctx *ConversationContext) AddUserMessage(content string) {
	ctx.messages = append(ctx.messages, anthropic.NewUserMessage(
		anthropic.NewTextBlock(content),
	))
	ctx.tokenEstimate += len(content) / 4 // Rough estimation
	ctx.TruncateIfNeeded()
}

// AddAssistantMessage adds an assistant message to the conversation.
func (ctx *ConversationContext) AddAssistantMessage(content string) {
	// Don't add message if content is empty - API rejects empty text blocks
	if len(content) == 0 {
		return
	}

	ctx.messages = append(ctx.messages, anthropic.NewAssistantMessage(
		anthropic.NewTextBlock(content),
	))
	ctx.tokenEstimate += len(content) / 4
	ctx.TruncateIfNeeded()
}

// AddAssistantMessageWithTools adds an assistant message with tool uses.
func (ctx *ConversationContext) AddAssistantMessageWithTools(content string, toolUses []ToolUse) {
	contentBlocks := []anthropic.ContentBlockParamUnion{}

	// Only add text block if content is not empty
	// API rejects empty text blocks
	if len(content) > 0 {
		contentBlocks = append(contentBlocks, anthropic.NewTextBlock(content))
	}

	// Add tool use blocks
	for _, use := range toolUses {
		contentBlocks = append(contentBlocks, anthropic.NewToolUseBlock(
			use.ID,
			use.Input,
			use.Name,
		))
	}

	ctx.messages = append(ctx.messages, anthropic.NewAssistantMessage(contentBlocks...))
	ctx.tokenEstimate += len(content)/4 + len(toolUses)*100 // Rough estimation
	ctx.TruncateIfNeeded()
}

// AddAssistantMessageWithToolUses is an alias for AddAssistantMessageWithTools.
func (ctx *ConversationContext) AddAssistantMessageWithToolUses(content string, toolUses []ToolUse) {
	ctx.AddAssistantMessageWithTools(content, toolUses)
}

// AddToolResultMessage adds a tool result message.
func (ctx *ConversationContext) AddToolResultMessage(result ToolResultMessage) {
	ctx.messages = append(ctx.messages, anthropic.NewUserMessage(
		anthropic.NewToolResultBlock(result.ToolUseID, result.Content, result.IsError),
	))
	ctx.tokenEstimate += len(result.Content) / 4
	ctx.TruncateIfNeeded()
}

// AddToolResults adds multiple tool result messages.
func (ctx *ConversationContext) AddToolResults(results []ToolResultMessage) {
	for _, result := range results {
		ctx.AddToolResultMessage(result)
	}
}

// GetMessages returns all messages in the conversation.
func (ctx *ConversationContext) GetMessages() []anthropic.MessageParam {
	return ctx.messages
}

// TruncateIfNeeded truncates old messages if token limit is exceeded.
func (ctx *ConversationContext) TruncateIfNeeded() {
	if ctx.tokenEstimate > ctx.maxTokens && len(ctx.messages) > 20 {
		// Keep last 20 messages (10 exchanges)
		ctx.messages = ctx.messages[len(ctx.messages)-20:]
		// Recalculate token estimate
		ctx.tokenEstimate = 0
		for range ctx.messages {
			// Rough estimation based on message type
			ctx.tokenEstimate += 500 // Assume 500 tokens per message on average
		}
	}
}

// AdventureContext holds the current adventure state.
type AdventureContext struct {
	basePath      string
	Adventure     *adventure.Adventure
	Party         *adventure.Party
	Characters    []*character.Character
	Inventory     *adventure.SharedInventory
	CurrentSession *adventure.Session
	RecentJournal []adventure.JournalEntry
	State         *adventure.GameState
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
		// Get last 10 entries
		start := 0
		if len(journal.Entries) > 10 {
			start = len(journal.Entries) - 10
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

	// Get last 10 entries
	start := 0
	if len(journal.Entries) > 10 {
		start = len(journal.Entries) - 10
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
