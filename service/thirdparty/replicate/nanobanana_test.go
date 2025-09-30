package replicate

import (
	"fmt"
	"math/rand"
	"github.com/QingsiLiu/baseComponents/service/image2image"
	"testing"
	"time"
)

// TestNanoBananaService_TaskRun 测试任务提交功能
func TestNanoBananaService_TaskRun(t *testing.T) {
	service := NewNanoBananaServiceWithKey("")
	req := &image2image.Image2ImageTaskRunReq{
		Prompt: "Make the sheets in the style of the logo. Make the scene natural.",
		ImageInputs: []string{
			"https://replicate.delivery/pbxt/NbYIclp4A5HWLsJ8lF5KgiYSNaLBBT1jUcYcHYQmN1uy5OnN/tmpcqc07f_q.png",
			"https://replicate.delivery/pbxt/NbYId45yH8s04sptdtPcGqFIhV7zS5GTcdS3TtNliyTAoYPO/Screenshot%202025-08-26%20at%205.30.12%E2%80%AFPM.png",
		},
		OutputFormat: "jpg",
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

// TestNanoBananaService_TaskCancel 测试任务取消功能
func TestNanoBananaService_TaskCancel(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewNanoBananaService()

	// 首先创建一个任务
	req := &image2image.Image2ImageTaskRunReq{
		Prompt: "Create a complex tattoo design with multiple elements",
		ImageInputs: []string{
			"https://example.com/complex.jpg",
		},
		OutputFormat: "jpg",
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

// TestNanoBananaService_TaskList 测试任务列表功能
func TestNanoBananaService_TaskList(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewNanoBananaService()

	tasks, err := service.TaskList()
	if err != nil {
		t.Fatalf("TaskList failed: %v", err)
	}

	t.Logf("Found %d NanoBanana tasks", len(tasks))

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

// TestNanoBananaService_TaskGetAndWait 测试任务提交并等待完成
func TestNanoBananaService_TaskGetAndWait(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewNanoBananaService()

	// 测试数据
	testPrompts := []struct {
		Prompt      string
		ImageInputs []string
		Description string
	}{
		{
			Prompt: "Create a traditional tattoo design with roses and skulls",
			ImageInputs: []string{
				"https://example.com/rose.jpg",
				"https://example.com/skull.jpg",
			},
			Description: "Traditional Rose and Skull",
		},
		{
			Prompt: "Modern minimalist tattoo with geometric patterns",
			ImageInputs: []string{
				"https://example.com/geometric.jpg",
			},
			Description: "Minimalist Geometric",
		},
	}

	// 结果结构
	type Result struct {
		Description string   `json:"description"`
		Prompt      string   `json:"prompt"`
		Images      []string `json:"images"`
		Error       string   `json:"error,omitempty"`
	}

	results := make([]Result, 0, len(testPrompts))

	// 循环处理每个 prompt
	for i, promptData := range testPrompts {
		t.Logf("处理第 %d/%d 个 prompt: %s", i+1, len(testPrompts), promptData.Description)

		req := &image2image.Image2ImageTaskRunReq{
			Prompt:       promptData.Prompt,
			ImageInputs:  promptData.ImageInputs,
			OutputFormat: "jpg",
		}

		result := Result{
			Description: promptData.Description,
			Prompt:      promptData.Prompt,
			Images:      []string{},
		}

		// 运行任务
		taskID, err := service.TaskRun(req)
		if err != nil {
			t.Logf("无法创建测试任务: %v", err)
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		// 等待任务完成，最多等待2分钟
		var task *image2image.Image2ImageTaskInfo
		timeout := time.After(2 * time.Minute)
		ticker := time.NewTicker(1 * time.Second)

		func() {
			defer ticker.Stop()
			for {
				select {
				case <-timeout:
					result.Error = "等待任务完成超时"
					return
				case <-ticker.C:
					task, err = service.TaskGet(taskID)
					if err != nil {
						result.Error = fmt.Sprintf("获取任务状态失败: %v", err)
						return
					}

					t.Logf("任务状态: %d", task.Status)

					if task.Status == image2image.TaskStatusCompleted || task.Status == image2image.TaskStatusFailed || task.Status == image2image.TaskStatusCanceled {
						return
					}
				}
			}
		}()

		// 处理任务结果
		if task != nil {
			if task.Status == image2image.TaskStatusCompleted {
				result.Images = task.Result
				t.Logf("任务完成，生成了 %d 张图片", len(task.Result))
			} else {
				result.Error = fmt.Sprintf("任务失败，状态: %d", task.Status)
			}
		}

		results = append(results, result)

		// 每次任务之间间隔一定时间，避免API限制
		if i < len(testPrompts)-1 {
			// 生成随机间隔时间，范围在3-10秒之间
			minSleep := 3
			maxSleep := 10
			randSleep := minSleep + rand.Intn(maxSleep-minSleep+1)
			sleepTime := time.Duration(randSleep) * time.Second
			t.Logf("等待随机时间 %v 后继续下一个任务...", sleepTime)
			time.Sleep(sleepTime)
		}
	}

	// 输出最终结果
	t.Logf("测试完成，处理了 %d 个任务", len(results))
	for i, result := range results {
		if result.Error != "" {
			t.Logf("任务 %d (%s) 失败: %s", i+1, result.Description, result.Error)
		} else {
			t.Logf("任务 %d (%s) 成功，生成了 %d 张图片", i+1, result.Description, len(result.Images))
		}
	}
}

// TestNanoBananaService_Integration 集成测试
func TestNanoBananaService_Integration(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewNanoBananaService()

	// 测试完整的工作流程
	req := &image2image.Image2ImageTaskRunReq{
		Prompt: "Create a beautiful tattoo design combining nature and geometric elements",
		ImageInputs: []string{
			"https://example.com/nature.jpg",
			"https://example.com/geometric.jpg",
		},
		OutputFormat: "jpg",
	}

	// 1. 提交任务
	taskId, err := service.TaskRun(req)
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	t.Logf("Task submitted: %s", taskId)

	// 2. 获取任务状态
	task, err := service.TaskGet(taskId)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	t.Logf("Task status: %d", task.Status)

	// 3. 列出任务
	tasks, err := service.TaskList()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	t.Logf("Total tasks: %d", len(tasks))

	// 验证我们的任务在列表中
	found := false
	for _, listTask := range tasks {
		if listTask.TaskId == taskId {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Submitted task %s not found in task list", taskId)
	}

	t.Log("Integration test completed successfully")
}
