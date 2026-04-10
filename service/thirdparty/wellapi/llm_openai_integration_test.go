package wellapi

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/QingsiLiu/baseComponents/service/llm"
)

func newOpenAIIntegrationService(t *testing.T) *OpenAIService {
	t.Helper()
	return &OpenAIService{client: newIntegrationClient(t)}
}

func TestOpenAIIntegrationTextGeneration(t *testing.T) {
	service := newOpenAIIntegrationService(t)

	for _, model := range []string{ModelGPT54Mini, ModelGPT54} {
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

func TestOpenAIIntegrationStructuredOutput(t *testing.T) {
	service := newOpenAIIntegrationService(t)

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGPT54Mini,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "Return a JSON object with fields name=assistant and ok=true."}}},
		},
		ResponseSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"ok":   map[string]any{"type": "boolean"},
			},
			"required": []string{"name", "ok"},
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

func TestOpenAIIntegrationFunctionCalling(t *testing.T) {
	service := newOpenAIIntegrationService(t)

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGPT54Mini,
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

func TestOpenAIIntegrationImageUnderstanding(t *testing.T) {
	service := newOpenAIIntegrationService(t)

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
		Model: ModelGPT54Mini,
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
