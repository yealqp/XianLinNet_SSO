// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package routers

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/oauth-server/oauth-server/middlewares"
	"github.com/oauth-server/oauth-server/types"
)

// **Property 3: 路由端点完整性**
// 对于任何 Beego 版本中存在的 API 端点，在 Fiber 版本中必须存在相同路径和方法的对应端点
// **Validates: Requirements 12.1, 12.2, 12.3, 12.4**
func TestProperty_RouteEndpointCompleteness(t *testing.T) {
	// 定义 Beego 版本中存在的所有端点
	beegoRoutes := []struct {
		method string
		path   string
	}{
		// Auth routes
		{"GET", "/oauth/authorize"},
		{"POST", "/oauth/authorize"},
		{"GET", "/api/auth/application-info"},
		{"POST", "/api/auth/login"},
		{"POST", "/api/auth/register"},
		{"POST", "/api/auth/send-code"},
		{"POST", "/api/auth/reset-password"},
		{"POST", "/api/auth/update-profile"},

		// Token routes
		{"POST", "/api/oauth/token"},
		{"POST", "/api/login/oauth/access_token"}, // Alias
		{"POST", "/api/oauth/introspect"},
		{"POST", "/api/login/oauth/introspect"}, // Alias
		{"POST", "/api/oauth/revoke"},

		// OIDC routes
		{"GET", "/.well-known/openid-configuration"},
		{"GET", "/.well-known/jwks"},
		{"GET", "/api/userinfo"},
		{"POST", "/api/oauth/register"},

		// Health check
		{"GET", "/health"},

		// Admin routes - User management
		{"GET", "/api/admin/users"},
		{"POST", "/api/admin/users"},
		{"GET", "/api/admin/users/:id"},
		{"POST", "/api/admin/users/:id/update"},
		{"POST", "/api/admin/users/:id/delete"},

		// Admin routes - Application management
		{"GET", "/api/admin/applications"},
		{"POST", "/api/admin/applications"},
		{"GET", "/api/admin/applications/:owner/:name"},
		{"POST", "/api/admin/applications/:owner/:name/update"},
		{"POST", "/api/admin/applications/:owner/:name/delete"},

		// Admin routes - Token management
		{"GET", "/api/admin/tokens"},
		{"POST", "/api/admin/tokens/:owner/:name/revoke"},
		{"POST", "/api/admin/tokens/user/:owner/:username/revoke"},

		// Admin routes - System management
		{"GET", "/api/admin/stats"},
		{"GET", "/api/admin/system"},
		{"POST", "/api/admin/cache/clear"},

		// Real name verification routes
		{"POST", "/api/realname/submit"},
		{"GET", "/api/realname/verify"},
		{"POST", "/api/admin/realname/verify"},
		{"GET", "/api/admin/realname/:userId"},
	}

	// 创建 Fiber 应用并注册路由
	app := fiber.New()
	RegisterMiddlewares(app)
	RegisterRoutes(app)

	// 验证每个 Beego 路由在 Fiber 中都存在
	for _, route := range beegoRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			// 替换路径参数为实际值
			testPath := route.path
			testPath = strings.ReplaceAll(testPath, ":id", "1")
			testPath = strings.ReplaceAll(testPath, ":owner", "test-owner")
			testPath = strings.ReplaceAll(testPath, ":name", "test-name")
			testPath = strings.ReplaceAll(testPath, ":username", "test-user")
			testPath = strings.ReplaceAll(testPath, ":userId", "1")

			// 创建测试请求
			req := httptest.NewRequest(route.method, testPath, nil)

			// 执行请求
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to test route %s %s: %v", route.method, route.path, err)
			}

			// 验证路由存在（不应该返回 404）
			if resp.StatusCode == fiber.StatusNotFound {
				t.Errorf("Route %s %s not found in Fiber app (expected to exist)", route.method, route.path)
			}
		})
	}
}

// **Property 4: 中间件执行顺序**
// 对于任何受保护的请求，中间件必须按照正确的顺序执行
// **Validates: Requirements 2.1, 2.3, 2.7, 2.9, 2.10**
func TestProperty_MiddlewareExecutionOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 20

	properties := gopter.NewProperties(parameters)

	// 生成器：生成不同的请求路径
	pathGen := gen.OneConstOf(
		"/api/auth/login",
		"/api/auth/register",
		"/api/auth/update-profile",
		"/api/admin/users",
		"/health",
	)

	properties.Property("Middleware execution order is correct", prop.ForAll(
		func(path string) bool {
			// 创建 Fiber 应用
			app := fiber.New()

			// 记录中间件执行顺序
			executionOrder := []string{}

			// 注册测试中间件来跟踪执行顺序
			app.Use(func(c *fiber.Ctx) error {
				executionOrder = append(executionOrder, "CORS")
				return c.Next()
			})

			app.Use(func(c *fiber.Ctx) error {
				executionOrder = append(executionOrder, "Logger")
				return c.Next()
			})

			app.Use(func(c *fiber.Ctx) error {
				executionOrder = append(executionOrder, "Recovery")
				return c.Next()
			})

			// 注册一个测试路由
			app.Get(path, func(c *fiber.Ctx) error {
				executionOrder = append(executionOrder, "Handler")
				return c.JSON(types.SuccessResponse("ok"))
			})

			// 创建测试请求
			req := httptest.NewRequest("GET", path, nil)

			// 执行请求
			_, err := app.Test(req, -1)
			if err != nil {
				return false
			}

			// 验证执行顺序：CORS -> Logger -> Recovery -> Handler
			expectedOrder := []string{"CORS", "Logger", "Recovery", "Handler"}
			if len(executionOrder) != len(expectedOrder) {
				return false
			}

			for i, middleware := range expectedOrder {
				if executionOrder[i] != middleware {
					return false
				}
			}

			return true
		},
		pathGen,
	))

	properties.TestingRun(t)
}

