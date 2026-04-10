package wellapi

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/QingsiLiu/baseComponents/service/llm"
)

type openAIEndpoint string

const (
	openAIEndpointResponses openAIEndpoint = "responses"
	openAIEndpointChat      openAIEndpoint = "chat"
)

// OpenAIService WellAPI OpenAI 兼容服务实现
type OpenAIService struct {
	client *Client
}

// NewOpenAIService 创建默认服务实例
func NewOpenAIService() llm.LLMService {
	return &OpenAIService{
		client: NewClient(),
	}
}

// NewOpenAIServiceWithKey 使用指定 API Key 创建服务实例
func NewOpenAIServiceWithKey(apiKey string) llm.LLMService {
	return &OpenAIService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 返回服务来源标识
func (s *OpenAIService) Source() string {
	return llm.SourceWellAPIOpenAI
}

// Generate 执行单次生成请求
func (s *OpenAIService) Generate(req *llm.GenerateReq) (*llm.GenerateResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if len(req.Messages) == 0 && strings.TrimSpace(req.SystemInstruction) == "" {
		return nil, fmt.Errorf("at least one message or system instruction is required")
	}
	if err := validateOpenAIReq(req); err != nil {
		return nil, err
	}

	models, explicitModel := s.resolveModels(req)
	errorsByModel := make([]error, 0, len(models))

	for index, model := range models {
		resp, err := s.generateWithRetry(req, model)
		if err == nil {
			return resp, nil
		}

		errorsByModel = append(errorsByModel, fmt.Errorf("%s: %w", model, err))
		if explicitModel || !isRetryableError(err) || index == len(models)-1 {
			break
		}
	}

	return nil, stderrors.Join(errorsByModel...)
}

func validateOpenAIReq(req *llm.GenerateReq) error {
	switch {
	case req.EnableURLContext:
		return fmt.Errorf("openai provider does not support EnableURLContext")
	case req.EnableGoogleSearch:
		return fmt.Errorf("openai provider does not support EnableGoogleSearch")
	case req.EnableCodeExecution:
		return fmt.Errorf("openai provider does not support EnableCodeExecution")
	}

	responseMIMEType := strings.TrimSpace(req.ResponseMIMEType)
	if responseMIMEType != "" && responseMIMEType != "application/json" {
		return fmt.Errorf("openai provider only supports application/json response mime type")
	}

	return nil
}

func (s *OpenAIService) resolveModels(req *llm.GenerateReq) ([]string, bool) {
	model := strings.TrimSpace(req.Model)
	if model != "" {
		return []string{model}, true
	}

	return []string{
		ModelGPT54Mini,
		ModelGPT5Mini20250807,
		ModelGPT54Nano,
	}, false
}

func (s *OpenAIService) generateWithRetry(req *llm.GenerateReq, model string) (*llm.GenerateResp, error) {
	attempts := s.client.retryMax
	if attempts <= 0 {
		attempts = DefaultRetryMax
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		resp, err := s.generateForModel(req, model)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if attempt == attempts || !isRetryableOpenAIError(err) {
			break
		}

		delay := s.client.retryBaseDelay
		if delay <= 0 {
			delay = DefaultRetryDelay
		}

		backoff := delay * time.Duration(1<<(attempt-1))
		timer := time.NewTimer(backoff)
		select {
		case <-timer.C:
		case <-context.Background().Done():
			timer.Stop()
		}
	}

	return nil, lastErr
}

func (s *OpenAIService) generateForModel(req *llm.GenerateReq, model string) (*llm.GenerateResp, error) {
	endpoints := preferredOpenAIEndpoints(model)
	errorsByEndpoint := make([]error, 0, len(endpoints))

	for index, endpoint := range endpoints {
		resp, err := s.generateOnEndpoint(req, model, endpoint)
		if err == nil {
			return resp, nil
		}

		errorsByEndpoint = append(errorsByEndpoint, fmt.Errorf("%s: %w", endpoint, err))
		if index == len(endpoints)-1 || !isEndpointCompatibilityError(err) {
			break
		}
	}

	return nil, stderrors.Join(errorsByEndpoint...)
}

func preferredOpenAIEndpoints(model string) []openAIEndpoint {
	switch strings.TrimSpace(model) {
	case ModelGPT54:
		return []openAIEndpoint{openAIEndpointResponses}
	case ModelGPT54Mini,
		ModelGPT54Mini20260317,
		ModelGPT54Nano,
		ModelGPT54Nano20260317,
		ModelGPT5Mini20250807,
		ModelGPT5Nano20250807:
		return []openAIEndpoint{openAIEndpointResponses, openAIEndpointChat}
	default:
		return []openAIEndpoint{openAIEndpointResponses, openAIEndpointChat}
	}
}

func (s *OpenAIService) generateOnEndpoint(req *llm.GenerateReq, model string, endpoint openAIEndpoint) (*llm.GenerateResp, error) {
	switch endpoint {
	case openAIEndpointResponses:
		wireReq, err := buildResponsesRequest(req, model)
		if err != nil {
			return nil, err
		}
		wireResp, err := s.client.CreateResponse(wireReq)
		if err != nil {
			return nil, err
		}
		return convertResponsesResponse(wireResp)
	case openAIEndpointChat:
		wireReq, err := buildChatCompletionRequest(req, model)
		if err != nil {
			return nil, err
		}
		wireResp, err := s.client.CreateChatCompletion(wireReq)
		if err != nil {
			return nil, err
		}
		return convertChatCompletionResponse(wireResp)
	default:
		return nil, fmt.Errorf("unsupported openai endpoint: %s", endpoint)
	}
}

func buildChatCompletionRequest(req *llm.GenerateReq, model string) (*ChatCompletionRequest, error) {
	messages, err := buildChatMessages(req)
	if err != nil {
		return nil, err
	}

	responseFormat, err := buildChatResponseFormat(req)
	if err != nil {
		return nil, err
	}

	wireReq := &ChatCompletionRequest{
		Model:          model,
		Messages:       messages,
		Temperature:    req.Temperature,
		TopP:           req.TopP,
		MaxTokens:      req.MaxOutputTokens,
		Tools:          buildChatTools(req.Tools),
		ResponseFormat: responseFormat,
	}
	if len(wireReq.Tools) > 0 {
		wireReq.ToolChoice = "auto"
	}

	return wireReq, nil
}

func buildResponsesRequest(req *llm.GenerateReq, model string) (*ResponsesRequest, error) {
	input, err := buildResponsesInput(req.Messages)
	if err != nil {
		return nil, err
	}

	textConfig, err := buildResponsesTextConfig(req)
	if err != nil {
		return nil, err
	}

	wireReq := &ResponsesRequest{
		Model:           model,
		Instructions:    strings.TrimSpace(req.SystemInstruction),
		Input:           input,
		Tools:           buildResponsesTools(req.Tools),
		Text:            textConfig,
		Temperature:     req.Temperature,
		TopP:            req.TopP,
		MaxOutputTokens: req.MaxOutputTokens,
		Reasoning:       buildResponsesReasoning(req.ThinkingBudget),
	}
	if len(wireReq.Tools) > 0 {
		wireReq.ToolChoice = "auto"
	}

	return wireReq, nil
}

func buildChatMessages(req *llm.GenerateReq) ([]ChatCompletionMessage, error) {
	messages := make([]ChatCompletionMessage, 0, len(req.Messages)+1)

	if strings.TrimSpace(req.SystemInstruction) != "" {
		messages = append(messages, ChatCompletionMessage{
			Role: "system",
			Content: []ChatCompletionContentPart{
				{Type: "text", Text: req.SystemInstruction},
			},
		})
	}

	for _, message := range req.Messages {
		content, err := buildChatMessageContent(message.Parts)
		if err != nil {
			return nil, err
		}
		if len(content) == 0 {
			continue
		}

		messages = append(messages, ChatCompletionMessage{
			Role:    message.Role,
			Content: content,
		})
	}

	return messages, nil
}

func buildChatMessageContent(parts []llm.Part) ([]ChatCompletionContentPart, error) {
	content := make([]ChatCompletionContentPart, 0, len(parts))

	for _, part := range parts {
		if strings.TrimSpace(part.Text) != "" {
			content = append(content, ChatCompletionContentPart{
				Type: "text",
				Text: part.Text,
			})
		}
		if strings.TrimSpace(part.InlineDataBase64) != "" {
			url, err := dataURLForPart(part)
			if err != nil {
				return nil, err
			}
			content = append(content, ChatCompletionContentPart{
				Type:     "image_url",
				ImageURL: &ChatCompletionImageURL{URL: url},
			})
		}
	}

	return content, nil
}

func buildResponsesInput(messages []llm.Message) ([]ResponsesInputItem, error) {
	input := make([]ResponsesInputItem, 0, len(messages))

	for _, message := range messages {
		content, err := buildResponsesInputContent(message.Parts)
		if err != nil {
			return nil, err
		}
		if len(content) == 0 {
			continue
		}

		input = append(input, ResponsesInputItem{
			Role:    message.Role,
			Content: content,
		})
	}

	return input, nil
}

func buildResponsesInputContent(parts []llm.Part) ([]ResponsesInputContent, error) {
	content := make([]ResponsesInputContent, 0, len(parts))

	for _, part := range parts {
		if strings.TrimSpace(part.Text) != "" {
			content = append(content, ResponsesInputContent{
				Type: "input_text",
				Text: part.Text,
			})
		}
		if strings.TrimSpace(part.InlineDataBase64) != "" {
			url, err := dataURLForPart(part)
			if err != nil {
				return nil, err
			}
			content = append(content, ResponsesInputContent{
				Type:     "input_image",
				ImageURL: url,
			})
		}
	}

	return content, nil
}

func buildChatTools(specs []llm.ToolSpec) []ChatCompletionTool {
	if len(specs) == 0 {
		return nil
	}

	tools := make([]ChatCompletionTool, 0, len(specs))
	for _, tool := range specs {
		tools = append(tools, ChatCompletionTool{
			Type: "function",
			Function: ChatCompletionToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}

	return tools
}

func buildResponsesTools(specs []llm.ToolSpec) []ResponsesTool {
	if len(specs) == 0 {
		return nil
	}

	tools := make([]ResponsesTool, 0, len(specs))
	for _, tool := range specs {
		tools = append(tools, ResponsesTool{
			Type:        "function",
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		})
	}

	return tools
}

func buildChatResponseFormat(req *llm.GenerateReq) (*ChatCompletionRespFormat, error) {
	responseMIMEType := strings.TrimSpace(req.ResponseMIMEType)

	if len(req.ResponseSchema) > 0 {
		schema := normalizeOpenAIJSONSchema(req.ResponseSchema)
		return &ChatCompletionRespFormat{
			Type: "json_schema",
			JSONSchema: &ChatCompletionRespJSONSchema{
				Name:   "response",
				Schema: schema,
				Strict: true,
			},
		}, nil
	}

	if responseMIMEType == "application/json" {
		return &ChatCompletionRespFormat{Type: "json_object"}, nil
	}

	return nil, nil
}

func buildResponsesTextConfig(req *llm.GenerateReq) (*ResponsesTextConfig, error) {
	responseMIMEType := strings.TrimSpace(req.ResponseMIMEType)

	if len(req.ResponseSchema) > 0 {
		schema := normalizeOpenAIJSONSchema(req.ResponseSchema)
		return &ResponsesTextConfig{
			Format: &ResponsesTextFormat{
				Type:   "json_schema",
				Name:   "response",
				Schema: schema,
				Strict: true,
			},
		}, nil
	}

	if responseMIMEType == "application/json" {
		return &ResponsesTextConfig{
			Format: &ResponsesTextFormat{Type: "json_object"},
		}, nil
	}

	return nil, nil
}

func buildResponsesReasoning(thinkingBudget *int) *ResponsesReasoning {
	if thinkingBudget == nil || *thinkingBudget <= 0 {
		return nil
	}

	return &ResponsesReasoning{
		Effort: thinkingBudgetToEffort(*thinkingBudget),
	}
}

func thinkingBudgetToEffort(thinkingBudget int) string {
	switch {
	case thinkingBudget <= 1024:
		return "minimal"
	case thinkingBudget <= 8192:
		return "low"
	case thinkingBudget <= 24576:
		return "medium"
	default:
		return "high"
	}
}

func convertChatCompletionResponse(resp *ChatCompletionResponse) (*llm.GenerateResp, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("response choices are empty")
	}

	raw := resp.Raw
	if len(raw) == 0 {
		marshaled, err := json.Marshal(resp)
		if err == nil {
			raw = marshaled
		}
	}

	choice := resp.Choices[0]
	result := &llm.GenerateResp{
		ModelVersion: resp.Model,
		FinishReason: choice.FinishReason,
		Raw:          raw,
	}

	if resp.Usage != nil {
		result.Usage = llm.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	text, parts := extractChatMessageText(choice.Message.Content)
	result.Text = text
	result.Parts = append(result.Parts, parts...)

	functionCalls, respParts, err := convertChatToolCalls(choice.Message.ToolCalls)
	if err != nil {
		return nil, err
	}
	result.FunctionCalls = append(result.FunctionCalls, functionCalls...)
	result.Parts = append(result.Parts, respParts...)

	return result, nil
}

func convertResponsesResponse(resp *ResponsesResponse) (*llm.GenerateResp, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	raw := resp.Raw
	if len(raw) == 0 {
		marshaled, err := json.Marshal(resp)
		if err == nil {
			raw = marshaled
		}
	}

	result := &llm.GenerateResp{
		ModelVersion: resp.Model,
		FinishReason: resp.Status,
		Raw:          raw,
	}

	if resp.Usage != nil {
		promptTokens := resp.Usage.PromptTokens
		if promptTokens == 0 {
			promptTokens = resp.Usage.InputTokens
		}
		completionTokens := resp.Usage.CompletionTokens
		if completionTokens == 0 {
			completionTokens = resp.Usage.OutputTokens
		}
		result.Usage = llm.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
		if resp.Usage.OutputTokensDetails != nil {
			result.Usage.ThoughtsTokens = resp.Usage.OutputTokensDetails.ReasoningTokens
		}
	}

	if strings.TrimSpace(resp.OutputText) != "" {
		result.Text = resp.OutputText
		result.Parts = append(result.Parts, llm.RespPart{Text: resp.OutputText})
	}

	for _, item := range resp.Output {
		switch item.Type {
		case "function_call":
			call, err := parseFunctionCall(item.Name, item.Arguments, item.CallID)
			if err != nil {
				return nil, err
			}
			result.FunctionCalls = append(result.FunctionCalls, call)
			result.Parts = append(result.Parts, llm.RespPart{FunctionCall: &call})
		default:
			if result.Text != "" {
				continue
			}
			for _, content := range item.Content {
				if content.Text == "" {
					continue
				}
				result.Text += content.Text
				result.Parts = append(result.Parts, llm.RespPart{Text: content.Text})
			}
		}
	}

	return result, nil
}

func extractChatMessageText(raw json.RawMessage) (string, []llm.RespPart) {
	if len(raw) == 0 {
		return "", nil
	}

	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		if plain == "" {
			return "", nil
		}
		return plain, []llm.RespPart{{Text: plain}}
	}

	var parts []ChatCompletionOutputPart
	if err := json.Unmarshal(raw, &parts); err == nil {
		var builder strings.Builder
		respParts := make([]llm.RespPart, 0, len(parts))
		for _, part := range parts {
			if part.Text == "" {
				continue
			}
			builder.WriteString(part.Text)
			respParts = append(respParts, llm.RespPart{Text: part.Text})
		}
		return builder.String(), respParts
	}

	return "", nil
}

func convertChatToolCalls(toolCalls []ChatCompletionToolCall) ([]llm.FunctionCall, []llm.RespPart, error) {
	if len(toolCalls) == 0 {
		return nil, nil, nil
	}

	functionCalls := make([]llm.FunctionCall, 0, len(toolCalls))
	respParts := make([]llm.RespPart, 0, len(toolCalls))

	for _, toolCall := range toolCalls {
		call, err := parseFunctionCall(toolCall.Function.Name, toolCall.Function.Arguments, toolCall.ID)
		if err != nil {
			return nil, nil, err
		}
		functionCalls = append(functionCalls, call)
		respParts = append(respParts, llm.RespPart{FunctionCall: &call})
	}

	return functionCalls, respParts, nil
}

func parseFunctionCall(name, arguments, id string) (llm.FunctionCall, error) {
	call := llm.FunctionCall{
		Name: name,
		Args: map[string]any{},
		ID:   id,
	}

	if strings.TrimSpace(arguments) == "" {
		return call, nil
	}

	if err := json.Unmarshal([]byte(arguments), &call.Args); err != nil {
		return llm.FunctionCall{}, fmt.Errorf("decode function call arguments: %w", err)
	}

	return call, nil
}

func dataURLForPart(part llm.Part) (string, error) {
	mimeType := strings.TrimSpace(part.MimeType)
	if mimeType == "" {
		return "", fmt.Errorf("mime type is required for inline data")
	}

	return "data:" + mimeType + ";base64," + part.InlineDataBase64, nil
}

func isRetryableOpenAIError(err error) bool {
	if isRetryableError(err) {
		return true
	}

	var netErr net.Error
	if stderrors.As(err, &netErr) {
		return true
	}

	return stderrors.Is(err, context.DeadlineExceeded)
}

func isEndpointCompatibilityError(err error) bool {
	var apiErr *APIError
	if !stderrors.As(err, &apiErr) || apiErr == nil {
		return false
	}

	if apiErr.StatusCode != 400 && apiErr.StatusCode != 404 {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(apiErr.Message + " " + apiErr.Raw))
	patterns := []string{
		"unsupported",
		"not support",
		"does not support",
		"only supports",
		"responses api",
		"chat completions",
		"chat/completions",
		"invalid model",
		"endpoint",
	}
	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}

	return false
}

func normalizeOpenAIJSONSchema(schema map[string]any) map[string]any {
	normalized, ok := deepCopyValue(schema).(map[string]any)
	if !ok {
		return map[string]any{}
	}

	normalizeSchemaNode(normalized)
	return normalized
}

func deepCopyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		copied := make(map[string]any, len(typed))
		for key, child := range typed {
			copied[key] = deepCopyValue(child)
		}
		return copied
	case []any:
		copied := make([]any, len(typed))
		for index, child := range typed {
			copied[index] = deepCopyValue(child)
		}
		return copied
	case []string:
		copied := make([]any, len(typed))
		for index, child := range typed {
			copied[index] = child
		}
		return copied
	default:
		return value
	}
}

