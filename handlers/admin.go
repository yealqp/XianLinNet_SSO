// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
	"github.com/oauth-server/oauth-server/types"
	"golang.org/x/crypto/bcrypt"
)

// HandleGetUsers 获取用户列表（需要管理员权限）
// Requirements: 8.1
func HandleGetUsers() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取 owner 参数（可选）
		owner := ctx.Query("owner", "")

		// 获取用户列表
		var users []*models.User
		var err error

		if owner != "" {
			// 如果指定了 owner，只查询该 owner 的用户，按 ID 升序排序，排除已删除的用户
			err = models.GetEngine().Where("owner = ? AND is_deleted = ?", owner, false).OrderBy("id ASC").Find(&users)
		} else {
			// 否则查询所有用户，按 ID 升序排序，排除已删除的用户
			err = models.GetEngine().Where("is_deleted = ?", false).OrderBy("id ASC").Find(&users)
		}

		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户列表失败"))
		}

		// 构造响应数据（不返回敏感信息）
		userList := make([]types.UserInfo, 0, len(users))
		for _, user := range users {
			userList = append(userList, types.UserInfo{
				ID:          user.Id,
				Email:       user.Email,
				Username:    user.Username,
				IsAdmin:     user.IsAdmin,
				IsRealName:  user.IsRealName,
				IsForbidden: user.IsForbidden,
				QQ:          user.QQ,
				Avatar:      user.Avatar,
			})
		}

		return ctx.JSON(types.SuccessResponse(userList))
	}
}

// HandleGetUser 获取单个用户（需要管理员权限）
// Requirements: 8.1
func HandleGetUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取用户 ID
		idStr := ctx.Params("id")
		if idStr == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户ID不能为空"))
		}

		// 转换为 int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 获取用户
		user, err := models.GetUserById(id)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil || user.IsDeleted {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 构造响应数据
		userInfo := types.UserInfo{
			ID:          user.Id,
			Email:       user.Email,
			Username:    user.Username,
			IsAdmin:     user.IsAdmin,
			IsRealName:  user.IsRealName,
			IsForbidden: user.IsForbidden,
			QQ:          user.QQ,
			Avatar:      user.Avatar,
		}

		return ctx.JSON(types.SuccessResponse(userInfo))
	}
}

// HandleCreateUser 创建用户（需要管理员权限）
// Requirements: 8.2
func HandleCreateUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析请求
		var req types.CreateUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证参数
		if req.Username == "" || req.Email == "" || req.Password == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户名、邮箱和密码不能为空"))
		}

		// 验证用户名长度
		if len(req.Username) < 3 || len(req.Username) > 50 {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户名长度必须在3-50个字符之间"))
		}

		// 验证密码长度
		if len(req.Password) < 6 {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("密码长度至少为6个字符"))
		}

		// 检查邮箱是否已存在（排除已删除的用户）
		existingUser, err := models.GetUserByEmail(req.Email)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("检查邮箱失败"))
		}
		if existingUser != nil && !existingUser.IsDeleted {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("邮箱已存在"))
		}

		// 哈希密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("密码加密失败"))
		}

		// 创建用户
		now := time.Now().Format("2006-01-02T15:04:05Z07:00")
		user := &models.User{
			Owner:       "built-in",
			CreatedTime: now,
			UpdatedTime: now,
			Username:    req.Username,
			Email:       req.Email,
			Password:    string(hashedPassword),
			QQ:          req.QQ,
			Avatar:      req.Avatar,
			IsAdmin:     false,
			IsForbidden: false,
			IsDeleted:   false,
			IsRealName:  false,
		}

		// 保存到数据库
		_, err = models.AddUser(user)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("创建用户失败"))
		}

		// 返回用户信息
		userInfo := types.UserInfo{
			ID:          user.Id,
			Email:       user.Email,
			Username:    user.Username,
			IsAdmin:     user.IsAdmin,
			IsRealName:  user.IsRealName,
			IsForbidden: user.IsForbidden,
			QQ:          user.QQ,
			Avatar:      user.Avatar,
		}

		return ctx.JSON(types.SuccessResponse(userInfo))
	}
}