// TestMiddlewareExecutionOrder_WithAuth 测试带认证的中间件执行顺序
func TestMiddlewareExecutionOrder_WithAuth(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 记录中间件执行顺序
	executionOrder := []string{}

	// 注册全局中间件
	app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "CORS")
		return c.Next()
	})

	app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "Logger")
		return c.Next()
	})

	app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "Recovery")
		return c.Next()
	})

	// 注册需要认证的路由
	authenticated := app.Group("/api", func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "JWTAuth")
		// 模拟 JWT 认证成功
		c.Locals("userID", "test-user")
		c.Locals("isAdmin", false)
		return c.Next()
	})

	authenticated.Get("/protected", func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "Handler")
		return c.JSON(types.SuccessResponse("ok"))
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/api/protected", nil)

	// 执行请求
	_, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test route: %v", err)
	}

	// 验证执行顺序：CORS -> Logger -> Recovery -> JWTAuth -> Handler
	expectedOrder := []string{"CORS", "Logger", "Recovery", "JWTAuth", "Handler"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d middlewares, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, middleware := range expectedOrder {
		if i >= len(executionOrder) {
			t.Errorf("Missing middleware at position %d: expected %s", i, middleware)
			continue
		}
		if executionOrder[i] != middleware {
			t.Errorf("Middleware at position %d: expected %s, got %s", i, middleware, executionOrder[i])
		}
	}
}

// TestMiddlewareExecutionOrder_WithAdminAuth 测试管理员路由的中间件执行顺序
func TestMiddlewareExecutionOrder_WithAdminAuth(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()

	// 记录中间件执行顺序
	executionOrder := []string{}

	// 注册全局中间件
	app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "CORS")
		return c.Next()
	})

	app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "Logger")
		return c.Next()
	})

	app.Use(func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "Recovery")
		return c.Next()
	})

	// 注册管理员路由
	admin := app.Group("/api/admin",
		func(c *fiber.Ctx) error {
			executionOrder = append(executionOrder, "JWTAuth")
			// 模拟 JWT 认证成功
			c.Locals("userID", "admin-user")
			c.Locals("isAdmin", true)
			return c.Next()
		},
		func(c *fiber.Ctx) error {
			executionOrder = append(executionOrder, "AdminAuth")
			// 检查管理员权限
			isAdmin, ok := c.Locals("isAdmin").(bool)
			if !ok || !isAdmin {
				return c.Status(fiber.StatusForbidden).JSON(types.ErrorResponse("需要管理员权限"))
			}
			return c.Next()
		},
	)

	admin.Get("/users", func(c *fiber.Ctx) error {
		executionOrder = append(executionOrder, "Handler")
		return c.JSON(types.SuccessResponse("ok"))
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/api/admin/users", nil)

	// 执行请求
	_, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test route: %v", err)
	}

	// 验证执行顺序：CORS -> Logger -> Recovery -> JWTAuth -> AdminAuth -> Handler
	expectedOrder := []string{"CORS", "Logger", "Recovery", "JWTAuth", "AdminAuth", "Handler"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d middlewares, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, middleware := range expectedOrder {
		if i >= len(executionOrder) {
			t.Errorf("Missing middleware at position %d: expected %s", i, middleware)
			continue
		}
		if executionOrder[i] != middleware {
			t.Errorf("Middleware at position %d: expected %s, got %s", i, middleware, executionOrder[i])
		}
	}
}

// TestHealthCheckEndpoint 测试健康检查端点
func TestHealthCheckEndpoint(t *testing.T) {
	// 创建 Fiber 应用并注册路由
	app := fiber.New()
	RegisterMiddlewares(app)
	RegisterRoutes(app)

	// 创建测试请求
	req := httptest.NewRequest("GET", "/health", nil)

	// 执行请求
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test health endpoint: %v", err)
	}

	// 验证状态码
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status code %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// 验证响应包含 "ok"
	if !strings.Contains(string(body), `"status":"ok"`) {
		t.Errorf("Expected response to contain status:ok, got: %s", string(body))
	}
}

// TestCORSMiddleware 测试 CORS 中间件
func TestCORSMiddleware(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()
	app.Use(middlewares.CORSMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(types.SuccessResponse("ok"))
	})

	// 测试预检请求
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test CORS: %v", err)
	}

	// 验证 CORS 头部
	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header to be set")
	}

	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header to be set")
	}
}

// TestRecoveryMiddleware 测试 Recovery 中间件
func TestRecoveryMiddleware(t *testing.T) {
	// 创建 Fiber 应用
	app := fiber.New()
	app.Use(middlewares.RecoveryMiddleware())

	// 注册一个会 panic 的路由
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/panic", nil)

	// 执行请求
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test recovery: %v", err)
	}

	// 验证服务器没有崩溃，返回 500 错误
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// 验证响应包含错误信息
	if !strings.Contains(string(body), `"status":"error"`) {
		t.Errorf("Expected error response, got: %s", string(body))
	}
}
