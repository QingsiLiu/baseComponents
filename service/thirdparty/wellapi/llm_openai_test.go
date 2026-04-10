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

func TestGeminiResolveModelsIgnoresThinkingBudget(t *testing.T) {
	service := &GeminiService{client: NewClientWithConfig(Config{APIKey: "test-key"})}
	thinkingBudget := 8192

	models, explicit := service.resolveModels(&llm.GenerateReq{
		ThinkingBudget: &thinkingBudget,
	})

	if explicit {
		t.Fatal("expected implicit model selection")
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 fallback models, got %d", len(models))
	}
	if models[0] != ModelGemini3FlashPreview || models[1] != ModelGemini31FlashLitePreview {
		t.Fatalf("unexpected model chain: %v", models)
	}
}

func TestBuildChatCompletionRequest(t *testing.T) {
	temperature := 0.2
	topP := 0.8

	wireReq, err := buildChatCompletionRequest(&llm.GenerateReq{
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
				Parameters:  map[string]any{"type": "object"},
			},
		},
		ResponseMIMEType: "application/json",
		Temperature:      &temperature,
		TopP:             &topP,
		MaxOutputTokens:  128,
	}, ModelGPT54Mini)
	if err != nil {
		t.Fatalf("buildChatCompletionRequest returned error: %v", err)
	}

	if wireReq.Model != ModelGPT54Mini {
		t.Fatalf("expected model %s, got %s", ModelGPT54Mini, wireReq.Model)
	}
	if len(wireReq.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(wireReq.Messages))
	}
	if wireReq.Messages[0].Role != "system" {
		t.Fatalf("expected first message to be system, got %s", wireReq.Messages[0].Role)
	}
	if len(wireReq.Messages[1].Content) != 2 {
		t.Fatalf("expected 2 content parts, got %d", len(wireReq.Messages[1].Content))
	}
	if got := wireReq.Messages[1].Content[1].ImageURL.URL; got != "data:image/png;base64,aGVsbG8=" {
		t.Fatalf("unexpected image url payload: %s", got)
	}
	if wireReq.ResponseFormat == nil || wireReq.ResponseFormat.Type != "json_object" {
		t.Fatalf("expected json_object response format, got %+v", wireReq.ResponseFormat)
	}
	if wireReq.ToolChoice != "auto" {
		t.Fatalf("expected tool choice auto, got %v", wireReq.ToolChoice)
	}
}

func TestBuildResponsesRequest(t *testing.T) {
	thinkingBudget := 9000

	wireReq, err := buildResponsesRequest(&llm.GenerateReq{
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
				Parameters:  map[string]any{"type": "object"},
			},
		},
		ResponseSchema: map[string]any{
			"type": "OBJECT",
			"properties": map[string]any{
				"name": map[string]any{"type": "STRING"},
			},
		},
		MaxOutputTokens: 256,
		ThinkingBudget:  &thinkingBudget,
	}, ModelGPT54Mini)
	if err != nil {
		t.Fatalf("buildResponsesRequest returned error: %v", err)
	}

	if wireReq.Model != ModelGPT54Mini {
		t.Fatalf("expected model %s, got %s", ModelGPT54Mini, wireReq.Model)
	}
	if wireReq.Instructions != "You are helpful." {
		t.Fatalf("unexpected instructions: %s", wireReq.Instructions)
	}
	if len(wireReq.Input) != 1 || len(wireReq.Input[0].Content) != 2 {
		t.Fatalf("unexpected input payload: %+v", wireReq.Input)
	}
	if wireReq.Input[0].Content[0].Type != "input_text" {
		t.Fatalf("expected first input part to be input_text, got %s", wireReq.Input[0].Content[0].Type)
	}
	if wireReq.Input[0].Content[1].ImageURL != "data:image/png;base64,aGVsbG8=" {
		t.Fatalf("unexpected image url payload: %s", wireReq.Input[0].Content[1].ImageURL)
	}
	if wireReq.Text == nil || wireReq.Text.Format == nil || wireReq.Text.Format.Type != "json_schema" {
		t.Fatalf("expected json_schema text config, got %+v", wireReq.Text)
	}
	if wireReq.Text.Format.Name != "response" || !wireReq.Text.Format.Strict {
		t.Fatalf("unexpected text format payload: %+v", wireReq.Text.Format)
	}
	if got := wireReq.Text.Format.Schema["type"]; got != "object" {
		t.Fatalf("expected normalized root type object, got %#v", got)
	}
	if got := wireReq.Text.Format.Schema["additionalProperties"]; got != false {
		t.Fatalf("expected additionalProperties=false, got %#v", got)
	}
	properties := wireReq.Text.Format.Schema["properties"].(map[string]any)
	nameSchema := properties["name"].(map[string]any)
	if got := nameSchema["type"]; got != "string" {
		t.Fatalf("expected normalized property type string, got %#v", got)
	}
	if wireReq.Reasoning == nil || wireReq.Reasoning.Effort != "medium" {
		t.Fatalf("expected medium reasoning effort, got %+v", wireReq.Reasoning)
	}
	if wireReq.ToolChoice != "auto" {
		t.Fatalf("expected tool choice auto, got %v", wireReq.ToolChoice)
	}
}

