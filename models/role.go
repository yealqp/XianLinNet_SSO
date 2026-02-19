// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import "fmt"

// Role represents a role in the system
type Role struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(500)" json:"description"`
	IsEnabled   bool   `json:"isEnabled"`

	// Role type: "system" (built-in) or "custom" (user-defined)
	Type string `xorm:"varchar(50)" json:"type"`

	// Organization this role belongs to
	Organization string `xorm:"varchar(100) index" json:"organization"`
}

func (r *Role) GetId() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// GetRole retrieves a role by owner and name
func GetRole(owner, name string) (*Role, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	role := Role{Owner: owner, Name: name}
	existed, err := engine.Get(&role)
	if err != nil {
		return nil, err
	}

	if existed {
		return &role, nil
	}
	return nil, nil
}

// GetRolesByOrganization retrieves all roles for an organization
func GetRolesByOrganization(organization string) ([]*Role, error) {
	roles := []*Role{}
	err := engine.Where("organization = ?", organization).Find(&roles)
	return roles, err
}

// AddRole creates a new role
func AddRole(role *Role) (bool, error) {
	affected, err := engine.Insert(role)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// UpdateRole updates an existing role
func UpdateRole(owner, name string, role *Role) (bool, error) {
	affected, err := engine.Where("owner = ? AND name = ?", owner, name).AllCols().Update(role)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// DeleteRole deletes a role
func DeleteRole(owner, name string) (bool, error) {
	// Also delete associated role permissions and user roles
	_, err := engine.Where("role_owner = ? AND role_name = ?", owner, name).Delete(&RolePermission{})
	if err != nil {
		return false, err
	}

	_, err = engine.Where("role_owner = ? AND role_name = ?", owner, name).Delete(&UserRole{})
	if err != nil {
		return false, err
	}

	affected, err := engine.Delete(&Role{Owner: owner, Name: name})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}
