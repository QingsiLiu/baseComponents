package wellapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QingsiLiu/baseComponents/service/llm"
)

func TestBuildGenerateContentRequest(t *testing.T) {
	service := &GeminiService{
		client: NewClientWithConfig(Config{APIKey: "test-key"}),
	}

	temperature := 0.3
	topP := 0.9
	thinkingBudget := 0
	req := &llm.GenerateReq{
		SystemInstruction: "You are helpful.",
		Messages: []llm.Message{
			{
				Role: "user",
				Parts: []llm.Part{
					{Text: "Describe this image."},
					{MimeType: "image/png", InlineDataBase64: "aGVsbG8="},
				},
			},
		},
		Tools: []llm.ToolSpec{
			{
				Name:        "schedule_meeting",
				Description: "Schedule a meeting",
				Parameters: map[string]any{
					"type": "object",
				},
			},
		},
		EnableURLContext:    true,
		EnableGoogleSearch:  true,
		EnableCodeExecution: true,
		ResponseSchema: map[string]any{
			"type": "OBJECT",
		},
		Temperature:     &temperature,
		TopP:            &topP,
		MaxOutputTokens: 256,
		ThinkingBudget:  &thinkingBudget,
	}

	wireReq := service.buildGenerateContentRequest(req, ModelGemini3FlashPreview)
	if wireReq.Model != ModelGemini3FlashPreview {
		t.Fatalf("expected model %s, got %s", ModelGemini3FlashPreview, wireReq.Model)
	}

	if wireReq.SystemInstruction == nil || len(wireReq.SystemInstruction.Parts) != 1 {
		t.Fatal("expected system instruction to be mapped")
	}

	if got := wireReq.SystemInstruction.Parts[0].Text; got != "You are helpful." {
		t.Fatalf("unexpected system instruction text: %s", got)
	}

	if len(wireReq.Contents) != 1 {
		t.Fatalf("expected 1 content message, got %d", len(wireReq.Contents))
	}

	if len(wireReq.Contents[0].Parts) != 2 {
		t.Fatalf("expected 2 parts in content, got %d", len(wireReq.Contents[0].Parts))
	}

	if got := wireReq.Contents[0].Parts[1].InlineData.MimeType; got != "image/png" {
		t.Fatalf("expected mime type image/png, got %s", got)
	}

	if wireReq.GenerationConfig == nil {
		t.Fatal("expected generation config to be set")
	}

	if got := wireReq.GenerationConfig.ResponseMIMEType; got != "application/json" {
		t.Fatalf("expected response mime type application/json, got %s", got)
	}

	if wireReq.GenerationConfig.ThinkingConfig == nil || wireReq.GenerationConfig.ThinkingConfig.ThinkingBudget != 0 {
		t.Fatal("expected thinking budget default to 0")
	}

	if len(wireReq.Tools) != 4 {
		t.Fatalf("expected 4 tool entries, got %d", len(wireReq.Tools))
	}

	if len(wireReq.Tools[0].FunctionDeclarations) != 1 {
		t.Fatalf("expected one function declaration, got %d", len(wireReq.Tools[0].FunctionDeclarations))
	}

	if wireReq.Tools[1].URLContext == nil {
		t.Fatal("expected URL context tool to be enabled")
	}

	if wireReq.Tools[2].GoogleSearch == nil {
		t.Fatal("expected google search tool to be enabled")
	}

	if wireReq.Tools[3].CodeExecution == nil {
		t.Fatal("expected code execution tool to be enabled")
	}
}

func TestConvertGenerateContentResponse(t *testing.T) {
	service := &GeminiService{
		client: NewClientWithConfig(Config{APIKey: "test-key"}),
	}

	wireResp := &GenerateContentResponse{
		ModelVersion: "gemini-3-flash-preview-nothinking",
		Raw:          []byte(`{"raw":"value"}`),
		Candidates: []Candidate{
			{
				FinishReason: "STOP",
				Content: Content{
					Role: "model",
					Parts: []Part{
						{Text: "Hello"},
						{FunctionCall: &WireFunctionCall{
							Name: "schedule_meeting",
							Args: map[string]any{"topic": "launch"},
							ID:   "call-1",
						}},
						{ExecutableCode: &ExecutableCode{Language: "PYTHON", Code: "print(1+2)"}},
					},
				},
			},
		},
		UsageMetadata: &UsageMetadata{
			PromptTokenCount:        10,
			ToolUsePromptTokenCount: 2,
			CandidatesTokenCount:    5,
			TotalTokenCount:         17,
			ThoughtsTokenCount:      3,
		},
	}

	resp, err := service.convertGenerateContentResponse(wireResp)
	if err != nil {
		t.Fatalf("convertGenerateContentResponse returned error: %v", err)
	}

	if resp.Text != "Hello" {
		t.Fatalf("expected text Hello, got %s", resp.Text)
	}

	if resp.ModelVersion != "gemini-3-flash-preview-nothinking" {
		t.Fatalf("unexpected model version: %s", resp.ModelVersion)
	}

	if len(resp.Parts) != 2 {
		t.Fatalf("expected 2 exposed response parts, got %d", len(resp.Parts))
	}

	if !resp.HasFunctionCalls() {
		t.Fatal("expected function calls to be present")
	}

	if resp.FunctionCalls[0].Name != "schedule_meeting" {
		t.Fatalf("unexpected function call name: %s", resp.FunctionCalls[0].Name)
	}

	if resp.Usage.PromptTokens != 12 {
		t.Fatalf("expected prompt tokens 12, got %d", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 5 {
		t.Fatalf("expected completion tokens 5, got %d", resp.Usage.CompletionTokens)
	}

	if resp.Usage.TotalTokens != 17 {
		t.Fatalf("expected total tokens 17, got %d", resp.Usage.TotalTokens)
	}

	if resp.Usage.ThoughtsTokens != 3 {
		t.Fatalf("expected thoughts tokens 3, got %d", resp.Usage.ThoughtsTokens)
	}
}

func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PathModels {
			http.NotFound(w, r)
			return
		}

		_ = json.NewEncoder(w).Encode(ListModelsResponse{
			Data: []Model{
				{ID: ModelGemini3FlashPreview},
				{ID: ModelGemini31FlashLitePreview},
			},
			Object:  "list",
			Success: true,
		})
	}))
	defer server.Close()

	client := NewClientWithConfig(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	resp, err := client.ListModels()
	if err != nil {
		t.Fatalf("ListModels returned error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 models, got %d", len(resp.Data))
	}
}

func TestStreamGenerateContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "alt=sse") {
			t.Fatalf("expected alt=sse query, got %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"Hello\"}]},\"finishReason\":\"STOP\"}],\"modelVersion\":\"test\"}\n\n"))
	}))
	defer server.Close()

	client := NewClientWithConfig(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	var gotText string
	err := client.StreamGenerateContent(&GenerateContentRequest{
		Model: "gemini-test",
		Contents: []Content{
			{
				Role:  "user",
				Parts: []Part{{Text: "hello"}},
			},
		},
	}, func(resp *GenerateContentResponse) error {
		gotText = resp.Candidates[0].Content.Parts[0].Text
		return nil
	})
	if err != nil {
		t.Fatalf("StreamGenerateContent returned error: %v", err)
	}

	if gotText != "Hello" {
		t.Fatalf("expected streamed text Hello, got %s", gotText)
	}
}

func TestGenerateFallbackWhenModelNotSpecified(t *testing.T) {
	var primaryCalls int32
	var fallbackCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1beta/models/" + ModelGemini3FlashPreview + ":generateContent":
			atomic.AddInt32(&primaryCalls, 1)
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"busy","type":"upstream_error","code":"429"}}`))
		case "/v1beta/models/" + ModelGemini31FlashLitePreview + ":generateContent":
			atomic.AddInt32(&fallbackCalls, 1)
			_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"fallback ok"}]},"finishReason":"STOP"}],"modelVersion":"lite"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := &GeminiService{
		client: NewClientWithConfig(Config{
			APIKey:         "test-key",
			BaseURL:        server.URL,
			RetryMax:       1,
			RetryBaseDelay: time.Millisecond,
		}),
	}

	resp, err := service.Generate(&llm.GenerateReq{
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "hello"}}},
		},
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if resp.Text != "fallback ok" {
		t.Fatalf("expected fallback text, got %s", resp.Text)
	}

	if atomic.LoadInt32(&primaryCalls) != 1 {
		t.Fatalf("expected primary model called once, got %d", primaryCalls)
	}

	if atomic.LoadInt32(&fallbackCalls) != 1 {
		t.Fatalf("expected fallback model called once, got %d", fallbackCalls)
	}
}

func TestGenerateExplicitModelRetriesWithoutFallback(t *testing.T) {
	var primaryCalls int32
	var fallbackCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1beta/models/" + ModelGemini31FlashPreview + ":generateContent":
			atomic.AddInt32(&primaryCalls, 1)
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"busy","type":"upstream_error","code":"429"}}`))
		case "/v1beta/models/" + ModelGemini31FlashLitePreview + ":generateContent":
			atomic.AddInt32(&fallbackCalls, 1)
			http.Error(w, "unexpected fallback", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := &GeminiService{
		client: NewClientWithConfig(Config{
			APIKey:         "test-key",
			BaseURL:        server.URL,
			RetryMax:       2,
			RetryBaseDelay: time.Millisecond,
		}),
	}

	_, err := service.Generate(&llm.GenerateReq{
		Model: ModelGemini31FlashPreview,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "hello"}}},
		},
	})
	if err == nil {
		t.Fatal("expected Generate to return an error")
	}

	if atomic.LoadInt32(&primaryCalls) != 2 {
		t.Fatalf("expected explicit model retried twice, got %d", primaryCalls)
	}

	if atomic.LoadInt32(&fallbackCalls) != 0 {
		t.Fatalf("expected fallback model to not be called, got %d", fallbackCalls)
	}
}
