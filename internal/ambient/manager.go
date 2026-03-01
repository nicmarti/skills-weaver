package ambient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	lyriaWSURL = "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1alpha.GenerativeService.BidiGenerateMusic"
	lyriaModel = "models/lyria-realtime-exp"

	// reconnect delay
	reconnectDelay = 5 * time.Second
)

// LyriaManager manages a persistent WebSocket connection to Lyria and broadcasts PCM audio.
type LyriaManager struct {
	apiKey      string
	mu          sync.RWMutex
	conn        *websocket.Conn
	subscribers map[string]chan []byte
	done        chan struct{}
	connected   bool
	reconnectCh chan struct{}
}

// NewLyriaManager creates a new LyriaManager.
func NewLyriaManager(apiKey string) *LyriaManager {
	return &LyriaManager{
		apiKey:      apiKey,
		subscribers: make(map[string]chan []byte),
		done:        make(chan struct{}),
		reconnectCh: make(chan struct{}, 1),
	}
}

// Connect establishes the WebSocket connection to Lyria and starts the receive loop.
func (m *LyriaManager) Connect(ctx context.Context) error {
	wsURL := fmt.Sprintf("%s?key=%s", lyriaWSURL, url.QueryEscape(m.apiKey))

	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
	}

	conn, resp, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("WebSocket dial failed (HTTP %d %s: %s): %w", resp.StatusCode, resp.Status, string(body), err)
		}
		return fmt.Errorf("WebSocket dial failed: %w", err)
	}

	// Send setup message
	setup := SetupMessage{Setup: LyriaSetup{Model: lyriaModel}}
	if err := conn.WriteJSON(setup); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send setup: %w", err)
	}

	// Wait for setupComplete
	if err := m.waitForSetupComplete(conn); err != nil {
		conn.Close()
		return fmt.Errorf("setup failed: %w", err)
	}

	// Send play command to start audio streaming
	play := PlaybackControlMessage{PlaybackControl: "PLAY"}
	if err := conn.WriteJSON(play); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send play: %w", err)
	}

	m.mu.Lock()
	m.conn = conn
	m.connected = true
	m.mu.Unlock()

	// Start receive loop in background
	go m.receiveLoop(conn)

	return nil
}

// waitForSetupComplete waits for the setupComplete message from Lyria.
func (m *LyriaManager) waitForSetupComplete(conn *websocket.Conn) error {
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read error waiting for setupComplete: %w", err)
		}

		var msg ServerMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		if msg.SetupComplete != nil {
			return nil
		}
	}
}

// SetScene sends new weighted prompts and config to Lyria to change the ambient music.
// Two separate messages are sent to match the SDK protocol:
//  1. clientContent with weightedPrompts
//  2. musicGenerationConfig with BPM and temperature
func (m *LyriaManager) SetScene(prompt string, bpm int, temp float64) error {
	m.mu.RLock()
	conn := m.conn
	connected := m.connected
	m.mu.RUnlock()

	if !connected || conn == nil {
		return fmt.Errorf("Lyria not connected")
	}

	contentMsg := ClientContentMessage{
		ClientContent: ClientContent{
			WeightedPrompts: []WeightedPrompt{{Text: prompt, Weight: 1.0}},
		},
	}
	if err := conn.WriteJSON(contentMsg); err != nil {
		return fmt.Errorf("failed to send weighted prompts: %w", err)
	}

	configMsg := MusicGenerationConfigMessage{
		MusicGenerationConfig: MusicGenerationConfig{BPM: bpm, Temperature: temp},
	}
	if err := conn.WriteJSON(configMsg); err != nil {
		return fmt.Errorf("failed to send music config: %w", err)
	}

	return nil
}

// Subscribe returns a channel for receiving PCM audio chunks and a cancel function.
func (m *LyriaManager) Subscribe() (id string, ch <-chan []byte, cancel func()) {
	id = uuid.New().String()
	audioCh := make(chan []byte, 64) // buffer to avoid dropping frames

	m.mu.Lock()
	m.subscribers[id] = audioCh
	m.mu.Unlock()

	cancel = func() {
		m.mu.Lock()
		delete(m.subscribers, id)
		m.mu.Unlock()
		// Drain channel
		for len(audioCh) > 0 {
			<-audioCh
		}
		close(audioCh)
	}

	return id, audioCh, cancel
}

// IsConnected returns whether Lyria is connected.
func (m *LyriaManager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// Stop shuts down the LyriaManager.
func (m *LyriaManager) Stop() {
	select {
	case <-m.done:
	default:
		close(m.done)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}
	m.connected = false
}

// receiveLoop reads audio chunks from Lyria and broadcasts to all subscribers.
func (m *LyriaManager) receiveLoop(conn *websocket.Conn) {
	defer func() {
		m.mu.Lock()
		if m.conn == conn {
			m.conn = nil
			m.connected = false
		}
		m.mu.Unlock()

		conn.Close()

		// Signal reconnect unless stopped
		select {
		case <-m.done:
		default:
			select {
			case m.reconnectCh <- struct{}{}:
			default:
			}
			go m.reconnectLoop()
		}
	}()

	for {
		select {
		case <-m.done:
			return
		default:
		}

		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("Lyria WebSocket read error: %v", err)
			}
			return
		}

		var msg ServerMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		if msg.ServerContent == nil {
			continue
		}

		for _, chunk := range msg.ServerContent.AudioChunks {
			if chunk.Data == "" {
				continue
			}

			pcmData, err := base64.StdEncoding.DecodeString(chunk.Data)
			if err != nil {
				log.Printf("Lyria: base64 decode error: %v", err)
				continue
			}

			m.mu.RLock()
			subCount := len(m.subscribers)
			m.mu.RUnlock()
			log.Printf("Lyria: audio chunk %d bytes → %d subscriber(s)", len(pcmData), subCount)

			m.broadcast(pcmData)
		}
	}
}

// broadcast sends PCM data to all subscribers, dropping frames if channels are full.
func (m *LyriaManager) broadcast(pcmData []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.subscribers {
		select {
		case ch <- pcmData:
		default:
			// Subscriber too slow, drop frame
		}
	}
}

// reconnectLoop attempts to reconnect to Lyria after a disconnection.
func (m *LyriaManager) reconnectLoop() {
	for {
		select {
		case <-m.done:
			return
		case <-time.After(reconnectDelay):
		}

		log.Printf("Lyria: attempting reconnect...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := m.Connect(ctx)
		cancel()

		if err != nil {
			log.Printf("Lyria reconnect failed: %v", err)
			continue
		}

		log.Printf("Lyria: reconnected successfully")
		return
	}
}
