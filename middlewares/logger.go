// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package middlewares

import (
	"encoding/json"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LogEntry 结构化日志条目
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Latency    string `json:"latency"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent,omitempty"`
	Error      string `json:"error,omitempty"`
}

// LoggerMiddleware 返回一个记录所有 HTTP 请求的中间件
// 使用结构化 JSON 格式记录：方法、路径、状态码、响应时间
func LoggerMiddleware() fiber.Handler {
	// 获取日志级别配置（默认为 info）
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return func(c *fiber.Ctx) error {
		// 记录开始时间
		start := time.Now()

		// 处理请求
		err := c.Next()

		// 计算响应时间
		latency := time.Since(start)

		// 构造日志条目
		logEntry := LogEntry{
			Timestamp:  time.Now().Format(time.RFC3339),
			Method:     c.Method(),
			Path:       c.Path(),
			StatusCode: c.Response().StatusCode(),
			Latency:    latency.String(),
			IP:         c.IP(),
			UserAgent:  c.Get("User-Agent"),
		}

		// 如果有错误，记录错误信息
		if err != nil {
			logEntry.Error = err.Error()
		}

		// 根据日志级别和状态码决定是否记录
		shouldLog := false
		switch logLevel {
		case "debug":
			shouldLog = true
		case "info":
			shouldLog = true
		case "warn":
			shouldLog = logEntry.StatusCode >= 400
		case "error":
			shouldLog = logEntry.StatusCode >= 500
		default:
			shouldLog = true
		}

		// 输出 JSON 格式的日志
		if shouldLog {
			logJSON, _ := json.Marshal(logEntry)
			os.Stdout.Write(logJSON)
			os.Stdout.Write([]byte("\n"))
		}

		return err
	}
}
