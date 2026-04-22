package edit

import (
	"context"

	"github.com/QingsiLiu/baseComponents/service/v2/core"
)

type Request struct {
	Prompt         string          `json:"prompt"`
	NegativePrompt string          `json:"negative_prompt,omitempty"`
	Images         []string        `json:"images"`
	Seed           int             `json:"seed,omitempty"`
	Strength       float64         `json:"strength,omitempty"`
	Count          int             `json:"count,omitempty"`
	AspectRatio    string          `json:"aspect_ratio,omitempty"`
	OutputFormat   string          `json:"output_format,omitempty"`
	OutputQuality  int             `json:"output_quality,omitempty"`
	SafetyMode     core.SafetyMode `json:"safety_mode,omitempty"`
}

type Image struct {
	URL string `json:"url,omitempty"`
}

type Result struct {
	Images []Image `json:"images"`
}

type Service interface {
	Run(ctx context.Context, target core.Target, req *Request) (*core.Operation[Result], error)
	Refresh(ctx context.Context, op *core.Operation[Result]) error
	Cancel(ctx context.Context, op *core.Operation[Result]) error
}
