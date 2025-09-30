// https://modelslab.com/models/modelslab/exterior-restorer
package modelslab

import (
	"fmt"
	"log"
	"strconv"

	"github.com/QingsiLiu/baseComponents/service/image2image"
)

type ExteriorService struct {
	client *Client
}

func NewExteriorService() image2image.Image2ImageService {
	return &ExteriorService{
		client: NewClient(),
	}
}

func NewExteriorServiceWithKey(apiKey string) image2image.Image2ImageService {
	return &ExteriorService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 实现Image2ImageService接口
func (i *ExteriorService) Source() string {
	return image2image.SourceModelsLabExterior
}

// TaskRun 实现Image2ImageService接口 - 提交图生图任务
func (i *ExteriorService) TaskRun(req *image2image.Image2ImageTaskRunReq) (taskId string, err error) {
	exteriorReq := i.convertToExteriorRequest(req)
	log.Printf("Exterior TaskRun request: %+v", exteriorReq)

	var resp TaskRunResponse
	err = i.client.PostAndDecode(ExteriorEndpoint, exteriorReq, &resp)
	if err != nil {
		log.Printf("Exterior TaskRun error: %v", err)
		return "", fmt.Errorf("exterior task run error: %w", err)
	}

	log.Printf("Exterior TaskRun response: %+v", resp)
	return strconv.Itoa(resp.ID), nil
}

// TaskGet 实现Image2ImageService接口 - 获取任务状态
func (i *ExteriorService) TaskGet(taskId string) (task *image2image.Image2ImageTaskInfo, err error) {
	req := i.client.CreateTaskGetRequest(taskId)
	log.Printf("Exterior TaskGet request: %+v", req)

	var resp TaskGetResponse
	err = i.client.PostAndDecode(FetchEndpoint, req, &resp)
	if err != nil {
		log.Printf("Exterior TaskGet error: %v", err)
		return nil, fmt.Errorf("exterior task get error: %w", err)
	}

	log.Printf("Exterior TaskGet response: %+v", resp)

	return i.convertToImage2ImageTaskInfo(&resp, taskId), nil
}

// TaskCancel 实现Image2ImageService接口 - 取消任务
func (i *ExteriorService) TaskCancel(taskId string) error {
	return fmt.Errorf("task cancellation not supported by ModelsLab Exterior API")
}

// TaskList 实现Image2ImageService接口 - 获取任务列表
func (i *ExteriorService) TaskList() ([]*image2image.Image2ImageTaskInfo, error) {
	return nil, fmt.Errorf("task listing not supported by ModelsLab Exterior API")
}

// convertToExteriorRequest 将通用请求转换为Exterior API请求
func (i *ExteriorService) convertToExteriorRequest(req *image2image.Image2ImageTaskRunReq) *ExteriorTaskRunRequest {
	exteriorReq := &ExteriorTaskRunRequest{
		Key:            i.client.GetAPIKey(),
		Prompt:         req.Prompt,
		NegativePrompt: "blurry, low resolution, bad lighting, poorly drawn furniture, distorted proportions, messy room, unrealistic colors, extra limbs, missing furniture, bad anatomy, low detail, pixelated, grainy, artifacts, oversaturated, asymmetry, ugly, cartoonish, out of frame, duplicate objects",
		Strength:       req.Strength,
		GuidanceScale:  float64(req.GuidanceScale),
		Base64:         false,
		Temp:           false,
	}

	if len(req.ImageInputs) > 0 {
		exteriorReq.InitImage = req.ImageInputs[0]
	}

	if req.Seed > 0 {
		exteriorReq.Seed = &req.Seed
	}

	if exteriorReq.Strength <= 0 {
		exteriorReq.Strength = 1
	}
	if exteriorReq.GuidanceScale <= 0 {
		exteriorReq.GuidanceScale = 8
	}
	if exteriorReq.NumInferenceSteps <= 0 {
		exteriorReq.NumInferenceSteps = 31
	}

	return exteriorReq
}

// convertToImage2ImageTaskInfo 将ModelsLab响应转换为Image2ImageTaskInfo
func (i *ExteriorService) convertToImage2ImageTaskInfo(resp *TaskGetResponse, taskId string) *image2image.Image2ImageTaskInfo {
	// 转换状态
	status := ConvertStatusToInt(resp.Status)

	// 转换输出
	result := ConvertOutputToStringSlice(resp.Output)

	// 创建任务信息
	taskInfo := &image2image.Image2ImageTaskInfo{
		TaskId:     taskId,
		Status:     status,
		Result:     result,
		Duration:   0, // ModelsLab API不直接提供执行时间
		CreateTime: 0,
		UpdateTime: 0,
	}

	return taskInfo
}

// ExteriorTaskRunRequest Exterior任务运行请求
type ExteriorTaskRunRequest struct {
	Key               string  `json:"key,omitempty"`
	InitImage         string  `json:"init_image"`
	Prompt            string  `json:"prompt"`
	NegativePrompt    string  `json:"negative_prompt,omitempty"`
	Seed              *int    `json:"seed,omitempty"`
	GuidanceScale     float64 `json:"guidance_scale,omitempty"`
	Strength          float64 `json:"strength,omitempty"`
	NumInferenceSteps int64   `json:"num_inference_steps,omitempty"`
	Base64            bool    `json:"base64,omitempty"`
	Temp              bool    `json:"temp,omitempty"`
}
