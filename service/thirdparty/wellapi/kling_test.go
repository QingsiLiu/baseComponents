package wellapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"
	legacykling "github.com/QingsiLiu/baseComponents/service/kling"
)

func TestKlingMotionControlTaskRunBuildsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PathKlingMotionControl {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %s", got)
		}

		var payload klingMotionControlRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}

		if payload.ImageURL != "https://example.com/image.png" {
			t.Fatalf("unexpected image_url: %s", payload.ImageURL)
		}
		if payload.VideoURL != "https://example.com/video.mp4" {
			t.Fatalf("unexpected video_url: %s", payload.VideoURL)
		}
		if payload.CharacterOrientation != "image" {
			t.Fatalf("unexpected character_orientation: %s", payload.CharacterOrientation)
		}
		if payload.Mode != "std" {
			t.Fatalf("expected default mode std, got %s", payload.Mode)
		}
		if payload.KeepOriginalSound != "no" {
			t.Fatalf("expected keep_original_sound no, got %s", payload.KeepOriginalSound)
		}
		if payload.CallbackURL != "https://example.com/callback" {
			t.Fatalf("unexpected callback_url: %s", payload.CallbackURL)
		}
		if payload.ExternalTaskID != "external-1" {
			t.Fatalf("unexpected external_task_id: %s", payload.ExternalTaskID)
		}

		_ = json.NewEncoder(w).Encode(klingTaskCreateResponse{
			Code:      0,
			Message:   "SUCCEED",
			RequestID: "req-1",
			Data: &klingTaskData{
				TaskID: "task-1",
			},
		})
	}))
	defer server.Close()

	service := &KlingMotionControlService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingMotionControl,
			PathKlingMotionControl,
			PathKlingMotionControlGetFmt,
			"WellAPI Kling Motion Control",
			NewClientWithConfig(Config{APIKey: "test-key", BaseURL: server.URL}),
		),
	}

	keepOriginalSound := false
	taskID, err := service.TaskRun(&aivideokling.KlingMotionControlTaskRunReq{
		Prompt:               "make the person follow the motion",
		ImageURL:             "https://example.com/image.png",
		VideoURL:             "https://example.com/video.mp4",
		KeepOriginalSound:    &keepOriginalSound,
		CharacterOrientation: "image",
		CallbackURL:          "https://example.com/callback",
		ExternalTaskID:       "external-1",
	})
	if err != nil {
		t.Fatalf("TaskRun returned error: %v", err)
	}
	if taskID != "task-1" {
		t.Fatalf("expected task id task-1, got %s", taskID)
	}
}

func TestKlingMotionControlTaskRunValidatesRequiredFields(t *testing.T) {
	service := &KlingMotionControlService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingMotionControl,
			PathKlingMotionControl,
			PathKlingMotionControlGetFmt,
			"WellAPI Kling Motion Control",
			NewClientWithConfig(Config{APIKey: "test-key", BaseURL: "https://example.com"}),
		),
	}

	cases := []struct {
		name string
		req  *aivideokling.KlingMotionControlTaskRunReq
		want string
	}{
		{
			name: "missing_image_url",
			req: &aivideokling.KlingMotionControlTaskRunReq{
				VideoURL:             "https://example.com/video.mp4",
				CharacterOrientation: "image",
			},
			want: "image_url is required",
		},
		{
			name: "missing_video_url",
			req: &aivideokling.KlingMotionControlTaskRunReq{
				ImageURL:             "https://example.com/image.png",
				CharacterOrientation: "image",
			},
			want: "video_url is required",
		},
		{
			name: "missing_character_orientation",
			req: &aivideokling.KlingMotionControlTaskRunReq{
				ImageURL: "https://example.com/image.png",
				VideoURL: "https://example.com/video.mp4",
			},
			want: "character_orientation is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.TaskRun(tc.req)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestKlingMotionControlTaskGetUsesQueryPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kling/v1/videos/motion-control/task-123" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(klingTaskQueryResponse{
			Code:      0,
			Message:   "SUCCEED",
			RequestID: "req-2",
			Data: &klingTaskData{
				TaskID:        "task-123",
				TaskStatus:    "processing",
				TaskStatusMsg: "working",
				CreatedAt:     1757584164490,
				UpdatedAt:     1757584174490,
				TaskResult: &klingTaskResult{
					Videos: []klingTaskVideo{{ID: "video-1", URL: "https://example.com/video.mp4", Duration: "5"}},
				},
			},
		})
	}))
	defer server.Close()

	service := &KlingMotionControlService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingMotionControl,
			PathKlingMotionControl,
			PathKlingMotionControlGetFmt,
			"WellAPI Kling Motion Control",
			NewClientWithConfig(Config{APIKey: "test-key", BaseURL: server.URL}),
		),
	}

	task, err := service.TaskGet("task-123")
	if err != nil {
		t.Fatalf("TaskGet returned error: %v", err)
	}
	if task.TaskID != "task-123" {
		t.Fatalf("unexpected task id: %s", task.TaskID)
	}
	if task.Status != aivideokling.TaskStatusRunning {
		t.Fatalf("expected running status, got %d", task.Status)
	}
	if len(task.Videos) != 1 || task.Videos[0].URL != "https://example.com/video.mp4" {
		t.Fatalf("unexpected videos: %#v", task.Videos)
	}
}

func TestKlingEffectsTaskRunBuildsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PathKlingEffects {
			http.NotFound(w, r)
			return
		}

		var payload klingEffectsRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}

		if payload.EffectScene != "balloon_parade" {
			t.Fatalf("unexpected effect_scene: %s", payload.EffectScene)
		}
		if payload.CallbackURL != "https://example.com/callback" {
			t.Fatalf("unexpected callback_url: %s", payload.CallbackURL)
		}
		if payload.ExternalTaskID != "external-2" {
			t.Fatalf("unexpected external_task_id: %s", payload.ExternalTaskID)
		}
		if payload.Input["model_name"] != "kling-v1-6" {
			t.Fatalf("unexpected model_name: %#v", payload.Input["model_name"])
		}
		if payload.Input["mode"] != "pro" {
			t.Fatalf("unexpected mode: %#v", payload.Input["mode"])
		}
		if payload.Input["duration"] != "5" {
			t.Fatalf("unexpected duration: %#v", payload.Input["duration"])
		}
		if payload.Input["image"] != "https://example.com/image.png" {
			t.Fatalf("unexpected image: %#v", payload.Input["image"])
		}
		if payload.Input["custom_field"] != "custom" {
			t.Fatalf("expected extra custom_field, got %#v", payload.Input["custom_field"])
		}
		if payload.Input["model_name"] == "should-not-override" {
			t.Fatal("extra field should not override typed model_name")
		}

		_ = json.NewEncoder(w).Encode(klingTaskCreateResponse{
			Code:    0,
			Message: "SUCCEED",
			Data: &klingTaskData{
				TaskID: "task-2",
			},
		})
	}))
	defer server.Close()

	service := &KlingEffectsService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingEffects,
			PathKlingEffects,
			PathKlingEffectsGetFmt,
			"WellAPI Kling Effects",
			NewClientWithConfig(Config{APIKey: "test-key", BaseURL: server.URL}),
		),
	}

	taskID, err := service.TaskRun(&aivideokling.KlingEffectsTaskRunReq{
		EffectScene: "balloon_parade",
		Input: aivideokling.KlingEffectsInput{
			ModelName: "kling-v1-6",
			Mode:      "pro",
			Duration:  "5",
			Image:     "https://example.com/image.png",
			Extra: map[string]any{
				"custom_field": "custom",
				"model_name":   "should-not-override",
			},
		},
		CallbackURL:    "https://example.com/callback",
		ExternalTaskID: "external-2",
	})
	if err != nil {
		t.Fatalf("TaskRun returned error: %v", err)
	}
	if taskID != "task-2" {
		t.Fatalf("expected task id task-2, got %s", taskID)
	}
}

func TestKlingEffectsTaskRunValidatesEffectScene(t *testing.T) {
	service := &KlingEffectsService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingEffects,
			PathKlingEffects,
			PathKlingEffectsGetFmt,
			"WellAPI Kling Effects",
			NewClientWithConfig(Config{APIKey: "test-key", BaseURL: "https://example.com"}),
		),
	}

	_, err := service.TaskRun(&aivideokling.KlingEffectsTaskRunReq{})
	if err == nil || !strings.Contains(err.Error(), "effect_scene is required") {
		t.Fatalf("expected effect_scene validation error, got %v", err)
	}
}

func TestKlingEffectsTaskGetUsesQueryPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kling/v1/videos/effects/task-456" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(klingTaskQueryResponse{
			Code:      0,
			Message:   "SUCCEED",
			RequestID: "req-3",
			Data: &klingTaskData{
				TaskID:        "task-456",
				TaskStatus:    "succeed",
				TaskStatusMsg: "done",
				CreatedAt:     1757584164490,
				UpdatedAt:     1757584184490,
				TaskResult: &klingTaskResult{
					Images: []klingTaskImage{{Index: 0, URL: "https://example.com/image.png"}},
					Videos: []klingTaskVideo{{ID: "video-2", URL: "https://example.com/video.mp4", Duration: "5"}},
				},
			},
		})
	}))
	defer server.Close()

	service := &KlingEffectsService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingEffects,
			PathKlingEffects,
			PathKlingEffectsGetFmt,
			"WellAPI Kling Effects",
			NewClientWithConfig(Config{APIKey: "test-key", BaseURL: server.URL}),
		),
	}

	task, err := service.TaskGet("task-456")
	if err != nil {
		t.Fatalf("TaskGet returned error: %v", err)
	}
	if task.Status != aivideokling.TaskStatusCompleted {
		t.Fatalf("expected completed status, got %d", task.Status)
	}
	if len(task.Images) != 1 || task.Images[0].URL != "https://example.com/image.png" {
		t.Fatalf("unexpected images: %#v", task.Images)
	}
}

func TestConvertKlingTaskInfoStatusMapping(t *testing.T) {
	cases := []struct {
		name   string
		status string
		want   int32
	}{
		{name: "submitted", status: "submitted", want: aivideokling.TaskStatusPending},
		{name: "processing", status: "processing", want: aivideokling.TaskStatusRunning},
		{name: "succeed", status: "succeed", want: aivideokling.TaskStatusCompleted},
		{name: "failed", status: "failed", want: aivideokling.TaskStatusFailed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			info := convertKlingTaskInfo(&klingTaskQueryResponse{
				Raw: []byte(`{"ok":true}`),
				Data: &klingTaskData{
					TaskID:        "task-status",
					TaskStatus:    tc.status,
					TaskStatusMsg: tc.status,
					CreatedAt:     1757584164490,
					UpdatedAt:     1757584174490,
					TaskResult: &klingTaskResult{
						Videos: []klingTaskVideo{{ID: "video-1", URL: "https://example.com/video.mp4", Duration: "5"}},
						Images: []klingTaskImage{{Index: 0, URL: "https://example.com/image.png"}},
					},
				},
			})

			if info.Status != tc.want {
				t.Fatalf("expected status %d, got %d", tc.want, info.Status)
			}
			if len(info.Videos) != 1 || len(info.Images) != 1 {
				t.Fatalf("expected one video and one image, got videos=%d images=%d", len(info.Videos), len(info.Images))
			}
			if len(info.Raw) == 0 {
				t.Fatal("expected raw response to be preserved")
			}
		})
	}
}

func TestLegacyKlingCompatibilityLayer(t *testing.T) {
	if legacykling.SourceWellAPIKlingMotionControl != aivideokling.SourceWellAPIKlingMotionControl {
		t.Fatal("expected legacy motion-control source to match new package")
	}
	if legacykling.ServiceTypeWellAPIKlingEffects != aivideokling.ServiceTypeWellAPIKlingEffects {
		t.Fatal("expected legacy effects service type to match new package")
	}
	if legacykling.ConvertTaskStatus("processing") != aivideokling.ConvertTaskStatus("processing") {
		t.Fatal("expected legacy status conversion to match new package")
	}
	if !legacykling.IsValidSource(legacykling.SourceWellAPIKlingEffects) {
		t.Fatal("expected legacy source validation to remain compatible")
	}
	if got := legacykling.GetServiceType(legacykling.SourceWellAPIKlingMotionControl); got != aivideokling.ServiceTypeWellAPIKlingMotionControl {
		t.Fatalf("unexpected legacy type lookup result: %s", got)
	}
}