// HandleUpdateUser 更新用户（需要管理员权限）
// Requirements: 8.3
func HandleUpdateUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取用户 ID
		idStr := ctx.Params("id")
		if idStr == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户ID不能为空"))
		}

		// 转换为 int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 解析请求
		var req types.UpdateUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 获取用户
		user, err := models.GetUserById(id)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil || user.IsDeleted {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 更新用户信息
		if req.Username != "" {
			if len(req.Username) < 3 || len(req.Username) > 50 {
				return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户名长度必须在3-50个字符之间"))
			}
			user.Username = req.Username
		}
		if req.Email != "" {
			// 检查新邮箱是否已被其他用户使用（排除已删除的用户）
			existingUser, err := models.GetUserByEmail(req.Email)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("检查邮箱失败"))
			}
			if existingUser != nil && existingUser.Id != id && !existingUser.IsDeleted {
				return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("邮箱已被其他用户使用"))
			}
			user.Email = req.Email
		}
		if req.QQ != "" {
			user.QQ = req.QQ
		}
		if req.Avatar != "" {
			user.Avatar = req.Avatar
		}

		// 更新时间戳
		user.UpdatedTime = time.Now().Format("2006-01-02T15:04:05Z07:00")

		// 保存到数据库
		_, err = models.UpdateUser(id, user)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("更新用户失败"))
		}

		// 返回更新后的用户信息
		userInfo := types.UserInfo{
			ID:          user.Id,
			Email:       user.Email,
			Username:    user.Username,
			IsAdmin:     user.IsAdmin,
			IsRealName:  user.IsRealName,
			IsForbidden: user.IsForbidden,
			QQ:          user.QQ,
			Avatar:      user.Avatar,
		}

		return ctx.JSON(types.SuccessResponse(userInfo))
	}
}

// HandleDeleteUser 删除用户（需要管理员权限）
// Requirements: 8.4
func HandleDeleteUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取用户 ID
		idStr := ctx.Params("id")
		if idStr == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户ID不能为空"))
		}

		// 转换为 int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 获取用户
		user, err := models.GetUserById(id)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil || user.IsDeleted {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 不能删除管理员
		if user.IsAdmin {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("不能删除管理员账户"))
		}

		// 标记为已删除（软删除）
		user.IsDeleted = true
		user.UpdatedTime = time.Now().Format("2006-01-02T15:04:05Z07:00")

		// 保存到数据库
		_, err = models.UpdateUser(id, user)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("删除用户失败"))
		}

		return ctx.JSON(types.SuccessResponse(map[string]string{
			"message": "用户已删除",
		}))
	}
}

// HandleGetApplications 获取应用列表（需要管理员权限）
// Requirements: 8.5
func HandleGetApplications() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取 owner 参数（可选）
		owner := ctx.Query("owner", "")

		// 获取应用列表
		var applications []*models.Application
		var err error

		if owner != "" {
			// 如果指定了 owner，只查询该 owner 的应用
			err = models.GetEngine().Where("owner = ?", owner).Find(&applications)
		} else {
			// 否则查询所有应用
			err = models.GetEngine().Find(&applications)
		}

		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取应用列表失败"))
		}

		return ctx.JSON(types.SuccessResponse(applications))
	}
}

// HandleGetApplication 获取单个应用（需要管理员权限）
// Requirements: 8.5
func HandleGetApplication() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取参数
		owner := ctx.Params("owner")
		name := ctx.Params("name")

		if owner == "" || name == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("owner 和 name 参数不能为空"))
		}

		// 获取应用
		application, err := models.GetApplication(owner, name)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取应用信息失败"))
		}
		if application == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("应用不存在"))
		}

		return ctx.JSON(types.SuccessResponse(application))
	}
}

// HandleCreateApplication 创建应用（需要管理员权限）
// Requirements: 8.6
func HandleCreateApplication() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析请求
		var req types.CreateApplicationRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证参数
		if req.Name == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("应用名称不能为空"))
		}

		// 检查应用是否已存在
		owner := "built-in"
		existingApp, err := models.GetApplication(owner, req.Name)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("检查应用失败"))
		}
		if existingApp != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("应用已存在"))
		}

		// 创建应用
		now := time.Now().Format("2006-01-02T15:04:05Z07:00")
		application := &models.Application{
			Owner:                owner,
			Name:                 req.Name,
			CreatedTime:          now,
			DisplayName:          req.DisplayName,
			Logo:                 req.Logo,
			Organization:         req.Organization,
			RedirectUris:         req.RedirectUris,
			GrantTypes:           req.GrantTypes,
			Scopes:               req.Scopes,
			ClientId:             models.GenerateClientId(),
			ClientSecret:         models.GenerateClientSecret(),
			TokenFormat:          "JWT",
			ExpireInHours:        168, // 7 days
			RefreshExpireInHours: 720, // 30 days
			EnablePassword:       true,
			EnableSignUp:         true,
		}

		// 保存到数据库
		_, err = models.AddApplication(application)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("创建应用失败"))
		}

		return ctx.JSON(types.SuccessResponse(application))
	}
}

