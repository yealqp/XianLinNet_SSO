// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package controllers

import (
	"fmt"

	"github.com/oauth-server/oauth-server/services"
)

type AuthController struct {
	BaseController
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

	// For demo purposes, auto-approve for admin user
	// In production, this should show a consent page
	userId := "admin"

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

	// Redirect back to client with code
	redirectUrl := fmt.Sprintf("%s?code=%s&state=%s", redirectUri, code.Code, state)
	c.Redirect(redirectUrl, 302)
}

// Login handles user login
// @router /api/login [post]
func (c *AuthController) Login() {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := c.GetRequestBody(&loginReq)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	// TODO: Implement actual login logic
	// For now, just return success for demo
	c.ResponseOk(map[string]string{
		"userId": loginReq.Username,
	})
}
