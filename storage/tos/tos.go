package tos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/QingsiLiu/baseComponents/storage"
	v2tos "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/codes"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

const defaultPreSignTTL = 15 * time.Minute

// Config 初始化TOS服务所需的配置
type Config struct {
	Endpoint       string
	Region         string
	AccessKey      string
	SecretKey      string
	SecurityToken  string
	PreSignExpires time.Duration
}

// TOSService 火山引擎对象存储服务
type TOSService struct {
	client       *v2tos.ClientV2
	ctx          context.Context
	preSignTTL   time.Duration
	defaultScope Config
}

// NewTOSService 创建一个新的TOS服务实例
func NewTOSService(cfg Config) (storage.StorageService, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("tos: endpoint is required")
	}

	options := make([]v2tos.ClientOption, 0, 3)
	if cfg.Region != "" {
		options = append(options, v2tos.WithRegion(cfg.Region))
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		cred := v2tos.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey)
		if cfg.SecurityToken != "" {
			cred.WithSecurityToken(cfg.SecurityToken)
		}
		options = append(options, v2tos.WithCredentials(cred))
	}

	client, err := v2tos.NewClientV2(cfg.Endpoint, options...)
	if err != nil {
		return nil, err
	}

	presignTTL := cfg.PreSignExpires
	if presignTTL <= 0 {
		presignTTL = defaultPreSignTTL
	}

	return &TOSService{
		client:       client,
		ctx:          context.Background(),
		preSignTTL:   presignTTL,
		defaultScope: cfg,
	}, nil
}

// ===== 基础文件操作 =====

// UploadObject 上传文件到TOS
func (t *TOSService) UploadObject(bucketName, fileKey string, data []byte) error {
	input := &v2tos.PutObjectV2Input{
		PutObjectBasicInput: v2tos.PutObjectBasicInput{
			Bucket:        bucketName,
			Key:           fileKey,
			ContentType:   storage.GetContentType(fileKey),
			ContentLength: int64(len(data)),
		},
		Content: bytes.NewReader(data),
	}

	_, err := t.client.PutObjectV2(t.ctx, input)
	return err
}

// UploadObjectStream 流式上传文件
func (t *TOSService) UploadObjectStream(bucketName, fileKey string, file io.Reader) error {
	input := &v2tos.PutObjectV2Input{
		PutObjectBasicInput: v2tos.PutObjectBasicInput{
			Bucket:      bucketName,
			Key:         fileKey,
			ContentType: storage.GetContentType(fileKey),
		},
		Content: file,
	}

	_, err := t.client.PutObjectV2(t.ctx, input)
	return err
}

// GetObject 获取文件内容
func (t *TOSService) GetObject(bucketName, fileKey string) ([]byte, error) {
	output, err := t.client.GetObjectV2(t.ctx, &v2tos.GetObjectV2Input{
		Bucket: bucketName,
		Key:    fileKey,
	})
	if err != nil {
		return nil, err
	}
	defer output.Content.Close()

	return io.ReadAll(output.Content)
}

// HeadObject 检查对象是否存在
func (t *TOSService) HeadObject(bucketName, fileKey string) bool {
	_, err := t.client.HeadObjectV2(t.ctx, &v2tos.HeadObjectV2Input{
		Bucket: bucketName,
		Key:    fileKey,
	})
	if err != nil {
		if v2tos.Code(err) == codes.NoSuchKey {
			return false
		}
		return false
	}
	return true
}

// DeleteObject 删除单个对象
func (t *TOSService) DeleteObject(bucketName, fileKey string) error {
	_, err := t.client.DeleteObjectV2(t.ctx, &v2tos.DeleteObjectV2Input{
		Bucket: bucketName,
		Key:    fileKey,
	})
	return err
}

// DeleteObjects 批量删除对象，返回删除失败的对象key
func (t *TOSService) DeleteObjects(bucketName string, fileKeys []string) ([]string, error) {
	if len(fileKeys) == 0 {
		return []string{}, nil
	}

	objects := make([]v2tos.ObjectTobeDeleted, 0, len(fileKeys))
	for _, key := range fileKeys {
		objects = append(objects, v2tos.ObjectTobeDeleted{Key: key})
	}

	output, err := t.client.DeleteMultiObjects(t.ctx, &v2tos.DeleteMultiObjectsInput{
		Bucket:  bucketName,
		Objects: objects,
	})
	if err != nil {
		return nil, err
	}

	var failed []string
	for _, delErr := range output.Error {
		failed = append(failed, delErr.Key)
	}
	return failed, nil
}

// ===== 文件管理操作 =====

