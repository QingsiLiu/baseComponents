# baseComponents

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/QingsiLiu/baseComponents)](https://goreportcard.com/report/github.com/QingsiLiu/baseComponents)

一个高质量的 Golang 基础组件库，旨在减少重复代码编写，提供各种常用的基础组件。

## 🚀 特性

- **模块化设计**: 按功能分类，便于按需使用
- **高性能**: 经过优化的实现，注重性能和内存使用
- **易于使用**: 简洁的 API 设计，丰富的文档和示例
- **全面测试**: 高测试覆盖率，确保代码质量
- **生产就绪**: 适用于生产环境的稳定组件
- **版本管理**: 遵循语义化版本控制，稳定的API

## 📦 已实现的组件模块

### 🛠️ 工具函数 (utils)
- **crypto**: 加密解密、哈希、签名验证
- **strings**: 字符串处理、格式化、验证
- **time**: 时间处理、格式化、时区转换
- **validation**: 数据验证、格式检查

### 💾 存储组件 (storage)
- **S3**: 完整的AWS S3文件管理器，支持文件上传下载、目录操作、预签名URL等
- **Local**: 本地文件存储（规划中）

### 📋 其他组件（规划中）
- **HTTP组件**: 客户端、服务器、中间件
- **数据库组件**: MySQL、Redis、MongoDB
- **基础设施组件**: 配置管理、日志、缓存、消息队列

## 🔧 安装

### 基础安装
```bash
go get github.com/QingsiLiu/baseComponents
```

### 按需安装特定组件
```bash
# 安装S3存储组件
go get github.com/QingsiLiu/baseComponents/storage/s3

# 安装工具函数
go get github.com/QingsiLiu/baseComponents/utils
```

### 版本管理
推荐使用特定版本标签：
```bash
# 安装特定版本
go get github.com/QingsiLiu/baseComponents@v1.0.0

# 安装最新稳定版本
go get github.com/QingsiLiu/baseComponents@latest
```

## 📖 快速开始

### 基础使用示例

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/QingsiLiu/baseComponents/utils"
    "github.com/QingsiLiu/baseComponents/storage/s3"
)

func main() {
    // 使用字符串工具
    if utils.IsValidEmail("user@example.com") {
        fmt.Println("邮箱格式有效")
    }
    
    // 使用S3存储组件
    s3Service, err := s3.NewS3Service("us-east-1")
    if err != nil {
        log.Fatal(err)
    }
    
    // 上传文件
    err = s3Service.UploadObject("my-bucket", "test.txt", []byte("Hello World"))
    if err != nil {
        log.Printf("上传失败: %v", err)
    }
}
```

### S3文件管理器使用示例

```go
package main

import (
    "github.com/QingsiLiu/baseComponents/storage"
    "github.com/QingsiLiu/baseComponents/storage/s3"
)

func main() {
    // 创建S3服务实例
    s3Service, err := s3.NewS3Service("us-east-1")
    if err != nil {
        panic(err)
    }
    
    bucket := "your-bucket"
    
    // 文件操作
    data := []byte("Hello, S3!")
    s3Service.UploadObject(bucket, "path/file.txt", data)
    
    // 列出文件
    listInput := &storage.ListObjectsInput{
        Bucket:  bucket,
        Prefix:  "path/",
        MaxKeys: 10,
    }
    result, _ := s3Service.ListObjects(listInput)
    
    // 生成预签名URL
    url, _ := s3Service.PreSignGetObject(bucket, "path/file.txt")
    fmt.Println("下载链接:", url)
}
```

## 🏗️ 开发

### 环境要求
- Go 1.21+
- Git
- Make (可选，用于运行开发命令)

### 本地开发设置

```bash
# 克隆项目
git clone https://github.com/QingsiLiu/baseComponents.git
cd baseComponents

# 安装依赖
go mod download

# 运行测试
go test ./...

# 格式化代码
go fmt ./...

# 代码检查
go vet ./...
```

### 常用开发命令

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 格式化代码
make fmt

# 代码检查
make lint

# 构建项目
make build

# 清理构建文件
make clean

# 查看所有可用命令
make help
```

