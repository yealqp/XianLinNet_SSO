// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
	"github.com/oauth-server/oauth-server/types"
)

// HandleLogin 处理用户登录请求
// Requirements: 3.1, 3.2, 3.3, 3.9, 3.10
func HandleLogin() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析 LoginRequest
		var req types.LoginRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证参数（email, password 非空）
		if req.Email == "" || req.Password == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("邮箱和密码不能为空"))
		}

		// 验证验证码（如果启用）
		if req.CaptchaToken != "" {
			valid, err := services.VerifyCaptcha(req.CaptchaToken)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("验证码验证失败"))
			}
			if !valid {
				return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("验证码无效"))
			}
		}

		// 调用 AuthService.Login()
		user, err := services.LoginUser(req.Email, req.Password)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse(err.Error()))
		}

		// 获取默认应用以生成 token
		// 使用内置应用 "admin/app-built-in"
		application, err := models.GetApplication("admin", "app-built-in")
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取应用信息失败"))
		}
		if application == nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("默认应用不存在"))
		}

		// 生成 JWT tokens
		accessToken, refreshToken, _, err := services.GenerateJwtToken(application, user, "openid profile email", "", "")
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("生成令牌失败"))
		}

		// 构造 LoginResponse
		loginResp := types.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int(application.ExpireInHours * 3600),
			TokenType:    "Bearer",
			User: types.UserInfo{
				ID:          user.Id,
				Email:       user.Email,
				Username:    user.Username,
				IsAdmin:     user.IsAdmin,
				IsRealName:  user.IsRealName,
				IsForbidden: user.IsForbidden,
				QQ:          user.QQ,
				Avatar:      user.Avatar,
			},
		}

		// 返回 LoginResponse
		return ctx.JSON(types.SuccessResponse(loginResp))
	}
}

// HandleRegister 处理用户注册请求
// Requirements: 3.4, 3.5
func HandleRegister() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析 RegisterRequest
		var req types.RegisterRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证参数
		if req.Username == "" || req.Email == "" || req.Password == "" || req.VerificationCode == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户名、邮箱、密码和验证码不能为空"))
		}

		// 验证用户名长度
		if len(req.Username) < 3 || len(req.Username) > 50 {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("用户名长度必须在3-50个字符之间"))
		}

		// 验证密码长度
		if len(req.Password) < 6 {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("密码长度至少为6个字符"))
		}

		// 验证邮箱验证码
		valid, err := services.VerifyCode(req.Email, req.VerificationCode, "register")
		if err != nil || !valid {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("验证码无效或已过期"))
		}

		// 注册用户
		user, err := services.RegisterUser(req.Email, req.Password, req.Username)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse(err.Error()))
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

// SendVerificationCodeRequest 发送验证码请求
type SendVerificationCodeRequest struct {
	Email        string `json:"email"`
	Purpose      string `json:"purpose"`      // "register" or "reset_password"
	CaptchaToken string `json:"captchaToken"` // 可选的验证码
}

// HandleSendVerificationCode 处理发送验证码请求
// Requirements: 3.6
func HandleSendVerificationCode() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析请求
		var req SendVerificationCodeRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证参数
		if req.Email == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("邮箱不能为空"))
		}

		if req.Purpose != "register" && req.Purpose != "reset_password" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的验证码用途"))
		}

		// 验证 Captcha（如果启用）
		if req.CaptchaToken != "" {
			valid, err := services.VerifyCaptcha(req.CaptchaToken)
			if err != nil || !valid {
				return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("验证码验证失败"))
			}
		}

		// 异步发送验证码邮件（不阻塞请求）
		go func() {
			code, err := services.SendVerificationEmail(req.Email, req.Purpose)
			if err != nil {
				log.Printf("Failed to send verification email to %s: %v", req.Email, err)
			} else if code != "" {
				// 开发环境下记录验证码
				log.Printf("Verification code sent to %s: %s", req.Email, code)
			}
		}()

		// 立即返回成功响应
		return ctx.JSON(types.SuccessResponse(map[string]string{
			"message": "验证码已发送，请查收邮件",
		}))
	}
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

