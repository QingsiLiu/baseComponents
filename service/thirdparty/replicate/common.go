package replicate

import (
	"os"
	"time"
)

// API基础配置
const (
	BaseURL          = "https://api.replicate.com/v1"
	DefaultTimeout   = 30 * time.Second
	APITokenEnvVar   = "REPLICATE_TOKEN"
)

// API路径常量
const (
	PathPredictions       = "/predictions"
	PathPredictionGet     = "/predictions/%s"
	PathPredictionCancel  = "/predictions/%s/cancel"
	PathModels           = "/models"
	PathModelGet         = "/models/%s"
)

// 任务状态常量
const (
	StatusStarting   = "starting"
	StatusProcessing = "processing"
	StatusSucceeded  = "succeeded"
	StatusFailed     = "failed"
	StatusCanceled   = "canceled"
)

// 获取API Token
func GetAPIToken() string {
	return os.Getenv(APITokenEnvVar)
}

// 状态转换函数 - 将Replicate状态转换为通用状态码
func ConvertStatusToInt(status string) int32 {
	switch status {
	case StatusSucceeded:
		return 2 // TaskStatusCompleted
	case StatusProcessing, StatusStarting:
		return 1 // TaskStatusRunning
	case StatusFailed:
		return 4 // TaskStatusFailed
	case StatusCanceled:
		return 3 // TaskStatusCanceled
	default:
		return 0 // TaskStatusPending
	}
}

// 检查状态是否为最终状态
func IsFinalStatus(status string) bool {
	return status == StatusSucceeded || status == StatusFailed || status == StatusCanceled
}