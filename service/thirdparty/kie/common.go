package kie

import (
	"os"
	"time"
)

const (
	// BaseURL 是 KIE API 的基础地址
	BaseURL = "https://api.kie.ai"

	// API 路径
	CreateTaskEndpoint = "/api/v1/jobs/createTask"
	RecordInfoEndpoint = "/api/v1/jobs/recordInfo"

	// 默认配置
	DefaultTimeout = 30 * time.Second
	APIKeyEnvVar   = "KIE_API_KEY"
)

// GetAPIKey 从环境变量中获取 KIE API Key
func GetAPIKey() string {
	return os.Getenv(APIKeyEnvVar)
}
