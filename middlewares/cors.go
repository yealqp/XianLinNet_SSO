// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"github.com/beego/beego/v2/server/web/context"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS)
func CORSMiddleware(ctx *context.Context) {
	// 允许的源
	origin := ctx.Input.Header("Origin")
	if origin == "" {
		origin = "*"
	}

	// 设置 CORS 响应头
	ctx.Output.Header("Access-Control-Allow-Origin", origin)
	ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	ctx.Output.Header("Access-Control-Allow-Credentials", "true")
	ctx.Output.Header("Access-Control-Max-Age", "86400")

	// 处理 OPTIONS 预检请求
	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.SetStatus(204)
		return
	}
}