// HandleUpdateApplication 更新应用（需要管理员权限）
// Requirements: 8.7
func HandleUpdateApplication() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取参数
		owner := ctx.Params("owner")
		name := ctx.Params("name")

		if owner == "" || name == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("owner 和 name 参数不能为空"))
		}

		// 解析请求
		var req types.CreateApplicationRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 获取应用
		application, err := models.GetApplication(owner, name)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取应用信息失败"))
		}
		if application == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("应用不存在"))
		}

		// 更新应用信息
		if req.DisplayName != "" {
			application.DisplayName = req.DisplayName
		}
		if req.Logo != "" {
			application.Logo = req.Logo
		}
		if req.Organization != "" {
			application.Organization = req.Organization
		}
		if len(req.RedirectUris) > 0 {
			application.RedirectUris = req.RedirectUris
		}
		if len(req.GrantTypes) > 0 {
			application.GrantTypes = req.GrantTypes
		}
		if len(req.Scopes) > 0 {
			application.Scopes = req.Scopes
		}

		// 保存到数据库
		_, err = models.UpdateApplication(owner, name, application)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("更新应用失败"))
		}

		return ctx.JSON(types.SuccessResponse(application))
	}
}

// HandleDeleteApplication 删除应用（需要管理员权限）
// Requirements: 8.8
func HandleDeleteApplication() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取参数
		owner := ctx.Params("owner")
		name := ctx.Params("name")

		if owner == "" || name == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("owner 和 name 参数不能为空"))
		}

		// 获取应用
		application, err := models.GetApplication(owner, name)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取应用信息失败"))
		}
		if application == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("应用不存在"))
		}

		// 删除应用
		_, err = models.DeleteApplication(owner, name)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("删除应用失败"))
		}

		// TODO: 删除相关的 tokens

		return ctx.JSON(types.SuccessResponse(map[string]string{
			"message": "应用已删除",
		}))
	}
}

// HandleGetTokens 获取 Token 列表（需要管理员权限）
// Requirements: 8.9
func HandleGetTokens() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取 owner 参数（可选）
		owner := ctx.Query("owner", "")

		// 获取 Token 列表
		var tokens []*models.Token
		var err error

		if owner != "" {
			// 如果指定了 owner，只查询该 owner 的 Token
			err = models.GetEngine().Where("owner = ?", owner).Find(&tokens)
		} else {
			// 否则查询所有 Token
			err = models.GetEngine().Find(&tokens)
		}

		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取Token列表失败"))
		}

		// 构造响应数据（不返回实际的 token 值）
		tokenList := make([]map[string]interface{}, 0, len(tokens))
		for _, token := range tokens {
			tokenList = append(tokenList, map[string]interface{}{
				"owner":       token.Owner,
				"name":        token.Name,
				"createdTime": token.CreatedTime,
				"application": token.Application,
				"user":        token.User,
				"expiresAt":   token.ExpiresAt,
				"scope":       token.Scope,
				"tokenType":   token.TokenType,
				"isRevoked":   token.IsRevoked(),
			})
		}

		return ctx.JSON(types.SuccessResponse(tokenList))
	}
}

// HandleRevokeToken 撤销指定 Token（需要管理员权限）
// Requirements: 8.10
func HandleRevokeToken() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取参数
		owner := ctx.Params("owner")
		name := ctx.Params("name")

		if owner == "" || name == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("owner 和 name 参数不能为空"))
		}

		// 获取 Token
		token, err := models.GetToken(owner, name)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取Token信息失败"))
		}
		if token == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("Token不存在"))
		}

		// 撤销 Token（设置 ExpiresIn 为 0）
		token.ExpiresIn = 0

		// 保存到数据库
		_, err = models.UpdateToken(owner, name, token)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("撤销Token失败"))
		}

		return ctx.JSON(types.SuccessResponse(map[string]string{
			"message": "Token已撤销",
		}))
	}
}

