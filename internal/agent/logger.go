package agent

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Logger handles debug logging for the agent with rotation and compression.
type Logger struct {
	file         *os.File
	enabled      bool
	sessionNum   int
	logPath      string
	maxSize      int64 // Max size in bytes before rotation
	maxRotations int   // Max number of rotated files to keep
	currentSize  int64 // Current file size
}

// SessionsData represents the sessions.json structure
type SessionsData struct {
	Sessions []struct {
		ID int `json:"id"`
	} `json:"sessions"`
}

// NewLogger creates a new logger that writes to the adventure directory.
// It creates a session-specific log file (sw-dm-session-N.log) to avoid huge monolithic logs.
func NewLogger(adventurePath string) (*Logger, error) {
	// Determine current session number
	sessionNum, err := getCurrentSessionNumber(adventurePath)
	if err != nil {
		// If can't determine session, use timestamp-based log
		sessionNum = 0
	}

	// Archive old monolithic sw-dm.log if it exists and is large
	oldLogPath := filepath.Join(adventurePath, "sw-dm.log")
	archiveOldLogIfNeeded(oldLogPath)

	// Create session-specific log file
	var logPath string
	if sessionNum > 0 {
		logPath = filepath.Join(adventurePath, fmt.Sprintf("sw-dm-session-%d.log", sessionNum))
	} else {
		// Fallback: use timestamp
		logPath = filepath.Join(adventurePath, fmt.Sprintf("sw-dm-%s.log", time.Now().Format("20060102-150405")))
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("creating log file: %w", err)
	}

	// Get current file size
	fileInfo, _ := file.Stat()
	currentSize := int64(0)
	if fileInfo != nil {
		currentSize = fileInfo.Size()
	}

	logger := &Logger{
		file:         file,
		enabled:      true,
		sessionNum:   sessionNum,
		logPath:      logPath,
		maxSize:      10 * 1024 * 1024, // 10MB default
		maxRotations: 5,                // Keep 5 rotated files
		currentSize:  currentSize,
	}

	// Write session start marker
	logger.LogSeparator()
	if sessionNum > 0 {
		logger.LogInfo(fmt.Sprintf("Session %d log started", sessionNum))
	} else {
		logger.LogInfo("New log started (session number unknown)")
	}
	logger.LogSeparator()

	// Clean up old rotated logs
	logger.cleanupOldLogs()

	return logger, nil
}

// getCurrentSessionNumber reads sessions.json and returns the last session ID + 1
// (representing the next/current session being played)
func getCurrentSessionNumber(adventurePath string) (int, error) {
	sessionsPath := filepath.Join(adventurePath, "sessions.json")

	// If sessions.json doesn't exist yet, this is session 1
	if _, err := os.Stat(sessionsPath); os.IsNotExist(err) {
		return 1, nil
	}

	data, err := os.ReadFile(sessionsPath)
	if err != nil {
		return 0, err
	}

	var sessionsData SessionsData
	if err := json.Unmarshal(data, &sessionsData); err != nil {
		return 0, err
	}

	// Find highest session ID
	maxID := 0
	for _, session := range sessionsData.Sessions {
		if session.ID > maxID {
			maxID = session.ID
		}
	}

	// Current session is last session ID + 1, or 1 if no sessions yet
	if maxID == 0 {
		return 1, nil
	}
	return maxID + 1, nil
}

// archiveOldLogIfNeeded archives sw-dm.log if it exists and is larger than 1MB
func archiveOldLogIfNeeded(logPath string) {
	info, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		// No old log, nothing to do
		return
	}
	if err != nil {
		// Can't stat, skip archiving
		return
	}

	// Archive if larger than 1MB (or always archive on first rotation)
	const maxSize = 1 * 1024 * 1024
	if info.Size() > maxSize {
		archivePath := fmt.Sprintf("%s.archived-%s", logPath, time.Now().Format("20060102-150405"))
		os.Rename(logPath, archivePath)
		fmt.Printf("Archived old log to: %s\n", archivePath)
	}
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
	separator := strings.Repeat("=", 80)
	l.write(fmt.Sprintf("\n%s\n\n", separator))
}

