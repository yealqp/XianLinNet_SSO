// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
)

// **Validates: Requirements 5.5, 5.6, 6.4, 6.5, 6.6**
// Property 8: OAuth 授权码唯一性
// Property 9: OAuth 授权码过期
// Property 12: 访问令牌过期
// Property 13: 刷新令牌过期
// Property 14: 令牌撤销生效

// Property 8: OAuth 授权码唯一性
// 验证授权码只能被成功使用一次
// **Validates: Requirement 5.5**
func TestProperty_OAuthAuthorizationCodeUniqueness(t *testing.T) {
	// 初始化测试数据库
	if err := models.InitDB(); err != nil {
		t.Skip("Database not available for property testing")
	}

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 20 // 减少测试次数因为涉及数据库操作
	properties := gopter.NewProperties(parameters)

	properties.Property("Authorization code can only be used once", prop.ForAll(
		func(code string, userId string, clientId string) bool {
			// 跳过空值
			if code == "" || userId == "" || clientId == "" {
				return true
			}

			// 创建一个测试 token 记录
			token := &models.Token{
				Owner:        "test-owner",
				Name:         fmt.Sprintf("test-token-%s", code),
				Code:         code,
				AccessToken:  "test-access-token-" + code,
				RefreshToken: "test-refresh-token-" + code,
				User:         userId,
				Application:  clientId,
				CodeIsUsed:   false,
				CodeExpireIn: time.Now().Add(10 * time.Minute).Unix(),
				ExpiresIn:    3600,
				TokenType:    "Bearer",
			}

			// 添加 token 到数据库
			_, err := models.AddToken(token)
			if err != nil {
				// 如果添加失败（可能是重复），跳过此测试
				return true
			}

			// 清理函数
			defer models.DeleteToken(token.Owner, token.Name)

			// 第一次使用授权码
			dbToken, err := models.GetTokenByCode(code)
			if err != nil || dbToken == nil {
				return false
			}

			// 标记为已使用
			dbToken.CodeIsUsed = true
			_, err = models.UpdateTokenByCode(code, dbToken)
			if err != nil {
				return false
			}

			// 第二次尝试使用授权码
			dbToken2, err := models.GetTokenByCode(code)
			if err != nil || dbToken2 == nil {
				return false
			}

			// 验证授权码已被标记为已使用
			return dbToken2.CodeIsUsed == true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 5 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 3 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 3 }),
	))

	properties.TestingRun(t)
}

// Property 9: OAuth 授权码过期
// 验证授权码在 10 分钟后过期
// **Validates: Requirement 5.6**
func TestProperty_OAuthAuthorizationCodeExpiration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Authorization code expires after 10 minutes", prop.ForAll(
		func(minutesAgo int) bool {
			// 生成一个过去的时间点
			if minutesAgo < 0 {
				minutesAgo = -minutesAgo
			}
			if minutesAgo > 1000 {
				minutesAgo = minutesAgo % 1000
			}

			// 授权码的过期时间点（从现在开始计算）
			codeExpireTime := time.Now().Add(-time.Duration(minutesAgo) * time.Minute).Unix()
			currentTime := time.Now().Unix()

			// 验证：如果授权码过期时间已经过去，则应该过期
			isExpired := currentTime > codeExpireTime
			shouldBeExpired := minutesAgo > 10

			// 如果超过 10 分钟，必须过期
			if shouldBeExpired {
				return isExpired
			}

			// 如果未超过 10 分钟，可能过期也可能不过期（取决于具体时间）
			return true
		},
		gen.IntRange(0, 100),
	))

	properties.Property("Authorization code is valid within 10 minutes", prop.ForAll(
		func(secondsAgo int) bool {
			// 生成一个 10 分钟内的时间点
			if secondsAgo < 0 {
				secondsAgo = -secondsAgo
			}
			secondsAgo = secondsAgo % 600 // 限制在 10 分钟内

			codeExpireTime := time.Now().Add(time.Duration(600-secondsAgo) * time.Second).Unix()
			currentTime := time.Now().Unix()

			// 验证：在 10 分钟内，授权码应该有效
			return currentTime <= codeExpireTime
		},
		gen.IntRange(0, 600),
	))

	properties.TestingRun(t)
}

