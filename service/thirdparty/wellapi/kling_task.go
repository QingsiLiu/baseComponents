package wellapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"
)

type klingTaskService struct {
	source          string
	createPath      string
	queryPathFormat string
	displayName     string
	client          *Client
}

func newKlingTaskService(source, createPath, queryPathFormat, displayName string, client *Client) *klingTaskService {
	return &klingTaskService{
		source:          source,
		createPath:      createPath,
		queryPathFormat: queryPathFormat,
		displayName:     displayName,
		client:          client,
	}
}

func (s *klingTaskService) Source() string {
	return s.source
}

func (s *klingTaskService) createTask(payload any) (string, error) {
	httpReq, err := s.client.newJSONRequest(http.MethodPost, s.client.baseURL+s.createPath, payload)
	if err != nil {
		return "", err
	}

	respBody, err := s.client.do(httpReq, payload)
	if err != nil {
		return "", err
	}

	var resp klingTaskCreateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("json decode error: %w", err)
	}
	resp.Raw = append([]byte(nil), respBody...)
	if resp.Code != 0 {
		return "", fmt.Errorf("%s api error code %d: %s", s.displayName, resp.Code, resp.Message)
	}

	if resp.Data == nil || resp.Data.TaskID == "" {
		return "", fmt.Errorf("%s response missing task_id", s.displayName)
	}

	return resp.Data.TaskID, nil
}

func (s *klingTaskService) getTask(taskID string) (*aivideokling.TaskInfo, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task id is required")
	}

	httpReq, err := http.NewRequest(http.MethodGet, s.client.baseURL+fmt.Sprintf(s.queryPathFormat, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("http request creation error: %w", err)
	}

	respBody, err := s.client.do(httpReq, nil)
	if err != nil {
		return nil, err
	}

	var resp klingTaskQueryResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}
	resp.Raw = append([]byte(nil), respBody...)
	if resp.Code != 0 {
		return nil, fmt.Errorf("%s api error code %d: %s", s.displayName, resp.Code, resp.Message)
	}

	if resp.Data == nil || resp.Data.TaskID == "" {
		return nil, fmt.Errorf("%s response missing task data", s.displayName)
	}

	return convertKlingTaskInfo(&resp), nil
}

func convertKlingTaskInfo(resp *klingTaskQueryResponse) *aivideokling.TaskInfo {
	if resp == nil || resp.Data == nil {
		return nil
	}

	info := &aivideokling.TaskInfo{
		TaskID:        resp.Data.TaskID,
		Status:        aivideokling.ConvertTaskStatus(resp.Data.TaskStatus),
		StatusMessage: firstNonEmpty(resp.Data.TaskStatusMsg, resp.Message),
		CreateTime:    unixMillisToSeconds(resp.Data.CreatedAt),
		UpdateTime:    unixMillisToSeconds(resp.Data.UpdatedAt),
		Raw:           append([]byte(nil), resp.Raw...),
	}

	if resp.Data.TaskResult != nil {
		if len(resp.Data.TaskResult.Videos) > 0 {
			info.Videos = make([]aivideokling.VideoResult, 0, len(resp.Data.TaskResult.Videos))
			for _, video := range resp.Data.TaskResult.Videos {
				info.Videos = append(info.Videos, aivideokling.VideoResult{
					ID:       video.ID,
					URL:      video.URL,
					Duration: video.Duration,
				})
			}
		}

		if len(resp.Data.TaskResult.Images) > 0 {
			info.Images = make([]aivideokling.ImageResult, 0, len(resp.Data.TaskResult.Images))
			for _, image := range resp.Data.TaskResult.Images {
				info.Images = append(info.Images, aivideokling.ImageResult{
					Index: image.Index,
					URL:   image.URL,
				})
			}
		}
	}

	return info
}

func unixMillisToSeconds(ms int64) int32 {
	if ms <= 0 {
		return 0
	}
	return int32(time.UnixMilli(ms).Unix())
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
