// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/types"
)

// HandleGetUserTokens 获取当前用户的 Token 列表
func HandleGetUserTokens() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(string)
		if !ok || userID == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("未授权"))
		}

		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		user, err := models.GetUserById(userIDInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		var tokens []*models.Token
		userIDStr := fmt.Sprintf("%d", userIDInt)
		log.Printf("[DEBUG] Querying tokens with user = '%s' (type: %T)", userIDStr, userIDStr)

		// Try with explicit CAST to ensure string comparison in PostgreSQL
		err = models.GetEngine().Where(`"user" = CAST(? AS VARCHAR)`, userIDStr).Find(&tokens)
		if err != nil {
			log.Printf("[ERROR] Failed to query tokens for user %s (ID: %s): %v", user.Username, userIDStr, err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取Token列表失败"))
		}

		log.Printf("[DEBUG] GetUserTokens - User: %s (ID: %s), Found %d tokens", user.Username, userIDStr, len(tokens))

		// Debug: Let's also check all tokens to see what's in the database
		var allTokens []*models.Token
		models.GetEngine().Limit(10).Find(&allTokens)
		log.Printf("[DEBUG] Total tokens in database (first 10): %d", len(allTokens))
		for i, t := range allTokens {
			log.Printf("[DEBUG] Token %d: User='%s', Application='%s', ExpiresIn=%d", i, t.User, t.Application, t.ExpiresIn)
		}

		// Try direct SQL query to debug
		var debugTokens []*models.Token
		models.GetEngine().SQL(`SELECT * FROM "token" WHERE "user" = '1' LIMIT 5`).Find(&debugTokens)
		log.Printf("[DEBUG] Direct SQL query found %d tokens", len(debugTokens))

		tokenList := make([]map[string]interface{}, 0, len(tokens))
		for _, token := range tokens {
			app, _ := models.GetApplication(token.Owner, token.Application)
			appDisplayName := token.Application
			appLogo := ""
			if app != nil {
				if app.DisplayName != "" {
					appDisplayName = app.DisplayName
				}
				appLogo = app.Logo
			}

			isExpired := false
			if token.ExpiresAt > 0 && time.Now().Unix() > token.ExpiresAt {
				isExpired = true
			}

			expiresAtStr := ""
			if token.ExpiresAt > 0 {
				expiresAtStr = time.Unix(token.ExpiresAt, 0).Format(time.RFC3339)
			}

			tokenList = append(tokenList, map[string]interface{}{
				"name":            token.Name,
				"createdTime":     token.CreatedTime,
				"application":     token.Application,
				"applicationName": appDisplayName,
				"applicationLogo": appLogo,
				"expiresAt":       expiresAtStr,
				"scope":           token.Scope,
				"tokenType":       token.TokenType,
				"isRevoked":       token.IsRevoked(),
				"isExpired":       isExpired,
			})
		}

		log.Printf("[DEBUG] Returning %d tokens", len(tokenList))
		return ctx.JSON(types.SuccessResponse(tokenList))
	}
}

// HandleRevokeUserToken 撤销当前用户的指定 Token
func HandleRevokeUserToken() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(string)
		if !ok || userID == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("未授权"))
		}

		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		user, err := models.GetUserById(userIDInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		tokenName := ctx.Params("name")
		if tokenName == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("Token名称不能为空"))
		}

		token, err := models.GetToken(user.Owner, tokenName)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取Token失败"))
		}
		if token == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("Token不存在"))
		}

		userIDStr := fmt.Sprintf("%d", userIDInt)
		if token.User != userIDStr {
			return ctx.Status(fiber.StatusForbidden).JSON(types.ErrorResponse("无权撤销此Token"))
		}

		token.ExpiresIn = 0
		_, err = models.UpdateToken(user.Owner, tokenName, token)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("撤销Token失败"))
		}

		return ctx.JSON(types.SuccessResponse(map[string]interface{}{
			"message": "Token已撤销",
		}))
	}
}

// HandleGetUserApplications 获取当前用户授权过的应用列表
func HandleGetUserApplications() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(string)
		if !ok || userID == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("未授权"))
		}

		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		user, err := models.GetUserById(userIDInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		var tokens []*models.Token
		userIDStr := fmt.Sprintf("%d", userIDInt)
		err = models.GetEngine().Where(`"user" = CAST(? AS VARCHAR)`, userIDStr).Find(&tokens)
		if err != nil {
			log.Printf("[ERROR] Failed to query tokens for user %s (ID: %s): %v", user.Username, userIDStr, err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取Token列表失败"))
		}

		log.Printf("[DEBUG] GetUserApplications - User: %s (ID: %s), Found %d tokens", user.Username, userIDStr, len(tokens))

		appMap := make(map[string]map[string]interface{})
		for _, token := range tokens {
			appKey := token.Owner + "/" + token.Application
			if _, exists := appMap[appKey]; !exists {
				app, _ := models.GetApplication(token.Owner, token.Application)

				appInfo := map[string]interface{}{
					"owner":       token.Owner,
					"name":        token.Application,
					"displayName": token.Application,
					"logo":        "",
					"description": "",
					"homepageUrl": "",
					"firstAuth":   token.CreatedTime,
					"lastAuth":    token.CreatedTime,
					"tokenCount":  1,
					"scopes":      []string{},
				}

				if app != nil {
					if app.DisplayName != "" {
						appInfo["displayName"] = app.DisplayName
					}
					appInfo["logo"] = app.Logo
					appInfo["description"] = app.Description
					appInfo["homepageUrl"] = app.HomepageUrl
				}

				if token.Scope != "" {
					appInfo["scopes"] = []string{token.Scope}
				}

				appMap[appKey] = appInfo
			} else {
				appInfo := appMap[appKey]
				appInfo["tokenCount"] = appInfo["tokenCount"].(int) + 1

				if token.CreatedTime > appInfo["lastAuth"].(string) {
					appInfo["lastAuth"] = token.CreatedTime
				}

				if token.CreatedTime < appInfo["firstAuth"].(string) {
					appInfo["firstAuth"] = token.CreatedTime
				}

				if token.Scope != "" {
					scopes := appInfo["scopes"].([]string)
					scopeExists := false
					for _, s := range scopes {
						if s == token.Scope {
							scopeExists = true
							break
						}
					}
					if !scopeExists {
						appInfo["scopes"] = append(scopes, token.Scope)
					}
				}
			}
		}

		appList := make([]map[string]interface{}, 0, len(appMap))
		for _, appInfo := range appMap {
			appList = append(appList, appInfo)
		}

		log.Printf("[DEBUG] Returning %d applications", len(appList))
		return ctx.JSON(types.SuccessResponse(appList))
	}
}