## 📁 项目结构

```
baseComponents/
├── storage/            # 存储组件
│   ├── s3/            # AWS S3 存储实现
│   │   ├── s3.go      # S3服务实现
│   │   ├── s3_test.go # S3测试文件
│   │   └── doc.md     # S3文档
│   └── storage.go     # 存储接口定义
├── utils/             # 工具函数
│   ├── crypto.go      # 加密相关工具
│   ├── strings.go     # 字符串处理工具
│   ├── strings_test.go # 字符串测试
│   ├── time.go        # 时间处理工具
│   └── validation.go  # 数据验证工具
├── examples/          # 使用示例
│   └── s3_file_manager_example.go # S3文件管理器示例
├── docs/              # 详细文档
│   └── S3_FILE_MANAGER.md # S3文件管理器文档
├── .github/           # GitHub配置
│   └── workflows/     # CI/CD工作流
├── go.mod             # Go模块文件
├── go.sum             # Go依赖锁定文件
├── Makefile           # 构建脚本
├── README.md          # 项目说明
└── LICENSE            # 许可证文件
```

## 🚀 在其他项目中使用

### 1. 作为依赖引入

在你的项目中创建或更新 `go.mod` 文件：

```bash
# 初始化新项目
go mod init your-project-name

# 添加baseComponents依赖
go get github.com/QingsiLiu/baseComponents@latest
```

### 2. 使用特定组件

```go
// main.go
package main

import (
    "log"
    "github.com/QingsiLiu/baseComponents/storage/s3"
    "github.com/QingsiLiu/baseComponents/utils"
)

func main() {
    // 使用S3存储组件
    s3Service, err := s3.NewS3Service("us-east-1")
    if err != nil {
        log.Fatal("Failed to create S3 service:", err)
    }
    
    // 使用工具函数
    if utils.IsValidEmail("test@example.com") {
        log.Println("Valid email address")
    }
}
```

### 3. 项目示例结构

```
your-project/
├── go.mod              # 包含baseComponents依赖
├── go.sum              # 依赖锁定文件
├── main.go             # 主程序
├── config/             # 配置文件
│   └── config.yaml
└── handlers/           # 业务处理器
    └── file_handler.go # 使用S3组件的文件处理器
```

### 4. 完整的集成示例

```go
// handlers/file_handler.go
package handlers

import (
    "fmt"
    "github.com/QingsiLiu/baseComponents/storage"
    "github.com/QingsiLiu/baseComponents/storage/s3"
)

type FileHandler struct {
    storage storage.StorageService
}

func NewFileHandler() (*FileHandler, error) {
    s3Service, err := s3.NewS3Service("us-east-1")
    if err != nil {
        return nil, fmt.Errorf("failed to create S3 service: %w", err)
    }
    
    return &FileHandler{
        storage: s3Service,
    }, nil
}

func (h *FileHandler) UploadFile(bucket, key string, data []byte) error {
    return h.storage.UploadObject(bucket, key, data)
}

func (h *FileHandler) ListFiles(bucket, prefix string) ([]storage.ObjectInfo, error) {
    input := &storage.ListObjectsInput{
        Bucket:  bucket,
        Prefix:  prefix,
        MaxKeys: 100,
    }
    
    result, err := h.storage.ListObjects(input)
    if err != nil {
        return nil, err
    }
    
    return result.Objects, nil
}
```

## 📋 版本管理和发布策略

### 语义化版本控制

