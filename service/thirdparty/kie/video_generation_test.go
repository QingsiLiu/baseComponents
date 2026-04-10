package kie

import (
	"encoding/json"
	"testing"

	"github.com/QingsiLiu/baseComponents/service/aivideo"
)

type kieAIVideoServiceCase struct {
	name           string
	service        *kieAIVideoTaskService
	expectedModel  string
	expectedSource string
}

func TestKling30VideoDefaultsAndExplicitMode(t *testing.T) {
	service := newKieAIVideoTaskService(
		aivideo.SourceKieKling30Video,
		kling30VideoModelName,
		"Kling 3.0 Video",
		NewClientWithKey(""),
		buildKling30VideoInput,
	)

	t.Run("defaults", func(t *testing.T) {
		payload := service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{})
		if payload.Model != kling30VideoModelName {
			t.Fatalf("expected default model %s, got %s", kling30VideoModelName, payload.Model)
		}

		input, ok := payload.Input.(*Kling30VideoInput)
		if !ok {
			t.Fatalf("expected *Kling30VideoInput, got %T", payload.Input)
		}
		if input.Mode != "std" {
			t.Fatalf("expected default mode std, got %s", input.Mode)
		}
	})

	t.Run("explicit_mode_and_callback", func(t *testing.T) {
		payload := service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
			Mode:        "pro",
			CallbackURL: "https://example.com/callback",
			Prompt:      "ignored",
			Resolution:  "720p",
		})
		if payload.CallbackURL != "https://example.com/callback" {
			t.Fatalf("expected callback url to be preserved, got %s", payload.CallbackURL)
		}

		input, ok := payload.Input.(*Kling30VideoInput)
		if !ok {
			t.Fatalf("expected *Kling30VideoInput, got %T", payload.Input)
		}
		if input.Mode != "pro" {
			t.Fatalf("expected mode pro, got %s", input.Mode)
		}
	})
}

func TestKling26ImageToVideoDefaultsAndExplicitFields(t *testing.T) {
	service := newKieAIVideoTaskService(
		aivideo.SourceKieKling26ImageToVideo,
		kling26ImageToVideoModelName,
		"Kling 2.6 Image To Video",
		NewClientWithKey(""),
		buildKling26ImageToVideoInput,
	)

	t.Run("defaults", func(t *testing.T) {
		payload := service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
			Prompt: "animate this image",
			Image:  "https://example.com/a.png",
		})
		if payload.Model != kling26ImageToVideoModelName {
			t.Fatalf("expected default model %s, got %s", kling26ImageToVideoModelName, payload.Model)
		}

		input, ok := payload.Input.(*Kling26ImageToVideoInput)
		if !ok {
			t.Fatalf("expected *Kling26ImageToVideoInput, got %T", payload.Input)
		}
		if input.Prompt != "animate this image" {
			t.Fatalf("expected prompt to be preserved, got %s", input.Prompt)
		}
		if len(input.ImageURLs) != 1 || input.ImageURLs[0] != "https://example.com/a.png" {
			t.Fatalf("unexpected image_urls: %#v", input.ImageURLs)
		}
		if input.Sound {
			t.Fatal("expected default sound false")
		}
		if input.Duration != "5" {
			t.Fatalf("expected default duration 5, got %s", input.Duration)
		}
	})

	t.Run("explicit_sound_duration_and_callback", func(t *testing.T) {
		payload := service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
			Prompt:             "animate this image",
			ReferenceImageURLs: []string{"https://example.com/a.png", "https://example.com/b.png"},
			GenerateAudio:      boolPtr(true),
			Duration:           10,
			CallbackURL:        "https://example.com/callback",
		})
		if payload.CallbackURL != "https://example.com/callback" {
			t.Fatalf("expected callback url to be preserved, got %s", payload.CallbackURL)
		}

		input, ok := payload.Input.(*Kling26ImageToVideoInput)
		if !ok {
			t.Fatalf("expected *Kling26ImageToVideoInput, got %T", payload.Input)
		}
		if len(input.ImageURLs) != 2 || input.ImageURLs[0] != "https://example.com/a.png" || input.ImageURLs[1] != "https://example.com/b.png" {
			t.Fatalf("unexpected image_urls: %#v", input.ImageURLs)
		}
		if !input.Sound {
			t.Fatal("expected sound true")
		}
		if input.Duration != "10" {
			t.Fatalf("expected duration 10, got %s", input.Duration)
		}
	})
}

