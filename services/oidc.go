// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/oauth-server/oauth-server/models"
)

// CreateOidcClient 创建新的 OIDC 客户端应用
// Requirements: 7.4
func CreateOidcClient(
	clientName string,
	redirectUris []string,
	grantTypes []string,
	responseTypes []string,
	scope string,
	tokenEndpointAuthMethod string,
	logoUri string,
) (string, string, error) {
	// 生成 client_id 和 client_secret
	clientId := models.GenerateClientId()
	clientSecret := models.GenerateClientSecret()

	// 确定是否为公开客户端（不需要 client_secret）
	isPublicClient := tokenEndpointAuthMethod == "none"
	if isPublicClient {
		clientSecret = "" // 公开客户端不需要 secret
	}

	// 创建 Application 对象
	application := &models.Application{
		Owner:                "built-in",
		Name:                 fmt.Sprintf("oidc-%s", clientId),
		CreatedTime:          models.GetCurrentTime(),
		DisplayName:          clientName,
		Logo:                 logoUri,
		HomepageUrl:          "",
		Description:          fmt.Sprintf("OIDC client: %s", clientName),
		Organization:         "built-in",
		ClientId:             clientId,
		ClientSecret:         clientSecret,
		RedirectUris:         redirectUris,
		TokenFormat:          "JWT",
		ExpireInHours:        1,
		RefreshExpireInHours: 168, // 7 days
		GrantTypes:           grantTypes,
		Tags:                 []string{"oidc"},
		EnablePassword:       true,
		EnableSignUp:         false,
		EnableCodeSignin:     false,
		Cert:                 "",
		Scopes:               strings.Split(scope, " "),
	}

	// 保存到数据库
	_, err := models.AddApplication(application)
	if err != nil {
		return "", "", fmt.Errorf("failed to create application: %w", err)
	}

	return clientId, clientSecret, nil
}

// GetCurrentTimestamp 返回当前时间戳（秒）
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