// HandleRevokeUserTokens 撤销用户所有 Token（需要管理员权限）
// Requirements: 8.11
func HandleRevokeUserTokens() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取参数
		owner := ctx.Params("owner")
		username := ctx.Params("username")

		if owner == "" || username == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("owner 和 username 参数不能为空"))
		}

		// 获取用户
		user, err := models.GetUserByUsername(username)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 获取用户的所有 Token
		var tokens []*models.Token
		err = models.GetEngine().Where("owner = ? AND user = ?", owner, username).Find(&tokens)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户Token列表失败"))
		}

		// 撤销所有 Token
		revokedCount := 0
		for _, token := range tokens {
			token.ExpiresIn = 0
			_, err = models.UpdateToken(token.Owner, token.Name, token)
			if err == nil {
				revokedCount++
			}
		}

		return ctx.JSON(types.SuccessResponse(map[string]interface{}{
			"message": "用户Token已撤销",
			"count":   revokedCount,
		}))
	}
}

// HandleGetStats 获取系统统计（需要管理员权限）
// Requirements: 8.12
func HandleGetStats() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 统计用户数（排除已删除的用户）
		var userCount int64
		userCount, err := models.GetEngine().Where("is_deleted = ?", false).Count(&models.User{})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("统计用户数失败"))
		}

		// 统计应用数
		var appCount int64
		appCount, err = models.GetEngine().Count(&models.Application{})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("统计应用数失败"))
		}

		// 统计 Token 数
		var tokenCount int64
		tokenCount, err = models.GetEngine().Count(&models.Token{})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("统计Token数失败"))
		}

		// 统计活动 Token 数（未过期的）
		var activeTokenCount int64
		activeTokenCount, err = models.GetEngine().Where("expires_at > ?", time.Now().Unix()).Count(&models.Token{})
		if err != nil {
			// 如果查询失败，使用总数作为活动数
			activeTokenCount = tokenCount
		}

		// 构造响应
		stats := map[string]interface{}{
			"userCount":        userCount,
			"applicationCount": appCount,
			"tokenCount":       tokenCount,
			"activeTokenCount": activeTokenCount,
		}

		return ctx.JSON(types.SuccessResponse(stats))
	}
}

// HandleGetSystemInfo 获取系统信息（需要管理员权限）
// Requirements: 8.13
func HandleGetSystemInfo() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 计算运行时间（秒）
		uptime := int64(time.Since(services.ServerStartTime).Seconds())

		// 检查 Redis 连接状态
		redisConnected := false
		if err := services.PingRedis(); err == nil {
			redisConnected = true
		}

		// 获取系统信息
		systemInfo := map[string]interface{}{
			"version":        "1.0.0",
			"uptime":         uptime,
			"redisConnected": redisConnected,
		}

		return ctx.JSON(types.SuccessResponse(systemInfo))
	}
}

// HandleClearCache 清除缓存（需要管理员权限）
// Requirements: 8.14
func HandleClearCache() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 清除 Redis 缓存
		err := services.ClearCache()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("清除缓存失败"))
		}

		return ctx.JSON(types.SuccessResponse(map[string]string{
			"message": "缓存已清除",
		}))
	}
}

// HandleBanUser 封禁用户（需要管理员权限）
func HandleBanUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取用户 ID
		idStr := ctx.Params("id")
		if idStr == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户ID不能为空"))
		}

		// 转换为 int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 获取用户
		user, err := models.GetUserById(id)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil || user.IsDeleted {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 不能封禁管理员
		if user.IsAdmin {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("不能封禁管理员账户"))
		}

		// 设置封禁状态
		user.IsForbidden = true
		user.UpdatedTime = time.Now().Format("2006-01-02T15:04:05Z07:00")

		// 保存到数据库
		_, err = models.UpdateUser(id, user)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("封禁用户失败"))
		}

		return ctx.JSON(types.SuccessResponse(map[string]interface{}{
			"message": "用户已封禁",
			"userId":  id,
		}))
	}
}

// HandleUnbanUser 解封用户（需要管理员权限）
func HandleUnbanUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 获取用户 ID
		idStr := ctx.Params("id")
		if idStr == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户ID不能为空"))
		}

		// 转换为 int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 获取用户
		user, err := models.GetUserById(id)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil || user.IsDeleted {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 设置解封状态
		user.IsForbidden = false
		user.UpdatedTime = time.Now().Format("2006-01-02T15:04:05Z07:00")

		// 保存到数据库
		_, err = models.UpdateUser(id, user)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("解封用户失败"))
		}

		return ctx.JSON(types.SuccessResponse(map[string]interface{}{
			"message": "用户已解封",
			"userId":  id,
		}))
	}
}
