package kie

import (
	"fmt"
	"strings"

	"github.com/QingsiLiu/baseComponents/service/aivideo"
	"github.com/QingsiLiu/baseComponents/utils"
)

type kieAIVideoTaskService struct {
	source      string
	modelName   string
	displayName string
	client      *Client
	buildInput  func(*aivideo.AIVideoTaskRunReq) interface{}
}

func newKieAIVideoTaskService(
	source string,
	modelName string,
	displayName string,
	client *Client,
	buildInput func(*aivideo.AIVideoTaskRunReq) interface{},
) *kieAIVideoTaskService {
	return &kieAIVideoTaskService{
		source:      source,
		modelName:   modelName,
		displayName: displayName,
		client:      client,
		buildInput:  buildInput,
	}
}

func (s *kieAIVideoTaskService) Source() string {
	return s.source
}

func (s *kieAIVideoTaskService) TaskRun(req *aivideo.AIVideoTaskRunReq) (string, error) {
	if req.Debug {
		return "mock_task_id_" + utils.RandomString(5), nil
	}

	resp, err := s.client.CreateTask(s.convertToCreateRequest(req))
	if err != nil {
		return "", err
	}
	if resp.Data == nil || resp.Data.TaskID == "" {
		return "", fmt.Errorf("missing task ID in response")
	}

	return resp.Data.TaskID, nil
}

func (s *kieAIVideoTaskService) TaskGet(taskID string) (*aivideo.AIVideoTaskInfo, error) {
	if strings.HasPrefix(taskID, "mock_task_id_") {
		return &aivideo.AIVideoTaskInfo{
			TaskId:   taskID,
			Status:   aivideo.TaskStatusCompleted,
			Result:   []string{"https://static.aiquickdraw.com/tools/example/1770648690994_jIU8D0i9.mp4"},
			Duration: 5,
		}, nil
	}

	resp, err := s.client.GetTaskRecord(taskID)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("task data is empty")
	}

	return s.convertToTaskInfo(resp.Data), nil
}

func (s *kieAIVideoTaskService) TaskCancel(taskID string) error {
	return fmt.Errorf("task cancellation not supported by KIE %s API", s.displayName)
}

func (s *kieAIVideoTaskService) TaskList() ([]*aivideo.AIVideoTaskInfo, error) {
	return nil, fmt.Errorf("task listing not supported by KIE %s API", s.displayName)
}

func (s *kieAIVideoTaskService) convertToCreateRequest(req *aivideo.AIVideoTaskRunReq) *TaskCreateRequest {
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = s.modelName
	}

	return &TaskCreateRequest{
		Model:       model,
		CallbackURL: req.CallbackURL,
		Input:       s.buildInput(req),
	}
}

func (s *kieAIVideoTaskService) convertToTaskInfo(detail *TaskRecordDetail) *aivideo.AIVideoTaskInfo {
	return &aivideo.AIVideoTaskInfo{
		TaskId:     detail.TaskID,
		Status:     ConvertStateToStatus(detail.State),
		Result:     ParseResultURLs(detail.ResultJSON),
		Duration:   ResolveTaskDuration(detail),
		CreateTime: UnixMillisToSeconds(detail.CreateTime),
		UpdateTime: ResolveTaskUpdateTime(detail),
	}
}

func boolValueOrDefault(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func stringValueOrDefault(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func intValueOrDefault(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}

func resolvePrimaryImageInputs(req *aivideo.AIVideoTaskRunReq) []string {
	if len(req.ReferenceImageURLs) > 0 {
		return req.ReferenceImageURLs
	}
	if req.Image == "" {
		return nil
	}
	return []string{req.Image}
}
