package replicate

import (
	"fmt"
	"log"
	"time"

	"github.com/QingsiLiu/baseComponents/service/text2image"
)

const (
	QwenImageModelVersion = "qwen/qwen-image"
)

// QwenImageService implements text2image.Text2ImageService for qwen/qwen-image
type QwenImageService struct {
	client *Client
}

func NewQwenImageService() text2image.Text2ImageService {
	return &QwenImageService{
		client: NewClient(),
	}
}

func NewQwenImageServiceWithKey(token string) text2image.Text2ImageService {
	return &QwenImageService{
		client: NewClientWithToken(token),
	}
}

// Source implements text2image.Text2ImageService.
func (q *QwenImageService) Source() string {
	return text2image.SourceReplicateQwenImage
}

// TaskRun implements text2image.Text2ImageService.
func (q *QwenImageService) TaskRun(req *text2image.Text2ImageTaskRunReq) (taskId string, err error) {
	input := q.convertToQwenImageInput(req)

	predReq := &PredictionRequest{
		Version: QwenImageModelVersion,
		Input:   input,
	}

	resp, err := q.client.CreatePrediction(predReq)
	if err != nil {
		log.Printf("Create Qwen Image prediction failed: %v", err)
		return "", fmt.Errorf("create prediction failed: %w", err)
	}

	return resp.ID, nil
}

// TaskGet implements text2image.Text2ImageService.
func (q *QwenImageService) TaskGet(taskId string) (task *text2image.Text2ImageTaskInfo, err error) {
	resp, err := q.client.GetPrediction(taskId)
	if err != nil {
		log.Printf("Get Qwen Image prediction failed: %v", err)
		return nil, fmt.Errorf("get prediction failed: %w", err)
	}

	return q.convertToTaskInfo(resp), nil
}

// TaskCancel implements text2image.Text2ImageService.
func (q *QwenImageService) TaskCancel(taskId string) error {
	_, err := q.client.CancelPrediction(taskId)
	if err != nil {
		log.Printf("Cancel Qwen Image prediction failed: %v", err)
		return fmt.Errorf("cancel prediction failed: %w", err)
	}

	return nil
}

// TaskList implements text2image.Text2ImageService.
func (q *QwenImageService) TaskList() ([]*text2image.Text2ImageTaskInfo, error) {
	predictions, err := q.client.ListPredictions()
	if err != nil {
		log.Printf("List Qwen Image predictions failed: %v", err)
		return nil, fmt.Errorf("list predictions failed: %w", err)
	}

	tasks := make([]*text2image.Text2ImageTaskInfo, 0, len(predictions))
	for _, pred := range predictions {
		if pred.Model == QwenImageModelVersion || pred.Version == QwenImageModelVersion {
			tasks = append(tasks, q.convertToTaskInfo(&pred))
		}
	}

	return tasks, nil
}

func (q *QwenImageService) convertToQwenImageInput(req *text2image.Text2ImageTaskRunReq) *QwenImageInput {
	return &QwenImageInput{
		Seed:                 req.Seed,
		Prompt:               req.Prompt,
		Guidance:             req.Guidance,
		Strength:             req.Strength,
		AspectRatio:          req.AspectRatio,
		OutputFormat:         req.OutputFormat,
		OutputQuality:        req.OutputQuality,
		NegativePrompt:       req.NegativePrompt,
		NumInferenceSteps:    req.NumInferenceSteps,
		DisableSafetyChecker: req.DisableSafetyChecker,
	}
}

func (q *QwenImageService) convertToTaskInfo(resp *PredictionResponse) *text2image.Text2ImageTaskInfo {
	var createTime, updateTime int32

	if resp.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, resp.CreatedAt); err == nil {
			createTime = int32(t.Unix())
		}
	}

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

type QwenImageInput struct {
	Seed                 int       `json:"seed,omitempty"`
	Image                string    `json:"image,omitempty"`
	Prompt               string    `json:"prompt"`
	GoFast               bool      `json:"go_fast,omitempty"`
	Guidance             float64   `json:"guidance,omitempty"`
	Strength             float64   `json:"strength,omitempty"`
	ImageSize            string    `json:"image_size,omitempty"`
	LoraScale            float64   `json:"lora_scale,omitempty"`
	AspectRatio          string    `json:"aspect_ratio,omitempty"`
	LoraWeights          string    `json:"lora_weights,omitempty"`
	OutputFormat         string    `json:"output_format,omitempty"`
	EnhancePrompt        bool      `json:"enhance_prompt,omitempty"`
	OutputQuality        int       `json:"output_quality,omitempty"`
	NegativePrompt       string    `json:"negative_prompt,omitempty"`
	ExtraLoraScale       []float64 `json:"extra_lora_scale,omitempty"`
	ExtraLoraWeights     []string  `json:"extra_lora_weights,omitempty"`
	NumInferenceSteps    int       `json:"num_inference_steps,omitempty"`
	DisableSafetyChecker bool      `json:"disable_safety_checker,omitempty"`
}
