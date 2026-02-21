// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCORSMiddleware(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 CORS 中间件
	app.Use(CORSMiddleware())

	// 添加测试路由
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 测试用例
	tests := []struct {
		name           string
		method         string
		origin         string
		expectedStatus int
		checkHeaders   map[string]string
	}{
		{
			name:           "GET request with origin",
			method:         "GET",
			origin:         "http://localhost:3000",
			expectedStatus: 200,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:           "OPTIONS preflight request",
			method:         "OPTIONS",
			origin:         "http://localhost:3000",
			expectedStatus: 204,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Methods":     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
				"Access-Control-Allow-Headers":     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "86400",
			},
		},
		{
			name:           "POST request without origin",
			method:         "POST",
			origin:         "",
			expectedStatus: 200,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建请求
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// 执行请求
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			// 验证状态码
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// 验证响应头
			for header, expectedValue := range tt.checkHeaders {
				actualValue := resp.Header.Get(header)
				if actualValue != expectedValue {
					t.Errorf("Header %s: expected %q, got %q", header, expectedValue, actualValue)
				}
			}

			// 读取响应体（确保没有错误）
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
		})
	}
}

func TestCORSMiddleware_ExposeHeaders(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 CORS 中间件
	app.Use(CORSMiddleware())

	// 添加测试路由
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Set("Content-Length", "100")
		return c.SendString("OK")
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 验证 Expose-Headers
	exposeHeaders := resp.Header.Get("Access-Control-Expose-Headers")
	if exposeHeaders != "Content-Length,Content-Type" {
		t.Errorf("Expected Expose-Headers %q, got %q", "Content-Length,Content-Type", exposeHeaders)
	}
}
