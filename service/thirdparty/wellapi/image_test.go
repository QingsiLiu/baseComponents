package wellapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImageServiceGenerateDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PathImagesGenerations {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %s", got)
		}

		var payload ImageGenerateReq
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}

		if payload.Model != ModelGPTImage2 {
			t.Fatalf("expected default model %s, got %s", ModelGPTImage2, payload.Model)
		}
		if payload.N != 1 {
			t.Fatalf("expected default n 1, got %d", payload.N)
		}
		if payload.Size != "auto" {
			t.Fatalf("expected default size auto, got %s", payload.Size)
		}
		if payload.ResponseFormat != "url" {
			t.Fatalf("expected default response format url, got %s", payload.ResponseFormat)
		}

		_ = json.NewEncoder(w).Encode(ImageGenerateResp{
			Created: 123,
			Data: []GeneratedImage{
				{
					URL:           "https://example.com/image.png",
					RevisedPrompt: "A revised prompt",
				},
			},
		})
	}))
	defer server.Close()

	service := &ImageService{
		client: NewClientWithConfig(Config{
			APIKey:  "test-key",
			BaseURL: server.URL,
		}),
	}

	resp, err := service.Generate(&ImageGenerateReq{
		Prompt: "A futuristic city skyline.",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if resp.Created != 123 {
		t.Fatalf("unexpected created timestamp: %d", resp.Created)
	}
	if len(resp.Data) != 1 || resp.Data[0].URL != "https://example.com/image.png" {
		t.Fatalf("unexpected response data: %#v", resp.Data)
	}
	if resp.Data[0].RevisedPrompt != "A revised prompt" {
		t.Fatalf("unexpected revised prompt: %s", resp.Data[0].RevisedPrompt)
	}
}

func TestImageServiceGenerateSupportsOverrides(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload ImageGenerateReq
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}

		if payload.Model != "gpt-image-2-preview" {
			t.Fatalf("unexpected model override: %s", payload.Model)
		}
		if payload.N != 2 {
			t.Fatalf("unexpected n: %d", payload.N)
		}
		if payload.Size != "1024x1536" {
			t.Fatalf("unexpected size: %s", payload.Size)
		}
		if payload.ResponseFormat != "b64_json" {
			t.Fatalf("unexpected response format: %s", payload.ResponseFormat)
		}

		_ = json.NewEncoder(w).Encode(ImageGenerateResp{
			Created: 456,
			Data: []GeneratedImage{
				{B64JSON: "aGVsbG8="},
			},
		})
	}))
	defer server.Close()

	service := &ImageService{
		client: NewClientWithConfig(Config{
			APIKey:  "test-key",
			BaseURL: server.URL,
		}),
	}

	resp, err := service.Generate(&ImageGenerateReq{
		Model:          "gpt-image-2-preview",
		Prompt:         "A portrait in watercolor.",
		N:              2,
		Size:           "1024x1536",
		ResponseFormat: "b64_json",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].B64JSON != "aGVsbG8=" {
		t.Fatalf("unexpected response data: %#v", resp.Data)
	}
}

func TestClientCreateImageGenerationReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "rate limited",
				"type":    "rate_limit_error",
				"code":    "too_many_requests",
			},
		})
	}))
	defer server.Close()

	client := NewClientWithConfig(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	_, err := client.CreateImageGeneration(&ImageGenerateReq{
		Model:  ModelGPTImage2,
		Prompt: "A lighthouse in fog.",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("unexpected status code: %d", apiErr.StatusCode)
	}
	if apiErr.Message != "rate limited" {
		t.Fatalf("unexpected error message: %s", apiErr.Message)
	}
}
