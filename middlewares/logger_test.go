// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestLoggerMiddleware(t *testing.T) {
	// 保存原始 stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// 创建管道捕获日志输出
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 Logger 中间件
	app.Use(LoggerMiddleware())

	// 添加测试路由
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Test-Agent")

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 关闭写入端并读取日志输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// 验证状态码
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 验证日志输出
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected log output, got empty string")
	}

	// 解析 JSON 日志
	var logEntry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Failed to parse log JSON: %v", err)
	}

	// 验证日志字段
	if logEntry.Method != "GET" {
		t.Errorf("Expected method GET, got %s", logEntry.Method)
	}
	if logEntry.Path != "/test" {
		t.Errorf("Expected path /test, got %s", logEntry.Path)
	}
	if logEntry.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", logEntry.StatusCode)
	}
	if logEntry.Latency == "" {
		t.Error("Expected latency to be recorded")
	}
}

func TestLoggerMiddleware_WithError(t *testing.T) {
	// 保存原始 stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// 创建管道捕获日志输出
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 Logger 中间件
	app.Use(LoggerMiddleware())

	// 添加测试路由（返回错误）
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "test error")
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/error", nil)

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 关闭写入端并读取日志输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// 验证状态码
	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// 解析 JSON 日志
	var logEntry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Failed to parse log JSON: %v", err)
	}

	// 验证错误被记录
	if logEntry.Error == "" {
		t.Error("Expected error to be logged")
	}
}

func TestLoggerMiddleware_LogLevels(t *testing.T) {
	tests := []struct {
		name       string
		logLevel   string
		statusCode int
		shouldLog  bool
	}{
		{"debug level logs everything", "debug", 200, true},
		{"info level logs everything", "info", 200, true},
		{"warn level logs 4xx", "warn", 400, true},
		{"warn level skips 2xx", "warn", 200, false},
		{"error level logs 5xx", "error", 500, true},
		{"error level skips 4xx", "error", 400, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置日志级别
			os.Setenv("LOG_LEVEL", tt.logLevel)
			defer os.Unsetenv("LOG_LEVEL")

			// 保存原始 stdout
			oldStdout := os.Stdout
			defer func() { os.Stdout = oldStdout }()

			// 创建管道捕获日志输出
			r, w, _ := os.Pipe()
			os.Stdout = w

			// 创建 Fiber 应用
			app := fiber.New()
			app.Use(LoggerMiddleware())

			// 添加测试路由
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.Status(tt.statusCode).SendString("OK")
			})

			// 执行请求
			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			resp.Body.Close()

			// 关闭写入端并读取日志输出
			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// 验证是否记录日志
			hasLog := buf.Len() > 0
			if hasLog != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, got hasLog=%v", tt.shouldLog, hasLog)
			}
		})
	}
}
