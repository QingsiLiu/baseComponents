package modelslab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	timeout    time.Duration
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		apiKey:     os.Getenv(APIKeyEnvVar),
		timeout:    DefaultTimeout,
	}
}

func NewClientWithKey(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		apiKey:     apiKey,
		timeout:    DefaultTimeout,
	}
}

func (c *Client) GetAPIKey() string {
	return c.apiKey
}

func (c *Client) Post(endpoint string, payload interface{}) (*http.Response, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	log.Printf("ModelsLab API request to %s: %s", endpoint, string(reqBody))

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("HTTP request creation error: %v", err)
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		return nil, fmt.Errorf("http request error: %w", err)
	}

	return resp, nil
}

func (c *Client) PostAndDecode(endpoint string, payload interface{}, result interface{}) error {
	resp, err := c.Post(endpoint, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Response body read error: %v", err)
		return fmt.Errorf("response body read error: %w", err)
	}

	log.Printf("ModelsLab API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf("API error (status %d): %v", resp.StatusCode, errResp)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		log.Printf("JSON decode error: %v", err)
		return fmt.Errorf("json decode error: %w", err)
	}

	return nil
}

func (c *Client) CreateTaskGetRequest(taskId string) *TaskGetRequest {
	return &TaskGetRequest{
		Key:       c.apiKey,
		RequestID: taskId,
	}
}

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
