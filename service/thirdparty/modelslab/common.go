package modelslab

import (
	"os"
	"strings"
	"time"
)

// API基础配置
const (
	BaseURL          = "https://modelslab.com/api/v6"
	PathText2Img     = "/images/text2img"
	PathInterior     = "/interior/make"
	PathExterior     = "/interior/exterior_restorer"
	PathFetch        = "/images/fetch"
	Text2ImgEndpoint = BaseURL + "/images/text2img"
	InteriorEndpoint = BaseURL + "/interior/make"
	ExteriorEndpoint = BaseURL + "/interior/exterior_restorer"
	FetchEndpoint    = BaseURL + "/images/fetch"
	DefaultTimeout   = 30 * time.Second
	APIKeyEnvVar     = "MODELSLAB_API_KEY"
	BaseURLEnvVar    = "MODELSLAB_BASE_URL"
)

func GetAPIKey() string {
	return os.Getenv(APIKeyEnvVar)
}

func GetBaseURL() string {
	baseURL := strings.TrimSpace(os.Getenv(BaseURLEnvVar))
	if baseURL == "" {
		return BaseURL
	}
	return strings.TrimRight(baseURL, "/")
}
