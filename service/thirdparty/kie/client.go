package kie

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Client 封装与 KIE API 的交互
type Client struct {
	httpClient *http.Client
	apiKey     string
	timeout    time.Duration
	baseURL    string
}

// NewClient 使用环境变量中的 API Key 创建客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		apiKey:     GetAPIKey(),
		timeout:    DefaultTimeout,
		baseURL:    BaseURL,
	}
}

// NewClientWithKey 使用指定 API Key 创建客户端
func NewClientWithKey(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		apiKey:     apiKey,
		timeout:    DefaultTimeout,
		baseURL:    BaseURL,
	}
}

// GetAPIKey 返回当前使用的 API Key
func (c *Client) GetAPIKey() string {
	return c.apiKey
}

// CreateTask 创建生成任务
func (c *Client) CreateTask(payload *TaskCreateRequest) (*TaskCreateResponse, error) {
	endpoint := c.baseURL + CreateTaskEndpoint

	reqBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("KIE JSON marshal error: %v", err)
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	log.Printf("KIE API request to %s: %s", endpoint, string(reqBody))

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("KIE HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("KIE HTTP request error: %v", err)
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("KIE response body read error: %v", err)
		return nil, fmt.Errorf("response body read error: %w", err)
	}

	log.Printf("KIE API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result TaskCreateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("KIE JSON decode error: %v", err)
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	if result.Code != http.StatusOK {
		return nil, fmt.Errorf("API error code %d: %s", result.Code, result.Message)
	}

	if result.Data == nil || result.Data.TaskID == "" {
		return nil, fmt.Errorf("API response missing task ID")
	}

	return &result, nil
}

// GetTaskRecord 查询任务详情
func (c *Client) GetTaskRecord(taskID string) (*TaskRecordResponse, error) {
	endpoint, err := url.Parse(c.baseURL + RecordInfoEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("taskId", taskID)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		log.Printf("KIE HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("KIE HTTP request error: %v", err)
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("KIE response body read error: %v", err)
		return nil, fmt.Errorf("response body read error: %w", err)
	}

	log.Printf("KIE API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result TaskRecordResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("KIE JSON decode error: %v", err)
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	if result.Code != http.StatusOK {
		return nil, fmt.Errorf("API error code %d: %s", result.Code, result.Message)
	}

	if result.Data == nil {
		return nil, fmt.Errorf("API response missing task data")
	}

	return &result, nil
}

// CheckLinkAvailability 检查链接是否可用
func (c *Client) CheckLinkAvailability(link string) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Head(link)
	if err != nil {
		log.Printf("KIE link availability check failed for %s: %v", link, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
