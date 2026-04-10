package wellapi

import "encoding/json"

// ChatCompletionRequest WellAPI OpenAI Chat Completions 请求
type ChatCompletionRequest struct {
	Model          string                    `json:"model"`
	Messages       []ChatCompletionMessage   `json:"messages"`
	Temperature    *float64                  `json:"temperature,omitempty"`
	TopP           *float64                  `json:"top_p,omitempty"`
	MaxTokens      int                       `json:"max_tokens,omitempty"`
	Tools          []ChatCompletionTool      `json:"tools,omitempty"`
	ToolChoice     any                       `json:"tool_choice,omitempty"`
	ResponseFormat *ChatCompletionRespFormat `json:"response_format,omitempty"`
}

// ChatCompletionMessage Chat 请求消息
type ChatCompletionMessage struct {
	Role    string                      `json:"role"`
	Content []ChatCompletionContentPart `json:"content"`
}

// ChatCompletionContentPart Chat 消息片段
type ChatCompletionContentPart struct {
	Type     string                  `json:"type"`
	Text     string                  `json:"text,omitempty"`
	ImageURL *ChatCompletionImageURL `json:"image_url,omitempty"`
}

// ChatCompletionImageURL Chat 图片输入
type ChatCompletionImageURL struct {
	URL string `json:"url"`
}

// ChatCompletionTool Chat 工具定义
type ChatCompletionTool struct {
	Type     string                     `json:"type"`
	Function ChatCompletionToolFunction `json:"function"`
}

// ChatCompletionToolFunction Chat 函数工具
type ChatCompletionToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// ChatCompletionRespFormat Chat 结构化输出格式
type ChatCompletionRespFormat struct {
	Type       string                        `json:"type"`
	JSONSchema *ChatCompletionRespJSONSchema `json:"json_schema,omitempty"`
}

// ChatCompletionRespJSONSchema Chat JSON Schema 包装
type ChatCompletionRespJSONSchema struct {
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`
	Strict bool           `json:"strict"`
}

// ChatCompletionResponse WellAPI OpenAI Chat Completions 响应
type ChatCompletionResponse struct {
	ID      string                 `json:"id,omitempty"`
	Object  string                 `json:"object,omitempty"`
	Created int64                  `json:"created,omitempty"`
	Model   string                 `json:"model,omitempty"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   *OpenAIUsage           `json:"usage,omitempty"`
	Raw     []byte                 `json:"-"`
}

// ChatCompletionChoice Chat 选择项
type ChatCompletionChoice struct {
	Index        int                           `json:"index,omitempty"`
	Message      ChatCompletionResponseMessage `json:"message"`
	FinishReason string                        `json:"finish_reason,omitempty"`
}

// ChatCompletionResponseMessage Chat 响应消息
type ChatCompletionResponseMessage struct {
	Role      string                   `json:"role,omitempty"`
	Content   json.RawMessage          `json:"content,omitempty"`
	ToolCalls []ChatCompletionToolCall `json:"tool_calls,omitempty"`
}

// ChatCompletionToolCall Chat 工具调用
type ChatCompletionToolCall struct {
	ID       string                         `json:"id,omitempty"`
	Type     string                         `json:"type,omitempty"`
	Function ChatCompletionToolCallFunction `json:"function"`
}

// ChatCompletionToolCallFunction Chat 工具调用函数信息
type ChatCompletionToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
}

// ChatCompletionOutputPart Chat 响应内容片段
type ChatCompletionOutputPart struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// ResponsesRequest WellAPI Responses API 请求
type ResponsesRequest struct {
	Model           string               `json:"model"`
	Instructions    string               `json:"instructions,omitempty"`
	Input           []ResponsesInputItem `json:"input"`
	Tools           []ResponsesTool      `json:"tools,omitempty"`
	ToolChoice      any                  `json:"tool_choice,omitempty"`
	Text            *ResponsesTextConfig `json:"text,omitempty"`
	Temperature     *float64             `json:"temperature,omitempty"`
	TopP            *float64             `json:"top_p,omitempty"`
	MaxOutputTokens int                  `json:"max_output_tokens,omitempty"`
	Reasoning       *ResponsesReasoning  `json:"reasoning,omitempty"`
}

// ResponsesInputItem Responses 输入消息
type ResponsesInputItem struct {
	Role    string                  `json:"role"`
	Content []ResponsesInputContent `json:"content"`
}

// ResponsesInputContent Responses 输入片段
type ResponsesInputContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// ResponsesTool Responses 工具定义
type ResponsesTool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// ResponsesTextConfig Responses 文本配置
type ResponsesTextConfig struct {
	Format    *ResponsesTextFormat `json:"format,omitempty"`
	Verbosity string               `json:"verbosity,omitempty"`
}

// ResponsesTextFormat Responses 结构化输出格式
type ResponsesTextFormat struct {
	Type   string         `json:"type"`
	Name   string         `json:"name,omitempty"`
	Schema map[string]any `json:"schema,omitempty"`
	Strict bool           `json:"strict,omitempty"`
}

// ResponsesReasoning Responses reasoning 配置
type ResponsesReasoning struct {
	Effort string `json:"effort,omitempty"`
}

// ResponsesResponse WellAPI Responses API 响应
type ResponsesResponse struct {
	ID         string                `json:"id,omitempty"`
	Object     string                `json:"object,omitempty"`
	CreatedAt  int64                 `json:"created_at,omitempty"`
	Status     string                `json:"status,omitempty"`
	Model      string                `json:"model,omitempty"`
	OutputText string                `json:"output_text,omitempty"`
	Output     []ResponsesOutputItem `json:"output,omitempty"`
	Usage      *ResponsesUsage       `json:"usage,omitempty"`
	Raw        []byte                `json:"-"`
}

// ResponsesOutputItem Responses 输出项
type ResponsesOutputItem struct {
	ID        string                   `json:"id,omitempty"`
	Type      string                   `json:"type,omitempty"`
	Status    string                   `json:"status,omitempty"`
	Name      string                   `json:"name,omitempty"`
	Arguments string                   `json:"arguments,omitempty"`
	CallID    string                   `json:"call_id,omitempty"`
	Content   []ResponsesOutputContent `json:"content,omitempty"`
}

// ResponsesOutputContent Responses 输出片段
type ResponsesOutputContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// OpenAIUsage OpenAI Chat Usage
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// ResponsesUsage Responses Usage
type ResponsesUsage struct {
	PromptTokens        int                          `json:"prompt_tokens,omitempty"`
	CompletionTokens    int                          `json:"completion_tokens,omitempty"`
	InputTokens         int                          `json:"input_tokens,omitempty"`
	OutputTokens        int                          `json:"output_tokens,omitempty"`
	TotalTokens         int                          `json:"total_tokens,omitempty"`
	OutputTokensDetails *ResponsesOutputTokensDetail `json:"output_tokens_details,omitempty"`
}

// ResponsesOutputTokensDetail Responses 输出 token 细节
type ResponsesOutputTokensDetail struct {
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}
