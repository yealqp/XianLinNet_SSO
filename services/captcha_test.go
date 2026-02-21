// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestVerifyCaptcha_Disabled tests captcha verification when CAPTCHA_ENABLED is false
// Requirements: 4.2
func TestVerifyCaptcha_Disabled(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("CAPTCHA_ENABLED")
	defer os.Setenv("CAPTCHA_ENABLED", originalEnabled)

	// Test with CAPTCHA_ENABLED = "false"
	os.Setenv("CAPTCHA_ENABLED", "false")
	result, err := VerifyCaptcha("any-token")
	if err != nil {
		t.Errorf("Expected no error when captcha is disabled, got: %v", err)
	}
	if !result {
		t.Error("Expected verification to pass when captcha is disabled")
	}

	// Test with CAPTCHA_ENABLED not set
	os.Unsetenv("CAPTCHA_ENABLED")
	result, err = VerifyCaptcha("any-token")
	if err != nil {
		t.Errorf("Expected no error when captcha is not set, got: %v", err)
	}
	if !result {
		t.Error("Expected verification to pass when captcha is not set")
	}
}

// TestVerifyCaptcha_ValidToken tests captcha verification with a valid token
// Requirements: 4.1, 4.3
func TestVerifyCaptcha_ValidToken(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("CAPTCHA_ENABLED")
	originalInstanceUrl := os.Getenv("CAPTCHA_INSTANCE_URL")
	originalSiteKey := os.Getenv("CAPTCHA_SITE_KEY")
	originalSecret := os.Getenv("CAPTCHA_SECRET")
	defer func() {
		os.Setenv("CAPTCHA_ENABLED", originalEnabled)
		os.Setenv("CAPTCHA_INSTANCE_URL", originalInstanceUrl)
		os.Setenv("CAPTCHA_SITE_KEY", originalSiteKey)
		os.Setenv("CAPTCHA_SECRET", originalSecret)
	}()

	// Create mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got: %s", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var reqBody CaptchaVerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify request body
		if reqBody.Secret != "test-secret" {
			t.Errorf("Expected secret 'test-secret', got: %s", reqBody.Secret)
		}
		if reqBody.Response != "valid-token" {
			t.Errorf("Expected response 'valid-token', got: %s", reqBody.Response)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CaptchaVerifyResponse{Success: true})
	}))
	defer mockServer.Close()

	// Set environment variables
	os.Setenv("CAPTCHA_ENABLED", "true")
	os.Setenv("CAPTCHA_INSTANCE_URL", mockServer.URL)
	os.Setenv("CAPTCHA_SITE_KEY", "test-site-key")
	os.Setenv("CAPTCHA_SECRET", "test-secret")

	// Test verification
	result, err := VerifyCaptcha("valid-token")
	if err != nil {
		t.Errorf("Expected no error for valid token, got: %v", err)
	}
	if !result {
		t.Error("Expected verification to pass for valid token")
	}
}

// TestVerifyCaptcha_InvalidToken tests captcha verification with an invalid token
// Requirements: 4.1, 4.4
func TestVerifyCaptcha_InvalidToken(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("CAPTCHA_ENABLED")
	originalInstanceUrl := os.Getenv("CAPTCHA_INSTANCE_URL")
	originalSiteKey := os.Getenv("CAPTCHA_SITE_KEY")
	originalSecret := os.Getenv("CAPTCHA_SECRET")
	defer func() {
		os.Setenv("CAPTCHA_ENABLED", originalEnabled)
		os.Setenv("CAPTCHA_INSTANCE_URL", originalInstanceUrl)
		os.Setenv("CAPTCHA_SITE_KEY", originalSiteKey)
		os.Setenv("CAPTCHA_SECRET", originalSecret)
	}()

	// Create mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return failure response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CaptchaVerifyResponse{Success: false})
	}))
	defer mockServer.Close()

	// Set environment variables
	os.Setenv("CAPTCHA_ENABLED", "true")
	os.Setenv("CAPTCHA_INSTANCE_URL", mockServer.URL)
	os.Setenv("CAPTCHA_SITE_KEY", "test-site-key")
	os.Setenv("CAPTCHA_SECRET", "test-secret")

	// Test verification
	result, err := VerifyCaptcha("invalid-token")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result {
		t.Error("Expected verification to fail for invalid token")
	}
}

