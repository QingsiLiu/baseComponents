package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// LLMService 通用大模型服务接口
type LLMService interface {
	Source() string
	Generate(req *GenerateReq) (*GenerateResp, error)
}

// GenerateReq 通用生成请求
type GenerateReq struct {
	Model               string         `json:"model"`
	SystemInstruction   string         `json:"system_instruction"`
	Messages            []Message      `json:"messages"`
	Tools               []ToolSpec     `json:"tools"`
	EnableURLContext    bool           `json:"enable_url_context"`
	EnableGoogleSearch  bool           `json:"enable_google_search"`
	EnableCodeExecution bool           `json:"enable_code_execution"`
	ResponseMIMEType    string         `json:"response_mime_type"`
	ResponseSchema      map[string]any `json:"response_schema"`
	Temperature         *float64       `json:"temperature,omitempty"`
	TopP                *float64       `json:"top_p,omitempty"`
	MaxOutputTokens     int            `json:"max_output_tokens"`
	ThinkingBudget      *int           `json:"thinking_budget,omitempty"`
	Debug               bool           `json:"debug"`
}

// ToolSpec 通用函数调用工具定义
type ToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Message 通用消息结构
type Message struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part 通用消息片段
type Part struct {
	Text             string `json:"text,omitempty"`
	MimeType         string `json:"mime_type,omitempty"`
	InlineDataBase64 string `json:"inline_data_base64,omitempty"`
}

// GenerateResp 通用生成响应
type GenerateResp struct {
	ModelVersion  string         `json:"model_version"`
	Text          string         `json:"text"`
	Parts         []RespPart     `json:"parts"`
	FunctionCalls []FunctionCall `json:"function_calls"`
	FinishReason  string         `json:"finish_reason"`
	Usage         Usage          `json:"usage"`
	Raw           []byte         `json:"raw"`
}

// RespPart 响应片段
type RespPart struct {
	Text         string        `json:"text,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// FunctionCall 函数调用信息
type FunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
	ID   string         `json:"id,omitempty"`
}

// Usage 统一 token 统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	ThoughtsTokens   int `json:"thoughts_tokens"`
}

// DecodeJSON 将响应文本解码为 JSON 结构
func (r *GenerateResp) DecodeJSON(v any) error {
	if r == nil {
		return fmt.Errorf("response is nil")
	}

	payload := strings.TrimSpace(r.Text)
	if payload == "" {
		return fmt.Errorf("response text is empty")
	}

	if strings.HasPrefix(payload, "```") {
		payload = strings.TrimPrefix(payload, "```json")
		payload = strings.TrimPrefix(payload, "```JSON")
		payload = strings.TrimPrefix(payload, "```")
		payload = strings.TrimSuffix(payload, "```")
		payload = strings.TrimSpace(payload)
	}

	if err := json.Unmarshal([]byte(payload), v); err != nil {
		return fmt.Errorf("decode response json: %w", err)
	}

	return nil
}

// HasFunctionCalls 是否包含函数调用
func (r *GenerateResp) HasFunctionCalls() bool {
	return r != nil && len(r.FunctionCalls) > 0
}

// FirstFunctionCall 返回第一个函数调用
func (r *GenerateResp) FirstFunctionCall() *FunctionCall {
	if !r.HasFunctionCalls() {
		return nil
	}
	return &r.FunctionCalls[0]
}
