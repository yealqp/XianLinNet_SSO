// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import "fmt"

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	Id          int64  `xorm:"pk autoincr" json:"id"`
	RoleOwner   string `xorm:"varchar(100) notnull index" json:"roleOwner"`
	RoleName    string `xorm:"varchar(100) notnull index" json:"roleName"`
	PermOwner   string `xorm:"varchar(100) notnull index" json:"permOwner"`
	PermName    string `xorm:"varchar(100) notnull index" json:"permName"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
}

func (rp *RolePermission) GetId() string {
	return fmt.Sprintf("%d", rp.Id)
}

// AddRolePermission assigns a permission to a role
func AddRolePermission(roleOwner, roleName, permOwner, permName string) (bool, error) {
	// Check if already exists
	exists, err := engine.Where("role_owner = ? AND role_name = ? AND perm_owner = ? AND perm_name = ?",
		roleOwner, roleName, permOwner, permName).Exist(&RolePermission{})
	if err != nil {
		return false, err
	}
	if exists {
		return false, fmt.Errorf("role permission already exists")
	}

	rp := &RolePermission{
		RoleOwner:   roleOwner,
		RoleName:    roleName,
		PermOwner:   permOwner,
		PermName:    permName,
		CreatedTime: GetCurrentTime(),
	}

	affected, err := engine.Insert(rp)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// RemoveRolePermission removes a permission from a role
func RemoveRolePermission(roleOwner, roleName, permOwner, permName string) (bool, error) {
	affected, err := engine.Where("role_owner = ? AND role_name = ? AND perm_owner = ? AND perm_name = ?",
		roleOwner, roleName, permOwner, permName).Delete(&RolePermission{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// GetRolePermissions retrieves all permissions for a role
func GetRolePermissions(roleOwner, roleName string) ([]*Permission, error) {
	var perms []*Permission

	err := engine.SQL(`
		SELECT p.* FROM permission p
		INNER JOIN role_permission rp ON p.owner = rp.perm_owner AND p.name = rp.perm_name
		WHERE rp.role_owner = ? AND rp.role_name = ?
	`, roleOwner, roleName).Find(&perms)

	return perms, err
}

// GetPermissionRoles retrieves all roles that have a specific permission
func GetPermissionRoles(permOwner, permName string) ([]*Role, error) {
	var roles []*Role

	err := engine.SQL(`
		SELECT r.* FROM role r
		INNER JOIN role_permission rp ON r.owner = rp.role_owner AND r.name = rp.role_name
		WHERE rp.perm_owner = ? AND rp.perm_name = ?
	`, permOwner, permName).Find(&roles)

	return roles, err
}
