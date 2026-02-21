// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oauth-server/oauth-server/models"
)

type Claims struct {
	Owner       string   `json:"owner"`
	CreatedTime string   `json:"createdTime"`
	Id          string   `json:"id"`
	Type        string   `json:"type"`
	Username    string   `json:"username"`
	Avatar      string   `json:"avatar"`
	Email       string   `json:"email"`
	QQ          string   `json:"qq"`
	IsRealName  bool     `json:"isRealName"`
	IsAdmin     bool     `json:"isAdmin"`
	Scope       string   `json:"scope"`
	Iss         string   `json:"iss"`
	Sub         string   `json:"sub"`
	Aud         []string `json:"aud"`
	Nonce       string   `json:"nonce,omitempty"`
	TokenUse    string   `json:"token_use"` // "access", "refresh", or "id"

	// OIDC Standard Claims
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Picture           string `json:"picture,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	UpdatedAt         int64  `json:"updated_at,omitempty"`

	jwt.RegisteredClaims
}

// GenerateIDToken generates an OIDC ID Token
func GenerateIDToken(application *models.Application, user *models.User, nonce string, accessToken string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(application.ExpireInHours) * time.Hour)

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

	// Generate unique JTI
	jti := fmt.Sprintf("id-%s-%d", models.GenerateClientId(), nowTime.UnixNano())
	notBefore := nowTime.Add(-10 * time.Second)

	// Create ID Token claims
	claims := Claims{
		Owner:             user.Owner,
		CreatedTime:       user.CreatedTime,
		Id:                fmt.Sprintf("%d", user.Id),
		Type:              user.Type,
		Username:          user.Username,
		Email:             user.Email,
		QQ:                user.QQ,
		IsRealName:        user.IsRealName,
		IsAdmin:           user.IsAdmin,
		Iss:               origin,
		Sub:               user.GetId(),
		Aud:               []string{application.ClientId},
		Nonce:             nonce,
		TokenUse:          "id",
		Name:              user.Username,
		PreferredUsername: user.Username,
		Picture:           user.Avatar,
		EmailVerified:     true, // 假设邮箱已验证
		UpdatedAt:         time.Now().Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			NotBefore: jwt.NewNumericDate(notBefore),
			ID:        jti,
		},
	}

	// Generate ID token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	idToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return idToken, nil
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

	// Generate unique JTI for access token using UUID + timestamp
	accessJti := fmt.Sprintf("%s-%d", models.GenerateClientId(), nowTime.UnixNano())

	// Set NotBefore to a few seconds before now to account for clock skew
	notBefore := nowTime.Add(-10 * time.Second)

	// Create claims for access token
	claims := Claims{
		Owner:             user.Owner,
		CreatedTime:       user.CreatedTime,
		Id:                fmt.Sprintf("%d", user.Id),
		Type:              user.Type,
		Username:          user.Username,
		Avatar:            user.Avatar,
		Email:             user.Email,
		QQ:                user.QQ,
		IsRealName:        user.IsRealName,
		IsAdmin:           user.IsAdmin,
		Scope:             scope,
		Iss:               origin,
		Sub:               user.GetId(),
		Aud:               []string{application.ClientId},
		Nonce:             nonce,
		TokenUse:          "access",
		Name:              user.Username,
		PreferredUsername: user.Username,
		Picture:           user.Avatar,
		EmailVerified:     true,
		UpdatedAt:         time.Now().Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			NotBefore: jwt.NewNumericDate(notBefore),
			ID:        accessJti,
		},
	}

	// Generate access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", "", err
	}

	// Small delay to ensure different timestamp for refresh token
	time.Sleep(1 * time.Millisecond)
	refreshNowTime := time.Now()

	// Generate unique JTI for refresh token using UUID + timestamp
	refreshJti := fmt.Sprintf("%s-%d", models.GenerateClientId(), refreshNowTime.UnixNano())

	// Set NotBefore to a few seconds before now to account for clock skew
	refreshNotBefore := refreshNowTime.Add(-10 * time.Second)

	// Generate refresh token with longer expiration
	// Create a new claims object for refresh token
	refreshClaims := Claims{
		Owner:       user.Owner,
		CreatedTime: user.CreatedTime,
		Id:          fmt.Sprintf("%d", user.Id),
		Type:        user.Type,
		Username:    user.Username,
		Avatar:      user.Avatar,
		Email:       user.Email,
		QQ:          user.QQ,
		IsRealName:  user.IsRealName,
		IsAdmin:     user.IsAdmin,
		Scope:       scope,
		Iss:         origin,
		Sub:         user.GetId(),
		Aud:         []string{application.ClientId},
		Nonce:       nonce,
		TokenUse:    "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpireTime),
			IssuedAt:  jwt.NewNumericDate(refreshNowTime),
			NotBefore: jwt.NewNumericDate(refreshNotBefore),
			ID:        refreshJti,
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", "", err
	}

	// Generate token name with nanosecond precision
	tokenName := fmt.Sprintf("token_%d_%d", user.Id, nowTime.UnixNano())

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

	// Get user by ID from claims
	userId, err := strconv.ParseInt(claims.Id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	user, err := models.GetUserById(userId)
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
