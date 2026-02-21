// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package controllers

import (
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/oauth-server/oauth-server/services"
)

type OidcController struct {
	BaseController
}

// Discovery handles OIDC discovery endpoint
// @router /.well-known/openid-configuration [get]
func (c *OidcController) Discovery() {
	origin, _ := web.AppConfig.String("origin")
	if origin == "" {
		origin = "http://localhost:8080"
	}

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
		"scopes_supported":                      []string{"openid", "profile", "email", "address", "offline_access"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
		"claims_supported":                      []string{"sub", "iss", "aud", "exp", "iat", "name", "email", "picture"},
		"code_challenge_methods_supported":      []string{"S256"},
	}

	c.Data["json"] = discovery
	c.ServeJSON()
}

// Jwks handles JWKS endpoint
// @router /.well-known/jwks [get]
func (c *OidcController) Jwks() {
	// 获取公钥的 JWK 表示
	jwk, err := services.GetPublicKeyJWK()
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.ResponseError("Failed to get public key")
		return
	}

	jwks := map[string]interface{}{
		"keys": []interface{}{jwk},
	}

	c.Data["json"] = jwks
	c.ServeJSON()
}

// UserInfo handles userinfo endpoint
// @router /api/userinfo [get]
func (c *OidcController) UserInfo() {
	// Get access token from Authorization header
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Missing authorization header")
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

	// Parse token to get claims (for scope checking)
	claims, err := services.ParseJwtToken(token)
	if err != nil {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid token claims")
		return
	}

	// Build user info based on requested scopes
	userInfo := map[string]interface{}{
		"sub": user.GetId(),
	}

	// Check scopes and add corresponding claims
	scopes := claims.Scope

	// Profile scope
	if contains(scopes, "profile") || contains(scopes, "openid") {
		userInfo["name"] = user.Username
		userInfo["preferred_username"] = user.Username
		userInfo["picture"] = user.Avatar
		userInfo["updated_at"] = time.Now().Unix()
	}

	// Email scope
	if contains(scopes, "email") || contains(scopes, "openid") {
		userInfo["email"] = user.Email
		userInfo["email_verified"] = true
	}

	// Custom claims
	userInfo["username"] = user.Username
	userInfo["qq"] = user.QQ
	userInfo["is_real_name"] = user.IsRealName
	userInfo["is_admin"] = user.IsAdmin

	c.Data["json"] = userInfo
	c.ServeJSON()
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)+1] == substr+" " || s[len(s)-len(substr)-1:] == " "+substr))
}

// Register handles dynamic client registration
// @router /api/oauth/register [post]
func (c *OidcController) Register() {
	var req struct {
		ClientName              string   `json:"client_name"`
		RedirectUris            []string `json:"redirect_uris"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
		Scope                   string   `json:"scope"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	}

	err := c.GetRequestBody(&req)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.ResponseError("Invalid request body")
		return
	}

	// Validate required fields
	if len(req.RedirectUris) == 0 {
		c.Ctx.Output.SetStatus(400)
		c.ResponseError("redirect_uris is required")
		return
	}

	// TODO: Implement actual DCR logic
	// For now, return a mock response
	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]interface{}{
		"client_id":     "generated-client-id",
		"client_secret": "generated-client-secret",
		"client_name":   req.ClientName,
		"redirect_uris": req.RedirectUris,
		"grant_types":   req.GrantTypes,
	}
	c.ServeJSON()
}
