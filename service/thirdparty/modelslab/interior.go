// https://modelslab.com/models/modelslab/interior
package modelslab

import (
	"fmt"
	"log"
	"strconv"

	"github.com/QingsiLiu/baseComponents/service/image2image"
)

type InteriorService struct {
	client *Client
}

func NewInteriorService() image2image.Image2ImageService {
	return &InteriorService{
		client: NewClient(),
	}
}

func NewInteriorServiceWithKey(apiKey string) image2image.Image2ImageService {
	return &InteriorService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 实现Image2ImageService接口
func (i *InteriorService) Source() string {
	return image2image.SourceModelsLabInterior
}

// TaskRun 实现Image2ImageService接口 - 提交图生图任务
func (i *InteriorService) TaskRun(req *image2image.Image2ImageTaskRunReq) (taskId string, err error) {
	interiorReq := i.convertToInteriorRequest(req)
	log.Printf("Interior TaskRun request: %+v", interiorReq)

	var resp TaskRunResponse
	err = i.client.PostAndDecode(InteriorEndpoint, interiorReq, &resp)
	if err != nil {
		log.Printf("Interior TaskRun error: %v", err)
		return "", fmt.Errorf("interior task run error: %w", err)
	}

	log.Printf("Interior TaskRun response: %+v", resp)
	return strconv.Itoa(resp.ID), nil
}

// TaskGet 实现Image2ImageService接口 - 获取任务状态
func (i *InteriorService) TaskGet(taskId string) (task *image2image.Image2ImageTaskInfo, err error) {
	req := i.client.CreateTaskGetRequest(taskId)
	log.Printf("Interior TaskGet request: %+v", req)

	var resp TaskGetResponse
	err = i.client.PostAndDecode(FetchEndpoint, req, &resp)
	if err != nil {
		log.Printf("Interior TaskGet error: %v", err)
		return nil, fmt.Errorf("interior task get error: %w", err)
	}

	log.Printf("Interior TaskGet response: %+v", resp)

	return i.convertToImage2ImageTaskInfo(&resp, taskId), nil
}

// TaskCancel 实现Image2ImageService接口 - 取消任务
func (i *InteriorService) TaskCancel(taskId string) error {
	return fmt.Errorf("task cancellation not supported by ModelsLab Interior API")
}

// TaskList 实现Image2ImageService接口 - 获取任务列表
func (i *InteriorService) TaskList() ([]*image2image.Image2ImageTaskInfo, error) {
	return nil, fmt.Errorf("task listing not supported by ModelsLab Interior API")
}

// convertToInteriorRequest 将通用请求转换为Interior API请求
func (i *InteriorService) convertToInteriorRequest(req *image2image.Image2ImageTaskRunReq) *InteriorTaskRunRequest {
	interiorReq := &InteriorTaskRunRequest{
		Key:            i.client.GetAPIKey(),
		Prompt:         req.Prompt,
		NegativePrompt: "blurry, low resolution, bad lighting, poorly drawn furniture, distorted proportions, messy room, unrealistic colors, extra limbs, missing furniture, bad anatomy, low detail, pixelated, grainy, artifacts, oversaturated, asymmetry, ugly, cartoonish, out of frame, duplicate objects",
		Strength:       req.Strength,
		GuidanceScale:  float64(req.GuidanceScale),
		Base64:         false,
		Temp:           false,
	}

	if len(req.ImageInputs) > 0 {
		interiorReq.InitImage = req.ImageInputs[0]
	}

	if req.Seed > 0 {
		interiorReq.Seed = &req.Seed
	}

	if interiorReq.Strength <= 0 {
		interiorReq.Strength = 1
	}
	if interiorReq.GuidanceScale <= 0 {
		interiorReq.GuidanceScale = 8
	}
	if interiorReq.NumInferenceSteps <= 0 {
		interiorReq.NumInferenceSteps = 31
	}

	return interiorReq
}

// convertToImage2ImageTaskInfo 将ModelsLab响应转换为Image2ImageTaskInfo
func (i *InteriorService) convertToImage2ImageTaskInfo(resp *TaskGetResponse, taskId string) *image2image.Image2ImageTaskInfo {
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

// InteriorTaskRunRequest Interior任务运行请求
type InteriorTaskRunRequest struct {
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