func normalizeSchemaNode(node map[string]any) {
	if node == nil {
		return
	}

	if rawType, ok := node["type"]; ok {
		node["type"] = normalizeSchemaType(rawType)
	}

	if isObjectSchema(node["type"]) {
		if _, ok := node["additionalProperties"]; !ok {
			node["additionalProperties"] = false
		}
	}

	if properties, ok := node["properties"].(map[string]any); ok {
		for key, value := range properties {
			child, ok := value.(map[string]any)
			if !ok {
				continue
			}
			normalizeSchemaNode(child)
			properties[key] = child
		}
	}

	if items, ok := node["items"].(map[string]any); ok {
		normalizeSchemaNode(items)
		node["items"] = items
	}

	for _, key := range []string{"anyOf", "oneOf", "allOf"} {
		values, ok := node[key].([]any)
		if !ok {
			continue
		}
		for index, value := range values {
			child, ok := value.(map[string]any)
			if !ok {
				continue
			}
			normalizeSchemaNode(child)
			values[index] = child
		}
		node[key] = values
	}
}

func normalizeSchemaType(value any) any {
	switch typed := value.(type) {
	case string:
		return strings.ToLower(typed)
	case []any:
		for index, item := range typed {
			if text, ok := item.(string); ok {
				typed[index] = strings.ToLower(text)
			}
		}
		return typed
	default:
		return value
	}
}

func isObjectSchema(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.EqualFold(typed, "object")
	case []any:
		for _, item := range typed {
			text, ok := item.(string)
			if ok && strings.EqualFold(text, "object") {
				return true
			}
		}
	}

	return false
}