func TestKling26TextToVideoDefaultsAndExplicitFields(t *testing.T) {
	service := newKieAIVideoTaskService(
		aivideo.SourceKieKling26TextToVideo,
		kling26TextToVideoModelName,
		"Kling 2.6 Text To Video",
		NewClientWithKey(""),
		buildKling26TextToVideoInput,
	)

	t.Run("defaults", func(t *testing.T) {
		payload := service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
			Prompt: "generate a live commerce clip",
		})
		if payload.Model != kling26TextToVideoModelName {
			t.Fatalf("expected default model %s, got %s", kling26TextToVideoModelName, payload.Model)
		}

		input, ok := payload.Input.(*Kling26TextToVideoInput)
		if !ok {
			t.Fatalf("expected *Kling26TextToVideoInput, got %T", payload.Input)
		}
		if input.Prompt != "generate a live commerce clip" {
			t.Fatalf("expected prompt to be preserved, got %s", input.Prompt)
		}
		if input.Sound {
			t.Fatal("expected default sound false")
		}
		if input.AspectRatio != "1:1" {
			t.Fatalf("expected default aspect_ratio 1:1, got %s", input.AspectRatio)
		}
		if input.Duration != "5" {
			t.Fatalf("expected default duration 5, got %s", input.Duration)
		}
	})

	t.Run("explicit_sound_aspect_ratio_duration_and_callback", func(t *testing.T) {
		payload := service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
			Prompt:        "generate a live commerce clip",
			GenerateAudio: boolPtr(true),
			AspectRatio:   "16:9",
			Duration:      10,
			CallbackURL:   "https://example.com/callback",
			Image:         "https://example.com/ignored.png",
		})
		if payload.CallbackURL != "https://example.com/callback" {
			t.Fatalf("expected callback url to be preserved, got %s", payload.CallbackURL)
		}

		input, ok := payload.Input.(*Kling26TextToVideoInput)
		if !ok {
			t.Fatalf("expected *Kling26TextToVideoInput, got %T", payload.Input)
		}
		if !input.Sound {
			t.Fatal("expected sound true")
		}
		if input.AspectRatio != "16:9" {
			t.Fatalf("expected aspect_ratio 16:9, got %s", input.AspectRatio)
		}
		if input.Duration != "10" {
			t.Fatalf("expected duration 10, got %s", input.Duration)
		}
	})
}

func TestSeedance2Family_SourceAndDefaults(t *testing.T) {
	cases := []kieAIVideoServiceCase{
		{
			name:           "seedance_2",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance2, seedance2ModelName, "Seedance 2", NewClientWithKey(""), buildSeedance2Input),
			expectedModel:  seedance2ModelName,
			expectedSource: aivideo.SourceKieSeedance2,
		},
		{
			name:           "seedance_2_fast",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance2Fast, seedance2FastModelName, "Seedance 2 Fast", NewClientWithKey(""), buildSeedance2Input),
			expectedModel:  seedance2FastModelName,
			expectedSource: aivideo.SourceKieSeedance2Fast,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.service.Source() != tc.expectedSource {
				t.Fatalf("expected source %s, got %s", tc.expectedSource, tc.service.Source())
			}

			payload := tc.service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
				Prompt: "make a sunset drive",
			})
			if payload.Model != tc.expectedModel {
				t.Fatalf("expected default model %s, got %s", tc.expectedModel, payload.Model)
			}

			input, ok := payload.Input.(*Seedance2Input)
			if !ok {
				t.Fatalf("expected *Seedance2Input, got %T", payload.Input)
			}
			if input.Resolution != "720p" {
				t.Fatalf("expected default resolution 720p, got %s", input.Resolution)
			}
			if input.AspectRatio != "16:9" {
				t.Fatalf("expected default aspect ratio 16:9, got %s", input.AspectRatio)
			}
			if input.Duration != 15 {
				t.Fatalf("expected default duration 15, got %d", input.Duration)
			}
			if input.ReturnLastFrame {
				t.Fatal("expected default return_last_frame false")
			}
			if !input.GenerateAudio {
				t.Fatal("expected default generate_audio true")
			}
			if input.WebSearch {
				t.Fatal("expected default web_search false")
			}
			if !input.NSFWChecker {
				t.Fatal("expected default nsfw_checker true")
			}
		})
	}
}

