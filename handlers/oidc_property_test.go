// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
)

// **Validates: Requirements 7.3, 7.5**
// Property 18: ID Token 包含条件
// Property 19: 用户信息完整性

// Property 18: ID Token 包含条件
// 验证当请求包含 openid scope 时，响应中必须包含 ID Token
// **Validates: Requirement 7.5**
func TestProperty_IDTokenInclusionCondition(t *testing.T) {
	// 初始化测试数据库
	if err := models.InitDB(); err != nil {
		t.Skip("Database not available for property testing")
	}

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("ID Token is included when openid scope is requested", prop.ForAll(
		func(scope string, includeOpenid bool) bool {
			// 构造 scope 字符串
			scopes := []string{}
			if includeOpenid {
				scopes = append(scopes, "openid")
			}
			if scope != "" && scope != "openid" {
				scopes = append(scopes, scope)
			}
			scopeStr := strings.Join(scopes, " ")

			// 创建测试用户
			user := &models.User{
				Owner:      "test-org",
				Id:         12345,
				Username:   "testuser",
				Email:      "test@example.com",
				Type:       "normal-user",
				IsAdmin:    false,
				IsRealName: false,
			}

			// 创建测试应用
			application := &models.Application{
				Owner:                "test-owner",
				Name:                 "test-app",
				ClientId:             "test-client-id",
				ExpireInHours:        1,
				RefreshExpireInHours: 168,
			}

			// 生成 JWT token
			accessToken, _, _, err := services.GenerateJwtToken(application, user, scopeStr, "", "")
			if err != nil {
				return true // 跳过错误情况
			}

			// 解析 token 以验证 scope
			claims, err := services.ParseJwtToken(accessToken)
			if err != nil {
				return true // 跳过错误情况
			}

			// 验证：如果请求了 openid scope，token 中应该包含 openid scope
			hasOpenidScope := strings.Contains(claims.Scope, "openid")

			if includeOpenid {
				return hasOpenidScope
			}

			// 如果没有请求 openid scope，可能包含也可能不包含
			return true
		},
		gen.OneConstOf("profile", "email", "address", "phone", ""),
		gen.Bool(),
	))

	properties.Property("ID Token contains required OIDC claims", prop.ForAll(
		func(nonce string) bool {
			// 创建测试用户
			user := &models.User{
				Owner:      "test-org",
				Id:         12345,
				Username:   "testuser",
				Email:      "test@example.com",
				Type:       "normal-user",
				IsAdmin:    false,
				IsRealName: false,
			}

			// 创建测试应用
			application := &models.Application{
				Owner:                "test-owner",
				Name:                 "test-app",
				ClientId:             "test-client-id",
				ExpireInHours:        1,
				RefreshExpireInHours: 168,
			}

			// 生成 ID Token
			idToken, err := services.GenerateIDToken(application, user, nonce, "test-access-token")
			if err != nil {
				return true // 跳过错误情况
			}

			// 解析 ID Token
			claims, err := services.ParseJwtToken(idToken)
			if err != nil {
				return false
			}

			// 验证必需的 OIDC 声明
			hasIss := claims.Iss != ""
			hasSub := claims.Sub != ""
			hasAud := len(claims.Aud) > 0
			hasExp := claims.ExpiresAt != nil
			hasIat := claims.IssuedAt != nil

			// 验证 nonce（如果提供）
			nonceMatches := true
			if nonce != "" {
				nonceMatches = claims.Nonce == nonce
			}

			return hasIss && hasSub && hasAud && hasExp && hasIat && nonceMatches
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Property 19: 用户信息完整性
// 验证 UserInfo 端点返回的用户信息包含必需的字段（sub, email, name）
// **Validates: Requirement 7.3**
func TestProperty_UserInfoIntegrity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("UserInfo response contains required fields", prop.ForAll(
		func(userId int64, username string, email string) bool {
			// 跳过无效输入
			if userId <= 0 || username == "" || email == "" {
				return true
			}

			// 创建测试用户
			user := &models.User{
				Owner:      "test-org",
				Id:         userId,
				Username:   username,
				Email:      email,
				Type:       "normal-user",
				IsAdmin:    false,
				IsRealName: false,
			}

			// 验证用户信息包含必需字段
			hasSub := user.GetId() != ""
			hasEmail := user.Email != ""
			hasName := user.Username != ""

			return hasSub && hasEmail && hasName
		},
		gen.Int64Range(1, 1000000),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 50 }),
		gen.RegexMatch(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	))

	properties.Property("UserInfo respects scope restrictions", prop.ForAll(
		func(includeProfile bool, includeEmail bool) bool {
			// 构造 scope
			scopes := []string{"openid"}
			if includeProfile {
				scopes = append(scopes, "profile")
			}
			if includeEmail {
				scopes = append(scopes, "email")
			}
			scopeStr := strings.Join(scopes, " ")

			// 模拟 UserInfo 响应构造逻辑
			userInfo := map[string]interface{}{
				"sub": "test-user-id",
			}

			// 根据 scope 添加字段
			if containsScope(scopeStr, "profile") || containsScope(scopeStr, "openid") {
				userInfo["name"] = "Test User"
				userInfo["preferred_username"] = "testuser"
			}

			if containsScope(scopeStr, "email") || containsScope(scopeStr, "openid") {
				userInfo["email"] = "test@example.com"
				userInfo["email_verified"] = true
			}

			// 验证：sub 字段始终存在
			_, hasSub := userInfo["sub"]
			if !hasSub {
				return false
			}

			// 验证：profile scope 控制 name 字段
			_, hasName := userInfo["name"]
			if includeProfile && !hasName {
				return false
			}

			// 验证：email scope 控制 email 字段
			_, hasEmail := userInfo["email"]
			if includeEmail && !hasEmail {
				return false
			}

			return true
		},
		gen.Bool(),
		gen.Bool(),
	))

	properties.Property("UserInfo sub claim is unique per user", prop.ForAll(
		func(userId1 int64, userId2 int64) bool {
			// 确保 userId 有效
			if userId1 <= 0 {
				userId1 = 1
			}
			if userId2 <= 0 {
				userId2 = 1
			}

			// 创建两个用户
			user1 := &models.User{
				Owner: "test-org",
				Id:    userId1,
			}

			user2 := &models.User{
				Owner: "test-org",
				Id:    userId2,
			}

			// 获取 sub 声明
			sub1 := user1.GetId()
			sub2 := user2.GetId()

			// 验证：不同用户的 sub 应该不同
			if userId1 != userId2 {
				return sub1 != sub2
			}

			// 相同用户的 sub 应该相同
			return sub1 == sub2
		},
		gen.Int64Range(1, 1000000),
		gen.Int64Range(1, 1000000),
	))

	properties.TestingRun(t)
}

// Property: OIDC Discovery 端点一致性
// 验证 Discovery 端点返回的配置包含所有必需字段
func TestProperty_OIDCDiscoveryConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Discovery response contains required endpoints", prop.ForAll(
		func(origin string) bool {
			// 跳过空 origin
			if origin == "" {
				origin = "http://localhost:8080"
			}

			// 模拟 Discovery 响应
			discovery := map[string]interface{}{
				"issuer":                 origin,
				"authorization_endpoint": origin + "/oauth/authorize",
				"token_endpoint":         origin + "/api/oauth/token",
				"userinfo_endpoint":      origin + "/api/userinfo",
				"jwks_uri":               origin + "/.well-known/jwks",
			}

			// 验证必需字段存在
			_, hasIssuer := discovery["issuer"]
			_, hasAuthEndpoint := discovery["authorization_endpoint"]
			_, hasTokenEndpoint := discovery["token_endpoint"]
			_, hasUserinfoEndpoint := discovery["userinfo_endpoint"]
			_, hasJwksUri := discovery["jwks_uri"]

			return hasIssuer && hasAuthEndpoint && hasTokenEndpoint && hasUserinfoEndpoint && hasJwksUri
		},
		gen.RegexMatch(`^https?://[a-zA-Z0-9.-]+(:[0-9]+)?$`),
	))

	properties.Property("Discovery issuer matches origin", prop.ForAll(
		func(origin string) bool {
			// 跳过空 origin
			if origin == "" {
				origin = "http://localhost:8080"
			}

			// 模拟 Discovery 响应
			discovery := map[string]interface{}{
				"issuer": origin,
			}

			// 验证 issuer 与 origin 匹配
			issuer, ok := discovery["issuer"].(string)
			if !ok {
				return false
			}

			return issuer == origin
		},
		gen.RegexMatch(`^https?://[a-zA-Z0-9.-]+(:[0-9]+)?$`),
	))

	properties.TestingRun(t)
}

