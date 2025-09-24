# Makefile for baseComponents
# Go基础组件库构建和开发工具

# 变量定义
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOMOD := $(GO) mod
GOVET := $(GO) vet

# 项目信息
PROJECT_NAME := baseComponents
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)

# 目录定义
BUILD_DIR := build
COVERAGE_DIR := coverage
DOCS_DIR := docs

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "$(BLUE)$(PROJECT_NAME) - Go基础组件库$(NC)"
	@echo "$(BLUE)版本: $(VERSION) ($(COMMIT_HASH))$(NC)"
	@echo ""
	@echo "$(YELLOW)可用命令:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发环境设置
.PHONY: setup
setup: ## 设置开发环境
	@echo "$(BLUE)设置开发环境...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)开发环境设置完成$(NC)"

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "$(BLUE)下载依赖...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) verify

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "$(BLUE)更新依赖...$(NC)"
	@$(GOMOD) get -u ./...
	@$(GOMOD) tidy

.PHONY: deps-clean
deps-clean: ## 清理未使用的依赖
	@echo "$(BLUE)清理依赖...$(NC)"
	@$(GOMOD) tidy

# 代码格式化
.PHONY: fmt
fmt: ## 格式化代码
	@echo "$(BLUE)格式化代码...$(NC)"
	@$(GOFMT) -s -w .
	@$(GO) mod tidy

.PHONY: fmt-check
fmt-check: ## 检查代码格式
	@echo "$(BLUE)检查代码格式...$(NC)"
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "$(RED)以下文件需要格式化:$(NC)"; \
		$(GOFMT) -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)代码格式检查通过$(NC)"

# 代码检查
.PHONY: vet
vet: ## 运行go vet
	@echo "$(BLUE)运行 go vet...$(NC)"
	@$(GOVET) ./...

.PHONY: lint
lint: ## 运行代码检查
	@echo "$(BLUE)运行代码检查...$(NC)"
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
	else \
		echo "$(YELLOW)golangci-lint 未安装，跳过代码检查$(NC)"; \
		echo "$(YELLOW)安装命令: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2$(NC)"; \
	fi

.PHONY: lint-install
lint-install: ## 安装golangci-lint
	@echo "$(BLUE)安装 golangci-lint...$(NC)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2

# 测试
.PHONY: test
test: ## 运行所有测试
	@echo "$(BLUE)运行测试...$(NC)"
	@$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## 运行短测试（跳过长时间运行的测试）
	@echo "$(BLUE)运行短测试...$(NC)"
	@$(GOTEST) -short -v ./...

.PHONY: test-race
test-race: ## 运行竞态检测测试
	@echo "$(BLUE)运行竞态检测测试...$(NC)"
	@$(GOTEST) -race -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(BLUE)运行测试并生成覆盖率报告...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "$(GREEN)覆盖率报告已生成: $(COVERAGE_DIR)/coverage.html$(NC)"

.PHONY: test-integration
test-integration: ## 运行集成测试
	@echo "$(BLUE)运行集成测试...$(NC)"
	@echo "$(YELLOW)请确保设置了必要的环境变量:$(NC)"
	@echo "  AWS_ACCESS_KEY_ID"
	@echo "  AWS_SECRET_ACCESS_KEY"
	@echo "  AWS_REGION"
	@echo "  AWS_TEST_BUCKET"
	@$(GOTEST) -v ./storage/s3 -run TestS3ServiceWithRealAWS

# 构建
.PHONY: build
build: ## 构建项目
	@echo "$(BLUE)构建项目...$(NC)"
	@$(GOBUILD) ./...

.PHONY: build-examples
build-examples: ## 构建示例程序
	@echo "$(BLUE)构建示例程序...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for example in examples/*.go; do \
		if [ -f "$$example" ]; then \
			name=$$(basename "$$example" .go); \
			echo "构建示例: $$name"; \
			$(GOBUILD) -o $(BUILD_DIR)/$$name $$example; \
		fi \
	done

# 清理
.PHONY: clean
clean: ## 清理构建文件
	@echo "$(BLUE)清理构建文件...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@$(GO) clean -cache
	@$(GO) clean -testcache

