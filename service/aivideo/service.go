package aivideo

// AIVideoService AI视频生成服务接口
type AIVideoService interface {
	Source() string
	TaskRun(req *AIVideoTaskRunReq) (taskId string, err error)
	TaskGet(taskId string) (task *AIVideoTaskInfo, err error)
	TaskCancel(taskId string) error
	TaskList() ([]*AIVideoTaskInfo, error)
}

// AIVideoTaskRunReq AI视频生成任务请求
type AIVideoTaskRunReq struct {
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`                    // 文本描述（必填）
	Quality        string  `json:"quality"`                   // 视频分辨率: 360p, 540p, 720p, 1080p
	Duration       int     `json:"duration"`                  // 视频时长（秒）: 5, 8
	AspectRatio    string  `json:"aspect_ratio"`              // 视频比例: 16:9, 9:16, 1:1
	Image          string  `json:"image"`                     // 首帧图片URL（可选）
	LastFrameImage string  `json:"last_frame_image"`          // 末帧图片URL（可选，需配合image使用）
	Effect         string  `json:"effect"`                    // 特殊效果（可选）
	NegativePrompt string  `json:"negative_prompt"`           // 负面提示词
	Seed           int     `json:"seed"`                      // 随机种子
	Debug          bool    `json:"debug"`
}

// AIVideoTaskInfo AI视频生成任务信息
type AIVideoTaskInfo struct {
	TaskId     string   `json:"task_id"`     // 任务ID
	Status     int32    `json:"status"`      // 任务状态
	Result     []string `json:"result"`      // 任务结果（视频URL列表）
	Duration   float64  `json:"duration"`    // 任务执行时间
	CreateTime int32    `json:"create_time"` // 创建时间
	UpdateTime int32    `json:"update_time"` // 更新时间
}

func (t *AIVideoTaskInfo) GetStatusName(status int32) string {
	switch status {
	case TaskStatusPending:
		return "pending"
	case TaskStatusRunning:
		return "running"
	case TaskStatusCompleted:
		return "completed"
	case TaskStatusCanceled:
		return "canceled"
	case TaskStatusFailed:
		return "failed"
	}
	return "unknown"
}
