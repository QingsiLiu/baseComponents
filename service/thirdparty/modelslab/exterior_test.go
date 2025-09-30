package modelslab

import (
	"github.com/QingsiLiu/baseComponents/service/image2image"
	"testing"
	"time"
)

func TestExteriorTaskRun(t *testing.T) {
	service := NewExteriorServiceWithKey("")
	req := &image2image.Image2ImageTaskRunReq{
		Prompt: `Redesign only the interior of the bedroom in the provided image into a glamorous, luxurious, and symmetrical Art Deco Style room. Highest Priority: The primary task is to introduce a complete and stylistically perfect Art Deco bed setup with realistic and proportionate dimensions as the room's focal point. THE BED: The Glamorous Centerpiece
Frame and Headboard: The bed must be the glamorous centerpiece, featuring a tall, symmetrical headboard with bold geometric shapes, sunburst patterns, or stepped forms. The frame is made of high-gloss lacquer or rich, dark wood with polished brass or chrome inlays.
Bedding: Bedding is luxurious silk or satin in a solid, jewel-toned color, looking smooth and opulent.
Placement and Size: The bed is perfectly centered to emphasize symmetry. Its size must be proportionate to the room. Surrounding Elements:
Furniture: Use symmetrical nightstands and a vanity with lacquered wood, mirrored surfaces, and chrome/brass hardware.
Materials & Color: The color palette is high-contrast: black, white, gold, silver, with rich jewel-toned accents (emerald green, sapphire blue).
Rug: A rug with a bold geometric pattern (chevrons, zig-zags, sunbursts).
Lighting: Stylized, geometric lighting fixtures made of brass or chrome.
Decor: Decor includes glamorous, multi-faceted mirrors and stylized art prints. CRITICAL INSTRUCTION:
DO NOT alter the architecture, structure, or permanent fixtures of the room. Final Image Style:
Render as photorealistic, with stylized, dramatic lighting that highlights the glossy surfaces and metallic details, creating a luxurious and glamorous atmosphere.`,
		ImageInputs: []string{"https://bbyy-homedesign.s3.us-east-1.amazonaws.com/input/175768613067l"},
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

	t.Logf("Task status: %s", task.GetStatusName(task.Status))

	// 如果任务已完成，检查结果
	if task.Status == image2image.TaskStatusCompleted {
		if len(task.Result) == 0 {
			t.Error("Completed task should have results")
		} else {
			t.Logf("Task id: %s, duration: %f", task.TaskId, task.Duration)
			t.Logf("Generated images: %v", task.Result)
		}
	}
}
