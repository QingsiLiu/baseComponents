package aivideo

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

// ServiceInfo 服务信息结构体
type ServiceInfo struct {
	Source      string // 完整的服务源标识符
	ServiceType string // 混淆后的服务类型标识符
}

// 服务配置表 - 所有服务信息的单一数据源
var serviceConfigs = []ServiceInfo{
	{Source: "owner", ServiceType: "0"},
	{Source: "replicate_pixverse_v5", ServiceType: "rpv5"},
	{Source: "kie_kling_2_6_image_to_video", ServiceType: "kk26iv"},
	{Source: "kie_kling_2_6_text_to_video", ServiceType: "kk26tv"},
	{Source: "kie_kling_3_0_video", ServiceType: "kk3v"},
	{Source: "kie_seedance_1_5_pro", ServiceType: "ks15p"},
	{Source: "kie_seedance_2", ServiceType: "ks2"},
	{Source: "kie_seedance_2_fast", ServiceType: "ks2f"},
}

// 常量定义 - 从配置表自动生成
const (
	SourceOwner                  = "owner"
	SourceReplicatePixverse      = "replicate_pixverse_v5"
	SourceKieKling26ImageToVideo = "kie_kling_2_6_image_to_video"
	SourceKieKling26TextToVideo  = "kie_kling_2_6_text_to_video"
	SourceKieKling30Video        = "kie_kling_3_0_video"
	SourceKieSeedance15Pro       = "kie_seedance_1_5_pro"
	SourceKieSeedance2           = "kie_seedance_2"
	SourceKieSeedance2Fast       = "kie_seedance_2_fast"
)

// 服务类型混淆映射常量
const (
	ServiceTypeOwner                  = "0"
	ServiceTypeReplicatePixverse      = "rpv5"
	ServiceTypeKieKling26ImageToVideo = "kk26iv"
	ServiceTypeKieKling26TextToVideo  = "kk26tv"
	ServiceTypeKieKling30Video        = "kk3v"
	ServiceTypeKieSeedance15Pro       = "ks15p"
	ServiceTypeKieSeedance2           = "ks2"
	ServiceTypeKieSeedance2Fast       = "ks2f"
)

// 初始化时自动生成的映射表和切片
var (
	serviceTypeMap   map[string]string
	serviceSourceMap map[string]string
	AllServiceSource []string
	AllServiceType   []string
)

// init 初始化函数，自动从配置表生成所有映射关系
func init() {
	registry := serviceregistry.New(toRegistryConfigs(serviceConfigs))
	serviceTypeMap = registry.TypeBySource()
	serviceSourceMap = registry.SourceByType()
	AllServiceSource = registry.AllServiceSource()
	AllServiceType = registry.AllServiceType()
}

// GetServiceType 获取混淆后的任务类型标识符
func GetServiceType(source string) string {
	if serviceType, ok := serviceTypeMap[source]; ok {
		return serviceType
	}
	return "unknown"
}

// GetServiceSource 根据服务类型获取服务源
func GetServiceSource(serviceType string) string {
	if source, ok := serviceSourceMap[serviceType]; ok {
		return source
	}
	return "unknown"
}

// IsValidSource 检查服务源是否有效
func IsValidSource(source string) bool {
	_, ok := serviceTypeMap[source]
	return ok
}

// IsValidServiceType 检查服务类型是否有效
func IsValidServiceType(serviceType string) bool {
	_, ok := serviceSourceMap[serviceType]
	return ok
}

// GetAllServiceConfigs 获取所有服务配置（只读）
func GetAllServiceConfigs() []ServiceInfo {
	// 返回副本以防止外部修改
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
