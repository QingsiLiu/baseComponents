package wellapi

// ListModelsResponse WellAPI 模型列表响应
type ListModelsResponse struct {
	Data    []Model `json:"data"`
	Object  string  `json:"object"`
	Success bool    `json:"success"`
}

// Model 单个模型信息
type Model struct {
	ID                     string   `json:"id"`
	Object                 string   `json:"object"`
	Created                int64    `json:"created"`
	OwnedBy                string   `json:"owned_by"`
	SupportedEndpointTypes []string `json:"supported_endpoint_types"`
}

// GenerateContentRequest Gemini 原生 generateContent 请求
type GenerateContentRequest struct {
	Model             string            `json:"-"`
	SystemInstruction *Content          `json:"systemInstruction,omitempty"`
	Contents          []Content         `json:"contents"`
	Tools             []Tool            `json:"tools,omitempty"`
	GenerationConfig  *GenerationConfig `json:"generationConfig,omitempty"`
}

// Content Gemini 内容结构
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts,omitempty"`
}

// Part Gemini 内容片段
type Part struct {
	Text                string               `json:"text,omitempty"`
	InlineData          *InlineData          `json:"inline_data,omitempty"`
	FunctionCall        *WireFunctionCall    `json:"functionCall,omitempty"`
	ExecutableCode      *ExecutableCode      `json:"executableCode,omitempty"`
	CodeExecutionResult *CodeExecutionResult `json:"codeExecutionResult,omitempty"`
	ThoughtSignature    string               `json:"thoughtSignature,omitempty"`
}

// InlineData 内联媒体数据
type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// GenerationConfig 生成参数
type GenerationConfig struct {
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"topP,omitempty"`
	MaxOutputTokens  int             `json:"maxOutputTokens,omitempty"`
	ResponseMIMEType string          `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]any  `json:"responseSchema,omitempty"`
	ThinkingConfig   *ThinkingConfig `json:"thinkingConfig,omitempty"`
}

// ThinkingConfig 思考配置
type ThinkingConfig struct {
	ThinkingBudget int `json:"thinkingBudget"`
}

// Tool Gemini 工具定义
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"`
	URLContext           *struct{}             `json:"UrlContext,omitempty"`
	GoogleSearch         *struct{}             `json:"googleSearch,omitempty"`
	CodeExecution        *struct{}             `json:"codeExecution,omitempty"`
}

// FunctionDeclaration 函数声明
type FunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// WireFunctionCall Gemini 原生函数调用
type WireFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
	ID   string         `json:"id,omitempty"`
}

// ExecutableCode 代码执行片段
type ExecutableCode struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

// CodeExecutionResult 代码执行结果
type CodeExecutionResult struct {
	Outcome string `json:"outcome"`
	Output  string `json:"output,omitempty"`
}

// GenerateContentResponse Gemini 原生响应
type GenerateContentResponse struct {
	Candidates    []Candidate    `json:"candidates"`
	CreateTime    string         `json:"createTime,omitempty"`
	ModelVersion  string         `json:"modelVersion,omitempty"`
	ResponseID    string         `json:"responseId,omitempty"`
	UsageMetadata *UsageMetadata `json:"usageMetadata,omitempty"`
	Raw           []byte         `json:"-"`
}

// Candidate 候选结果
type Candidate struct {
	Index         int            `json:"index,omitempty"`
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason,omitempty"`
	FinishMessage string         `json:"finishMessage,omitempty"`
	AvgLogprobs   float64        `json:"avgLogprobs,omitempty"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// UsageMetadata Gemini token 使用信息
type UsageMetadata struct {
	PromptTokenCount        int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount    int `json:"candidatesTokenCount,omitempty"`
	TotalTokenCount         int `json:"totalTokenCount,omitempty"`
	ThoughtsTokenCount      int `json:"thoughtsTokenCount,omitempty"`
	ToolUsePromptTokenCount int `json:"toolUsePromptTokenCount,omitempty"`
}

// SafetyRating 安全分级
type SafetyRating struct {
	Category         string  `json:"category,omitempty"`
	Probability      string  `json:"probability,omitempty"`
	ProbabilityScore float64 `json:"probabilityScore,omitempty"`
	Severity         string  `json:"severity,omitempty"`
	SeverityScore    float64 `json:"severityScore,omitempty"`
}

type apiErrorEnvelope struct {
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   any    `json:"param"`
		Code    any    `json:"code"`
	} `json:"error"`
}

// APIError WellAPI 错误响应
type APIError struct {
	StatusCode int
	Message    string
	Type       string
	Param      string
	Code       string
	Raw        string
}

func (e *APIError) Error() string {
	if e == nil {
		return "wellapi api error"
	}

	if e.Type != "" {
		return e.Type + ": " + e.Message
	}

	if e.Message != "" {
		return e.Message
	}

	return "wellapi api error"
}

// IsRetryable 判断是否可重试
func (e *APIError) IsRetryable() bool {
	if e == nil {
		return false
	}

	if e.StatusCode == 429 || e.StatusCode >= 500 {
		return true
	}

	return e.Type == "upstream_error"
}
