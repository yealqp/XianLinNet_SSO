// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beego/beego/v2/server/web"
)

// CaptchaVerifyRequest represents the request to verify captcha
type CaptchaVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
}

// CaptchaVerifyResponse represents the response from captcha verification
type CaptchaVerifyResponse struct {
	Success bool `json:"success"`
}

// VerifyCaptcha verifies the captcha token with Cap.js server
func VerifyCaptcha(token string) (bool, error) {
	// Get captcha configuration from app.conf
	enabled, _ := web.AppConfig.Bool("captchaEnabled")
	if !enabled {
		// If captcha is disabled, always return true
		return true, nil
	}

	instanceUrl, _ := web.AppConfig.String("captchaInstanceUrl")
	siteKey, _ := web.AppConfig.String("captchaSiteKey")
	secret, _ := web.AppConfig.String("captchaSecret")

	if instanceUrl == "" || siteKey == "" || secret == "" {
		return false, fmt.Errorf("captcha configuration is incomplete")
	}

	// Build verification URL
	verifyUrl := fmt.Sprintf("%s/%s/siteverify", instanceUrl, siteKey)

	// Prepare request body
	reqBody := CaptchaVerifyRequest{
		Secret:   secret,
		Response: token,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send POST request
	resp, err := client.Post(verifyUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to verify captcha: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var result CaptchaVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to parse response: %v", err)
	}

	return result.Success, nil
}
