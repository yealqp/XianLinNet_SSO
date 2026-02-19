// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import "fmt"

// Permission represents a permission in the system
type Permission struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(500)" json:"description"`

	// Resource type: "user", "application", "token", "organization", "role", "permission", etc.
	Resource string `xorm:"varchar(100) index" json:"resource"`

	// Action: "read", "write", "delete", "manage", "*" (all)
	Action string `xorm:"varchar(100)" json:"action"`

	// Effect: "allow" or "deny"
	Effect string `xorm:"varchar(50)" json:"effect"`

	// IsEnabled indicates if this permission is active
	IsEnabled bool `json:"isEnabled"`
}

func (p *Permission) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}

// GetPermission retrieves a permission by owner and name
func GetPermission(owner, name string) (*Permission, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	perm := Permission{Owner: owner, Name: name}
	existed, err := engine.Get(&perm)
	if err != nil {
		return nil, err
	}

	if existed {
		return &perm, nil
	}
	return nil, nil
}

// GetAllPermissions retrieves all permissions
func GetAllPermissions(owner string) ([]*Permission, error) {
	perms := []*Permission{}
	err := engine.Where("owner = ?", owner).Find(&perms)
	return perms, err
}

// AddPermission creates a new permission
func AddPermission(perm *Permission) (bool, error) {
	affected, err := engine.Insert(perm)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// UpdatePermission updates an existing permission
func UpdatePermission(owner, name string, perm *Permission) (bool, error) {
	affected, err := engine.Where("owner = ? AND name = ?", owner, name).AllCols().Update(perm)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// DeletePermission deletes a permission
func DeletePermission(owner, name string) (bool, error) {
	// Also delete associated role permissions
	_, err := engine.Where("permission_owner = ? AND permission_name = ?", owner, name).Delete(&RolePermission{})
	if err != nil {
		return false, err
	}

	affected, err := engine.Delete(&Permission{Owner: owner, Name: name})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}
