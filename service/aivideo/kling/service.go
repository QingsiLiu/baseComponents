package kling

import (
	"strings"

	"github.com/QingsiLiu/baseComponents/internal/taskstatus"
)

// MotionControlService 可灵动作控制服务接口。
type MotionControlService interface {
	Source() string
	TaskRun(req *KlingMotionControlTaskRunReq) (taskID string, err error)
	TaskGet(taskID string) (*TaskInfo, error)
}

// EffectsService 可灵视频特效服务接口。
type EffectsService interface {
	Source() string
	TaskRun(req *KlingEffectsTaskRunReq) (taskID string, err error)
	TaskGet(taskID string) (*TaskInfo, error)
}

// KlingMotionControlTaskRunReq 动作控制任务请求。
type KlingMotionControlTaskRunReq struct {
	Prompt               string `json:"prompt"`
	ImageURL             string `json:"image_url"`
	VideoURL             string `json:"video_url"`
	KeepOriginalSound    *bool  `json:"keep_original_sound,omitempty"`
	CharacterOrientation string `json:"character_orientation"`
	Mode                 string `json:"mode"`
	CallbackURL          string `json:"callback_url"`
	ExternalTaskID       string `json:"external_task_id"`
}

// KlingEffectsTaskRunReq 视频特效任务请求。
type KlingEffectsTaskRunReq struct {
	EffectScene    string            `json:"effect_scene"`
	Input          KlingEffectsInput `json:"input"`
	CallbackURL    string            `json:"callback_url"`
	ExternalTaskID string            `json:"external_task_id"`
}

// KlingEffectsInput 视频特效输入。
type KlingEffectsInput struct {
	ModelName string         `json:"model_name"`
	Mode      string         `json:"mode"`
	Duration  string         `json:"duration"`
	Image     string         `json:"image"`
	Images    []string       `json:"images"`
	Extra     map[string]any `json:"extra"`
}

// TaskInfo 可灵异步任务统一信息。
type TaskInfo struct {
	TaskID        string        `json:"task_id"`
	Status        int32         `json:"status"`
	StatusMessage string        `json:"status_message"`
	CreateTime    int32         `json:"create_time"`
	UpdateTime    int32         `json:"update_time"`
	Videos        []VideoResult `json:"videos"`
	Images        []ImageResult `json:"images"`
	Raw           []byte        `json:"raw"`
}

// VideoResult 任务输出视频。
type VideoResult struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Duration string `json:"duration"`
}

// ImageResult 任务输出图片。
type ImageResult struct {
	Index int    `json:"index"`
	URL   string `json:"url"`
}

func (t *TaskInfo) GetStatusName(status int32) string {
	return taskstatus.Name(status)
}

// ConvertTaskStatus 将 WellAPI Kling 任务状态转换为统一状态码。
func ConvertTaskStatus(status string) int32 {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "submitted":
		return TaskStatusPending
	case "processing":
		return TaskStatusRunning
	case "succeed":
		return TaskStatusCompleted
	case "failed":
		return TaskStatusFailed
	default:
		return TaskStatusPending
	}
}
