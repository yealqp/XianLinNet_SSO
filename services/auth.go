// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"fmt"
	"time"

	"github.com/oauth-server/oauth-server/models"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword checks if the provided password matches the hashed password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	// Simple email validation
	if len(email) < 3 || len(email) > 100 {
		return false
	}
	// Check for @ symbol
	atCount := 0
	for _, c := range email {
		if c == '@' {
			atCount++
		}
	}
	return atCount == 1
}

// RegisterUser registers a new user with email
func RegisterUser(email, password, username string) (*models.User, error) {
	// Validate email
	if !ValidateEmail(email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Check if user already exists
	existingUser, err := models.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	now := time.Now().Format(time.RFC3339)
	user := &models.User{
		Owner:       "built-in",
		CreatedTime: now,
		UpdatedTime: now,
		Type:        "normal-user",
		Password:    hashedPassword,
		Username:    username,
		Email:       email,
		IsRealName:  false, // Default to not real-name verified
		IsAdmin:     false,
		IsForbidden: false,
		IsDeleted:   false,
	}

	// Save user to database
	_, err = models.AddUser(user)
	if err != nil {
		return nil, err
	}

	// Assign default role: unverified-user
	_, err = models.AddUserRole(user.Id, "admin", "unverified-user")
	if err != nil {
		// Log error but don't fail registration
		fmt.Printf("Warning: Failed to assign default role to user: %v\n", err)
	}

	return user, nil
}

// LoginUser authenticates a user with email and password
func LoginUser(email, password string) (*models.User, error) {
	// Validate email
	if !ValidateEmail(email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Get user by email
	user, err := models.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if user is forbidden
	if user.IsForbidden {
		return nil, fmt.Errorf("account is disabled")
	}

	// Check if user is deleted
	if user.IsDeleted {
		return nil, fmt.Errorf("account not found")
	}

	// Check password
	if !CheckPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	return user, nil
}

// ResetPassword resets user password
func ResetPassword(email, newPassword string) error {
	// Validate email
	if !ValidateEmail(email) {
		return fmt.Errorf("invalid email format")
	}

	// Get user by email
	user, err := models.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	user.Password = hashedPassword
	user.UpdatedTime = time.Now().Format(time.RFC3339)

	_, err = models.UpdateUser(user.Id, user)
	if err != nil {
		return err
	}

	return nil
}
