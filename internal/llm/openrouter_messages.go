package llm

import (
	"encoding/json"
	"fmt"

	oropenrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/optionalnullable"
)

// buildChatRequest converts a neutral Request into the SDK ChatRequest.
// Aliases are resolved centrally so every caller sends concrete model IDs.
func buildChatRequest(req Request, stream bool) (components.ChatRequest, error) {
	model := ResolveModel(req.Model)

	messages, err := messagesToChat(req)
	if err != nil {
		return components.ChatRequest{}, err
	}

	chatReq := components.ChatRequest{
		Model:    oropenrouter.String(model),
		Messages: messages,
		Stream:   oropenrouter.Bool(stream),
	}

	if req.MaxCompletionTokens > 0 {
		chatReq.MaxCompletionTokens = optionalnullable.From(oropenrouter.Int64(req.MaxCompletionTokens))
	}
	if req.Temperature != nil {
		chatReq.Temperature = optionalnullable.From(oropenrouter.Pointer(*req.Temperature))
	}
	if req.SessionID != "" {
		chatReq.SessionID = oropenrouter.String(req.SessionID)
	}
	if req.RequireParameters {
		chatReq.Provider = optionalnullable.From(oropenrouter.Pointer(components.ProviderPreferences{
			RequireParameters: optionalnullable.From(oropenrouter.Bool(true)),
		}))
	}
	if req.JSONResponse != nil {
		rf, err := jsonResponseFormatToChat(*req.JSONResponse)
		if err != nil {
			return components.ChatRequest{}, err
		}
		chatReq.ResponseFormat = &rf
	}

	tools := toolsToChat(req)
	if len(tools) > 0 {
		chatReq.Tools = tools
	}

	return chatReq, nil
}

// jsonResponseFormatToChat maps the neutral JSON control onto the SDK union.
// A schema selects the strict json_schema mode; otherwise the lighter
// json_object mode only guarantees valid JSON.
func jsonResponseFormatToChat(fmtCfg JSONResponseFormat) (components.ResponseFormat, error) {
	if fmtCfg.Schema != nil {
		name := fmtCfg.Name
		if name == "" {
			name = "skillsweaver_response"
		}
		schema := components.ChatFormatJSONSchemaConfig{
			Type: components.ChatFormatJSONSchemaConfigTypeJSONSchema,
			JSONSchema: components.ChatJSONSchemaConfig{
				Name:   name,
				Schema: fmtCfg.Schema,
			},
		}
		if fmtCfg.Strict {
			schema.JSONSchema.Strict = optionalnullable.From(oropenrouter.Bool(true))
		}
		return components.CreateResponseFormatJSONSchema(schema), nil
	}
	return components.CreateResponseFormatJSONObject(components.ChatFormatJSONObjectConfig{
		Type: components.ChatFormatJSONObjectConfigTypeJSONObject,
	}), nil
}

// messagesToChat converts neutral history into Chat wire messages. The system
// prompt is prepended as the first message. A single neutral message can expand
// into several wire messages (a tool-role message emits one message per result).
func messagesToChat(req Request) ([]components.ChatMessages, error) {
	messages := make([]components.ChatMessages, 0, len(req.Messages)+1)

	if req.System != "" {
		system := components.ChatSystemMessage{
			Role:    components.ChatSystemMessageRoleSystem,
			Content: components.CreateChatSystemMessageContentStr(req.System),
		}
		messages = append(messages, components.CreateChatMessagesSystem(system))
	}

	for _, msg := range req.Messages {
		switch msg.Role {
		case RoleSystem:
			system := components.ChatSystemMessage{
				Role:    components.ChatSystemMessageRoleSystem,
				Content: components.CreateChatSystemMessageContentStr(msg.Text),
			}
			messages = append(messages, components.CreateChatMessagesSystem(system))
		case RoleUser:
			messages = append(messages, userMessageToChat(msg))
		case RoleAssistant:
			assistant, err := assistantMessageToChat(msg)
			if err != nil {
				return nil, err
			}
			messages = append(messages, assistant)
		case RoleTool:
			for _, result := range msg.ToolResults {
				tool := components.ChatToolMessage{
					Role:       components.ChatToolMessageRoleTool,
					ToolCallID: result.ToolCallID,
					Content:    components.CreateChatToolMessageContentStr(result.Content),
				}
				messages = append(messages, components.CreateChatMessagesTool(tool))
			}
		default:
			return nil, fmt.Errorf("llm: unsupported message role %q", msg.Role)
		}
	}

	return messages, nil
}

func userMessageToChat(msg Message) components.ChatMessages {
	content := components.ChatUserMessageContent{}
	images := make([]ImagePart, 0, len(msg.Images))
	for _, img := range msg.Images {
		if img.Base64 != "" {
			images = append(images, img)
		}
	}
	if len(images) == 0 {
		content = components.CreateChatUserMessageContentStr(msg.Text)
	} else {
		items := []components.ChatContentItems{
			components.CreateChatContentItemsText(components.ChatContentText{Text: msg.Text}),
		}
		for _, img := range images {
			items = append(items, components.CreateChatContentItemsImageURL(components.ChatContentImage{
				ImageURL: components.ChatContentImageImageURL{URL: img.DataURL()},
			}))
		}
		content = components.CreateChatUserMessageContentArrayOfChatContentItems(items)
	}

	user := components.ChatUserMessage{
		Role:    components.ChatUserMessageRoleUser,
		Content: content,
	}
	return components.CreateChatMessagesUser(user)
}

