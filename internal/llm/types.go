// Package llm provides a provider-neutral boundary for LLM calls used by the
// SkillsWeaver game engine. Provider SDKs (currently the official OpenRouter Go
// SDK) may only be imported by this package; agents, tools, web handlers, and
// utility packages consume the neutral types defined here.
package llm

import (
	"encoding/json"
)

// Role identifies the sender of a conversation message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ImagePart is a base64-encoded image attached to a user message.
type ImagePart struct {
	MediaType string // e.g. "image/png"
	Base64    string
	// ResourceRef identifies a re-loadable local resource, not serialized image bytes.
	ResourceRef string
}

// DataURL renders the image as a data URL accepted by Chat image parts.
func (p ImagePart) DataURL() string {
	return "data:" + p.MediaType + ";base64," + p.Base64
}

// ToolCall is a tool invocation requested by the model. Arguments stays a raw
// JSON payload until the tool execution boundary.
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// ToolResult is the outcome of executing a tool call.
type ToolResult struct {
	ToolCallID string
	Content    string
	IsError    bool
}

// Message is a provider-neutral conversation message.
type Message struct {
	Role        Role
	Text        string
	Images      []ImagePart
	ToolCalls   []ToolCall
	ToolResults []ToolResult
	// CacheBreakpoint marks the message as a prompt-cache boundary: the
	// prefix ending at this message becomes cacheable (Anthropic-style
	// cache_control; OpenRouter converts for other providers). Meaningful
	// for user messages in multi-turn conversations where later turns
	// replay the same prefix.
	CacheBreakpoint bool
}

// UserMessage creates a user text message.
func UserMessage(text string) Message {
	return Message{Role: RoleUser, Text: text}
}

// UserMessageWithImage creates a user message with text and attached images.
func UserMessageWithImage(text string, images ...ImagePart) Message {
	return Message{Role: RoleUser, Text: text, Images: images}
}

// CachedUserMessage creates a user message marked as a prompt-cache
// breakpoint. In multi-turn conversations, replaying this message in later
// turns lets the provider serve the shared prefix from cache.
func CachedUserMessage(text string, images ...ImagePart) Message {
	m := UserMessageWithImage(text, images...)
	m.CacheBreakpoint = true
	return m
}

// AssistantMessage creates an assistant text message.
func AssistantMessage(text string) Message {
	return Message{Role: RoleAssistant, Text: text}
}

// AssistantToolCallMessage creates an assistant message requesting tool calls.
// The text may be empty (Chat sends content-less assistant messages with tools).
func AssistantToolCallMessage(text string, calls []ToolCall) Message {
	return Message{Role: RoleAssistant, Text: text, ToolCalls: calls}
}

// ToolResultMessage creates a tool-role message for a single tool result.
func ToolResultMessage(result ToolResult) Message {
	return Message{Role: RoleTool, ToolResults: []ToolResult{result}}
}

// ToolDefinition describes a client-executable function for the model.
// Parameters is the complete JSON Schema object for the tool input.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// JSONResponseFormat requests a machine-readable JSON response instead of
// free-form text. When Schema is set the provider enforces that JSON Schema
// (structured output); otherwise the provider only guarantees a bare JSON
// object. Requests carrying it must set RequireParameters so endpoints
// without response-format support fail routing loudly instead of silently
// returning prose.
type JSONResponseFormat struct {
	// Name labels the schema for providers that require it (a-z, 0-9, dashes).
	Name string
	// Schema is the complete JSON Schema object; nil selects json_object mode.
	Schema map[string]interface{}
	// Strict asks the provider to reject outputs violating the schema.
	Strict bool
}

// Request describes a single LLM turn.
type Request struct {
	Model               string
	System              string
	Messages            []Message
	Tools               []ToolDefinition
	MaxCompletionTokens int64
	Temperature         *float64
	SessionID           string
	RequireParameters   bool
	// JSONResponse optionally switches the response to structured JSON output.
	JSONResponse *JSONResponseFormat
	// ReasoningEffort optionally constrains the thinking budget of reasoning
	// models (OpenRouter reasoning.effort). Empty leaves the provider
	// default. Accepted values: none|minimal|low|medium|high|xhigh|max.
	ReasoningEffort string
	// CacheLastMessage marks the final conversation message as an
	// Anthropic-style prompt-cache breakpoint (conversational caching):
	// the whole prefix — system, tools, prior history — becomes cacheable
	// and the breakpoint advances as the conversation grows, so the next
	// call in the loop reads the replayed prefix at the cache-read price
	// instead of full input price. Enable only on call sites that replay a
	// growing conversation (agent tool loops, persisted nested-agent chats);
	// a one-shot call would pay the cache-write surcharge for nothing.
	CacheLastMessage bool
}

// FinishReason classifies why a generation ended.
type FinishReason string

const (
	FinishNone          FinishReason = ""
	FinishStop          FinishReason = "stop"
	FinishToolCalls     FinishReason = "tool_calls"
	FinishLength        FinishReason = "length"
	FinishContentFilter FinishReason = "content_filter"
	FinishError         FinishReason = "error"
)

// Usage reports provider accounting for one request (aggregate; provider
// sub-calls such as server tools are not separable).
type Usage struct {
	PromptTokens             int64
	CompletionTokens         int64
	TotalTokens              int64
	CachedTokens             int64
	CacheWriteTokens         int64
	ReasoningTokens          int64
	Cost                     float64
	ServerToolCallsRequested int64
	ServerToolCallsExecuted  int64
	// Present reports whether the provider returned usage information.
	Present bool
}

// Response is the outcome of one model call.
type Response struct {
	ID           string // provider generation/completion ID
	Model        string // actually routed model (may differ from requested)
	FinishReason FinishReason
	Message      Message // assistant message with text and/or tool calls
	Usage        Usage
}