func TestSeedance2Family_ResolutionPrecedenceAndLegacyImage(t *testing.T) {
	cases := []kieAIVideoServiceCase{
		{
			name:           "seedance_2",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance2, seedance2ModelName, "Seedance 2", NewClientWithKey(""), buildSeedance2Input),
			expectedModel:  seedance2ModelName,
			expectedSource: aivideo.SourceKieSeedance2,
		},
		{
			name:           "seedance_2_fast",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance2Fast, seedance2FastModelName, "Seedance 2 Fast", NewClientWithKey(""), buildSeedance2Input),
			expectedModel:  seedance2FastModelName,
			expectedSource: aivideo.SourceKieSeedance2Fast,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := tc.service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
				Prompt:     "animate this frame",
				Resolution: "720p",
				Quality:    "480p",
				Image:      "https://example.com/frame.png",
			})

			input, ok := payload.Input.(*Seedance2Input)
			if !ok {
				t.Fatalf("expected *Seedance2Input, got %T", payload.Input)
			}
			if input.Resolution != "720p" {
				t.Fatalf("expected resolution from Resolution field, got %s", input.Resolution)
			}
			if len(input.ReferenceImageURLs) != 1 || input.ReferenceImageURLs[0] != "https://example.com/frame.png" {
				t.Fatalf("expected legacy image to populate reference_image_urls, got %#v", input.ReferenceImageURLs)
			}
		})
	}
}

func TestSeedance2Family_ExplicitFields(t *testing.T) {
	cases := []kieAIVideoServiceCase{
		{
			name:           "seedance_2",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance2, seedance2ModelName, "Seedance 2", NewClientWithKey(""), buildSeedance2Input),
			expectedModel:  seedance2ModelName,
			expectedSource: aivideo.SourceKieSeedance2,
		},
		{
			name:           "seedance_2_fast",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance2Fast, seedance2FastModelName, "Seedance 2 Fast", NewClientWithKey(""), buildSeedance2Input),
			expectedModel:  seedance2FastModelName,
			expectedSource: aivideo.SourceKieSeedance2Fast,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &aivideo.AIVideoTaskRunReq{
				Model:              tc.expectedModel,
				Prompt:             "add synchronized narration",
				AspectRatio:        "9:16",
				Duration:           12,
				CallbackURL:        "https://example.com/callback",
				ReferenceImageURLs: []string{"https://example.com/a.png"},
				ReferenceVideoURLs: []string{"https://example.com/a.mp4"},
				ReferenceAudioURLs: []string{"https://example.com/a.wav"},
				ReturnLastFrame:    boolPtr(true),
				GenerateAudio:      boolPtr(false),
				WebSearch:          boolPtr(true),
				NSFWChecker:        boolPtr(false),
			}

			payload := tc.service.convertToCreateRequest(req)
			if payload.Model != tc.expectedModel {
				t.Fatalf("expected model %s, got %s", tc.expectedModel, payload.Model)
			}
			if payload.CallbackURL != req.CallbackURL {
				t.Fatalf("expected callback url %s, got %s", req.CallbackURL, payload.CallbackURL)
			}

			input, ok := payload.Input.(*Seedance2Input)
			if !ok {
				t.Fatalf("expected *Seedance2Input, got %T", payload.Input)
			}
			if len(input.ReferenceImageURLs) != 1 || input.ReferenceImageURLs[0] != req.ReferenceImageURLs[0] {
				t.Fatalf("unexpected reference_image_urls: %#v", input.ReferenceImageURLs)
			}
			if len(input.ReferenceVideoURLs) != 1 || input.ReferenceVideoURLs[0] != req.ReferenceVideoURLs[0] {
				t.Fatalf("unexpected reference_video_urls: %#v", input.ReferenceVideoURLs)
			}
			if len(input.ReferenceAudioURLs) != 1 || input.ReferenceAudioURLs[0] != req.ReferenceAudioURLs[0] {
				t.Fatalf("unexpected reference_audio_urls: %#v", input.ReferenceAudioURLs)
			}
			if !input.ReturnLastFrame {
				t.Fatal("expected return_last_frame true")
			}
			if input.GenerateAudio {
				t.Fatal("expected generate_audio false")
			}
			if !input.WebSearch {
				t.Fatal("expected web_search true")
			}
			if input.NSFWChecker {
				t.Fatal("expected nsfw_checker false")
			}
		})
	}
}