func assistantMessageToChat(msg Message) (components.ChatMessages, error) {
	assistant := components.ChatAssistantMessage{
		Role: components.ChatAssistantMessageRoleAssistant,
	}

	// Leave content unset for tool-call-only assistant messages; Chat wire
	// format allows assistant messages without content.
	if msg.Text != "" {
		text := msg.Text
		assistant.Content = optionalnullable.From(oropenrouter.Pointer(components.CreateChatAssistantMessageContentStr(text)))
	}

	for _, call := range msg.ToolCalls {
		args := string(call.Arguments)
		if args == "" || args == "null" {
			args = "{}"
		}
		assistant.ToolCalls = append(assistant.ToolCalls, components.ChatToolCall{
			ID:   call.ID,
			Type: components.ChatToolCallTypeFunction,
			Function: components.ChatToolCallFunction{
				Name:      call.Name,
				Arguments: args,
			},
		})
	}

	return components.CreateChatMessagesAssistant(assistant), nil
}

// toolsToChat converts neutral tool definitions to Chat function tools.
func toolsToChat(req Request) []components.ChatFunctionTool {
	tools := make([]components.ChatFunctionTool, 0, len(req.Tools))

	for _, tool := range req.Tools {
		def := components.ChatFunctionToolFunctionFunction{
			Name:       tool.Name,
			Parameters: tool.Parameters,
		}
		if tool.Description != "" {
			def.Description = oropenrouter.String(tool.Description)
		}
		tools = append(tools, components.CreateChatFunctionToolChatFunctionToolFunction(
			components.ChatFunctionToolFunction{
				Type:     components.ChatFunctionToolTypeFunction,
				Function: def,
			},
		))
	}

	return tools
}

// usageFromChat maps SDK usage accounting onto neutral Usage.
func usageFromChat(u *components.ChatUsage) Usage {
	usage := Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}

	if details, ok := u.GetPromptTokensDetails().Get(); ok && details != nil {
		if details.CachedTokens != nil {
			usage.CachedTokens = *details.CachedTokens
		}
		if details.CacheWriteTokens != nil {
			usage.CacheWriteTokens = *details.CacheWriteTokens
		}
	}
	if details, ok := u.GetCompletionTokensDetails().Get(); ok && details != nil {
		if v, ok := details.GetReasoningTokens().Get(); ok && v != nil {
			usage.ReasoningTokens = *v
		}
	}
	if v, ok := u.GetCost().Get(); ok && v != nil {
		usage.Cost = *v
	}
	if details, ok := u.GetServerToolUseDetails().Get(); ok && details != nil {
		if v, ok := details.GetToolCallsRequested().Get(); ok && v != nil {
			usage.ServerToolCallsRequested = *v
		}
		if v, ok := details.GetToolCallsExecuted().Get(); ok && v != nil {
			usage.ServerToolCallsExecuted = *v
		}
	}

	usage.Present = true
	return usage
}

// chatResultToResponse maps a non-streaming SDK result onto the neutral type.
func chatResultToResponse(r *components.ChatResult) (Response, error) {
	resp := Response{ID: r.ID, Model: r.Model}

	if r.Usage != nil {
		resp.Usage = usageFromChat(r.Usage)
	}

	if len(r.Choices) == 0 {
		return Response{}, &Error{Kind: ErrKindInvalidRequest, Status: 200, Message: "response contained no choices"}
	}
	choice := r.Choices[0]
	resp.FinishReason = mapFinishReason(choice.FinishReason)

	resp.Message.Role = RoleAssistant
	if content, ok := choice.Message.Content.Get(); ok && content != nil {
		if content.Str != nil {
			resp.Message.Text = *content.Str
		}
		for _, item := range content.ArrayOfChatContentItems {
			if item.ChatContentText != nil {
				resp.Message.Text += item.ChatContentText.Text
			}
		}
	}

	for _, call := range choice.Message.ToolCalls {
		args := json.RawMessage(call.Function.Arguments)
		if len(args) == 0 {
			args = json.RawMessage("{}")
		}
		resp.Message.ToolCalls = append(resp.Message.ToolCalls, ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: args,
		})
	}

	if len(resp.Message.ToolCalls) > 0 && resp.FinishReason == FinishNone {
		resp.FinishReason = FinishToolCalls
	}

	return resp, nil
}

func mapFinishReason(reason *components.ChatFinishReasonEnum) FinishReason {
	if reason == nil {
		return FinishNone
	}
	switch *reason {
	case components.ChatFinishReasonEnumStop:
		return FinishStop
	case components.ChatFinishReasonEnumToolCalls:
		return FinishToolCalls
	case components.ChatFinishReasonEnumLength:
		return FinishLength
	case components.ChatFinishReasonEnumContentFilter:
		return FinishContentFilter
	case components.ChatFinishReasonEnumError:
		return FinishError
	default:
		return FinishReason(string(*reason))
	}
}
