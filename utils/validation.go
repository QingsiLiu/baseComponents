package utils

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// IsEmail 验证邮箱格式
func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsURL 验证URL格式
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsIP 验证IP地址格式
func IsIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsIPv4 验证IPv4地址格式
func IsIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// IsIPv6 验证IPv6地址格式
func IsIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}

// IsPhone 验证手机号格式（中国大陆）
func IsPhone(phone string) bool {
	// 中国大陆手机号正则表达式
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// IsIDCard 验证身份证号格式（中国大陆）
func IsIDCard(idCard string) bool {
	// 18位身份证号正则表达式
	pattern := `^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`
	matched, _ := regexp.MatchString(pattern, idCard)
	if !matched {
		return false
	}

	// 验证校验码
	return validateIDCardChecksum(idCard)
}

// validateIDCardChecksum 验证身份证校验码
func validateIDCardChecksum(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}

	// 权重因子
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	// 校验码对应表
	checkCodes := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}

	sum := 0
	for i := 0; i < 17; i++ {
		digit, err := strconv.Atoi(string(idCard[i]))
		if err != nil {
			return false
		}
		sum += digit * weights[i]
	}

	checkIndex := sum % 11
	expectedCheck := checkCodes[checkIndex]
	actualCheck := strings.ToUpper(string(idCard[17]))

	return expectedCheck == actualCheck
}

// IsNumeric 验证是否为数字
func IsNumeric(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

// IsInteger 验证是否为整数
func IsInteger(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

// IsAlpha 验证是否只包含字母
func IsAlpha(str string) bool {
	for _, r := range str {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return len(str) > 0
}

// IsAlphaNumeric 验证是否只包含字母和数字
func IsAlphaNumeric(str string) bool {
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return len(str) > 0
}

// IsLength 验证字符串长度是否在指定范围内
func IsLength(str string, min, max int) bool {
	length := len([]rune(str))
	return length >= min && length <= max
}

// IsMinLength 验证字符串最小长度
func IsMinLength(str string, min int) bool {
	return len([]rune(str)) >= min
}

// IsMaxLength 验证字符串最大长度
func IsMaxLength(str string, max int) bool {
	return len([]rune(str)) <= max
}

// IsPassword 验证密码强度
// 至少8位，包含大小写字母、数字和特殊字符
func IsPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsWeakPassword 验证弱密码（只需要字母和数字）
func IsWeakPassword(password string) bool {
	if len(password) < 6 {
		return false
	}

	hasLetter := false
	hasDigit := false

	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}

// IsCreditCard 验证信用卡号格式（Luhn算法）
func IsCreditCard(cardNumber string) bool {
	// 移除空格和连字符
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// 检查是否只包含数字
	if !IsNumeric(cardNumber) {
		return false
	}

	// 检查长度（通常13-19位）
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}

	// Luhn算法验证
	return luhnCheck(cardNumber)
}

// luhnCheck Luhn算法验证
func luhnCheck(cardNumber string) bool {
	sum := 0
	alternate := false

	// 从右到左遍历
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(cardNumber[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// IsJSON 验证是否为有效的JSON格式
func IsJSON(str string) bool {
	str = strings.TrimSpace(str)
	return (strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) ||
		(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]"))
}

// IsHexColor 验证是否为有效的十六进制颜色值
func IsHexColor(color string) bool {
	pattern := `^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`
	matched, _ := regexp.MatchString(pattern, color)
	return matched
}

// IsMAC 验证MAC地址格式
func IsMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil
}
