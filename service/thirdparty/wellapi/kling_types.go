package wellapi

import "encoding/json"

type klingTaskCreateResponse struct {
	Code      int            `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Data      *klingTaskData `json:"data,omitempty"`
	Raw       []byte         `json:"-"`
}

type klingTaskQueryResponse struct {
	Code      int            `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Data      *klingTaskData `json:"data,omitempty"`
	Raw       []byte         `json:"-"`
}

type klingTaskData struct {
	TaskID        string           `json:"task_id"`
	TaskStatus    string           `json:"task_status"`
	TaskStatusMsg string           `json:"task_status_msg"`
	TaskInfo      json.RawMessage  `json:"task_info,omitempty"`
	TaskResult    *klingTaskResult `json:"task_result,omitempty"`
	CreatedAt     int64            `json:"created_at"`
	UpdatedAt     int64            `json:"updated_at"`
}

type klingTaskResult struct {
	Images []klingTaskImage `json:"images,omitempty"`
	Videos []klingTaskVideo `json:"videos,omitempty"`
}

type klingTaskImage struct {
	Index int    `json:"index"`
	URL   string `json:"url"`
}

type klingTaskVideo struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Duration string `json:"duration"`
}

type klingMotionControlRequest struct {
	Prompt               string `json:"prompt,omitempty"`
	ImageURL             string `json:"image_url"`
	VideoURL             string `json:"video_url"`
	KeepOriginalSound    string `json:"keep_original_sound,omitempty"`
	CharacterOrientation string `json:"character_orientation"`
	Mode                 string `json:"mode"`
	CallbackURL          string `json:"callback_url,omitempty"`
	ExternalTaskID       string `json:"external_task_id,omitempty"`
}

type klingEffectsRequest struct {
	EffectScene    string         `json:"effect_scene"`
	Input          map[string]any `json:"input,omitempty"`
	CallbackURL    string         `json:"callback_url,omitempty"`
	ExternalTaskID string         `json:"external_task_id,omitempty"`
}
