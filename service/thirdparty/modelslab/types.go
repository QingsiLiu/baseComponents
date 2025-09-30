package modelslab

import "strconv"

// 任务状态常量
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusSuccess    = "success"
	TaskStatusFailed     = "failed"
)

// 通用请求和响应类型

// TaskGetRequest 任务查询请求
type TaskGetRequest struct {
	Key       string `json:"key"`
	RequestID string `json:"request_id"`	
}

// TaskRunResponse 任务提交响应
type TaskRunResponse struct {
	Status              string   `json:"status"`
	GenerationTime      float64  `json:"generationTime"`
	ID                  int      `json:"id"`
	Output              []string `json:"output"`
	ProxyLinks          []string `json:"proxy_links"`
	NSFWContentDetected string   `json:"nsfw_content_detected"`
	WebhookStatus       string   `json:"webhook_status"`
	Tip                 string   `json:"tip"`
}

// TaskGetResponse 任务查询响应
type TaskGetResponse struct {
	Status     string      `json:"status"`
	ID         int         `json:"id"`
	Output     interface{} `json:"output"` // 可能是[]string或空字符串
	ProxyLinks []string    `json:"proxy_links"`
	Message    string      `json:"message"` // 错误信息
	Tip        string      `json:"tip"`
}

// TaskInfo 通用任务信息
type TaskInfo struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Output    interface{} `json:"output"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

// ConvertToTaskInfo 将API响应转换为通用任务信息
func ConvertToTaskInfo(resp *TaskGetResponse) *TaskInfo {
	return &TaskInfo{
		ID:      strconv.Itoa(resp.ID),
		Status:  resp.Status,
		Message: resp.Message,
		Output:  resp.Output,
	}
}

// ConvertStatusToInt 将字符串状态转换为整数状态（兼容现有接口）
func ConvertStatusToInt(status string) int32 {
	switch status {
	case TaskStatusSuccess:
		return 2 // TaskStatusCompleted
	case TaskStatusProcessing:
		return 1 // TaskStatusRunning
	case TaskStatusFailed:
		return 4 // TaskStatusFailed
	default:
		return 0 // TaskStatusPending
	}
}

// ConvertOutputToStringSlice 将interface{}类型的output转换为字符串切片
func ConvertOutputToStringSlice(output interface{}) []string {
	if output == nil {
		return nil
	}

	switch v := output.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				result[i] = str
			}
		}
		return result
	case string:
		if v != "" {
			return []string{v}
		}
		return nil
	default:
		return nil
	}
}
