package replicate

import (
	"fmt"
	"log"
	"github.com/QingsiLiu/baseComponents/service/text2image"
	"time"
)

const (
	FluxSchnellModelVersion = "black-forest-labs/flux-schnell"
)

type FluxSchnellService struct {
	client *Client
}

func NewFluxSchnellService() text2image.Text2ImageService {
	return &FluxSchnellService{
		client: NewClient(),
	}
}

func NewFluxSchnellServiceWithKey(token string) text2image.Text2ImageService {
	return &FluxSchnellService{
		client: NewClientWithToken(token),
	}
}

// Source implements text2image.Text2ImageService.
func (f *FluxSchnellService) Source() string {
	return text2image.SourceReplicateFluxSchnell
}

// TaskRun implements text2image.Text2ImageService.
func (f *FluxSchnellService) TaskRun(req *text2image.Text2ImageTaskRunReq) (taskId string, err error) {
	input := f.convertToFluxSchnellInput(req)

	// 创建预测任务
	predReq := &PredictionRequest{
		Version: FluxSchnellModelVersion,
		Input:   input,
	}

	resp, err := f.client.CreatePrediction(predReq)
	if err != nil {
		log.Printf("Create FluxSchnell prediction failed: %v", err)
		return "", fmt.Errorf("create prediction failed: %w", err)
	}

	return resp.ID, nil
}

// TaskGet implements text2image.Text2ImageService.
func (f *FluxSchnellService) TaskGet(taskId string) (task *text2image.Text2ImageTaskInfo, err error) {
	resp, err := f.client.GetPrediction(taskId)
	if err != nil {
		log.Printf("Get FluxSchnell prediction failed: %v", err)
		return nil, fmt.Errorf("get prediction failed: %w", err)
	}

	return f.convertToTaskInfo(resp), nil
}

// TaskCancel implements text2image.Text2ImageService.
func (f *FluxSchnellService) TaskCancel(taskId string) error {
	_, err := f.client.CancelPrediction(taskId)
	if err != nil {
		log.Printf("Cancel FluxSchnell prediction failed: %v", err)
		return fmt.Errorf("cancel prediction failed: %w", err)
	}

	return nil
}

// TaskList implements text2image.Text2ImageService.
func (f *FluxSchnellService) TaskList() ([]*text2image.Text2ImageTaskInfo, error) {
	predictions, err := f.client.ListPredictions()
	if err != nil {
		log.Printf("List FluxSchnell predictions failed: %v", err)
		return nil, fmt.Errorf("list predictions failed: %w", err)
	}

	tasks := make([]*text2image.Text2ImageTaskInfo, 0, len(predictions))
	for _, pred := range predictions {
		// 只返回 FluxSchnell 模型的任务
		if pred.Model == FluxSchnellModelVersion || pred.Version == FluxSchnellModelVersion {
			tasks = append(tasks, f.convertToTaskInfo(&pred))
		}
	}

	return tasks, nil
}

// convertToFluxSchnellInput 将通用请求转换为 Flux Schnell 输入格式
func (f *FluxSchnellService) convertToFluxSchnellInput(req *text2image.Text2ImageTaskRunReq) *FluxSchnellInput {
	return &FluxSchnellInput{
		Seed:              req.Seed,
		Prompt:            req.Prompt,
		Megapixels:        "1",
		SpeedMode:         req.SpeedMode,
		NumOutputs:        req.NumOutputs,
		AspectRatio:       req.AspectRatio,
		OutputFormat:      "jpg",
		OutputQuality:     req.OutputQuality,
		NumInferenceSteps: req.NumInferenceSteps,
	}
}

// convertToTaskInfo 将 PredictionResponse 转换为 Text2ImageTaskInfo
func (f *FluxSchnellService) convertToTaskInfo(resp *PredictionResponse) *text2image.Text2ImageTaskInfo {
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

type FluxSchnellInput struct {
	Seed              int    `json:"seed,omitempty"`
	Prompt            string `json:"prompt"`
	Megapixels        string `json:"megapixels,omitempty"`
	SpeedMode         string `json:"speed_mode,omitempty"`
	NumOutputs        int    `json:"num_outputs,omitempty"`
	AspectRatio       string `json:"aspect_ratio,omitempty"`
	OutputFormat      string `json:"output_format,omitempty"`
	OutputQuality     int    `json:"output_quality,omitempty"`
	NumInferenceSteps int    `json:"num_inference_steps,omitempty"`
}