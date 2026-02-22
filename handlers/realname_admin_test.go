// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestHandleAdminGetRealNameInfo tests the admin get real name info handler
func TestHandleAdminGetRealNameInfo(t *testing.T) {
	tests := []struct {
		name           string
		userId         string
		expectedStatus int
	}{
		{
			name:           "Valid user ID",
			userId:         "1",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "Invalid user ID - empty",
			userId:         "",
			expectedStatus: fiber.StatusNotFound, // Fiber returns 404 for missing params
		},
		{
			name:           "Invalid user ID - non-numeric",
			userId:         "abc",
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			// Register the handler
			app.Get("/admin/realname/:userId", HandleAdminGetRealNameInfo())

			// Create request
			url := "/admin/realname/" + tt.userId
			req := httptest.NewRequest("GET", url, nil)

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
