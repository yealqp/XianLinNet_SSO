// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package controllers

import (
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/services"
)

type PermissionController struct {
	BaseController
}

// checkPermission checks if the current user has the required permission
func (c *PermissionController) checkPermission(resource, action string) bool {
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Missing authorization header")
		return false
	}

	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid authorization header")
		return false
	}

	user, err := services.ValidateToken(token)
	if err != nil {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("Invalid token")
		return false
	}

	// Check permission
	hasPermission, err := services.CheckPermission(user.Owner, user.Name, resource, action)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.ResponseError(err.Error())
		return false
	}

	if !hasPermission {
		c.Ctx.Output.SetStatus(403)
		c.ResponseError("Insufficient permissions")
		return false
	}

	return true
}

// ==================== Role Management ====================

// GetRoles lists all roles
// @router /api/roles [get]
func (c *PermissionController) GetRoles() {
	if !c.checkPermission("role", "read") {
		return
	}

	owner := c.GetString("owner", "admin")
	organization := c.GetString("organization")

	var roles []*models.Role
	var err error

	if organization != "" {
		roles, err = models.GetRolesByOrganization(organization)
	} else {
		err = models.GetEngine().Where("owner = ?", owner).Find(&roles)
	}

	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(roles)
}

// GetRole gets a specific role
// @router /api/roles/:owner/:name [get]
func (c *PermissionController) GetRole() {
	if !c.checkPermission("role", "read") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	role, err := models.GetRole(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if role == nil {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Role not found")
		return
	}

	c.ResponseOk(role)
}

// CreateRole creates a new role
// @router /api/roles [post]
func (c *PermissionController) CreateRole() {
	if !c.checkPermission("role", "write") {
		return
	}

	var role models.Role
	err := c.GetRequestBody(&role)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	if role.Owner == "" {
		role.Owner = "admin"
	}
	if role.CreatedTime == "" {
		role.CreatedTime = models.GetCurrentTime()
	}
	if role.Type == "" {
		role.Type = "custom"
	}

	affected, err := models.AddRole(&role)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.ResponseError("Failed to create role")
		return
	}

	c.ResponseOk(role)
}

// UpdateRole updates a role
// @router /api/roles/:owner/:name [put]
func (c *PermissionController) UpdateRole() {
	if !c.checkPermission("role", "write") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	var role models.Role
	err := c.GetRequestBody(&role)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	role.UpdatedTime = models.GetCurrentTime()

	affected, err := models.UpdateRole(owner, name, &role)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Role not found")
		return
	}

	c.ResponseOk(role)
}

// DeleteRole deletes a role
// @router /api/roles/:owner/:name [delete]
func (c *PermissionController) DeleteRole() {
	if !c.checkPermission("role", "write") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	// Prevent deletion of system roles
	role, _ := models.GetRole(owner, name)
	if role != nil && role.Type == "system" {
		c.ResponseError("Cannot delete system roles")
		return
	}

	affected, err := models.DeleteRole(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Role not found")
		return
	}

	c.ResponseOk()
}

// ==================== Permission Management ====================

// GetPermissions lists all permissions
// @router /api/permissions [get]
func (c *PermissionController) GetPermissions() {
	if !c.checkPermission("permission", "read") {
		return
	}

	owner := c.GetString("owner", "admin")

	perms, err := models.GetAllPermissions(owner)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(perms)
}

