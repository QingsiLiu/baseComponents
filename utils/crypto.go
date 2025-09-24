package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// MD5 计算MD5哈希值
func MD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// MD5String 计算字符串的MD5哈希值
func MD5String(s string) string {
	return MD5([]byte(s))
}

// SHA1 计算SHA1哈希值
func SHA1(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA1String 计算字符串的SHA1哈希值
func SHA1String(s string) string {
	return SHA1([]byte(s))
}

// SHA256 计算SHA256哈希值
func SHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SHA256String 计算字符串的SHA256哈希值
func SHA256String(s string) string {
	return SHA256([]byte(s))
}

// SHA512 计算SHA512哈希值
func SHA512(data []byte) string {
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:])
}

// SHA512String 计算字符串的SHA512哈希值
func SHA512String(s string) string {
	return SHA512([]byte(s))
}

// Base64Encode Base64编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64EncodeString 对字符串进行Base64编码
func Base64EncodeString(s string) string {
	return Base64Encode([]byte(s))
}

// Base64Decode Base64解码
func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// Base64DecodeString Base64解码为字符串
func Base64DecodeString(s string) (string, error) {
	data, err := Base64Decode(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Base64URLEncode URL安全的Base64编码
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode URL安全的Base64解码
func Base64URLDecode(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(s)
}

// GenerateRandomBytes 生成指定长度的随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateRandomString 生成指定长度的随机字符串（Base64编码）
func GenerateRandomString(length int) (string, error) {
	// 计算需要的字节数（Base64编码后长度会增加约1/3）
	byteLength := (length * 3) / 4
	if (length*3)%4 != 0 {
		byteLength++
	}

	bytes, err := GenerateRandomBytes(byteLength)
	if err != nil {
		return "", err
	}

	encoded := Base64Encode(bytes)
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	return encoded, nil
}

// AESEncrypt AES加密
func AESEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 生成随机IV
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// AESDecrypt AES解密
func AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// AESEncryptString AES加密字符串
func AESEncryptString(plaintext, key string) (string, error) {
	keyBytes := []byte(key)
	// 确保密钥长度为16, 24, 或 32 字节
	if len(keyBytes) < 16 {
		// 如果密钥太短，用SHA256哈希
		hash := sha256.Sum256(keyBytes)
		keyBytes = hash[:]
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	} else if len(keyBytes) > 24 {
		keyBytes = keyBytes[:24]
	} else if len(keyBytes) > 16 {
		keyBytes = keyBytes[:16]
	}

	encrypted, err := AESEncrypt([]byte(plaintext), keyBytes)
	if err != nil {
		return "", err
	}

	return Base64Encode(encrypted), nil
}

// AESDecryptString AES解密字符串
func AESDecryptString(ciphertext, key string) (string, error) {
	keyBytes := []byte(key)
	// 确保密钥长度为16, 24, 或 32 字节
	if len(keyBytes) < 16 {
		// 如果密钥太短，用SHA256哈希
		hash := sha256.Sum256(keyBytes)
		keyBytes = hash[:]
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	} else if len(keyBytes) > 24 {
		keyBytes = keyBytes[:24]
	} else if len(keyBytes) > 16 {
		keyBytes = keyBytes[:16]
	}

	encrypted, err := Base64Decode(ciphertext)
	if err != nil {
		return "", err
	}

	decrypted, err := AESDecrypt(encrypted, keyBytes)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
