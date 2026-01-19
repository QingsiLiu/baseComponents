package replicate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiToken   string
	timeout    time.Duration
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		apiToken:   GetAPIToken(),
		timeout:    DefaultTimeout,
	}
}

func NewClientWithToken(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		apiToken:   token,
		timeout:    DefaultTimeout,
	}
}

func (c *Client) GetAPIToken() string {
	return c.apiToken
}

func (c *Client) CreatePrediction(req *PredictionRequest) (*PredictionResponse, error) {
	endpoint := BaseURL + PathPredictions

	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	log.Printf("Replicate API request to %s: %s", endpoint, string(reqBody))

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiToken)
	httpReq.Header.Set("Prefer", "respond-async")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Response body read error: %v", err)
		return nil, fmt.Errorf("response body read error: %w", err)
	}

	log.Printf("Replicate API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, errResp.Title, errResp.Detail)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result PredictionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON decode error: %v", err)
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return &result, nil
}

// GetPrediction 获取预测任务信息
func (c *Client) GetPrediction(predictionID string) (*PredictionResponse, error) {
	endpoint := BaseURL + fmt.Sprintf(PathPredictionGet, predictionID)

	httpReq, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Response body read error: %v", err)
		return nil, fmt.Errorf("response body read error: %w", err)
	}

	log.Printf("Replicate API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, errResp.Title, errResp.Detail)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result PredictionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON decode error: %v", err)
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return &result, nil
}

// CancelPrediction 取消预测任务
func (c *Client) CancelPrediction(predictionID string) (*PredictionResponse, error) {
	endpoint := BaseURL + fmt.Sprintf(PathPredictionCancel, predictionID)

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		log.Printf("HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Response body read error: %v", err)
		return nil, fmt.Errorf("response body read error: %w", err)
	}

	log.Printf("Replicate API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, errResp.Title, errResp.Detail)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result PredictionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON decode error: %v", err)
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return &result, nil
}

// ListPredictions 列出预测任务
func (c *Client) ListPredictions() ([]PredictionResponse, error) {
	endpoint := BaseURL + PathPredictions

	httpReq, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Response body read error: %v", err)
		return nil, fmt.Errorf("response body read error: %w", err)
	}

	log.Printf("Replicate API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, errResp.Title, errResp.Detail)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Results []PredictionResponse `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("JSON decode error: %v", err)
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return result.Results, nil
}

// === 便捷方法 ===

// CreateText2ImageTask 创建文本转图像任务
func (c *Client) CreateText2ImageTask(model string, input interface{}) (string, error) {
	req := &PredictionRequest{
		Version: model,
		Input:   input,
	}

	resp, err := c.CreatePrediction(req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// CreateImage2ImageTask 创建图像转图像任务
func (c *Client) CreateImage2ImageTask(model string, input interface{}) (string, error) {
	req := &PredictionRequest{
		Version: model,
		Input:   input,
	}

	resp, err := c.CreatePrediction(req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// WaitForCompletion 等待任务完成
func (c *Client) WaitForCompletion(predictionID string, maxWaitTime time.Duration) (*PredictionResponse, error) {
	startTime := time.Now()
	checkInterval := 2 * time.Second

	for {
		if time.Since(startTime) > maxWaitTime {
			return nil, fmt.Errorf("task timeout after %v", maxWaitTime)
		}

		resp, err := c.GetPrediction(predictionID)
		if err != nil {
			return nil, err
		}

		if IsFinalStatus(resp.Status) {
			return resp, nil
		}

		time.Sleep(checkInterval)

		// 动态调整检查间隔
		if time.Since(startTime) > 30*time.Second {
			checkInterval = 5 * time.Second
		}
	}
}

// CheckLinkAvailability 检查链接可用性
func (c *Client) CheckLinkAvailability(url string) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Head(url)
	if err != nil {
		log.Printf("Link availability check failed for %s: %v", url, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
