// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import "fmt"

type Organization struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName   string `xorm:"varchar(100)" json:"displayName"`
	WebsiteUrl    string `xorm:"varchar(100)" json:"websiteUrl"`
	Logo          string `xorm:"varchar(200)" json:"logo"`
	PasswordType  string `xorm:"varchar(100)" json:"passwordType"`
	PasswordSalt  string `xorm:"varchar(100)" json:"passwordSalt"`
	DefaultAvatar string `xorm:"varchar(200)" json:"defaultAvatar"`
	EnableSignUp  bool   `json:"enableSignUp"`
	DcrPolicy     string `xorm:"varchar(100)" json:"dcrPolicy"` // Dynamic Client Registration policy
}

func (o *Organization) GetId() string {
	return fmt.Sprintf("%s/%s", o.Owner, o.Name)
}

func GetOrganization(owner, name string) (*Organization, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	org := Organization{Owner: owner, Name: name}
	existed, err := engine.Get(&org)
	if err != nil {
		return nil, err
	}

	if existed {
		return &org, nil
	}
	return nil, nil
}

func AddOrganization(org *Organization) (bool, error) {
	affected, err := engine.Insert(org)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func UpdateOrganization(owner, name string, org *Organization) (bool, error) {
	affected, err := engine.Where("owner = ? AND name = ?", owner, name).AllCols().Update(org)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteOrganization(owner, name string) (bool, error) {
	affected, err := engine.Delete(&Organization{Owner: owner, Name: name})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}
