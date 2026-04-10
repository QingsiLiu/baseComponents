package kie

import (
	"strconv"

	"github.com/QingsiLiu/baseComponents/service/aivideo"
)

const (
	seedance15ProModelName = "bytedance/seedance-1.5-pro"
	seedance2ModelName     = "bytedance/seedance-2"
	seedance2FastModelName = "bytedance/seedance-2-fast"
)

// Seedance15ProService KIE Seedance 1.5 Pro 视频生成服务实现。
type Seedance15ProService struct {
	*kieAIVideoTaskService
}

// Seedance2Service KIE Seedance 2 视频生成服务实现。
type Seedance2Service struct {
	*kieAIVideoTaskService
}

// Seedance2FastService KIE Seedance 2 Fast 视频生成服务实现。
type Seedance2FastService struct {
	*kieAIVideoTaskService
}

func NewSeedance15ProService() aivideo.AIVideoService {
	return &Seedance15ProService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance15Pro,
			seedance15ProModelName,
			"Seedance 1.5 Pro",
			NewClient(),
			buildSeedance15ProInput,
		),
	}
}

func NewSeedance15ProServiceWithKey(apiKey string) aivideo.AIVideoService {
	return &Seedance15ProService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance15Pro,
			seedance15ProModelName,
			"Seedance 1.5 Pro",
			NewClientWithKey(apiKey),
			buildSeedance15ProInput,
		),
	}
}

func NewSeedance2Service() aivideo.AIVideoService {
	return &Seedance2Service{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance2,
			seedance2ModelName,
			"Seedance 2",
			NewClient(),
			buildSeedance2Input,
		),
	}
}

func NewSeedance2ServiceWithKey(apiKey string) aivideo.AIVideoService {
	return &Seedance2Service{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance2,
			seedance2ModelName,
			"Seedance 2",
			NewClientWithKey(apiKey),
			buildSeedance2Input,
		),
	}
}

func NewSeedance2FastService() aivideo.AIVideoService {
	return &Seedance2FastService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance2Fast,
			seedance2FastModelName,
			"Seedance 2 Fast",
			NewClient(),
			buildSeedance2Input,
		),
	}
}

func NewSeedance2FastServiceWithKey(apiKey string) aivideo.AIVideoService {
	return &Seedance2FastService{
		kieAIVideoTaskService: newKieAIVideoTaskService(
			aivideo.SourceKieSeedance2Fast,
			seedance2FastModelName,
			"Seedance 2 Fast",
			NewClientWithKey(apiKey),
			buildSeedance2Input,
		),
	}
}

type Seedance2Input struct {
	Prompt             string   `json:"prompt,omitempty"`
	ReferenceImageURLs []string `json:"reference_image_urls,omitempty"`
	ReferenceVideoURLs []string `json:"reference_video_urls,omitempty"`
	ReferenceAudioURLs []string `json:"reference_audio_urls,omitempty"`
	ReturnLastFrame    bool     `json:"return_last_frame"`
	GenerateAudio      bool     `json:"generate_audio"`
	Resolution         string   `json:"resolution,omitempty"`
	AspectRatio        string   `json:"aspect_ratio,omitempty"`
	Duration           int      `json:"duration,omitempty"`
	WebSearch          bool     `json:"web_search"`
	NSFWChecker        bool     `json:"nsfw_checker"`
}

type Seedance15ProInput struct {
	Prompt        string   `json:"prompt,omitempty"`
	InputURLs     []string `json:"input_urls,omitempty"`
	AspectRatio   string   `json:"aspect_ratio,omitempty"`
	Resolution    string   `json:"resolution,omitempty"`
	Duration      string   `json:"duration,omitempty"`
	FixedLens     bool     `json:"fixed_lens"`
	GenerateAudio bool     `json:"generate_audio"`
	NSFWChecker   bool     `json:"nsfw_checker"`
}

func buildSeedance2Input(req *aivideo.AIVideoTaskRunReq) interface{} {
	return &Seedance2Input{
		Prompt:             req.Prompt,
		ReferenceImageURLs: resolvePrimaryImageInputs(req),
		ReferenceVideoURLs: req.ReferenceVideoURLs,
		ReferenceAudioURLs: req.ReferenceAudioURLs,
		ReturnLastFrame:    boolValueOrDefault(req.ReturnLastFrame, false),
		GenerateAudio:      boolValueOrDefault(req.GenerateAudio, true),
		Resolution:         seedanceResolution(req),
		AspectRatio:        stringValueOrDefault(req.AspectRatio, "16:9"),
		Duration:           intValueOrDefault(req.Duration, 15),
		WebSearch:          boolValueOrDefault(req.WebSearch, false),
		NSFWChecker:        boolValueOrDefault(req.NSFWChecker, true),
	}
}

func buildSeedance15ProInput(req *aivideo.AIVideoTaskRunReq) interface{} {
	return &Seedance15ProInput{
		Prompt:        req.Prompt,
		InputURLs:     resolvePrimaryImageInputs(req),
		AspectRatio:   stringValueOrDefault(req.AspectRatio, "16:9"),
		Resolution:    seedanceResolution(req),
		Duration:      seedance15ProDuration(req.Duration),
		FixedLens:     boolValueOrDefault(req.FixedLens, false),
		GenerateAudio: boolValueOrDefault(req.GenerateAudio, true),
		NSFWChecker:   boolValueOrDefault(req.NSFWChecker, true),
	}
}

func seedanceResolution(req *aivideo.AIVideoTaskRunReq) string {
	if req.Resolution != "" {
		return req.Resolution
	}
	if req.Quality != "" {
		return req.Quality
	}
	return "720p"
}

func seedance15ProDuration(duration int) string {
	if duration == 0 {
		return "8"
	}
	return strconv.Itoa(duration)
}
