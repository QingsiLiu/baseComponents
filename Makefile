# 简化的 Makefile for baseComponents

.PHONY: help test build fmt clean release

# 默认目标
.DEFAULT_GOAL := help

help: ## 显示帮助信息
	@echo "baseComponents - Go基础组件库"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## 运行测试
	@echo "运行测试..."
	@go test ./...

build: ## 构建项目
	@echo "构建项目..."
	@go build ./...

fmt: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...

clean: ## 清理构建文件
	@echo "清理构建文件..."
	@go clean ./...

release: ## 发布新版本 (用法: make release VERSION=v0.2.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "请指定版本号: make release VERSION=v0.2.0"; \
		exit 1; \
	fi
	@./scripts/release.sh $(VERSION)