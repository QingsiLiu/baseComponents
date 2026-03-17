package wellapi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// StreamHandler 处理流式响应分片
type StreamHandler func(*GenerateContentResponse) error

// Client WellAPI 客户端
type Client struct {
	httpClient     *http.Client
	apiKey         string
	timeout        time.Duration
	baseURL        string
	retryMax       int
	retryBaseDelay time.Duration
}

// NewClient 使用环境变量创建默认客户端
func NewClient() *Client {
	return NewClientWithConfig(Config{})
}

// NewClientWithKey 使用指定 API Key 创建客户端
func NewClientWithKey(apiKey string) *Client {
	return NewClientWithConfig(Config{APIKey: apiKey})
}

// NewClientWithConfig 使用自定义配置创建客户端
func NewClientWithConfig(cfg Config) *Client {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		apiKey = GetAPIKey()
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = GetBaseURL()
	} else {
		baseURL = strings.TrimRight(baseURL, "/")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	retryMax := cfg.RetryMax
	if retryMax <= 0 {
		retryMax = DefaultRetryMax
	}

	retryBaseDelay := cfg.RetryBaseDelay
	if retryBaseDelay <= 0 {
		retryBaseDelay = DefaultRetryDelay
	}

	return &Client{
		httpClient:     &http.Client{Timeout: timeout},
		apiKey:         apiKey,
		timeout:        timeout,
		baseURL:        baseURL,
		retryMax:       retryMax,
		retryBaseDelay: retryBaseDelay,
	}
}

// GetAPIKey 返回当前使用的 API Key
func (c *Client) GetAPIKey() string {
	return c.apiKey
}

// ListModels 获取模型列表
func (c *Client) ListModels() (*ListModelsResponse, error) {
	endpoint := c.baseURL + PathModels
	httpReq, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("WellAPI HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	respBody, err := c.do(httpReq, nil)
	if err != nil {
		return nil, err
	}

	var result ListModelsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("WellAPI JSON decode error: %v", err)
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return &result, nil
}

// GenerateContent 调用 Gemini 原生 generateContent
func (c *Client) GenerateContent(req *GenerateContentRequest) (*GenerateContentResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, fmt.Errorf("model is required")
	}

	endpoint := c.baseURL + fmt.Sprintf(PathGenerateFormat, req.Model)
	httpReq, err := c.newJSONRequest(http.MethodPost, endpoint, req)
	if err != nil {
		return nil, err
	}

	respBody, err := c.do(httpReq, req)
	if err != nil {
		return nil, err
	}

	var result GenerateContentResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("WellAPI JSON decode error: %v", err)
		return nil, fmt.Errorf("json decode error: %w", err)
	}
	result.Raw = append([]byte(nil), respBody...)

	return &result, nil
}

// StreamGenerateContent 调用 Gemini 原生流式接口
func (c *Client) StreamGenerateContent(req *GenerateContentRequest, handler StreamHandler) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	if handler == nil {
		return fmt.Errorf("stream handler is nil")
	}
	if strings.TrimSpace(req.Model) == "" {
		return fmt.Errorf("model is required")
	}

	endpoint := c.baseURL + fmt.Sprintf(PathStreamFormat, req.Model) + "?alt=sse"
	httpReq, err := c.newJSONRequest(http.MethodPost, endpoint, req)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("WellAPI HTTP request error: %v", err)
		return fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Printf("WellAPI response body read error: %v", readErr)
			return fmt.Errorf("response body read error: %w", readErr)
		}
		log.Printf("WellAPI API response (status %d): %s", resp.StatusCode, truncateForLog(string(body)))
		return c.buildAPIError(resp.StatusCode, body)
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	eventLines := make([]string, 0, 4)
	flushEvent := func() error {
		if len(eventLines) == 0 {
			return nil
		}

		payload := strings.Join(eventLines, "\n")
		eventLines = eventLines[:0]
		payload = strings.TrimSpace(payload)
		if payload == "" {
			return nil
		}

		var chunk GenerateContentResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			log.Printf("WellAPI stream chunk decode error: %v", err)
			return fmt.Errorf("stream chunk decode error: %w", err)
		}
		chunk.Raw = []byte(payload)

		if err := handler(&chunk); err != nil {
			return err
		}

		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := flushEvent(); err != nil {
				return err
			}
			continue
		}

		if strings.HasPrefix(line, "data:") {
			eventLines = append(eventLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream read error: %w", err)
	}

	return flushEvent()
}

func (c *Client) newJSONRequest(method, endpoint string, payload interface{}) (*http.Request, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("WellAPI JSON marshal error: %v", err)
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	log.Printf("WellAPI request to %s: %s", endpoint, truncateForLog(string(reqBody)))

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("WellAPI HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	return req, nil
}

func (c *Client) do(req *http.Request, payload interface{}) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("WellAPI HTTP request error: %v", err)
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("WellAPI response body read error: %v", err)
		return nil, fmt.Errorf("response body read error: %w", err)
	}

	log.Printf("WellAPI API response (status %d): %s", resp.StatusCode, truncateForLog(string(body)))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, c.buildAPIError(resp.StatusCode, body)
	}

	return body, nil
}

func (c *Client) buildAPIError(statusCode int, body []byte) error {
	var env apiErrorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Error != nil {
		return &APIError{
			StatusCode: statusCode,
			Message:    env.Error.Message,
			Type:       env.Error.Type,
			Param:      fmt.Sprint(env.Error.Param),
			Code:       fmt.Sprint(env.Error.Code),
			Raw:        string(body),
		}
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    string(body),
		Raw:        string(body),
	}
}

func truncateForLog(value string) string {
	const maxLen = 2048
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "...(truncated)"
}
