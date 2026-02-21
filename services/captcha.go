// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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
// Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6
func VerifyCaptcha(token string) (bool, error) {
	// Check CAPTCHA_ENABLED environment variable (Requirement 4.1, 4.2)
	enabled := os.Getenv("CAPTCHA_ENABLED")
	if strings.ToLower(enabled) != "true" {
		// If captcha is disabled or not set, always return true
		return true, nil
	}

	// Read Cap.js configuration from environment variables (Requirement 4.6)
	instanceUrl := os.Getenv("CAPTCHA_INSTANCE_URL")
	siteKey := os.Getenv("CAPTCHA_SITE_KEY")
	secret := os.Getenv("CAPTCHA_SECRET")

	if instanceUrl == "" || siteKey == "" || secret == "" {
		return false, fmt.Errorf("captcha configuration is incomplete")
	}

	// Build verification URL (Requirement 4.5)
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

	// Create HTTP client with timeout (handle network errors gracefully)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send POST request to Cap.js server (Requirement 4.5)
	resp, err := client.Post(verifyUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Handle network errors (Requirement 4.4)
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

	// Return verification result (Requirement 4.3, 4.4)
	return result.Success, nil
}
