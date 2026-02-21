// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/types"
)

// RecoveryLogEntry panic 恢复日志条目
type RecoveryLogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	IP         string `json:"ip"`
	Error      string `json:"error"`
	StackTrace string `json:"stack_trace"`
}

// RecoveryMiddleware 返回一个捕获 panic 的中间件
// 当处理请求时发生 panic，返回 500 错误并记录堆栈跟踪
func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// 获取堆栈跟踪
				stackTrace := string(debug.Stack())

				// 构造恢复日志条目
				logEntry := RecoveryLogEntry{
					Timestamp:  time.Now().Format(time.RFC3339),
					Level:      "error",
					Message:    "Panic recovered",
					Method:     c.Method(),
					Path:       c.Path(),
					IP:         c.IP(),
					Error:      fmt.Sprintf("%v", r),
					StackTrace: stackTrace,
				}

				// 输出 JSON 格式的错误日志
				logJSON, _ := json.Marshal(logEntry)
				os.Stderr.Write(logJSON)
				os.Stderr.Write([]byte("\n"))

				// 返回 500 错误响应
				c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("服务器内部错误"))
			}
		}()

		// 继续处理请求
		return c.Next()
	}
}
