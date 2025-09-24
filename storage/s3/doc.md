# S3 文件管理器文档

## 概述

S3 文件管理器是一个基于 AWS SDK v2 构建的完整文件管理解决方案，提供了丰富的文件操作功能，包括基础的上传下载、高级的文件管理、目录操作、预签名URL生成等功能。

## 功能特性

### 🚀 基础文件操作
- **文件上传**: 支持字节数组和流式上传
- **文件下载**: 高效的文件下载功能
- **文件检查**: 快速检查文件是否存在
- **文件删除**: 单个和批量删除功能

### 📁 文件管理功能
- **列出对象**: 支持前缀过滤和分页的对象列表
- **复制对象**: 在存储桶内或跨存储桶复制文件
- **移动对象**: 文件移动和重命名功能
- **元数据管理**: 获取和设置对象元数据

### 🗂️ 目录操作
- **创建文件夹**: 创建虚拟文件夹结构
- **删除文件夹**: 递归删除文件夹及其内容
- **列出文件夹**: 获取指定路径下的文件夹列表

### 🔗 预签名URL功能
- **上传预签名URL**: 生成安全的上传链接
- **下载预签名URL**: 生成临时下载链接
- **删除预签名URL**: 生成删除操作链接
- **批量预签名URL**: 批量生成多个文件的预签名URL

### 🔐 高级功能
- **ACL管理**: 设置和获取对象访问控制列表
- **自定义元数据**: 为对象添加自定义元数据
- **直接下载URL**: 生成公共访问的下载链接

## 快速开始

### 1. 环境配置

设置必要的环境变量：

```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"  # 可选，默认为 us-east-1
```

### 2. 创建服务实例

```go
import (
    "github.com/QingsiLiu/baseComponents/storage/s3"
)

// 创建普通实例
s3Service, err := s3.NewS3Service("us-east-1")
if err != nil {
    log.Fatal(err)
}

// 或使用单例模式
s3Service := s3.NewS3Svc("us-east-1")
```

### 3. 基础使用示例

```go
bucketName := "your-bucket-name"
key := "path/to/your/file.txt"
data := []byte("Hello, S3!")

// 上传文件
err := s3Service.UploadObject(bucketName, key, data)
if err != nil {
    log.Printf("上传失败: %v", err)
}

// 下载文件
downloadedData, err := s3Service.GetObject(bucketName, key)
if err != nil {
    log.Printf("下载失败: %v", err)
}

// 检查文件是否存在
exists := s3Service.HeadObject(bucketName, key)
if exists {
    fmt.Println("文件存在")
}
```

## 详细API文档

### 基础文件操作

#### UploadObject
上传字节数组到S3

```go
func (s *S3Service) UploadObject(bucket, key string, data []byte) error
```

**参数:**
- `bucket`: 存储桶名称
- `key`: 对象键（文件路径）
- `data`: 要上传的字节数据

#### UploadObjectStream
流式上传文件到S3

```go
func (s *S3Service) UploadObjectStream(bucket, key string, reader io.Reader) error
```

**参数:**
- `bucket`: 存储桶名称
- `key`: 对象键（文件路径）
- `reader`: 数据流读取器

#### GetObject
从S3下载文件

```go
func (s *S3Service) GetObject(bucket, key string) ([]byte, error)
```

**返回:**
- `[]byte`: 文件内容
- `error`: 错误信息

#### HeadObject
检查对象是否存在

```go
func (s *S3Service) HeadObject(bucket, key string) bool
```

**返回:**
- `bool`: 文件是否存在

### 文件管理功能

#### ListObjects
列出存储桶中的对象

```go
func (s *S3Service) ListObjects(input *storage.ListObjectsInput) (*storage.ListObjectsOutput, error)
```

**输入结构:**
```go
type ListObjectsInput struct {
    Bucket      string
    Prefix      string
    Delimiter   string
    MaxKeys     int32
    StartAfter  string
}
```

**输出结构:**
```go
type ListObjectsOutput struct {
    Objects           []ObjectInfo
    CommonPrefixes    []string
    IsTruncated       bool
    NextContinuationToken string
}
```

#### CopyObject
复制对象

```go
func (s *S3Service) CopyObject(input *storage.CopyObjectInput) error
```

**输入结构:**
```go
type CopyObjectInput struct {
    SourceBucket      string
    SourceKey         string
    DestinationBucket string
    DestinationKey    string
    Metadata          map[string]string
}
```

#### MoveObject
移动对象

```go
func (s *S3Service) MoveObject(sourceBucket, sourceKey, destBucket, destKey string) error
```

#### DeleteObject
删除单个对象

```go
func (s *S3Service) DeleteObject(bucket, key string) error
```

#### DeleteObjects
批量删除对象

```go
func (s *S3Service) DeleteObjects(bucket string, keys []string) ([]string, error)
```

**返回:**
- `[]string`: 成功删除的对象键列表
- `error`: 错误信息

### 目录操作

#### CreateFolder
创建文件夹

```go
func (s *S3Service) CreateFolder(bucket, folderPath string) error
```

**注意:** `folderPath` 必须以 `/` 结尾

#### DeleteFolder
删除文件夹及其所有内容

```go
func (s *S3Service) DeleteFolder(bucket, folderPath string) error
```

#### ListFolders
列出指定路径下的文件夹

```go
func (s *S3Service) ListFolders(bucket, prefix string) ([]string, error)
```

### 预签名URL功能

#### PreSignPutObject
生成上传预签名URL

```go
func (s *S3Service) PreSignPutObject(bucket, key string) (string, error)
```

#### PreSignGetObject
生成下载预签名URL

