package llm

import "testing"

func TestServiceTypeMappings(t *testing.T) {
	if !IsValidSource(SourceWellAPIGemini) {
		t.Fatalf("expected source %s to be valid", SourceWellAPIGemini)
	}

	if got := GetServiceType(SourceWellAPIGemini); got != ServiceTypeWellAPIGemini {
		t.Fatalf("expected service type %s, got %s", ServiceTypeWellAPIGemini, got)
	}

	if got := GetServiceSource(ServiceTypeWellAPIGemini); got != SourceWellAPIGemini {
		t.Fatalf("expected source %s, got %s", SourceWellAPIGemini, got)
	}
}

func TestGenerateRespDecodeJSON(t *testing.T) {
	resp := &GenerateResp{
		Text: `{"name":"assistant","ok":true}`,
	}

	var out struct {
		Name string `json:"name"`
		OK   bool   `json:"ok"`
	}

	if err := resp.DecodeJSON(&out); err != nil {
		t.Fatalf("DecodeJSON returned error: %v", err)
	}

	if out.Name != "assistant" || !out.OK {
		t.Fatalf("unexpected decoded payload: %+v", out)
	}
}

func TestGenerateRespHelpers(t *testing.T) {
	resp := &GenerateResp{
		FunctionCalls: []FunctionCall{
			{
				Name: "schedule_meeting",
				Args: map[string]any{
					"topic": "launch",
				},
			},
		},
	}

	if !resp.HasFunctionCalls() {
		t.Fatal("expected function calls to be present")
	}

	call := resp.FirstFunctionCall()
	if call == nil {
		t.Fatal("expected first function call to be non-nil")
	}

	if call.Name != "schedule_meeting" {
		t.Fatalf("expected function name schedule_meeting, got %s", call.Name)
	}
}
