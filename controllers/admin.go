// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package controllers

import (
	"context"
	"strconv"
	"time"

	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
)

var serverStartTime = time.Now()

type AdminController struct {
	BaseController
}

// Middleware to check admin authentication
func (c *AdminController) checkAdmin() bool {
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Missing authorization header")
		return false
	}

	// Extract token
	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid authorization header")
		return false
	}

	// Validate token
	user, err := services.ValidateToken(token)
	if err != nil {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid token")
		return false
	}

	if !user.IsAdmin {
		c.Ctx.Output.SetStatus(403)
		c.ResponseError("Admin access required")
		return false
	}

	return true
}

// ==================== User Management ====================

// GetUsers lists all users
// @router /api/admin/users [get]
func (c *AdminController) GetUsers() {
	if !c.checkAdmin() {
		return
	}

	users := []models.User{}
	err := models.GetEngine().Find(&users)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Mask sensitive data
	for i := range users {
		users[i].Password = "***"
	}

	c.ResponseOk(users)
}

// GetUser gets a specific user
// @router /api/admin/users/:id [get]
func (c *AdminController) GetUser() {
	if !c.checkAdmin() {
		return
	}

	idStr := c.GetString(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.ResponseError("Invalid user ID")
		return
	}

	user, err := models.GetUserById(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if user == nil {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("User not found")
		return
	}

	user.Password = "***"
	c.ResponseOk(user)
}

// CreateUser creates a new user
// @router /api/admin/users [post]
func (c *AdminController) CreateUser() {
	if !c.checkAdmin() {
		return
	}

	var user models.User
	err := c.GetRequestBody(&user)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	// Set defaults
	if user.Owner == "" {
		user.Owner = "built-in"
	}
	if user.CreatedTime == "" {
		user.CreatedTime = time.Now().Format(time.RFC3339)
	}
	// Id is auto-increment, no need to set it manually

	// TODO: Hash password

	affected, err := models.AddUser(&user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.ResponseError("Failed to create user")
		return
	}

	// Invalidate cache
	services.InvalidateUserCache(user.GetId())

	c.ResponseOk(user)
}

// UpdateUser updates a user
// @router /api/admin/users/:id [put]
func (c *AdminController) UpdateUser() {
	if !c.checkAdmin() {
		return
	}

	idStr := c.GetString(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.ResponseError("Invalid user ID")
		return
	}

	var user models.User
	err = c.GetRequestBody(&user)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	user.Id = id
	affected, err := models.UpdateUser(id, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("User not found")
		return
	}

	// Invalidate cache
	services.InvalidateUserCache(user.GetId())

	c.ResponseOk(user)
}

// DeleteUser deletes a user
// @router /api/admin/users/:id [delete]
func (c *AdminController) DeleteUser() {
	if !c.checkAdmin() {
		return
	}

	idStr := c.GetString(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.ResponseError("Invalid user ID")
		return
	}

	// Get user first for cache invalidation
	user, err := models.GetUserById(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if user == nil {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("User not found")
		return
	}

	affected, err := models.DeleteUser(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("User not found")
		return
	}

	// Invalidate cache
	services.InvalidateUserCache(user.GetId())

	c.ResponseOk()
}

// ==================== Application Management ====================

// GetApplications lists all applications
// @router /api/admin/applications [get]
func (c *AdminController) GetApplications() {
	if !c.checkAdmin() {
		return
	}

	owner := c.GetString("owner", "admin")

	apps := []models.Application{}
	err := models.GetEngine().Find(&apps, &models.Application{Owner: owner})
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Ensure array fields are not nil
	for i := range apps {
		if apps[i].RedirectUris == nil {
			apps[i].RedirectUris = []string{}
		}
		if apps[i].GrantTypes == nil {
			apps[i].GrantTypes = []string{}
		}
		if apps[i].Scopes == nil {
			apps[i].Scopes = []string{}
		}
		if apps[i].Tags == nil {
			apps[i].Tags = []string{}
		}
	}

	c.ResponseOk(apps)
}

// GetApplication gets a specific application
// @router /api/admin/applications/:owner/:name [get]
func (c *AdminController) GetApplication() {
	if !c.checkAdmin() {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	app, err := models.GetApplication(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if app == nil {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Application not found")
		return
	}

	// Ensure array fields are not nil
	if app.RedirectUris == nil {
		app.RedirectUris = []string{}
	}
	if app.GrantTypes == nil {
		app.GrantTypes = []string{}
	}
	if app.Scopes == nil {
		app.Scopes = []string{}
	}
	if app.Tags == nil {
		app.Tags = []string{}
	}

	c.ResponseOk(app)
}

// CreateApplication creates a new application
// @router /api/admin/applications [post]
func (c *AdminController) CreateApplication() {
	if !c.checkAdmin() {
		return
	}

	var app models.Application
	err := c.GetRequestBody(&app)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	// Set defaults
	if app.Owner == "" {
		app.Owner = "admin"
	}
	if app.Organization == "" {
		app.Organization = "built-in"
	}
	if app.CreatedTime == "" {
		app.CreatedTime = time.Now().Format(time.RFC3339)
	}
	if app.ExpireInHours == 0 {
		app.ExpireInHours = 168 // 7 days
	}
	if app.RefreshExpireInHours == 0 {
		app.RefreshExpireInHours = 720 // 30 days
	}
	if app.TokenFormat == "" {
		app.TokenFormat = "JWT"
	}
	if len(app.GrantTypes) == 0 {
		app.GrantTypes = []string{"authorization_code", "password", "client_credentials", "refresh_token"}
	}
	if len(app.Scopes) == 0 {
		app.Scopes = []string{"openid", "profile", "email"}
	}

	affected, err := models.AddApplication(&app)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.ResponseError("Failed to create application")
		return
	}

	c.ResponseOk(app)
}

// UpdateApplication updates an application
// @router /api/admin/applications/:owner/:name [put]
func (c *AdminController) UpdateApplication() {
	if !c.checkAdmin() {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	// Get existing application
	existingApp, err := models.GetApplication(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if existingApp == nil {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Application not found")
		return
	}

	// Parse update request
	var updateData map[string]interface{}
	err = c.GetRequestBody(&updateData)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	// Update only provided fields
	if val, ok := updateData["name"]; ok {
		if strVal, ok := val.(string); ok && strVal != "" {
			existingApp.Name = strVal
		}
	}
	if val, ok := updateData["displayName"]; ok {
		if strVal, ok := val.(string); ok {
			existingApp.DisplayName = strVal
		}
	}
	if val, ok := updateData["logo"]; ok {
		if strVal, ok := val.(string); ok {
			existingApp.Logo = strVal
		}
	}
	if val, ok := updateData["redirectUris"]; ok {
		if arrVal, ok := val.([]interface{}); ok {
			uris := make([]string, 0, len(arrVal))
			for _, v := range arrVal {
				if strVal, ok := v.(string); ok {
					uris = append(uris, strVal)
				}
			}
			existingApp.RedirectUris = uris
		}
	}
	if val, ok := updateData["grantTypes"]; ok {
		if arrVal, ok := val.([]interface{}); ok {
			types := make([]string, 0, len(arrVal))
			for _, v := range arrVal {
				if strVal, ok := v.(string); ok {
					types = append(types, strVal)
				}
			}
			existingApp.GrantTypes = types
		}
	}
	if val, ok := updateData["scopes"]; ok {
		if arrVal, ok := val.([]interface{}); ok {
			scopes := make([]string, 0, len(arrVal))
			for _, v := range arrVal {
				if strVal, ok := v.(string); ok {
					scopes = append(scopes, strVal)
				}
			}
			existingApp.Scopes = scopes
		}
	}
	if val, ok := updateData["organization"]; ok {
		if strVal, ok := val.(string); ok {
			existingApp.Organization = strVal
		}
	}

	affected, err := models.UpdateApplication(owner, name, existingApp)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Application not found")
		return
	}

	c.ResponseOk(existingApp)
}

// DeleteApplication deletes an application
// @router /api/admin/applications/:owner/:name [delete]
func (c *AdminController) DeleteApplication() {
	if !c.checkAdmin() {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	affected, err := models.DeleteApplication(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Application not found")
		return
	}

	c.ResponseOk()
}

// ==================== Token Management ====================

// GetTokens lists all tokens
// @router /api/admin/tokens [get]
func (c *AdminController) GetTokens() {
	if !c.checkAdmin() {
		return
	}

	owner := c.GetString("owner", "admin")

	tokens := []models.Token{}
	err := models.GetEngine().Limit(100).Desc("created_time").Find(&tokens, &models.Token{Owner: owner})
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(tokens)
}

// RevokeToken revokes a token
// @router /api/admin/tokens/:owner/:name [delete]
func (c *AdminController) RevokeToken() {
	if !c.checkAdmin() {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	affected, err := models.DeleteToken(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Token not found")
		return
	}

	// Invalidate cache
	services.DeleteCachedToken(name)

	c.ResponseOk()
}

// RevokeUserTokens revokes all tokens for a user
// @router /api/admin/tokens/user/:owner/:username [delete]
func (c *AdminController) RevokeUserTokens() {
	if !c.checkAdmin() {
		return
	}

	owner := c.GetString(":owner")
	username := c.GetString(":username")

	affected, err := models.GetEngine().Where("organization = ? AND user = ?", owner, username).Delete(&models.Token{})
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(map[string]interface{}{
		"revoked": affected,
	})
}

// ==================== Statistics ====================

// GetStats returns system statistics
// @router /api/admin/stats [get]
func (c *AdminController) GetStats() {
	if !c.checkAdmin() {
		return
	}

	// Count users
	userCount, _ := models.GetEngine().Count(&models.User{})

	// Count applications
	appCount, _ := models.GetEngine().Count(&models.Application{})

	// Count tokens
	tokenCount, _ := models.GetEngine().Count(&models.Token{})

	// Count active tokens (not revoked and not expired)
	now := time.Now().Unix()
	activeTokenCount, _ := models.GetEngine().
		Where("expires_in > ?", 0).
		Where("expires_at = ? OR expires_at > ?", 0, now).
		Count(&models.Token{})

	stats := map[string]interface{}{
		"userCount":        userCount,
		"applicationCount": appCount,
		"tokenCount":       tokenCount,
		"activeTokenCount": activeTokenCount,
	}

	c.ResponseOk(stats)
}

// GetSystemInfo returns system information
// @router /api/admin/system [get]
func (c *AdminController) GetSystemInfo() {
	if !c.checkAdmin() {
		return
	}

	// Calculate uptime in seconds
	uptime := int64(time.Since(serverStartTime).Seconds())

	info := map[string]interface{}{
		"version":        "1.0.0",
		"uptime":         uptime,
		"redisConnected": false,
	}

	// Check Redis connection
	if redisClient := services.GetRedisClient(); redisClient != nil {
		_, err := redisClient.Ping(context.Background()).Result()
		info["redisConnected"] = err == nil
	}

	c.ResponseOk(info)
}

// ClearCache clears Redis cache
// @router /api/admin/cache/clear [post]
func (c *AdminController) ClearCache() {
	if !c.checkAdmin() {
		return
	}

	redisClient := services.GetRedisClient()
	if redisClient == nil {
		c.ResponseError("Redis not configured")
		return
	}

	// Clear all cache keys
	err := redisClient.FlushDB(context.Background()).Err()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(map[string]string{
		"message": "Cache cleared successfully",
	})
}
