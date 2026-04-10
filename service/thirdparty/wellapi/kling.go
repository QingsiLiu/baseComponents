package wellapi

import (
	"fmt"
	"strings"

	aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"
)

// KlingMotionControlService WellAPI Kling 动作控制服务实现。
type KlingMotionControlService struct {
	task *klingTaskService
}

// KlingEffectsService WellAPI Kling 视频特效服务实现。
type KlingEffectsService struct {
	task *klingTaskService
}

func NewKlingMotionControlService() aivideokling.MotionControlService {
	return &KlingMotionControlService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingMotionControl,
			PathKlingMotionControl,
			PathKlingMotionControlGetFmt,
			"WellAPI Kling Motion Control",
			NewClient(),
		),
	}
}

func NewKlingMotionControlServiceWithKey(apiKey string) aivideokling.MotionControlService {
	return &KlingMotionControlService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingMotionControl,
			PathKlingMotionControl,
			PathKlingMotionControlGetFmt,
			"WellAPI Kling Motion Control",
			NewClientWithKey(apiKey),
		),
	}
}

func NewKlingEffectsService() aivideokling.EffectsService {
	return &KlingEffectsService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingEffects,
			PathKlingEffects,
			PathKlingEffectsGetFmt,
			"WellAPI Kling Effects",
			NewClient(),
		),
	}
}

func NewKlingEffectsServiceWithKey(apiKey string) aivideokling.EffectsService {
	return &KlingEffectsService{
		task: newKlingTaskService(
			aivideokling.SourceWellAPIKlingEffects,
			PathKlingEffects,
			PathKlingEffectsGetFmt,
			"WellAPI Kling Effects",
			NewClientWithKey(apiKey),
		),
	}
}

func (s *KlingMotionControlService) Source() string {
	return s.task.Source()
}

func (s *KlingMotionControlService) TaskRun(req *aivideokling.KlingMotionControlTaskRunReq) (string, error) {
	if req == nil {
		return "", fmt.Errorf("request is nil")
	}
	if strings.TrimSpace(req.ImageURL) == "" {
		return "", fmt.Errorf("image_url is required")
	}
	if strings.TrimSpace(req.VideoURL) == "" {
		return "", fmt.Errorf("video_url is required")
	}
	if strings.TrimSpace(req.CharacterOrientation) == "" {
		return "", fmt.Errorf("character_orientation is required")
	}

	payload := &klingMotionControlRequest{
		Prompt:               req.Prompt,
		ImageURL:             req.ImageURL,
		VideoURL:             req.VideoURL,
		CharacterOrientation: req.CharacterOrientation,
		Mode:                 defaultMode(req.Mode),
		CallbackURL:          req.CallbackURL,
		ExternalTaskID:       req.ExternalTaskID,
	}
	if req.KeepOriginalSound != nil {
		payload.KeepOriginalSound = boolToYesNo(*req.KeepOriginalSound)
	}

	return s.task.createTask(payload)
}

func (s *KlingMotionControlService) TaskGet(taskID string) (*aivideokling.TaskInfo, error) {
	return s.task.getTask(taskID)
}

func (s *KlingEffectsService) Source() string {
	return s.task.Source()
}

func (s *KlingEffectsService) TaskRun(req *aivideokling.KlingEffectsTaskRunReq) (string, error) {
	if req == nil {
		return "", fmt.Errorf("request is nil")
	}
	if strings.TrimSpace(req.EffectScene) == "" {
		return "", fmt.Errorf("effect_scene is required")
	}

	payload := &klingEffectsRequest{
		EffectScene:    req.EffectScene,
		Input:          buildKlingEffectsInput(req.Input),
		CallbackURL:    req.CallbackURL,
		ExternalTaskID: req.ExternalTaskID,
	}

	return s.task.createTask(payload)
}

func (s *KlingEffectsService) TaskGet(taskID string) (*aivideokling.TaskInfo, error) {
	return s.task.getTask(taskID)
}

func buildKlingEffectsInput(input aivideokling.KlingEffectsInput) map[string]any {
	var result map[string]any

	setIfNotEmpty := func(key, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		if result == nil {
			result = make(map[string]any)
		}
		result[key] = value
	}

	setIfNotEmpty("model_name", input.ModelName)
	setIfNotEmpty("mode", input.Mode)
	setIfNotEmpty("duration", input.Duration)
	setIfNotEmpty("image", input.Image)
	if len(input.Images) > 0 {
		if result == nil {
			result = make(map[string]any)
		}
		result["images"] = append([]string(nil), input.Images...)
	}

	if len(input.Extra) > 0 {
		if result == nil {
			result = make(map[string]any)
		}
		for key, value := range input.Extra {
			if _, exists := result[key]; exists {
				continue
			}
			result[key] = value
		}
	}

	return result
}

func defaultMode(mode string) string {
	if strings.TrimSpace(mode) == "" {
		return "std"
	}
	return mode
}

func boolToYesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
