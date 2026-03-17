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

// GeminiService WellAPI Gemini 服务实现
type GeminiService struct {
	client *Client
}

// NewGeminiService 创建默认服务实例
func NewGeminiService() llm.LLMService {
	return &GeminiService{
		client: NewClient(),
	}
}

// NewGeminiServiceWithKey 使用指定 API Key 创建服务实例
func NewGeminiServiceWithKey(apiKey string) llm.LLMService {
	return &GeminiService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 返回服务来源标识
func (s *GeminiService) Source() string {
	return llm.SourceWellAPIGemini
}

// Generate 执行单次生成请求
func (s *GeminiService) Generate(req *llm.GenerateReq) (*llm.GenerateResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if len(req.Messages) == 0 && strings.TrimSpace(req.SystemInstruction) == "" {
		return nil, fmt.Errorf("at least one message or system instruction is required")
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

func (s *GeminiService) resolveModels(req *llm.GenerateReq) ([]string, bool) {
	model := strings.TrimSpace(req.Model)
	if model != "" {
		return []string{model}, true
	}

	thinkingBudget := 0
	if req.ThinkingBudget != nil {
		thinkingBudget = *req.ThinkingBudget
	}

	if thinkingBudget > 0 {
		return []string{ModelGemini3FlashPreviewThinking}, false
	}

	return []string{
		ModelGemini3FlashPreview,
		ModelGemini31FlashLitePreview,
	}, false
}

func (s *GeminiService) generateWithRetry(req *llm.GenerateReq, model string) (*llm.GenerateResp, error) {
	attempts := s.client.retryMax
	if attempts <= 0 {
		attempts = DefaultRetryMax
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		wireReq := s.buildGenerateContentRequest(req, model)
		wireResp, err := s.client.GenerateContent(wireReq)
		if err == nil {
			return s.convertGenerateContentResponse(wireResp)
		}

		lastErr = err
		if attempt == attempts || !isRetryableError(err) {
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

func (s *GeminiService) buildGenerateContentRequest(req *llm.GenerateReq, model string) *GenerateContentRequest {
	wireReq := &GenerateContentRequest{
		Model:    model,
		Contents: make([]Content, 0, len(req.Messages)),
	}

	if strings.TrimSpace(req.SystemInstruction) != "" {
		wireReq.SystemInstruction = &Content{
			Parts: []Part{{Text: req.SystemInstruction}},
		}
	}

	for _, message := range req.Messages {
		wireReq.Contents = append(wireReq.Contents, convertMessage(message))
	}

	wireReq.Tools = convertTools(req)
	wireReq.GenerationConfig = convertGenerationConfig(req)

	return wireReq
}

func (s *GeminiService) convertGenerateContentResponse(resp *GenerateContentResponse) (*llm.GenerateResp, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("response candidates are empty")
	}

	raw := resp.Raw
	if len(raw) == 0 {
		marshaled, err := json.Marshal(resp)
		if err == nil {
			raw = marshaled
		}
	}

	candidate := resp.Candidates[0]
	result := &llm.GenerateResp{
		ModelVersion: resp.ModelVersion,
		FinishReason: candidate.FinishReason,
		Raw:          raw,
	}

	if resp.UsageMetadata != nil {
		result.Usage = llm.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount + resp.UsageMetadata.ToolUsePromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
			ThoughtsTokens:   resp.UsageMetadata.ThoughtsTokenCount,
		}
	}

	var builder strings.Builder
	for _, part := range candidate.Content.Parts {
		switch {
		case part.Text != "":
			builder.WriteString(part.Text)
			result.Parts = append(result.Parts, llm.RespPart{Text: part.Text})
		case part.FunctionCall != nil:
			call := llm.FunctionCall{
				Name: part.FunctionCall.Name,
				Args: part.FunctionCall.Args,
				ID:   part.FunctionCall.ID,
			}
			result.FunctionCalls = append(result.FunctionCalls, call)
			result.Parts = append(result.Parts, llm.RespPart{FunctionCall: &call})
		}
	}

	result.Text = builder.String()
	return result, nil
}

func convertMessage(message llm.Message) Content {
	content := Content{
		Role:  message.Role,
		Parts: make([]Part, 0, len(message.Parts)),
	}

	for _, part := range message.Parts {
		wirePart := Part{}
		if strings.TrimSpace(part.Text) != "" {
			wirePart.Text = part.Text
		}
		if strings.TrimSpace(part.InlineDataBase64) != "" {
			wirePart.InlineData = &InlineData{
				MimeType: part.MimeType,
				Data:     part.InlineDataBase64,
			}
		}
		if wirePart.Text == "" && wirePart.InlineData == nil {
			continue
		}
		content.Parts = append(content.Parts, wirePart)
	}

	return content
}

func convertTools(req *llm.GenerateReq) []Tool {
	tools := make([]Tool, 0, 4)

	if len(req.Tools) > 0 {
		declarations := make([]FunctionDeclaration, 0, len(req.Tools))
		for _, tool := range req.Tools {
			declarations = append(declarations, FunctionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			})
		}
		tools = append(tools, Tool{
			FunctionDeclarations: declarations,
		})
	}

	if req.EnableURLContext {
		tools = append(tools, Tool{URLContext: &struct{}{}})
	}
	if req.EnableGoogleSearch {
		tools = append(tools, Tool{GoogleSearch: &struct{}{}})
	}
	if req.EnableCodeExecution {
		tools = append(tools, Tool{CodeExecution: &struct{}{}})
	}

	if len(tools) == 0 {
		return nil
	}

	return tools
}

func convertGenerationConfig(req *llm.GenerateReq) *GenerationConfig {
	responseMIMEType := strings.TrimSpace(req.ResponseMIMEType)
	if responseMIMEType == "" && len(req.ResponseSchema) > 0 {
		responseMIMEType = "application/json"
	}

	thinkingBudget := 0
	if req.ThinkingBudget != nil {
		thinkingBudget = *req.ThinkingBudget
	}

	config := &GenerationConfig{
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		MaxOutputTokens:  req.MaxOutputTokens,
		ResponseMIMEType: responseMIMEType,
		ResponseSchema:   req.ResponseSchema,
		ThinkingConfig: &ThinkingConfig{
			ThinkingBudget: thinkingBudget,
		},
	}

	if config.Temperature == nil && config.TopP == nil && config.MaxOutputTokens == 0 && config.ResponseMIMEType == "" && len(config.ResponseSchema) == 0 {
		return &GenerationConfig{
			ThinkingConfig: &ThinkingConfig{ThinkingBudget: thinkingBudget},
		}
	}

	return config
}

func isRetryableError(err error) bool {
	var apiErr *APIError
	if stderrors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}

	var netErr net.Error
	if stderrors.As(err, &netErr) {
		return true
	}

	return stderrors.Is(err, context.DeadlineExceeded)
}
