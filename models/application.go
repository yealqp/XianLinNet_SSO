// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import (
	"fmt"
	"regexp"
	"strings"
)

type Application struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName      string   `xorm:"varchar(100)" json:"displayName"`
	Logo             string   `xorm:"varchar(200)" json:"logo"`
	HomepageUrl      string   `xorm:"varchar(100)" json:"homepageUrl"`
	Description      string   `xorm:"varchar(100)" json:"description"`
	Organization     string   `xorm:"varchar(100)" json:"organization"`
	Cert             string   `xorm:"varchar(100)" json:"cert"`
	EnablePassword   bool     `json:"enablePassword"`
	EnableSignUp     bool     `json:"enableSignUp"`
	EnableCodeSignin bool     `json:"enableCodeSignin"`
	GrantTypes       []string `xorm:"varchar(1000) json" json:"grantTypes"`
	Tags             []string `xorm:"mediumtext json" json:"tags"`

	ClientId             string   `xorm:"varchar(100)" json:"clientId"`
	ClientSecret         string   `xorm:"varchar(100)" json:"clientSecret"`
	RedirectUris         []string `xorm:"varchar(1000) json" json:"redirectUris"`
	TokenFormat          string   `xorm:"varchar(100)" json:"tokenFormat"`
	ExpireInHours        float64  `json:"expireInHours"`
	RefreshExpireInHours float64  `json:"refreshExpireInHours"`
	Scopes               []string `xorm:"varchar(1000) json" json:"scopes"`
}

func (a *Application) GetId() string {
	return fmt.Sprintf("%s/%s", a.Owner, a.Name)
}

func (a *Application) IsRedirectUriValid(redirectUri string) bool {
	if redirectUri == "" {
		return false
	}

	for _, targetUri := range a.RedirectUris {
		if targetUri == "" {
			continue
		}

		// First try exact match (most secure)
		if targetUri == redirectUri {
			return true
		}

		// If configured URI starts with "regex:", use regex matching
		if strings.HasPrefix(targetUri, "regex:") {
			pattern := targetUri[6:]
			matched, err := regexp.MatchString(pattern, redirectUri)
			if err == nil && matched {
				return true
			}
			continue
		}

		// If configured URI starts with "prefix:", use prefix matching
		if strings.HasPrefix(targetUri, "prefix:") {
			prefix := targetUri[7:]
			if strings.HasPrefix(redirectUri, prefix) {
				return true
			}
			continue
		}

		// Legacy: regex matching without prefix (less secure, kept for compatibility)
		targetUriRegex := regexp.MustCompile(targetUri)
		if targetUriRegex.MatchString(redirectUri) {
			return true
		}
	}
	return false
}

func GetApplication(owner, name string) (*Application, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	app := Application{Owner: owner, Name: name}
	existed, err := engine.Get(&app)
	if err != nil {
		return nil, err
	}

	if existed {
		return &app, nil
	}
	return nil, nil
}

func GetApplicationByClientId(clientId string) (*Application, error) {
	app := Application{ClientId: clientId}
	existed, err := engine.Get(&app)
	if err != nil {
		return nil, err
	}

	if existed {
		return &app, nil
	}
	return nil, nil
}

func AddApplication(app *Application) (bool, error) {
	// Set defaults for required fields
	if app.ClientId == "" {
		app.ClientId = GenerateClientId()
	}
	if app.ClientSecret == "" {
		app.ClientSecret = GenerateClientSecret()
	}
	if app.ExpireInHours == 0 {
		app.ExpireInHours = 168 // 7 days
	}
	if app.RefreshExpireInHours == 0 {
		app.RefreshExpireInHours = 720 // 30 days
	}
	if app.TokenFormat == "" {
		app.TokenFormat = "JWT"
	}

	affected, err := engine.Insert(app)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func UpdateApplication(owner, name string, app *Application) (bool, error) {
	affected, err := engine.Where("owner = ? AND name = ?", owner, name).AllCols().Update(app)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteApplication(owner, name string) (bool, error) {
	affected, err := engine.Delete(&Application{Owner: owner, Name: name})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func IsGrantTypeValid(method string, grantTypes []string) bool {
	if method == "authorization_code" {
		return true
	}
	for _, m := range grantTypes {
		if m == method {
			return true
		}
	}
	return false
}
