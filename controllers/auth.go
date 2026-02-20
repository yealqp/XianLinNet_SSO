// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package controllers

import (
	"fmt"
	"time"

	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
)

type AuthController struct {
	BaseController
}

// GetApplicationInfo returns public application information by client_id
// @router /api/auth/application-info [get]
func (c *AuthController) GetApplicationInfo() {
	clientId := c.GetString("client_id")

	if clientId == "" {
		c.ResponseError("client_id is required")
		return
	}

	// Get application by client_id
	app, err := models.GetApplicationByClientId(clientId)
	if err != nil {
		c.ResponseError("Failed to get application")
		return
	}

	if app == nil {
		c.ResponseError("Application not found")
		return
	}

	// Return only public information (don't expose client_secret)
	c.ResponseOk(map[string]interface{}{
		"name":         app.Name,
		"displayName":  app.DisplayName,
		"logo":         app.Logo,
		"description":  app.Description,
		"homepageUrl":  app.HomepageUrl,
		"organization": app.Organization,
	})
}

// Authorize handles OAuth authorization requests
// @router /oauth/authorize [get]
func (c *AuthController) Authorize() {
	clientId := c.GetString("client_id")
	responseType := c.GetString("response_type")
	redirectUri := c.GetString("redirect_uri")
	scope := c.GetString("scope")
	state := c.GetString("state")
	nonce := c.GetString("nonce")
	codeChallenge := c.GetString("code_challenge")
	_ = c.GetString("code_challenge_method") // codeChallengeMethod - reserved for future use
	resource := c.GetString("resource")

	// Validate parameters
	msg, _, err := services.CheckOAuthLogin(clientId, responseType, redirectUri, scope, state)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if msg != "" {
		c.ResponseError(msg)
		return
	}

	// Get current user from session/token
	// Check Authorization header
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Authentication required")
		return
	}

	// Extract token
	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid authorization header")
		return
	}

	// Validate token and get user
	user, err := services.ValidateToken(token)
	if err != nil {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid token")
		return
	}

	// Convert user ID to string for OAuth code generation
	userId := fmt.Sprintf("%d", user.Id)

	code, err := services.GetOAuthCode(
		userId,
		clientId,
		responseType,
		redirectUri,
		scope,
		state,
		nonce,
		codeChallenge,
		resource,
	)

	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if code.Message != "" {
		c.ResponseError(code.Message)
		return
	}

	// Return authorization code and redirect URI
	c.ResponseOk(map[string]interface{}{
		"code":         code.Code,
		"state":        state,
		"redirect_uri": fmt.Sprintf("%s?code=%s&state=%s", redirectUri, code.Code, state),
	})
}

// Login handles user login with email and password
// @router /api/auth/login [post]
func (c *AuthController) Login() {
	var loginReq struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		CaptchaToken string `json:"captchaToken"`
	}

	err := c.GetRequestBody(&loginReq)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	// Validate input
	if loginReq.Email == "" || loginReq.Password == "" {
		c.ResponseError("Email and password are required")
		return
	}

	// Verify captcha token
	if loginReq.CaptchaToken == "" {
		c.ResponseError("Captcha token is required")
		return
	}

	captchaValid, err := services.VerifyCaptcha(loginReq.CaptchaToken)
	if err != nil {
		fmt.Printf("Captcha verification error: %v\n", err)
		c.ResponseError("Captcha verification failed")
		return
	}
	if !captchaValid {
		c.ResponseError("Invalid captcha")
		return
	}

	// Authenticate user
	user, err := services.LoginUser(loginReq.Email, loginReq.Password)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Get default application for token generation
	app, err := models.GetApplication("admin", "app-built-in")
	if err != nil || app == nil {
		c.ResponseError("Application not found")
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, _, err := services.GenerateJwtToken(
		app,
		user,
		"openid profile email",
		"",
		"",
	)
	if err != nil {
		c.ResponseError("Failed to generate token")
		return
	}

	// Return tokens and user info
	c.ResponseOk(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    app.ExpireInHours * 3600,
		"user": map[string]interface{}{
			"id":         user.Id,
			"email":      user.Email,
			"username":   user.Username,
			"isAdmin":    user.IsAdmin,
			"isRealName": user.IsRealName,
			"qq":         user.QQ,
			"avatar":     user.Avatar,
		},
	})
}

// Register handles user registration
// @router /api/auth/register [post]
func (c *AuthController) Register() {
	var registerReq struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		Username         string `json:"username"`
		VerificationCode string `json:"verificationCode"`
	}

	err := c.GetRequestBody(&registerReq)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	// Validate input
	if registerReq.Email == "" {
		c.ResponseError("Email is required")
		return
	}
	if registerReq.Password == "" {
		c.ResponseError("Password is required")
		return
	}
	if len(registerReq.Password) < 6 {
		c.ResponseError("Password must be at least 6 characters")
		return
	}
	if registerReq.VerificationCode == "" {
		c.ResponseError("Verification code is required")
		return
	}

	// Verify email code
	valid, err := services.VerifyCode(registerReq.Email, registerReq.VerificationCode, "register")
	if err != nil || !valid {
		c.ResponseError("Invalid or expired verification code")
		return
	}

	// Set default username if not provided
	if registerReq.Username == "" {
		registerReq.Username = registerReq.Email
	}

	// Register user
	user, err := services.RegisterUser(registerReq.Email, registerReq.Password, registerReq.Username)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(map[string]interface{}{
		"id":       user.Id,
		"email":    user.Email,
		"username": user.Username,
	})
}

