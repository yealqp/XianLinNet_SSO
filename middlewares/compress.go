// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

// CompressMiddleware 实现响应压缩
// 对大于 1KB 的响应启用 gzip 压缩
func CompressMiddleware() fiber.Handler {
	return compress.New(compress.Config{
		Level: compress.LevelDefault, // 默认压缩级别（平衡速度和压缩率）
		// 只压缩大于 1KB 的响应
		Next: func(c *fiber.Ctx) bool {
			// 如果响应体小于 1KB，跳过压缩
			return c.Response().Header.ContentLength() < 1024
		},
	})
}
