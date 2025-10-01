package gcs

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/QingsiLiu/baseComponents/storage"
)

// GCSClient Google Cloud Storage 客户端
type GCSClient struct {
	client    *gcs.Client
	projectID string
	ctx       context.Context
}

// NewGCSClient 创建新的 GCS 客户端
func NewGCSClient(projectID string, credentialsFile string) (*GCSClient, error) {
	ctx := context.Background()
	
	var client *gcs.Client
	var err error
	
	if credentialsFile != "" {
		client, err = gcs.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	} else {
		// 使用默认凭据（环境变量或服务账号）
		client, err = gcs.NewClient(ctx)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %v", err)
	}

	return &GCSClient{
		client:    client,
		projectID: projectID,
		ctx:       ctx,
	}, nil
}

// Close 关闭客户端连接
func (g *GCSClient) Close() error {
	return g.client.Close()
}

// ===== 基础文件操作 =====

// UploadObject 上传文件到存储
func (g *GCSClient) UploadObject(bucketName, fileKey string, data []byte) error {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	writer := obj.NewWriter(g.ctx)
	defer writer.Close()
	
	// 设置内容类型
	writer.ContentType = storage.GetContentType(fileKey)
	
	_, err := writer.Write(data)
	return err
}

// UploadObjectStream 流式上传文件到存储
func (g *GCSClient) UploadObjectStream(bucketName, fileKey string, file io.Reader) error {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	writer := obj.NewWriter(g.ctx)
	defer writer.Close()
	
	// 设置内容类型
	writer.ContentType = storage.GetContentType(fileKey)
	
	_, err := io.Copy(writer, file)
	return err
}

// GetObject 获取文件
func (g *GCSClient) GetObject(bucketName, fileKey string) ([]byte, error) {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	reader, err := obj.NewReader(g.ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	return io.ReadAll(reader)
}

// HeadObject 检查对象是否存在
func (g *GCSClient) HeadObject(bucketName, fileKey string) bool {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	_, err := obj.Attrs(g.ctx)
	return err == nil
}

// DeleteObject 删除单个对象
func (g *GCSClient) DeleteObject(bucketName, fileKey string) error {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	return obj.Delete(g.ctx)
}

// DeleteObjects 批量删除对象
func (g *GCSClient) DeleteObjects(bucketName string, fileKeys []string) ([]string, error) {
	var failedKeys []string
	bucket := g.client.Bucket(bucketName)
	
	for _, key := range fileKeys {
		obj := bucket.Object(key)
		if err := obj.Delete(g.ctx); err != nil {
			failedKeys = append(failedKeys, key)
		}
	}
	
	return failedKeys, nil
}

// ===== 文件管理操作 =====

// ListObjects 列举对象
func (g *GCSClient) ListObjects(input *storage.ListObjectsInput) (*storage.ListObjectsOutput, error) {
	bucket := g.client.Bucket(input.Bucket)
	
	query := &gcs.Query{
		Prefix:    input.Prefix,
		Delimiter: input.Delimiter,
	}
	
	if input.StartAfter != "" {
		query.StartOffset = input.StartAfter
	}
	
	it := bucket.Objects(g.ctx, query)
	
	var objects []storage.ObjectInfo
	var commonPrefixes []string
	var keyCount int32
	
	for {
		if input.MaxKeys > 0 && keyCount >= input.MaxKeys {
			break
		}
		
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		
		if attrs.Prefix != "" {
			// 这是一个公共前缀（目录）
			commonPrefixes = append(commonPrefixes, attrs.Prefix)
		} else {
			// 这是一个对象
			objects = append(objects, storage.ObjectInfo{
				Key:          attrs.Name,
				Size:         attrs.Size,
				LastModified: attrs.Updated,
				ETag:         attrs.Etag,
				ContentType:  attrs.ContentType,
				IsDir:        strings.HasSuffix(attrs.Name, "/"),
			})
		}
		keyCount++
	}
	
	return &storage.ListObjectsOutput{
		Objects:        objects,
		CommonPrefixes: commonPrefixes,
		IsTruncated:    false, // GCS 不直接支持分页，这里简化处理
		KeyCount:       keyCount,
	}, nil
}

// CopyObject 复制对象
func (g *GCSClient) CopyObject(input *storage.CopyObjectInput) error {
	srcBucket := g.client.Bucket(input.SourceBucket)
	srcObj := srcBucket.Object(input.SourceKey)
	
	dstBucket := g.client.Bucket(input.DestinationBucket)
	dstObj := dstBucket.Object(input.DestinationKey)
	
	copier := dstObj.CopierFrom(srcObj)
	
	if input.ContentType != "" {
		copier.ContentType = input.ContentType
	}
	
	if input.Metadata != nil {
		copier.Metadata = input.Metadata
	}
	
	_, err := copier.Run(g.ctx)
	return err
}

// MoveObject 移动对象（复制后删除源对象）
func (g *GCSClient) MoveObject(sourceBucket, sourceKey, destBucket, destKey string) error {
	// 先复制
	copyInput := &storage.CopyObjectInput{
		SourceBucket:      sourceBucket,
		SourceKey:         sourceKey,
		DestinationBucket: destBucket,
		DestinationKey:    destKey,
	}
	
	if err := g.CopyObject(copyInput); err != nil {
		return err
	}
	
	// 再删除源对象
	return g.DeleteObject(sourceBucket, sourceKey)
}

// GetObjectMetadata 获取对象元数据
func (g *GCSClient) GetObjectMetadata(bucketName, fileKey string) (*storage.ObjectMetadata, error) {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	attrs, err := obj.Attrs(g.ctx)
	if err != nil {
		return nil, err
	}
	
	return &storage.ObjectMetadata{
		ContentType:          attrs.ContentType,
		ContentLength:        attrs.Size,
		LastModified:         attrs.Updated,
		ETag:                 attrs.Etag,
		Metadata:             attrs.Metadata,
		StorageClass:         string(attrs.StorageClass),
		ServerSideEncryption: attrs.KMSKeyName,
	}, nil
}

// ===== 目录操作 =====

// CreateFolder 创建文件夹（通过创建以/结尾的空对象）
func (g *GCSClient) CreateFolder(bucketName, folderPath string) error {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}
	
	return g.UploadObject(bucketName, folderPath, []byte{})
}