// LogInfo logs an info message.
func (l *Logger) LogInfo(message string) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	l.write(fmt.Sprintf("[%s] INFO: %s\n", timestamp, message))
}

// LogUserMessage logs a user message.
func (l *Logger) LogUserMessage(message string) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	l.write(fmt.Sprintf("\n[%s] USER: %s\n\n", timestamp, message))
}

// LogToolCall logs a tool invocation with its parameters.
func (l *Logger) LogToolCall(toolName string, toolID string, params map[string]interface{}) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var output strings.Builder
	output.WriteString(fmt.Sprintf("[%s] TOOL CALL: %s (ID: %s)\n", timestamp, toolName, toolID))

	// Format parameters as JSON for readability
	if paramsJSON, err := json.MarshalIndent(params, "  ", "  "); err == nil {
		output.WriteString(fmt.Sprintf("  Parameters:\n  %s\n", string(paramsJSON)))
	} else {
		output.WriteString(fmt.Sprintf("  Parameters: %v\n", params))
	}

	l.write(output.String())
}

// LogCLICommand logs the equivalent CLI command for a tool call.
func (l *Logger) LogCLICommand(command string) {
	if !l.enabled || l.file == nil || command == "" {
		return
	}
	l.write(fmt.Sprintf("  Equivalent CLI:\n  %s\n", command))
}

// LogToolResult logs a tool execution result.
func (l *Logger) LogToolResult(toolName string, toolID string, result interface{}) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var output strings.Builder
	output.WriteString(fmt.Sprintf("[%s] TOOL RESULT: %s (ID: %s)\n", timestamp, toolName, toolID))

	// Format result as JSON for readability
	if resultJSON, err := json.MarshalIndent(result, "  ", "  "); err == nil {
		output.WriteString(fmt.Sprintf("  Result:\n  %s\n\n", string(resultJSON)))
	} else {
		output.WriteString(fmt.Sprintf("  Result: %v\n\n", result))
	}

	l.write(output.String())
}

// LogError logs an error.
func (l *Logger) LogError(context string, err error) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	l.write(fmt.Sprintf("[%s] ERROR in %s: %v\n", timestamp, context, err))
}

// LogAssistantResponse logs the assistant's text response.
func (l *Logger) LogAssistantResponse(content string) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if content != "" {
		l.write(fmt.Sprintf("[%s] ASSISTANT: %s\n", timestamp, content))
	}
}

// LogAgentInvocation logs a nested agent invocation with its details.
func (l *Logger) LogAgentInvocation(agentName string, invocationID string, question string, contextInfo string, response string, duration time.Duration, tokensUsed int) {
	if !l.enabled || l.file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var output strings.Builder
	output.WriteString(fmt.Sprintf("\n[%s] AGENT INVOCATION: %s (ID: %s)\n", timestamp, agentName, invocationID))
	output.WriteString(fmt.Sprintf("  Question:\n  %s\n\n", indent(question, 2)))

	if contextInfo != "" {
		output.WriteString(fmt.Sprintf("  Context:\n  %s\n\n", indent(contextInfo, 2)))
	}

	output.WriteString(fmt.Sprintf("  Response:\n  %s\n\n", indent(response, 2)))
	output.WriteString(fmt.Sprintf("  Duration: %.1fs\n", duration.Seconds()))
	output.WriteString(fmt.Sprintf("  Tokens Used: ~%d\n", tokensUsed))

	l.write(output.String())
}

// indent indents each line of text by the specified number of spaces.
func indent(text string, spaces int) string {
	prefix := ""
	for i := 0; i < spaces; i++ {
		prefix += " "
	}

	lines := []string{}
	for _, line := range splitLines(text) {
		lines = append(lines, prefix+line)
	}
	return joinLines(lines)
}

