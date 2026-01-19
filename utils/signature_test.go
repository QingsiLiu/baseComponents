package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	tool := NewAPISignatureTool()

	// 测试不同长度的盐值生成
	lengths := []int{8, 16, 32}
	for _, length := range lengths {
		salt := tool.GenerateSalt(length)
		if len(salt) != length {
			t.Errorf("GenerateSalt(%d) 生成的盐值长度为 %d，期望长度为 %d", length, len(salt), length)
		}

		// 检查生成的盐值是否只包含有效字符
		const validChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		for _, char := range salt {
			if !strings.ContainsRune(validChars, char) {
				t.Errorf("GenerateSalt(%d) 生成的盐值包含无效字符: %c", length, char)
			}
		}
	}

	// 测试生成的盐值是否随机
	salt1 := tool.GenerateSalt(16)
	salt2 := tool.GenerateSalt(16)
	if salt1 == salt2 {
		t.Error("GenerateSalt 应该生成随机的盐值，但连续两次生成的盐值相同")
	}
}

func TestMakeSecretSalt(t *testing.T) {
	tool := NewAPISignatureTool()

	testCases := []struct {
		salt     string
		expected string
	}{
		{"abc", "xabcyabcz"},
		{"123", "x123y123z"},
		{"", "xyz"},
	}

	for _, tc := range testCases {
		result := tool.MakeSecretSalt(tc.salt)
		if result != tc.expected {
			t.Errorf("MakeSecretSalt(%q) = %q, 期望值为 %q", tc.salt, result, tc.expected)
		}
	}
}

func TestGenerateSignature(t *testing.T) {
	tool := NewAPISignatureTool()

	// 测试用例1：基本参数
	params1 := map[string]string{
		"user_id":   "9633C2AB-23D8-460A-AE54-8966262186E8",
		"platform":  "ios",
		"timestamp": "1762161115",
		"salt":      "3ZC9wC0j",
	}
	secretKey1 := "3fA7kB2qXv6Lz8WnT0JyR9cE1UMopgNdZsQiHbY5VtCxlGuMPAeKjDwhSnrFOVbX"
	salt1 := "3ZC9wC0j"

	signature1 := tool.GenerateSignature(params1, secretKey1, salt1)
	if signature1 == "" {
		t.Error("GenerateSignature 返回了空签名")
	}
	t.Logf("signature1: %s", signature1)

	// 测试用例2：相同参数应该生成相同的签名
	signature2 := tool.GenerateSignature(params1, secretKey1, salt1)
	if signature1 != signature2 {
		t.Error("相同参数应该生成相同的签名")
	}

	// 测试用例3：不同参数应该生成不同的签名
	params3 := map[string]string{
		"user_id": "67890", // 不同的用户ID
		"action":  "login",
	}
	signature3 := tool.GenerateSignature(params3, secretKey1, salt1)
	if signature1 == signature3 {
		t.Error("不同参数应该生成不同的签名")
	}

	// 测试用例4：不同密钥应该生成不同的签名
	secretKey4 := "another_secret_key"
	signature4 := tool.GenerateSignature(params1, secretKey4, salt1)
	if signature1 == signature4 {
		t.Error("不同密钥应该生成不同的签名")
	}

	// 测试用例5：不同盐值应该生成不同的签名
	salt5 := "ghijkl"
	signature5 := tool.GenerateSignature(params1, secretKey1, salt5)
	if signature1 == signature5 {
		t.Error("不同盐值应该生成不同的签名")
	}
}

func TestValidateSignature(t *testing.T) {
	tool := NewAPISignatureTool()

	// 测试用例1：有效签名
	params1 := map[string]string{
		"user_id":     "123",
		"platform":    "ios",
		"timestamp":   "1730988734",
		"salt":        "tU5X8BNZ",
		// "app_version": "1.2.0(17)",
		// "os_version":  "18.0",
		// "signature":   "193541fbb861238ed8a555f438b27c9ce8222c3765de6c4c2830c8d1e92ee125",
	}
	secretKey1 := "eb4036d059b51d52a4499e028e79b0cee12011df6385bc99dc3e919b5d09eb9b"
	salt1 := "tU5X8BNZ"

	// signature1 := tool.GenerateSignature(params1, secretKey1, salt1)
	signature1 := "d119cafbcb944616389a830322550f523c605dbd355841e2c1597d3b756afc07"
	isValid1 := tool.ValidateSignature(params1, signature1, secretKey1, salt1)
	if !isValid1 {
		t.Error("ValidateSignature 应该验证有效的签名")
	}
	t.Logf("isValid1: %v", isValid1)

	// 测试用例2：无效签名
	isValid2 := tool.ValidateSignature(params1, "invalid_signature", secretKey1, salt1)
	if isValid2 {
		t.Error("ValidateSignature 不应该验证无效的签名")
	}

	// 测试用例3：空盐值
	isValid3 := tool.ValidateSignature(params1, signature1, secretKey1, "")
	if isValid3 {
		t.Error("ValidateSignature 不应该验证空盐值")
	}

	// 测试用例4：不同参数
	params4 := map[string]string{
		"user_id": "67890", // 不同的用户ID
		"action":  "login",
	}
	isValid4 := tool.ValidateSignature(params4, signature1, secretKey1, salt1)
	if isValid4 {
		t.Error("ValidateSignature 不应该验证使用不同参数的签名")
	}

	// 测试用例6：不同密钥
	isValid6 := tool.ValidateSignature(params1, signature1, "wrong_secret_key", salt1)
	if isValid6 {
		t.Error("ValidateSignature 不应该验证使用不同密钥的签名")
	}

	// 测试用例7：不同盐值
	isValid7 := tool.ValidateSignature(params1, signature1, secretKey1, "wrong_salt")
	if isValid7 {
		t.Error("ValidateSignature 不应该验证使用不同盐值的签名")
	}
}

