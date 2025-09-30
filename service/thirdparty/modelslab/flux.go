// https://modelslab.com/models/modelslab/flux-text-to-image
package modelslab

import (
	"fmt"
	"log"
	"strconv"

	"github.com/QingsiLiu/baseComponents/service/text2image"
)

type FluxService struct {
	client *Client
}

func NewFluxService() text2image.Text2ImageService {
	return &FluxService{
		client: NewClient(),
	}
}

func NewFluxServiceWithKey(apiKey string) text2image.Text2ImageService {
	return &FluxService{
		client: NewClientWithKey(apiKey),
	}
}

// Source 实现Text2ImageService接口
func (f *FluxService) Source() string {
	return text2image.SourceModelslabFlux
}

// TaskRun 实现Text2ImageService接口 - 提交文本转图像任务
func (f *FluxService) TaskRun(req *text2image.Text2ImageTaskRunReq) (taskId string, err error) {
	fluxReq := f.convertToFluxRequest(req)
	log.Printf("Flux TaskRun request: %+v", fluxReq)

	var resp TaskRunResponse
	err = f.client.PostAndDecode(Text2ImgEndpoint, fluxReq, &resp)
	if err != nil {
		log.Printf("Flux TaskRun error: %v", err)
		return "", fmt.Errorf("flux task run error: %w", err)
	}

	log.Printf("Flux TaskRun response: %+v", resp)
	return strconv.Itoa(resp.ID), nil
}

// TaskGet 实现Text2ImageService接口 - 获取任务状态
func (f *FluxService) TaskGet(taskId string) (task *text2image.Text2ImageTaskInfo, err error) {
	req := f.client.CreateTaskGetRequest(taskId)
	log.Printf("Flux TaskGet request: %+v", req)

	var resp TaskGetResponse
	err = f.client.PostAndDecode(FetchEndpoint, req, &resp)
	if err != nil {
		log.Printf("Flux TaskGet error: %v", err)
		return nil, fmt.Errorf("flux task get error: %w", err)
	}

	log.Printf("Flux TaskGet response: %+v", resp)

	return f.convertToText2ImageTaskInfo(&resp, taskId), nil
}

// TaskCancel 实现Text2ImageService接口 - 取消任务
func (f *FluxService) TaskCancel(taskId string) error {
	return fmt.Errorf("task cancellation not supported by ModelsLab Flux API")
}

// TaskList 实现Text2ImageService接口 - 获取任务列表
func (f *FluxService) TaskList() ([]*text2image.Text2ImageTaskInfo, error) {
	return nil, fmt.Errorf("task listing not supported by ModelsLab Flux API")
}

// convertToFluxRequest 将通用请求转换为Flux API请求
func (f *FluxService) convertToFluxRequest(req *text2image.Text2ImageTaskRunReq) *FluxTaskRunRequest {
	fluxReq := &FluxTaskRunRequest{
		Key:               f.client.GetAPIKey(),
		ModelID:           "flux",
		Prompt:            req.Prompt,
		NegativePrompt:    req.NegativePrompt,
		Width:             req.ImageWidth,
		Height:            req.ImageHeight,
		Samples:           req.NumOutputs,
		NumInferenceSteps: req.NumInferenceSteps,
		SafetyChecker:     "no",
		GuidanceScale:     7.5,
		Panorama:          "no",
		SelfAttention:     "yes",
		Tomesd:            "yes",
		Scheduler:         "DPMSolverMultistepScheduler",
	}

	// 设置种子
	if req.Seed > 0 {
		fluxReq.Seed = &req.Seed
	}

	if fluxReq.Width <= 0 {
		fluxReq.Width = 1024
	}
	if fluxReq.Height <= 0 {
		fluxReq.Height = 1024
	}

	if fluxReq.Samples <= 0 {
		fluxReq.Samples = 1
	}

	if fluxReq.NumInferenceSteps <= 0 {
		fluxReq.NumInferenceSteps = 20
	}

	return fluxReq
}

// convertToText2ImageTaskInfo 将ModelsLab响应转换为Text2ImageTaskInfo
func (f *FluxService) convertToText2ImageTaskInfo(resp *TaskGetResponse, taskId string) *text2image.Text2ImageTaskInfo {
	// 转换状态
	status := ConvertStatusToInt(resp.Status)

	// 转换输出
	result := ConvertOutputToStringSlice(resp.Output)

	// 创建任务信息
	taskInfo := &text2image.Text2ImageTaskInfo{
		TaskId:     taskId,
		Status:     status,
		Result:     result,
		Duration:   0, // ModelsLab API不直接提供执行时间
		CreateTime: 0,
		UpdateTime: 0,
	}

	return taskInfo
}

// FluxTaskRunRequest Flux任务运行请求
type FluxTaskRunRequest struct {
	Key               string  `json:"key,omitempty"`
	ModelID           string  `json:"model_id"`
	Prompt            string  `json:"prompt"`
	NegativePrompt    string  `json:"negative_prompt,omitempty"`
	Width             int     `json:"width,omitempty"`
	Height            int     `json:"height,omitempty"`
	Samples           int     `json:"samples,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
	SafetyChecker     string  `json:"safety_checker,omitempty"`
	EnhancePrompt     string  `json:"enhance_prompt,omitempty"`
	Seed              *int    `json:"seed,omitempty"`
	GuidanceScale     float64 `json:"guidance_scale,omitempty"`
	Panorama          string  `json:"panorama,omitempty"`
	SelfAttention     string  `json:"self_attention,omitempty"`
	LoraModel         *string `json:"lora_model,omitempty"`
	Tomesd            string  `json:"tomesd"`
	ClipSkip          *string `json:"clip_skip,omitempty"`
	UseKarrasSigmas   *string `json:"use_karras_sigmas,omitempty"`
	Vae               *string `json:"vae,omitempty"`
	LoraStrength      *string `json:"lora_strength,omitempty"`
	Scheduler         string  `json:"scheduler"`
	Webhook           *string `json:"webhook,omitempty"`
	TrackID           *string `json:"track_id,omitempty"`
}