package replicate

import (
	"fmt"
	"github.com/QingsiLiu/baseComponents/service/image2image"
	"testing"
	"time"
)

// TestControlNetService_TaskRun 测试任务提交功能
func TestControlNetService_TaskRun(t *testing.T) {
	service := NewControlNetServiceWithKey("")
	req := &image2image.Image2ImageTaskRunReq{
		Prompt:            "Create a photorealistic interior image of a Coastal style living room. Coastal design evokes a relaxed, beachside feel through a light and airy color palette, primarily featuring shades of white, blue, and sand. It emphasizes natural light and incorporates organic materials like jute, rattan, and weathered wood.",
		ImageInputs:       []string{"https://example.com/test-image.jpg"},
		Seed:              42,
		GuidanceScale:     7,
		NumInferenceSteps: 20,
		OutputFormat:      "jpg",
		OutputQuality:     80,
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
}

// TestControlNetService_convertToControlNetInput 测试请求转换功能
func TestControlNetService_convertToControlNetInput(t *testing.T) {
	service := &ControlNetService{}

	req := &image2image.Image2ImageTaskRunReq{
		Prompt:            "test prompt",
		ImageInputs:       []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
		GuidanceScale:     15,
		NumInferenceSteps: 30,
	}

	input := service.convertToControlNetInput(req)

	if input.Prompt != req.Prompt {
		t.Errorf("Expected prompt %s, got %s", req.Prompt, input.Prompt)
	}

	if input.Image != req.ImageInputs[0] {
		t.Errorf("Expected image %s, got %s", req.ImageInputs[0], input.Image)
	}

	if input.Scale != req.GuidanceScale {
		t.Errorf("Expected scale %d, got %d", req.GuidanceScale, input.Scale)
	}

	if input.DDimSteps != req.NumInferenceSteps {
		t.Errorf("Expected ddim_steps %d, got %d", req.NumInferenceSteps, input.DDimSteps)
	}

	if input.ImageResolution != "768" {
		t.Errorf("Expected image_resolution 768, got %s", input.ImageResolution)
	}
}

// TestControlNetService_convertToControlNetInput_Defaults 测试默认值
func TestControlNetService_convertToControlNetInput_Defaults(t *testing.T) {
	service := &ControlNetService{}

	req := &image2image.Image2ImageTaskRunReq{
		Prompt: "test prompt with defaults",
	}

	input := service.convertToControlNetInput(req)

	if input.DDimSteps != 40 {
		t.Errorf("Expected default ddim_steps 40, got %d", input.DDimSteps)
	}

	if input.ImageResolution != "768" {
		t.Errorf("Expected default image_resolution 768, got %s", input.ImageResolution)
	}

	if input.Scale != 25 {
		t.Errorf("Expected default scale 25, got %d", input.Scale)
	}

	if input.Image != "" {
		t.Errorf("Expected empty image when no inputs provided, got %s", input.Image)
	}
}

// TestControlNetService_TaskCancel 测试任务取消功能
func TestControlNetService_TaskCancel(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewControlNetService()

	// 首先创建一个任务
	req := &image2image.Image2ImageTaskRunReq{
		Prompt:      "test image for cancellation",
		ImageInputs: []string{"https://example.com/test-image.jpg"},
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

// TestControlNetService_TaskList 测试任务列表功能
func TestControlNetService_TaskList(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewControlNetService()

	tasks, err := service.TaskList()
	if err != nil {
		t.Fatalf("TaskList failed: %v", err)
	}

	t.Logf("Found %d ControlNet tasks", len(tasks))

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

// TestControlNetService_TaskGetAndWait 测试任务提交并等待完成
func TestControlNetService_TaskGetAndWait(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewControlNetService()

	// 测试数据
	testPrompts := []struct {
		Prompt      string
		ImageInput  string
		Description string
	}{
		{
			Prompt:      "modern minimalist living room with clean lines and neutral colors",
			ImageInput:  "https://example.com/living-room.jpg",
			Description: "Modern Living Room",
		},
		{
			Prompt:      "cozy bedroom with warm lighting and wooden furniture",
			ImageInput:  "https://example.com/bedroom.jpg",
			Description: "Cozy Bedroom",
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
			Prompt:      promptData.Prompt,
			ImageInputs: []string{promptData.ImageInput},
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

// TestControlNetService_Integration 集成测试
func TestControlNetService_Integration(t *testing.T) {
	// 跳过需要真实API token的测试
	if GetAPIToken() == "" {
		t.Skip("Skipping test: REPLICATE_TOKEN environment variable not set")
	}

	service := NewControlNetService()

	// 测试完整的工作流程
	req := &image2image.Image2ImageTaskRunReq{
		Prompt:            "elegant bathroom with spa-like features and natural stone",
		ImageInputs:       []string{"https://example.com/bathroom.jpg"},
		GuidanceScale:     7,
		NumInferenceSteps: 20,
		OutputFormat:      "jpg",
		OutputQuality:     80,
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