package modelslab

import (
	"github.com/QingsiLiu/baseComponents/service/text2image"
	"testing"
	"time"
)

func TestFluxTaskRun(t *testing.T) {
	service := NewFluxServiceWithKey("")
	req := &text2image.Text2ImageTaskRunReq{
		Prompt: "A simple red circle on white background",
	}

	taskId, err := service.TaskRun(req)
	if err != nil {
		t.Fatalf("TaskRun failed: %v", err)
	}

	if taskId == "" {
		t.Fatal("TaskRun returned empty task ID")
	}

	t.Logf("Task submitted successfully with ID: %s", taskId)

	// 等待一段时间让任务开始处理
	time.Sleep(2 * time.Second)

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

	t.Logf("Task status: %s", task.GetStatusName(task.Status))

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
