package kling

import (
	aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"
	"github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

type Config struct {
	APIKey string
}

type MotionControlService = aivideokling.MotionControlService
type EffectsService = aivideokling.EffectsService
type MotionControlRequest = aivideokling.KlingMotionControlTaskRunReq
type EffectsRequest = aivideokling.KlingEffectsTaskRunReq
type EffectsInput = aivideokling.KlingEffectsInput
type TaskInfo = aivideokling.TaskInfo
type VideoResult = aivideokling.VideoResult
type ImageResult = aivideokling.ImageResult

func NewMotionControlService(cfg Config) MotionControlService {
	if cfg.APIKey != "" {
		return wellapi.NewKlingMotionControlServiceWithKey(cfg.APIKey)
	}
	return wellapi.NewKlingMotionControlService()
}

func NewEffectsService(cfg Config) EffectsService {
	if cfg.APIKey != "" {
		return wellapi.NewKlingEffectsServiceWithKey(cfg.APIKey)
	}
	return wellapi.NewKlingEffectsService()
}
