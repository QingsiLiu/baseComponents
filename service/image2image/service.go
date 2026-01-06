package image2image

// Image2ImageService 图生图服务接口
type Image2ImageService interface {
	Source() string
	TaskRun(req *Image2ImageTaskRunReq) (taskId string, err error)
	TaskGet(taskId string) (task *Image2ImageTaskInfo, err error)
	TaskCancel(taskId string) error
	TaskList() ([]*Image2ImageTaskInfo, error)
}

// Image2ImageTaskRunReq 图生图任务请求
type Image2ImageTaskRunReq struct {
	Model             string   `json:"model"`
	ImageInputs       []string `json:"image_inputs"`
	Seed              int      `json:"seed"`
	Prompt            string   `json:"prompt"`
	NegativePrompt    string   `json:"negative_prompt"`
	Strength          float64  `json:"strength"`
	GuidanceScale     int      `json:"guidance_scale"`
	OutputImageSize   string   `json:"output_image_size"` // 输出图片尺寸，如 512x512；或比例，例如 3:4
	OutputFormat      string   `json:"output_format"`     // 输出格式，如 jpg, png
	OutputQuality     int      `json:"output_quality"`
	Resolution        string   `json:"resolution"`
	NumInferenceSteps int      `json:"num_inference_steps"`
	Debug             bool     `json:"debug"`
}

// Image2ImageTaskInfo 图生图任务信息
type Image2ImageTaskInfo struct {
	TaskId     string   `json:"task_id"`     // 任务ID
	Status     int32    `json:"status"`      // 任务状态
	Result     []string `json:"result"`      // 任务结果
	Duration   float64  `json:"duration"`    // 任务执行时间
	CreateTime int32    `json:"create_time"` // 创建时间
	UpdateTime int32    `json:"update_time"` // 更新时间
}

func (t *Image2ImageTaskInfo) GetStatusName(status int32) string {
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
