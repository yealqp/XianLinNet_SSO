// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/services"
	"github.com/oauth-server/oauth-server/types"
)

// HandleToken 处理 OAuth Token 请求
// 支持 authorization_code、refresh_token、password 授权类型
// Requirements: 5.4, 5.5, 5.6, 5.8, 6.1, 6.2
func HandleToken() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析 TokenRequest
		var req types.TokenRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证必需参数
		if req.GrantType == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("grant_type 不能为空"))
		}

		// 根据授权类型验证其他必需参数
		switch req.GrantType {
		case "authorization_code":
			if req.Code == "" || req.ClientId == "" {
				return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("code 和 client_id 不能为空"))
			}
		case "refresh_token":
			if req.RefreshToken == "" {
				return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("refresh_token 不能为空"))
			}
		case "password":
			if req.Username == "" || req.Password == "" || req.ClientId == "" {
				return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("username、password 和 client_id 不能为空"))
			}
		default:
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("不支持的 grant_type"))
		}

		// 调用 OAuth 服务获取 token
		result, err := services.GetOAuthToken(
			req.GrantType,
			req.ClientId,
			req.ClientSecret,
			req.Code,
			"", // code_verifier (PKCE)
			req.Scope,
			req.Username,
			req.Password,
			req.RefreshToken,
			"", // resource
		)

		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse(err.Error()))
		}

		// 检查是否返回了 TokenError
		if tokenError, ok := result.(*services.TokenError); ok {
			// 根据错误类型返回适当的 HTTP 状态码
			statusCode := fiber.StatusBadRequest
			switch tokenError.Error {
			case services.InvalidClient:
				statusCode = fiber.StatusUnauthorized
			case services.InvalidGrant:
				statusCode = fiber.StatusBadRequest
			case services.UnauthorizedClient:
				statusCode = fiber.StatusUnauthorized
			case services.UnsupportedGrantType:
				statusCode = fiber.StatusBadRequest
			}

			return ctx.Status(statusCode).JSON(map[string]interface{}{
				"error":             tokenError.Error,
				"error_description": tokenError.ErrorDescription,
			})
		}

		// 返回成功的 TokenResponse
		return ctx.JSON(result)
	}
}

// HandleIntrospect 处理 Token 自省请求
// 验证 token 有效性并返回元数据
// Requirements: 6.3
func HandleIntrospect() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析 token 参数
		token := ctx.FormValue("token")
		if token == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("token 不能为空"))
		}

		// 解析和验证 token
		claims, err := services.ParseJwtToken(token)
		if err != nil {
			// Token 无效，返回 active: false
			return ctx.JSON(map[string]interface{}{
				"active": false,
			})
		}

		// 返回 token 元数据
		return ctx.JSON(map[string]interface{}{
			"active":     true,
			"scope":      claims.Scope,
			"client_id":  claims.Aud[0],
			"username":   claims.Username,
			"token_type": "Bearer",
			"exp":        claims.ExpiresAt.Unix(),
			"iat":        claims.IssuedAt.Unix(),
			"sub":        claims.Sub,
			"iss":        claims.Iss,
		})
	}
}

// HandleRevoke 处理 Token 撤销请求
// 将 token 标记为无效
// Requirements: 6.4
func HandleRevoke() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析 token 参数
		token := ctx.FormValue("token")
		if token == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("token 不能为空"))
		}

		tokenTypeHint := ctx.FormValue("token_type_hint")

		// 解析 token 以获取信息
		claims, err := services.ParseJwtToken(token)
		if err != nil {
			// RFC 7009: 即使 token 无效也返回成功
			return ctx.JSON(types.SuccessResponse(map[string]interface{}{
				"revoked": true,
			}))
		}

		// 查找并撤销 token
		err = services.RevokeToken(token, tokenTypeHint)
		if err != nil {
			// 即使撤销失败也返回成功（按照 RFC 7009）
			return ctx.JSON(types.SuccessResponse(map[string]interface{}{
				"revoked": true,
				"jti":     claims.Id,
			}))
		}

		// 返回成功响应
		return ctx.JSON(types.SuccessResponse(map[string]interface{}{
			"revoked": true,
			"jti":     claims.Id,
		}))
	}
}
