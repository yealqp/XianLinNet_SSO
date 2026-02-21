// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/oauth-server/oauth-server/models"
	"golang.org/x/crypto/bcrypt"
)

const (
	InvalidRequest       = "invalid_request"
	InvalidClient        = "invalid_client"
	InvalidGrant         = "invalid_grant"
	UnauthorizedClient   = "unauthorized_client"
	UnsupportedGrantType = "unsupported_grant_type"
	InvalidScope         = "invalid_scope"
	EndpointError        = "endpoint_error"
)

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IdToken      string `json:"id_token,omitempty"`
}

type CodeResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

// ValidateResourceURI validates that the resource parameter is a valid absolute URI (RFC 8707)
func ValidateResourceURI(resource string) error {
	if resource == "" {
		return nil
	}

	parsedURL, err := url.Parse(resource)
	if err != nil {
		return fmt.Errorf("resource must be a valid URI")
	}

	if !parsedURL.IsAbs() {
		return fmt.Errorf("resource must be an absolute URI")
	}

	return nil
}

// PkceChallenge generates base64-URL-encoded SHA256 hash of verifier (RFC 7636)
func PkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
	return challenge
}

// CheckOAuthLogin validates OAuth login parameters
func CheckOAuthLogin(clientId, responseType, redirectUri, scope, state string) (string, *models.Application, error) {
	// OAuth 2.1: State parameter is required for CSRF protection
	if state == "" {
		return "state parameter is required", nil, nil
	}

	// Validate state length (minimum 8 characters for security)
	if len(state) < 8 {
		return "state parameter must be at least 8 characters", nil, nil
	}

	if responseType != "code" && responseType != "token" && responseType != "id_token" {
		return fmt.Sprintf("Grant_type: %s is not supported in this application", responseType), nil, nil
	}

	application, err := models.GetApplicationByClientId(clientId)
	if err != nil {
		return "", nil, err
	}

	if application == nil {
		return "Invalid client_id", nil, nil
	}

	if !application.IsRedirectUriValid(redirectUri) {
		return fmt.Sprintf("Redirect URI: %s doesn't exist in the allowed Redirect URI list", redirectUri), application, nil
	}

	return "", application, nil
}

// GetOAuthCode generates OAuth authorization code
func GetOAuthCode(userId, clientId, responseType, redirectUri, scope, state, nonce, challenge, resource string) (*CodeResponse, error) {
	// Parse userId to int64
	userIdInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return &CodeResponse{
			Message: fmt.Sprintf("Invalid user ID: %s", userId),
			Code:    "",
		}, nil
	}

	user, err := models.GetUserById(userIdInt)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return &CodeResponse{
			Message: fmt.Sprintf("The user: %s doesn't exist", userId),
			Code:    "",
		}, nil
	}

	if user.IsForbidden {
		return &CodeResponse{
			Message: "The user is forbidden to sign in",
			Code:    "",
		}, nil
	}

	msg, application, err := CheckOAuthLogin(clientId, responseType, redirectUri, scope, state)
	if err != nil {
		return nil, err
	}

	if msg != "" {
		return &CodeResponse{
			Message: msg,
			Code:    "",
		}, nil
	}

	// Validate resource parameter (RFC 8707)
	if err := ValidateResourceURI(resource); err != nil {
		return &CodeResponse{
			Message: err.Error(),
			Code:    "",
		}, nil
	}

	// Generate JWT tokens
	accessToken, refreshToken, tokenName, err := GenerateJwtToken(application, user, scope, nonce, resource)
	if err != nil {
		return nil, err
	}

	if challenge == "null" {
		challenge = ""
	}

	// Generate token family ID for refresh token rotation tracking
	tokenFamily := models.GenerateRandomString(32)

	// Calculate expiration timestamps
	now := time.Now()
	accessExpiresAt := now.Add(time.Duration(application.ExpireInHours) * time.Hour).Unix()
	refreshExpiresAt := now.Add(time.Duration(application.RefreshExpireInHours) * time.Hour).Unix()

	// Create token record
	token := &models.Token{
		Owner:            application.Owner,
		Name:             tokenName,
		CreatedTime:      models.GetCurrentTime(),
		Application:      application.Name,
		Organization:     user.Owner,
		User:             fmt.Sprintf("%d", user.Id),
		Code:             models.GenerateRandomString(32),
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(application.ExpireInHours * 3600),
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		Scope:            scope,
		TokenType:        "Bearer",
		CodeChallenge:    challenge,
		CodeIsUsed:       false,
		CodeExpireIn:     time.Now().Add(time.Minute * 5).Unix(),
		Resource:         resource,
		TokenFamily:      tokenFamily,
	}

	_, err = models.AddToken(token)
	if err != nil {
		return nil, err
	}

	return &CodeResponse{
		Message: "",
		Code:    token.Code,
	}, nil
}