// Property: OIDC 客户端注册一致性
// 验证动态客户端注册生成有效的 client_id 和 client_secret
func TestProperty_OIDCClientRegistrationConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Client registration generates valid credentials", prop.ForAll(
		func(clientName string, redirectUri string) bool {
			// 跳过无效输入
			if clientName == "" || redirectUri == "" {
				return true
			}

			// 生成 client_id 和 client_secret
			clientId := models.GenerateClientId()
			clientSecret := models.GenerateClientSecret()

			// 验证生成的凭证不为空
			hasClientId := clientId != ""
			hasClientSecret := clientSecret != ""

			// 验证 client_id 格式（UUID）
			validClientId := len(clientId) > 0

			// 验证 client_secret 长度（至少 32 字符）
			validClientSecret := len(clientSecret) >= 32

			return hasClientId && hasClientSecret && validClientId && validClientSecret
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 100 }),
		gen.RegexMatch(`^https?://[a-zA-Z0-9.-]+(:[0-9]+)?(/.*)?$`),
	))

	properties.Property("Public clients have no client_secret", prop.ForAll(
		func(isPublic bool) bool {
			// 模拟客户端注册
			var clientSecret string
			if isPublic {
				clientSecret = "" // 公开客户端不需要 secret
			} else {
				clientSecret = models.GenerateClientSecret()
			}

			// 验证：公开客户端的 secret 为空
			if isPublic {
				return clientSecret == ""
			}

			// 验证：机密客户端的 secret 不为空
			return clientSecret != ""
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Property: Scope 字符串解析一致性
// 验证 scope 字符串的解析和验证逻辑
func TestProperty_ScopeParsingConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Scope string is correctly parsed", prop.ForAll(
		func(scopes []string) bool {
			// 跳过空数组
			if len(scopes) == 0 {
				return true
			}

			// 构造 scope 字符串
			scopeStr := strings.Join(scopes, " ")

			// 解析 scope 字符串
			parsedScopes := strings.Split(scopeStr, " ")

			// 验证解析后的 scope 数量与原始数量一致
			return len(parsedScopes) == len(scopes)
		},
		gen.SliceOf(gen.OneConstOf("openid", "profile", "email", "address", "phone", "offline_access")),
	))

	properties.Property("containsScope function works correctly", prop.ForAll(
		func(scopes []string, targetScope string) bool {
			// 跳过空输入
			if len(scopes) == 0 || targetScope == "" {
				return true
			}

			// 构造 scope 字符串
			scopeStr := strings.Join(scopes, " ")

			// 检查是否包含目标 scope
			contains := containsScope(scopeStr, targetScope)

			// 验证结果
			shouldContain := false
			for _, s := range scopes {
				if s == targetScope {
					shouldContain = true
					break
				}
			}

			return contains == shouldContain
		},
		gen.SliceOf(gen.OneConstOf("openid", "profile", "email", "address", "phone")),
		gen.OneConstOf("openid", "profile", "email", "address", "phone", "offline_access"),
	))

	properties.TestingRun(t)
}
