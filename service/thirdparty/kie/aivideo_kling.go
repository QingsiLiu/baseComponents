package kie

import (
	"strconv"
	"strings"

	"github.com/QingsiLiu/baseComponents/service/aivideo"
)

const (
	kling26ImageToVideoModelName = "kling-2.6/image-to-video"
	kling26TextToVideoModelName  = "kling-2.6/text-to-video"
	kling30VideoModelName        = "kling-3.0/video"
)

// Kling30VideoService KIE Kling 3.0 Video 视频生成服务实现。
type Kling30VideoService struct {
	*kieAIVideoTaskService
}

// Kling26ImageToVideoService KIE Kling 2.6 图生视频服务实现。
type Kling26ImageToVideoService struct {
	*kieAIVideoTaskService
}

// Kling26TextToVideoService KIE Kling 2.6 文生视频服务实现。
type Kling26TextToVideoService struct {
	*kieAIVideoTaskService
}

func NewKling30VideoService() aivideo.AIVideoService {
	return &Kling30VideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling30Video,
			kling30VideoModelName,
			"Kling 3.0 Video",
			NewClient(),
			buildKling30VideoInput,
		),
	}
}

func NewKling30VideoServiceWithKey(apiKey string) aivideo.AIVideoService {
	return &Kling30VideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling30Video,
			kling30VideoModelName,
			"Kling 3.0 Video",
			NewClientWithKey(apiKey),
			buildKling30VideoInput,
		),
	}
}

func NewKling26ImageToVideoService() aivideo.AIVideoService {
	return &Kling26ImageToVideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling26ImageToVideo,
			kling26ImageToVideoModelName,
			"Kling 2.6 Image To Video",
			NewClient(),
			buildKling26ImageToVideoInput,
		),
	}
}

func NewKling26ImageToVideoServiceWithKey(apiKey string) aivideo.AIVideoService {
	return &Kling26ImageToVideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling26ImageToVideo,
			kling26ImageToVideoModelName,
			"Kling 2.6 Image To Video",
			NewClientWithKey(apiKey),
			buildKling26ImageToVideoInput,
		),
	}
}

func NewKling26TextToVideoService() aivideo.AIVideoService {
	return &Kling26TextToVideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling26TextToVideo,
			kling26TextToVideoModelName,
			"Kling 2.6 Text To Video",
			NewClient(),
			buildKling26TextToVideoInput,
		),
	}
}

func NewKling26TextToVideoServiceWithKey(apiKey string) aivideo.AIVideoService {
	return &Kling26TextToVideoService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieKling26TextToVideo,
			kling26TextToVideoModelName,
			"Kling 2.6 Text To Video",
			NewClientWithKey(apiKey),
			buildKling26TextToVideoInput,
		),
	}
}

type Kling30VideoInput struct {
	Mode string `json:"mode"`
}

type Kling26ImageToVideoInput struct {
	Prompt    string   `json:"prompt,omitempty"`
	ImageURLs []string `json:"image_urls,omitempty"`
	Sound     bool     `json:"sound"`
	Duration  string   `json:"duration,omitempty"`
}

type Kling26TextToVideoInput struct {
	Prompt      string `json:"prompt,omitempty"`
	Sound       bool   `json:"sound"`
	AspectRatio string `json:"aspect_ratio,omitempty"`
	Duration    string `json:"duration,omitempty"`
}

func buildKling30VideoInput(req *aivideo.AIVideoTaskRunReq) interface{} {
	return &Kling30VideoInput{
		Mode: klingMode(req.Mode),
	}
}

func buildKling26ImageToVideoInput(req *aivideo.AIVideoTaskRunReq) interface{} {
	return &Kling26ImageToVideoInput{
		Prompt:    req.Prompt,
		ImageURLs: resolvePrimaryImageInputs(req),
		Sound:     boolValueOrDefault(req.GenerateAudio, false),
		Duration:  kling26Duration(req.Duration, "5"),
	}
}

func buildKling26TextToVideoInput(req *aivideo.AIVideoTaskRunReq) interface{} {
	return &Kling26TextToVideoInput{
		Prompt:      req.Prompt,
		Sound:       boolValueOrDefault(req.GenerateAudio, false),
		AspectRatio: stringValueOrDefault(req.AspectRatio, "1:1"),
		Duration:    kling26Duration(req.Duration, "5"),
	}
}

func klingMode(mode string) string {
	if strings.TrimSpace(mode) == "" {
		return "std"
	}
	return mode
}

func kling26Duration(duration int, fallback string) string {
	if duration == 0 {
		return fallback
	}
	return strconv.Itoa(duration)
}
