package kie

import (
	"encoding/json"

	"github.com/QingsiLiu/baseComponents/service/image2image"
)

// 任务状态常量
const (
	TaskStateWaiting    = "waiting"
	TaskStateQueuing    = "queuing"
	TaskStateGenerating = "generating"
	TaskStateSuccess    = "success"
	TaskStateFail       = "fail"
)

// TaskCreateRequest 创建任务请求
type TaskCreateRequest struct {
	Model       string      `json:"model"`
	CallbackURL string      `json:"callBackUrl,omitempty"`
	Input       interface{} `json:"input"`
}

// NanoBananaInput Nano Banana 模型输入
type NanoBananaInput struct {
	Prompt       string   `json:"prompt"`
	OutputFormat string   `json:"output_format,omitempty"`
	ImageSize    string   `json:"image_size,omitempty"`
	ImageURLs    []string `json:"image_urls,omitempty"`
}

type NanoBananaProInput struct {
	Prompt       string   `json:"prompt"`
	ImageInput   []string `json:"image_input,omitempty"`
	AspectRatio  string   `json:"aspect_ratio,omitempty"`
	Resolution   string   `json:"resolution,omitempty"`
	OutputFormat string   `json:"output_format,omitempty"`
}

// TaskCreateResponse 创建任务响应
type TaskCreateResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *TaskCreatePayload `json:"data,omitempty"`
}

// TaskCreatePayload 创建任务成功载荷
type TaskCreatePayload struct {
	TaskID string `json:"taskId"`
}

// TaskRecordResponse 任务记录响应
type TaskRecordResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *TaskRecordDetail `json:"data,omitempty"`
}

// TaskRecordDetail 任务详情
type TaskRecordDetail struct {
	TaskID       string `json:"taskId"`
	Model        string `json:"model"`
	State        string `json:"state"`
	Param        string `json:"param"`
	ResultJSON   string `json:"resultJson"`
	FailCode     string `json:"failCode"`
	FailMsg      string `json:"failMsg"`
	CompleteTime int64  `json:"completeTime"`
	CreateTime   int64  `json:"createTime"`
	UpdateTime   int64  `json:"updateTime"`
}

// TaskResultEnvelope 任务结果
type TaskResultEnvelope struct {
	ResultUrls []string `json:"resultUrls"`
	ResultURL  string   `json:"resultUrl"`
	URLs       []string `json:"urls"`
	URL        string   `json:"url"`
}

// ConvertStateToStatus 将任务状态转换为统一状态码
func ConvertStateToStatus(state string) int32 {
	switch state {
	case TaskStateSuccess:
		return image2image.TaskStatusCompleted
	case TaskStateGenerating:
		return image2image.TaskStatusRunning
	case TaskStateFail:
		return image2image.TaskStatusFailed
	case TaskStateWaiting, TaskStateQueuing:
		return image2image.TaskStatusPending
	default:
		return image2image.TaskStatusPending
	}
}

// ParseResultURLs 从 resultJson 字段提取结果链接
func ParseResultURLs(resultJSON string) []string {
	if resultJSON == "" {
		return nil
	}

	var envelope TaskResultEnvelope
	if err := json.Unmarshal([]byte(resultJSON), &envelope); err != nil {
		return nil
	}

	urls := make([]string, 0, len(envelope.ResultUrls)+len(envelope.URLs)+2)

	if len(envelope.ResultUrls) > 0 {
		urls = append(urls, envelope.ResultUrls...)
	}
	if envelope.ResultURL != "" {
		urls = append(urls, envelope.ResultURL)
	}
	if len(envelope.URLs) > 0 {
		urls = append(urls, envelope.URLs...)
	}
	if envelope.URL != "" {
		urls = append(urls, envelope.URL)
	}

	if len(urls) == 0 {
		return nil
	}

	// 去重
	seen := make(map[string]struct{}, len(urls))
	unique := make([]string, 0, len(urls))
	for _, u := range urls {
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		unique = append(unique, u)
	}

	if len(unique) == 0 {
		return nil
	}

	return unique
}
