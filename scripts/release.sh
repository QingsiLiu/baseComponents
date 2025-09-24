#!/bin/bash

# 简化的版本发布脚本
# 用法: ./scripts/release.sh [版本号]
# 示例: ./scripts/release.sh v0.2.0

set -e

VERSION=${1:-""}

if [ -z "$VERSION" ]; then
    echo "请提供版本号，例如: ./scripts/release.sh v0.2.0"
    exit 1
fi

echo "🚀 开始发布版本 $VERSION..."

# 运行测试
echo "📋 运行测试..."
go test ./...

# 创建标签并推送
echo "🏷️  创建标签 $VERSION..."
git tag $VERSION
git push origin $VERSION

echo "✅ 版本 $VERSION 发布完成!"
echo "🔗 查看发布: https://github.com/QingsiLiu/baseComponents/releases"