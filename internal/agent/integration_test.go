package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
)

// TestIntegration_AgentInvocationFlow tests the complete agent invocation flow with mock client.
func TestIntegration_AgentInvocationFlow(t *testing.T) {
	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	adventureCtx := createTestAdventureContext(t, tmpDir)

	// Create mock client
	mockClient := NewMockAnthropicClient()
	mockService := mockClient.GetMockMessagesService()
	mockService.SetResponse("What is the armor class formula?", "Armor Class (AC) is 10 + Dexterity modifier + armor bonus.")
	mockService.SetResponse("What about saving throws?", "Saving throws are d20 + ability modifier + proficiency bonus if proficient.")

	// Create agent manager with mock client
	personaLoader := NewPersonaLoader()
	logger, _ := NewLogger(tmpDir)
	clientFactory := func(apiKey string) anthropicClient {
		return mockClient
	}
	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

	// Create test personas
	createTestPersonas(t, am)

	// Test 1: Invoke rules-keeper
	response1, err := am.InvokeAgent("rules-keeper", "What is the armor class formula?", "", 1)
	if err != nil {
		t.Fatalf("Failed to invoke rules-keeper: %v", err)
	}

	if response1 == "" {
		t.Error("Expected non-empty response from rules-keeper")
	}

	// Verify agent was created and tracked
	state1, exists := am.GetNestedAgentState("rules-keeper")
	if !exists {
		t.Fatal("Expected rules-keeper to be tracked after invocation")
	}

	if state1.invocationCount != 1 {
		t.Errorf("Expected invocation count 1, got: %d", state1.invocationCount)
	}

	// Test 2: Second invocation should reuse same agent
	response2, err := am.InvokeAgent("rules-keeper", "What about saving throws?", "", 1)
	if err != nil {
		t.Fatalf("Failed second invocation: %v", err)
	}

	if response2 == "" {
		t.Error("Expected non-empty response from second invocation")
	}

	// Verify invocation count increased
	state2, _ := am.GetNestedAgentState("rules-keeper")
	if state2.invocationCount != 2 {
		t.Errorf("Expected invocation count 2, got: %d", state2.invocationCount)
	}

	// Verify conversation history was maintained
	messages := state2.conversationCtx.GetMessages()
	if len(messages) < 4 { // 2 user + 2 assistant messages
		t.Errorf("Expected at least 4 messages in history, got: %d", len(messages))
	}
}

// TestIntegration_MultipleAgents tests invoking different agents with mock client.
func TestIntegration_MultipleAgents(t *testing.T) {
	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	adventureCtx := createTestAdventureContext(t, tmpDir)

	// Create mock client with generic responses
	mockClient := NewMockAnthropicClient()
	personaLoader := NewPersonaLoader()
	logger, _ := NewLogger(tmpDir)
	clientFactory := func(apiKey string) anthropicClient {
		return mockClient
	}
	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

	createTestPersonas(t, am)

	// Invoke multiple agents
	agents := []string{"rules-keeper", "character-creator", "world-keeper"}
	for _, agentName := range agents {
		response, err := am.InvokeAgent(agentName, "Test question", "", 1)
		if err != nil {
			t.Errorf("Failed to invoke %s: %v", agentName, err)
			continue
		}

		if response == "" {
			t.Errorf("Expected non-empty response from %s", agentName)
		}
	}

	// Verify all agents are tracked
	activeAgents := am.ListNestedAgents()
	if len(activeAgents) != 3 {
		t.Errorf("Expected 3 active agents, got: %d", len(activeAgents))
	}
}

