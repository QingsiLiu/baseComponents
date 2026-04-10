package kie

import (
	"encoding/json"
	"time"
)

// 任务状态常量
const (
	TaskStateWaiting    = "waiting"
	TaskStateQueuing    = "queuing"
	TaskStateGenerating = "generating"
	TaskStateSuccess    = "success"
	TaskStateFail       = "fail"
)

const (
	TaskStatusPending int32 = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusCanceled
	TaskStatusFailed
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

type NanoBanana2Input struct {
	Prompt       string   `json:"prompt"`
	ImageInput   []string `json:"image_input,omitempty"`
	AspectRatio  string   `json:"aspect_ratio,omitempty"`
	GoogleSearch bool     `json:"google_search"`
	Resolution   string   `json:"resolution,omitempty"`
	OutputFormat string   `json:"output_format,omitempty"`
}

// TaskCreateResponse 创建任务响应
type TaskCreateResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message,omitempty"`
	Msg     string             `json:"msg,omitempty"`
	Data    *TaskCreatePayload `json:"data,omitempty"`
}

// TaskCreatePayload 创建任务成功载荷
type TaskCreatePayload struct {
	TaskID string `json:"taskId"`
}

// TaskRecordResponse 任务记录响应
type TaskRecordResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message,omitempty"`
	Msg     string            `json:"msg,omitempty"`
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
	CostTime     int64  `json:"costTime"`
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
		return TaskStatusCompleted
	case TaskStateGenerating:
		return TaskStatusRunning
	case TaskStateFail:
		return TaskStatusFailed
	case TaskStateWaiting, TaskStateQueuing:
		return TaskStatusPending
	default:
		return TaskStatusPending
	}
}

func (r *TaskCreateResponse) GetMessage() string {
	if r == nil {
		return ""
	}
	if r.Message != "" {
		return r.Message
	}
	return r.Msg
}

func (r *TaskRecordResponse) GetMessage() string {
	if r == nil {
		return ""
	}
	if r.Message != "" {
		return r.Message
	}
	return r.Msg
}

func UnixMillisToSeconds(ms int64) int32 {
	if ms <= 0 {
		return 0
	}
	return int32(time.UnixMilli(ms).Unix())
}

func ResolveTaskUpdateTime(detail *TaskRecordDetail) int32 {
	if detail == nil {
		return 0
	}
	switch {
	case detail.UpdateTime > 0:
		return UnixMillisToSeconds(detail.UpdateTime)
	case detail.CompleteTime > 0:
		return UnixMillisToSeconds(detail.CompleteTime)
	case detail.CreateTime > 0:
		return UnixMillisToSeconds(detail.CreateTime)
	default:
		return 0
	}
}

func ResolveTaskDuration(detail *TaskRecordDetail) float64 {
	if detail == nil {
		return 0
	}
	if detail.CostTime > 0 {
		return (time.Duration(detail.CostTime) * time.Millisecond).Seconds()
	}

	var endTime int64
	switch {
	case detail.CompleteTime > 0:
		endTime = detail.CompleteTime
	case detail.UpdateTime > 0:
		endTime = detail.UpdateTime
	}
	if endTime > 0 && detail.CreateTime > 0 && endTime >= detail.CreateTime {
		return (time.Duration(endTime-detail.CreateTime) * time.Millisecond).Seconds()
	}

	return 0
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