// ListObjects 列举对象
func (t *TOSService) ListObjects(input *storage.ListObjectsInput) (*storage.ListObjectsOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("tos: list objects input is nil")
	}

	listInput := &v2tos.ListObjectsType2Input{
		Bucket:            input.Bucket,
		Prefix:            input.Prefix,
		Delimiter:         input.Delimiter,
		StartAfter:        input.StartAfter,
		ContinuationToken: input.ContinuationToken,
	}
	if input.MaxKeys > 0 {
		listInput.MaxKeys = int(input.MaxKeys)
	}

	resp, err := t.client.ListObjectsType2(t.ctx, listInput)
	if err != nil {
		return nil, err
	}

	output := &storage.ListObjectsOutput{
		IsTruncated:           resp.IsTruncated,
		KeyCount:              int32(resp.KeyCount),
		NextContinuationToken: resp.NextContinuationToken,
	}

	for _, obj := range resp.Contents {
		output.Objects = append(output.Objects, storage.ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ETag:         obj.ETag,
			IsDir:        strings.HasSuffix(obj.Key, "/"),
		})
	}

	for _, prefix := range resp.CommonPrefixes {
		output.CommonPrefixes = append(output.CommonPrefixes, prefix.Prefix)
	}

	return output, nil
}

// CopyObject 复制对象
func (t *TOSService) CopyObject(input *storage.CopyObjectInput) error {
	if input == nil {
		return fmt.Errorf("tos: copy object input is nil")
	}

	copyInput := &v2tos.CopyObjectInput{
		Bucket:            input.DestinationBucket,
		Key:               input.DestinationKey,
		SrcBucket:         input.SourceBucket,
		SrcKey:            input.SourceKey,
		ContentType:       input.ContentType,
		MetadataDirective: enum.MetadataDirectiveCopy,
	}
	if len(input.Metadata) > 0 {
		copyInput.MetadataDirective = enum.MetadataDirectiveReplace
		copyInput.Meta = input.Metadata
	}

	_, err := t.client.CopyObject(t.ctx, copyInput)
	return err
}

// MoveObject 移动对象（先复制后删除）
func (t *TOSService) MoveObject(sourceBucket, sourceKey, destBucket, destKey string) error {
	copyInput := &storage.CopyObjectInput{
		SourceBucket:      sourceBucket,
		SourceKey:         sourceKey,
		DestinationBucket: destBucket,
		DestinationKey:    destKey,
	}
	if err := t.CopyObject(copyInput); err != nil {
		return err
	}
	return t.DeleteObject(sourceBucket, sourceKey)
}

// GetObjectMetadata 获取对象元数据
func (t *TOSService) GetObjectMetadata(bucketName, fileKey string) (*storage.ObjectMetadata, error) {
	resp, err := t.client.HeadObjectV2(t.ctx, &v2tos.HeadObjectV2Input{
		Bucket: bucketName,
		Key:    fileKey,
	})
	if err != nil {
		return nil, err
	}

	meta := &storage.ObjectMetadata{
		ContentType:          resp.ContentType,
		ContentLength:        resp.ContentLength,
		LastModified:         resp.LastModified,
		ETag:                 resp.ETag,
		Metadata:             metadataToMap(resp.Meta),
		StorageClass:         string(resp.StorageClass),
		ServerSideEncryption: resp.ServerSideEncryption,
	}
	return meta, nil
}

// ===== 目录操作 =====

// CreateFolder 创建空目录（通过创建空对象实现）
func (t *TOSService) CreateFolder(bucketName, folderPath string) error {
	key := ensureTrailingSlash(folderPath)
	return t.UploadObject(bucketName, key, []byte{})
}

// DeleteFolder 删除目录及其所有内容
func (t *TOSService) DeleteFolder(bucketName, folderPath string) error {
	prefix := ensureTrailingSlash(folderPath)
	token := ""

	for {
		resp, err := t.client.ListObjectsType2(t.ctx, &v2tos.ListObjectsType2Input{
			Bucket:            bucketName,
			Prefix:            prefix,
			ContinuationToken: token,
			MaxKeys:           1000,
		})
		if err != nil {
			return err
		}

		var keys []string
		for _, obj := range resp.Contents {
			keys = append(keys, obj.Key)
		}

		if len(keys) > 0 {
			if _, err := t.DeleteObjects(bucketName, keys); err != nil {
				return err
			}
		}

		if !resp.IsTruncated {
			break
		}
		token = resp.NextContinuationToken
	}

	return nil
}

// ListFolders 列举指定前缀下的“目录”
func (t *TOSService) ListFolders(bucketName, prefix string) ([]string, error) {
	resp, err := t.client.ListObjectsType2(t.ctx, &v2tos.ListObjectsType2Input{
		Bucket:    bucketName,
		Prefix:    prefix,
		Delimiter: "/",
		MaxKeys:   1000,
	})
	if err != nil {
		return nil, err
	}

	folders := make([]string, 0, len(resp.CommonPrefixes))
	for _, common := range resp.CommonPrefixes {
		folders = append(folders, common.Prefix)
	}
	return folders, nil
}

