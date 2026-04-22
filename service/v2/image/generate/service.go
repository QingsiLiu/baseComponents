package generate

import (
	"context"

	"github.com/QingsiLiu/baseComponents/service/v2/core"
)

type Request struct {
	Prompt         string          `json:"prompt"`
	NegativePrompt string          `json:"negative_prompt,omitempty"`
	Seed           int             `json:"seed,omitempty"`
	Count          int             `json:"count,omitempty"`
	AspectRatio    string          `json:"aspect_ratio,omitempty"`
	Width          int             `json:"width,omitempty"`
	Height         int             `json:"height,omitempty"`
	OutputFormat   string          `json:"output_format,omitempty"`
	OutputQuality  int             `json:"output_quality,omitempty"`
	SafetyMode     core.SafetyMode `json:"safety_mode,omitempty"`
}

type Image struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type Result struct {
	Images []Image `json:"images"`
}

type Service interface {
	Run(ctx context.Context, target core.Target, req *Request) (*core.Operation[Result], error)
	Refresh(ctx context.Context, op *core.Operation[Result]) error
	Cancel(ctx context.Context, op *core.Operation[Result]) error
}
