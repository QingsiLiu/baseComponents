package wellapi

import (
	"os"
	"strings"
	"time"
)

const (
	BaseURL               = "https://wellapi.ai"
	DefaultTimeout        = 30 * time.Second
	DefaultRetryMax       = 3
	DefaultRetryDelay     = 200 * time.Millisecond
	APIKeyEnvVar          = "WELLAPI_API_KEY"
	BaseURLEnvVar         = "WELLAPI_BASE_URL"
	PathModels            = "/v1/models"
	PathChatCompletions   = "/v1/chat/completions"
	PathImagesGenerations = "/v1/images/generations"
	PathResponses         = "/v1/responses"
	PathGenerateFormat    = "/v1beta/models/%s:generateContent"
	PathStreamFormat      = "/v1beta/models/%s:streamGenerateContent"
)

const (
	ModelGemini25Flash               = "gemini-2.5-flash"
	ModelGemini25Pro                 = "gemini-2.5-pro"
	ModelGemini3FlashPreview         = "gemini-3-flash-preview"
	ModelGemini3FlashPreviewThinking = "gemini-3-flash-preview-thinking"
	ModelGemini3ProPreview           = "gemini-3-pro-preview"
	ModelGemini31FlashPreview        = "gemini-3.1-flash-preview"
	ModelGemini31FlashLitePreview    = "gemini-3.1-flash-lite-preview"
	ModelGemini31ProPreview          = "gemini-3.1-pro-preview"
)

const (
	ModelGPT54Mini         = "gpt-5.4-mini"
	ModelGPT54Mini20260317 = "gpt-5.4-mini-2026-03-17"
	ModelGPT54Nano         = "gpt-5.4-nano"
	ModelGPT54Nano20260317 = "gpt-5.4-nano-2026-03-17"
	ModelGPT54             = "gpt-5.4"
	ModelGPT5Mini20250807  = "gpt-5-mini-2025-08-07"
	ModelGPT5Nano20250807  = "gpt-5-nano-2025-08-07"
	ModelGPTImage2         = "gpt-image-2"
)

// Config WellAPI 客户端配置
type Config struct {
	APIKey         string
	BaseURL        string
	Timeout        time.Duration
	RetryMax       int
	RetryBaseDelay time.Duration
}

// GetAPIKey 从环境变量中获取 API Key
func GetAPIKey() string {
	return os.Getenv(APIKeyEnvVar)
}

// GetBaseURL 从环境变量中获取 BaseURL
func GetBaseURL() string {
	baseURL := strings.TrimSpace(os.Getenv(BaseURLEnvVar))
	if baseURL == "" {
		return BaseURL
	}
	return strings.TrimRight(baseURL, "/")
}
