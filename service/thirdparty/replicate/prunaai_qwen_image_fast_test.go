package replicate

import (
	"github.com/QingsiLiu/baseComponents/service/text2image"
	"testing"
)

// TestNanoBananaService_Integration 集成测试
func TestPrunaAIQwenImageFastService_Integration(t *testing.T) {
	service := NewPrunaAIQwenImageFastServiceWithKey("")

	// 测试完整的工作流程
	req := &text2image.Text2ImageTaskRunReq{
		Prompt:               "Cherries, carbonated water, macro, professional color grading, clean sharp focus, commercial high quality, magazine winning photography, hyper realistic, uhd, 8K. A fancy typography saying \"QWEN IMAGE FAST\" in the middle of the image.",
		Seed:                 0,
		Guidance:             0.62,
		AspectRatio:          "1:1",
		DisableSafetyChecker: true,
		Debug:                false,
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
	t.Logf("Task status: %v", task)


	t.Log("Integration test completed successfully")
}
