// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"fmt"
	"strings"

	"github.com/oauth-server/oauth-server/models"
)

// CheckPermission checks if a user has a specific permission
func CheckPermission(userId int64, resource, action string) (bool, error) {
	// Admin users have all permissions
	user, err := models.GetUserById(userId)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}
	if user.IsAdmin {
		return true, nil
	}

	// Get user's permissions through roles
	perms, err := models.GetUserPermissions(userId)
	if err != nil {
		return false, err
	}

	// Check if user has the required permission
	for _, perm := range perms {
		if !perm.IsEnabled {
			continue
		}

		// Check resource match (exact or wildcard)
		resourceMatch := perm.Resource == resource || perm.Resource == "*"

		// Check action match (exact or wildcard)
		actionMatch := perm.Action == action || perm.Action == "*"

		if resourceMatch && actionMatch {
			// Check effect
			if perm.Effect == "allow" {
				return true, nil
			} else if perm.Effect == "deny" {
				return false, nil
			}
		}
	}

	return false, nil
}

// CheckMultiplePermissions checks if a user has any of the specified permissions
func CheckMultiplePermissions(userId int64, permissions []string) (bool, error) {
	for _, perm := range permissions {
		parts := strings.Split(perm, ":")
		if len(parts) != 2 {
			continue
		}
		resource := parts[0]
		action := parts[1]

		hasPermission, err := CheckPermission(userId, resource, action)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	return false, nil
}

// GetUserEffectivePermissions returns all effective permissions for a user
func GetUserEffectivePermissions(userId int64) (map[string][]string, error) {
	perms, err := models.GetUserPermissions(userId)
	if err != nil {
		return nil, err
	}

	// Group permissions by resource
	result := make(map[string][]string)
	for _, perm := range perms {
		if !perm.IsEnabled || perm.Effect != "allow" {
			continue
		}

		if _, exists := result[perm.Resource]; !exists {
			result[perm.Resource] = []string{}
		}
		result[perm.Resource] = append(result[perm.Resource], perm.Action)
	}

	return result, nil
}

// HasRole checks if a user has a specific role
func HasRole(userId int64, roleOwner, roleName string) (bool, error) {
	roles, err := models.GetUserRoles(userId)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Owner == roleOwner && role.Name == roleName && role.IsEnabled {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyRole checks if a user has any of the specified roles
func HasAnyRole(userId int64, roleNames []string) (bool, error) {
	roles, err := models.GetUserRoles(userId)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if !role.IsEnabled {
			continue
		}
		for _, roleName := range roleNames {
			if role.Name == roleName {
				return true, nil
			}
		}
	}

	return false, nil
}

// InitializeDefaultRolesAndPermissions creates default roles and permissions
func InitializeDefaultRolesAndPermissions(owner, organization string) error {
	// Create default permissions
	defaultPermissions := []models.Permission{
		// User permissions
		{Owner: owner, Name: "user-read", DisplayName: "Read Users", Resource: "user", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: owner, Name: "user-write", DisplayName: "Write Users", Resource: "user", Action: "write", Effect: "allow", IsEnabled: true},
		{Owner: owner, Name: "user-delete", DisplayName: "Delete Users", Resource: "user", Action: "delete", Effect: "allow", IsEnabled: true},

		// Application permissions
		{Owner: owner, Name: "app-read", DisplayName: "Read Applications", Resource: "application", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: owner, Name: "app-write", DisplayName: "Write Applications", Resource: "application", Action: "write", Effect: "allow", IsEnabled: true},
		{Owner: owner, Name: "app-delete", DisplayName: "Delete Applications", Resource: "application", Action: "delete", Effect: "allow", IsEnabled: true},

		// Token permissions
		{Owner: owner, Name: "token-read", DisplayName: "Read Tokens", Resource: "token", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: owner, Name: "token-revoke", DisplayName: "Revoke Tokens", Resource: "token", Action: "revoke", Effect: "allow", IsEnabled: true},

		// Organization permissions
		{Owner: owner, Name: "org-read", DisplayName: "Read Organizations", Resource: "organization", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: owner, Name: "org-write", DisplayName: "Write Organizations", Resource: "organization", Action: "write", Effect: "allow", IsEnabled: true},

		// Role permissions
		{Owner: owner, Name: "role-read", DisplayName: "Read Roles", Resource: "role", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: owner, Name: "role-write", DisplayName: "Write Roles", Resource: "role", Action: "write", Effect: "allow", IsEnabled: true},

		// Permission permissions
		{Owner: owner, Name: "perm-read", DisplayName: "Read Permissions", Resource: "permission", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: owner, Name: "perm-write", DisplayName: "Write Permissions", Resource: "permission", Action: "write", Effect: "allow", IsEnabled: true},

		// Admin permission (all resources, all actions)
		{Owner: owner, Name: "admin-all", DisplayName: "Admin All", Resource: "*", Action: "*", Effect: "allow", IsEnabled: true},
	}

	for _, perm := range defaultPermissions {
		perm.CreatedTime = models.GetCurrentTime()
		exists, _ := models.GetPermission(perm.Owner, perm.Name)
		if exists == nil {
			models.AddPermission(&perm)
		}
	}

	// Create default roles
	defaultRoles := []models.Role{
		{Owner: owner, Name: "admin", DisplayName: "Administrator", Description: "Full system access", Type: "system", Organization: organization, IsEnabled: true},
		{Owner: owner, Name: "user-manager", DisplayName: "User Manager", Description: "Manage users", Type: "system", Organization: organization, IsEnabled: true},
		{Owner: owner, Name: "app-manager", DisplayName: "Application Manager", Description: "Manage applications", Type: "system", Organization: organization, IsEnabled: true},
		{Owner: owner, Name: "viewer", DisplayName: "Viewer", Description: "Read-only access", Type: "system", Organization: organization, IsEnabled: true},
		{Owner: owner, Name: "user", DisplayName: "Regular User", Description: "Basic user access", Type: "system", Organization: organization, IsEnabled: true},
	}

	for _, role := range defaultRoles {
		role.CreatedTime = models.GetCurrentTime()
		exists, _ := models.GetRole(role.Owner, role.Name)
		if exists == nil {
			models.AddRole(&role)
		}
	}

	// Assign permissions to roles
	rolePermissions := map[string][]string{
		"admin":        {"admin-all"},
		"user-manager": {"user-read", "user-write", "user-delete", "role-read"},
		"app-manager":  {"app-read", "app-write", "app-delete", "token-read", "token-revoke"},
		"viewer":       {"user-read", "app-read", "token-read", "org-read", "role-read", "perm-read"},
		"user":         {"user-read"},
	}

	for roleName, permNames := range rolePermissions {
		for _, permName := range permNames {
			models.AddRolePermission(owner, roleName, owner, permName)
		}
	}

	return nil
}