// TestVerifyCaptcha_IncompleteConfig tests captcha verification with incomplete configuration
// Requirements: 4.1
func TestVerifyCaptcha_IncompleteConfig(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("CAPTCHA_ENABLED")
	originalInstanceUrl := os.Getenv("CAPTCHA_INSTANCE_URL")
	originalSiteKey := os.Getenv("CAPTCHA_SITE_KEY")
	originalSecret := os.Getenv("CAPTCHA_SECRET")
	defer func() {
		os.Setenv("CAPTCHA_ENABLED", originalEnabled)
		os.Setenv("CAPTCHA_INSTANCE_URL", originalInstanceUrl)
		os.Setenv("CAPTCHA_SITE_KEY", originalSiteKey)
		os.Setenv("CAPTCHA_SECRET", originalSecret)
	}()

	// Test with missing CAPTCHA_INSTANCE_URL
	os.Setenv("CAPTCHA_ENABLED", "true")
	os.Unsetenv("CAPTCHA_INSTANCE_URL")
	os.Setenv("CAPTCHA_SITE_KEY", "test-site-key")
	os.Setenv("CAPTCHA_SECRET", "test-secret")

	result, err := VerifyCaptcha("any-token")
	if err == nil {
		t.Error("Expected error for incomplete configuration")
	}
	if result {
		t.Error("Expected verification to fail for incomplete configuration")
	}

	// Test with missing CAPTCHA_SITE_KEY
	os.Setenv("CAPTCHA_INSTANCE_URL", "http://example.com")
	os.Unsetenv("CAPTCHA_SITE_KEY")
	os.Setenv("CAPTCHA_SECRET", "test-secret")

	result, err = VerifyCaptcha("any-token")
	if err == nil {
		t.Error("Expected error for incomplete configuration")
	}
	if result {
		t.Error("Expected verification to fail for incomplete configuration")
	}

	// Test with missing CAPTCHA_SECRET
	os.Setenv("CAPTCHA_SITE_KEY", "test-site-key")
	os.Unsetenv("CAPTCHA_SECRET")

	result, err = VerifyCaptcha("any-token")
	if err == nil {
		t.Error("Expected error for incomplete configuration")
	}
	if result {
		t.Error("Expected verification to fail for incomplete configuration")
	}
}

// TestVerifyCaptcha_NetworkError tests captcha verification with network errors
// Requirements: 4.4
func TestVerifyCaptcha_NetworkError(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("CAPTCHA_ENABLED")
	originalInstanceUrl := os.Getenv("CAPTCHA_INSTANCE_URL")
	originalSiteKey := os.Getenv("CAPTCHA_SITE_KEY")
	originalSecret := os.Getenv("CAPTCHA_SECRET")
	defer func() {
		os.Setenv("CAPTCHA_ENABLED", originalEnabled)
		os.Setenv("CAPTCHA_INSTANCE_URL", originalInstanceUrl)
		os.Setenv("CAPTCHA_SITE_KEY", originalSiteKey)
		os.Setenv("CAPTCHA_SECRET", originalSecret)
	}()

	// Set environment variables with invalid URL
	os.Setenv("CAPTCHA_ENABLED", "true")
	os.Setenv("CAPTCHA_INSTANCE_URL", "http://invalid-url-that-does-not-exist.local")
	os.Setenv("CAPTCHA_SITE_KEY", "test-site-key")
	os.Setenv("CAPTCHA_SECRET", "test-secret")

	// Test verification
	result, err := VerifyCaptcha("any-token")
	if err == nil {
		t.Error("Expected error for network failure")
	}
	if result {
		t.Error("Expected verification to fail for network error")
	}
}

