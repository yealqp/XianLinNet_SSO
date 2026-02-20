// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import (
	"fmt"
)

type User struct {
	Id          int64  `xorm:"pk autoincr" json:"id"`
	Owner       string `xorm:"varchar(100) notnull index" json:"owner"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	Type        string            `xorm:"varchar(100)" json:"type"`
	Password    string            `xorm:"varchar(150)" json:"password"`
	Username    string            `xorm:"varchar(100)" json:"username"`
	Avatar      string            `xorm:"text" json:"avatar"`
	Email       string            `xorm:"varchar(100) unique index" json:"email"`
	QQ          string            `xorm:"'qq' varchar(20)" json:"qq"`
	IsRealName  bool              `json:"isRealName"`
	RealName    string            `xorm:"text" json:"-"` // 加密存储的真实姓名，不返回给前端
	IDCard      string            `xorm:"text" json:"-"` // 加密存储的身份证号，不返回给前端
	CountryCode string            `xorm:"varchar(6)" json:"countryCode"`
	IsAdmin     bool              `json:"isAdmin"`
	IsForbidden bool              `json:"isForbidden"`
	IsDeleted   bool              `json:"isDeleted"`
	Properties  map[string]string `xorm:"text json" json:"properties"`

	// OAuth fields
	SignupApplication    string `xorm:"varchar(100)" json:"signupApplication"`
	AccessToken          string `xorm:"mediumtext" json:"accessToken"`
	OriginalToken        string `xorm:"mediumtext" json:"originalToken"`
	OriginalRefreshToken string `xorm:"mediumtext" json:"originalRefreshToken"`
}

func (u *User) GetId() string {
	return fmt.Sprintf("%d", u.Id)
}

func GetUserById(id int64) (*User, error) {
	if id == 0 {
		return nil, nil
	}

	user := User{Id: id}
	existed, err := engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	}
	return nil, nil
}

func GetUserByEmail(email string) (*User, error) {
	if email == "" {
		return nil, nil
	}

	user := User{Email: email}
	existed, err := engine.Get(&user)
	if err != nil {
		return nil, err
	}

	if existed {
		return &user, nil
	}
	return nil, nil
}

func GetUserByUsername(username string) (*User, error) {
	if username == "" {
		return nil, nil
	}

	user := User{Username: username}
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
	// Try to get user by email or username
	user, err := GetUserByEmail(field)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	user, err = GetUserByUsername(field)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	return nil, nil
}

func GetUsers(owner string) ([]*User, error) {
	users := []*User{}
	err := engine.Where("owner = ?", owner).Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func AddUser(user *User) (bool, error) {
	affected, err := engine.Insert(user)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func UpdateUser(id int64, user *User) (bool, error) {
	affected, err := engine.ID(id).AllCols().Update(user)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteUser(id int64) (bool, error) {
	affected, err := engine.ID(id).Delete(&User{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}
