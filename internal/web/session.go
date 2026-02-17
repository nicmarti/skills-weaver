package web

import (
	"fmt"
	"sync"
	"time"

	"dungeons/internal/agent"
)

const (
	// SessionTTL is how long a session remains active without activity.
	// Set to 2 hours to allow for breaks during gameplay.
	SessionTTL = 2 * time.Hour
	// CleanupInterval is how often to check for expired sessions.
	CleanupInterval = 10 * time.Minute
)

// OutputRedirector is an OutputHandler that redirects to another OutputHandler.
// This allows us to change where output goes without recreating the agent.
type OutputRedirector struct {
	target *WebOutput
	mu     sync.RWMutex
}

// NewOutputRedirector creates a new output redirector.
func NewOutputRedirector() *OutputRedirector {
	return &OutputRedirector{
		target: NewWebOutput(),
	}
}

// SetTarget sets the target WebOutput for redirection.
func (r *OutputRedirector) SetTarget(target *WebOutput) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.target = target
}

// GetTarget returns the current target WebOutput.
func (r *OutputRedirector) GetTarget() *WebOutput {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.target
}

// OnTextChunk implements OutputHandler.
func (r *OutputRedirector) OnTextChunk(text string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnTextChunk(text)
	}
}

// OnToolStart implements OutputHandler.
func (r *OutputRedirector) OnToolStart(toolName, toolID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnToolStart(toolName, toolID)
	}
}

// OnToolComplete implements OutputHandler.
func (r *OutputRedirector) OnToolComplete(toolName string, result interface{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnToolComplete(toolName, result)
	}
}

// OnAgentInvocationStart implements OutputHandler.
func (r *OutputRedirector) OnAgentInvocationStart(agentName string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnAgentInvocationStart(agentName)
	}
}

// OnAgentInvocationComplete implements OutputHandler.
func (r *OutputRedirector) OnAgentInvocationComplete(agentName string, duration time.Duration) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnAgentInvocationComplete(agentName, duration)
	}
}

// OnError implements OutputHandler.
func (r *OutputRedirector) OnError(err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnError(err)
	}
}

// OnComplete implements OutputHandler.
func (r *OutputRedirector) OnComplete() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnComplete()
	}
}

// OnLocationUpdate implements LocationUpdateNotifier.
func (r *OutputRedirector) OnLocationUpdate(location string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnLocationUpdate(location)
	}
}

// OnMapGenerated forwards map generation events to the target.
func (r *OutputRedirector) OnMapGenerated(location, mapPath string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.target != nil {
		r.target.OnMapGenerated(location, mapPath)
	}
}

// Session represents an active game session for an adventure.
type Session struct {
	Slug           string
	Agent          *agent.Agent
	AdventureCtx   *agent.AdventureContext
	outputRedirect *OutputRedirector
	LastActivity   time.Time
	mu             sync.Mutex
	processing     bool
}

// SessionManager manages game sessions, one per adventure.
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration
	apiKey   string
	stopCh   chan struct{}
}

// NewSessionManager creates a new session manager.
func NewSessionManager(apiKey string) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		ttl:      SessionTTL,
		apiKey:   apiKey,
		stopCh:   make(chan struct{}),
	}
	go sm.cleanupLoop()
	return sm
}

// GetOrCreateSession returns an existing session or creates a new one.
func (sm *SessionManager) GetOrCreateSession(slug string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check for existing session
	if session, exists := sm.sessions[slug]; exists {
		session.LastActivity = time.Now()
		return session, nil
	}

	// Create new session
	session, err := sm.createSession(slug)
	if err != nil {
		return nil, err
	}

	sm.sessions[slug] = session
	return session, nil
}

// createSession creates a new session for an adventure.
func (sm *SessionManager) createSession(slug string) (*Session, error) {
	// Load adventure context
	adventureCtx, err := agent.LoadAdventureContext("data/adventures", slug)
	if err != nil {
		return nil, fmt.Errorf("failed to load adventure: %w", err)
	}

	// Create output redirector
	outputRedirect := NewOutputRedirector()

	// Create agent with the redirector as output handler
	dmAgent, err := agent.New(sm.apiKey, adventureCtx, outputRedirect)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &Session{
		Slug:           slug,
		Agent:          dmAgent,
		AdventureCtx:   adventureCtx,
		outputRedirect: outputRedirect,
		LastActivity:   time.Now(),
	}, nil
}

// GetSession returns an existing session if it exists.
func (sm *SessionManager) GetSession(slug string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[slug]
	if exists {
		session.LastActivity = time.Now()
	}
	return session, exists
}

// RemoveSession removes a session.
func (sm *SessionManager) RemoveSession(slug string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[slug]; exists {
		if output := session.outputRedirect.GetTarget(); output != nil {
			output.Close()
		}
		delete(sm.sessions, slug)
	}
}

// cleanupLoop periodically removes expired sessions.
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.cleanupExpired()
		case <-sm.stopCh:
			return
		}
	}
}

// cleanupExpired removes sessions that have been inactive for too long.
func (sm *SessionManager) cleanupExpired() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for slug, session := range sm.sessions {
		if now.Sub(session.LastActivity) > sm.ttl {
			if output := session.outputRedirect.GetTarget(); output != nil {
				output.Close()
			}
			delete(sm.sessions, slug)
		}
	}
}

// Stop stops the session manager cleanup loop.
func (sm *SessionManager) Stop() {
	close(sm.stopCh)
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, session := range sm.sessions {
		if output := session.outputRedirect.GetTarget(); output != nil {
			output.Close()
		}
	}
}

// IsProcessing returns whether the session is currently processing a message.
func (s *Session) IsProcessing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.processing
}

// ProcessMessage processes a user message and returns the WebOutput to read from.
// The WebOutput channel will receive events as they are generated.
func (s *Session) ProcessMessage(message string) (*WebOutput, error) {
	s.mu.Lock()
	if s.processing {
		s.mu.Unlock()
		return nil, fmt.Errorf("session is already processing a message")
	}
	s.processing = true

	// Reload adventure context to get latest state
	// This ensures the DM has up-to-date journal/inventory even if modified externally
	if err := s.AdventureCtx.Reload(); err != nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("failed to reload adventure context: %w", err)
	}

	// Create new WebOutput for this message
	output := NewWebOutput()
	s.outputRedirect.SetTarget(output)
	s.mu.Unlock()

	// Process in goroutine so caller can start reading events immediately
	go func() {
		defer func() {
			s.mu.Lock()
			s.processing = false
			s.mu.Unlock()
		}()

		if err := s.Agent.ProcessUserMessage(message); err != nil {
			output.OnError(err)
		}
		// Output.OnComplete() is called by the agent
	}()

	return output, nil
}

// GetCurrentOutput returns the current WebOutput being used.
func (s *Session) GetCurrentOutput() *WebOutput {
	return s.outputRedirect.GetTarget()
}