// splitLines splits text into lines.
func splitLines(text string) []string {
	result := []string{}
	current := ""
	for _, ch := range text {
		if ch == '\n' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// joinLines joins lines with newlines.
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}

// write wraps file writes with rotation check and size tracking.
func (l *Logger) write(data string) error {
	if !l.enabled || l.file == nil {
		return nil
	}

	// Check if we need to rotate before writing
	dataSize := int64(len(data))
	if l.currentSize+dataSize > l.maxSize {
		if err := l.rotate(); err != nil {
			fmt.Printf("Warning: Failed to rotate log: %v\n", err)
			// Continue writing to current file
		}
	}

	// Write data
	n, err := l.file.WriteString(data)
	if err != nil {
		return err
	}

	l.currentSize += int64(n)
	return nil
}

// rotate rotates the current log file, compresses the old one, and opens a new one.
func (l *Logger) rotate() error {
	// Close current file
	if l.file != nil {
		l.file.Close()
	}

	// Rotate existing rotated files (.1 -> .2, .2 -> .3, etc.)
	for i := l.maxRotations - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d.gz", l.logPath, i)
		newPath := fmt.Sprintf("%s.%d.gz", l.logPath, i+1)

		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Compress current log to .1.gz
	rotatedPath := fmt.Sprintf("%s.1.gz", l.logPath)
	if err := l.compressFile(l.logPath, rotatedPath); err != nil {
		fmt.Printf("Warning: Failed to compress log: %v\n", err)
		// Fallback: just rename without compression
		os.Rename(l.logPath, l.logPath+".1")
	}

	// Open new log file
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	l.file = file
	l.currentSize = 0

	// Log rotation event
	l.LogInfo("Log rotated - previous log compressed")

	return nil
}

// compressFile compresses srcPath to destPath using gzip.
func (l *Logger) compressFile(srcPath, destPath string) error {
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(destFile)
	defer gzWriter.Close()

	// Copy and compress
	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	// Remove original file after successful compression
	if err := os.Remove(srcPath); err != nil {
		fmt.Printf("Warning: Failed to remove original log file: %v\n", err)
	}

	return nil
}

// cleanupOldLogs removes rotated logs older than maxRotations.
func (l *Logger) cleanupOldLogs() {
	baseName := filepath.Base(l.logPath)
	dirName := filepath.Dir(l.logPath)

	// Find all rotated log files
	pattern := fmt.Sprintf("%s.*.gz", baseName)
	matches, err := filepath.Glob(filepath.Join(dirName, pattern))
	if err != nil {
		return
	}

	// Sort by modification time (oldest first)
	sort.Slice(matches, func(i, j int) bool {
		infoI, _ := os.Stat(matches[i])
		infoJ, _ := os.Stat(matches[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Remove files beyond maxRotations
	if len(matches) > l.maxRotations {
		for i := 0; i < len(matches)-l.maxRotations; i++ {
			if err := os.Remove(matches[i]); err != nil {
				fmt.Printf("Warning: Failed to remove old log %s: %v\n", matches[i], err)
			} else {
				fmt.Printf("Removed old log: %s\n", filepath.Base(matches[i]))
			}
		}
	}

	// Also clean up old .archived-* files
	archivedPattern := fmt.Sprintf("%s.archived-*", baseName)
	archivedMatches, err := filepath.Glob(filepath.Join(dirName, archivedPattern))
	if err == nil {
		for _, archived := range archivedMatches {
			// Keep archived files for 30 days
			info, err := os.Stat(archived)
			if err == nil && time.Since(info.ModTime()) > 30*24*time.Hour {
				os.Remove(archived)
				fmt.Printf("Removed old archived log: %s\n", filepath.Base(archived))
			}
		}
	}
}

// SetMaxSize sets the maximum log file size before rotation.
func (l *Logger) SetMaxSize(sizeInMB int) {
	l.maxSize = int64(sizeInMB) * 1024 * 1024
}

// SetMaxRotations sets the maximum number of rotated log files to keep.
func (l *Logger) SetMaxRotations(count int) {
	l.maxRotations = count
}
