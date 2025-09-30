package replicate

import (
	"github.com/QingsiLiu/baseComponents/service/text2image"
	"testing"
)

// TestFlux1DevServiceTaskRun 测试实际的任务提交功能
func TestFlux1DevServiceTaskRun(t *testing.T) {
	service := NewFlux1DevServiceWithKey("")
	req := &text2image.Text2ImageTaskRunReq{
		Prompt:            "a professional logo design for the brand 'coastal hues'. the creative concept features a watercolor logo illustration and the brand name.",
		Seed:              -1,
		Guidance:          3.5,
		ImageSize:         1024,
		SpeedMode:         "Lightly Juiced 🍊 (more consistent)",
		AspectRatio:       "1:1",
		OutputFormat:      "jpg",
		OutputQuality:     80,
		NumInferenceSteps: 28,
	}

	taskId, err := service.TaskRun(req)
	if err != nil {
		t.Fatalf("TaskRun failed: %v", err)
	}

	if taskId == "" {
		t.Fatal("TaskRun returned empty task ID")
	}

	t.Logf("Task submitted successfully with ID: %s", taskId)

	// 测试获取任务状态
	task, err := service.TaskGet(taskId)
	if err != nil {
		t.Fatalf("TaskGet failed: %v", err)
	}

	if task.TaskId != taskId {
		t.Errorf("Expected task ID %s, got %s", taskId, task.TaskId)
	}

	// 检查任务状态是否有效
	validStatuses := []int32{
		text2image.TaskStatusPending,
		text2image.TaskStatusRunning,
		text2image.TaskStatusCompleted,
		text2image.TaskStatusFailed,
	}

	statusValid := false
	for _, status := range validStatuses {
		if task.Status == status {
			statusValid = true
			break
		}
	}

	if !statusValid {
		t.Errorf("Invalid task status: %d", task.Status)
	}

	t.Logf("Task status: %d", task.Status)

	// 如果任务已完成，检查结果
	if task.Status == text2image.TaskStatusCompleted {
		if len(task.Result) == 0 {
			t.Error("Completed task should have results")
		} else {
			t.Logf("Task id: %s, duration: %f", task.TaskId, task.Duration)
			t.Logf("Generated images: %v", task.Result)
		}
	}
}

// TestFlux1DevServiceTaskCancel 测试任务取消功能
func TestFlux1DevServiceTaskCancel(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewFlux1DevService()

	// 首先创建一个任务
	req := &text2image.Text2ImageTaskRunReq{
		Prompt: "test image for cancellation",
	}

	taskId, err := service.TaskRun(req)
	if err != nil {
		t.Fatalf("TaskRun failed: %v", err)
	}

	// 尝试取消任务
	err = service.TaskCancel(taskId)
	if err != nil {
		t.Logf("TaskCancel failed (this may be expected if task completed quickly): %v", err)
	} else {
		t.Logf("Task %s cancelled successfully", taskId)
	}
}

// TestFlux1DevServiceTaskList 测试任务列表功能
func TestFlux1DevServiceTaskList(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewFlux1DevService()

	tasks, err := service.TaskList()
	if err != nil {
		t.Fatalf("TaskList failed: %v", err)
	}

	t.Logf("Found %d Flux1Dev tasks", len(tasks))

	// 检查返回的任务格式
	for i, task := range tasks {
		if i >= 3 { // 只检查前3个任务
			break
		}

		if task.TaskId == "" {
			t.Errorf("Task %d has empty TaskId", i)
		}

		if task.Status < 0 || task.Status > 4 {
			t.Errorf("Task %d has invalid status: %d", i, task.Status)
		}

		t.Logf("Task %d: ID=%s, Status=%d", i, task.TaskId, task.Status)
	}
}
