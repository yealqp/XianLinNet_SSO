// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

type Provider struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName  string `xorm:"varchar(100)" json:"displayName"`
	Category     string `xorm:"varchar(100)" json:"category"` // OAuth, SAML, etc.
	Type         string `xorm:"varchar(100)" json:"type"`     // GitHub, Google, etc.
	SubType      string `xorm:"varchar(100)" json:"subType"`
	ClientId     string `xorm:"varchar(200)" json:"clientId"`
	ClientSecret string `xorm:"varchar(200)" json:"clientSecret"`
	HostUrl      string `xorm:"varchar(200)" json:"hostUrl"`
	AuthUrl      string `xorm:"varchar(200)" json:"authUrl"`
	TokenUrl     string `xorm:"varchar(200)" json:"tokenUrl"`
	UserInfoUrl  string `xorm:"varchar(200)" json:"userInfoUrl"`
}

func GetProvider(owner, name string) (*Provider, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	provider := Provider{Owner: owner, Name: name}
	existed, err := engine.Get(&provider)
	if err != nil {
		return nil, err
	}

	if existed {
		return &provider, nil
	}
	return nil, nil
}

func AddProvider(provider *Provider) (bool, error) {
	affected, err := engine.Insert(provider)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}
