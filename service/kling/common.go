package kling

import aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"

const (
	TaskStatusPending   int32 = aivideokling.TaskStatusPending
	TaskStatusRunning   int32 = aivideokling.TaskStatusRunning
	TaskStatusCompleted int32 = aivideokling.TaskStatusCompleted
	TaskStatusCanceled  int32 = aivideokling.TaskStatusCanceled
	TaskStatusFailed    int32 = aivideokling.TaskStatusFailed
)

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.ServiceInfo.
type ServiceInfo = aivideokling.ServiceInfo

const (
	SourceOwner                     = aivideokling.SourceOwner
	SourceWellAPIKlingMotionControl = aivideokling.SourceWellAPIKlingMotionControl
	SourceWellAPIKlingEffects       = aivideokling.SourceWellAPIKlingEffects
)

const (
	ServiceTypeOwner                     = aivideokling.ServiceTypeOwner
	ServiceTypeWellAPIKlingMotionControl = aivideokling.ServiceTypeWellAPIKlingMotionControl
	ServiceTypeWellAPIKlingEffects       = aivideokling.ServiceTypeWellAPIKlingEffects
)

var (
	AllServiceSource = append([]string(nil), aivideokling.AllServiceSource...)
	AllServiceType   = append([]string(nil), aivideokling.AllServiceType...)
)

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.GetServiceType.
func GetServiceType(source string) string {
	return aivideokling.GetServiceType(source)
}

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.GetServiceSource.
func GetServiceSource(serviceType string) string {
	return aivideokling.GetServiceSource(serviceType)
}

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.IsValidSource.
func IsValidSource(source string) bool {
	return aivideokling.IsValidSource(source)
}

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.IsValidServiceType.
func IsValidServiceType(serviceType string) bool {
	return aivideokling.IsValidServiceType(serviceType)
}

// Deprecated: use github.com/QingsiLiu/baseComponents/service/aivideo/kling.GetAllServiceConfigs.
func GetAllServiceConfigs() []ServiceInfo {
	return aivideokling.GetAllServiceConfigs()
}