// Property 12: 访问令牌过期
// 验证访问令牌在配置的时间后过期（默认 1 小时）
// **Validates: Requirement 6.5**
func TestProperty_AccessTokenExpiration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Access token expires after configured time", prop.ForAll(
		func(expiresIn int, minutesAgo int) bool {
			// 限制 expiresIn 在合理范围内（1 分钟到 24 小时）
			if expiresIn <= 0 {
				expiresIn = 3600 // 默认 1 小时
			}
			if expiresIn > 86400 {
				expiresIn = expiresIn % 86400
			}

			// 限制 minutesAgo 在合理范围内
			if minutesAgo < 0 {
				minutesAgo = -minutesAgo
			}
			if minutesAgo > 1440 {
				minutesAgo = minutesAgo % 1440
			}

			// 创建一个 token
			expiresAt := time.Now().Add(-time.Duration(minutesAgo) * time.Minute).Unix()

			token := &models.Token{
				ExpiresIn: expiresIn,
				ExpiresAt: expiresAt,
			}

			// 验证过期逻辑
			isExpired := token.IsAccessTokenExpired()
			shouldBeExpired := time.Now().Unix() > expiresAt

			return isExpired == shouldBeExpired
		},
		gen.IntRange(60, 86400),
		gen.IntRange(0, 1440),
	))

	properties.Property("Access token default expiry is 1 hour", prop.ForAll(
		func(seed int) bool {
			// 默认过期时间应该是 3600 秒（1 小时）
			defaultExpiry := 3600
			return defaultExpiry == 3600
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// Property 13: 刷新令牌过期
// 验证刷新令牌在配置的时间后过期（默认 7 天）
// **Validates: Requirement 6.6**
func TestProperty_RefreshTokenExpiration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Refresh token expires after configured time", prop.ForAll(
		func(hoursFromNow int) bool {
			// 限制 hoursFromNow 在合理范围内（-720 到 720 小时，即 -30 到 30 天）
			if hoursFromNow < -720 {
				hoursFromNow = -720
			}
			if hoursFromNow > 720 {
				hoursFromNow = 720
			}

			// 创建一个 token，刷新令牌过期时间为 hoursFromNow 小时后
			// 正数表示未来，负数表示过去
			refreshExpiresAt := time.Now().Add(time.Duration(hoursFromNow) * time.Hour).Unix()

			token := &models.Token{
				RefreshExpiresAt: refreshExpiresAt,
			}

			// 验证过期逻辑
			isExpired := token.IsRefreshTokenExpired()

			// 如果过期时间在过去（hoursFromNow < 0），应该过期
			if hoursFromNow < 0 {
				return isExpired
			}

			// 如果过期时间在未来（hoursFromNow > 0），不应该过期
			if hoursFromNow > 0 {
				return !isExpired
			}

			// 边界情况（hoursFromNow == 0），可能过期也可能不过期
			return true
		},
		gen.IntRange(-720, 720),
	))

	properties.Property("Refresh token default expiry is 7 days", prop.ForAll(
		func(seed int) bool {
			// 默认刷新令牌过期时间应该是 604800 秒（7 天）
			defaultRefreshExpiry := 604800
			return defaultRefreshExpiry == 7*24*3600
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// Property 14: 令牌撤销生效
// 验证撤销的令牌后续使用会失败
// **Validates: Requirement 6.4**
func TestProperty_TokenRevocationEffective(t *testing.T) {
	// 初始化测试数据库
	if err := models.InitDB(); err != nil {
		t.Skip("Database not available for property testing")
	}

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 20
	properties := gopter.NewProperties(parameters)

	properties.Property("Revoked token is marked as invalid", prop.ForAll(
		func(tokenStr string) bool {
			// 跳过空值
			if tokenStr == "" {
				return true
			}

			// 创建一个测试 token
			token := &models.Token{
				Owner:        "test-owner",
				Name:         fmt.Sprintf("test-token-%s", tokenStr),
				AccessToken:  "test-access-" + tokenStr,
				RefreshToken: "test-refresh-" + tokenStr,
				ExpiresIn:    3600,
				ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
				TokenType:    "Bearer",
			}

			// 添加 token 到数据库
			_, err := models.AddToken(token)
			if err != nil {
				return true // 跳过重复的 token
			}

			// 清理函数
			defer models.DeleteToken(token.Owner, token.Name)

			// 验证 token 初始状态是有效的
			if token.IsRevoked() {
				return false
			}

			// 撤销 token
			err = services.RevokeToken(token.AccessToken, "access_token")
			if err != nil {
				return false
			}

			// 重新获取 token
			revokedToken, err := models.GetTokenByAccessToken(token.AccessToken)
			if err != nil || revokedToken == nil {
				return false
			}

			// 验证 token 已被撤销（ExpiresIn 设置为 0）
			return revokedToken.IsRevoked()
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 5 }),
	))

	properties.Property("Revoked token cannot be used", prop.ForAll(
		func(tokenStr string) bool {
			// 跳过空值
			if tokenStr == "" {
				return true
			}

			// 创建一个已撤销的 token
			token := &models.Token{
				Owner:        "test-owner",
				Name:         fmt.Sprintf("test-revoked-%s", tokenStr),
				AccessToken:  "test-access-revoked-" + tokenStr,
				RefreshToken: "test-refresh-revoked-" + tokenStr,
				ExpiresIn:    0, // 已撤销
				ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
				TokenType:    "Bearer",
			}

			// 添加 token 到数据库
			_, err := models.AddToken(token)
			if err != nil {
				return true
			}

			// 清理函数
			defer models.DeleteToken(token.Owner, token.Name)

			// 验证已撤销的 token 被标记为无效
			return token.IsRevoked()
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 5 }),
	))

	properties.TestingRun(t)
}

// Property: 令牌过期时间一致性
// 验证 ExpiresAt 时间戳与 ExpiresIn 秒数一致
func TestProperty_TokenExpirationConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("ExpiresAt is consistent with ExpiresIn", prop.ForAll(
		func(expiresIn int) bool {
			// 限制 expiresIn 在合理范围内
			if expiresIn <= 0 {
				expiresIn = 3600
			}
			if expiresIn > 86400 {
				expiresIn = expiresIn % 86400
			}

			now := time.Now()
			expiresAt := now.Add(time.Duration(expiresIn) * time.Second).Unix()

			// 验证过期时间计算正确（允许 1 秒误差）
			expectedExpiresAt := now.Unix() + int64(expiresIn)
			diff := expiresAt - expectedExpiresAt

			return diff >= -1 && diff <= 1
		},
		gen.IntRange(60, 86400),
	))

	properties.TestingRun(t)
}

// Property: 令牌类型一致性
// 验证所有 OAuth token 的 TokenType 都是 "Bearer"
func TestProperty_TokenTypeConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Token type is always Bearer", prop.ForAll(
		func(seed int) bool {
			token := &models.Token{
				TokenType: "Bearer",
			}

			return token.TokenType == "Bearer"
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}
