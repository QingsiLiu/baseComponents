package text2image

// Text2ImageService 文本生成图像服务接口
type Text2ImageService interface {
	Source() string
	TaskRun(req *Text2ImageTaskRunReq) (taskId string, err error)
	TaskGet(taskId string) (task *Text2ImageTaskInfo, err error)
	TaskCancel(taskId string) error
	TaskList() ([]*Text2ImageTaskInfo, error)
}

// Text2ImageTaskReq 文本生成图像任务请求
// 常见文生图参数，未来根据实际业务持续补充
type Text2ImageTaskRunReq struct {
	Model                string  `json:"model"`
	Style                string  `json:"style"`
	ImageWidth           int     `json:"image_width"`
	ImageHeight          int     `json:"image_height"`
	ImageSize            int     `json:"image_size"`
	Seed                 int     `json:"seed"`
	Prompt               string  `json:"prompt"`
	NegativePrompt       string  `json:"negative_prompt"`
	Guidance             float64 `json:"guidance"`
	Megapixels           int     `json:"megapixels"`
	SpeedMode            string  `json:"speed_mode"`
	NumOutputs           int     `json:"num_outputs"`
	AspectRatio          string  `json:"aspect_ratio"`
	OutputFormat         string  `json:"output_format"`
	OutputQuality        int     `json:"output_quality"`
	NumInferenceSteps    int     `json:"num_inference_steps"`
	DisableSafetyChecker bool    `json:"disable_safety_checker"`
	Debug                bool    `json:"debug"`
}

// Text2ImageTaskInfo 文本生成图像任务信息
type Text2ImageTaskInfo struct {
	TaskId     string   `json:"task_id"`     // 任务ID
	Status     int32    `json:"status"`      // 任务状态
	Result     []string `json:"result"`      // 任务结果
	Duration   float64  `json:"duration"`    // 任务执行时间
	CreateTime int32    `json:"create_time"` // 创建时间
	UpdateTime int32    `json:"update_time"` // 更新时间
}

func (t *Text2ImageTaskInfo) GetStatusName(status int32) string {
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
