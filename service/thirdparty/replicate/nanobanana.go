package replicate

import (
	"github.com/QingsiLiu/baseComponents/service/image2image"
	"time"
)

const (
	// NanoBanana 模型版本
	nanoBananaModelVersion = "google/nano-banana"
)

// NanoBananaService NanoBanana 服务实现
type NanoBananaService struct {
	client *Client
}

// NewNanoBananaService 创建 NanoBanana 服务实例
func NewNanoBananaService() image2image.Image2ImageService {
	return &NanoBananaService{
		client: NewClient(),
	}
}

// NewNanoBananaServiceWithKey 使用指定 API Key 创建 NanoBanana 服务实例
func NewNanoBananaServiceWithKey(apiKey string) image2image.Image2ImageService {
	return &NanoBananaService{
		client: NewClientWithToken(apiKey),
	}
}

// Source 返回服务来源标识
func (s *NanoBananaService) Source() string {
	return image2image.SourceReplicateNanoBanana
}

// TaskRun 提交 NanoBanana 任务
func (s *NanoBananaService) TaskRun(req *image2image.Image2ImageTaskRunReq) (string, error) {
	input := s.convertToNanoBananaInput(req)

	predReq := &PredictionRequest{
		Version: nanoBananaModelVersion,
		Input:   input,
	}

	resp, err := s.client.CreatePrediction(predReq)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// TaskGet 获取任务状态
func (s *NanoBananaService) TaskGet(taskId string) (*image2image.Image2ImageTaskInfo, error) {
	resp, err := s.client.GetPrediction(taskId)
	if err != nil {
		return nil, err
	}

	return s.convertToTaskInfo(resp), nil
}

// TaskCancel 取消任务
func (s *NanoBananaService) TaskCancel(taskId string) error {
	_, err := s.client.CancelPrediction(taskId)
	return err
}

// TaskList 获取任务列表
func (s *NanoBananaService) TaskList() ([]*image2image.Image2ImageTaskInfo, error) {
	predictions, err := s.client.ListPredictions()
	if err != nil {
		return nil, err
	}

	tasks := make([]*image2image.Image2ImageTaskInfo, 0, len(predictions))
	for _, pred := range predictions {
		tasks = append(tasks, s.convertToTaskInfo(&pred))
	}

	return tasks, nil
}

// convertToNanoBananaInput 转换请求参数为 NanoBanana 输入格式
func (s *NanoBananaService) convertToNanoBananaInput(req *image2image.Image2ImageTaskRunReq) *NanoBananaInput {
	input := &NanoBananaInput{
		Prompt:       req.Prompt,
		ImageInput:   req.ImageInputs,
		OutputFormat: req.OutputFormat,
	}

	return input
}

// convertToTaskInfo 转换 Replicate 响应为任务信息
func (s *NanoBananaService) convertToTaskInfo(resp *PredictionResponse) *image2image.Image2ImageTaskInfo {
	task := &image2image.Image2ImageTaskInfo{
		TaskId: resp.ID,
		Status: ConvertStatusToInt(resp.Status),
	}

	// 解析时间字符串
	if resp.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, resp.CreatedAt); err == nil {
			task.CreateTime = int32(t.Unix())
			task.UpdateTime = int32(t.Unix())
		}
	}

	// 设置更新时间
	if resp.CompletedAt != "" {
		if t, err := time.Parse(time.RFC3339, resp.CompletedAt); err == nil {
			task.UpdateTime = int32(t.Unix())
		}
	} else if resp.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, resp.StartedAt); err == nil {
			task.UpdateTime = int32(t.Unix())
		}
	}

	// 设置执行时长
	if resp.Metrics.PredictTime > 0 {
		task.Duration = resp.Metrics.PredictTime
	}

	// 处理输出结果
	if resp.Output != nil {
		if outputStr, ok := resp.Output.(string); ok && outputStr != "" {
			task.Result = []string{outputStr}
		} else if outputSlice, ok := resp.Output.([]interface{}); ok {
			for _, item := range outputSlice {
				if str, ok := item.(string); ok {
					task.Result = append(task.Result, str)
				}
			}
		}
	}

	return task
}

type NanoBananaInput struct {
	Prompt       string   `json:"prompt"`
	ImageInput   []string `json:"image_input"`
	OutputFormat string   `json:"output_format"`
}
