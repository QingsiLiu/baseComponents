package wellapi

import (
	"encoding/json"
	"fmt"
)

// CreateChatCompletion 调用 WellAPI OpenAI Chat Completions
func (c *Client) CreateChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	httpReq, err := c.newJSONRequest("POST", c.baseURL+PathChatCompletions, req)
	if err != nil {
		return nil, err
	}

	respBody, err := c.do(httpReq, req)
	if err != nil {
		return nil, err
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}
	result.Raw = append([]byte(nil), respBody...)

	return &result, nil
}

// CreateResponse 调用 WellAPI OpenAI Responses API
func (c *Client) CreateResponse(req *ResponsesRequest) (*ResponsesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	httpReq, err := c.newJSONRequest("POST", c.baseURL+PathResponses, req)
	if err != nil {
		return nil, err
	}

	respBody, err := c.do(httpReq, req)
	if err != nil {
		return nil, err
	}

	var result ResponsesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}
	result.Raw = append([]byte(nil), respBody...)

	return &result, nil
}
