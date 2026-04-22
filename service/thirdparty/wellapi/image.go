package wellapi

import (
	"fmt"
	"strings"
)

// ImageService WellAPI 同步图片生成服务实现。
type ImageService struct {
	client *Client
}

// NewImageService 创建默认服务实例。
func NewImageService() *ImageService {
	return &ImageService{
		client: NewClient(),
	}
}

// NewImageServiceWithKey 使用指定 API Key 创建服务实例。
func NewImageServiceWithKey(apiKey string) *ImageService {
	return &ImageService{
		client: NewClientWithKey(apiKey),
	}
}

// Generate 执行一次同步图片生成请求。
func (s *ImageService) Generate(req *ImageGenerateReq) (*ImageGenerateResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	wireReq := *req
	if strings.TrimSpace(wireReq.Model) == "" {
		wireReq.Model = ModelGPTImage2
	}
	if wireReq.N <= 0 {
		wireReq.N = 1
	}
	if strings.TrimSpace(wireReq.Size) == "" {
		wireReq.Size = "auto"
	}
	if strings.TrimSpace(wireReq.ResponseFormat) == "" {
		wireReq.ResponseFormat = "url"
	}

	return s.client.CreateImageGeneration(&wireReq)
}