// DeleteFolder 删除文件夹及其所有内容
func (g *GCSClient) DeleteFolder(bucketName, folderPath string) error {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}
	
	// 列出所有以该前缀开头的对象
	input := &storage.ListObjectsInput{
		Bucket: bucketName,
		Prefix: folderPath,
	}
	
	output, err := g.ListObjects(input)
	if err != nil {
		return err
	}
	
	// 删除所有对象
	var keys []string
	for _, obj := range output.Objects {
		keys = append(keys, obj.Key)
	}
	
	if len(keys) > 0 {
		_, err = g.DeleteObjects(bucketName, keys)
	}
	
	return err
}

// ListFolders 列举文件夹
func (g *GCSClient) ListFolders(bucketName, prefix string) ([]string, error) {
	input := &storage.ListObjectsInput{
		Bucket:    bucketName,
		Prefix:    prefix,
		Delimiter: "/",
	}
	
	output, err := g.ListObjects(input)
	if err != nil {
		return nil, err
	}
	
	return output.CommonPrefixes, nil
}

// ===== 预签名URL操作 =====

// PreSignPutObject 生成预签名上传URL
func (g *GCSClient) PreSignPutObject(bucketName, fileKey string) (string, error) {
	opts := &gcs.SignedURLOptions{
		Scheme:  gcs.SigningSchemeV4,
		Method:  "PUT",
		Expires: time.Now().Add(15 * time.Minute),
	}
	
	return g.client.Bucket(bucketName).SignedURL(fileKey, opts)
}

// BatchPreSignPutObject 批量生成预签名上传URL
func (g *GCSClient) BatchPreSignPutObject(bucketName string, fileKeys []string, isWholeKey bool) map[string]string {
	result := make(map[string]string)
	
	for _, key := range fileKeys {
		url, err := g.PreSignPutObject(bucketName, key)
		if err == nil {
			if isWholeKey {
				result[key] = url
			} else {
				// 如果不是完整键，可能需要处理键名
				result[key] = url
			}
		}
	}
	
	return result
}

// PreSignGetObject 生成预签名获取URL
func (g *GCSClient) PreSignGetObject(bucketName, fileKey string) (string, error) {
	opts := &gcs.SignedURLOptions{
		Scheme:  gcs.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	}
	
	return g.client.Bucket(bucketName).SignedURL(fileKey, opts)
}

// PreSignDeleteObject 生成预签名删除URL
func (g *GCSClient) PreSignDeleteObject(bucketName, fileKey string) (string, error) {
	opts := &gcs.SignedURLOptions{
		Scheme:  gcs.SigningSchemeV4,
		Method:  "DELETE",
		Expires: time.Now().Add(15 * time.Minute),
	}
	
	return g.client.Bucket(bucketName).SignedURL(fileKey, opts)
}

// ===== 高级功能 =====

// SetObjectACL 设置对象访问控制列表
func (g *GCSClient) SetObjectACL(bucketName, fileKey, acl string) error {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	var aclRule gcs.ACLRule
	
	switch acl {
	case "public-read":
		aclRule = gcs.ACLRule{
			Entity: gcs.AllUsers,
			Role:   gcs.RoleReader,
		}
	case "private":
		// 移除公共访问权限
		return obj.ACL().Delete(g.ctx, gcs.AllUsers)
	default:
		return fmt.Errorf("unsupported ACL: %s", acl)
	}
	
	return obj.ACL().Set(g.ctx, aclRule.Entity, aclRule.Role)
}

// GetObjectACL 获取对象访问控制列表
func (g *GCSClient) GetObjectACL(bucketName, fileKey string) (string, error) {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	rules, err := obj.ACL().List(g.ctx)
	if err != nil {
		return "", err
	}
	
	// 检查是否有公共读取权限
	for _, rule := range rules {
		if rule.Entity == gcs.AllUsers && rule.Role == gcs.RoleReader {
			return "public-read", nil
		}
	}
	
	return "private", nil
}

// SetObjectMetadata 设置对象元数据
func (g *GCSClient) SetObjectMetadata(bucketName, fileKey string, metadata map[string]string) error {
	bucket := g.client.Bucket(bucketName)
	obj := bucket.Object(fileKey)
	
	attrs := gcs.ObjectAttrsToUpdate{
		Metadata: metadata,
	}
	
	_, err := obj.Update(g.ctx, attrs)
	return err
}

// GenerateDownloadURL 生成直接下载链接（公共读取）
func (g *GCSClient) GenerateDownloadURL(bucketName, fileKey string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, fileKey)
}