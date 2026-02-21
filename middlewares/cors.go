// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware returns a Fiber CORS middleware handler
// It handles Cross-Origin Resource Sharing (CORS) for the Vue 3 frontend
func CORSMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		// AllowOrigins: 允许的源，生产环境应该限制为具体的前端域名
		AllowOrigins: "*",

		// AllowMethods: 允许的 HTTP 方法
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,PATCH",

		// AllowHeaders: 允许的请求头
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Requested-With",

		// AllowCredentials: 允许携带凭证（cookies, authorization headers）
		AllowCredentials: true,

		// ExposeHeaders: 暴露给前端的响应头
		ExposeHeaders: "Content-Length,Content-Type",

		// MaxAge: 预检请求的缓存时间（秒）
		MaxAge: 86400, // 24 hours
	})
}
