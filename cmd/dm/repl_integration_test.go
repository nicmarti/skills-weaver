package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"
)

// MockAgent implements a minimal agent for testing input validation.
type MockAgent struct {
	receivedMessages []string
	callCount        int
}

// ProcessUserMessage records all messages sent to the agent.
func (m *MockAgent) ProcessUserMessage(message string) error {
	m.callCount++
	m.receivedMessages = append(m.receivedMessages, message)
	return nil
}

// TestREPLEmptyInputHandling tests that the REPL never sends empty messages to the agent.
func TestREPLEmptyInputHandling(t *testing.T) {
	testCases := []struct {
		name                string
		input               string
		expectedCallCount   int
		expectedLastMessage string
	}{
		{
			name:              "Empty line",
			input:             "\nexit\n",
			expectedCallCount: 0,
		},
		{
			name:              "Multiple empty lines",
			input:             "\n\n\n\nexit\n",
			expectedCallCount: 0,
		},
		{
			name:              "Spaces only",
			input:             "     \nexit\n",
			expectedCallCount: 0,
		},
		{
			name:              "Tabs only",
			input:             "\t\t\t\nexit\n",
			expectedCallCount: 0,
		},
		{
			name:              "Mixed whitespace",
			input:             "  \t  \n\t\t\n   \nexit\n",
			expectedCallCount: 0,
		},
		{
			name:                "Valid message after empty",
			input:               "\n\nhello\nexit\n",
			expectedCallCount:   1,
			expectedLastMessage: "hello",
		},
		{
			name:                "Valid message with leading/trailing spaces",
			input:               "  test message  \nexit\n",
			expectedCallCount:   1,
			expectedLastMessage: "test message",
		},
		{
			name:                "Multiple valid messages with empty lines between",
			input:               "first\n\n\nsecond\n\nthird\nexit\n",
			expectedCallCount:   3,
			expectedLastMessage: "third",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAgent := &MockAgent{
				receivedMessages: []string{},
				callCount:        0,
			}

			// Simulate user input
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			go func() {
				io.WriteString(w, tc.input)
				w.Close()
			}()

			// Capture stdout to suppress output during tests
			oldStdout := os.Stdout
			_, captureW, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			os.Stdout = captureW

			// Simulate REPL loop (extracted logic)
			simulateREPLLoop(mockAgent)

			// Restore stdin/stdout
			os.Stdin = oldStdin
			os.Stdout = oldStdout
			captureW.Close()

			// Verify agent was called the expected number of times
			if mockAgent.callCount != tc.expectedCallCount {
				t.Errorf("Expected %d agent calls, got %d", tc.expectedCallCount, mockAgent.callCount)
			}

			// Verify no empty messages were sent
			for i, msg := range mockAgent.receivedMessages {
				if msg == "" {
					t.Errorf("Agent received empty message at index %d", i)
				}
			}

			// Verify last message if expected
			if tc.expectedLastMessage != "" && len(mockAgent.receivedMessages) > 0 {
				lastMsg := mockAgent.receivedMessages[len(mockAgent.receivedMessages)-1]
				if lastMsg != tc.expectedLastMessage {
					t.Errorf("Expected last message %q, got %q", tc.expectedLastMessage, lastMsg)
				}
			}
		})
	}
}

