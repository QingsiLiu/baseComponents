package wellapi

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CreateImageGeneration 调用 WellAPI 图片生成接口。
func (c *Client) CreateImageGeneration(req *ImageGenerateReq) (*ImageGenerateResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, fmt.Errorf("model is required")
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	httpReq, err := c.newJSONRequest("POST", c.baseURL+PathImagesGenerations, req)
	if err != nil {
		return nil, err
	}

	respBody, err := c.do(httpReq, req)
	if err != nil {
		return nil, err
	}

	var result ImageGenerateResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return &result, nil
}
