// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/routers"
	"github.com/oauth-server/oauth-server/services"
	"github.com/oauth-server/oauth-server/types"
)

func main() {
	// 步骤 0: 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	// 步骤 1: 初始化数据库
	err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// 自动同步表结构
	err = models.InitTables()
	if err != nil {
		log.Printf("Warning: Failed to sync tables: %v", err)
	}

	// 自动初始化数据（创建默认管理员等）
	err = models.InitData()
	if err != nil {
		log.Printf("Warning: Failed to initialize data: %v", err)
	}

	// 步骤 2: 初始化 Redis（可选）
	err = services.InitRedis()
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
		log.Println("Continuing without Redis cache...")
	} else {
		log.Println("Redis cache initialized successfully")
	}

	// 步骤 3: 初始化 RSA 密钥
	err = services.InitRSAKeys()
	if err != nil {
		log.Fatalf("Failed to initialize RSA keys: %v", err)
	}
	log.Println("RSA keys initialized successfully")

	// 检查是否为初始化命令
	if len(os.Args) > 1 && os.Args[1] == "init" {
		log.Println("Database initialization completed!")
		return
	}

	// 步骤 4: 创建 Fiber 应用实例
	app := fiber.New(fiber.Config{
		// 服务器配置
		ReadTimeout:  getEnvDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getEnvDuration("WRITE_TIMEOUT", 10*time.Second),
		BodyLimit:    getEnvInt("BODY_LIMIT", 4*1024*1024), // 4MB

		// 自定义错误处理器
		ErrorHandler: customErrorHandler,

		// 禁用启动消息（我们会自定义）
		DisableStartupMessage: true,

		// 应用名称
		AppName: "OAuth Server (Fiber v2)",
	})

	// 步骤 5: 注册中间件
	routers.RegisterMiddlewares(app)
	log.Println("Middlewares registered successfully")

	// 步骤 6: 注册路由
	routers.RegisterRoutes(app)
	log.Println("Routes registered successfully")

	// 步骤 7: 启动服务器
	port := getEnv("PORT", "8080")
	log.Printf("OAuth Server starting on port %s...", port)
	log.Printf("Server URL: http://localhost:%s", port)
	log.Printf("Health check: http://localhost:%s/health", port)

	// 步骤 8: 实现优雅关闭
	// 创建一个 channel 来监听系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// 在 goroutine 中启动服务器
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	<-quit
	log.Println("Shutting down server...")

	// 优雅关闭服务器（等待现有连接完成）
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// customErrorHandler 自定义错误处理器
// 捕获所有未处理的错误，返回统一的错误响应格式
// Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7, 11.8
func customErrorHandler(ctx *fiber.Ctx, err error) error {
	// 默认状态码为 500
	code := fiber.StatusInternalServerError

	// 如果是 Fiber 错误，使用其状态码
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// 根据状态码返回适当的错误消息
	var message string
	switch code {
	case fiber.StatusBadRequest:
		message = "请求数据无效"
	case fiber.StatusUnauthorized:
		message = "认证失败"
	case fiber.StatusForbidden:
		message = "权限不足"
	case fiber.StatusNotFound:
		message = "资源不存在"
	case fiber.StatusRequestEntityTooLarge:
		message = "请求数据过大"
	case fiber.StatusInternalServerError:
		message = "服务器内部错误"
	default:
		message = "请求处理失败"
	}

	// 记录错误日志（不泄露系统内部信息）
	log.Printf("Error [%d]: %s - %v", code, ctx.Path(), err)

	// 返回统一的错误响应格式
	return ctx.Status(code).JSON(types.ErrorResponse(message))
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数类型的环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDuration 获取时间间隔类型的环境变量
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
