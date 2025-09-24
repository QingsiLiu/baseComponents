package s3

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QingsiLiu/baseComponents/storage"
	"github.com/aws/aws-sdk-go-v2/config"
)

// TestNewS3Service 测试创建S3服务实例
func TestNewS3Service(t *testing.T) {
	service, err := NewS3Service("us-east-1")
	if err != nil {
		t.Fatalf("Failed to create S3 service: %v", err)
	}

	if service == nil {
		t.Fatal("S3 service is nil")
	}

	if service.client == nil {
		t.Fatal("S3 client is nil")
	}

	if service.uploader == nil {
		t.Fatal("S3 uploader is nil")
	}
}

// TestNewS3Svc 测试单例模式
func TestNewS3Svc(t *testing.T) {
	service1 := NewS3Svc("us-east-1")
	service2 := NewS3Svc("us-east-1")

	if service1 != service2 {
		t.Fatal("NewS3Svc should return the same instance (singleton)")
	}
}

// TestS3ServiceWithRealAWS 集成测试（需要真实的AWS凭证和测试桶）
func TestS3ServiceWithRealAWS(t *testing.T) {
	// 检查环境变量
	bucketName := os.Getenv("TEST_S3_BUCKET")
	if bucketName == "" {
		t.Skip("Skipping integration test: TEST_S3_BUCKET environment variable not set")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	service, err := NewS3Service(region)
	if err != nil {
		t.Fatalf("Failed to create S3 service: %v", err)
	}

	testKey := fmt.Sprintf("test-file-%d.txt", time.Now().Unix())
	testData := []byte("Hello, S3 File Manager!")

	t.Run("UploadAndDownload", func(t *testing.T) {
		// 上传文件
		err := service.UploadObject(bucketName, testKey, testData)
		if err != nil {
			t.Fatalf("Failed to upload object: %v", err)
		}

		// 检查文件是否存在
		exists := service.HeadObject(bucketName, testKey)
		if !exists {
			t.Fatal("Object should exist after upload")
		}

		// 下载文件
		downloadedData, err := service.GetObject(bucketName, testKey)
		if err != nil {
			t.Fatalf("Failed to get object: %v", err)
		}

		if !bytes.Equal(testData, downloadedData) {
			t.Fatal("Downloaded data doesn't match uploaded data")
		}

		// 清理
		defer func() {
			err := service.DeleteObject(bucketName, testKey)
			if err != nil {
				t.Logf("Warning: Failed to cleanup test object: %v", err)
			}
		}()
	})

	t.Run("ListObjects", func(t *testing.T) {
		// 创建测试文件
		testPrefix := fmt.Sprintf("test-list-%d/", time.Now().Unix())
		testFiles := []string{
			testPrefix + "file1.txt",
			testPrefix + "file2.txt",
			testPrefix + "subfolder/file3.txt",
		}

		// 上传测试文件
		for _, key := range testFiles {
			err := service.UploadObject(bucketName, key, []byte("test content"))
			if err != nil {
				t.Fatalf("Failed to upload test file %s: %v", key, err)
			}
		}

		// 列出对象
		listInput := &storage.ListObjectsInput{
			Bucket:  bucketName,
			Prefix:  testPrefix,
			MaxKeys: 10,
		}

		result, err := service.ListObjects(listInput)
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		if len(result.Objects) < 3 {
			t.Fatalf("Expected at least 3 objects, got %d", len(result.Objects))
		}

		// 清理
		defer func() {
			for _, key := range testFiles {
				err := service.DeleteObject(bucketName, key)
				if err != nil {
					t.Logf("Warning: Failed to cleanup test file %s: %v", key, err)
				}
			}
		}()
	})

	t.Run("CopyAndMoveObject", func(t *testing.T) {
		sourceKey := fmt.Sprintf("test-source-%d.txt", time.Now().Unix())
		copyKey := fmt.Sprintf("test-copy-%d.txt", time.Now().Unix())
		moveKey := fmt.Sprintf("test-move-%d.txt", time.Now().Unix())

		// 上传源文件
		err := service.UploadObject(bucketName, sourceKey, testData)
		if err != nil {
			t.Fatalf("Failed to upload source object: %v", err)
		}

		// 复制对象
		copyInput := &storage.CopyObjectInput{
			SourceBucket:      bucketName,
			SourceKey:         sourceKey,
			DestinationBucket: bucketName,
			DestinationKey:    copyKey,
		}

		err = service.CopyObject(copyInput)
		if err != nil {
			t.Fatalf("Failed to copy object: %v", err)
		}

		// 验证复制的文件存在
		exists := service.HeadObject(bucketName, copyKey)
		if !exists {
			t.Fatal("Copied object should exist")
		}

		// 移动对象
		err = service.MoveObject(bucketName, sourceKey, bucketName, moveKey)
		if err != nil {
			t.Fatalf("Failed to move object: %v", err)
		}

		// 验证移动后的状态
		sourceExists := service.HeadObject(bucketName, sourceKey)
		moveExists := service.HeadObject(bucketName, moveKey)

		if sourceExists {
			t.Fatal("Source object should not exist after move")
		}
		if !moveExists {
			t.Fatal("Moved object should exist")
		}

		// 清理
		defer func() {
			service.DeleteObject(bucketName, copyKey)
			service.DeleteObject(bucketName, moveKey)
		}()
	})

	t.Run("FolderOperations", func(t *testing.T) {
		folderPath := fmt.Sprintf("test-folder-%d/", time.Now().Unix())

		// 创建文件夹
		err := service.CreateFolder(bucketName, folderPath)
		if err != nil {
			t.Fatalf("Failed to create folder: %v", err)
		}

		// 验证文件夹存在
		exists := service.HeadObject(bucketName, folderPath)
		if !exists {
			t.Fatal("Folder should exist after creation")
		}

		// 在文件夹中创建文件
		fileInFolder := folderPath + "test-file.txt"
		err = service.UploadObject(bucketName, fileInFolder, testData)
		if err != nil {
			t.Fatalf("Failed to upload file in folder: %v", err)
		}

		// 列出文件夹
		folders, err := service.ListFolders(bucketName, "")
		if err != nil {
			t.Fatalf("Failed to list folders: %v", err)
		}

		found := false
		for _, folder := range folders {
			if strings.HasPrefix(folder, "test-folder-") {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("Created folder should be in the list")
		}

		// 删除文件夹
		err = service.DeleteFolder(bucketName, folderPath)
		if err != nil {
			t.Fatalf("Failed to delete folder: %v", err)
		}

		// 验证文件夹和文件都被删除
		folderExists := service.HeadObject(bucketName, folderPath)
		fileExists := service.HeadObject(bucketName, fileInFolder)

		if folderExists || fileExists {
			t.Fatal("Folder and its contents should be deleted")
		}
	})

	t.Run("PreSignedURLs", func(t *testing.T) {
		testKey := fmt.Sprintf("test-presign-%d.txt", time.Now().Unix())

		// 生成上传预签名URL
		putURL, err := service.PreSignPutObject(bucketName, testKey)
		if err != nil {
			t.Fatalf("Failed to generate pre-signed PUT URL: %v", err)
		}

		if putURL == "" {
			t.Fatal("Pre-signed PUT URL should not be empty")
		}

		// 上传文件用于测试下载预签名URL
		err = service.UploadObject(bucketName, testKey, testData)
		if err != nil {
			t.Fatalf("Failed to upload test object: %v", err)
		}

		// 生成下载预签名URL
		getURL, err := service.PreSignGetObject(bucketName, testKey)
		if err != nil {
			t.Fatalf("Failed to generate pre-signed GET URL: %v", err)
		}

		if getURL == "" {
			t.Fatal("Pre-signed GET URL should not be empty")
		}

		// 生成删除预签名URL
		deleteURL, err := service.PreSignDeleteObject(bucketName, testKey)
		if err != nil {
			t.Fatalf("Failed to generate pre-signed DELETE URL: %v", err)
		}

		if deleteURL == "" {
			t.Fatal("Pre-signed DELETE URL should not be empty")
		}

		// 测试GenerateDownloadURL
		downloadURL := service.GenerateDownloadURL(bucketName, testKey)
		if downloadURL == "" {
			t.Fatal("Download URL should not be empty")
		}

		// 清理
		defer func() {
			service.DeleteObject(bucketName, testKey)
		}()
	})

	t.Run("BatchPreSignedURLs", func(t *testing.T) {
		fileKeys := []string{
			fmt.Sprintf("batch-test-1-%d.txt", time.Now().Unix()),
			fmt.Sprintf("batch-test-2-%d.txt", time.Now().Unix()),
			fmt.Sprintf("batch-test-3-%d.txt", time.Now().Unix()),
		}

		// 生成批量预签名URL
		urlMap := service.BatchPreSignPutObject(bucketName, fileKeys, true)

		if len(urlMap) != len(fileKeys) {
			t.Fatalf("Expected %d URLs, got %d", len(fileKeys), len(urlMap))
		}

		for _, key := range fileKeys {
			if url, exists := urlMap[key]; !exists || url == "" {
				t.Fatalf("Missing or empty URL for key: %s", key)
			}
		}
	})

	t.Run("ObjectMetadata", func(t *testing.T) {
		testKey := fmt.Sprintf("test-metadata-%d.txt", time.Now().Unix())

		// 上传文件
		err := service.UploadObject(bucketName, testKey, testData)
		if err != nil {
			t.Fatalf("Failed to upload object: %v", err)
		}

		// 获取对象元数据
		metadata, err := service.GetObjectMetadata(bucketName, testKey)
		if err != nil {
			t.Fatalf("Failed to get object metadata: %v", err)
		}

		if metadata.ContentLength != int64(len(testData)) {
			t.Fatalf("Expected content length %d, got %d", len(testData), metadata.ContentLength)
		}

		if metadata.ETag == "" {
			t.Fatal("ETag should not be empty")
		}

		// 设置对象元数据
		customMetadata := map[string]string{
			"author":      "test-user",
			"description": "test file for metadata",
		}

		err = service.SetObjectMetadata(bucketName, testKey, customMetadata)
		if err != nil {
			t.Fatalf("Failed to set object metadata: %v", err)
		}

		// 验证元数据已设置
		updatedMetadata, err := service.GetObjectMetadata(bucketName, testKey)
		if err != nil {
			t.Fatalf("Failed to get updated metadata: %v", err)
		}

		for key, expectedValue := range customMetadata {
			if actualValue, exists := updatedMetadata.Metadata[key]; !exists || actualValue != expectedValue {
				t.Fatalf("Expected metadata %s=%s, got %s=%s", key, expectedValue, key, actualValue)
			}
		}

		// 清理
		defer func() {
			service.DeleteObject(bucketName, testKey)
		}()
	})

	t.Run("BatchDeleteObjects", func(t *testing.T) {
		// 创建多个测试文件
		testKeys := []string{
			fmt.Sprintf("batch-delete-1-%d.txt", time.Now().Unix()),
			fmt.Sprintf("batch-delete-2-%d.txt", time.Now().Unix()),
			fmt.Sprintf("batch-delete-3-%d.txt", time.Now().Unix()),
		}

		// 上传测试文件
		for _, key := range testKeys {
			err := service.UploadObject(bucketName, key, testData)
			if err != nil {
				t.Fatalf("Failed to upload test file %s: %v", key, err)
			}
		}

		// 批量删除
		deletedKeys, err := service.DeleteObjects(bucketName, testKeys)
		if err != nil {
			t.Fatalf("Failed to batch delete objects: %v", err)
		}

		if len(deletedKeys) != len(testKeys) {
			t.Fatalf("Expected %d deleted keys, got %d", len(testKeys), len(deletedKeys))
		}

		// 验证文件已删除
		for _, key := range testKeys {
			exists := service.HeadObject(bucketName, key)
			if exists {
				t.Fatalf("Object %s should be deleted", key)
			}
		}
	})
}

// BenchmarkS3ServiceCreation 性能测试
func BenchmarkS3ServiceCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		service, err := NewS3Service("us-east-1")
		if err != nil {
			b.Fatalf("Failed to create S3 service: %v", err)
		}
		_ = service
	}
}

// BenchmarkS3ServiceSingleton 单例性能测试
func BenchmarkS3ServiceSingleton(b *testing.B) {
	for i := 0; i < b.N; i++ {
		service := NewS3Svc("us-east-1")
		_ = service
	}
}

// TestConfigLoading 测试配置加载
func TestConfigLoading(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		t.Logf("Warning: Failed to load AWS config: %v", err)
		return
	}

	if cfg.Region == "" {
		t.Log("Warning: No AWS region configured")
	}
}
