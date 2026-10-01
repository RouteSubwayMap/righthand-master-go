.PHONY: build run clean test auth memo task gateway

# 默认目标
all: build

# 安装依赖
deps:
	go mod download
	go mod tidy

# 构建所有服务
build: build-auth build-memo build-task build-gateway

build-auth:
	@echo "构建认证服务..."
	go build -o bin/auth cmd/auth/main.go

build-memo:
	@echo "构建备忘录服务..."
	go build -o bin/memo cmd/memo/main.go

build-task:
	@echo "构建任务服务..."
	go build -o bin/task cmd/task/main.go

build-gateway:
	@echo "构建API网关..."
	go build -o bin/gateway cmd/gateway/main.go

# 运行服务
run: run-gateway

run-auth:
	@echo "启动认证服务..."
	go run cmd/auth/main.go

run-memo:
	@echo "启动备忘录服务..."
	go run cmd/memo/main.go

run-task:
	@echo "启动任务服务..."
	go run cmd/task/main.go

run-gateway:
	@echo "启动API网关..."
	go run cmd/gateway/main.go

# 测试
test:
	@echo "运行测试..."
	go test -v ./...

# 清理
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	go clean

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golint ./...

# 帮助
help:
	@echo "RightHand Go 后端构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  make deps      - 安装依赖"
	@echo "  make build     - 构建所有服务"
	@echo "  make build-auth - 构建认证服务"
	@echo "  make run-auth  - 运行认证服务"
	@echo "  make test      - 运行测试"
	@echo "  make clean     - 清理构建文件"
	@echo "  make fmt       - 格式化代码"
