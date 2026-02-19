// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oauth-server/oauth-server/models"
)

type Claims struct {
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	CreatedTime string   `json:"createdTime"`
	Id          string   `json:"id"`
	Type        string   `json:"type"`
	DisplayName string   `json:"displayName"`
	Avatar      string   `json:"avatar"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	IsAdmin     bool     `json:"isAdmin"`
	Scope       string   `json:"scope"`
	Iss         string   `json:"iss"`
	Sub         string   `json:"sub"`
	Aud         []string `json:"aud"`
	Nonce       string   `json:"nonce,omitempty"`
	jwt.RegisteredClaims
}

// GenerateJwtToken generates access and refresh tokens
func GenerateJwtToken(application *models.Application, user *models.User, scope, nonce, resource string) (string, string, string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(application.ExpireInHours) * time.Hour)
	refreshExpireTime := nowTime.Add(time.Duration(application.RefreshExpireInHours) * time.Hour)

	// Get JWT secret from config
	jwtSecret, _ := web.AppConfig.String("jwtSecret")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-in-production"
	}

	// Get origin from config
	origin, _ := web.AppConfig.String("origin")
	if origin == "" {
		origin = "http://localhost:8080"
	}

	// Create claims
	claims := Claims{
		Owner:       user.Owner,
		Name:        user.Name,
		CreatedTime: user.CreatedTime,
		Id:          user.Id,
		Type:        user.Type,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Email:       user.Email,
		Phone:       user.Phone,
		IsAdmin:     user.IsAdmin,
		Scope:       scope,
		Iss:         origin,
		Sub:         user.GetId(),
		Aud:         []string{application.ClientId},
		Nonce:       nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			NotBefore: jwt.NewNumericDate(nowTime),
			ID:        models.GenerateRandomString(16),
		},
	}

	// Generate access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", "", err
	}

	// Generate refresh token with longer expiration
	refreshClaims := claims
	refreshClaims.ExpiresAt = jwt.NewNumericDate(refreshExpireTime)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", "", err
	}

	// Generate token name
	tokenName := fmt.Sprintf("token_%s_%d", user.Name, nowTime.Unix())

	return accessToken, refreshTokenString, tokenName, nil
}

// ParseJwtToken parses and validates a JWT token
func ParseJwtToken(tokenString string) (*Claims, error) {
	jwtSecret, _ := web.AppConfig.String("jwtSecret")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-in-production"
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateToken validates a token and returns user info
func ValidateToken(tokenString string) (*models.User, error) {
	claims, err := ParseJwtToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Get user from database
	user, err := models.GetUser(claims.Owner, claims.Name)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.IsForbidden {
		return nil, fmt.Errorf("user is forbidden")
	}

	return user, nil
}
