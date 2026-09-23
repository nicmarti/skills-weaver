package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OpenRouterTeam/go-sdk/retry"
)

func testConfig(t *testing.T) Config {
	t.Helper()
	cfg := DefaultModels()
	cfg.APIKey = "or-test-key"
	return cfg
}

func newRetryNone() *retry.Config {
	r := retry.Config{Strategy: "none"}
	return &r
}

// readRequestBody decodes a JSON request body into a generic map.
func readRequestBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode request body: %v", err)
	}
	return body
}

func chatResultFixture() string {
	return `{
		"id": "chatcmpl-test-1",
		"created": 1700000000,
		"object": "chat.completion",
		"model": "anthropic/claude-sonnet-5",
		"choices": [{
			"index": 0,
			"finish_reason": "tool_calls",
			"message": {
				"role": "assistant",
				"content": "Rolling the dice.",
				"tool_calls": [{
					"id": "call_1",
					"type": "function",
					"function": {"name": "roll_dice", "arguments": "{\"notation\":\"1d20+5\"}"}
				}]
			}
		}],
		"usage": {
			"prompt_tokens": 1200,
			"completion_tokens": 80,
			"total_tokens": 1280,
			"prompt_tokens_details": {"cached_tokens": 900, "cache_write_tokens": 0},
			"completion_tokens_details": {"reasoning_tokens": 12},
			"cost": 0.0042,
			"server_tool_use_details": {"tool_calls_requested": 1, "tool_calls_executed": 1}
		}
	}`
}