func TestSeedance15ProDefaultsAndMapping(t *testing.T) {
	service := newKieAIVideoTaskService(
		aivideo.SourceKieSeedance15Pro,
		seedance15ProModelName,
		"Seedance 1.5 Pro",
		NewClientWithKey(""),
		buildSeedance15ProInput,
	)

	payload := service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
		Prompt: "cinematic tavern scene",
	})
	if payload.Model != seedance15ProModelName {
		t.Fatalf("expected default model %s, got %s", seedance15ProModelName, payload.Model)
	}
	if payload.CallbackURL != "" {
		t.Fatalf("expected empty callback url, got %s", payload.CallbackURL)
	}

	input, ok := payload.Input.(*Seedance15ProInput)
	if !ok {
		t.Fatalf("expected *Seedance15ProInput, got %T", payload.Input)
	}
	if input.AspectRatio != "16:9" {
		t.Fatalf("expected default aspect ratio 16:9, got %s", input.AspectRatio)
	}
	if input.Resolution != "720p" {
		t.Fatalf("expected default resolution 720p, got %s", input.Resolution)
	}
	if input.Duration != "8" {
		t.Fatalf("expected default duration 8, got %s", input.Duration)
	}
	if input.FixedLens {
		t.Fatal("expected default fixed_lens false")
	}
	if !input.GenerateAudio {
		t.Fatal("expected default generate_audio true")
	}
	if !input.NSFWChecker {
		t.Fatal("expected default nsfw_checker true")
	}
}

func TestSeedance15ProExplicitFields(t *testing.T) {
	service := newKieAIVideoTaskService(
		aivideo.SourceKieSeedance15Pro,
		seedance15ProModelName,
		"Seedance 1.5 Pro",
		NewClientWithKey(""),
		buildSeedance15ProInput,
	)

	req := &aivideo.AIVideoTaskRunReq{
		Model:              seedance15ProModelName,
		Prompt:             "cinematic tavern scene",
		AspectRatio:        "9:16",
		Resolution:         "1080p",
		Duration:           12,
		Image:              "https://example.com/legacy.png",
		CallbackURL:        "https://example.com/callback",
		FixedLens:          boolPtr(true),
		GenerateAudio:      boolPtr(false),
		NSFWChecker:        boolPtr(false),
		ReferenceVideoURLs: []string{"https://example.com/ignored.mp4"},
		ReferenceAudioURLs: []string{"https://example.com/ignored.wav"},
		WebSearch:          boolPtr(true),
		ReturnLastFrame:    boolPtr(true),
	}

	payload := service.convertToCreateRequest(req)
	if payload.Model != seedance15ProModelName {
		t.Fatalf("expected model %s, got %s", seedance15ProModelName, payload.Model)
	}
	if payload.CallbackURL != req.CallbackURL {
		t.Fatalf("expected callback url %s, got %s", req.CallbackURL, payload.CallbackURL)
	}

	input, ok := payload.Input.(*Seedance15ProInput)
	if !ok {
		t.Fatalf("expected *Seedance15ProInput, got %T", payload.Input)
	}
	if len(input.InputURLs) != 1 || input.InputURLs[0] != req.Image {
		t.Fatalf("expected legacy image to populate input_urls, got %#v", input.InputURLs)
	}
	if input.AspectRatio != "9:16" {
		t.Fatalf("expected aspect ratio 9:16, got %s", input.AspectRatio)
	}
	if input.Resolution != "1080p" {
		t.Fatalf("expected resolution 1080p, got %s", input.Resolution)
	}
	if input.Duration != "12" {
		t.Fatalf("expected duration 12, got %s", input.Duration)
	}
	if !input.FixedLens {
		t.Fatal("expected fixed_lens true")
	}
	if input.GenerateAudio {
		t.Fatal("expected generate_audio false")
	}
	if input.NSFWChecker {
		t.Fatal("expected nsfw_checker false")
	}
}