// HandleResetPassword 处理重置密码请求
// Requirements: 3.7
func HandleResetPassword() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析请求
		var req ResetPasswordRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 验证参数
		if req.Email == "" || req.Code == "" || req.NewPassword == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("邮箱、验证码和新密码不能为空"))
		}

		// 验证密码长度
		if len(req.NewPassword) < 6 {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("密码长度至少为6个字符"))
		}

		// 验证验证码
		valid, err := services.VerifyCode(req.Email, req.Code, "reset_password")
		if err != nil || !valid {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("验证码无效或已过期"))
		}

		// 重置密码
		err = services.ResetPassword(req.Email, req.NewPassword)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse(err.Error()))
		}

		// 返回成功
		return ctx.JSON(types.SuccessResponse(map[string]string{
			"message": "密码重置成功",
		}))
	}
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Username string `json:"username,omitempty"`
	QQ       string `json:"qq,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// HandleUpdateProfile 处理更新个人资料请求
// Requirements: 3.8
func HandleUpdateProfile() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 需要 JWT 认证，从 context 获取用户 ID
		userID, ok := ctx.Locals("userID").(string)
		if !ok || userID == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("未授权"))
		}

		// 解析更新请求
		var req UpdateProfileRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的请求数据"))
		}

		// 转换 userID 为 int64
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("无效的用户ID"))
		}

		// 获取用户
		user, err := models.GetUserById(userIDInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取用户信息失败"))
		}
		if user == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("用户不存在"))
		}

		// 更新用户信息
		if req.Username != "" {
			user.Username = req.Username
		}
		if req.QQ != "" {
			user.QQ = req.QQ
		}
		if req.Avatar != "" {
			user.Avatar = req.Avatar
		}

		// 调用 AuthService.UpdateProfile()
		_, err = models.UpdateUser(userIDInt, user)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("更新用户信息失败"))
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

// HandleGetApplicationInfo 处理获取应用信息请求
// Requirements: 5.7
func HandleGetApplicationInfo() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 解析 client_id 参数
		clientID := ctx.Query("client_id")
		if clientID == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("client_id 参数不能为空"))
		}

		// 调用 AuthService.GetApplicationInfo()
		application, err := models.GetApplicationByClientId(clientID)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("获取应用信息失败"))
		}
		if application == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(types.ErrorResponse("应用不存在"))
		}

		// 返回应用名称、Logo、组织信息
		appInfo := map[string]interface{}{
			"name":         application.Name,
			"displayName":  application.DisplayName,
			"logo":         application.Logo,
			"organization": application.Organization,
			"homepageUrl":  application.HomepageUrl,
			"description":  application.Description,
		}

		return ctx.JSON(types.SuccessResponse(appInfo))
	}
}

// HandleAuthorize 处理 OAuth 授权请求
// Requirements: 5.1, 5.2, 5.3
func HandleAuthorize() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 需要 JWT 认证，从 context 获取用户 ID
		userID, ok := ctx.Locals("userID").(string)
		if !ok || userID == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse("未授权"))
		}

		// 获取请求参数
		clientID := ctx.Query("client_id")
		redirectURI := ctx.Query("redirect_uri")
		responseType := ctx.Query("response_type", "code")
		scope := ctx.Query("scope", "openid profile email")
		state := ctx.Query("state")
		nonce := ctx.Query("nonce")
		codeChallenge := ctx.Query("code_challenge")
		resource := ctx.Query("resource")

		// 验证 client_id 和 redirect_uri
		if clientID == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("client_id 参数不能为空"))
		}
		if redirectURI == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse("redirect_uri 参数不能为空"))
		}

		// 验证 OAuth 参数
		msg, application, err := services.CheckOAuthLogin(clientID, responseType, redirectURI, scope, state)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("验证失败"))
		}
		if msg != "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse(msg))
		}

		// 处理 GET 请求（显示授权页面信息）
		if ctx.Method() == "GET" {
			// 返回授权页面所需信息
			authInfo := map[string]interface{}{
				"clientId":    clientID,
				"redirectUri": redirectURI,
				"scope":       scope,
				"state":       state,
				"application": map[string]interface{}{
					"name":         application.Name,
					"displayName":  application.DisplayName,
					"logo":         application.Logo,
					"organization": application.Organization,
				},
			}
			return ctx.JSON(types.SuccessResponse(authInfo))
		}

		// 处理 POST 请求（用户同意授权）
		if ctx.Method() == "POST" {
			// 生成授权码
			codeResp, err := services.GetOAuthCode(userID, clientID, responseType, redirectURI, scope, state, nonce, codeChallenge, resource)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse("生成授权码失败"))
			}

			if codeResp.Message != "" {
				return ctx.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse(codeResp.Message))
			}

			// 构建重定向 URL
			redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, codeResp.Code)
			if state != "" {
				redirectURL += fmt.Sprintf("&state=%s", state)
			}

			// 返回 JSON 响应，包含重定向 URL 和授权码
			return ctx.JSON(types.SuccessResponse(map[string]interface{}{
				"code":         codeResp.Code,
				"redirect_uri": redirectURL,
				"state":        state,
			}))
		}

		return ctx.Status(fiber.StatusMethodNotAllowed).JSON(types.ErrorResponse("不支持的请求方法"))
	}
}
