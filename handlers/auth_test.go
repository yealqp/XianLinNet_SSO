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

// TestHandleLogin tests the login handler parameter validation
func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Missing email",
			requestBody: types.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing password",
			requestBody: types.LoginRequest{
				Email:    "test@example.com",
				Password: "",
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
			app.Post("/login", HandleLogin())

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

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
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

// TestHandleRegister tests the register handler parameter validation
func TestHandleRegister(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Missing username",
			requestBody: types.RegisterRequest{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing email",
			requestBody: types.RegisterRequest{
				Username: "testuser",
				Email:    "",
				Password: "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing password",
			requestBody: types.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Username too short",
			requestBody: types.RegisterRequest{
				Username: "ab",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Password too short",
			requestBody: types.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "12345",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/register", HandleRegister())

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
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

// TestHandleSendVerificationCode tests the send verification code handler
func TestHandleSendVerificationCode(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid email",
			requestBody: SendVerificationCodeRequest{
				Email: "test@example.com",
			},
			expectedStatus: fiber.StatusOK,
			expectedError:  false,
		},
		{
			name: "Missing email",
			requestBody: SendVerificationCodeRequest{
				Email: "",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/send-code", HandleSendVerificationCode())

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest("POST", "/send-code", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestHandleResetPassword tests the reset password handler parameter validation
func TestHandleResetPassword(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Missing email",
			requestBody: ResetPasswordRequest{
				Email:       "",
				Code:        "123456",
				NewPassword: "newpassword123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing code",
			requestBody: ResetPasswordRequest{
				Email:       "test@example.com",
				Code:        "",
				NewPassword: "newpassword123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Password too short",
			requestBody: ResetPasswordRequest{
				Email:       "test@example.com",
				Code:        "123456",
				NewPassword: "12345",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/reset-password", HandleResetPassword())

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest("POST", "/reset-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestHandleUpdateProfile tests the update profile handler parameter validation
func TestHandleUpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		userID         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Missing userID",
			requestBody: UpdateProfileRequest{
				Username: "newusername",
			},
			userID:         "",
			expectedStatus: fiber.StatusUnauthorized,
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

			app.Post("/update-profile", HandleUpdateProfile())

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest("POST", "/update-profile", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestHandleGetApplicationInfo tests the get application info handler parameter validation
func TestHandleGetApplicationInfo(t *testing.T) {
	tests := []struct {
		name           string
		clientID       string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Missing client_id",
			clientID:       "",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/application-info", HandleGetApplicationInfo())

			url := "/application-info"
			if tt.clientID != "" {
				url += "?client_id=" + tt.clientID
			}

			req := httptest.NewRequest("GET", url, nil)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestHandleAuthorize tests the authorize handler parameter validation
func TestHandleAuthorize(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		userID         string
		queryParams    map[string]string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Missing client_id",
			method: "GET",
			userID: "1",
			queryParams: map[string]string{
				"redirect_uri": "http://localhost:3000/callback",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "Missing redirect_uri",
			method: "GET",
			userID: "1",
			queryParams: map[string]string{
				"client_id": "test-client",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "Missing userID (not authenticated)",
			method: "GET",
			userID: "",
			queryParams: map[string]string{
				"client_id":    "test-client",
				"redirect_uri": "http://localhost:3000/callback",
			},
			expectedStatus: fiber.StatusUnauthorized,
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

			app.Get("/authorize", HandleAuthorize())
			app.Post("/authorize", HandleAuthorize())

			url := "/authorize?"
			for key, value := range tt.queryParams {
				url += key + "=" + value + "&"
			}

			req := httptest.NewRequest(tt.method, url, nil)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
