package replicate

import (
	"fmt"
	"log"
	"time"

	"github.com/QingsiLiu/baseComponents/service/aivideo"
)

const (
	PixverseV5ModelVersion = "pixverse/pixverse-v5"
)

type PixverseV5Service struct {
	client *Client
}

func NewPixverseV5Service() aivideo.AIVideoService {
	return &PixverseV5Service{
		client: NewClient(),
	}
}

func NewPixverseV5ServiceWithKey(token string) aivideo.AIVideoService {
	return &PixverseV5Service{
		client: NewClientWithToken(token),
	}
}

// Source implements aivideo.AIVideoService.
func (p *PixverseV5Service) Source() string {
	return aivideo.SourceReplicatePixverse
}

// TaskRun implements aivideo.AIVideoService.
func (p *PixverseV5Service) TaskRun(req *aivideo.AIVideoTaskRunReq) (taskId string, err error) {
	input := p.convertToPixverseInput(req)

	// 创建预测任务
	predReq := &PredictionRequest{
		Version: PixverseV5ModelVersion,
		Input:   input,
	}

	resp, err := p.client.CreatePrediction(predReq)
	if err != nil {
		log.Printf("Create PixverseV5 prediction failed: %v", err)
		return "", fmt.Errorf("create prediction failed: %w", err)
	}

	return resp.ID, nil
}

// TaskGet implements aivideo.AIVideoService.
func (p *PixverseV5Service) TaskGet(taskId string) (task *aivideo.AIVideoTaskInfo, err error) {
	resp, err := p.client.GetPrediction(taskId)
	if err != nil {
		log.Printf("Get PixverseV5 prediction failed: %v", err)
		return nil, fmt.Errorf("get prediction failed: %w", err)
	}

	return p.convertToTaskInfo(resp), nil
}

// TaskCancel implements aivideo.AIVideoService.
func (p *PixverseV5Service) TaskCancel(taskId string) error {
	_, err := p.client.CancelPrediction(taskId)
	if err != nil {
		log.Printf("Cancel PixverseV5 prediction failed: %v", err)
		return fmt.Errorf("cancel prediction failed: %w", err)
	}

	return nil
}

// TaskList implements aivideo.AIVideoService.
func (p *PixverseV5Service) TaskList() ([]*aivideo.AIVideoTaskInfo, error) {
	predictions, err := p.client.ListPredictions()
	if err != nil {
		log.Printf("List PixverseV5 predictions failed: %v", err)
		return nil, fmt.Errorf("list predictions failed: %w", err)
	}

	tasks := make([]*aivideo.AIVideoTaskInfo, 0, len(predictions))
	for _, pred := range predictions {
		// 只返回 PixverseV5 模型的任务
		if pred.Model == PixverseV5ModelVersion || pred.Version == PixverseV5ModelVersion {
			tasks = append(tasks, p.convertToTaskInfo(&pred))
		}
	}

	return tasks, nil
}

// convertToPixverseInput 将通用请求转换为 Pixverse V5 输入格式
func (p *PixverseV5Service) convertToPixverseInput(req *aivideo.AIVideoTaskRunReq) *PixverseV5Input {
	input := &PixverseV5Input{
		Prompt:         req.Prompt,
		Quality:        req.Quality,
		Duration:       req.Duration,
		AspectRatio:    req.AspectRatio,
		NegativePrompt: req.NegativePrompt,
		Seed:           req.Seed,
	}

	// 处理可选参数
	if req.Image != "" {
		input.Image = req.Image
	}

	if req.LastFrameImage != "" {
		input.LastFrameImage = req.LastFrameImage
	}

	if req.Effect != "" {
		input.Effect = req.Effect
	}

	return input
}

// convertToTaskInfo 将 PredictionResponse 转换为 AIVideoTaskInfo
func (p *PixverseV5Service) convertToTaskInfo(resp *PredictionResponse) *aivideo.AIVideoTaskInfo {
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

	return &aivideo.AIVideoTaskInfo{
		TaskId:     resp.ID,
		Status:     ConvertStatusToInt(resp.Status),
		Result:     result,
		Duration:   resp.Metrics.PredictTime,
		CreateTime: createTime,
		UpdateTime: updateTime,
	}
}

// PixverseV5Input Pixverse V5 模型输入参数
type PixverseV5Input struct {
	Prompt         string `json:"prompt"`                    // 文本描述（必填）
	Quality        string `json:"quality,omitempty"`         // 视频分辨率: 360p, 540p, 720p, 1080p (默认: 540p)
	Duration       int    `json:"duration,omitempty"`        // 视频时长（秒）: 5, 8 (默认: 5)
	AspectRatio    string `json:"aspect_ratio,omitempty"`    // 视频比例: 16:9, 9:16, 1:1 (默认: 16:9)
	Image          string `json:"image,omitempty"`           // 首帧图片URL（可选）
	LastFrameImage string `json:"last_frame_image,omitempty"` // 末帧图片URL（可选，需配合image使用）
	Effect         string `json:"effect,omitempty"`          // 特殊效果（可选）
	NegativePrompt string `json:"negative_prompt,omitempty"` // 负面提示词
	Seed           int    `json:"seed,omitempty"`            // 随机种子
}
