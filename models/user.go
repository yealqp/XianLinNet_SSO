// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import (
	"fmt"
)

type User struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(255) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	Id                string            `xorm:"varchar(100) index" json:"id"`
	Type              string            `xorm:"varchar(100)" json:"type"`
	Password          string            `xorm:"varchar(150)" json:"password"`
	PasswordSalt      string            `xorm:"varchar(100)" json:"passwordSalt"`
	DisplayName       string            `xorm:"varchar(100)" json:"displayName"`
	Avatar            string            `xorm:"text" json:"avatar"`
	Email             string            `xorm:"varchar(100) index" json:"email"`
	EmailVerified     bool              `json:"emailVerified"`
	Phone             string            `xorm:"varchar(100) index" json:"phone"`
	CountryCode       string            `xorm:"varchar(6)" json:"countryCode"`
	IsAdmin           bool              `json:"isAdmin"`
	IsForbidden       bool              `json:"isForbidden"`
	IsDeleted         bool              `json:"isDeleted"`
	SignupApplication string            `xorm:"varchar(100)" json:"signupApplication"`
	Properties        map[string]string `xorm:"text json" json:"properties"`

	// OAuth fields
	AccessToken          string `xorm:"mediumtext" json:"accessToken"`
	OriginalToken        string `xorm:"mediumtext" json:"originalToken"`
	OriginalRefreshToken string `xorm:"mediumtext" json:"originalRefreshToken"`
}

func (u *User) GetId() string {
	return fmt.Sprintf("%s/%s", u.Owner, u.Name)
}

func GetUser(owner, name string) (*User, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	user := User{Owner: owner, Name: name}
	existed, err := engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	}
	return nil, nil
}

func GetUserById(owner, id string) (*User, error) {
	if owner == "" || id == "" {
		return nil, nil
	}

	user := User{Owner: owner, Id: id}
	existed, err := engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	}
	return nil, nil
}

func GetUserByEmail(owner, email string) (*User, error) {
	if owner == "" || email == "" {
		return nil, nil
	}

	user := User{Owner: owner, Email: email}
	existed, err := engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	}
	return nil, nil
}

func GetUserByFields(owner, field string) (*User, error) {
	// Try to get user by name, email, or phone
	user, err := GetUser(owner, field)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	user, err = GetUserByEmail(owner, field)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	return nil, nil
}

func AddUser(user *User) (bool, error) {
	affected, err := engine.Insert(user)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func UpdateUser(owner, name string, user *User) (bool, error) {
	affected, err := engine.Where("owner = ? AND name = ?", owner, name).AllCols().Update(user)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteUser(owner, name string) (bool, error) {
	affected, err := engine.Delete(&User{Owner: owner, Name: name})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}
