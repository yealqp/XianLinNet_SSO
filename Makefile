.PHONY: help build run init test clean docker docker-run docker-stop install dev prod

# 默认目标
help:
	@echo "OAuth Server - Makefile Commands"
	@echo ""
	@echo "Development:"
	@echo "  make install    - Install dependencies"
	@echo "  make init       - Initialize database"
	@echo "  make dev        - Run in development mode"
	@echo "  make run        - Run the server"
	@echo "  make test       - Run tests"
	@echo ""
	@echo "Build:"
	@echo "  make build      - Build the binary"
	@echo "  make clean      - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker     - Build Docker image"
	@echo "  make docker-run - Run with Docker Compose"
	@echo "  make docker-stop- Stop Docker Compose"
	@echo ""
	@echo "Production:"
	@echo "  make prod       - Build for production"
	@echo ""

# 安装依赖
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed!"

# 初始化数据库
init:
	@echo "Initializing database..."
	@if [ ! -f conf/app.conf ]; then \
		echo "Creating conf/app.conf from example..."; \
		cp conf/app.conf.example conf/app.conf; \
	fi
	go run main.go init
	@echo "Database initialized!"

# 开发模式运行
dev:
	@echo "Starting in development mode..."
	@if [ ! -f conf/app.conf ]; then \
		echo "Creating conf/app.conf from example..."; \
		cp conf/app.conf.example conf/app.conf; \
	fi
	go run main.go

# 运行服务器
run: build
	@echo "Starting OAuth Server..."
ifeq ($(OS),Windows_NT)
	oauth-server.exe
else
	./oauth-server
endif

# 构建
build:
	@echo "Building OAuth Server..."
ifeq ($(OS),Windows_NT)
	go build -o oauth-server.exe main.go
	@echo "Build complete: oauth-server.exe"
else
	go build -o oauth-server main.go
	@echo "Build complete: ./oauth-server"
endif

# 生产环境构建
prod:
	@echo "Building for production..."
	CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o oauth-server main.go
	@echo "Production build complete!"

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 运行测试脚本
test-api:
	@echo "Running API tests..."
	@chmod +x test_oauth.sh
	./test_oauth.sh

# 清理
clean:
	@echo "Cleaning..."
ifeq ($(OS),Windows_NT)
	@if exist oauth-server.exe del /f oauth-server.exe
	@if exist *.db del /f *.db
	@if exist *.log del /f *.log
else
	rm -f oauth-server
	rm -f *.db
	rm -f *.log
endif
	@echo "Clean complete!"

# Docker构建
docker:
	@echo "Building Docker image..."
	docker build -t oauth-server:latest .
	@echo "Docker image built!"

# Docker Compose运行
docker-run:
	@echo "Starting with Docker Compose..."
	docker-compose up -d
	@echo "Waiting for services to start..."
	sleep 5
	docker-compose exec oauth-server ./oauth-server init
	@echo "OAuth Server is running!"
	@echo "Access at: http://localhost:8080"

# Docker Compose停止
docker-stop:
	@echo "Stopping Docker Compose..."
	docker-compose down
	@echo "Stopped!"

# 查看Docker日志
docker-logs:
	docker-compose logs -f oauth-server

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted!"

# 代码检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 生成文档
docs:
	@echo "Generating documentation..."
	@if command -v godoc > /dev/null; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with:"; \
		echo "  go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# 数据库迁移
migrate:
	@echo "Running database migrations..."
	go run main.go init
	@echo "Migrations complete!"

# 备份数据库
backup:
	@echo "Backing up database..."
	@mkdir -p backups
	@if [ -f oauth_server.db ]; then \
		cp oauth_server.db backups/oauth_server_$(shell date +%Y%m%d_%H%M%S).db; \
		echo "SQLite backup created!"; \
	else \
		echo "No SQLite database found. For MySQL/PostgreSQL, use your database backup tools."; \
	fi

# 查看版本
version:
	@echo "OAuth Server"
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

# 健康检查
health:
	@echo "Checking server health..."
	@curl -s http://localhost:8080/health | jq . || echo "Server not responding"

# 快速启动（安装+初始化+运行）
quickstart: install init dev

# 完整测试（构建+测试+API测试）
test-all: build test test-api

# 部署准备
deploy-prep: clean prod
	@echo "Deployment package ready!"
	@echo "Binary: ./oauth-server"
	@echo "Config: conf/app.conf.example"
