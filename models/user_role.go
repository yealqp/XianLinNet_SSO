// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import "fmt"

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	Id          int64  `xorm:"pk autoincr" json:"id"`
	UserOwner   string `xorm:"varchar(100) notnull index" json:"userOwner"`
	UserName    string `xorm:"varchar(255) notnull index" json:"userName"`
	RoleOwner   string `xorm:"varchar(100) notnull index" json:"roleOwner"`
	RoleName    string `xorm:"varchar(100) notnull index" json:"roleName"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
}

func (ur *UserRole) GetId() string {
	return fmt.Sprintf("%d", ur.Id)
}

// AddUserRole assigns a role to a user
func AddUserRole(userOwner, userName, roleOwner, roleName string) (bool, error) {
	// Check if already exists
	exists, err := engine.Where("user_owner = ? AND user_name = ? AND role_owner = ? AND role_name = ?",
		userOwner, userName, roleOwner, roleName).Exist(&UserRole{})
	if err != nil {
		return false, err
	}
	if exists {
		return false, fmt.Errorf("user role already exists")
	}

	ur := &UserRole{
		UserOwner:   userOwner,
		UserName:    userName,
		RoleOwner:   roleOwner,
		RoleName:    roleName,
		CreatedTime: GetCurrentTime(),
	}

	affected, err := engine.Insert(ur)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// RemoveUserRole removes a role from a user
func RemoveUserRole(userOwner, userName, roleOwner, roleName string) (bool, error) {
	affected, err := engine.Where("user_owner = ? AND user_name = ? AND role_owner = ? AND role_name = ?",
		userOwner, userName, roleOwner, roleName).Delete(&UserRole{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// GetUserRoles retrieves all roles for a user
func GetUserRoles(userOwner, userName string) ([]*Role, error) {
	var roles []*Role

	err := engine.SQL(`
		SELECT r.* FROM role r
		INNER JOIN user_role ur ON r.owner = ur.role_owner AND r.name = ur.role_name
		WHERE ur.user_owner = ? AND ur.user_name = ?
	`, userOwner, userName).Find(&roles)

	return roles, err
}

// GetRoleUsers retrieves all users that have a specific role
func GetRoleUsers(roleOwner, roleName string) ([]*User, error) {
	var users []*User

	err := engine.SQL(`
		SELECT u.* FROM user u
		INNER JOIN user_role ur ON u.owner = ur.user_owner AND u.name = ur.user_name
		WHERE ur.role_owner = ? AND ur.role_name = ?
	`, roleOwner, roleName).Find(&users)

	return users, err
}

// GetUserPermissions retrieves all permissions for a user (through their roles)
func GetUserPermissions(userOwner, userName string) ([]*Permission, error) {
	var perms []*Permission

	err := engine.SQL(`
		SELECT DISTINCT p.* FROM permission p
		INNER JOIN role_permission rp ON p.owner = rp.perm_owner AND p.name = rp.perm_name
		INNER JOIN user_role ur ON rp.role_owner = ur.role_owner AND rp.role_name = ur.role_name
		WHERE ur.user_owner = ? AND ur.user_name = ? AND p.is_enabled = 1
	`, userOwner, userName).Find(&perms)

	return perms, err
}
