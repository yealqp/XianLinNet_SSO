// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/services"
	"github.com/oauth-server/oauth-server/types"
)

// HandleDiscovery 处理 OIDC Discovery 端点
// 返回 /.well-known/openid-configuration
// Requirements: 7.1
func HandleDiscovery() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取服务器的 origin（issuer）
		origin := os.Getenv("ORIGIN")
		if origin == "" {
			// 从请求中构建 origin
			scheme := "http"
			if ctx.Protocol() == "https" {
				scheme = "https"
			}
			origin = scheme + "://" + ctx.Hostname()
		}

		// 构造 OIDC Discovery 响应
		discovery := map[string]interface{}{
			"issuer":                                origin,
			"authorization_endpoint":                origin + "/oauth/authorize",
			"token_endpoint":                        origin + "/api/oauth/token",
			"userinfo_endpoint":                     origin + "/api/userinfo",
			"jwks_uri":                              origin + "/.well-known/jwks",
			"registration_endpoint":                 origin + "/api/oauth/register",
			"introspection_endpoint":                origin + "/api/oauth/introspect",
			"revocation_endpoint":                   origin + "/api/oauth/revoke",
			"response_types_supported":              []string{"code", "token", "id_token", "code token", "code id_token", "token id_token", "code token id_token"},
			"response_modes_supported":              []string{"query", "fragment", "form_post"},
			"grant_types_supported":                 []string{"authorization_code", "implicit", "password", "client_credentials", "refresh_token", "urn:ietf:params:oauth:grant-type:token-exchange"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"HS256", "RS256"},
			"scopes_supported":                      []string{"openid", "profile", "email", "address", "phone", "offline_access"},
			"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post", "none"},
			"claims_supported":                      []string{"sub", "iss", "aud", "exp", "iat", "name", "email", "picture", "preferred_username", "email_verified", "updated_at"},
			"code_challenge_methods_supported":      []string{"S256", "plain"},
		}

		return ctx.JSON(discovery)
	}
}

// HandleJwks 处理 JWKS 端点
// 返回 JSON Web Key Set
// Requirements: 7.2
func HandleJwks() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取公钥的 JWK 表示
		jwk, err := services.GetPublicKeyJWK()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取公钥失败"))
		}

		// 构造 JWKS 响应
		jwks := map[string]interface{}{
			"keys": []interface{}{jwk},
		}

		return ctx.JSON(jwks)
	}
}

// HandleUserInfo 处理 UserInfo 端点
// 需要 JWT 认证，返回用户基本信息
// Requirements: 7.3
func HandleUserInfo() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 从 Authorization 头提取 token
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("缺少认证令牌"))
		}

		// 提取 Bearer token
		token := ""
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			token = authHeader[7:]
		} else {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("无效的认证格式"))
		}

		// 验证 token 并获取用户信息
		user, err := services.ValidateToken(token)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("令牌无效"))
		}

		// 解析 token 以获取 scope 信息
		claims, err := services.ParseJwtToken(token)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("令牌声明无效"))
		}

		// 构造用户信息响应（基于 scope）
		userInfo := map[string]interface{}{
			"sub": user.GetId(),
		}

		// 检查 scope 并添加相应的声明
		scopes := claims.Scope

		// Profile scope - 包含头像
		if containsScope(scopes, "profile") || containsScope(scopes, "openid") {
			if user.Avatar != "" {
				userInfo["picture"] = user.Avatar
			}
		}

		// Email scope - 包含邮箱信息
		if containsScope(scopes, "email") || containsScope(scopes, "openid") {
			userInfo["email"] = user.Email
		}

		// 添加自定义声明（始终包含）
		userInfo["id"] = user.Id
		userInfo["username"] = user.Username
		if user.QQ != "" {
			userInfo["qq"] = user.QQ
		}
		if user.Avatar != "" {
			userInfo["avatar"] = user.Avatar
		}
		userInfo["is_real_name"] = user.IsRealName
		userInfo["is_admin"] = user.IsAdmin

		// 返回标准的 ApiResponse 格式
		return ctx.JSON(types.SuccessResponse(userInfo))
	}
}

// HandleOidcRegister 处理 OIDC 动态客户端注册
// 解析客户端注册请求，创建新的 OIDC 客户端应用
// Requirements: 7.4
func HandleOidcRegister() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析注册请求
		var req struct {
			ClientName              string   `json:"client_name"`
			RedirectUris            []string `json:"redirect_uris"`
			GrantTypes              []string `json:"grant_types"`
			ResponseTypes           []string `json:"response_types"`
			Scope                   string   `json:"scope"`
			TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
			LogoUri                 string   `json:"logo_uri"`
			Contacts                []string `json:"contacts"`
		}

		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证必需字段
		if req.ClientName == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("client_name 是必需的"))
		}

		if len(req.RedirectUris) == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("redirect_uris 是必需的"))
		}

		// 设置默认值
		if len(req.GrantTypes) == 0 {
			req.GrantTypes = []string{"authorization_code"}
		}

		if len(req.ResponseTypes) == 0 {
			req.ResponseTypes = []string{"code"}
		}

		if req.Scope == "" {
			req.Scope = "openid profile email"
		}

		if req.TokenEndpointAuthMethod == "" {
			req.TokenEndpointAuthMethod = "client_secret_basic"
		}

		// 调用服务层创建 OIDC 客户端
		clientId, clientSecret, err := services.CreateOidcClient(
			req.ClientName,
			req.RedirectUris,
			req.GrantTypes,
			req.ResponseTypes,
			req.Scope,
			req.TokenEndpointAuthMethod,
			req.LogoUri,
		)

		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("创建客户端失败: " + err.Error()))
		}

		// 构造响应
		response := map[string]interface{}{
			"client_id":                  clientId,
			"client_secret":              clientSecret,
			"client_name":                req.ClientName,
			"redirect_uris":              req.RedirectUris,
			"grant_types":                req.GrantTypes,
			"response_types":             req.ResponseTypes,
			"scope":                      req.Scope,
			"token_endpoint_auth_method": req.TokenEndpointAuthMethod,
			"client_id_issued_at":        services.GetCurrentTimestamp(),
			"client_secret_expires_at":   0, // 不过期
		}

		if req.LogoUri != "" {
			response["logo_uri"] = req.LogoUri
		}

		if len(req.Contacts) > 0 {
			response["contacts"] = req.Contacts
		}

		return ctx.Status(fiber.StatusCreated).JSON(response)
	}
}

// containsScope 检查 scope 字符串中是否包含指定的 scope
func containsScope(scopes, scope string) bool {
	if scopes == "" {
		return false
	}
	scopeList := strings.Split(scopes, " ")
	for _, s := range scopeList {
		if s == scope {
			return true
		}
	}
	return false
}
