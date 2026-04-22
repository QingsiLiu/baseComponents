package wellapi

// ImageGenerateReq WellAPI 同步图片生成请求。
type ImageGenerateReq struct {
	Model          string `json:"model,omitempty"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

// GeneratedImage WellAPI 图像对象。
type GeneratedImage struct {
	B64JSON       string `json:"b64_json,omitempty"`
	URL           string `json:"url,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageGenerateResp WellAPI 同步图片生成响应。
type ImageGenerateResp struct {
	Created int64            `json:"created,omitempty"`
	Data    []GeneratedImage `json:"data"`
}