func TestSeedance15ProReferenceImagesOverrideLegacyImage(t *testing.T) {
	service := newKieAIVideoTaskService(
		aivideo.SourceKieSeedance15Pro,
		seedance15ProModelName,
		"Seedance 1.5 Pro",
		NewClientWithKey(""),
		buildSeedance15ProInput,
	)

	payload := service.convertToCreateRequest(&aivideo.AIVideoTaskRunReq{
		Prompt:             "cinematic tavern scene",
		Image:              "https://example.com/legacy.png",
		ReferenceImageURLs: []string{"https://example.com/a.png", "https://example.com/b.png"},
	})

	input, ok := payload.Input.(*Seedance15ProInput)
	if !ok {
		t.Fatalf("expected *Seedance15ProInput, got %T", payload.Input)
	}
	if len(input.InputURLs) != 2 || input.InputURLs[0] != "https://example.com/a.png" || input.InputURLs[1] != "https://example.com/b.png" {
		t.Fatalf("unexpected input_urls: %#v", input.InputURLs)
	}
}

func TestKieAIVideoServices_convertToTaskInfo_UsesCostTimeAndParsesResult(t *testing.T) {
	cases := []kieAIVideoServiceCase{
		{
			name:           "kling_2_6_image_to_video",
			service:        newKieAIVideoTaskService(aivideo.SourceKieKling26ImageToVideo, kling26ImageToVideoModelName, "Kling 2.6 Image To Video", NewClientWithKey(""), buildKling26ImageToVideoInput),
			expectedModel:  kling26ImageToVideoModelName,
			expectedSource: aivideo.SourceKieKling26ImageToVideo,
		},
		{
			name:           "kling_2_6_text_to_video",
			service:        newKieAIVideoTaskService(aivideo.SourceKieKling26TextToVideo, kling26TextToVideoModelName, "Kling 2.6 Text To Video", NewClientWithKey(""), buildKling26TextToVideoInput),
			expectedModel:  kling26TextToVideoModelName,
			expectedSource: aivideo.SourceKieKling26TextToVideo,
		},
		{
			name:           "kling_3_0_video",
			service:        newKieAIVideoTaskService(aivideo.SourceKieKling30Video, kling30VideoModelName, "Kling 3.0 Video", NewClientWithKey(""), buildKling30VideoInput),
			expectedModel:  kling30VideoModelName,
			expectedSource: aivideo.SourceKieKling30Video,
		},
		{
			name:           "seedance_1_5_pro",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance15Pro, seedance15ProModelName, "Seedance 1.5 Pro", NewClientWithKey(""), buildSeedance15ProInput),
			expectedModel:  seedance15ProModelName,
			expectedSource: aivideo.SourceKieSeedance15Pro,
		},
		{
			name:           "seedance_2",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance2, seedance2ModelName, "Seedance 2", NewClientWithKey(""), buildSeedance2Input),
			expectedModel:  seedance2ModelName,
			expectedSource: aivideo.SourceKieSeedance2,
		},
		{
			name:           "seedance_2_fast",
			service:        newKieAIVideoTaskService(aivideo.SourceKieSeedance2Fast, seedance2FastModelName, "Seedance 2 Fast", NewClientWithKey(""), buildSeedance2Input),
			expectedModel:  seedance2FastModelName,
			expectedSource: aivideo.SourceKieSeedance2Fast,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			task := tc.service.convertToTaskInfo(&TaskRecordDetail{
				TaskID:     "task-1",
				State:      TaskStateSuccess,
				ResultJSON: `{"resultUrls":["https://example.com/video.mp4"]}`,
				CostTime:   2500,
				CreateTime: 1757584164490,
			})
			if task.TaskId != "task-1" {
				t.Fatalf("expected task id task-1, got %s", task.TaskId)
			}
			if task.Status != aivideo.TaskStatusCompleted {
				t.Fatalf("expected completed status, got %d", task.Status)
			}
			if len(task.Result) != 1 || task.Result[0] != "https://example.com/video.mp4" {
				t.Fatalf("unexpected result urls: %#v", task.Result)
			}
			if task.Duration != 2.5 {
				t.Fatalf("expected duration 2.5, got %v", task.Duration)
			}
			if task.CreateTime == 0 {
				t.Fatal("expected create time to be populated")
			}
		})
	}
}

