package replicate

import (
	"fmt"
	"log"
	"github.com/QingsiLiu/baseComponents/service/image2image"
	"strconv"
	"time"

	"github.com/duke-git/lancet/v2/slice"
)

const (
	ControlNetModelVersion = "854e8727697a057c525cdb45ab037f64ecca770a1769cc52287c2e56472a247b"
)

type ControlNetService struct {
	client *Client
}

func NewControlNetService() image2image.Image2ImageService {
	return &ControlNetService{
		client: NewClient(),
	}
}

func NewControlNetServiceWithKey(token string) image2image.Image2ImageService {
	return &ControlNetService{
		client: NewClientWithToken(token),
	}
}

// Source implements image2image.Image2ImageService.
func (c *ControlNetService) Source() string {
	return image2image.SourceReplicateControlNet
}

// TaskRun implements image2image.Image2ImageService.
func (c *ControlNetService) TaskRun(req *image2image.Image2ImageTaskRunReq) (taskId string, err error) {
	input := c.convertToControlNetInput(req)

	// 创建预测任务
	predReq := &PredictionRequest{
		Version: ControlNetModelVersion,
		Input:   input,
	}

	resp, err := c.client.CreatePrediction(predReq)
	if err != nil {
		log.Printf("Create ControlNet prediction failed: %v", err)
		return "", fmt.Errorf("create prediction failed: %w", err)
	}

	return resp.ID, nil
}

// TaskGet implements image2image.Image2ImageService.
func (c *ControlNetService) TaskGet(taskId string) (task *image2image.Image2ImageTaskInfo, err error) {
	resp, err := c.client.GetPrediction(taskId)
	if err != nil {
		log.Printf("Get ControlNet prediction failed: %v", err)
		return nil, fmt.Errorf("get prediction failed: %w", err)
	}

	return c.convertToTaskInfo(resp), nil
}

// TaskCancel implements image2image.Image2ImageService.
func (c *ControlNetService) TaskCancel(taskId string) error {
	_, err := c.client.CancelPrediction(taskId)
	if err != nil {
		log.Printf("Cancel ControlNet prediction failed: %v", err)
		return fmt.Errorf("cancel prediction failed: %w", err)
	}

	return nil
}

// TaskList implements image2image.Image2ImageService.
func (c *ControlNetService) TaskList() ([]*image2image.Image2ImageTaskInfo, error) {
	predictions, err := c.client.ListPredictions()
	if err != nil {
		log.Printf("List ControlNet predictions failed: %v", err)
		return nil, fmt.Errorf("list predictions failed: %w", err)
	}

	tasks := make([]*image2image.Image2ImageTaskInfo, 0, len(predictions))
	for _, pred := range predictions {
		tasks = append(tasks, c.convertToTaskInfo(&pred))
	}

	return tasks, nil
}

// ControlNetInput ControlNet 模型输入参数
type ControlNetInput struct {
	Prompt          string `json:"prompt"`
	Image           string `json:"image"`
	DDimSteps       int    `json:"ddim_steps"`
	ImageResolution string `json:"image_resolution"`
	Scale           int    `json:"scale"`
}

// convertToControlNetInput 将通用请求转换为 ControlNet 输入格式
func (c *ControlNetService) convertToControlNetInput(req *image2image.Image2ImageTaskRunReq) *ControlNetInput {
	input := &ControlNetInput{
		Prompt:          req.Prompt,
		DDimSteps:       req.NumInferenceSteps,
		ImageResolution: strconv.Itoa(req.OutputQuality),
		Scale:           req.GuidanceScale,
	}

	if len(req.ImageInputs) > 0 {
		input.Image = req.ImageInputs[0]
	}

	return input
}

// convertToTaskInfo 将 PredictionResponse 转换为 Image2ImageTaskInfo
func (c *ControlNetService) convertToTaskInfo(resp *PredictionResponse) *image2image.Image2ImageTaskInfo {
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

	// 如果结果有多个元素且第一个元素需要删除（根据原始实现）
	if len(result) >= 2 {
		result = slice.DeleteAt(result, 0)
	}

	return &image2image.Image2ImageTaskInfo{
		TaskId:     resp.ID,
		Status:     ConvertStatusToInt(resp.Status),
		Result:     result,
		Duration:   resp.Metrics.PredictTime,
		CreateTime: createTime,
		UpdateTime: updateTime,
	}
}
