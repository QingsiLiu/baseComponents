package storage

import (
	"io"
	"strings"
	"time"
)

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string    `json:"key"`          // 对象键名
	Size         int64     `json:"size"`         // 文件大小（字节）
	LastModified time.Time `json:"lastModified"` // 最后修改时间
	ETag         string    `json:"etag"`         // ETag值
	ContentType  string    `json:"contentType"`  // 内容类型
	IsDir        bool      `json:"isDir"`        // 是否为目录
}

// ListObjectsInput 列举对象输入参数
type ListObjectsInput struct {
	Bucket            string `json:"bucket"`            // 存储桶名称
	Prefix            string `json:"prefix"`            // 前缀过滤
	Delimiter         string `json:"delimiter"`         // 分隔符（用于目录结构）
	MaxKeys           int32  `json:"maxKeys"`           // 最大返回数量
	StartAfter        string `json:"startAfter"`        // 起始位置
	ContinuationToken string `json:"continuationToken"` // 分页令牌
}

// ListObjectsOutput 列举对象输出结果
type ListObjectsOutput struct {
	Objects               []ObjectInfo `json:"objects"`               // 对象列表
	CommonPrefixes        []string     `json:"commonPrefixes"`        // 公共前缀（目录）
	IsTruncated           bool         `json:"isTruncated"`           // 是否被截断
	NextContinuationToken string       `json:"nextContinuationToken"` // 下一页令牌
	KeyCount              int32        `json:"keyCount"`              // 返回的键数量
}

// CopyObjectInput 复制对象输入参数
type CopyObjectInput struct {
	SourceBucket      string            `json:"sourceBucket"`      // 源存储桶
	SourceKey         string            `json:"sourceKey"`         // 源对象键
	DestinationBucket string            `json:"destinationBucket"` // 目标存储桶
	DestinationKey    string            `json:"destinationKey"`    // 目标对象键
	Metadata          map[string]string `json:"metadata"`          // 元数据
	ContentType       string            `json:"contentType"`       // 内容类型
}

// ObjectMetadata 对象元数据
type ObjectMetadata struct {
	ContentType          string            `json:"contentType"`          // 内容类型
	ContentLength        int64             `json:"contentLength"`        // 内容长度
	LastModified         time.Time         `json:"lastModified"`         // 最后修改时间
	ETag                 string            `json:"etag"`                 // ETag值
	Metadata             map[string]string `json:"metadata"`             // 用户自定义元数据
	StorageClass         string            `json:"storageClass"`         // 存储类别
	ServerSideEncryption string            `json:"serverSideEncryption"` // 服务端加密
}

// StorageService 对象存储服务接口
type StorageService interface {
	// ===== 基础文件操作 =====

	// UploadObject 上传文件到存储
	UploadObject(bucketName, fileKey string, data []byte) error

	// UploadObjectStream 流式上传文件到存储
	UploadObjectStream(bucketName, fileKey string, file io.Reader) error

	// GetObject 获取文件
	GetObject(bucketName, fileKey string) ([]byte, error)

	// HeadObject 检查对象是否存在
	HeadObject(bucketName, fileKey string) bool

	// DeleteObject 删除单个对象
	DeleteObject(bucketName, fileKey string) error

	// DeleteObjects 批量删除对象
	DeleteObjects(bucketName string, fileKeys []string) ([]string, error)

	// ===== 文件管理操作 =====

	// ListObjects 列举对象
	ListObjects(input *ListObjectsInput) (*ListObjectsOutput, error)

	// CopyObject 复制对象
	CopyObject(input *CopyObjectInput) error

	// MoveObject 移动对象（复制后删除源对象）
	MoveObject(sourceBucket, sourceKey, destBucket, destKey string) error

	// GetObjectMetadata 获取对象元数据
	GetObjectMetadata(bucketName, fileKey string) (*ObjectMetadata, error)

	// ===== 目录操作 =====

	// CreateFolder 创建文件夹（通过创建以/结尾的空对象）
	CreateFolder(bucketName, folderPath string) error

	// DeleteFolder 删除文件夹及其所有内容
	DeleteFolder(bucketName, folderPath string) error

	// ListFolders 列举文件夹
	ListFolders(bucketName, prefix string) ([]string, error)

	// ===== 预签名URL操作 =====

	// PreSignPutObject 生成预签名上传URL
	PreSignPutObject(bucketName, fileKey string) (string, error)

	// BatchPreSignPutObject 批量生成预签名上传URL
	BatchPreSignPutObject(bucketName string, fileKeys []string, isWholeKey bool) map[string]string

	// PreSignGetObject 生成预签名获取URL
	PreSignGetObject(bucketName, fileKey string) (string, error)

	// PreSignDeleteObject 生成预签名删除URL
	PreSignDeleteObject(bucketName, fileKey string) (string, error)

	// ===== 高级功能 =====

	// SetObjectACL 设置对象访问控制列表
	SetObjectACL(bucketName, fileKey, acl string) error

	// GetObjectACL 获取对象访问控制列表
	GetObjectACL(bucketName, fileKey string) (string, error)

	// SetObjectMetadata 设置对象元数据
	SetObjectMetadata(bucketName, fileKey string, metadata map[string]string) error

	// GenerateDownloadURL 生成直接下载链接（公共读取）
	GenerateDownloadURL(bucketName, fileKey string) string
}

func GetContentType(fileName string) string {
	contentType := "image/jpeg" // 默认为jpeg
	if strings.HasSuffix(strings.ToLower(fileName), ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".gif") {
		contentType = "image/gif"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".webp") {
		contentType = "image/webp"
	} else if strings.HasSuffix(strings.ToLower(fileName), ".svg") {
		contentType = "image/svg+xml"
	}
	return contentType
}
