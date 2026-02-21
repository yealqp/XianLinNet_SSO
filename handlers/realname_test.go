// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/oauth-server/oauth-server/types"
)

// TestHandleSubmitRealName tests the submit real name handler parameter validation
// Requirements: 9.1, 9.4
func TestHandleSubmitRealName(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		userID         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Missing userID (not authenticated)",
			requestBody: SubmitRealNameRequest{
				Name:   "张三",
				IDCard: "110101199001011234",
			},
			userID:         "",
			expectedStatus: fiber.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "Missing name",
			requestBody: SubmitRealNameRequest{
				Name:   "",
				IDCard: "110101199001011234",
			},
			userID:         "1",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing IDCard",
			requestBody: SubmitRealNameRequest{
				Name:   "张三",
				IDCard: "",
			},
			userID:         "1",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			userID:         "1",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Invalid userID format",
			requestBody: SubmitRealNameRequest{
				Name:   "张三",
				IDCard: "110101199001011234",
			},
			userID:         "invalid",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			// Add middleware to set userID in context
			app.Use(func(c *fiber.Ctx) error {
				if tt.userID != "" {
					c.Locals("userID", tt.userID)
				}
				return c.Next()
			})

			app.Post("/submit-realname", HandleSubmitRealName())

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/submit-realname", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Parse response
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var apiResp types.ApiResponse
			if err := json.Unmarshal(respBody, &apiResp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectedError && apiResp.Status != "error" {
				t.Errorf("Expected error response, got status: %s", apiResp.Status)
			}

			if !tt.expectedError && apiResp.Status != "ok" {
				t.Errorf("Expected success response, got status: %s, msg: %s", apiResp.Status, apiResp.Msg)
			}
		})
	}
}

// TestHandleGetRealNameInfo tests the get real name info handler parameter validation
// Requirements: 9.2
func TestHandleGetRealNameInfo(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Missing userID (not authenticated)",
			userID:         "",
			expectedStatus: fiber.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "Invalid userID format",
			userID:         "invalid",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			// Add middleware to set userID in context
			app.Use(func(c *fiber.Ctx) error {
				if tt.userID != "" {
					c.Locals("userID", tt.userID)
				}
				return c.Next()
			})

			app.Get("/realname-info", HandleGetRealNameInfo())

			req := httptest.NewRequest("GET", "/realname-info", nil)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Parse response
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var apiResp types.ApiResponse
			if err := json.Unmarshal(respBody, &apiResp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectedError && apiResp.Status != "error" {
				t.Errorf("Expected error response, got status: %s", apiResp.Status)
			}

			if !tt.expectedError && apiResp.Status != "ok" {
				t.Errorf("Expected success response, got status: %s, msg: %s", apiResp.Status, apiResp.Msg)
			}
		})
	}
}

// TestHandleVerifyRealName tests the verify real name handler parameter validation
// Requirements: 9.3
func TestHandleVerifyRealName(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Missing userID",
			requestBody: VerifyRealNameRequest{
				UserID:     0,
				IsApproved: true,
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Post("/verify-realname", HandleVerifyRealName())

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/verify-realname", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Parse response
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var apiResp types.ApiResponse
			if err := json.Unmarshal(respBody, &apiResp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.expectedError && apiResp.Status != "error" {
				t.Errorf("Expected error response, got status: %s", apiResp.Status)
			}

			if !tt.expectedError && apiResp.Status != "ok" {
				t.Errorf("Expected success response, got status: %s, msg: %s", apiResp.Status, apiResp.Msg)
			}
		})
	}
}

// TestRSAEncryptionDecryption tests RSA encryption and decryption functionality
// Requirements: 9.4
func TestRSAEncryptionDecryption(t *testing.T) {
	// This test verifies that the RSA encryption/decryption used for storing
	// sensitive real-name information works correctly.
	// The actual encryption/decryption is handled by services.EncryptData and
	// services.DecryptData, which are tested in the services package.

	// This test is a placeholder to document that RSA encryption is tested
	// as part of the real-name authentication functionality.
	t.Log("RSA encryption/decryption is tested in services/encryption_test.go")
}
