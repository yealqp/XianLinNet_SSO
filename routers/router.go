// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package routers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/oauth-server/oauth-server/controllers"
)

func Init() {
	// Auth routes
	authCtrl := &controllers.AuthController{}
	web.Router("/oauth/authorize", authCtrl, "get:Authorize")
	web.Router("/api/login", authCtrl, "post:Login")

	// Token routes
	tokenCtrl := &controllers.TokenController{}
	web.Router("/api/oauth/token", tokenCtrl, "post:Token")
	web.Router("/api/login/oauth/access_token", tokenCtrl, "post:Token") // Alias for compatibility
	web.Router("/api/oauth/introspect", tokenCtrl, "post:Introspect")
	web.Router("/api/login/oauth/introspect", tokenCtrl, "post:Introspect") // Alias
	web.Router("/api/oauth/revoke", tokenCtrl, "post:Revoke")

	// OIDC routes
	oidcCtrl := &controllers.OidcController{}
	web.Router("/.well-known/openid-configuration", oidcCtrl, "get:Discovery")
	web.Router("/.well-known/jwks", oidcCtrl, "get:Jwks")
	web.Router("/api/userinfo", oidcCtrl, "get:UserInfo")
	web.Router("/api/oauth/register", oidcCtrl, "post:Register")

	// Health check
	web.Router("/health", &controllers.BaseController{}, "get:ResponseOk")

	// Admin routes
	adminCtrl := &controllers.AdminController{}

	// User management
	web.Router("/api/admin/users", adminCtrl, "get:GetUsers;post:CreateUser")
	web.Router("/api/admin/users/:owner/:name", adminCtrl, "get:GetUser;put:UpdateUser;delete:DeleteUser")

	// Application management
	web.Router("/api/admin/applications", adminCtrl, "get:GetApplications;post:CreateApplication")
	web.Router("/api/admin/applications/:owner/:name", adminCtrl, "get:GetApplication;put:UpdateApplication;delete:DeleteApplication")

	// Token management
	web.Router("/api/admin/tokens", adminCtrl, "get:GetTokens")
	web.Router("/api/admin/tokens/:owner/:name", adminCtrl, "delete:RevokeToken")
	web.Router("/api/admin/tokens/user/:owner/:username", adminCtrl, "delete:RevokeUserTokens")

	// Statistics and system
	web.Router("/api/admin/stats", adminCtrl, "get:GetStats")
	web.Router("/api/admin/system", adminCtrl, "get:GetSystemInfo")
	web.Router("/api/admin/cache/clear", adminCtrl, "post:ClearCache")

	// Permission management routes
	permCtrl := &controllers.PermissionController{}

	// Role management
	web.Router("/api/roles", permCtrl, "get:GetRoles;post:CreateRole")
	web.Router("/api/roles/:owner/:name", permCtrl, "get:GetRole;put:UpdateRole;delete:DeleteRole")
	web.Router("/api/roles/:owner/:name/permissions", permCtrl, "get:GetRolePermissions;post:AddRolePermission")
	web.Router("/api/roles/:owner/:name/permissions/:permOwner/:permName", permCtrl, "delete:RemoveRolePermission")

	// Permission management
	web.Router("/api/permissions", permCtrl, "get:GetPermissions;post:CreatePermission")
	web.Router("/api/permissions/:owner/:name", permCtrl, "get:GetPermission;put:UpdatePermission;delete:DeletePermission")

	// User role and permission management
	web.Router("/api/users/:owner/:name/roles", permCtrl, "get:GetUserRoles;post:AddUserRole")
	web.Router("/api/users/:owner/:name/roles/:roleOwner/:roleName", permCtrl, "delete:RemoveUserRole")
	web.Router("/api/users/:owner/:name/permissions", permCtrl, "get:GetUserPermissions")
}