// GetPermission gets a specific permission
// @router /api/permissions/:owner/:name [get]
func (c *PermissionController) GetPermission() {
	if !c.checkPermission("permission", "read") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	perm, err := models.GetPermission(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if perm == nil {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Permission not found")
		return
	}

	c.ResponseOk(perm)
}

// CreatePermission creates a new permission
// @router /api/permissions [post]
func (c *PermissionController) CreatePermission() {
	if !c.checkPermission("permission", "write") {
		return
	}

	var perm models.Permission
	err := c.GetRequestBody(&perm)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	if perm.Owner == "" {
		perm.Owner = "admin"
	}
	if perm.CreatedTime == "" {
		perm.CreatedTime = models.GetCurrentTime()
	}
	if perm.Effect == "" {
		perm.Effect = "allow"
	}

	affected, err := models.AddPermission(&perm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.ResponseError("Failed to create permission")
		return
	}

	c.ResponseOk(perm)
}

// UpdatePermission updates a permission
// @router /api/permissions/:owner/:name [put]
func (c *PermissionController) UpdatePermission() {
	if !c.checkPermission("permission", "write") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	var perm models.Permission
	err := c.GetRequestBody(&perm)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	affected, err := models.UpdatePermission(owner, name, &perm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Permission not found")
		return
	}

	c.ResponseOk(perm)
}

// DeletePermission deletes a permission
// @router /api/permissions/:owner/:name [delete]
func (c *PermissionController) DeletePermission() {
	if !c.checkPermission("permission", "write") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	affected, err := models.DeletePermission(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Permission not found")
		return
	}

	c.ResponseOk()
}

// ==================== Role-Permission Assignment ====================

// GetRolePermissions gets all permissions for a role
// @router /api/roles/:owner/:name/permissions [get]
func (c *PermissionController) GetRolePermissions() {
	if !c.checkPermission("role", "read") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	perms, err := models.GetRolePermissions(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(perms)
}

// AddRolePermission assigns a permission to a role
// @router /api/roles/:owner/:name/permissions [post]
func (c *PermissionController) AddRolePermission() {
	if !c.checkPermission("role", "write") {
		return
	}

	roleOwner := c.GetString(":owner")
	roleName := c.GetString(":name")

	var req struct {
		PermOwner string `json:"permOwner"`
		PermName  string `json:"permName"`
	}

	err := c.GetRequestBody(&req)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	affected, err := models.AddRolePermission(roleOwner, roleName, req.PermOwner, req.PermName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.ResponseError("Failed to add permission to role")
		return
	}

	c.ResponseOk()
}

// RemoveRolePermission removes a permission from a role
// @router /api/roles/:owner/:name/permissions/:permOwner/:permName [delete]
func (c *PermissionController) RemoveRolePermission() {
	if !c.checkPermission("role", "write") {
		return
	}

	roleOwner := c.GetString(":owner")
	roleName := c.GetString(":name")
	permOwner := c.GetString(":permOwner")
	permName := c.GetString(":permName")

	affected, err := models.RemoveRolePermission(roleOwner, roleName, permOwner, permName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("Role permission not found")
		return
	}

	c.ResponseOk()
}

// ==================== User-Role Assignment ====================

// GetUserRoles gets all roles for a user
// @router /api/users/:owner/:name/roles [get]
func (c *PermissionController) GetUserRoles() {
	if !c.checkPermission("user", "read") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	roles, err := models.GetUserRoles(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(roles)
}

// AddUserRole assigns a role to a user
// @router /api/users/:owner/:name/roles [post]
func (c *PermissionController) AddUserRole() {
	if !c.checkPermission("user", "write") {
		return
	}

	userOwner := c.GetString(":owner")
	userName := c.GetString(":name")

	var req struct {
		RoleOwner string `json:"roleOwner"`
		RoleName  string `json:"roleName"`
	}

	err := c.GetRequestBody(&req)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	affected, err := models.AddUserRole(userOwner, userName, req.RoleOwner, req.RoleName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.ResponseError("Failed to add role to user")
		return
	}

	// Invalidate user cache
	services.InvalidateUserCache(models.GetId(userOwner, userName))

	c.ResponseOk()
}

// RemoveUserRole removes a role from a user
// @router /api/users/:owner/:name/roles/:roleOwner/:roleName [delete]
func (c *PermissionController) RemoveUserRole() {
	if !c.checkPermission("user", "write") {
		return
	}

	userOwner := c.GetString(":owner")
	userName := c.GetString(":name")
	roleOwner := c.GetString(":roleOwner")
	roleName := c.GetString(":roleName")

	affected, err := models.RemoveUserRole(userOwner, userName, roleOwner, roleName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !affected {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("User role not found")
		return
	}

	// Invalidate user cache
	services.InvalidateUserCache(models.GetId(userOwner, userName))

	c.ResponseOk()
}

// GetUserPermissions gets all effective permissions for a user
// @router /api/users/:owner/:name/permissions [get]
func (c *PermissionController) GetUserPermissions() {
	if !c.checkPermission("user", "read") {
		return
	}

	owner := c.GetString(":owner")
	name := c.GetString(":name")

	perms, err := services.GetUserEffectivePermissions(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(perms)
}