// TestIntegration_StatePersistenceAcrossSessions tests saving and loading agent states with mock client.
func TestIntegration_StatePersistenceAcrossSessions(t *testing.T) {
	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	adventureCtx := createTestAdventureContext(t, tmpDir)
	statesPath := filepath.Join(tmpDir, "agent-states.json")

	// Create shared mock client
	mockClient := NewMockAnthropicClient()
	clientFactory := func(apiKey string) anthropicClient {
		return mockClient
	}

	// Session 1: Create agent and invoke it
	{
		personaLoader := NewPersonaLoader()
		logger, _ := NewLogger(tmpDir)
		am1 := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

		createTestPersonas(t, am1)

		// Invoke agent
		_, err := am1.InvokeAgent("rules-keeper", "First question", "", 1)
		if err != nil {
			t.Fatalf("Failed first invocation: %v", err)
		}

		// Save states
		if err := am1.SaveAgentStates(statesPath); err != nil {
			t.Fatalf("Failed to save states: %v", err)
		}
	}

	// Verify states file exists
	if _, err := os.Stat(statesPath); os.IsNotExist(err) {
		t.Fatal("States file was not created")
	}

	// Session 2: Load states and verify continuity
	{
		personaLoader := NewPersonaLoader()
		logger, _ := NewLogger(tmpDir)
		am2 := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

		createTestPersonas(t, am2)

		// Load states
		if err := am2.LoadAgentStates(statesPath); err != nil {
			t.Fatalf("Failed to load states: %v", err)
		}

		// Verify agent state was restored
		state, exists := am2.GetNestedAgentState("rules-keeper")
		if !exists {
			t.Fatal("Expected rules-keeper to exist after loading")
		}

		if state.invocationCount != 1 {
			t.Errorf("Expected invocation count 1, got: %d", state.invocationCount)
		}

		// Note: Conversation history is not fully persisted (TODO in SerializeConversationContext)
		// so the conversation will start fresh, but metadata (invocation count, last invoked) is preserved

		// Make second invocation
		_, err := am2.InvokeAgent("rules-keeper", "Second question", "", 1)
		if err != nil {
			t.Fatalf("Failed second invocation after reload: %v", err)
		}

		// Verify count increased
		state2, _ := am2.GetNestedAgentState("rules-keeper")
		if state2.invocationCount != 2 {
			t.Errorf("Expected invocation count 2 after reload, got: %d", state2.invocationCount)
		}
	}
}

// TestIntegration_RecursionPrevention tests that nested agents cannot invoke other agents.
func TestIntegration_RecursionPrevention(t *testing.T) {
	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	apiKey := "test-key"
	adventureCtx := createTestAdventureContext(t, tmpDir)

	personaLoader := NewPersonaLoader()
	am := NewAgentManager(apiKey, adventureCtx, nil, nil, personaLoader)

	// Attempt to invoke at depth 2 (nested agent trying to invoke another)
	_, err := am.InvokeAgent("rules-keeper", "Test", "", 2)
	if err == nil {
		t.Fatal("Expected error for depth 2 invocation, got nil")
	}

	// Verify it's a recursion error
	var recursionErr *ErrRecursionLimit
	if e, ok := err.(*ErrRecursionLimit); ok {
		recursionErr = e
	}

	if recursionErr == nil {
		t.Errorf("Expected ErrRecursionLimit, got: %v", err)
	}

	if recursionErr.CurrentDepth != 2 {
		t.Errorf("Expected depth 2, got: %d", recursionErr.CurrentDepth)
	}

	if recursionErr.MaxDepth != 1 {
		t.Errorf("Expected max depth 1, got: %d", recursionErr.MaxDepth)
	}
}

// TestIntegration_InvalidAgentHandling tests proper error handling for invalid agents.
func TestIntegration_InvalidAgentHandling(t *testing.T) {
	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	apiKey := "test-key"
	adventureCtx := createTestAdventureContext(t, tmpDir)

	personaLoader := NewPersonaLoader()
	am := NewAgentManager(apiKey, adventureCtx, nil, nil, personaLoader)

	// Try to invoke dungeon-master (not allowed as nested agent)
	_, err := am.InvokeAgent("dungeon-master", "Test", "", 1)
	if err == nil {
		t.Fatal("Expected error for dungeon-master invocation, got nil")
	}

	var notFoundErr *ErrAgentNotFound
	if e, ok := err.(*ErrAgentNotFound); ok {
		notFoundErr = e
	}

	if notFoundErr == nil {
		t.Errorf("Expected ErrAgentNotFound, got: %v", err)
	}

	if notFoundErr.AgentName != "dungeon-master" {
		t.Errorf("Expected agent name 'dungeon-master', got: %s", notFoundErr.AgentName)
	}

	// Try completely invalid agent
	_, err = am.InvokeAgent("invalid-agent", "Test", "", 1)
	if err == nil {
		t.Fatal("Expected error for invalid agent, got nil")
	}

	if _, ok := err.(*ErrAgentNotFound); !ok {
		t.Errorf("Expected ErrAgentNotFound for invalid agent, got: %T", err)
	}
}

