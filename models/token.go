// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Token struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Application  string `xorm:"varchar(100)" json:"application"`
	Organization string `xorm:"varchar(100)" json:"organization"`
	User         string `xorm:"varchar(100)" json:"user"`

	Code             string `xorm:"varchar(100) index" json:"code"`
	AccessToken      string `xorm:"mediumtext" json:"accessToken"`
	RefreshToken     string `xorm:"mediumtext" json:"refreshToken"`
	AccessTokenHash  string `xorm:"varchar(100) index" json:"accessTokenHash"`
	RefreshTokenHash string `xorm:"varchar(100) index" json:"refreshTokenHash"`
	ExpiresIn        int    `json:"expiresIn"`        // Token有效期长度（秒）
	ExpiresAt        int64  `json:"expiresAt"`        // Access Token过期时间戳
	RefreshExpiresAt int64  `json:"refreshExpiresAt"` // Refresh Token过期时间戳
	Scope            string `xorm:"varchar(100)" json:"scope"`
	TokenType        string `xorm:"varchar(100)" json:"tokenType"`
	CodeChallenge    string `xorm:"varchar(100)" json:"codeChallenge"`
	CodeIsUsed       bool   `json:"codeIsUsed"`
	CodeExpireIn     int64  `json:"codeExpireIn"`
	Resource         string `xorm:"varchar(255)" json:"resource"`

	// OAuth 2.1 security enhancements
	RefreshTokenUsed bool   `json:"refreshTokenUsed"`
	TokenFamily      string `xorm:"varchar(100) index" json:"tokenFamily"`
}

func (t *Token) GetId() string {
	return fmt.Sprintf("%s/%s", t.Owner, t.Name)
}

func getTokenHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	res := hex.EncodeToString(hash[:])
	if len(res) > 64 {
		return res[:64]
	}
	return res
}

func (t *Token) PopulateHashes() {
	if t.AccessTokenHash == "" && t.AccessToken != "" {
		t.AccessTokenHash = getTokenHash(t.AccessToken)
	}
	if t.RefreshTokenHash == "" && t.RefreshToken != "" {
		t.RefreshTokenHash = getTokenHash(t.RefreshToken)
	}
}

func GetToken(owner, name string) (*Token, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	token := Token{Owner: owner, Name: name}
	existed, err := engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if existed {
		return &token, nil
	}
	return nil, nil
}

func GetTokenByCode(code string) (*Token, error) {
	token := Token{Code: code}
	existed, err := engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if existed {
		return &token, nil
	}
	return nil, nil
}

func GetTokenByAccessToken(accessToken string) (*Token, error) {
	token := Token{AccessTokenHash: getTokenHash(accessToken)}
	existed, err := engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if existed {
		return &token, nil
	}
	return nil, nil
}

func GetTokenByRefreshToken(refreshToken string) (*Token, error) {
	token := Token{RefreshTokenHash: getTokenHash(refreshToken)}
	existed, err := engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if existed {
		return &token, nil
	}
	return nil, nil
}

func AddToken(token *Token) (bool, error) {
	token.PopulateHashes()

	affected, err := engine.Insert(token)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func UpdateToken(owner, name string, token *Token) (bool, error) {
	token.PopulateHashes()

	affected, err := engine.Where("owner = ? AND name = ?", owner, name).AllCols().Update(token)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func UpdateTokenByCode(code string, token *Token) (bool, error) {
	affected, err := engine.Where("code = ?", code).Cols("code_is_used").Update(token)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteToken(owner, name string) (bool, error) {
	affected, err := engine.Delete(&Token{Owner: owner, Name: name})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// GetTokensByFamily retrieves all tokens in a token family
func GetTokensByFamily(tokenFamily string) ([]*Token, error) {
	if tokenFamily == "" {
		return nil, nil
	}

	var tokens []*Token
	err := engine.Where("token_family = ?", tokenFamily).Find(&tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// RevokeTokenFamily revokes all tokens in a token family (for refresh token reuse detection)
func RevokeTokenFamily(tokenFamily string) error {
	if tokenFamily == "" {
		return nil
	}

	_, err := engine.Where("token_family = ?", tokenFamily).Cols("expires_in").Update(&Token{ExpiresIn: 0})
	return err
}

// IsAccessTokenExpired checks if the access token is expired
func (t *Token) IsAccessTokenExpired() bool {
	if t.ExpiresAt == 0 {
		return false // No expiration set
	}
	return time.Now().Unix() > t.ExpiresAt
}

// IsRefreshTokenExpired checks if the refresh token is expired
func (t *Token) IsRefreshTokenExpired() bool {
	if t.RefreshExpiresAt == 0 {
		return false // No expiration set
	}
	return time.Now().Unix() > t.RefreshExpiresAt
}

// IsRevoked checks if the token has been revoked
func (t *Token) IsRevoked() bool {
	return t.ExpiresIn <= 0
}
