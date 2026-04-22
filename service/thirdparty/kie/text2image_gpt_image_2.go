package kie

import (
	"fmt"
	"strings"

	"github.com/QingsiLiu/baseComponents/service/text2image"
	"github.com/QingsiLiu/baseComponents/utils"
)

const (
	gptImage2Text2ImageModelName = "gpt-image-2-text-to-image"
)

// GPTImage2Text2ImageService KIE GPT Image-2 文生图服务实现。
type GPTImage2Text2ImageService struct {
	client *Client
}

// NewGPTImage2Text2ImageService 创建默认服务实例。
func NewGPTImage2Text2ImageService() text2image.Text2ImageService {
	return &GPTImage2Text2ImageService{
		client: NewClient(),
	}
}

// NewGPTImage2Text2ImageServiceWithKey 使用指定 API Key 创建服务实例。
func NewGPTImage2Text2ImageServiceWithKey(apiKey string) text2image.Text2ImageService {
	return &GPTImage2Text2ImageService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 返回服务来源标识。
func (s *GPTImage2Text2ImageService) Source() string {
	return text2image.SourceKieGPTImage2Text2Image
}

// TaskRun 提交任务。
func (s *GPTImage2Text2ImageService) TaskRun(req *text2image.Text2ImageTaskRunReq) (string, error) {
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

// TaskGet 查询任务。
func (s *GPTImage2Text2ImageService) TaskGet(taskID string) (*text2image.Text2ImageTaskInfo, error) {
	if strings.HasPrefix(taskID, "mock_task_id_") {
		return &text2image.Text2ImageTaskInfo{
			TaskId:   taskID,
			Status:   text2image.TaskStatusCompleted,
			Result:   []string{"https://kie.ai/cdn-cgi/image/width=1920,quality=85,fit=scale-down,format=webp/https://file.aiquickdraw.com/custom-page/akr/section-images/1756260298615p09gs2nz.webp"},
			Duration: 5,
		}, nil
	}

	resp, err := s.client.GetTaskRecord(taskID)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("task data is empty")
	}

	return s.convertToTaskInfo(resp.Data), nil
}

// TaskCancel 取消任务（KIE 暂不支持）。
func (s *GPTImage2Text2ImageService) TaskCancel(taskID string) error {
	return fmt.Errorf("task cancellation not supported by KIE GPT Image-2 Text2Image API")
}

// TaskList 列出任务（KIE 暂不支持）。
func (s *GPTImage2Text2ImageService) TaskList() ([]*text2image.Text2ImageTaskInfo, error) {
	return nil, fmt.Errorf("task listing not supported by KIE GPT Image-2 Text2Image API")
}

func (s *GPTImage2Text2ImageService) convertToCreateRequest(req *text2image.Text2ImageTaskRunReq) *TaskCreateRequest {
	model := req.Model
	if model == "" {
		model = gptImage2Text2ImageModelName
	}

	return &TaskCreateRequest{
		Model: model,
		Input: &GPTImage2Text2ImageInput{
			Prompt:      req.Prompt,
			NSFWChecker: !req.DisableSafetyChecker,
		},
	}
}

func (s *GPTImage2Text2ImageService) convertToTaskInfo(detail *TaskRecordDetail) *text2image.Text2ImageTaskInfo {
	return &text2image.Text2ImageTaskInfo{
		TaskId:     detail.TaskID,
		Status:     ConvertStateToStatus(detail.State),
		Result:     ParseResultURLs(detail.ResultJSON),
		CreateTime: UnixMillisToSeconds(detail.CreateTime),
		UpdateTime: ResolveTaskUpdateTime(detail),
		Duration:   ResolveTaskDuration(detail),
	}
}

// GPTImage2Text2ImageInput GPT Image-2 文生图模型输入。
type GPTImage2Text2ImageInput struct {
	Prompt      string `json:"prompt"`
	NSFWChecker bool   `json:"nsfw_checker"`
}