// TestIntegration_AgentStatistics tests statistics collection with mock client.
func TestIntegration_AgentStatistics(t *testing.T) {
	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	adventureCtx := createTestAdventureContext(t, tmpDir)

	mockClient := NewMockAnthropicClient()
	personaLoader := NewPersonaLoader()
	logger, _ := NewLogger(tmpDir)
	clientFactory := func(apiKey string) anthropicClient {
		return mockClient
	}
	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

	createTestPersonas(t, am)

	// Invoke agents multiple times
	am.InvokeAgent("rules-keeper", "Question 1", "", 1)
	am.InvokeAgent("rules-keeper", "Question 2", "", 1)
	am.InvokeAgent("character-creator", "Question 1", "", 1)

	// Get statistics
	stats := am.GetStatistics()

	totalAgents, ok := stats["total_agents"].(int)
	if !ok || totalAgents != 2 {
		t.Errorf("Expected total_agents=2, got: %v", stats["total_agents"])
	}

	agentStats, ok := stats["agents"].(map[string]map[string]interface{})
	if !ok {
		t.Fatal("Expected agents map in statistics")
	}

	// Verify rules-keeper stats
	rkStats, ok := agentStats["rules-keeper"]
	if !ok {
		t.Fatal("Expected rules-keeper in statistics")
	}

	invCount, _ := rkStats["invocation_count"].(int)
	if invCount != 2 {
		t.Errorf("Expected rules-keeper invocation_count=2, got: %v", invCount)
	}

	// Verify character-creator stats
	ccStats, ok := agentStats["character-creator"]
	if !ok {
		t.Fatal("Expected character-creator in statistics")
	}

	invCount2, _ := ccStats["invocation_count"].(int)
	if invCount2 != 1 {
		t.Errorf("Expected character-creator invocation_count=1, got: %v", invCount2)
	}
}

// TestIntegration_AgentClearing tests clearing individual and all agents with mock client.
func TestIntegration_AgentClearing(t *testing.T) {
	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	adventureCtx := createTestAdventureContext(t, tmpDir)

	mockClient := NewMockAnthropicClient()
	personaLoader := NewPersonaLoader()
	logger, _ := NewLogger(tmpDir)
	clientFactory := func(apiKey string) anthropicClient {
		return mockClient
	}
	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

	createTestPersonas(t, am)

	// Create multiple agents
	am.InvokeAgent("rules-keeper", "Test", "", 1)
	am.InvokeAgent("character-creator", "Test", "", 1)

	// Verify both exist
	if len(am.ListNestedAgents()) != 2 {
		t.Error("Expected 2 agents before clearing")
	}

	// Clear one agent
	am.ClearNestedAgent("rules-keeper")

	// Verify only character-creator remains
	agents := am.ListNestedAgents()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent after clearing rules-keeper, got: %d", len(agents))
	}

	if agents[0] != "character-creator" {
		t.Errorf("Expected character-creator to remain, got: %s", agents[0])
	}

	// Clear all agents
	am.ClearAllNestedAgents()

	// Verify none remain
	if len(am.ListNestedAgents()) != 0 {
		t.Error("Expected 0 agents after clearing all")
	}
}

