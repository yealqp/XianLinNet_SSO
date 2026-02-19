// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateClientId generates a random client ID
func GenerateClientId() string {
	return uuid.New().String()
}

// GenerateClientSecret generates a random client secret
func GenerateClientSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GenerateRandomString generates a random string
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// GetCurrentTime returns current time in RFC3339 format
func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

// GetId returns formatted ID
func GetId(owner, name string) string {
	return fmt.Sprintf("%s/%s", owner, name)
}

// ParseId parses owner and name from ID
func ParseId(id string) (string, string, error) {
	var owner, name string
	_, err := fmt.Sscanf(id, "%s/%s", &owner, &name)
	if err != nil {
		return "", "", err
	}
	return owner, name, nil
}