func TestThinkingBudgetToEffort(t *testing.T) {
	cases := []struct {
		budget int
		want   string
	}{
		{budget: 1, want: "minimal"},
		{budget: 1024, want: "minimal"},
		{budget: 2048, want: "low"},
		{budget: 9000, want: "medium"},
		{budget: 50000, want: "high"},
	}

	for _, tc := range cases {
		if got := thinkingBudgetToEffort(tc.budget); got != tc.want {
			t.Fatalf("thinkingBudgetToEffort(%d) = %s, want %s", tc.budget, got, tc.want)
		}
	}
}

func TestConvertChatCompletionResponse(t *testing.T) {
	resp, err := convertChatCompletionResponse(&ChatCompletionResponse{
		Model: "gpt-5.4-mini",
		Raw:   []byte(`{"raw":"value"}`),
		Choices: []ChatCompletionChoice{
			{
				FinishReason: "stop",
				Message: ChatCompletionResponseMessage{
					Role:    "assistant",
					Content: json.RawMessage(`"hello"`),
					ToolCalls: []ChatCompletionToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: ChatCompletionToolCallFunction{
								Name:      "schedule_meeting",
								Arguments: `{"topic":"launch"}`,
							},
						},
					},
				},
			},
		},
		Usage: &OpenAIUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	})
	if err != nil {
		t.Fatalf("convertChatCompletionResponse returned error: %v", err)
	}

	if resp.Text != "hello" {
		t.Fatalf("expected text hello, got %q", resp.Text)
	}
	if resp.FinishReason != "stop" {
		t.Fatalf("unexpected finish reason: %s", resp.FinishReason)
	}
	if len(resp.FunctionCalls) != 1 || resp.FunctionCalls[0].Name != "schedule_meeting" {
		t.Fatalf("unexpected function calls: %+v", resp.FunctionCalls)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
}

func TestConvertResponsesResponse(t *testing.T) {
	resp, err := convertResponsesResponse(&ResponsesResponse{
		Model:  "gpt-5.4-mini",
		Raw:    []byte(`{"raw":"value"}`),
		Status: "completed",
		Output: []ResponsesOutputItem{
			{
				Type: "message",
				Content: []ResponsesOutputContent{
					{Type: "output_text", Text: "hello"},
				},
			},
			{
				Type:      "function_call",
				Name:      "schedule_meeting",
				Arguments: `{"topic":"launch"}`,
				CallID:    "call_1",
			},
		},
		Usage: &ResponsesUsage{
			InputTokens:  10,
			OutputTokens: 6,
			TotalTokens:  16,
			OutputTokensDetails: &ResponsesOutputTokensDetail{
				ReasoningTokens: 3,
			},
		},
	})
	if err != nil {
		t.Fatalf("convertResponsesResponse returned error: %v", err)
	}

	if resp.Text != "hello" {
		t.Fatalf("expected text hello, got %q", resp.Text)
	}
	if resp.FinishReason != "completed" {
		t.Fatalf("unexpected finish reason: %s", resp.FinishReason)
	}
	if len(resp.FunctionCalls) != 1 || resp.FunctionCalls[0].Name != "schedule_meeting" {
		t.Fatalf("unexpected function calls: %+v", resp.FunctionCalls)
	}
	if resp.Usage.ThoughtsTokens != 3 {
		t.Fatalf("unexpected thoughts tokens: %+v", resp.Usage)
	}
}

