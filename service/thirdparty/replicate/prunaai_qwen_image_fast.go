package replicate

import (
	"fmt"
	"log"
	"time"

	"github.com/QingsiLiu/baseComponents/service/text2image"
)

const (
	PrunaAIQwenImageFastModelVersion = "prunaai/qwen-image-fast:01b324d214eb4870ff424dc4215c067759c4c01a8751e327a434e2b16054db2f"
	defaultAspectRatio               = "1:1"
)

// PrunaAIQwenImageFastService implements text2image.Text2ImageService for prunaai/qwen-image-fast
type PrunaAIQwenImageFastService struct {
	client *Client
}

func NewPrunaAIQwenImageFastService() text2image.Text2ImageService {
	return &PrunaAIQwenImageFastService{
		client: NewClient(),
	}
}

func NewPrunaAIQwenImageFastServiceWithKey(token string) text2image.Text2ImageService {
	return &PrunaAIQwenImageFastService{
		client: NewClientWithToken(token),
	}
}

// Source implements text2image.Text2ImageService.
func (p *PrunaAIQwenImageFastService) Source() string {
	return text2image.SourceReplicatePrunaAIQwenImageFast
}

// TaskRun implements text2image.Text2ImageService.
func (p *PrunaAIQwenImageFastService) TaskRun(req *text2image.Text2ImageTaskRunReq) (taskId string, err error) {
	input := p.convertToQwenImageFastInput(req)

	predReq := &PredictionRequest{
		Version: PrunaAIQwenImageFastModelVersion,
		Input:   input,
	}

	resp, err := p.client.CreatePrediction(predReq)
	if err != nil {
		log.Printf("Create PrunaAI Qwen Image Fast prediction failed: %v", err)
		return "", fmt.Errorf("create prediction failed: %w", err)
	}

	return resp.ID, nil
}

// TaskGet implements text2image.Text2ImageService.
func (p *PrunaAIQwenImageFastService) TaskGet(taskId string) (task *text2image.Text2ImageTaskInfo, err error) {
	resp, err := p.client.GetPrediction(taskId)
	if err != nil {
		log.Printf("Get PrunaAI Qwen Image Fast prediction failed: %v", err)
		return nil, fmt.Errorf("get prediction failed: %w", err)
	}

	return p.convertToTaskInfo(resp), nil
}

// TaskCancel implements text2image.Text2ImageService.
func (p *PrunaAIQwenImageFastService) TaskCancel(taskId string) error {
	_, err := p.client.CancelPrediction(taskId)
	if err != nil {
		log.Printf("Cancel PrunaAI Qwen Image Fast prediction failed: %v", err)
		return fmt.Errorf("cancel prediction failed: %w", err)
	}

	return nil
}

// TaskList implements text2image.Text2ImageService.
func (p *PrunaAIQwenImageFastService) TaskList() ([]*text2image.Text2ImageTaskInfo, error) {
	predictions, err := p.client.ListPredictions()
	if err != nil {
		log.Printf("List PrunaAI Qwen Image Fast predictions failed: %v", err)
		return nil, fmt.Errorf("list predictions failed: %w", err)
	}

	tasks := make([]*text2image.Text2ImageTaskInfo, 0, len(predictions))
	for _, pred := range predictions {
		if pred.Model == PrunaAIQwenImageFastModelVersion || pred.Version == PrunaAIQwenImageFastModelVersion {
			tasks = append(tasks, p.convertToTaskInfo(&pred))
		}
	}

	return tasks, nil
}

func (p *PrunaAIQwenImageFastService) convertToQwenImageFastInput(req *text2image.Text2ImageTaskRunReq) *PrunaAIQwenImageFastInput {
	input := &PrunaAIQwenImageFastInput{
		Prompt:               req.Prompt,
		Width:                req.ImageWidth,
		Height:               req.ImageHeight,
		AspectRatio:          req.AspectRatio,
		Creativity:           req.Guidance,
		Seed:                 0,
		DisableSafetyChecker: req.DisableSafetyChecker,
	}

	if input.AspectRatio == "" {
		input.AspectRatio = defaultAspectRatio
	}

	return input
}

func (p *PrunaAIQwenImageFastService) convertToTaskInfo(resp *PredictionResponse) *text2image.Text2ImageTaskInfo {
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

type PrunaAIQwenImageFastInput struct {
	Prompt               string  `json:"prompt"`
	Creativity           float64 `json:"creativity,omitempty"`
	AspectRatio          string  `json:"aspect_ratio,omitempty"`
	Width                int     `json:"width,omitempty"`
	Height               int     `json:"height,omitempty"`
	Seed                 int     `json:"seed,omitempty"`
	DisableSafetyChecker bool    `json:"disable_safety_checker,omitempty"`
}