// SendVerificationCode sends verification code to email
// @router /api/auth/send-code [post]
func (c *AuthController) SendVerificationCode() {
	var req struct {
		Email        string `json:"email"`
		Purpose      string `json:"purpose"` // "register" or "reset_password"
		CaptchaToken string `json:"captchaToken"`
	}

	// Debug: log raw request body
	fmt.Printf("Raw request body: %s\n", string(c.Ctx.Input.RequestBody))

	err := c.GetRequestBody(&req)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		c.ResponseError("Invalid request body")
		return
	}

	fmt.Printf("Parsed request - Email: %s, Purpose: %s\n", req.Email, req.Purpose)

	// Verify captcha token
	if req.CaptchaToken == "" {
		c.ResponseError("Captcha token is required")
		return
	}

	captchaValid, err := services.VerifyCaptcha(req.CaptchaToken)
	if err != nil {
		fmt.Printf("Captcha verification error: %v\n", err)
		c.ResponseError("Captcha verification failed")
		return
	}
	if !captchaValid {
		c.ResponseError("Invalid captcha")
		return
	}

	// Validate input
	if req.Email == "" {
		c.ResponseError("Email is required")
		return
	}
	if req.Purpose != "register" && req.Purpose != "reset_password" {
		c.ResponseError("Invalid purpose")
		return
	}

	// For registration, check if email already exists
	if req.Purpose == "register" {
		existingUser, err := models.GetUserByEmail(req.Email)
		if err != nil {
			c.ResponseError("Failed to check email")
			return
		}
		if existingUser != nil {
			c.ResponseError("Email already registered")
			return
		}
	}

	// For password reset, check if email exists
	if req.Purpose == "reset_password" {
		existingUser, err := models.GetUserByEmail(req.Email)
		if err != nil {
			c.ResponseError("Failed to check email")
			return
		}
		if existingUser == nil {
			c.ResponseError("Email not found")
			return
		}
	}

	// Send verification code
	code, err := services.SendVerificationEmail(req.Email, req.Purpose)
	if err != nil {
		c.ResponseError("Failed to send verification code")
		return
	}

	// In development, return the code for testing
	// In production, don't return the code
	c.ResponseOk(map[string]interface{}{
		"message": "Verification code sent to your email",
		"code":    code, // Remove this in production
	})
}

// ResetPassword handles password reset
// @router /api/auth/reset-password [post]
func (c *AuthController) ResetPassword() {
	var req struct {
		Email            string `json:"email"`
		VerificationCode string `json:"verificationCode"`
		NewPassword      string `json:"newPassword"`
	}

	err := c.GetRequestBody(&req)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" {
		c.ResponseError("Email is required")
		return
	}
	if req.VerificationCode == "" {
		c.ResponseError("Verification code is required")
		return
	}
	if req.NewPassword == "" {
		c.ResponseError("New password is required")
		return
	}
	if len(req.NewPassword) < 6 {
		c.ResponseError("Password must be at least 6 characters")
		return
	}

	// Verify email code
	valid, err := services.VerifyCode(req.Email, req.VerificationCode, "reset_password")
	if err != nil || !valid {
		c.ResponseError("Invalid or expired verification code")
		return
	}

	// Reset password
	err = services.ResetPassword(req.Email, req.NewPassword)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(map[string]interface{}{
		"message": "Password reset successfully",
	})
}

// UpdateProfile 更新用户个人资料
func (c *AuthController) UpdateProfile() {
	var req struct {
		UserId   int64  `json:"userId"`
		Username string `json:"username"`
		QQ       string `json:"qq"`
		Avatar   string `json:"avatar"`
	}

	err := c.GetRequestBody(&req)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	if req.UserId == 0 {
		c.ResponseError("用户ID不能为空")
		return
	}

	// 获取用户
	user, err := models.GetUserById(req.UserId)
	if err != nil || user == nil {
		c.ResponseError("用户不存在")
		return
	}

	// 更新字段
	if req.Username != "" {
		user.Username = req.Username
	}
	user.QQ = req.QQ
	user.Avatar = req.Avatar
	user.UpdatedTime = time.Now().Format(time.RFC3339)

	// 保存到数据库
	_, err = models.UpdateUser(req.UserId, user)
	if err != nil {
		c.ResponseError("更新失败: " + err.Error())
		return
	}

	c.ResponseOk(map[string]interface{}{
		"message": "个人资料更新成功",
		"user": map[string]interface{}{
			"id":         user.Id,
			"username":   user.Username,
			"email":      user.Email,
			"qq":         user.QQ,
			"avatar":     user.Avatar,
			"isRealName": user.IsRealName,
		},
	})
}