// TestIntegration_LoggingOfInvocations tests that agent invocations are logged with mock client.
func TestIntegration_LoggingOfInvocations(t *testing.T) {
	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	adventureCtx := createTestAdventureContext(t, tmpDir)

	mockClient := NewMockAnthropicClient()
	personaLoader := NewPersonaLoader()
	logger, _ := NewLogger(tmpDir)
	clientFactory := func(apiKey string) anthropicClient {
		return mockClient
	}
	am := NewAgentManagerWithClientFactory("test-key", adventureCtx, logger, nil, personaLoader, clientFactory)

	createTestPersonas(t, am)

	// Invoke agent
	question := "What is the armor class formula?"
	am.InvokeAgent("rules-keeper", question, "Test context", 1)

	// Close logger to flush to disk
	logger.Close()

	// Wait a moment for file writes to complete
	time.Sleep(100 * time.Millisecond)

	// Read log file
	logFiles, err := filepath.Glob(filepath.Join(tmpDir, "sw-dm*.log"))
	if err != nil || len(logFiles) == 0 {
		t.Fatal("Expected log file to be created")
	}

	logData, err := os.ReadFile(logFiles[0])
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(logData)

	// Verify log contains agent invocation details
	if !contains(logContent, "AGENT INVOCATION") {
		t.Error("Expected log to contain 'AGENT INVOCATION'")
	}

	if !contains(logContent, "rules-keeper") {
		t.Error("Expected log to contain agent name")
	}

	if !contains(logContent, question) {
		t.Error("Expected log to contain question")
	}

	if !contains(logContent, "Test context") {
		t.Error("Expected log to contain context")
	}
}

// Helper functions

func setupIntegrationTest(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func createTestAdventureContext(t *testing.T, basePath string) *AdventureContext {
	t.Helper()

	adv := &adventure.Adventure{
		Name:        "Test Adventure",
		Description: "Integration test adventure",
	}
	adv.SetBasePath(basePath)

	return &AdventureContext{
		Adventure: adv,
		Party: &adventure.Party{
			Characters: []string{},
		},
		Inventory: &adventure.SharedInventory{
			Gold: 100,
		},
		State: &adventure.GameState{
			CurrentLocation: "Test Location",
		},
		Characters:    []*character.Character{},
		RecentJournal: []adventure.JournalEntry{},
	}
}

// TestIntegration_RealAPI_Optional is an optional test that verifies real Anthropic API integration.
// This test only runs when ANTHROPIC_API_KEY is set and can be slow.
// Use it to verify that the real API integration still works correctly.
func TestIntegration_RealAPI_Optional(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping real API test: ANTHROPIC_API_KEY not set (this is optional)")
	}

	// Only run this test if explicitly requested
	if os.Getenv("RUN_REAL_API_TESTS") == "" {
		t.Skip("Skipping real API test: RUN_REAL_API_TESTS not set (set to 1 to enable)")
	}

	tmpDir, cleanup := setupIntegrationTest(t)
	defer cleanup()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	adventureCtx := createTestAdventureContext(t, tmpDir)

	// Create agent manager with real API client
	personaLoader := NewPersonaLoader()
	logger, _ := NewLogger(tmpDir)
	am := NewAgentManager(apiKey, adventureCtx, logger, nil, personaLoader)

	// Create test personas
	createTestPersonas(t, am)

	// Test real API call
	response, err := am.InvokeAgent("rules-keeper", "What is the armor class formula in D&D 5e?", "", 1)
	if err != nil {
		t.Fatalf("Real API call failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response from real API")
	}

	t.Logf("Real API response: %s", response)

	// Verify agent was created and tracked
	state, exists := am.GetNestedAgentState("rules-keeper")
	if !exists {
		t.Fatal("Expected rules-keeper to be tracked after real API invocation")
	}

	if state.invocationCount != 1 {
		t.Errorf("Expected invocation count 1, got: %d", state.invocationCount)
	}

	// Verify metrics were tracked
	if state.metrics.TotalTokensUsed <= 0 {
		t.Error("Expected positive token usage from real API")
	}
}
