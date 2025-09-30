package replicate

import (
	"github.com/QingsiLiu/baseComponents/service/text2image"
	"testing"
)

func TestFluxSchnellService_TaskRun(t *testing.T) {
	service := NewFluxSchnellServiceWithKey("")
	req := &text2image.Text2ImageTaskRunReq{
		Prompt:            "a cat jumping in the air to catch a bird",
		NumOutputs:        1,
	}

	// 注意：这个测试需要有效的 API 密钥才能真正运行
	// 在实际环境中，你可能需要模拟 HTTP 请求
	_, err := service.TaskRun(req)
	if err != nil {
		t.Logf("TaskRun() error = %v (expected without valid API key)", err)
	}
}

func TestFluxSchnellService_TaskCancel(t *testing.T) {
	service := NewFluxSchnellService()
	taskId := "test-task-id"

	// 注意：这个测试需要有效的 API 密钥才能真正运行
	err := service.TaskCancel(taskId)
	if err != nil {
		t.Logf("TaskCancel() error = %v (expected without valid API key)", err)
	}
}

func TestFluxSchnellService_TaskGet(t *testing.T) {
	service := NewFluxSchnellService()
	taskId := "test-task-id"

	// 注意：这个测试需要有效的 API 密钥才能真正运行
	_, err := service.TaskGet(taskId)
	if err != nil {
		t.Logf("TaskGet() error = %v (expected without valid API key)", err)
	}
}

func TestFluxSchnellService_TaskList(t *testing.T) {
	service := NewFluxSchnellService()

	// 注意：这个测试需要有效的 API 密钥才能真正运行
	_, err := service.TaskList()
	if err != nil {
		t.Logf("TaskList() error = %v (expected without valid API key)", err)
	}
}

func TestFluxSchnellService_convertToTaskInfo(t *testing.T) {
	service := &FluxSchnellService{}
	resp := &PredictionResponse{
		ID:          "test-id",
		Status:      "succeeded",
		CreatedAt:   "2023-01-01T00:00:00Z",
		StartedAt:   "2023-01-01T00:01:00Z",
		CompletedAt: "2023-01-01T00:02:00Z",
		Output:      []interface{}{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
		Metrics: struct {
			PredictTime float64 `json:"predict_time,omitempty"`
		}{
			PredictTime: 120.5,
		},
	}

	taskInfo := service.convertToTaskInfo(resp)

	if taskInfo.TaskId != resp.ID {
		t.Errorf("TaskId = %v, want %v", taskInfo.TaskId, resp.ID)
	}
	if taskInfo.Status != ConvertStatusToInt(resp.Status) {
		t.Errorf("Status = %v, want %v", taskInfo.Status, ConvertStatusToInt(resp.Status))
	}
	if len(taskInfo.Result) != 2 {
		t.Errorf("Result length = %v, want %v", len(taskInfo.Result), 2)
	}
	if taskInfo.Duration != resp.Metrics.PredictTime {
		t.Errorf("Duration = %v, want %v", taskInfo.Duration, resp.Metrics.PredictTime)
	}
	if taskInfo.CreateTime == 0 {
		t.Error("CreateTime should not be 0")
	}
	if taskInfo.UpdateTime == 0 {
		t.Error("UpdateTime should not be 0")
	}
}