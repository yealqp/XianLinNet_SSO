// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package routers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/handlers"
	"github.com/oauth-server/oauth-server/middlewares"
	"github.com/oauth-server/oauth-server/types"
)

// RegisterMiddlewares 注册全局中间件
// 按照执行顺序注册：CORS -> Compress -> Logger -> Recovery
func RegisterMiddlewares(app *fiber.App) {
	// CORS 中间件 - 处理跨域请求
	// 注意：CORS 已在 nginx 层处理，这里不再重复添加以避免 CORS 头重复
	// app.Use(middlewares.CORSMiddleware())

	// Compress 中间件 - 响应压缩（对大于 1KB 的响应启用 gzip）
	app.Use(middlewares.CompressMiddleware())

	// Logger 中间件 - 记录所有请求
	app.Use(middlewares.LoggerMiddleware())

	// Recovery 中间件 - 捕获 panic
	app.Use(middlewares.RecoveryMiddleware())
}

// RegisterRoutes 注册所有路由
// 确保路由路径与 Beego 版本完全一致，保持前端兼容性
func RegisterRoutes(app *fiber.App) {
	// 健康检查端点
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(types.ApiResponse{Status: "ok"})
	})

	// OAuth 授权端点（需要 JWT 认证）
	app.Get("/oauth/authorize", middlewares.JWTAuthMiddleware(), handlers.HandleAuthorize())
	app.Post("/oauth/authorize", middlewares.JWTAuthMiddleware(), handlers.HandleAuthorize())

	// OIDC Discovery 端点
	app.Get("/.well-known/openid-configuration", handlers.HandleDiscovery())
	app.Get("/.well-known/jwks", handlers.HandleJwks())

	// API 路由组
	api := app.Group("/api")

	// ========== 认证路由（公开，无需认证） ==========
	api.Post("/auth/login", handlers.HandleLogin())
	api.Post("/auth/register", handlers.HandleRegister())
	api.Post("/auth/send-code", handlers.HandleSendVerificationCode())
	api.Post("/auth/reset-password", handlers.HandleResetPassword())
	api.Get("/auth/application-info", handlers.HandleGetApplicationInfo())

	// ========== Token 路由（公开） ==========
	api.Post("/oauth/token", handlers.HandleToken())
	api.Post("/login/oauth/access_token", handlers.HandleToken()) // 兼容别名
	api.Post("/oauth/introspect", handlers.HandleIntrospect())
	api.Post("/login/oauth/introspect", handlers.HandleIntrospect()) // 兼容别名
	api.Post("/oauth/revoke", handlers.HandleRevoke())
	api.Post("/oauth/register", handlers.HandleOidcRegister())

	// ========== 需要认证的路由 ==========
	api.Post("/auth/update-profile", middlewares.JWTAuthMiddleware(), handlers.HandleUpdateProfile())
	api.Get("/userinfo", middlewares.JWTAuthMiddleware(), handlers.HandleUserInfo())

	// ========== 管理员路由（需要 JWT 认证 + 管理员权限） ==========
	admin := api.Group("/admin", middlewares.JWTAuthMiddleware(), middlewares.AdminAuthMiddleware())

	// 用户管理
	admin.Get("/users", handlers.HandleGetUsers())
	admin.Post("/users", handlers.HandleCreateUser())
	admin.Get("/users/:id", handlers.HandleGetUser())
	admin.Post("/users/:id/update", handlers.HandleUpdateUser())
	admin.Post("/users/:id/delete", handlers.HandleDeleteUser())

	// 应用管理
	admin.Get("/applications", handlers.HandleGetApplications())
	admin.Post("/applications", handlers.HandleCreateApplication())
	admin.Get("/applications/:owner/:name", handlers.HandleGetApplication())
	admin.Post("/applications/:owner/:name/update", handlers.HandleUpdateApplication())
	admin.Post("/applications/:owner/:name/delete", handlers.HandleDeleteApplication())

	// Token 管理
	admin.Get("/tokens", handlers.HandleGetTokens())
	admin.Post("/tokens/:owner/:name/revoke", handlers.HandleRevokeToken())
	admin.Post("/tokens/user/:owner/:username/revoke", handlers.HandleRevokeUserTokens())

	// 系统管理
	admin.Get("/stats", handlers.HandleGetStats())
	admin.Get("/system", handlers.HandleGetSystemInfo())
	admin.Post("/cache/clear", handlers.HandleClearCache())

	// ========== 实名认证路由 ==========
	// 提交实名认证（需要 JWT 认证）
	api.Post("/realname/submit", middlewares.JWTAuthMiddleware(), handlers.HandleSubmitRealName())
	api.Get("/realname/verify", middlewares.JWTAuthMiddleware(), handlers.HandleGetRealNameInfo())

	// 管理员验证实名信息（需要管理员权限）
	admin.Post("/realname/verify", handlers.HandleVerifyRealName())
	admin.Get("/realname/:userId", handlers.HandleAdminGetRealNameInfo())
}