func TestComplete_SendsNeutralWireFormat(t *testing.T) {
	var capturedBody map[string]interface{}
	var capturedAuth, capturedReferer, capturedTitle string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody = readRequestBody(t, r)
		capturedAuth = r.Header.Get("Authorization")
		capturedReferer = r.Header.Get("HTTP-Referer")
		capturedTitle = r.Header.Get("X-Title")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(chatResultFixture()))
	}))
	defer server.Close()

	cfg := testConfig(t)
	cfg.HTTPReferer = "https://skillsweaver.example"
	cfg.AppName = "SkillsWeaver"
	client, err := newOpenRouterClient(cfg, nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	temp := 0.7
	req := Request{
		Model:               "sonnet",
		System:              "Tu es le Maître du Donjon.",
		MaxCompletionTokens: 4096,
		Temperature:         &temp,
		SessionID:           "adv-1/side-0",
		Messages: []Message{
			UserMessage("Attaque le gobelin"),
			UserMessageWithImage("Voici la carte", ImagePart{MediaType: "image/png", Base64: "QUJD"}),
			AssistantToolCallMessage("", []ToolCall{
				{ID: "call_9", Name: "get_monster", Arguments: json.RawMessage(`{"name":"goblin"}`)},
			}),
			ToolResultMessage(ToolResult{ToolCallID: "call_9", Content: `{"ac":13}`}),
		},
		Tools: []ToolDefinition{
			{
				Name:        "roll_dice",
				Description: "Lance des dés",
				Parameters: map[string]interface{}{
					"type":                 "object",
					"properties":           map[string]interface{}{"notation": map[string]string{"type": "string"}},
					"required":             []string{"notation"},
					"additionalProperties": false,
				},
			},
		},
		Advisor: &AdvisorConfig{Model: ModelOpus5, Instructions: "Conseiller stratégique.", ForwardTranscript: true, MaxCompletionTokens: 4096},
	}

	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		t.Fatalf("Complete() error: %v", err)
	}

	if capturedBody == nil {
		t.Fatal("request body was not captured")
	}
	if capturedAuth != "Bearer or-test-key" {
		t.Errorf("Authorization header = %q", capturedAuth)
	}
	if capturedReferer != "https://skillsweaver.example" {
		t.Errorf("HTTP-Referer header = %q", capturedReferer)
	}
	if capturedTitle != "SkillsWeaver" {
		t.Errorf("X-Title header = %q", capturedTitle)
	}

	if got := capturedBody["model"]; got != ModelSonnet5 {
		t.Errorf("wire model = %v, want %v (alias resolved)", got, ModelSonnet5)
	}
	if got := capturedBody["stream"].(bool); got {
		t.Error("stream should be false for Complete")
	}
	if got, ok := capturedBody["max_completion_tokens"].(float64); !ok || got != 4096 {
		t.Errorf("max_completion_tokens = %v", capturedBody["max_completion_tokens"])
	}
	if got, ok := capturedBody["session_id"].(string); !ok || got != "adv-1/side-0" {
		t.Errorf("session_id = %v", capturedBody["session_id"])
	}
	if got, ok := capturedBody["temperature"].(float64); !ok || got != 0.7 {
		t.Errorf("temperature = %v", capturedBody["temperature"])
	}

	msgs, ok := capturedBody["messages"].([]interface{})
	if !ok {
		t.Fatal("messages is not an array")
	}
	if len(msgs) != 5 {
		t.Fatalf("len(messages) = %d, want 5 (system + user + user image + assistant tools + tool result)", len(msgs))
	}

	system := msgs[0].(map[string]interface{})
	if system["role"] != "system" || system["content"] != "Tu es le Maître du Donjon." {
		t.Errorf("system message = %v", system)
	}

	userText := msgs[1].(map[string]interface{})
	if userText["role"] != "user" || userText["content"] != "Attaque le gobelin" {
		t.Errorf("user text message = %v", userText)
	}

	userImage := msgs[2].(map[string]interface{})
	if userImage["role"] != "user" {
		t.Fatalf("message 2 role = %v", userImage["role"])
	}
	parts, ok := userImage["content"].([]interface{})
	if !ok || len(parts) != 2 {
		t.Fatalf("user image content parts = %v", userImage["content"])
	}
	textPart := parts[0].(map[string]interface{})
	if textPart["type"] != "text" || textPart["text"] != "Voici la carte" {
		t.Errorf("text part = %v", textPart)
	}
	imgPart := parts[1].(map[string]interface{})
	if imgPart["type"] != "image_url" {
		t.Errorf("image part type = %v", imgPart["type"])
	}
	url := imgPart["image_url"].(map[string]interface{})["url"].(string)
	if !strings.HasPrefix(url, "data:image/png;base64,QUJD") {
		t.Errorf("image data URL = %q", url)
	}

	assistant := msgs[3].(map[string]interface{})
	calls, ok := assistant["tool_calls"].([]interface{})
	if !ok || len(calls) != 1 {
		t.Fatalf("assistant tool_calls = %v", assistant["tool_calls"])
	}
	call := calls[0].(map[string]interface{})
	if call["id"] != "call_9" {
		t.Errorf("tool call id = %v", call["id"])
	}
	fn := call["function"].(map[string]interface{})
	if fn["name"] != "get_monster" {
		t.Errorf("tool call name = %v", fn["name"])
	}
	if fn["arguments"] != `{"name":"goblin"}` {
		t.Errorf("tool call arguments = %v (want JSON string)", fn["arguments"])
	}

	toolMsg := msgs[4].(map[string]interface{})
	if toolMsg["role"] != "tool" || toolMsg["tool_call_id"] != "call_9" {
		t.Errorf("tool result message = %v", toolMsg)
	}

	tools, ok := capturedBody["tools"].([]interface{})
	if !ok || len(tools) != 2 {
		t.Fatalf("tools = %v (want local function + advisor)", capturedBody["tools"])
	}
	localTool := tools[0].(map[string]interface{})
	if localTool["type"] != "function" {
		t.Errorf("tools[0].type = %v", localTool["type"])
	}
	localFn := localTool["function"].(map[string]interface{})
	params := localFn["parameters"].(map[string]interface{})
	if _, ok := params["additionalProperties"]; !ok {
		t.Error("complete JSON Schema should be forwarded (additionalProperties missing)")
	}

	advisorTool := tools[1].(map[string]interface{})
	if advisorTool["type"] != "openrouter:advisor" {
		t.Errorf("tools[1].type = %v", advisorTool["type"])
	}
	advParams := advisorTool["parameters"].(map[string]interface{})
	if advParams["model"] != ModelOpus5 {
		t.Errorf("advisor model = %v", advParams["model"])
	}
	if advParams["forward_transcript"] != true {
		t.Errorf("advisor forward_transcript = %v", advParams["forward_transcript"])
	}
	if _, ok := advParams["max_completion_tokens"]; !ok {
		t.Error("advisor max_completion_tokens missing")
	}

	// Response mapping
	if resp.ID != "chatcmpl-test-1" {
		t.Errorf("resp.ID = %q", resp.ID)
	}
	if resp.Model != ModelSonnet5 {
		t.Errorf("resp.Model = %q", resp.Model)
	}
	if resp.FinishReason != FinishToolCalls {
		t.Errorf("resp.FinishReason = %q", resp.FinishReason)
	}
	if len(resp.Message.ToolCalls) != 1 || resp.Message.ToolCalls[0].Name != "roll_dice" {
		t.Fatalf("resp tool calls = %+v", resp.Message.ToolCalls)
	}
	if string(resp.Message.ToolCalls[0].Arguments) != `{"notation":"1d20+5"}` {
		t.Errorf("resp tool arguments = %s", resp.Message.ToolCalls[0].Arguments)
	}
	if !resp.Usage.Present {
		t.Fatal("usage should be present")
	}
	if resp.Usage.PromptTokens != 1200 || resp.Usage.CompletionTokens != 80 || resp.Usage.TotalTokens != 1280 {
		t.Errorf("usage tokens = %+v", resp.Usage)
	}
	if resp.Usage.CachedTokens != 900 || resp.Usage.ReasoningTokens != 12 {
		t.Errorf("usage details = %+v", resp.Usage)
	}
	if resp.Usage.Cost != 0.0042 {
		t.Errorf("usage cost = %v", resp.Usage.Cost)
	}
	if resp.Usage.ServerToolCallsRequested != 1 || resp.Usage.ServerToolCallsExecuted != 1 {
		t.Errorf("server tool usage = %+v", resp.Usage)
	}
}

