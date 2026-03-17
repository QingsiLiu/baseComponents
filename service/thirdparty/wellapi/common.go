package wellapi

import (
	"os"
	"strings"
	"time"
)

const (
	BaseURL            = "https://wellapi.ai"
	DefaultTimeout     = 30 * time.Second
	DefaultRetryMax    = 3
	DefaultRetryDelay  = 200 * time.Millisecond
	APIKeyEnvVar       = "WELLAPI_API_KEY"
	BaseURLEnvVar      = "WELLAPI_BASE_URL"
	PathModels         = "/v1/models"
	PathGenerateFormat = "/v1beta/models/%s:generateContent"
	PathStreamFormat   = "/v1beta/models/%s:streamGenerateContent"
)

const (
	ModelGemini3FlashPreview         = "gemini-3-flash-preview"
	ModelGemini31FlashPreview        = "gemini-3.1-flash-preview"
	ModelGemini31FlashLitePreview    = "gemini-3.1-flash-lite-preview"
	ModelGemini3FlashPreviewThinking = "gemini-3-flash-preview-thinking"
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
