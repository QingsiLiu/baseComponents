package utils

// Platform 定义支持的平台类型
type Platform string

// 平台类型常量
const (
	PlatformIOS Platform = "ios"
	PlatformWEB Platform = "web"
)

// IsValidPlatform 检查平台是否有效
func IsValidPlatform(platform string) bool {
	switch Platform(platform) {
	case PlatformIOS, PlatformWEB:
		return true
	default:
		return false
	}
}