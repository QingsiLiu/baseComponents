package kie

import (
	"testing"

	"github.com/QingsiLiu/baseComponents/service/image2image"
)

func TestNanoBananaService_TaskRun(t *testing.T) {
	service := NewNanoBananaServiceWithKey("")
	req := &image2image.Image2ImageTaskRunReq{
		Prompt: "Make the sheets in the style of the logo. Make the scene natural.",
		ImageInputs: []string{
			"https://replicate.delivery/pbxt/NbYIclp4A5HWLsJ8lF5KgiYSNaLBBT1jUcYcHYQmN1uy5OnN/tmpcqc07f_q.png",
			"https://replicate.delivery/pbxt/NbYId45yH8s04sptdtPcGqFIhV7zS5GTcdS3TtNliyTAoYPO/Screenshot%202025-08-26%20at%205.30.12%E2%80%AFPM.png",
		},
		OutputFormat: "jpeg",
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
		image2image.TaskStatusPending,
		image2image.TaskStatusRunning,
		image2image.TaskStatusCompleted,
		image2image.TaskStatusFailed,
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
	t.Logf("Task: %v", task)
}
