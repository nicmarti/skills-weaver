// Package web implements a Gin-based web interface for SkillsWeaver.
package web

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SSEEvent represents a Server-Sent Event to be sent to the client.
type SSEEvent struct {
	Event string `json:"event"` // text, tool_start, tool_complete, agent_start, agent_complete, error, complete, image
	Data  string `json:"data"`  // HTML fragment or text content
}

// WebOutput implements the OutputHandler interface for SSE streaming.
// It sends events through a channel that the SSE handler reads from.
type WebOutput struct {
	eventChan chan SSEEvent
	mu        sync.Mutex
	closed    bool
}

// NewWebOutput creates a new WebOutput with a buffered event channel.
func NewWebOutput() *WebOutput {
	return &WebOutput{
		eventChan: make(chan SSEEvent, 100),
		closed:    false,
	}
}

// Events returns the event channel for reading SSE events.
func (w *WebOutput) Events() <-chan SSEEvent {
	return w.eventChan
}

// Close closes the event channel.
func (w *WebOutput) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		w.closed = true
		close(w.eventChan)
	}
}

// IsClosed returns whether the output channel is closed.
func (w *WebOutput) IsClosed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closed
}

// sendEvent safely sends an event to the channel.
func (w *WebOutput) sendEvent(event SSEEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.closed {
		select {
		case w.eventChan <- event:
		default:
			// Channel full, drop event
		}
	}
}

// OnTextChunk is called when a text chunk is received from the agent.
// This is called frequently during streaming.
// Text is JSON-encoded to preserve newlines in SSE format.
func (w *WebOutput) OnTextChunk(text string) {
	// Encode text as JSON to preserve newlines
	jsonText, _ := json.Marshal(text)
	w.sendEvent(SSEEvent{
		Event: "text",
		Data:  string(jsonText),
	})
}

// OnToolStart is called when a tool starts executing.
func (w *WebOutput) OnToolStart(toolName, toolID string) {
	data := map[string]string{
		"tool_name": toolName,
		"tool_id":   toolID,
	}
	jsonData, _ := json.Marshal(data)
	w.sendEvent(SSEEvent{
		Event: "tool_start",
		Data:  string(jsonData),
	})
}

// OnToolComplete is called when a tool finishes executing.
func (w *WebOutput) OnToolComplete(toolName string, result interface{}) {
	// Extract display message if available
	displayMsg := toolName + " complete"
	if m, ok := result.(map[string]interface{}); ok {
		if display, ok := m["display"].(string); ok {
			displayMsg = display
		}
		// Check if this is an image result
		if imagePath, ok := m["image_path"].(string); ok {
			// Send special image event
			imageData := map[string]string{
				"tool_name":  toolName,
				"image_path": imagePath,
			}
			if prompt, ok := m["prompt"].(string); ok {
				imageData["prompt"] = prompt
			}
			jsonData, _ := json.Marshal(imageData)
			w.sendEvent(SSEEvent{
				Event: "image",
				Data:  string(jsonData),
			})
		}
	}

	data := map[string]string{
		"tool_name": toolName,
		"display":   displayMsg,
	}
	jsonData, _ := json.Marshal(data)
	w.sendEvent(SSEEvent{
		Event: "tool_complete",
		Data:  string(jsonData),
	})
}

// OnAgentInvocationStart is called when invoking a nested agent.
func (w *WebOutput) OnAgentInvocationStart(agentName string) {
	data := map[string]string{
		"agent_name": agentName,
	}
	jsonData, _ := json.Marshal(data)
	w.sendEvent(SSEEvent{
		Event: "agent_start",
		Data:  string(jsonData),
	})
}

// OnAgentInvocationComplete is called when a nested agent invocation completes.
func (w *WebOutput) OnAgentInvocationComplete(agentName string, duration time.Duration) {
	data := map[string]interface{}{
		"agent_name":  agentName,
		"duration_ms": duration.Milliseconds(),
	}
	jsonData, _ := json.Marshal(data)
	w.sendEvent(SSEEvent{
		Event: "agent_complete",
		Data:  string(jsonData),
	})
}

// OnError is called when an error occurs.
func (w *WebOutput) OnError(err error) {
	data := map[string]string{
		"error": err.Error(),
	}
	jsonData, _ := json.Marshal(data)
	w.sendEvent(SSEEvent{
		Event: "error",
		Data:  string(jsonData),
	})
}

// OnComplete is called when the agent finishes processing.
func (w *WebOutput) OnComplete() {
	w.sendEvent(SSEEvent{
		Event: "complete",
		Data:  fmt.Sprintf(`{"timestamp":%d}`, time.Now().UnixMilli()),
	})
}

// OnLocationUpdate is called when the party's location changes.
// This sends a location_update SSE event to trigger mini-map refresh.
func (w *WebOutput) OnLocationUpdate(location string) {
	data := map[string]string{"location": location}
	jsonData, _ := json.Marshal(data)
	w.sendEvent(SSEEvent{
		Event: "location_update",
		Data:  string(jsonData),
	})
}

// OnMapGenerated is called when a map is generated for a location.
// This sends a map_generated SSE event to trigger mini-map refresh.
func (w *WebOutput) OnMapGenerated(location, mapPath string) {
	data := map[string]string{
		"location": location,
		"map_path": mapPath,
	}
	jsonData, _ := json.Marshal(data)
	w.sendEvent(SSEEvent{
		Event: "map_generated",
		Data:  string(jsonData),
	})
}