# 文档
.PHONY: docs
docs: ## 生成文档
	@echo "$(BLUE)生成文档...$(NC)"
	@mkdir -p $(DOCS_DIR)
	@$(GO) doc -all ./... > $(DOCS_DIR)/api.txt
	@echo "$(GREEN)API文档已生成: $(DOCS_DIR)/api.txt$(NC)"

.PHONY: docs-serve
docs-serve: ## 启动文档服务器
	@echo "$(BLUE)启动文档服务器...$(NC)"
	@if command -v godoc >/dev/null 2>&1; then \
		echo "$(GREEN)文档服务器启动在 http://localhost:6060$(NC)"; \
		godoc -http=:6060; \
	else \
		echo "$(YELLOW)godoc 未安装$(NC)"; \
		echo "$(YELLOW)安装命令: go install golang.org/x/tools/cmd/godoc@latest$(NC)"; \
	fi

# 版本管理
.PHONY: version
version: ## 显示版本信息
	@echo "$(BLUE)版本信息:$(NC)"
	@echo "  项目: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  提交: $(COMMIT_HASH)"
	@echo "  构建时间: $(BUILD_TIME)"

.PHONY: version-next
version-next: ## 计算下一个版本号
	@./scripts/version.sh next

.PHONY: version-current
version-current: ## 显示当前版本
	@./scripts/version.sh current

# 发布
.PHONY: release-check
release-check: fmt-check vet lint test ## 发布前检查
	@echo "$(GREEN)发布前检查通过$(NC)"

.PHONY: release-patch
release-patch: release-check ## 发布补丁版本
	@./scripts/release.sh $$(./scripts/version.sh next patch | grep "下一个版本" | awk '{print $$3}' | sed 's/[()]//g') stable

.PHONY: release-minor
release-minor: release-check ## 发布次版本
	@./scripts/release.sh $$(./scripts/version.sh next minor | grep "下一个版本" | awk '{print $$3}' | sed 's/[()]//g') stable

.PHONY: release-major
release-major: release-check ## 发布主版本
	@./scripts/release.sh $$(./scripts/version.sh next major | grep "下一个版本" | awk '{print $$3}' | sed 's/[()]//g') stable

.PHONY: release-alpha
release-alpha: release-check ## 发布Alpha版本
	@./scripts/release.sh $$(./scripts/version.sh next alpha | grep "下一个版本" | awk '{print $$3}' | sed 's/[()]//g') prerelease

.PHONY: release-beta
release-beta: release-check ## 发布Beta版本
	@./scripts/release.sh $$(./scripts/version.sh next beta | grep "下一个版本" | awk '{print $$3}' | sed 's/[()]//g') prerelease

# 开发工具
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "$(BLUE)安装开发工具...$(NC)"
	@go install golang.org/x/tools/cmd/godoc@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
	@echo "$(GREEN)开发工具安装完成$(NC)"

.PHONY: check-tools
check-tools: ## 检查开发工具
	@echo "$(BLUE)检查开发工具...$(NC)"
	@echo -n "Go: "; $(GO) version
	@echo -n "golangci-lint: "; if command -v $(GOLINT) >/dev/null 2>&1; then $(GOLINT) version; else echo "$(RED)未安装$(NC)"; fi
	@echo -n "godoc: "; if command -v godoc >/dev/null 2>&1; then echo "$(GREEN)已安装$(NC)"; else echo "$(RED)未安装$(NC)"; fi

# 完整检查
.PHONY: check-all
check-all: fmt-check vet lint test ## 运行所有检查
	@echo "$(GREEN)所有检查通过$(NC)"

# 快速开发循环
.PHONY: dev
dev: fmt vet test ## 开发模式（格式化、检查、测试）
	@echo "$(GREEN)开发检查完成$(NC)"

# CI/CD 相关
.PHONY: ci
ci: deps fmt-check vet lint test build ## CI流水线
	@echo "$(GREEN)CI流水线完成$(NC)"

.PHONY: pre-commit
pre-commit: fmt vet lint test ## 提交前检查
	@echo "$(GREEN)提交前检查完成$(NC)"