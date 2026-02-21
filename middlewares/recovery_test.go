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
	"github.com/oauth-server/oauth-server/types"
)

func TestRecoveryMiddleware_CatchesPanic(t *testing.T) {
	// 保存原始 stderr
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()

	// 创建管道捕获错误日志输出
	r, w, _ := os.Pipe()
	os.Stderr = w

	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 Recovery 中间件
	app.Use(RecoveryMiddleware())

	// 添加会触发 panic 的路由
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/panic", nil)

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 关闭写入端并读取错误日志输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// 验证状态码为 500
	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	// 验证响应格式
	var apiResp types.ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", apiResp.Status)
	}

	if apiResp.Msg != "服务器内部错误" {
		t.Errorf("Expected msg '服务器内部错误', got '%s'", apiResp.Msg)
	}

	// 验证错误日志被记录
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected error log output, got empty string")
	}

	// 解析 JSON 错误日志
	var logEntry RecoveryLogEntry
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Failed to parse error log JSON: %v", err)
	}

	// 验证日志字段
	if logEntry.Level != "error" {
		t.Errorf("Expected level 'error', got '%s'", logEntry.Level)
	}
	if logEntry.Message != "Panic recovered" {
		t.Errorf("Expected message 'Panic recovered', got '%s'", logEntry.Message)
	}
	if logEntry.Error != "test panic" {
		t.Errorf("Expected error 'test panic', got '%s'", logEntry.Error)
	}
	if logEntry.StackTrace == "" {
		t.Error("Expected stack trace to be recorded")
	}
}

func TestRecoveryMiddleware_NormalRequest(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 Recovery 中间件
	app.Use(RecoveryMiddleware())

	// 添加正常路由
	app.Get("/normal", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/normal", nil)

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 验证状态码为 200
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", string(body))
	}
}

func TestRecoveryMiddleware_ServerContinuesRunning(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 Recovery 中间件
	app.Use(RecoveryMiddleware())

	// 添加会触发 panic 的路由
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	// 添加正常路由
	app.Get("/normal", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 第一个请求触发 panic
	req1 := httptest.NewRequest("GET", "/panic", nil)
	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("Failed to execute first request: %v", err)
	}
	resp1.Body.Close()

	// 验证第一个请求返回 500
	if resp1.StatusCode != 500 {
		t.Errorf("Expected status 500 for panic request, got %d", resp1.StatusCode)
	}

	// 第二个请求应该正常工作（验证服务器继续运行）
	req2 := httptest.NewRequest("GET", "/normal", nil)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("Failed to execute second request: %v", err)
	}
	defer resp2.Body.Close()

	// 验证第二个请求返回 200
	if resp2.StatusCode != 200 {
		t.Errorf("Expected status 200 for normal request, got %d", resp2.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", string(body))
	}
}
