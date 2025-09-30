package replicate

import (
	"fmt"
	"log"
	"github.com/QingsiLiu/baseComponents/service/text2image"
	"time"
)

const (
	Flux1DevModelVersion = "prunaai/flux.1-dev:b0306d92aa025bb747dc74162f3c27d6ed83798e08e5f8977adf3d859d0536a3"
)

type Flux1DevService struct {
	client *Client
}

func NewFlux1DevService() text2image.Text2ImageService {
	return &Flux1DevService{
		client: NewClient(),
	}
}

func NewFlux1DevServiceWithKey(token string) text2image.Text2ImageService {
	return &Flux1DevService{
		client: NewClientWithToken(token),
	}
}

// Source implements text2image.Text2ImageService.
func (f *Flux1DevService) Source() string {
	return text2image.SourceReplicateFlux1Dev
}

// TaskRun implements text2image.Text2ImageService.
func (f *Flux1DevService) TaskRun(req *text2image.Text2ImageTaskRunReq) (taskId string, err error) {
	input := f.convertToFluxDevInput(req)

	// 创建预测任务
	predReq := &PredictionRequest{
		Version: Flux1DevModelVersion,
		Input:   input,
	}

	resp, err := f.client.CreatePrediction(predReq)
	if err != nil {
		log.Printf("Create Flux1Dev prediction failed: %v", err)
		return "", fmt.Errorf("create prediction failed: %w", err)
	}

	return resp.ID, nil
}

// TaskGet implements text2image.Text2ImageService.
func (f *Flux1DevService) TaskGet(taskId string) (task *text2image.Text2ImageTaskInfo, err error) {
	resp, err := f.client.GetPrediction(taskId)
	if err != nil {
		log.Printf("Get Flux1Dev prediction failed: %v", err)
		return nil, fmt.Errorf("get prediction failed: %w", err)
	}

	return f.convertToTaskInfo(resp), nil
}

// TaskCancel implements text2image.Text2ImageService.
func (f *Flux1DevService) TaskCancel(taskId string) error {
	_, err := f.client.CancelPrediction(taskId)
	if err != nil {
		log.Printf("Cancel Flux1Dev prediction failed: %v", err)
		return fmt.Errorf("cancel prediction failed: %w", err)
	}

	return nil
}

// TaskList implements text2image.Text2ImageService.
func (f *Flux1DevService) TaskList() ([]*text2image.Text2ImageTaskInfo, error) {
	predictions, err := f.client.ListPredictions()
	if err != nil {
		log.Printf("List Flux1Dev predictions failed: %v", err)
		return nil, fmt.Errorf("list predictions failed: %w", err)
	}

	tasks := make([]*text2image.Text2ImageTaskInfo, 0, len(predictions))
	for _, pred := range predictions {
		// 只返回 Flux1Dev 模型的任务
		if pred.Model == Flux1DevModelVersion || pred.Version == Flux1DevModelVersion {
			tasks = append(tasks, f.convertToTaskInfo(&pred))
		}
	}

	return tasks, nil
}

// convertToFluxDevInput 将通用请求转换为 Flux Dev 输入格式
func (f *Flux1DevService) convertToFluxDevInput(req *text2image.Text2ImageTaskRunReq) *FluxDevInput {
	input := &FluxDevInput{
		Seed:              req.Seed,
		Prompt:            req.Prompt,
		Guidance:          req.Guidance,
		ImageSize:         req.ImageSize,
		SpeedMode:         req.SpeedMode,
		AspectRatio:       req.AspectRatio,
		OutputFormat:      req.OutputFormat,
		OutputQuality:     req.OutputQuality,
		NumInferenceSteps: req.NumInferenceSteps,
	}

	return input
}

// convertToTaskInfo 将 PredictionResponse 转换为 Text2ImageTaskInfo
func (f *Flux1DevService) convertToTaskInfo(resp *PredictionResponse) *text2image.Text2ImageTaskInfo {
	// 解析时间字符串为Unix时间戳
	var createTime, updateTime int32

	if resp.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, resp.CreatedAt); err == nil {
			createTime = int32(t.Unix())
		}
	}

	// 优先使用completed_at，其次是started_at，最后是created_at
	if resp.CompletedAt != "" {
		if t, err := time.Parse(time.RFC3339, resp.CompletedAt); err == nil {
			updateTime = int32(t.Unix())
		}
	} else if resp.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, resp.StartedAt); err == nil {
			updateTime = int32(t.Unix())
		}
	} else {
		updateTime = createTime
	}

	// 转换输出结果
	result := ConvertOutputToStringSlice(resp.Output)

	return &text2image.Text2ImageTaskInfo{
		TaskId:     resp.ID,
		Status:     ConvertStatusToInt(resp.Status),
		Result:     result,
		Duration:   resp.Metrics.PredictTime,
		CreateTime: createTime,
		UpdateTime: updateTime,
	}
}

type FluxDevInput struct {
	Seed              int     `json:"seed,omitempty"`
	Prompt            string  `json:"prompt"`
	Guidance          float64 `json:"guidance,omitempty"`
	ImageSize         int     `json:"image_size,omitempty"`
	SpeedMode         string  `json:"speed_mode,omitempty"`
	AspectRatio       string  `json:"aspect_ratio,omitempty"`
	OutputFormat      string  `json:"output_format,omitempty"`
	OutputQuality     int     `json:"output_quality,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
}