本项目遵循 [语义化版本控制](https://semver.org/lang/zh-CN/) 规范：

- **主版本号 (MAJOR)**: 不兼容的API修改
- **次版本号 (MINOR)**: 向下兼容的功能性新增
- **修订号 (PATCH)**: 向下兼容的问题修正

### 版本发布流程

1. **开发阶段**: 在 `develop` 分支进行功能开发
2. **测试阶段**: 创建 `release/vX.Y.Z` 分支进行测试
3. **发布阶段**: 合并到 `main` 分支并打标签

### 使用特定版本

```bash
# 使用最新稳定版本
go get github.com/QingsiLiu/baseComponents@latest

# 使用特定版本
go get github.com/QingsiLiu/baseComponents@v1.2.3

# 使用特定分支
go get github.com/QingsiLiu/baseComponents@develop
```

### 版本兼容性

- **v1.x.x**: 稳定版本，保证API兼容性
- **v0.x.x**: 开发版本，API可能会有变化
- **vX.Y.Z-alpha/beta/rc**: 预发布版本

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./storage/s3

# 运行测试并显示详细输出
go test -v ./...

# 运行测试并生成覆盖率报告
go test -cover ./...

# 生成HTML覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 集成测试

对于需要真实AWS环境的测试，设置环境变量：

```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
export AWS_TEST_BUCKET="your-test-bucket"

# 运行集成测试
go test ./storage/s3 -v -run TestS3ServiceWithRealAWS
```

## 📚 文档和示例

### 详细文档

- [S3文件管理器文档](docs/S3_FILE_MANAGER.md) - 完整的S3组件使用指南
- [S3文件管理器快速入门](README_S3_FILE_MANAGER.md) - S3组件快速开始指南

### 示例代码

- [S3文件管理器完整示例](examples/s3_file_manager_example.go) - 展示所有S3功能的完整示例

### API文档

使用 `go doc` 查看API文档：

```bash
# 查看包文档
go doc github.com/QingsiLiu/baseComponents/storage/s3

# 查看特定函数文档
go doc github.com/QingsiLiu/baseComponents/storage/s3.NewS3Service
```

## 🔧 配置和环境变量

### S3组件配置

```bash
# AWS凭证（必需）
export AWS_ACCESS_KEY_ID="your-access-key-id"
export AWS_SECRET_ACCESS_KEY="your-secret-access-key"

# AWS区域（可选，默认us-east-1）
export AWS_REGION="us-east-1"

# 测试用存储桶（仅测试时需要）
export AWS_TEST_BUCKET="your-test-bucket"
```

### 配置文件示例

```yaml
# config.yaml
aws:
  region: "us-east-1"
  access_key_id: "${AWS_ACCESS_KEY_ID}"
  secret_access_key: "${AWS_SECRET_ACCESS_KEY}"

storage:
  default_bucket: "my-app-storage"
  upload_timeout: "30s"
```

## 🛡️ 安全最佳实践

### 1. 凭证管理
- ✅ 使用环境变量存储敏感信息
- ✅ 在生产环境使用IAM角色
- ❌ 不要在代码中硬编码凭证

### 2. 权限控制
- ✅ 遵循最小权限原则
- ✅ 定期审查和更新权限
- ✅ 使用预签名URL限制访问时间

### 3. 数据保护
- ✅ 启用传输加密（HTTPS）
- ✅ 考虑使用服务端加密
- ✅ 定期备份重要数据

## 🐛 故障排除

### 常见问题

1. **模块导入错误**
   ```
   Error: module not found
   ```
   解决方案：确保使用正确的模块路径和版本

2. **AWS凭证错误**
   ```
   Error: NoCredentialProviders
   ```
   解决方案：设置正确的AWS环境变量

3. **版本冲突**
   ```
   Error: version conflict
   ```
   解决方案：使用 `go mod tidy` 清理依赖

### 获取帮助

- 📖 查看[详细文档](docs/)
- 🐛 提交[Issue](https://github.com/QingsiLiu/baseComponents/issues)
- 💬 参与[讨论](https://github.com/QingsiLiu/baseComponents/discussions)

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 贡献方式

1. **报告问题**: 发现bug或有改进建议
2. **提交代码**: 修复bug或添加新功能
3. **改进文档**: 完善文档和示例
4. **分享经验**: 分享使用心得和最佳实践

### 贡献流程

1. Fork 项目到你的GitHub账户
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建Pull Request

### 代码规范

- 遵循Go官方代码规范
- 添加必要的测试用例
- 更新相关文档
- 确保所有测试通过

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。

---

## 🌟 致谢

感谢所有贡献者和使用者的支持！

如果这个项目对你有帮助，请给个 ⭐️ Star！