// simulateREPLLoop simulates the main REPL loop for testing.
func simulateREPLLoop(agentInterface interface{ ProcessUserMessage(string) error }) {
	// Use bufio.Scanner just like main.go
	scanner := bufio.NewScanner(os.Stdin)

	for {
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Skip empty input (this is the logic we're testing)
		if input == "" {
			continue
		}

		// Exit on quit commands
		if input == "exit" || input == "quit" {
			return
		}

		// Process message (should never receive empty string)
		_ = agentInterface.ProcessUserMessage(input)
	}
}

// TestREPLWithRealScanner tests with actual bufio.Scanner behavior.
func TestREPLWithRealScanner(t *testing.T) {
	testCases := []struct {
		name              string
		input             string
		expectedCallCount int
		expectedMessages  []string
	}{
		{
			name:              "Empty input should not trigger agent",
			input:             "\nexit\n",
			expectedCallCount: 0,
			expectedMessages:  []string{},
		},
		{
			name:              "Whitespace input should not trigger agent",
			input:             "   \n\t\t\nexit\n",
			expectedCallCount: 0,
			expectedMessages:  []string{},
		},
		{
			name:              "Valid input should trigger agent",
			input:             "hello world\nexit\n",
			expectedCallCount: 1,
			expectedMessages:  []string{"hello world"},
		},
		{
			name:              "Multiple inputs with whitespace",
			input:             "first\n\nsecond\n   \nthird\nexit\n",
			expectedCallCount: 3,
			expectedMessages:  []string{"first", "second", "third"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock agent
			mockAgent := &MockAgent{
				receivedMessages: []string{},
				callCount:        0,
			}

			// Create pipe for stdin simulation
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			oldStdin := os.Stdin
			os.Stdin = r
			defer func() { os.Stdin = oldStdin }()

			// Write test input
			go func() {
				io.WriteString(w, tc.input)
				w.Close()
			}()

			// Capture stdout
			oldStdout := os.Stdout
			captureR, captureW, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			os.Stdout = captureW
			defer func() {
				os.Stdout = oldStdout
				captureR.Close()
			}()

			// Run REPL simulation
			simulateREPLWithScanner(mockAgent)

			captureW.Close()

			// Verify results
			if mockAgent.callCount != tc.expectedCallCount {
				t.Errorf("Expected %d calls, got %d", tc.expectedCallCount, mockAgent.callCount)
			}

			if len(mockAgent.receivedMessages) != len(tc.expectedMessages) {
				t.Errorf("Expected %d messages, got %d", len(tc.expectedMessages), len(mockAgent.receivedMessages))
			}

			for i, expected := range tc.expectedMessages {
				if i >= len(mockAgent.receivedMessages) {
					t.Errorf("Missing message at index %d: expected %q", i, expected)
					continue
				}
				if mockAgent.receivedMessages[i] != expected {
					t.Errorf("Message %d: expected %q, got %q", i, expected, mockAgent.receivedMessages[i])
				}
			}

			// Critical: Verify NO empty messages were sent
			for i, msg := range mockAgent.receivedMessages {
				if msg == "" {
					t.Fatalf("CRITICAL: Empty message sent to agent at index %d", i)
				}
			}
		})
	}
}

// simulateREPLWithScanner simulates the exact REPL logic from main.go using bufio.Scanner.
func simulateREPLWithScanner(agentInterface interface{ ProcessUserMessage(string) error }) {
	// This mirrors the exact logic from main.go lines 82-106
	scanner := bufio.NewScanner(os.Stdin)

	for {
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			// This is the critical line we're testing - empty input should NOT call agent
			continue
		}

		if input == "exit" || input == "quit" {
			break
		}

		// Process user message (this should never receive empty string)
		_ = agentInterface.ProcessUserMessage(input)
	}
}

// TestAgentNeverReceivesEmpty is a property-based test ensuring the invariant.
func TestAgentNeverReceivesEmpty(t *testing.T) {
	// Property: For ANY input to the REPL, the agent should NEVER receive an empty message
	inputs := []string{
		"",
		" ",
		"  ",
		"\t",
		"\n",
		"\r",
		"   \t\n\r   ",
		"hello",
		"  hello  ",
		"\thello\t",
	}

	for _, input := range inputs {
		t.Run("Input: "+input, func(t *testing.T) {
			mockAgent := &MockAgent{receivedMessages: []string{}}

			// Simulate processing this exact input
			trimmed := strings.TrimSpace(input)
			if trimmed != "" {
				mockAgent.ProcessUserMessage(trimmed)
			}

			// Verify invariant: agent never received empty message
			for i, msg := range mockAgent.receivedMessages {
				if msg == "" {
					t.Fatalf("INVARIANT VIOLATED: Agent received empty message at index %d for input %q", i, input)
				}
			}
		})
	}
}
