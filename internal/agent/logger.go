package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Logger handles debug logging for the agent.
type Logger struct {
	file    *os.File
	enabled bool
}

// NewLogger creates a new logger that writes to the adventure directory.
func NewLogger(adventurePath string) (*Logger, error) {
	logPath := filepath.Join(adventurePath, "sw-dm.log")

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("creating log file: %w", err)
	}

	logger := &Logger{
		file:    file,
		enabled: true,
	}

	// Write session start marker
	logger.LogSeparator()
	logger.LogInfo("New session started")
	logger.LogSeparator()

	return logger, nil
}

// Close closes the log file.
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// LogSeparator writes a visual separator.
func (l *Logger) LogSeparator() {
	if !l.enabled || l.file == nil {
		return
	}
	separator := ""
	for i := 0; i < 80; i++ {
		separator += "="
	}
	l.file.WriteString(fmt.Sprintf("\n%s\n\n", separator))
}

// LogInfo logs an info message.
func (l *Logger) LogInfo(message string) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	l.file.WriteString(fmt.Sprintf("[%s] INFO: %s\n", timestamp, message))
}

// LogUserMessage logs a user message.
func (l *Logger) LogUserMessage(message string) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	l.file.WriteString(fmt.Sprintf("\n[%s] USER: %s\n\n", timestamp, message))
}

// LogToolCall logs a tool invocation with its parameters.
func (l *Logger) LogToolCall(toolName string, toolID string, params map[string]interface{}) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	l.file.WriteString(fmt.Sprintf("[%s] TOOL CALL: %s (ID: %s)\n", timestamp, toolName, toolID))

	// Format parameters as JSON for readability
	if paramsJSON, err := json.MarshalIndent(params, "  ", "  "); err == nil {
		l.file.WriteString(fmt.Sprintf("  Parameters:\n  %s\n", string(paramsJSON)))
	} else {
		l.file.WriteString(fmt.Sprintf("  Parameters: %v\n", params))
	}
}

// LogCLICommand logs the equivalent CLI command for a tool call.
func (l *Logger) LogCLICommand(command string) {
	if !l.enabled || l.file == nil || command == "" {
		return
	}
	l.file.WriteString(fmt.Sprintf("  Equivalent CLI:\n  %s\n", command))
}

// LogToolResult logs a tool execution result.
func (l *Logger) LogToolResult(toolName string, toolID string, result interface{}) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	l.file.WriteString(fmt.Sprintf("[%s] TOOL RESULT: %s (ID: %s)\n", timestamp, toolName, toolID))

	// Format result as JSON for readability
	if resultJSON, err := json.MarshalIndent(result, "  ", "  "); err == nil {
		l.file.WriteString(fmt.Sprintf("  Result:\n  %s\n\n", string(resultJSON)))
	} else {
		l.file.WriteString(fmt.Sprintf("  Result: %v\n\n", result))
	}
}

// LogError logs an error.
func (l *Logger) LogError(context string, err error) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	l.file.WriteString(fmt.Sprintf("[%s] ERROR in %s: %v\n", timestamp, context, err))
}

// LogAssistantResponse logs the assistant's text response.
func (l *Logger) LogAssistantResponse(content string) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if content != "" {
		l.file.WriteString(fmt.Sprintf("[%s] ASSISTANT: %s\n", timestamp, content))
	}
}