func TestChatRequest_RequireParameters(t *testing.T) {
	req, err := buildChatRequest(Request{
		Model:             ModelSonnet5,
		Messages:          []Message{UserMessage("test")},
		RequireParameters: true,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	provider, ok := decoded["provider"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected provider field in request: %s", string(body))
	}
	if reqParam, ok := provider["require_parameters"].(bool); !ok || !reqParam {
		t.Errorf("require_parameters = %v, want true", provider["require_parameters"])
	}
}

func TestChatRequest_PersistedImageReferenceWithoutBytesIsTextOnly(t *testing.T) {
	req, err := buildChatRequest(Request{
		Model: ModelSonnet5,
		Messages: []Message{UserMessageWithImage("[image non persistée]", ImagePart{
			ResourceRef: "world-map",
		})},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	messages := decoded["messages"].([]interface{})
	if content := messages[0].(map[string]interface{})["content"]; content != "[image non persistée]" {
		t.Fatalf("unavailable image reference must not produce an invalid data URL: %v", content)
	}
}

func TestComplete_ZeroChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"chatcmpl-0","object":"chat.completion","model":"m","choices":[]}`))
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	_, err = client.Complete(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}})
	if err == nil {
		t.Fatal("Complete() should fail on zero choices")
	}
	if llmErr, ok := err.(*Error); ok && llmErr.Kind != ErrKindInvalidRequest {
		t.Errorf("kind = %q, want invalid_request", llmErr.Kind)
	}
}

func TestComplete_ErrorNormalization(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		wantKind ErrorKind
	}{
		{"401 auth", 401, `{"error":{"code":401,"message":"Invalid API key"}}`, ErrKindAuth},
		{"402 payment", 402, `{"error":{"code":402,"message":"Insufficient credits"}}`, ErrKindPayment},
		{"429 rate limit", 429, `{"error":{"code":429,"message":"Rate limit exceeded"}}`, ErrKindRateLimit},
		{"529 overloaded", 529, `{"error":{"code":529,"message":"Provider returned error"}}`, ErrKindOverloaded},
		{"400 invalid request", 400, `{"error":{"code":400,"message":"Invalid request"}}`, ErrKindInvalidRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer server.Close()

			client, err := newOpenRouterClient(testConfig(t), newRetryNone(), server.URL)
			if err != nil {
				t.Fatalf("newOpenRouterClient: %v", err)
			}

			_, callErr := client.Complete(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}})
			if callErr == nil {
				t.Fatalf("Complete() should fail with status %d", tc.status)
			}
			llmErr, ok := callErr.(*Error)
			if !ok {
				t.Fatalf("error type = %T, want *llm.Error", callErr)
			}
			if llmErr.Kind != tc.wantKind {
				t.Errorf("kind = %q, want %q", llmErr.Kind, tc.wantKind)
			}
			if llmErr.Status != tc.status {
				t.Errorf("status = %d, want %d", llmErr.Status, tc.status)
			}
		})
	}
}

func TestComplete_APIErrorFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(500)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), newRetryNone(), server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	_, err = client.Complete(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}})
	if err == nil {
		t.Fatal("Complete() should fail on 500 text/plain")
	}
	llmErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T, want *llm.Error", err)
	}
	if llmErr.Kind != ErrKindProvider {
		t.Errorf("kind = %q, want provider", llmErr.Kind)
	}
	if !llmErr.Retryable {
		t.Error("500 should be retryable")
	}
}

func TestComplete_Retries503ThenSucceeds(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		if hits == 1 {
			w.WriteHeader(503)
			w.Write([]byte(`{"error":{"code":503,"message":"Service temporarily unavailable"}}`))
			return
		}
		w.Write([]byte(chatResultFixture()))
	}))
	defer server.Close()

	client, err := newOpenRouterClient(testConfig(t), nil, server.URL)
	if err != nil {
		t.Fatalf("newOpenRouterClient: %v", err)
	}

	resp, err := client.Complete(context.Background(), Request{Model: ModelSonnet5, Messages: []Message{UserMessage("hi")}})
	if err != nil {
		t.Fatalf("Complete() should succeed after 503 retry: %v", err)
	}
	if resp.ID != "chatcmpl-test-1" {
		t.Errorf("resp.ID = %q", resp.ID)
	}
}
