package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/QingsiLiu/baseComponents/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Service S3存储服务
type S3Service struct {
	client   *s3.Client
	uploader *manager.Uploader
}

var s3Svc *S3Service
var onceS3Svc sync.Once

// NewS3Svc 创建一个新的S3服务实例（单例模式）
func NewS3Svc(region string) storage.StorageService {
	onceS3Svc.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			panic(fmt.Sprintf("failed to load AWS config: %v", err))
		}

		client := s3.NewFromConfig(cfg)
		uploader := manager.NewUploader(client)

		s3Svc = &S3Service{
			client:   client,
			uploader: uploader,
		}
	})
	return s3Svc
}

// NewS3Service 创建S3服务实例（非单例）
func NewS3Service(region string) (*S3Service, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	return &S3Service{
		client:   client,
		uploader: uploader,
	}, nil
}

// UploadObject 上传文件到S3
func (s *S3Service) UploadObject(bucketName, fileKey string, data []byte) error {
	contentType := storage.GetContentType(fileKey)

	_, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(fileKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})

	return err
}

// UploadObjectStream 流式上传文件到S3
func (s *S3Service) UploadObjectStream(bucketName, fileKey string, file io.Reader) error {
	contentType := storage.GetContentType(fileKey)

	_, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(fileKey),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	return err
}

// GetObject 从S3获取文件
func (s *S3Service) GetObject(bucketName, fileKey string) ([]byte, error) {
	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// PreSignPutObject 生成预签名上传URL
func (s *S3Service) PreSignPutObject(bucketName, fileKey string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(15 * time.Minute)
	})

	if err != nil {
		return "", err
	}

	return request.URL, nil
}

// BatchPreSignPutObject 批量生成预签名上传URL
func (s *S3Service) BatchPreSignPutObject(bucketName string, fileKeys []string, isWholeKey bool) map[string]string {
	result := make(map[string]string)
	presignClient := s3.NewPresignClient(s.client)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, fileKey := range fileKeys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			actualKey := key
			if !isWholeKey {
				// 如果不是完整key，可能需要添加前缀或后缀
				actualKey = key
			}

			request, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(actualKey),
			}, func(opts *s3.PresignOptions) {
				opts.Expires = time.Duration(15 * time.Minute)
			})

			mu.Lock()
			if err != nil {
				result[key] = ""
			} else {
				result[key] = request.URL
			}
			mu.Unlock()
		}(fileKey)
	}

	wg.Wait()
	return result
}

// PreSignGetObject 生成预签名获取URL
func (s *S3Service) PreSignGetObject(bucketName, fileKey string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(15 * time.Minute)
	})

	if err != nil {
		return "", err
	}

	return request.URL, nil
}

// HeadObject 检查对象是否存在
func (s *S3Service) HeadObject(bucketName, fileKey string) bool {
	_, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	})

	if err != nil {
		// 检查是否是NoSuchKey错误或NotFound错误
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return false
		}
		// 其他错误也认为对象不存在
		return false
	}

	return true
}

// DeleteObject 删除单个对象
func (s *S3Service) DeleteObject(bucketName, fileKey string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	})
	return err
}

// DeleteObjects 批量删除对象
func (s *S3Service) DeleteObjects(bucketName string, fileKeys []string) ([]string, error) {
	if len(fileKeys) == 0 {
		return []string{}, nil
	}

	// 构建删除对象列表
	var objectsToDelete []types.ObjectIdentifier
	for _, key := range fileKeys {
		objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
			Key: aws.String(key),
		})
	}

	// 执行批量删除
	result, err := s.client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{
			Objects: objectsToDelete,
		},
	})

	if err != nil {
		return nil, err
	}

	// 收集成功删除的对象键
	var deletedKeys []string
	for _, deleted := range result.Deleted {
		if deleted.Key != nil {
			deletedKeys = append(deletedKeys, *deleted.Key)
		}
	}

	return deletedKeys, nil
}

// ListObjects 列出对象
func (s *S3Service) ListObjects(input *storage.ListObjectsInput) (*storage.ListObjectsOutput, error) {
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(input.Bucket),
	}

	if input.Prefix != "" {
		listInput.Prefix = aws.String(input.Prefix)
	}
	if input.Delimiter != "" {
		listInput.Delimiter = aws.String(input.Delimiter)
	}
	if input.MaxKeys > 0 {
		listInput.MaxKeys = aws.Int32(input.MaxKeys)
	}
	if input.ContinuationToken != "" {
		listInput.ContinuationToken = aws.String(input.ContinuationToken)
	}

	result, err := s.client.ListObjectsV2(context.TODO(), listInput)
	if err != nil {
		return nil, err
	}

	output := &storage.ListObjectsOutput{
		IsTruncated: aws.ToBool(result.IsTruncated),
		KeyCount:    aws.ToInt32(result.KeyCount),
	}

	if result.NextContinuationToken != nil {
		output.NextContinuationToken = *result.NextContinuationToken
	}

	// 转换对象信息
	for _, obj := range result.Contents {
		objInfo := storage.ObjectInfo{
			Key:          *obj.Key,
			Size:         aws.ToInt64(obj.Size),
			LastModified: *obj.LastModified,
		}
		if obj.ETag != nil {
			objInfo.ETag = *obj.ETag
		}
		output.Objects = append(output.Objects, objInfo)
	}

	// 转换公共前缀（文件夹）
	for _, prefix := range result.CommonPrefixes {
		if prefix.Prefix != nil {
			output.CommonPrefixes = append(output.CommonPrefixes, *prefix.Prefix)
		}
	}

	return output, nil
}

