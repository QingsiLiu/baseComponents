package kie

import (
	"fmt"
	"strings"
	"time"

	"github.com/QingsiLiu/baseComponents/service/image2image"
	"github.com/QingsiLiu/baseComponents/utils"
)

const (
	nanoBananaModelName     = "google/nano-banana"
	nanoBananaEditModelName = "google/nano-banana-edit"
	nanoBananaProModelName  = "nano-banana-pro"
)

// NanoBananaService KIE Nano Banana 服务实现
type NanoBananaService struct {
	client *Client
}

// NewNanoBananaService 创建默认服务实例
func NewNanoBananaService() image2image.Image2ImageService {
	return &NanoBananaService{
		client: NewClient(),
	}
}

// NewNanoBananaServiceWithKey 使用指定 API Key 创建服务实例
func NewNanoBananaServiceWithKey(apiKey string) image2image.Image2ImageService {
	return &NanoBananaService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 返回服务来源标识
func (s *NanoBananaService) Source() string {
	return image2image.SourceKieNanoBanana
}

// TaskRun 提交任务
func (s *NanoBananaService) TaskRun(req *image2image.Image2ImageTaskRunReq) (string, error) {
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
func (s *NanoBananaService) TaskGet(taskId string) (*image2image.Image2ImageTaskInfo, error) {
	if strings.HasPrefix(taskId, "mock_task_id_") {
		return &image2image.Image2ImageTaskInfo{
			TaskId:   taskId,
			Status:   image2image.TaskStatusCompleted,
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
func (s *NanoBananaService) TaskCancel(taskId string) error {
	return fmt.Errorf("task cancellation not supported by KIE NanoBanana API")
}

// TaskList 列出任务（KIE 暂不支持）
func (s *NanoBananaService) TaskList() ([]*image2image.Image2ImageTaskInfo, error) {
	return nil, fmt.Errorf("task listing not supported by KIE NanoBanana API")
}

func (s *NanoBananaService) convertToCreateRequest(req *image2image.Image2ImageTaskRunReq) *TaskCreateRequest {
	model := req.Model
	switch model {
	case "":
		if len(req.ImageInputs) > 0 {
			model = nanoBananaEditModelName
		} else {
			model = nanoBananaModelName
		}

		input := &NanoBananaInput{
			Prompt:       req.Prompt,
			OutputFormat: req.OutputFormat,
			ImageURLs:    req.ImageInputs,
			ImageSize:    req.OutputImageSize,
		}

		if input.OutputFormat == "" {
			input.OutputFormat = "png"
		}

		return &TaskCreateRequest{
			Model: model,
			Input: input,
		}
	case nanoBananaProModelName:
		model = nanoBananaProModelName
		input := &NanoBananaProInput{
			Prompt:       req.Prompt,
			ImageInput:   req.ImageInputs,
			AspectRatio:  req.OutputImageSize,
			Resolution:   req.Resolution,
			OutputFormat: req.OutputFormat,
		}
		if input.Resolution == "" {
			input.Resolution = "1K"
		}
		if input.OutputFormat == "" {
			input.OutputFormat = "png"
		}
		return &TaskCreateRequest{
			Model: model,
			Input: input,
		}
	}
	return nil
}

func (s *NanoBananaService) convertToTaskInfo(detail *TaskRecordDetail) *image2image.Image2ImageTaskInfo {
	task := &image2image.Image2ImageTaskInfo{
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
