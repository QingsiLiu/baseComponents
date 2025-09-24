#!/bin/bash

# 版本发布脚本
# 用法: ./scripts/release.sh [版本号] [发布类型]
# 示例: ./scripts/release.sh 1.0.0 stable
#       ./scripts/release.sh 1.1.0-alpha.1 prerelease

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 检查参数
if [ $# -lt 1 ]; then
    print_message $RED "错误: 请提供版本号"
    echo "用法: $0 <版本号> [发布类型]"
    echo "示例: $0 1.0.0 stable"
    echo "      $0 1.1.0-alpha.1 prerelease"
    exit 1
fi

VERSION=$1
RELEASE_TYPE=${2:-"stable"}

# 验证版本号格式
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+(\.[0-9]+)?)?$ ]]; then
    print_message $RED "错误: 版本号格式不正确"
    echo "正确格式: X.Y.Z 或 X.Y.Z-alpha.N"
    exit 1
fi

# 检查是否在主分支
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    print_message $YELLOW "警告: 当前不在主分支 (当前分支: $CURRENT_BRANCH)"
    read -p "是否继续? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_message $RED "发布已取消"
        exit 1
    fi
fi

# 检查工作目录是否干净
if ! git diff-index --quiet HEAD --; then
    print_message $RED "错误: 工作目录不干净，请先提交所有更改"
    git status --porcelain
    exit 1
fi

# 检查是否有未推送的提交
if [ $(git rev-list HEAD...origin/$(git branch --show-current) --count) -ne 0 ]; then
    print_message $RED "错误: 有未推送的提交，请先推送到远程仓库"
    exit 1
fi

print_message $BLUE "开始发布版本 v$VERSION..."

# 运行测试
print_message $BLUE "运行测试..."
if ! go test ./...; then
    print_message $RED "错误: 测试失败"
    exit 1
fi

# 检查代码格式
print_message $BLUE "检查代码格式..."
if ! go fmt ./...; then
    print_message $RED "错误: 代码格式检查失败"
    exit 1
fi

# 代码静态检查
print_message $BLUE "运行代码静态检查..."
if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run; then
        print_message $YELLOW "警告: 静态检查发现问题"
        read -p "是否继续发布? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_message $RED "发布已取消"
            exit 1
        fi
    fi
else
    print_message $YELLOW "警告: 未安装 golangci-lint，跳过静态检查"
fi

# 构建检查
print_message $BLUE "检查构建..."
if ! go build ./...; then
    print_message $RED "错误: 构建失败"
    exit 1
fi

# 更新 CHANGELOG.md
print_message $BLUE "更新 CHANGELOG.md..."
if [ -f CHANGELOG.md ]; then
    # 创建备份
    cp CHANGELOG.md CHANGELOG.md.bak
    
    # 获取当前日期
    CURRENT_DATE=$(date +%Y-%m-%d)
    
    # 更新版本信息
    sed -i.tmp "s/## \[未发布\]/## [未发布]\n\n### 新增\n- 准备下一个版本\n\n## [$VERSION] - $CURRENT_DATE/" CHANGELOG.md
    rm CHANGELOG.md.tmp
    
    print_message $GREEN "CHANGELOG.md 已更新"
else
    print_message $YELLOW "警告: 未找到 CHANGELOG.md 文件"
fi

# 提交更改
if git diff --quiet; then
    print_message $BLUE "没有需要提交的更改"
else
    print_message $BLUE "提交版本更新..."
    git add .
    git commit -m "chore: release v$VERSION"
fi

# 创建标签
print_message $BLUE "创建标签 v$VERSION..."
if [ "$RELEASE_TYPE" = "prerelease" ]; then
    git tag -a "v$VERSION" -m "Release v$VERSION (预发布版本)"
else
    git tag -a "v$VERSION" -m "Release v$VERSION"
fi

# 推送到远程仓库
print_message $BLUE "推送到远程仓库..."
git push origin $(git branch --show-current)
git push origin "v$VERSION"

print_message $GREEN "✅ 版本 v$VERSION 发布成功!"

# 显示后续步骤
print_message $BLUE "后续步骤:"
echo "1. 检查 GitHub Actions 是否正常运行"
echo "2. 验证 GitHub Release 是否创建成功"
echo "3. 等待 Go Proxy 缓存更新 (通常需要几分钟)"
echo "4. 测试新版本: go get github.com/QingsiLiu/baseComponents@v$VERSION"

# 清理备份文件
if [ -f CHANGELOG.md.bak ]; then
    rm CHANGELOG.md.bak
fi

print_message $GREEN "发布完成! 🎉"