// TestVerifyCaptcha_InvalidResponse tests captcha verification with invalid server response
// Requirements: 4.4
func TestVerifyCaptcha_InvalidResponse(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("CAPTCHA_ENABLED")
	originalInstanceUrl := os.Getenv("CAPTCHA_INSTANCE_URL")
	originalSiteKey := os.Getenv("CAPTCHA_SITE_KEY")
	originalSecret := os.Getenv("CAPTCHA_SECRET")
	defer func() {
		os.Setenv("CAPTCHA_ENABLED", originalEnabled)
		os.Setenv("CAPTCHA_INSTANCE_URL", originalInstanceUrl)
		os.Setenv("CAPTCHA_SITE_KEY", originalSiteKey)
		os.Setenv("CAPTCHA_SECRET", originalSecret)
	}()

	// Create mock HTTP server that returns invalid JSON
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer mockServer.Close()

	// Set environment variables
	os.Setenv("CAPTCHA_ENABLED", "true")
	os.Setenv("CAPTCHA_INSTANCE_URL", mockServer.URL)
	os.Setenv("CAPTCHA_SITE_KEY", "test-site-key")
	os.Setenv("CAPTCHA_SECRET", "test-secret")

	// Test verification
	result, err := VerifyCaptcha("any-token")
	if err == nil {
		t.Error("Expected error for invalid response")
	}
	if result {
		t.Error("Expected verification to fail for invalid response")
	}
}

// TestVerifyCaptcha_CaseInsensitiveEnabled tests that CAPTCHA_ENABLED is case-insensitive
// Requirements: 4.1, 4.2
func TestVerifyCaptcha_CaseInsensitiveEnabled(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("CAPTCHA_ENABLED")
	originalInstanceUrl := os.Getenv("CAPTCHA_INSTANCE_URL")
	originalSiteKey := os.Getenv("CAPTCHA_SITE_KEY")
	originalSecret := os.Getenv("CAPTCHA_SECRET")
	defer func() {
		os.Setenv("CAPTCHA_ENABLED", originalEnabled)
		os.Setenv("CAPTCHA_INSTANCE_URL", originalInstanceUrl)
		os.Setenv("CAPTCHA_SITE_KEY", originalSiteKey)
		os.Setenv("CAPTCHA_SECRET", originalSecret)
	}()

	// Create mock HTTP server for enabled tests
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CaptchaVerifyResponse{Success: true})
	}))
	defer mockServer.Close()

	testCases := []struct {
		name            string
		value           string
		shouldBeEnabled bool
	}{
		{"lowercase true", "true", true},
		{"uppercase TRUE", "TRUE", true},
		{"mixed case True", "True", true},
		{"lowercase false", "false", false},
		{"uppercase FALSE", "FALSE", false},
		{"mixed case False", "False", false},
		{"empty string", "", false},
		{"random value", "yes", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("CAPTCHA_ENABLED", tc.value)

			if tc.shouldBeEnabled {
				// Set up mock configuration for enabled tests
				os.Setenv("CAPTCHA_INSTANCE_URL", mockServer.URL)
				os.Setenv("CAPTCHA_SITE_KEY", "test-site-key")
				os.Setenv("CAPTCHA_SECRET", "test-secret")
			}

			result, err := VerifyCaptcha("any-token")

			if tc.shouldBeEnabled {
				// When enabled, should call the server and get result
				if err != nil {
					t.Errorf("Expected no error when enabled, got: %v", err)
				}
				if !result {
					t.Error("Expected verification to pass when enabled with valid mock server")
				}
			} else {
				// When disabled, should always pass without error
				if err != nil {
					t.Errorf("Expected no error when disabled, got: %v", err)
				}
				if !result {
					t.Error("Expected verification to pass when disabled")
				}
			}
		})
	}
}
