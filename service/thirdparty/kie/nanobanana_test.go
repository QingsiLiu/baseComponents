package kie

import (
	"testing"

	"github.com/QingsiLiu/baseComponents/service/image2image"
)

func boolPtr(v bool) *bool {
	return &v
}

func TestNanoBananaService_convertToCreateRequest_DefaultAndExplicitModels(t *testing.T) {
	service := &NanoBananaService{}

	t.Run("default_with_image_uses_edit_model", func(t *testing.T) {
		req := &image2image.Image2ImageTaskRunReq{
			Prompt:      "keep style",
			ImageInputs: []string{"https://example.com/a.png"},
		}

		payload := service.convertToCreateRequest(req)
		if payload == nil {
			t.Fatal("payload should not be nil")
		}
		if payload.Model != nanoBananaEditModelName {
			t.Fatalf("expected model %s, got %s", nanoBananaEditModelName, payload.Model)
		}

		input, ok := payload.Input.(*NanoBananaInput)
		if !ok {
			t.Fatalf("expected *NanoBananaInput, got %T", payload.Input)
		}

		if input.OutputFormat != "png" {
			t.Fatalf("expected default output format png, got %s", input.OutputFormat)
		}
		if len(input.ImageURLs) != 1 || input.ImageURLs[0] != "https://example.com/a.png" {
			t.Fatalf("unexpected image urls: %#v", input.ImageURLs)
		}
	})

	t.Run("explicit_base_model_still_uses_base_input", func(t *testing.T) {
		req := &image2image.Image2ImageTaskRunReq{
			Model:        nanoBananaModelName,
			Prompt:       "draw cat",
			OutputFormat: "jpeg",
		}

		payload := service.convertToCreateRequest(req)
		if payload == nil {
			t.Fatal("payload should not be nil")
		}
		if payload.Model != nanoBananaModelName {
			t.Fatalf("expected model %s, got %s", nanoBananaModelName, payload.Model)
		}

		input, ok := payload.Input.(*NanoBananaInput)
		if !ok {
			t.Fatalf("expected *NanoBananaInput, got %T", payload.Input)
		}
		if input.OutputFormat != "jpeg" {
			t.Fatalf("expected output format jpeg, got %s", input.OutputFormat)
		}
	})
}

func TestNanoBananaService_convertToCreateRequest_NanoBananaPro(t *testing.T) {
	service := &NanoBananaService{}
	req := &image2image.Image2ImageTaskRunReq{
		Model:           nanoBananaProModelName,
		Prompt:          "make it cinematic",
		OutputImageSize: "16:9",
	}

	payload := service.convertToCreateRequest(req)
	if payload == nil {
		t.Fatal("payload should not be nil")
	}
	if payload.Model != nanoBananaProModelName {
		t.Fatalf("expected model %s, got %s", nanoBananaProModelName, payload.Model)
	}

	input, ok := payload.Input.(*NanoBananaProInput)
	if !ok {
		t.Fatalf("expected *NanoBananaProInput, got %T", payload.Input)
	}
	if input.AspectRatio != "16:9" {
		t.Fatalf("expected aspect ratio 16:9, got %s", input.AspectRatio)
	}
	if input.Resolution != "1K" {
		t.Fatalf("expected default resolution 1K, got %s", input.Resolution)
	}
	if input.OutputFormat != "png" {
		t.Fatalf("expected default output format png, got %s", input.OutputFormat)
	}
}

func TestNanoBananaService_convertToCreateRequest_NanoBanana2(t *testing.T) {
	service := &NanoBananaService{}

	t.Run("defaults", func(t *testing.T) {
		req := &image2image.Image2ImageTaskRunReq{
			Model:  nanoBanana2ModelName,
			Prompt: "translate text",
			ImageInputs: []string{
				"https://example.com/in.png",
			},
		}

		payload := service.convertToCreateRequest(req)
		if payload == nil {
			t.Fatal("payload should not be nil")
		}
		if payload.Model != nanoBanana2ModelName {
			t.Fatalf("expected model %s, got %s", nanoBanana2ModelName, payload.Model)
		}

		input, ok := payload.Input.(*NanoBanana2Input)
		if !ok {
			t.Fatalf("expected *NanoBanana2Input, got %T", payload.Input)
		}

		if input.AspectRatio != "auto" {
			t.Fatalf("expected default aspect ratio auto, got %s", input.AspectRatio)
		}
		if input.Resolution != "1K" {
			t.Fatalf("expected default resolution 1K, got %s", input.Resolution)
		}
		if input.OutputFormat != "jpg" {
			t.Fatalf("expected default output format jpg, got %s", input.OutputFormat)
		}
		if input.GoogleSearch {
			t.Fatal("expected default google_search false")
		}
	})

	t.Run("with_google_search_true", func(t *testing.T) {
		req := &image2image.Image2ImageTaskRunReq{
			Model:        nanoBanana2ModelName,
			Prompt:       "translate text",
			ImageInputs:  []string{"https://example.com/in.png"},
			OutputFormat: "png",
			Resolution:   "2K",
			GoogleSearch: boolPtr(true),
		}

		payload := service.convertToCreateRequest(req)
		if payload == nil {
			t.Fatal("payload should not be nil")
		}

		input, ok := payload.Input.(*NanoBanana2Input)
		if !ok {
			t.Fatalf("expected *NanoBanana2Input, got %T", payload.Input)
		}
		if !input.GoogleSearch {
			t.Fatal("expected google_search true")
		}
		if input.Resolution != "2K" {
			t.Fatalf("expected resolution 2K, got %s", input.Resolution)
		}
		if input.OutputFormat != "png" {
			t.Fatalf("expected output format png, got %s", input.OutputFormat)
		}
	})
}

func TestNanoBananaService_TaskRun(t *testing.T) {
	service := NewNanoBananaServiceWithKey("")
	req := &image2image.Image2ImageTaskRunReq{
		Prompt: "Make the sheets in the style of the logo. Make the scene natural.",
		ImageInputs: []string{
			"https://replicate.delivery/pbxt/NbYIclp4A5HWLsJ8lF5KgiYSNaLBBT1jUcYcHYQmN1uy5OnN/tmpcqc07f_q.png",
			"https://replicate.delivery/pbxt/NbYId45yH8s04sptdtPcGqFIhV7zS5GTcdS3TtNliyTAoYPO/Screenshot%202025-08-26%20at%205.30.12%E2%80%AFPM.png",
		},
		OutputFormat: "jpeg",
	}

	taskId, err := service.TaskRun(req)
	if err != nil {
		t.Fatalf("TaskRun failed: %v", err)
	}

	if taskId == "" {
		t.Fatal("TaskRun returned empty task ID")
	}

	t.Logf("Task submitted successfully with ID: %s", taskId)

	// 测试获取任务状态
	task, err := service.TaskGet(taskId)
	if err != nil {
		t.Fatalf("TaskGet failed: %v", err)
	}

	if task.TaskId != taskId {
		t.Errorf("Expected task ID %s, got %s", taskId, task.TaskId)
	}

	// 检查任务状态是否有效
	validStatuses := []int32{
		image2image.TaskStatusPending,
		image2image.TaskStatusRunning,
		image2image.TaskStatusCompleted,
		image2image.TaskStatusFailed,
	}

	statusValid := false
	for _, status := range validStatuses {
		if task.Status == status {
			statusValid = true
			break
		}
	}

	if !statusValid {
		t.Errorf("Invalid task status: %d", task.Status)
	}

	t.Logf("Task status: %d", task.Status)
	t.Logf("Task: %v", task)
}
