package text2image

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
	{Source: "replicate_flux-schnell", ServiceType: "rfc"},
	{Source: "replicate_flux1dev", ServiceType: "rf1d"},
	{Source: "replicate_qwen_image", ServiceType: "rqi"},
	{Source: "modelslab_flux", ServiceType: "mlf"},
	{Source: "replicate_prunaai_qwen_image_fast", ServiceType: "rpqf"},
	{Source: "kie_qwen_image_text2image", ServiceType: "kqi"},
	{Source: "kie_ideogram_v3_text2image", ServiceType: "kiv3"},
	{Source: "kie_gpt_image_2_text2image", ServiceType: "kgi2"},
}

// 常量定义 - 从配置表自动生成
const (
	SourceOwner                         = "owner"
	SourceReplicateFluxSchnell          = "replicate_flux-schnell"
	SourceReplicateFlux1Dev             = "replicate_flux1dev"
	SourceReplicateQwenImage            = "replicate_qwen_image"
	SourceModelslabFlux                 = "modelslab_flux"
	SourceReplicatePrunaAIQwenImageFast = "replicate_prunaai_qwen_image_fast"
	SourceKieQwenImageText2Image        = "kie_qwen_image_text2image"
	SourceKieIdeogramV3Text2Image       = "kie_ideogram_v3_text2image"
	SourceKieGPTImage2Text2Image        = "kie_gpt_image_2_text2image"
)

// 服务类型混淆映射常量
const (
	ServiceTypeOwner                         = "0"
	ServiceTypeReplicateFluxSchnell          = "rfc"
	ServiceTypeReplicateFlux1Dev             = "rf1d"
	ServiceTypeReplicateQwenImage            = "rqi"
	ServiceTypeModelslabFlux                 = "mlf"
	ServiceTypeReplicatePrunaAIQwenImageFast = "rpqf"
	ServiceTypeKieQwenImageText2Image        = "kqi"
	ServiceTypeKieIdeogramV3Text2Image       = "kiv3"
	ServiceTypeKieGPTImage2Text2Image        = "kgi2"
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