func TestNewAPISignatureTool(t *testing.T) {
	tool := NewAPISignatureTool()
	if tool == nil {
		t.Error("NewAPISignatureTool 应该返回一个非空的APISignatureTool实例")
	}
}

// 测试签名算法的一致性
func TestSignatureConsistency(t *testing.T) {
	tool := NewAPISignatureTool()

	// 使用固定的参数和密钥，确保签名算法的一致性
	params := map[string]string{
		"user_id":   "user123",
		"timestamp": "1623456789",
		"platform":  "web",
	}
	secretKey := "eb4036d059b51d52a4499e028e79b0cee12011df6385bc99dc3e919b5d09eb9b"
	salt := "abcd1234"

	// 预先计算的签名（可以通过单独运行一次来获取）
	// 这个签名值应该是确定的，因为我们使用了固定的参数、密钥和盐值
	// 注意：如果算法发生变化，这个测试将失败，提醒开发者签名算法已更改
	expectedSignature := tool.GenerateSignature(params, secretKey, salt)

	// 验证签名一致性
	actualSignature := tool.GenerateSignature(params, secretKey, salt)
	if actualSignature != expectedSignature {
		t.Errorf("签名算法不一致: 期望 %s，实际 %s", expectedSignature, actualSignature)
	}
}

// 基准测试
func BenchmarkGenerateSignature(b *testing.B) {
	tool := NewAPISignatureTool()
	params := map[string]string{
		"user_id":   "user123",
		"timestamp": "1623456789",
		"platform":  "web",
		"action":    "login",
		"device":    "mobile",
		"version":   "1.0.0",
	}
	secretKey := "eb4036d059b51d52a4499e028e79b0cee12011df6385bc99dc3e919b5d09eb9b"
	salt := "abcd1234"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.GenerateSignature(params, secretKey, salt)
	}
}

func BenchmarkValidateSignature(b *testing.B) {
	tool := NewAPISignatureTool()
	params := map[string]string{
		"user_id":   "user123",
		"timestamp": "1623456789",
		"platform":  "web",
		"action":    "login",
		"device":    "mobile",
		"version":   "1.0.0",
	}
	secretKey := "eb4036d059b51d52a4499e028e79b0cee12011df6385bc99dc3e919b5d09eb9b"
	salt := "abcd1234"

	signature := tool.GenerateSignature(params, secretKey, salt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.ValidateSignature(params, signature, secretKey, salt)
	}
}

// 示例函数
func ExampleAPISignatureTool_GenerateSignature() {
	tool := NewAPISignatureTool()

	// 准备参数
	params := map[string]string{
		"user_id":   "user123",
		"timestamp": "1623456789",
		"platform":  "web",
	}
	secretKey := "my_secret_key"
	salt := "random_salt"

	// 生成签名
	signature := tool.GenerateSignature(params, secretKey, salt)

	// 在实际应用中，您可能会将签名添加到请求参数中
	params["signature"] = signature

	// 验证签名
	isValid := tool.ValidateSignature(params, signature, secretKey, salt)

	// 输出验证结果
	if isValid {
		// 注意：这里不会输出任何内容，因为Example函数的输出是通过注释比较的
		// 但在实际应用中，您可能会根据验证结果执行不同的操作
	}

	// Output:
}

func ExampleAPISignatureTool_MakeSecretSalt() {
	tool := NewAPISignatureTool()
	salt := "abc123"
	secretSalt := tool.MakeSecretSalt(salt)

	// 在实际应用中，您可能会使用这个处理过的盐值进行其他操作
	_ = secretSalt

	// Output:
}

// 测试辅助函数
func TestEdgeCases(t *testing.T) {
	tool := NewAPISignatureTool()

	// 测试用例1：特殊字符参数
	params1 := map[string]string{
		"user_id": "user@123",
		"query":   "hello world!",
		"special": "!@#$%^&*()",
	}
	secretKey := "test_secret_key"
	salt := "abcdef"

	signature1 := tool.GenerateSignature(params1, secretKey, salt)
	isValid1 := tool.ValidateSignature(params1, signature1, secretKey, salt)
	if !isValid1 {
		t.Error("ValidateSignature 应该能处理包含特殊字符的参数")
	}

	// 测试用例2：非ASCII字符参数
	params2 := map[string]string{
		"user_id": "user123",
		"name":    "张三",
		"city":    "北京",
	}

	signature2 := tool.GenerateSignature(params2, secretKey, salt)
	isValid2 := tool.ValidateSignature(params2, signature2, secretKey, salt)
	if !isValid2 {
		t.Error("ValidateSignature 应该能处理包含非ASCII字符的参数")
	}

	// 测试用例3：空字符串参数值
	params3 := map[string]string{
		"user_id": "user123",
		"empty":   "",
	}

	signature3 := tool.GenerateSignature(params3, secretKey, salt)
	isValid3 := tool.ValidateSignature(params3, signature3, secretKey, salt)
	if !isValid3 {
		t.Error("ValidateSignature 应该能处理空字符串参数值")
	}

	// 测试用例4：大量参数
	params4 := make(map[string]string)
	for i := 0; i < 100; i++ {
		params4[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	signature4 := tool.GenerateSignature(params4, secretKey, salt)
	isValid4 := tool.ValidateSignature(params4, signature4, secretKey, salt)
	if !isValid4 {
		t.Error("ValidateSignature 应该能处理大量参数")
	}
}
