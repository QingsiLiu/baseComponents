package kie

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QingsiLiu/baseComponents/service/text2image"
)

func TestGPTImage2Text2ImageService_convertToCreateRequest(t *testing.T) {
	service := &GPTImage2Text2ImageService{}

	t.Run("defaults", func(t *testing.T) {
		payload := service.convertToCreateRequest(&text2image.Text2ImageTaskRunReq{
			Prompt: "A cinematic night city poster.",
		})

		if payload.Model != gptImage2Text2ImageModelName {
			t.Fatalf("expected model %s, got %s", gptImage2Text2ImageModelName, payload.Model)
		}

		input, ok := payload.Input.(*GPTImage2Text2ImageInput)
		if !ok {
			t.Fatalf("expected *GPTImage2Text2ImageInput, got %T", payload.Input)
		}
		if input.Prompt != "A cinematic night city poster." {
			t.Fatalf("unexpected prompt: %s", input.Prompt)
		}
		if !input.NSFWChecker {
			t.Fatal("expected nsfw_checker to default to true")
		}
	})

	t.Run("disable_safety_checker_maps_to_false", func(t *testing.T) {
		payload := service.convertToCreateRequest(&text2image.Text2ImageTaskRunReq{
			Model:                "custom-model",
			Prompt:               "A watercolor cat.",
			DisableSafetyChecker: true,
		})

		if payload.Model != "custom-model" {
			t.Fatalf("expected explicit model custom-model, got %s", payload.Model)
		}

		input := payload.Input.(*GPTImage2Text2ImageInput)
		if input.NSFWChecker {
			t.Fatal("expected nsfw_checker to be false when safety checker is disabled")
		}
	})
}

func TestGPTImage2Text2ImageService_TaskGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != RecordInfoEndpoint {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("taskId"); got != "task_gpt_image_2" {
			t.Fatalf("unexpected taskId query: %s", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %s", got)
		}

		_ = json.NewEncoder(w).Encode(TaskRecordResponse{
			Code: 200,
			Msg:  "success",
			Data: &TaskRecordDetail{
				TaskID:     "task_gpt_image_2",
				Model:      gptImage2Text2ImageModelName,
				State:      TaskStateSuccess,
				ResultJSON: `{"resultUrls":["https://example.com/image-1.png"],"resultUrl":"https://example.com/image-1.png","urls":["https://example.com/image-2.png"]}`,
				CostTime:   2500,
				CreateTime: 1710000000000,
				UpdateTime: 1710000004000,
			},
		})
	}))
	defer server.Close()

	service := &GPTImage2Text2ImageService{
		client: &Client{
			httpClient: server.Client(),
			apiKey:     "test-key",
			timeout:    DefaultTimeout,
			baseURL:    server.URL,
		},
	}

	task, err := service.TaskGet("task_gpt_image_2")
	if err != nil {
		t.Fatalf("TaskGet returned error: %v", err)
	}

	if task.TaskId != "task_gpt_image_2" {
		t.Fatalf("unexpected task id: %s", task.TaskId)
	}
	if task.Status != text2image.TaskStatusCompleted {
		t.Fatalf("expected completed status, got %d", task.Status)
	}
	if len(task.Result) != 2 {
		t.Fatalf("expected 2 unique result urls, got %#v", task.Result)
	}
	if task.Result[0] != "https://example.com/image-1.png" || task.Result[1] != "https://example.com/image-2.png" {
		t.Fatalf("unexpected result urls: %#v", task.Result)
	}
	if task.CreateTime == 0 || task.UpdateTime == 0 {
		t.Fatalf("expected non-zero timestamps, got create=%d update=%d", task.CreateTime, task.UpdateTime)
	}
	if task.Duration != 2.5 {
		t.Fatalf("expected duration 2.5, got %v", task.Duration)
	}
}