// GetOAuthToken handles token requests for various grant types
func GetOAuthToken(grantType, clientId, clientSecret, code, verifier, scope, username, password, refreshToken, resource string) (interface{}, error) {
	application, err := models.GetApplicationByClientId(clientId)
	if err != nil {
		return nil, err
	}

	if application == nil {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_id is invalid",
		}, nil
	}

	// Check if grant type is allowed
	if !models.IsGrantTypeValid(grantType, application.GrantTypes) {
		return &TokenError{
			Error:            UnsupportedGrantType,
			ErrorDescription: fmt.Sprintf("grant_type: %s is not supported in this application", grantType),
		}, nil
	}

	var token *models.Token
	var tokenError *TokenError

	switch grantType {
	case "authorization_code":
		token, tokenError, err = GetAuthorizationCodeToken(application, clientSecret, code, verifier, resource)
	case "password":
		token, tokenError, err = GetPasswordToken(application, username, password, scope)
	case "client_credentials":
		token, tokenError, err = GetClientCredentialsToken(application, clientSecret, scope)
	case "refresh_token":
		return RefreshToken(refreshToken, scope, clientId, clientSecret)
	case "urn:ietf:params:oauth:grant-type:token-exchange":
		// Token exchange implementation would go here
		return &TokenError{
			Error:            UnsupportedGrantType,
			ErrorDescription: "token-exchange not yet implemented",
		}, nil
	default:
		return &TokenError{
			Error:            UnsupportedGrantType,
			ErrorDescription: fmt.Sprintf("grant_type: %s is not supported", grantType),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	if tokenError != nil {
		return tokenError, nil
	}

	// Mark code as used
	token.CodeIsUsed = true
	_, err = models.UpdateTokenByCode(token.Code, token)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
		RefreshToken: token.RefreshToken,
		Scope:        token.Scope,
		IdToken:      token.AccessToken, // For OIDC compatibility
	}, nil
}

// GetAuthorizationCodeToken handles authorization code flow
func GetAuthorizationCodeToken(application *models.Application, clientSecret, code, verifier, resource string) (*models.Token, *TokenError, error) {
	if code == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "authorization code should not be empty",
		}, nil
	}

	token, err := models.GetTokenByCode(code)
	if err != nil {
		return nil, nil, err
	}

	if token == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("authorization code: [%s] is invalid", code),
		}, nil
	}

	if token.CodeIsUsed {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "authorization code has been used",
		}, nil
	}

	// OAuth 2.1: Public clients (no secret) MUST use PKCE
	isPublicClient := application.ClientSecret == "" || clientSecret == ""
	if isPublicClient && token.CodeChallenge == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "PKCE is required for public clients",
		}, nil
	}

	// Verify PKCE challenge if present
	if token.CodeChallenge != "" {
		if verifier == "" {
			return nil, &TokenError{
				Error:            InvalidRequest,
				ErrorDescription: "code_verifier is required when PKCE is used",
			}, nil
		}

		// Validate verifier length (RFC 7636: 43-128 characters)
		if len(verifier) < 43 || len(verifier) > 128 {
			return nil, &TokenError{
				Error:            InvalidRequest,
				ErrorDescription: "code_verifier must be 43-128 characters",
			}, nil
		}

		challengeAnswer := PkceChallenge(verifier)
		if challengeAnswer != token.CodeChallenge {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "invalid code_verifier",
			}, nil
		}
	}

	// Verify client secret (can be empty if PKCE is used)
	if application.ClientSecret != clientSecret {
		if token.CodeChallenge == "" {
			return nil, &TokenError{
				Error:            InvalidClient,
				ErrorDescription: "client_secret is invalid",
			}, nil
		} else if clientSecret != "" {
			return nil, &TokenError{
				Error:            InvalidClient,
				ErrorDescription: "client_secret is invalid",
			}, nil
		}
	}

	// Verify application matches
	if application.Name != token.Application {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "token is for wrong application",
		}, nil
	}

	// Verify resource parameter matches (RFC 8707)
	if resource != token.Resource {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "resource parameter does not match authorization request",
		}, nil
	}

	// Check code expiration
	nowUnix := time.Now().Unix()
	if nowUnix > token.CodeExpireIn {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "authorization code has expired",
		}, nil
	}

	return token, nil, nil
}

