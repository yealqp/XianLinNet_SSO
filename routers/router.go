// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package routers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/oauth-server/oauth-server/controllers"
	"github.com/oauth-server/oauth-server/middlewares"
)

func Init() {
	// 注册 CORS 中间件
	web.InsertFilter("*", web.BeforeRouter, middlewares.CORSFilter)

	// Auth routes
	authCtrl := &controllers.AuthController{}
	web.Router("/oauth/authorize", authCtrl, "get:Authorize")
	web.Router("/api/auth/application-info", authCtrl, "get:GetApplicationInfo")
	web.Router("/api/auth/login", authCtrl, "post:Login")
	web.Router("/api/auth/register", authCtrl, "post:Register")
	web.Router("/api/auth/send-code", authCtrl, "post:SendVerificationCode")
	web.Router("/api/auth/reset-password", authCtrl, "post:ResetPassword")
	web.Router("/api/auth/update-profile", authCtrl, "post:UpdateProfile")

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
	web.Router("/api/admin/users/:id", adminCtrl, "get:GetUser")
	web.Router("/api/admin/users/:id/update", adminCtrl, "post:UpdateUser")
	web.Router("/api/admin/users/:id/delete", adminCtrl, "post:DeleteUser")

	// Application management
	web.Router("/api/admin/applications", adminCtrl, "get:GetApplications;post:CreateApplication")
	web.Router("/api/admin/applications/:owner/:name", adminCtrl, "get:GetApplication")
	web.Router("/api/admin/applications/:owner/:name/update", adminCtrl, "post:UpdateApplication")
	web.Router("/api/admin/applications/:owner/:name/delete", adminCtrl, "post:DeleteApplication")

	// Token management
	web.Router("/api/admin/tokens", adminCtrl, "get:GetTokens")
	web.Router("/api/admin/tokens/:owner/:name/revoke", adminCtrl, "post:RevokeToken")
	web.Router("/api/admin/tokens/user/:owner/:username/revoke", adminCtrl, "post:RevokeUserTokens")

	// Statistics and system
	web.Router("/api/admin/stats", adminCtrl, "get:GetStats")
	web.Router("/api/admin/system", adminCtrl, "get:GetSystemInfo")
	web.Router("/api/admin/cache/clear", adminCtrl, "post:ClearCache")

	// Permission management routes
	permCtrl := &controllers.PermissionController{}

	// Role management
	web.Router("/api/roles", permCtrl, "get:GetRoles;post:CreateRole")
	web.Router("/api/roles/:owner/:name", permCtrl, "get:GetRole")
	web.Router("/api/roles/:owner/:name/update", permCtrl, "post:UpdateRole")
	web.Router("/api/roles/:owner/:name/delete", permCtrl, "post:DeleteRole")
	web.Router("/api/roles/:owner/:name/permissions", permCtrl, "get:GetRolePermissions;post:AddRolePermission")
	web.Router("/api/roles/:owner/:name/permissions/remove", permCtrl, "post:RemoveRolePermission")

	// Permission management
	web.Router("/api/permissions", permCtrl, "get:GetPermissions;post:CreatePermission")
	web.Router("/api/permissions/:owner/:name", permCtrl, "get:GetPermission")
	web.Router("/api/permissions/:owner/:name/update", permCtrl, "post:UpdatePermission")
	web.Router("/api/permissions/:owner/:name/delete", permCtrl, "post:DeletePermission")

	// User role and permission management
	web.Router("/api/users/:id/roles", permCtrl, "get:GetUserRoles;post:AddUserRole")
	web.Router("/api/users/:id/roles/remove", permCtrl, "post:RemoveUserRole")
	web.Router("/api/users/:id/permissions", permCtrl, "get:GetUserPermissions")

	// Real name verification routes
	realnameCtrl := &controllers.RealNameController{}
	web.Router("/api/realname/verify", realnameCtrl, "post:VerifyRealName")
	web.Router("/api/realname/submit", realnameCtrl, "post:SubmitRealName")
	web.Router("/api/admin/realname/:userId", realnameCtrl, "get:GetRealNameInfo")
}
