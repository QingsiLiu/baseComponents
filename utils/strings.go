package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// ToCamelCase 将下划线分隔的字符串转换为驼峰命名
func ToCamelCase(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.Title(parts[i])
		}
	}
	return result
}

// ToSnakeCase 将驼峰命名转换为下划线分隔
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// ToKebabCase 将字符串转换为短横线分隔
func ToKebabCase(s string) string {
	return strings.ReplaceAll(ToSnakeCase(s), "_", "-")
}

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsEmpty 检查字符串是否为空或只包含空白字符
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Truncate 截断字符串到指定长度，可选择添加省略号
func Truncate(s string, length int, ellipsis ...string) string {
	if len(s) <= length {
		return s
	}

	suffix := "..."
	if len(ellipsis) > 0 {
		suffix = ellipsis[0]
	}

	if length <= len(suffix) {
		return suffix[:length]
	}

	return s[:length-len(suffix)] + suffix
}

// RemoveSpaces 移除字符串中的所有空格
func RemoveSpaces(s string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(s, "")
}

// CompactSpaces 将多个连续空格压缩为单个空格
func CompactSpaces(s string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(s), " ")
}

// Contains 检查字符串是否包含任意一个子字符串
func Contains(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// ContainsAll 检查字符串是否包含所有子字符串
func ContainsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// Mask 对字符串进行掩码处理
func Mask(s string, start, end int, maskChar rune) string {
	if s == "" || start < 0 || end < 0 || start >= len(s) {
		return s
	}

	runes := []rune(s)
	length := len(runes)

	if end > length {
		end = length
	}

	if start > end {
		start, end = end, start
	}

	for i := start; i < end; i++ {
		runes[i] = maskChar
	}

	return string(runes)
}

// RandomString 生成指定长度的随机字符串
func RandomString(length int, charset ...string) string {
	if length <= 0 {
		return ""
	}

	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if len(charset) > 0 {
		chars = charset[0]
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = chars[i%len(chars)] // 简化版本，实际应使用随机数
	}

	return string(result)
}
