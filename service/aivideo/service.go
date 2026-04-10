package aivideo

import "github.com/QingsiLiu/baseComponents/internal/taskstatus"

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
	Model         string `json:"model"`
	Prompt        string `json:"prompt"`         // 文本描述（必填）
	Resolution    string `json:"resolution"`     // 推荐的视频分辨率字段
	Duration      int    `json:"duration"`       // 视频时长（秒）: 5, 8
	AspectRatio   string `json:"aspect_ratio"`   // 视频比例: 16:9, 9:16, 1:1
	Image         string `json:"image"`          // 首帧图片URL（可选）
	CallbackURL   string `json:"callback_url"`   // 任务完成回调地址
	GenerateAudio *bool  `json:"generate_audio"` // 是否生成音频
	Debug         bool   `json:"debug"`

	// 以下字段为 provider-specific 兼容字段。
	// 第一轮结构重构后，后续新增模型默认不再继续向这里增加私有字段。
	Mode               string   `json:"mode"`                 // 兼容字段：模型模式开关，例如 std/pro
	Quality            string   `json:"quality"`              // 兼容字段：旧调用的视频分辨率字段
	LastFrameImage     string   `json:"last_frame_image"`     // 末帧图片URL（可选，需配合image使用）
	ReferenceImageURLs []string `json:"reference_image_urls"` // 多张参考图片URL
	ReferenceVideoURLs []string `json:"reference_video_urls"` // 参考视频URL
	ReferenceAudioURLs []string `json:"reference_audio_urls"` // 参考音频URL
	ReturnLastFrame    *bool    `json:"return_last_frame"`    // 是否返回末帧图像
	FixedLens          *bool    `json:"fixed_lens"`           // 是否固定镜头
	WebSearch          *bool    `json:"web_search"`           // 是否启用联网搜索
	NSFWChecker        *bool    `json:"nsfw_checker"`         // 是否启用 NSFW 检查
	Effect             string   `json:"effect"`               // 特殊效果（可选）
	NegativePrompt     string   `json:"negative_prompt"`      // 负面提示词
	Seed               int      `json:"seed"`                 // 随机种子
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
	return taskstatus.Name(status)
}
