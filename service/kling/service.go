package kling

import aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.MotionControlService.
type MotionControlService = aivideokling.MotionControlService

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.EffectsService.
type EffectsService = aivideokling.EffectsService

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.KlingMotionControlTaskRunReq.
type KlingMotionControlTaskRunReq = aivideokling.KlingMotionControlTaskRunReq

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.KlingEffectsTaskRunReq.
type KlingEffectsTaskRunReq = aivideokling.KlingEffectsTaskRunReq

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.KlingEffectsInput.
type KlingEffectsInput = aivideokling.KlingEffectsInput

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.TaskInfo.
type TaskInfo = aivideokling.TaskInfo

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.VideoResult.
type VideoResult = aivideokling.VideoResult

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.ImageResult.
type ImageResult = aivideokling.ImageResult

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.ConvertTaskStatus.
func ConvertTaskStatus(status string) int32 {
	return aivideokling.ConvertTaskStatus(status)
}