func TestOpenAIServiceRejectsUnsupportedFlags(t *testing.T) {
	service := &OpenAIService{client: NewClientWithConfig(Config{APIKey: "test-key"})}

	_, err := service.Generate(&llm.GenerateReq{
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "hello"}}},
		},
		EnableURLContext: true,
	})
	if err == nil {
		t.Fatal("expected Generate to return an error")
	}
	if !strings.Contains(err.Error(), "EnableURLContext") {
		t.Fatalf("expected error to mention EnableURLContext, got %v", err)
	}
}

func TestOpenAIServiceFallbackFromResponsesToChat(t *testing.T) {
	var responsesCalls int32
	var chatCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PathResponses:
			atomic.AddInt32(&responsesCalls, 1)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"model does not support responses api","type":"invalid_request_error"}}`))
		case PathChatCompletions:
			atomic.AddInt32(&chatCalls, 1)
			_, _ = w.Write([]byte(`{"id":"chatcmpl","model":"gpt-5.4-mini","choices":[{"index":0,"message":{"role":"assistant","content":"chat ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := &OpenAIService{
		client: NewClientWithConfig(Config{
			APIKey:         "test-key",
			BaseURL:        server.URL,
			RetryMax:       1,
			RetryBaseDelay: time.Millisecond,
		}),
	}

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGPT54Mini,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "hello"}}},
		},
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if resp.Text != "chat ok" {
		t.Fatalf("expected fallback text chat ok, got %q", resp.Text)
	}
	if atomic.LoadInt32(&responsesCalls) != 1 || atomic.LoadInt32(&chatCalls) != 1 {
		t.Fatalf("unexpected endpoint calls responses=%d chat=%d", responsesCalls, chatCalls)
	}
}

func TestOpenAIServiceUsesResponsesOnlyForGPT54(t *testing.T) {
	var responsesCalls int32
	var chatCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PathResponses:
			atomic.AddInt32(&responsesCalls, 1)
			_, _ = w.Write([]byte(`{"id":"resp","model":"gpt-5.4","status":"completed","output_text":"response ok","usage":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}`))
		case PathChatCompletions:
			atomic.AddInt32(&chatCalls, 1)
			http.Error(w, "unexpected chat call", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := &OpenAIService{
		client: NewClientWithConfig(Config{
			APIKey:         "test-key",
			BaseURL:        server.URL,
			RetryMax:       1,
			RetryBaseDelay: time.Millisecond,
		}),
	}

	resp, err := service.Generate(&llm.GenerateReq{
		Model: ModelGPT54,
		Messages: []llm.Message{
			{Role: "user", Parts: []llm.Part{{Text: "hello"}}},
		},
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if resp.Text != "response ok" {
		t.Fatalf("expected response text, got %q", resp.Text)
	}
	if atomic.LoadInt32(&responsesCalls) != 1 {
		t.Fatalf("expected one responses call, got %d", responsesCalls)
	}
	if atomic.LoadInt32(&chatCalls) != 0 {
		t.Fatalf("expected no chat calls, got %d", chatCalls)
	}
}

func TestOpenAIServiceDefaultModelFallback(t *testing.T) {
	var responsesCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != PathResponses {
			http.NotFound(w, r)
			return
		}

		atomic.AddInt32(&responsesCalls, 1)

		var req ResponsesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		switch req.Model {
		case ModelGPT54Mini:
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"busy","type":"upstream_error","code":"429"}}`))
		case ModelGPT5Mini20250807:
			_, _ = w.Write([]byte(`{"id":"resp","model":"gpt-5-mini-2025-08-07","status":"completed","output_text":"fallback ok","usage":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}`))
		default:
			t.Fatalf("unexpected model requested: %s", req.Model)
		}
	}))
	defer server.Close()

	service := &OpenAIService{
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
		t.Fatalf("expected fallback text, got %q", resp.Text)
	}
	if atomic.LoadInt32(&responsesCalls) != 2 {
		t.Fatalf("expected two responses calls, got %d", responsesCalls)
	}
}
