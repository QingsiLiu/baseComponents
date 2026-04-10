package kling

import (
	"github.com/QingsiLiu/baseComponents/internal/serviceregistry"
	"github.com/QingsiLiu/baseComponents/internal/taskstatus"
)

const (
	TaskStatusPending   int32 = taskstatus.Pending
	TaskStatusRunning   int32 = taskstatus.Running
	TaskStatusCompleted int32 = taskstatus.Completed
	TaskStatusCanceled  int32 = taskstatus.Canceled
	TaskStatusFailed    int32 = taskstatus.Failed
)

// ServiceInfo 服务信息结构体。
type ServiceInfo struct {
	Source      string // 完整的服务源标识符
	ServiceType string // 混淆后的服务类型标识符
}

var serviceConfigs = []ServiceInfo{
	{Source: "owner", ServiceType: "0"},
	{Source: "wellapi_kling_motion_control", ServiceType: "wkmc"},
	{Source: "wellapi_kling_effects", ServiceType: "wkef"},
}

const (
	SourceOwner                     = "owner"
	SourceWellAPIKlingMotionControl = "wellapi_kling_motion_control"
	SourceWellAPIKlingEffects       = "wellapi_kling_effects"
)

const (
	ServiceTypeOwner                     = "0"
	ServiceTypeWellAPIKlingMotionControl = "wkmc"
	ServiceTypeWellAPIKlingEffects       = "wkef"
)

var (
	serviceTypeMap   map[string]string
	serviceSourceMap map[string]string
	AllServiceSource []string
	AllServiceType   []string
)

func init() {
	registry := serviceregistry.New(toRegistryConfigs(serviceConfigs))
	serviceTypeMap = registry.TypeBySource()
	serviceSourceMap = registry.SourceByType()
	AllServiceSource = registry.AllServiceSource()
	AllServiceType = registry.AllServiceType()
}

// GetServiceType 获取混淆后的任务类型标识符。
func GetServiceType(source string) string {
	if serviceType, ok := serviceTypeMap[source]; ok {
		return serviceType
	}
	return "unknown"
}

// GetServiceSource 根据服务类型获取服务源。
func GetServiceSource(serviceType string) string {
	if source, ok := serviceSourceMap[serviceType]; ok {
		return source
	}
	return "unknown"
}

// IsValidSource 检查服务源是否有效。
func IsValidSource(source string) bool {
	_, ok := serviceTypeMap[source]
	return ok
}

// IsValidServiceType 检查服务类型是否有效。
func IsValidServiceType(serviceType string) bool {
	_, ok := serviceSourceMap[serviceType]
	return ok
}

// GetAllServiceConfigs 获取所有服务配置（只读）。
func GetAllServiceConfigs() []ServiceInfo {
	configs := make([]ServiceInfo, len(serviceConfigs))
	copy(configs, serviceConfigs)
	return configs
}

func toRegistryConfigs(configs []ServiceInfo) []serviceregistry.Config {
	out := make([]serviceregistry.Config, 0, len(configs))
	for _, config := range configs {
		out = append(out, serviceregistry.Config{
			Source:      config.Source,
			ServiceType: config.ServiceType,
		})
	}
	return out
}
