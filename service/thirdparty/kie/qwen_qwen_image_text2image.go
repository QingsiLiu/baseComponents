package kie

import (
	"fmt"
	"strings"
	"time"

	"github.com/QingsiLiu/baseComponents/service/text2image"
	"github.com/QingsiLiu/baseComponents/utils"
)

const (
	qwenText2ImageModelName = "qwen/text-to-image"
)

// QwenText2ImageService KIE Qwen Text-to-Image 服务实现
type QwenText2ImageService struct {
	client *Client
}

// NewQwenText2ImageService 创建默认服务实例
func NewQwenText2ImageService() text2image.Text2ImageService {
	return &QwenText2ImageService{
		client: NewClient(),
	}
}

// NewQwenText2ImageServiceWithKey 使用指定 API Key 创建服务实例
func NewQwenText2ImageServiceWithKey(apiKey string) text2image.Text2ImageService {
	return &QwenText2ImageService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 返回服务来源标识
func (s *QwenText2ImageService) Source() string {
	return text2image.SourceKieQwenImageText2Image
}

// TaskRun 提交任务
func (s *QwenText2ImageService) TaskRun(req *text2image.Text2ImageTaskRunReq) (string, error) {
	if req.Debug {
		return "mock_task_id_" + utils.RandomString(5), nil
	}
	payload := s.convertToCreateRequest(req)

	resp, err := s.client.CreateTask(payload)
	if err != nil {
		return "", err
	}

	if resp.Data == nil || resp.Data.TaskID == "" {
		return "", fmt.Errorf("missing task ID in response")
	}

	return resp.Data.TaskID, nil
}

// TaskGet 查询任务
func (s *QwenText2ImageService) TaskGet(taskId string) (*text2image.Text2ImageTaskInfo, error) {
	if strings.HasPrefix(taskId, "mock_task_id_") {
		return &text2image.Text2ImageTaskInfo{
			TaskId:   taskId,
			Status:   text2image.TaskStatusCompleted,
			Result:   []string{"https://kie.ai/cdn-cgi/image/width=1920,quality=85,fit=scale-down,format=webp/https://file.aiquickdraw.com/custom-page/akr/section-images/1756260298615p09gs2nz.webp"},
			Duration: 5,
		}, nil
	}
	resp, err := s.client.GetTaskRecord(taskId)
	if err != nil {
		return nil, err
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("task data is empty")
	}

	return s.convertToTaskInfo(resp.Data), nil
}

// TaskCancel 取消任务（KIE 暂不支持）
func (s *QwenText2ImageService) TaskCancel(taskId string) error {
	return fmt.Errorf("task cancellation not supported by KIE Qwen Text2Image API")
}

// TaskList 列出任务（KIE 暂不支持）
func (s *QwenText2ImageService) TaskList() ([]*text2image.Text2ImageTaskInfo, error) {
	return nil, fmt.Errorf("task listing not supported by KIE Qwen Text2Image API")
}

func (s *QwenText2ImageService) convertToCreateRequest(req *text2image.Text2ImageTaskRunReq) *TaskCreateRequest {
	model := req.Model
	if model == "" {
		model = qwenText2ImageModelName
	}

	input := &QwenText2ImageInput{
		Prompt:            req.Prompt,
		ImageSize:         req.AspectRatio,
		NumInferenceSteps: req.NumInferenceSteps,
		Seed:              req.Seed,
		GuidanceScale:     req.Guidance,
		EnableSafety:      mapDisableSafetyToEnable(req.DisableSafetyChecker),
		OutputFormat:      req.OutputFormat,
		NegativePrompt:    req.NegativePrompt,
		Acceleration:      normalizeAcceleration(req.SpeedMode),
	}

	return &TaskCreateRequest{
		Model: model,
		Input: input,
	}
}

func (s *QwenText2ImageService) convertToTaskInfo(detail *TaskRecordDetail) *text2image.Text2ImageTaskInfo {
	task := &text2image.Text2ImageTaskInfo{
		TaskId: detail.TaskID,
		Status: ConvertStateToStatus(detail.State),
		Result: ParseResultURLs(detail.ResultJSON),
	}

	if detail.CreateTime > 0 {
		task.CreateTime = int32(time.UnixMilli(detail.CreateTime).Unix())
	}
	if detail.UpdateTime > 0 {
		task.UpdateTime = int32(time.UnixMilli(detail.UpdateTime).Unix())
	}

	var endTime int64
	switch {
	case detail.CompleteTime > 0:
		endTime = detail.CompleteTime
	case detail.UpdateTime > 0:
		endTime = detail.UpdateTime
	}

	if endTime > 0 && detail.CreateTime > 0 && endTime >= detail.CreateTime {
		task.Duration = (time.Duration(endTime-detail.CreateTime) * time.Millisecond).Seconds()
	}

	return task
}

func normalizeAcceleration(speedMode string) string {
	switch strings.ToLower(strings.TrimSpace(speedMode)) {
	case "none", "regular", "high":
		return strings.ToLower(strings.TrimSpace(speedMode))
	default:
		return ""
	}
}

func mapDisableSafetyToEnable(disable bool) *bool {
	if !disable {
		return nil
	}
	enable := false
	return &enable
}

// QwenText2ImageInput Qwen 文生图模型输入
type QwenText2ImageInput struct {
	Prompt            string  `json:"prompt"`
	ImageSize         string  `json:"image_size,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
	Seed              int     `json:"seed,omitempty"`
	GuidanceScale     float64 `json:"guidance_scale,omitempty"`
	EnableSafety      *bool   `json:"enable_safety_checker,omitempty"`
	OutputFormat      string  `json:"output_format,omitempty"`
	NegativePrompt    string  `json:"negative_prompt,omitempty"`
	Acceleration      string  `json:"acceleration,omitempty"`
}