// ===== 预签名URL操作 =====

// PreSignPutObject 生成预签名上传链接
func (t *TOSService) PreSignPutObject(bucketName, fileKey string) (string, error) {
	resp, err := t.client.PreSignedURL(&v2tos.PreSignedURLInput{
		HTTPMethod: enum.HttpMethodPut,
		Bucket:     bucketName,
		Key:        fileKey,
		Expires:    int64(t.preSignTTL.Seconds()),
	})
	if err != nil {
		return "", err
	}
	return resp.SignedUrl, nil
}

// BatchPreSignPutObject 批量生成预签名上传URL
func (t *TOSService) BatchPreSignPutObject(bucketName string, fileKeys []string, isWholeKey bool) map[string]string {
	result := make(map[string]string, len(fileKeys))
	for _, key := range fileKeys {
		actualKey := key
		if !isWholeKey {
			actualKey = key
		}
		url, err := t.PreSignPutObject(bucketName, actualKey)
		if err != nil {
			log.Printf("tos: failed to pre-sign put url for %s: %v", actualKey, err)
			result[key] = ""
			continue
		}
		result[key] = url
	}
	return result
}

// PreSignGetObject 生成预签名下载链接
func (t *TOSService) PreSignGetObject(bucketName, fileKey string) (string, error) {
	resp, err := t.client.PreSignedURL(&v2tos.PreSignedURLInput{
		HTTPMethod: enum.HttpMethodGet,
		Bucket:     bucketName,
		Key:        fileKey,
		Expires:    int64(t.preSignTTL.Seconds()),
	})
	if err != nil {
		return "", err
	}
	return resp.SignedUrl, nil
}

// PreSignDeleteObject 生成预签名删除链接
func (t *TOSService) PreSignDeleteObject(bucketName, fileKey string) (string, error) {
	resp, err := t.client.PreSignedURL(&v2tos.PreSignedURLInput{
		HTTPMethod: enum.HttpMethodDelete,
		Bucket:     bucketName,
		Key:        fileKey,
		Expires:    int64(t.preSignTTL.Seconds()),
	})
	if err != nil {
		return "", err
	}
	return resp.SignedUrl, nil
}

// ===== 高级功能 =====

// SetObjectACL 设置对象ACL
func (t *TOSService) SetObjectACL(bucketName, fileKey, acl string) error {
	if acl == "" {
		acl = string(enum.ACLPrivate)
	}
	_, err := t.client.PutObjectACL(t.ctx, &v2tos.PutObjectACLInput{
		Bucket: bucketName,
		Key:    fileKey,
		ACL:    enum.ACLType(acl),
	})
	return err
}

// GetObjectACL 获取对象ACL
func (t *TOSService) GetObjectACL(bucketName, fileKey string) (string, error) {
	output, err := t.client.GetObjectACL(t.ctx, &v2tos.GetObjectACLInput{
		Bucket: bucketName,
		Key:    fileKey,
	})
	if err != nil {
		return "", err
	}

	for _, grant := range output.Grants {
		if grant.GranteeV2.Canned == enum.CannedAllUsers {
			switch grant.Permission {
			case enum.PermissionRead:
				return string(enum.ACLPublicRead), nil
			case enum.PermissionWrite:
				return string(enum.ACLPublicReadWrite), nil
			}
		}
	}
	return string(enum.ACLPrivate), nil
}

// SetObjectMetadata 设置对象自定义元数据
func (t *TOSService) SetObjectMetadata(bucketName, fileKey string, metadata map[string]string) error {
	if metadata == nil {
		metadata = map[string]string{}
	}
	_, err := t.client.SetObjectMeta(t.ctx, &v2tos.SetObjectMetaInput{
		Bucket: bucketName,
		Key:    fileKey,
		Meta:   metadata,
	})
	return err
}

// GenerateDownloadURL 生成下载链接（默认走预签名）
func (t *TOSService) GenerateDownloadURL(bucketName, fileKey string) string {
	url, err := t.PreSignGetObject(bucketName, fileKey)
	if err != nil {
		return ""
	}
	return url
}

// ===== 辅助方法 =====

func metadataToMap(meta v2tos.Metadata) map[string]string {
	if meta == nil {
		return nil
	}
	result := make(map[string]string)
	meta.Range(func(key, value string) bool {
		result[key] = value
		return true
	})
	return result
}

func ensureTrailingSlash(path string) string {
	if path == "" {
		return ""
	}
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}
