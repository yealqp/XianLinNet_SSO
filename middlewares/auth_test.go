// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oauth-server/oauth-server/services"
	"github.com/oauth-server/oauth-server/types"
)

// 生成测试用的 JWT token，使用与 services.Claims 相同的结构
func generateTestToken(userID string, email string, isAdmin bool, isRealName bool, expired bool) string {
	// 使用默认的 JWT secret（与 services.ParseJwtToken 中的默认值相同）
	jwtSecret := "default-secret-key-change-in-production"

	now := time.Now()
	expireTime := now.Add(1 * time.Hour)
	if expired {
		expireTime = now.Add(-1 * time.Hour) // 已过期
	}

	// 使用与 services.Claims 完全相同的结构
	claims := services.Claims{
		Owner:      "test-owner",
		Id:         userID,
		Username:   "testuser",
		Email:      email,
		IsRealName: isRealName,
		IsAdmin:    isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))
	return tokenString
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 JWT 认证中间件
	app.Use(JWTAuthMiddleware())

	// 添加测试路由
	app.Get("/protected", func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		email := c.Locals("email")
		isAdmin := c.Locals("isAdmin")
		isRealName := c.Locals("isRealName")

		return c.JSON(fiber.Map{
			"userID":     userID,
			"email":      email,
			"isAdmin":    isAdmin,
			"isRealName": isRealName,
		})
	})

	// 生成有效的 token
	token := generateTestToken("123", "test@example.com", true, true, false)

	// 创建请求
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 验证状态码
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 验证响应内容
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["userID"] != "123" {
		t.Errorf("Expected userID '123', got '%v'", result["userID"])
	}
	if result["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", result["email"])
	}
	if result["isAdmin"] != true {
		t.Errorf("Expected isAdmin true, got %v", result["isAdmin"])
	}
	if result["isRealName"] != true {
		t.Errorf("Expected isRealName true, got %v", result["isRealName"])
	}
}

func TestJWTAuthMiddleware_MissingToken(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 JWT 认证中间件
	app.Use(JWTAuthMiddleware())

	// 添加测试路由
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 创建请求（不带 Authorization 头）
	req := httptest.NewRequest("GET", "/protected", nil)

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 验证状态码为 401
	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	// 验证响应格式
	var apiResp types.ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", apiResp.Status)
	}
	if apiResp.Msg != "缺少认证令牌" {
		t.Errorf("Expected msg '缺少认证令牌', got '%s'", apiResp.Msg)
	}
}

func TestJWTAuthMiddleware_InvalidFormat(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 JWT 认证中间件
	app.Use(JWTAuthMiddleware())

	// 添加测试路由
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 创建请求（Authorization 格式错误）
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 验证状态码为 401
	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	// 验证响应格式
	var apiResp types.ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Msg != "无效的认证格式" {
		t.Errorf("Expected msg '无效的认证格式', got '%s'", apiResp.Msg)
	}
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 JWT 认证中间件
	app.Use(JWTAuthMiddleware())

	// 添加测试路由
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 创建请求（使用无效的 token）
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 验证状态码为 401
	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	// 验证响应格式
	var apiResp types.ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Msg != "令牌验证失败" {
		t.Errorf("Expected msg '令牌验证失败', got '%s'", apiResp.Msg)
	}
}

func TestAdminAuthMiddleware_AdminUser(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 JWT 认证中间件和管理员权限中间件
	app.Use(JWTAuthMiddleware())
	app.Use(AdminAuthMiddleware())

	// 添加测试路由
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendString("Admin OK")
	})

	// 生成管理员 token
	token := generateTestToken("123", "admin@example.com", true, true, false)

	// 创建请求
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

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
}

func TestAdminAuthMiddleware_NonAdminUser(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 JWT 认证中间件和管理员权限中间件
	app.Use(JWTAuthMiddleware())
	app.Use(AdminAuthMiddleware())

	// 添加测试路由
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendString("Admin OK")
	})

	// 生成非管理员 token
	token := generateTestToken("123", "user@example.com", false, true, false)

	// 创建请求
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 验证状态码为 403
	if resp.StatusCode != 403 {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}

	// 验证响应格式
	var apiResp types.ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Msg != "需要管理员权限" {
		t.Errorf("Expected msg '需要管理员权限', got '%s'", apiResp.Msg)
	}
}

func TestRealNameAuthMiddleware_RealNameUser(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 JWT 认证中间件和实名认证中间件
	app.Use(JWTAuthMiddleware())
	app.Use(RealNameAuthMiddleware())

	// 添加测试路由
	app.Get("/realname", func(c *fiber.Ctx) error {
		return c.SendString("RealName OK")
	})

	// 生成已实名认证的 token
	token := generateTestToken("123", "user@example.com", false, true, false)

	// 创建请求
	req := httptest.NewRequest("GET", "/realname", nil)
	req.Header.Set("Authorization", "Bearer "+token)

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
}

func TestRealNameAuthMiddleware_NonRealNameUser(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册 JWT 认证中间件和实名认证中间件
	app.Use(JWTAuthMiddleware())
	app.Use(RealNameAuthMiddleware())

	// 添加测试路由
	app.Get("/realname", func(c *fiber.Ctx) error {
		return c.SendString("RealName OK")
	})

	// 生成未实名认证的 token
	token := generateTestToken("123", "user@example.com", false, false, false)

	// 创建请求
	req := httptest.NewRequest("GET", "/realname", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// 执行请求
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 验证状态码为 403
	if resp.StatusCode != 403 {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}

	// 验证响应格式
	var apiResp types.ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if apiResp.Msg != "需要完成实名认证" {
		t.Errorf("Expected msg '需要完成实名认证', got '%s'", apiResp.Msg)
	}
}

func TestMiddlewareChain(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 注册中间件链：JWT -> Admin
	app.Use(JWTAuthMiddleware())
	app.Use(AdminAuthMiddleware())

	// 添加测试路由
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Admin access granted",
			"userID":  c.Locals("userID"),
		})
	})

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Valid admin token",
			token:          generateTestToken("123", "admin@example.com", true, true, false),
			expectedStatus: 200,
			expectedMsg:    "",
		},
		{
			name:           "Valid non-admin token",
			token:          generateTestToken("123", "user@example.com", false, true, false),
			expectedStatus: 403,
			expectedMsg:    "需要管理员权限",
		},
		{
			name:           "Invalid token",
			token:          "invalid.token.here",
			expectedStatus: 401,
			expectedMsg:    "令牌验证失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/admin", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.token))

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedMsg != "" {
				var apiResp types.ApiResponse
				if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if apiResp.Msg != tt.expectedMsg {
					t.Errorf("Expected msg '%s', got '%s'", tt.expectedMsg, apiResp.Msg)
				}
			}
		})
	}
}
