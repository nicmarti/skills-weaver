package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// MockAnthropicClient implements a mock Anthropic API client for testing.
type MockAnthropicClient struct {
	messages *MockMessagesService
}

// GetMessages returns the mock messages service.
func (m *MockAnthropicClient) GetMessages() messagesService {
	return m.messages
}

// MockMessagesService implements a mock Messages service.
type MockMessagesService struct {
	// Responses maps agent questions to predefined responses
	Responses map[string]string

	// CallCount tracks number of API calls made
	CallCount int

	// LastParams stores the last parameters passed to New()
	LastParams *anthropic.MessageNewParams

	// SimulateError when set to true, returns an error
	SimulateError bool

	// ErrorMessage is the error to return when SimulateError is true
	ErrorMessage string
}

// NewMockAnthropicClient creates a new mock client with sensible defaults.
func NewMockAnthropicClient() *MockAnthropicClient {
	return &MockAnthropicClient{
		messages: &MockMessagesService{
			Responses: make(map[string]string),
			CallCount: 0,
		},
	}
}

// GetMockMessagesService returns the underlying mock messages service for test configuration.
func (m *MockAnthropicClient) GetMockMessagesService() *MockMessagesService {
	return m.messages
}

// New implements the Messages.New method for testing.
func (m *MockMessagesService) New(ctx context.Context, params anthropic.MessageNewParams, opts ...option.RequestOption) (*anthropic.Message, error) {
	m.CallCount++
	m.LastParams = &params

	// Simulate error if configured
	if m.SimulateError {
		errMsg := m.ErrorMessage
		if errMsg == "" {
			errMsg = "mock API error"
		}
		return nil, fmt.Errorf("%s", errMsg)
	}

	// Extract user message to determine response (simplified for mock)
	userMessage := fmt.Sprintf("Request #%d", m.CallCount)

	// Get predefined response or use default
	responseText := m.getResponse(userMessage)

	// Create mock API response using JSON unmarshaling to properly construct ContentBlockUnion
	// This avoids issues with the complex union type
	messageJSON := fmt.Sprintf(`{
		"id": "mock-msg-%d",
		"type": "message",
		"role": "assistant",
		"model": "claude-haiku-4-5",
		"content": [{"type": "text", "text": %q}],
		"usage": {"input_tokens": 100, "output_tokens": 50},
		"stop_reason": "end_turn"
	}`, m.CallCount, responseText)

	var message anthropic.Message
	if err := json.Unmarshal([]byte(messageJSON), &message); err != nil {
		return nil, fmt.Errorf("failed to create mock message: %w", err)
	}

	return &message, nil
}

// getResponse returns a predefined response based on the question.
func (m *MockMessagesService) getResponse(question string) string {
	// Check for exact match
	if response, ok := m.Responses[question]; ok {
		return response
	}

	// Check for partial match (case-insensitive contains)
	questionLower := mockToLower(question)
	for key, response := range m.Responses {
		if mockContains(questionLower, mockToLower(key)) {
			return response
		}
	}

	// Default responses based on common questions
	if mockContains(questionLower, "armor class") || mockContains(questionLower, "ac") {
		return "Armor Class (AC) is calculated as 10 + Dexterity modifier + armor bonus + shield bonus."
	}

	if mockContains(questionLower, "saving throw") {
		return "Saving throws are d20 + ability modifier + proficiency bonus (if proficient)."
	}

	if mockContains(questionLower, "character") {
		return "To create a character, choose species, class, roll stats (4d6 keep highest 3), select skills, and pick equipment."
	}

	if mockContains(questionLower, "world") || mockContains(questionLower, "lore") {
		return "The world has four major factions: Valdorine (maritime), Karvath (military), Lumenciel (religious), and Astrene (scholarly)."
	}

	// Generic fallback
	return "This is a mock response from the " + extractAgentName(question) + " agent. The question was: " + question
}

// extractAgentName attempts to identify which agent should respond based on context.
func extractAgentName(question string) string {
	questionLower := mockToLower(question)

	if mockContains(questionLower, "rule") || mockContains(questionLower, "combat") || mockContains(questionLower, "spell") {
		return "rules-keeper"
	}

	if mockContains(questionLower, "character") || mockContains(questionLower, "class") || mockContains(questionLower, "species") {
		return "character-creator"
	}

	if mockContains(questionLower, "world") || mockContains(questionLower, "lore") || mockContains(questionLower, "faction") {
		return "world-keeper"
	}

	return "assistant"
}

// SetResponse sets a predefined response for a specific question.
func (m *MockMessagesService) SetResponse(question, response string) {
	m.Responses[question] = response
}

// Reset resets the mock state (call count, responses, etc.).
func (m *MockMessagesService) Reset() {
	m.CallCount = 0
	m.Responses = make(map[string]string)
	m.LastParams = nil
	m.SimulateError = false
	m.ErrorMessage = ""
}

// Helper functions for mock

func mockToLower(s string) string {
	result := ""
	for _, ch := range s {
		if ch >= 'A' && ch <= 'Z' {
			result += string(ch + 32)
		} else {
			result += string(ch)
		}
	}
	return result
}

func mockContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || mockFindSubstring(s, substr))
}

func mockFindSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
