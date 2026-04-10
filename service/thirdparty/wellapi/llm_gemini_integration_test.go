package wellapi

import (
	"bytes"
	"encoding/base64"
	stderrors "errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QingsiLiu/baseComponents/service/llm"
)

func newIntegrationClient(t *testing.T) *Client {
	t.Helper()

	apiKey := strings.TrimSpace(os.Getenv(APIKeyEnvVar))
	if apiKey == "" {
		t.Skip("WELLAPI_API_KEY is not set")
	}

	return NewClientWithConfig(Config{
		APIKey:         apiKey,
		Timeout:        60 * time.Second,
		RetryMax:       1,
		RetryBaseDelay: 200 * time.Millisecond,
	})
}

func newIntegrationService(t *testing.T) *GeminiService {
	t.Helper()
	return &GeminiService{client: newIntegrationClient(t)}
}

func TestWellAPIIntegrationListModels(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.ListModels()
	if err != nil {
		t.Fatalf("ListModels returned error: %v", err)
	}

	want := map[string]bool{
		ModelGemini3FlashPreview:         false,
		ModelGemini31FlashPreview:        false,
		ModelGemini31FlashLitePreview:    false,
		ModelGemini3FlashPreviewThinking: false,
	}

	for _, model := range resp.Data {
		if _, ok := want[model.ID]; ok {
			want[model.ID] = true
		}
	}

	for model, seen := range want {
		if !seen {
			t.Fatalf("expected model %s in list response", model)
		}
	}
}

func TestGeminiIntegrationTextGeneration(t *testing.T) {
	service := newIntegrationService(t)

	cases := []string{
		ModelGemini3FlashPreview,
		ModelGemini31FlashLitePreview,
	}

	for _, model := range cases {
		resp, err := service.Generate(&llm.GenerateReq{
			Model: model,
			Messages: []llm.Message{
				{Role: "user", Parts: []llm.Part{{Text: "Reply with OK only."}}},
			},
			MaxOutputTokens: 16,
		})
		if err != nil {
			t.Fatalf("Generate returned error for model %s: %v", model, err)
		}

		if !strings.Contains(strings.ToUpper(resp.Text), "OK") {
			t.Fatalf("expected response to contain OK for model %s, got %q", model, resp.Text)
		}
	}
}

func TestGeminiIntegrationStructuredOutput(t *testing.T) {
	service := newIntegrationService(t)

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGemini3FlashPreview,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "Return a JSON object with fields name=assistant and ok=true."}}},
		},
		ResponseSchema: map[string]any{
			"type": "OBJECT",
			"properties": map[string]any{
				"name": map[string]any{"type": "STRING"},
				"ok":   map[string]any{"type": "BOOLEAN"},
			},
			"required":         []string{"name", "ok"},
			"propertyOrdering": []string{"name", "ok"},
		},
		MaxOutputTokens: 64,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	var out struct {
		Name string `json:"name"`
		OK   bool   `json:"ok"`
	}

	if err := resp.DecodeJSON(&out); err != nil {
		t.Fatalf("DecodeJSON returned error: %v", err)
	}

	if out.Name != "assistant" || !out.OK {
		t.Fatalf("unexpected structured output: %+v", out)
	}
}

func TestGeminiIntegrationFunctionCalling(t *testing.T) {
	service := newIntegrationService(t)

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGemini3FlashPreview,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "Schedule a meeting with Alice tomorrow at 10:00 about launch."}}},
		},
		Tools: []llm.ToolSpec{
			{
				Name:        "schedule_meeting",
				Description: "Schedule a meeting",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"attendees": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"date":      map[string]any{"type": "string"},
						"time":      map[string]any{"type": "string"},
						"topic":     map[string]any{"type": "string"},
					},
					"required": []string{"attendees", "date", "time", "topic"},
				},
			},
		},
		MaxOutputTokens: 128,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	call := resp.FirstFunctionCall()
	if call == nil {
		t.Fatal("expected at least one function call")
	}

	if call.Name != "schedule_meeting" {
		t.Fatalf("unexpected function call name: %s", call.Name)
	}
}

func TestGeminiIntegrationImageUnderstanding(t *testing.T) {
	service := newIntegrationService(t)

	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGemini3FlashPreview,
		Messages: []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Text: "What is the dominant color in this image? Reply with one lowercase English word."},
					{MimeType: "image/png", InlineDataBase64: base64.StdEncoding.EncodeToString(buf.Bytes())},
				},
			},
		},
		MaxOutputTokens: 16,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if !strings.Contains(strings.ToLower(resp.Text), "red") {
		t.Fatalf("expected response to mention red, got %q", resp.Text)
	}
}

func TestGeminiIntegrationGemini31FlashPreview(t *testing.T) {
	service := newIntegrationService(t)

	_, err := service.Generate(&llm.GenerateReq{
		Model: ModelGemini31FlashPreview,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "Reply with OK only."}}},
		},
		MaxOutputTokens: 16,
	})
	if err == nil {
		return
	}

	var apiErr *APIError
	if strings.Contains(err.Error(), "429") || (strings.Contains(err.Error(), "upstream_error")) {
		t.Skipf("gemini-3.1-flash-preview currently unavailable: %v", err)
	}
	if stderrors.As(err, &apiErr) && apiErr != nil && apiErr.IsRetryable() {
		t.Skipf("gemini-3.1-flash-preview currently unavailable: %v", err)
	}

	t.Fatalf("Generate returned unexpected error: %v", err)
}

func TestGeminiIntegrationGemini25Flash(t *testing.T) {
	service := newIntegrationService(t)

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGemini25Flash,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "Reply with OK only."}}},
		},
		MaxOutputTokens: 16,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if !strings.Contains(strings.ToUpper(resp.Text), "OK") {
		t.Fatalf("expected response to contain OK, got %q", resp.Text)
	}
}

func TestGeminiIntegrationGemini31FlashLitePreview(t *testing.T) {
	service := newIntegrationService(t)

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGemini31FlashLitePreview,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "Reply with OK only."}}},
		},
		MaxOutputTokens: 16,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if !strings.Contains(strings.ToUpper(resp.Text), "OK") {
		t.Fatalf("expected response to contain OK, got %q", resp.Text)
	}
}