// CopyObject 复制对象
func (s *S3Service) CopyObject(input *storage.CopyObjectInput) error {
	copySource := fmt.Sprintf("%s/%s", input.SourceBucket, input.SourceKey)
	
	_, err := s.client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:     aws.String(input.DestinationBucket),
		Key:        aws.String(input.DestinationKey),
		CopySource: aws.String(copySource),
	})
	
	return err
}

// MoveObject 移动对象（复制后删除源对象）
func (s *S3Service) MoveObject(sourceBucket, sourceKey, destBucket, destKey string) error {
	// 先复制对象
	copyInput := &storage.CopyObjectInput{
		SourceBucket:      sourceBucket,
		SourceKey:         sourceKey,
		DestinationBucket: destBucket,
		DestinationKey:    destKey,
	}
	
	err := s.CopyObject(copyInput)
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	// 删除源对象
	err = s.DeleteObject(sourceBucket, sourceKey)
	if err != nil {
		return fmt.Errorf("failed to delete source object: %w", err)
	}

	return nil
}

// GetObjectMetadata 获取对象元数据
func (s *S3Service) GetObjectMetadata(bucketName, fileKey string) (*storage.ObjectMetadata, error) {
	result, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return nil, err
	}

	metadata := &storage.ObjectMetadata{
		ContentLength: aws.ToInt64(result.ContentLength),
		LastModified:  *result.LastModified,
		Metadata:      make(map[string]string),
	}

	if result.ContentType != nil {
		metadata.ContentType = *result.ContentType
	}
	if result.ETag != nil {
		metadata.ETag = *result.ETag
	}
	if result.StorageClass != "" {
		metadata.StorageClass = string(result.StorageClass)
	}

	// 复制用户定义的元数据
	for k, v := range result.Metadata {
		metadata.Metadata[k] = v
	}

	return metadata, nil
}

// CreateFolder 创建文件夹（通过创建一个以/结尾的空对象）
func (s *S3Service) CreateFolder(bucketName, folderPath string) error {
	// 确保文件夹路径以/结尾
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(folderPath),
		Body:   bytes.NewReader([]byte{}),
	})

	return err
}

// DeleteFolder 删除文件夹及其所有内容
func (s *S3Service) DeleteFolder(bucketName, folderPath string) error {
	// 确保文件夹路径以/结尾
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	// 列出文件夹下的所有对象
	listInput := &storage.ListObjectsInput{
		Bucket:  bucketName,
		Prefix:  folderPath,
		MaxKeys: 1000,
	}

	var allKeys []string
	for {
		result, err := s.ListObjects(listInput)
		if err != nil {
			return err
		}

		// 收集所有对象键
		for _, obj := range result.Objects {
			allKeys = append(allKeys, obj.Key)
		}

		// 如果没有更多对象，退出循环
		if !result.IsTruncated {
			break
		}

		listInput.ContinuationToken = result.NextContinuationToken
	}

	// 批量删除所有对象
	if len(allKeys) > 0 {
		_, err := s.DeleteObjects(bucketName, allKeys)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListFolders 列出文件夹
func (s *S3Service) ListFolders(bucketName, prefix string) ([]string, error) {
	listInput := &storage.ListObjectsInput{
		Bucket:    bucketName,
		Prefix:    prefix,
		Delimiter: "/",
		MaxKeys:   1000,
	}

	result, err := s.ListObjects(listInput)
	if err != nil {
		return nil, err
	}

	return result.CommonPrefixes, nil
}

// PreSignDeleteObject 生成删除对象的预签名URL
func (s *S3Service) PreSignDeleteObject(bucketName, fileKey string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	
	request, err := presignClient.PresignDeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute // 默认15分钟过期
	})
	
	if err != nil {
		return "", err
	}
	
	return request.URL, nil
}

// SetObjectACL 设置对象ACL
func (s *S3Service) SetObjectACL(bucketName, fileKey, acl string) error {
	_, err := s.client.PutObjectAcl(context.TODO(), &s3.PutObjectAclInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
		ACL:    types.ObjectCannedACL(acl),
	})
	return err
}

// GetObjectACL 获取对象ACL
func (s *S3Service) GetObjectACL(bucketName, fileKey string) (string, error) {
	result, err := s.client.GetObjectAcl(context.TODO(), &s3.GetObjectAclInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return "", err
	}

	// 简化返回，实际使用中可能需要更详细的ACL信息
	if result.Owner != nil && result.Owner.DisplayName != nil {
		return *result.Owner.DisplayName, nil
	}
	
	return "unknown", nil
}

// SetObjectMetadata 设置对象元数据
func (s *S3Service) SetObjectMetadata(bucketName, fileKey string, metadata map[string]string) error {
	// 获取当前对象信息
	headResult, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return err
	}

	// 复制对象并设置新的元数据
	copySource := fmt.Sprintf("%s/%s", bucketName, fileKey)
	_, err = s.client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:            aws.String(bucketName),
		Key:               aws.String(fileKey),
		CopySource:        aws.String(copySource),
		Metadata:          metadata,
		MetadataDirective: types.MetadataDirectiveReplace,
		ContentType:       headResult.ContentType,
	})

	return err
}

// GenerateDownloadURL 生成下载URL（与PreSignGetObject相同）
func (s *S3Service) GenerateDownloadURL(bucketName, fileKey string) string {
	url, err := s.PreSignGetObject(bucketName, fileKey)
	if err != nil {
		return ""
	}
	return url
}
