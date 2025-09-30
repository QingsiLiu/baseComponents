package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"net/url"
	"sort"
	"strings"
	"time"
)

// APISignatureTool 提供API签名验证功能
type APISignatureTool struct{}

// GenerateSalt 生成指定长度的随机盐值
func (APISignatureTool) GenerateSalt(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// MakeSecretSalt 处理盐值
func (APISignatureTool) MakeSecretSalt(salt string) string {
	return "x" + salt + "y" + salt + "z"
}

// GenerateSignature 生成API签名
func (t APISignatureTool) GenerateSignature(params map[string]string, secretKey string, salt string) string {
	// 添加盐值到参数
	paramsWithSalt := make(map[string]string)
	for k, v := range params {
		paramsWithSalt[k] = v
	}
	paramsWithSalt["salt"] = t.MakeSecretSalt(salt)

	// 按键名排序参数
	var keys []string
	for k := range paramsWithSalt {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建参数字符串
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(paramsWithSalt[k]))
	}
	encodedParams := strings.Join(pairs, "&")

	// 生成签名
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(encodedParams))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// ValidateSignature 验证API签名
func (t APISignatureTool) ValidateSignature(params map[string]string, receivedSignature string, secretKey string, salt string) bool {
	if salt == "" {
		return false // 没有盐值，验证失败
	}

	// 移除signature参数以避免重复添加
	paramsWithoutSignature := make(map[string]string)
	for k, v := range params {
		if k != "signature" {
			paramsWithoutSignature[k] = v
		}
	}

	// 生成签名
	expectedSignature := t.GenerateSignature(paramsWithoutSignature, secretKey, salt)

	// 比较生成的签名和接收到的签名
	return hmac.Equal([]byte(expectedSignature), []byte(receivedSignature))
}

// NewAPISignatureTool 创建API签名工具实例
func NewAPISignatureTool() *APISignatureTool {
	return &APISignatureTool{}
}
