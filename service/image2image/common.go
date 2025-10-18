package image2image

const (
	TaskStatusPending = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusCanceled
	TaskStatusFailed
)

// ServiceInfo 服务信息结构体
type ServiceInfo struct {
	Source      string // 完整的服务源标识符
	ServiceType string // 混淆后的服务类型标识符
}

// 服务配置表 - 所有服务信息的单一数据源
var serviceConfigs = []ServiceInfo{
	{Source: "owner", ServiceType: "0"},
	{Source: "replicate_nano_banana", ServiceType: "rnb"},
	{Source: "replicate_controlnet", ServiceType: "rcn"},
	{Source: "modelslab_interior", ServiceType: "mli"},
	{Source: "modelslab_exterior", ServiceType: "mle"},
	{Source: "kie_nano_banana", ServiceType: "knb"},
}

// 常量定义 - 从配置表自动生成
const (
	SourceOwner               = "owner"
	SourceReplicateNanoBanana = "replicate_nano_banana"
	SourceReplicateControlNet = "replicate_controlnet"
	SourceModelsLabInterior   = "modelslab_interior"
	SourceModelsLabExterior   = "modelslab_exterior"
	SourceKieNanoBanana       = "kie_nano_banana"
)

const (
	ServiceTypeOwner               = "0"
	ServiceTypeModelsLabInterior   = "mli"
	ServiceTypeModelsLabExterior   = "mle"
	ServiceTypeReplicateNanoBanana = "rnb"
	ServiceTypeReplicateControlNet = "rcn"
	ServiceTypeKieNanoBanana       = "knb"
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
	serviceTypeMap = make(map[string]string, len(serviceConfigs))
	serviceSourceMap = make(map[string]string, len(serviceConfigs))
	AllServiceSource = make([]string, 0, len(serviceConfigs))
	AllServiceType = make([]string, 0, len(serviceConfigs))

	for _, config := range serviceConfigs {
		serviceTypeMap[config.Source] = config.ServiceType
		serviceSourceMap[config.ServiceType] = config.Source
		AllServiceSource = append(AllServiceSource, config.Source)
		AllServiceType = append(AllServiceType, config.ServiceType)
	}
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