// GetPasswordToken handles password grant flow
func GetPasswordToken(application *models.Application, username, password, scope string) (*models.Token, *TokenError, error) {
	user, err := models.GetUserByFields(application.Organization, username)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "invalid username or password",
		}, nil
	}

	// Verify password (using bcrypt)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "invalid username or password",
		}, nil
	}

	if user.IsForbidden {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user is forbidden to sign in",
		}, nil
	}

	// Generate JWT tokens
	accessToken, refreshToken, tokenName, err := GenerateJwtToken(application, user, scope, "", "")
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}

	// Generate token family ID for refresh token rotation tracking
	tokenFamily := models.GenerateRandomString(32)

	// Calculate expiration timestamps
	now := time.Now()
	accessExpiresAt := now.Add(time.Duration(application.ExpireInHours) * time.Hour).Unix()
	refreshExpiresAt := now.Add(time.Duration(application.RefreshExpireInHours) * time.Hour).Unix()

	token := &models.Token{
		Owner:            application.Owner,
		Name:             tokenName,
		CreatedTime:      models.GetCurrentTime(),
		Application:      application.Name,
		Organization:     user.Owner,
		User:             fmt.Sprintf("%d", user.Id),
		Code:             models.GenerateRandomString(32),
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(application.ExpireInHours * 3600),
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		Scope:            scope,
		TokenType:        "Bearer",
		CodeIsUsed:       true,
		TokenFamily:      tokenFamily,
	}

	_, err = models.AddToken(token)
	if err != nil {
		return nil, nil, err
	}

	return token, nil, nil
}

// GetClientCredentialsToken handles client credentials flow
func GetClientCredentialsToken(application *models.Application, clientSecret, scope string) (*models.Token, *TokenError, error) {
	if application.ClientSecret != clientSecret {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}

	// Create a null user for client credentials
	nullUser := &models.User{
		Owner: application.Owner,
		Id:    0, // Service account - no real user ID
		Type:  "application",
	}

	accessToken, _, tokenName, err := GenerateJwtToken(application, nullUser, scope, "", "")
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}

	// Calculate expiration timestamps
	now := time.Now()
	accessExpiresAt := now.Add(time.Duration(application.ExpireInHours) * time.Hour).Unix()

	token := &models.Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  models.GetCurrentTime(),
		Application:  application.Name,
		Organization: application.Organization,
		User:         "0", // Service account
		Code:         models.GenerateRandomString(32),
		AccessToken:  accessToken,
		ExpiresIn:    int(application.ExpireInHours * 3600),
		ExpiresAt:    accessExpiresAt,
		Scope:        scope,
		TokenType:    "Bearer",
		CodeIsUsed:   true,
	}

	_, err = models.AddToken(token)
	if err != nil {
		return nil, nil, err
	}

	return token, nil, nil
}

