package kie

import (
	"fmt"
	"strings"

	"github.com/QingsiLiu/baseComponents/service/text2image"
	"github.com/QingsiLiu/baseComponents/utils"
)

const (
	ideogramV3Text2ImageModelName = "ideogram/v3-text-to-image"
)

// IdeogramV3Text2ImageService KIE Ideogram V3 Text-to-Image 服务实现
type IdeogramV3Text2ImageService struct {
	client *Client
}

// NewIdeogramV3Text2ImageService 创建默认服务实例
func NewIdeogramV3Text2ImageService() text2image.Text2ImageService {
	return &IdeogramV3Text2ImageService{
		client: NewClient(),
	}
}

// NewIdeogramV3Text2ImageServiceWithKey 使用指定 API Key 创建服务实例
func NewIdeogramV3Text2ImageServiceWithKey(apiKey string) text2image.Text2ImageService {
	return &IdeogramV3Text2ImageService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 返回服务来源标识
func (s *IdeogramV3Text2ImageService) Source() string {
	return text2image.SourceKieIdeogramV3Text2Image
}

// TaskRun 提交任务
func (s *IdeogramV3Text2ImageService) TaskRun(req *text2image.Text2ImageTaskRunReq) (string, error) {
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
func (s *IdeogramV3Text2ImageService) TaskGet(taskId string) (*text2image.Text2ImageTaskInfo, error) {
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
func (s *IdeogramV3Text2ImageService) TaskCancel(taskId string) error {
	return fmt.Errorf("task cancellation not supported by KIE Ideogram V3 Text2Image API")
}

// TaskList 列出任务（KIE 暂不支持）
func (s *IdeogramV3Text2ImageService) TaskList() ([]*text2image.Text2ImageTaskInfo, error) {
	return nil, fmt.Errorf("task listing not supported by KIE Ideogram V3 Text2Image API")
}

func (s *IdeogramV3Text2ImageService) convertToCreateRequest(req *text2image.Text2ImageTaskRunReq) *TaskCreateRequest {
	model := req.Model
	if model == "" {
		model = ideogramV3Text2ImageModelName
	}

	input := &IdeogramV3Text2ImageInput{
		Prompt:         req.Prompt,
		RenderingSpeed: normalizeRenderingSpeed(req.SpeedMode),
		Style:          req.Style,
		ImageSize:      req.AspectRatio,
		Seed:           req.Seed,
		NegativePrompt: req.NegativePrompt,
	}

	return &TaskCreateRequest{
		Model: model,
		Input: input,
	}
}

func (s *IdeogramV3Text2ImageService) convertToTaskInfo(detail *TaskRecordDetail) *text2image.Text2ImageTaskInfo {
	task := &text2image.Text2ImageTaskInfo{
		TaskId:     detail.TaskID,
		Status:     ConvertStateToStatus(detail.State),
		Result:     ParseResultURLs(detail.ResultJSON),
		CreateTime: UnixMillisToSeconds(detail.CreateTime),
		UpdateTime: ResolveTaskUpdateTime(detail),
		Duration:   ResolveTaskDuration(detail),
	}

	return task
}

func normalizeRenderingSpeed(speedMode string) string {
	switch strings.ToUpper(strings.TrimSpace(speedMode)) {
	case "TURBO", "BALANCED", "QUALITY":
		return strings.ToUpper(strings.TrimSpace(speedMode))
	default:
		return ""
	}
}

// IdeogramV3Text2ImageInput Ideogram V3 文生图模型输入
type IdeogramV3Text2ImageInput struct {
	Prompt         string `json:"prompt"`
	RenderingSpeed string `json:"rendering_speed,omitempty"`
	Style          string `json:"style,omitempty"`
	ExpandPrompt   bool   `json:"expand_prompt,omitempty"`
	ImageSize      string `json:"image_size,omitempty"`
	Seed           int    `json:"seed,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
}
