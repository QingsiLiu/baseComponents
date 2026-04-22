package generate

import (
	"context"

	"github.com/QingsiLiu/baseComponents/service/v2/core"
)

type Request struct {
	Prompt          string          `json:"prompt"`
	ReferenceImages []string        `json:"reference_images,omitempty"`
	ReferenceVideos []string        `json:"reference_videos,omitempty"`
	Seed            int             `json:"seed,omitempty"`
	DurationSeconds int             `json:"duration_seconds,omitempty"`
	AspectRatio     string          `json:"aspect_ratio,omitempty"`
	Resolution      string          `json:"resolution,omitempty"`
	GenerateAudio   *bool           `json:"generate_audio,omitempty"`
	SafetyMode      core.SafetyMode `json:"safety_mode,omitempty"`
}

type Video struct {
	URL string `json:"url,omitempty"`
}

type Image struct {
	URL string `json:"url,omitempty"`
}

type Result struct {
	Videos []Video `json:"videos,omitempty"`
	Images []Image `json:"images,omitempty"`
}

type Service interface {
	Run(ctx context.Context, target core.Target, req *Request) (*core.Operation[Result], error)
	Refresh(ctx context.Context, op *core.Operation[Result]) error
	Cancel(ctx context.Context, op *core.Operation[Result]) error
}