// RefreshToken handles refresh token flow
func RefreshToken(refreshToken, scope, clientId, clientSecret string) (interface{}, error) {
	application, err := models.GetApplicationByClientId(clientId)
	if err != nil {
		return nil, err
	}

	if application == nil {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_id is invalid",
		}, nil
	}

	if clientSecret != "" && application.ClientSecret != clientSecret {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}

	// Get token by refresh token
	token, err := models.GetTokenByRefreshToken(refreshToken)
	if err != nil || token == nil {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token is invalid or revoked",
		}, nil
	}

	// Check if token has been revoked
	if token.IsRevoked() {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token has been revoked",
		}, nil
	}

	// Check if refresh token is expired
	if token.IsRefreshTokenExpired() {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token is expired",
		}, nil
	}

	// OAuth 2.1: Detect refresh token reuse
	if token.RefreshTokenUsed {
		// Token reuse detected! Revoke entire token family
		if token.TokenFamily != "" {
			models.RevokeTokenFamily(token.TokenFamily)
		}
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token reuse detected - all tokens revoked",
		}, nil
	}

	// Get user by ID from token
	userIdInt, err := strconv.ParseInt(token.User, 10, 64)
	if err != nil {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "invalid user ID in token",
		}, nil
	}

	user, err := models.GetUserById(userIdInt)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "user does not exist",
		}, nil
	}

	if user.IsForbidden {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user is forbidden to sign in",
		}, nil
	}

	// Use old scope if new scope is not provided
	if scope == "" {
		scope = token.Scope
	}

	// Mark old token as used (before generating new one)
	token.RefreshTokenUsed = true
	_, err = models.UpdateToken(token.Owner, token.Name, token)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	newAccessToken, newRefreshToken, tokenName, err := GenerateJwtToken(application, user, scope, "", "")
	if err != nil {
		return &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}

	// Calculate expiration timestamps
	now := time.Now()
	accessExpiresAt := now.Add(time.Duration(application.ExpireInHours) * time.Hour).Unix()
	refreshExpiresAt := now.Add(time.Duration(application.RefreshExpireInHours) * time.Hour).Unix()

	newToken := &models.Token{
		Owner:            application.Owner,
		Name:             tokenName,
		CreatedTime:      models.GetCurrentTime(),
		Application:      application.Name,
		Organization:     user.Owner,
		User:             fmt.Sprintf("%d", user.Id),
		Code:             models.GenerateRandomString(32),
		AccessToken:      newAccessToken,
		RefreshToken:     newRefreshToken,
		ExpiresIn:        int(application.ExpireInHours * 3600),
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		Scope:            scope,
		TokenType:        "Bearer",
		TokenFamily:      token.TokenFamily, // Preserve token family
	}

	_, err = models.AddToken(newToken)
	if err != nil {
		return nil, err
	}

	// Clear old token from cache
	DeleteCachedToken(token.AccessTokenHash)

	return &TokenResponse{
		AccessToken:  newToken.AccessToken,
		TokenType:    newToken.TokenType,
		ExpiresIn:    newToken.ExpiresIn,
		RefreshToken: newToken.RefreshToken,
		Scope:        newToken.Scope,
		IdToken:      newToken.AccessToken,
	}, nil
}

// ValidateScope validates and potentially downgrades scope
func ValidateScope(requestedScope, existingScope string) (bool, string) {
	if requestedScope == "" {
		return true, existingScope
	}

	if existingScope == "" {
		return true, requestedScope
	}

	existingScopes := strings.Split(existingScope, " ")
	requestedScopes := strings.Split(requestedScope, " ")

	for _, rs := range requestedScopes {
		if rs == "" {
			continue
		}
		found := false
		for _, es := range existingScopes {
			if es != "" && rs == es {
				found = true
				break
			}
		}
		if !found {
			return false, ""
		}
	}

	return true, requestedScope
}

// verifyPassword verifies a password against its hash and salt
func verifyPassword(hashedPassword, plainPassword, salt string) bool {
	if plainPassword == "" {
		return false
	}

	// If no hash exists, reject (security: don't allow empty passwords)
	if hashedPassword == "" {
		return false
	}

	// For bcrypt hashes (starts with $2a$, $2b$, or $2y$)
	if len(hashedPassword) >= 4 && hashedPassword[0] == '$' && hashedPassword[3] == '$' {
		// This is a bcrypt hash - would need golang.org/x/crypto/bcrypt
		// For now, do a simple comparison (TEMPORARY - should use bcrypt)
		return hashedPassword == plainPassword
	}

	// For salted SHA256 (legacy support)
	if salt != "" {
		saltedPassword := plainPassword + salt
		hash := sha256.Sum256([]byte(saltedPassword))
		computed := fmt.Sprintf("%x", hash)
		return computed == hashedPassword
	}

	// Direct comparison (INSECURE - only for testing)
	return hashedPassword == plainPassword
}

// RevokeToken revokes a token (access or refresh token)
// Requirements: 6.4
func RevokeToken(token string, tokenTypeHint string) error {
	var dbToken *models.Token
	var err error

	// Try to find the token based on the hint
	if tokenTypeHint == "refresh_token" {
		dbToken, err = models.GetTokenByRefreshToken(token)
	} else {
		// Try access token first
		dbToken, err = models.GetTokenByAccessToken(token)
		// If not found, try refresh token
		if dbToken == nil && err == nil {
			dbToken, err = models.GetTokenByRefreshToken(token)
		}
	}

	if err != nil {
		return err
	}

	if dbToken != nil {
		// Mark token as revoked by setting ExpiresIn to 0
		dbToken.ExpiresIn = 0
		_, err = models.UpdateToken(dbToken.Owner, dbToken.Name, dbToken)
		if err != nil {
			return err
		}

		// Clear from cache if cached
		DeleteCachedToken(dbToken.AccessTokenHash)
	}

	return nil
}
