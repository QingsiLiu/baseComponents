package serviceregistry

// Config 服务注册配置。
type Config struct {
	Source      string
	ServiceType string
}

// Registry 维护 source/type 双向映射。
type Registry struct {
	typeBySource map[string]string
	sourceByType map[string]string
	allSources   []string
	allTypes     []string
}

// New 基于配置创建注册表。
func New(configs []Config) *Registry {
	registry := &Registry{
		typeBySource: make(map[string]string, len(configs)),
		sourceByType: make(map[string]string, len(configs)),
		allSources:   make([]string, 0, len(configs)),
		allTypes:     make([]string, 0, len(configs)),
	}

	for _, config := range configs {
		registry.typeBySource[config.Source] = config.ServiceType
		registry.sourceByType[config.ServiceType] = config.Source
		registry.allSources = append(registry.allSources, config.Source)
		registry.allTypes = append(registry.allTypes, config.ServiceType)
	}

	return registry
}

// GetServiceType 获取混淆后的任务类型标识符。
func (r *Registry) GetServiceType(source string) string {
	if serviceType, ok := r.typeBySource[source]; ok {
		return serviceType
	}
	return "unknown"
}

// GetServiceSource 根据服务类型获取服务源。
func (r *Registry) GetServiceSource(serviceType string) string {
	if source, ok := r.sourceByType[serviceType]; ok {
		return source
	}
	return "unknown"
}

// IsValidSource 检查服务源是否有效。
func (r *Registry) IsValidSource(source string) bool {
	_, ok := r.typeBySource[source]
	return ok
}

// IsValidServiceType 检查服务类型是否有效。
func (r *Registry) IsValidServiceType(serviceType string) bool {
	_, ok := r.sourceByType[serviceType]
	return ok
}

// AllServiceSource 返回所有 source 的副本。
func (r *Registry) AllServiceSource() []string {
	out := make([]string, len(r.allSources))
	copy(out, r.allSources)
	return out
}

// AllServiceType 返回所有 service type 的副本。
func (r *Registry) AllServiceType() []string {
	out := make([]string, len(r.allTypes))
	copy(out, r.allTypes)
	return out
}

// TypeBySource 返回 source -> type 映射的副本。
func (r *Registry) TypeBySource() map[string]string {
	out := make(map[string]string, len(r.typeBySource))
	for source, serviceType := range r.typeBySource {
		out[source] = serviceType
	}
	return out
}

// SourceByType 返回 type -> source 映射的副本。
func (r *Registry) SourceByType() map[string]string {
	out := make(map[string]string, len(r.sourceByType))
	for serviceType, source := range r.sourceByType {
		out[serviceType] = source
	}
	return out
}