```go
func (s *S3Service) PreSignGetObject(bucket, key string) (string, error)
```

#### PreSignDeleteObject
生成删除预签名URL

```go
func (s *S3Service) PreSignDeleteObject(bucket, key string) (string, error)
```

#### BatchPreSignPutObject
批量生成上传预签名URL

```go
func (s *S3Service) BatchPreSignPutObject(bucket string, keys []string, isPublic bool) map[string]string
```

**参数:**
- `bucket`: 存储桶名称
- `keys`: 对象键列表
- `isPublic`: 是否为公共访问

**返回:**
- `map[string]string`: 键值对映射，键为对象键，值为预签名URL

### 高级功能

#### GetObjectMetadata
获取对象元数据

```go
func (s *S3Service) GetObjectMetadata(bucket, key string) (*storage.ObjectMetadata, error)
```

**返回结构:**
```go
type ObjectMetadata struct {
    ContentLength   int64
    ContentType     string
    ETag            string
    LastModified    time.Time
    Metadata        map[string]string
}
```

#### SetObjectMetadata
设置对象自定义元数据

```go
func (s *S3Service) SetObjectMetadata(bucket, key string, metadata map[string]string) error
```

#### SetObjectACL
设置对象访问控制列表

```go
func (s *S3Service) SetObjectACL(bucket, key, acl string) error
```

**支持的ACL值:**
- `private`: 私有访问
- `public-read`: 公共读取
- `public-read-write`: 公共读写
- `authenticated-read`: 认证用户读取

#### GetObjectACL
获取对象访问控制列表

```go
func (s *S3Service) GetObjectACL(bucket, key string) (string, error)
```

#### GenerateDownloadURL
生成公共下载URL

```go
func (s *S3Service) GenerateDownloadURL(bucket, key string) string
```

## 使用示例

### 完整的文件管理示例

查看 `examples/s3_file_manager_example.go` 文件，其中包含了所有功能的详细使用示例：

- 基础文件操作演示
- 文件管理功能演示
- 目录操作演示
- 预签名URL功能演示
- 高级功能演示

### 运行示例

```bash
# 设置环境变量
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"

# 运行示例
go run examples/s3_file_manager_example.go
```

## 错误处理

所有方法都返回标准的Go错误类型。建议在生产环境中进行适当的错误处理：

```go
if err := s3Service.UploadObject(bucket, key, data); err != nil {
    // 根据错误类型进行不同的处理
    if strings.Contains(err.Error(), "NoSuchBucket") {
        log.Printf("存储桶不存在: %v", err)
    } else if strings.Contains(err.Error(), "AccessDenied") {
        log.Printf("访问被拒绝: %v", err)
    } else {
        log.Printf("上传失败: %v", err)
    }
}
```

## 性能优化建议

### 1. 使用单例模式
对于频繁的S3操作，建议使用单例模式以减少客户端创建开销：

```go
s3Service := s3.NewS3Svc("us-east-1")
```

### 2. 流式上传大文件
对于大文件，使用流式上传以减少内存使用：

```go
file, err := os.Open("large-file.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

err = s3Service.UploadObjectStream(bucket, key, file)
```

### 3. 批量操作
对于多个文件的操作，使用批量方法以提高效率：

```go
// 批量删除
deletedKeys, err := s3Service.DeleteObjects(bucket, keys)

// 批量生成预签名URL
urlMap := s3Service.BatchPreSignPutObject(bucket, keys, false)
```

### 4. 合理使用前缀和分页
在列出大量对象时，使用前缀过滤和分页：

```go
listInput := &storage.ListObjectsInput{
    Bucket:  bucket,
    Prefix:  "logs/2024/",
    MaxKeys: 1000,
}
```

## 安全注意事项

### 1. 凭证管理
- 永远不要在代码中硬编码AWS凭证
- 使用环境变量或AWS凭证文件
- 在生产环境中使用IAM角色

### 2. 访问控制
- 合理设置对象ACL
- 使用预签名URL时设置适当的过期时间
- 定期审查和更新访问权限

### 3. 数据加密
- 考虑使用服务端加密（SSE）
- 对敏感数据使用客户端加密

## 测试

运行测试套件：

```bash
# 运行所有测试
go test ./storage/s3/ -v

# 运行集成测试（需要真实的AWS凭证）
export AWS_TEST_BUCKET="your-test-bucket"
go test ./storage/s3/ -v -run TestS3ServiceWithRealAWS
```

## 故障排除

### 常见问题

1. **凭证错误**
   ```
   Error: NoCredentialProviders
   ```
   解决方案：确保设置了正确的AWS凭证

2. **存储桶不存在**
   ```
   Error: NoSuchBucket
   ```
   解决方案：确保存储桶名称正确且存在

3. **权限不足**
   ```
   Error: AccessDenied
   ```
   解决方案：检查IAM权限设置

4. **区域不匹配**
   ```
   Error: AuthorizationHeaderMalformed
   ```
   解决方案：确保指定了正确的AWS区域

### 调试技巧

启用AWS SDK调试日志：

```go
import (
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
)

cfg, err := config.LoadDefaultConfig(context.TODO(),
    config.WithRegion(region),
    config.WithClientLogMode(aws.LogRetries|aws.LogRequest|aws.LogResponse),
)
```

## 版本历史

- **v2.0.0**: 升级到AWS SDK v2，添加完整的文件管理器功能
- **v1.0.0**: 基于AWS SDK v1的基础实现

## 贡献

欢迎提交问题和改进建议！请确保：

1. 遵循现有的代码风格
2. 添加适当的测试
3. 更新相关文档

## 许可证

本项目采用 MIT 许可证。详见 LICENSE 文件。