func TestNewKieAIVideoServices(t *testing.T) {
	kling26ImageToVideo, ok := NewKling26ImageToVideoService().(*Kling26ImageToVideoService)
	if !ok {
		t.Fatalf("expected *Kling26ImageToVideoService, got %T", NewKling26ImageToVideoService())
	}
	if kling26ImageToVideo.Source() != aivideo.SourceKieKling26ImageToVideo {
		t.Fatalf("expected source %s, got %s", aivideo.SourceKieKling26ImageToVideo, kling26ImageToVideo.Source())
	}

	kling26TextToVideo, ok := NewKling26TextToVideoService().(*Kling26TextToVideoService)
	if !ok {
		t.Fatalf("expected *Kling26TextToVideoService, got %T", NewKling26TextToVideoService())
	}
	if kling26TextToVideo.Source() != aivideo.SourceKieKling26TextToVideo {
		t.Fatalf("expected source %s, got %s", aivideo.SourceKieKling26TextToVideo, kling26TextToVideo.Source())
	}

	kling30Video, ok := NewKling30VideoService().(*Kling30VideoService)
	if !ok {
		t.Fatalf("expected *Kling30VideoService, got %T", NewKling30VideoService())
	}
	if kling30Video.Source() != aivideo.SourceKieKling30Video {
		t.Fatalf("expected source %s, got %s", aivideo.SourceKieKling30Video, kling30Video.Source())
	}

	seedance15Pro, ok := NewSeedance15ProService().(*Seedance15ProService)
	if !ok {
		t.Fatalf("expected *Seedance15ProService, got %T", NewSeedance15ProService())
	}
	if seedance15Pro.Source() != aivideo.SourceKieSeedance15Pro {
		t.Fatalf("expected source %s, got %s", aivideo.SourceKieSeedance15Pro, seedance15Pro.Source())
	}

	seedance2, ok := NewSeedance2Service().(*Seedance2Service)
	if !ok {
		t.Fatalf("expected *Seedance2Service, got %T", NewSeedance2Service())
	}
	if seedance2.Source() != aivideo.SourceKieSeedance2 {
		t.Fatalf("expected source %s, got %s", aivideo.SourceKieSeedance2, seedance2.Source())
	}

	seedance2Fast, ok := NewSeedance2FastService().(*Seedance2FastService)
	if !ok {
		t.Fatalf("expected *Seedance2FastService, got %T", NewSeedance2FastService())
	}
	if seedance2Fast.Source() != aivideo.SourceKieSeedance2Fast {
		t.Fatalf("expected source %s, got %s", aivideo.SourceKieSeedance2Fast, seedance2Fast.Source())
	}
}

func TestTaskResponseMessageCompatibility(t *testing.T) {
	var createRespWithMsg TaskCreateResponse
	if err := json.Unmarshal([]byte(`{"code":200,"msg":"success","data":{"taskId":"abc"}}`), &createRespWithMsg); err != nil {
		t.Fatalf("unmarshal create response with msg failed: %v", err)
	}
	if createRespWithMsg.GetMessage() != "success" {
		t.Fatalf("expected message success, got %s", createRespWithMsg.GetMessage())
	}

	var createRespWithMessage TaskCreateResponse
	if err := json.Unmarshal([]byte(`{"code":200,"message":"created","data":{"taskId":"abc"}}`), &createRespWithMessage); err != nil {
		t.Fatalf("unmarshal create response with message failed: %v", err)
	}
	if createRespWithMessage.GetMessage() != "created" {
		t.Fatalf("expected message created, got %s", createRespWithMessage.GetMessage())
	}

	var recordRespWithMessage TaskRecordResponse
	if err := json.Unmarshal([]byte(`{"code":200,"message":"ok","data":{"taskId":"abc"}}`), &recordRespWithMessage); err != nil {
		t.Fatalf("unmarshal record response with message failed: %v", err)
	}
	if recordRespWithMessage.GetMessage() != "ok" {
		t.Fatalf("expected message ok, got %s", recordRespWithMessage.GetMessage())
	}

	var recordRespWithMsg TaskRecordResponse
	if err := json.Unmarshal([]byte(`{"code":200,"msg":"queued","data":{"taskId":"abc"}}`), &recordRespWithMsg); err != nil {
		t.Fatalf("unmarshal record response with msg failed: %v", err)
	}
	if recordRespWithMsg.GetMessage() != "queued" {
		t.Fatalf("expected message queued, got %s", recordRespWithMsg.GetMessage())
	}
}
