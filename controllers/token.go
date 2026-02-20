// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package controllers

import (
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
)

type TokenController struct {
	BaseController
}

// Token handles OAuth token requests
// @router /api/oauth/token [post]
func (c *TokenController) Token() {
	grantType := c.GetString("grant_type")
	clientId := c.GetString("client_id")
	clientSecret := c.GetString("client_secret")
	code := c.GetString("code")
	verifier := c.GetString("code_verifier")
	scope := c.GetString("scope")
	username := c.GetString("username")
	password := c.GetString("password")
	refreshToken := c.GetString("refresh_token")
	resource := c.GetString("resource")

	result, err := services.GetOAuthToken(
		grantType,
		clientId,
		clientSecret,
		code,
		verifier,
		scope,
		username,
		password,
		refreshToken,
		resource,
	)

	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = result
	c.ServeJSON()
}

// Introspect handles token introspection requests
// @router /api/oauth/introspect [post]
func (c *TokenController) Introspect() {
	token := c.GetString("token")
	_ = c.GetString("token_type_hint") // tokenTypeHint - reserved for future use

	if token == "" {
		c.ResponseError("token is required")
		return
	}

	// Parse and validate token
	claims, err := services.ParseJwtToken(token)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"active": false,
		}
		c.ServeJSON()
		return
	}

	// Return introspection response
	c.Data["json"] = map[string]interface{}{
		"active":     true,
		"scope":      claims.Scope,
		"client_id":  claims.Aud[0],
		"username":   claims.Username,
		"token_type": "Bearer",
		"exp":        claims.ExpiresAt.Unix(),
		"iat":        claims.IssuedAt.Unix(),
		"sub":        claims.Sub,
		"iss":        claims.Iss,
	}
	c.ServeJSON()
}

// Revoke handles token revocation requests
// @router /api/oauth/revoke [post]
func (c *TokenController) Revoke() {
	token := c.GetString("token")
	tokenTypeHint := c.GetString("token_type_hint")

	if token == "" {
		c.ResponseError("token is required")
		return
	}

	// Parse token to get information
	claims, err := services.ParseJwtToken(token)
	if err != nil {
		// RFC 7009: Even if token is invalid, return success
		c.ResponseOk()
		return
	}

	// Find and revoke the token
	var dbToken *models.Token
	if tokenTypeHint == "refresh_token" {
		dbToken, _ = models.GetTokenByRefreshToken(token)
	} else {
		// Try access token first
		dbToken, _ = models.GetTokenByAccessToken(token)
		// If not found, try refresh token
		if dbToken == nil {
			dbToken, _ = models.GetTokenByRefreshToken(token)
		}
	}

	if dbToken != nil {
		// Mark token as revoked by setting ExpiresIn to 0
		dbToken.ExpiresIn = 0
		models.UpdateToken(dbToken.Owner, dbToken.Name, dbToken)

		// Clear from cache if cached
		services.DeleteCachedToken(dbToken.AccessTokenHash)
	}

	// Always return success per RFC 7009
	c.ResponseOk(map[string]interface{}{
		"revoked": true,
		"jti":     claims.Id,
	})
}
