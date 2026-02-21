// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/services"
	"github.com/oauth-server/oauth-server/types"
)

// JWTAuthMiddleware 返回一个 JWT 认证中间件
// 验证 Authorization 头中的 Bearer token，并将用户信息存储到 ctx.Locals
func JWTAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 提取 Authorization 头
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("缺少认证令牌"))
		}

		// 解析 Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("无效的认证格式"))
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证 JWT 并获取用户信息
		claims, err := services.ParseJwtToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("令牌验证失败"))
		}

		// 检查令牌是否过期（ParseJwtToken 已经验证了过期时间）
		// 这里不需要额外检查，因为 jwt.Parse 会自动验证 exp claim

		// 存储用户信息到上下文
		c.Locals("userID", claims.Id)
		c.Locals("email", claims.Email)
		c.Locals("username", claims.Username)
		c.Locals("isAdmin", claims.IsAdmin)
		c.Locals("isRealName", claims.IsRealName)
		c.Locals("owner", claims.Owner)

		// 继续处理请求
		return c.Next()
	}
}

// AdminAuthMiddleware 返回一个管理员权限验证中间件
// 检查用户是否具有管理员权限，必须在 JWTAuthMiddleware 之后使用
func AdminAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 从上下文获取 isAdmin 标志
		isAdmin, ok := c.Locals("isAdmin").(bool)
		if !ok || !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(types.ErrorResponse("需要管理员权限"))
		}

		// 继续处理请求
		return c.Next()
	}
}

// RealNameAuthMiddleware 返回一个实名认证验证中间件
// 检查用户是否已完成实名认证，必须在 JWTAuthMiddleware 之后使用
func RealNameAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 从上下文获取 isRealName 标志
		isRealName, ok := c.Locals("isRealName").(bool)
		if !ok || !isRealName {
			return c.Status(fiber.StatusForbidden).JSON(types.ErrorResponse("需要完成实名认证"))
		}

		// 继续处理请求
		return c.Next()
	}
}
