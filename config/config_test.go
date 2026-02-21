package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Test with default values
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify default values
	if config.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Server.Port)
	}

	if config.Database.Host != "localhost" {
		t.Errorf("Expected default DB host localhost, got %s", config.Database.Host)
	}

	if config.JWT.AccessTokenExpiry != 3600 {
		t.Errorf("Expected default access token expiry 3600, got %d", config.JWT.AccessTokenExpiry)
	}
}

func TestGetEnv(t *testing.T) {
	// Test with existing env var
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected test_value, got %s", result)
	}

	// Test with non-existing env var
	result = getEnv("NON_EXISTING_VAR", "default")
	if result != "default" {
		t.Errorf("Expected default, got %s", result)
	}
}

func TestGetIntEnv(t *testing.T) {
	// Test with valid integer
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getIntEnv("TEST_INT", 10)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test with invalid integer
	os.Setenv("TEST_INT_INVALID", "not_a_number")
	defer os.Unsetenv("TEST_INT_INVALID")

	result = getIntEnv("TEST_INT_INVALID", 10)
	if result != 10 {
		t.Errorf("Expected default 10, got %d", result)
	}

	// Test with non-existing env var
	result = getIntEnv("NON_EXISTING_INT", 10)
	if result != 10 {
		t.Errorf("Expected default 10, got %d", result)
	}
}

func TestGetBoolEnv(t *testing.T) {
	// Test with true
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	result := getBoolEnv("TEST_BOOL", false)
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}

	// Test with false
	os.Setenv("TEST_BOOL_FALSE", "false")
	defer os.Unsetenv("TEST_BOOL_FALSE")

	result = getBoolEnv("TEST_BOOL_FALSE", true)
	if result != false {
		t.Errorf("Expected false, got %v", result)
	}

	// Test with invalid boolean
	os.Setenv("TEST_BOOL_INVALID", "not_a_bool")
	defer os.Unsetenv("TEST_BOOL_INVALID")

	result = getBoolEnv("TEST_BOOL_INVALID", true)
	if result != true {
		t.Errorf("Expected default true, got %v", result)
	}
}

func TestGetDurationEnv(t *testing.T) {
	// Test with valid duration
	os.Setenv("TEST_DURATION", "5s")
	defer os.Unsetenv("TEST_DURATION")

	result := getDurationEnv("TEST_DURATION", 10*time.Second)
	if result != 5*time.Second {
		t.Errorf("Expected 5s, got %v", result)
	}

	// Test with invalid duration
	os.Setenv("TEST_DURATION_INVALID", "not_a_duration")
	defer os.Unsetenv("TEST_DURATION_INVALID")

	result = getDurationEnv("TEST_DURATION_INVALID", 10*time.Second)
	if result != 10*time.Second {
		t.Errorf("Expected default 10s, got %v", result)
	}
}

func TestConfigStructure(t *testing.T) {
	// Set some environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DB_HOST", "testdb")
	os.Setenv("JWT_SECRET", "test_secret")
	os.Setenv("CAPTCHA_ENABLED", "true")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("CAPTCHA_ENABLED")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify custom values
	if config.Server.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", config.Server.Port)
	}

	if config.Database.Host != "testdb" {
		t.Errorf("Expected DB host testdb, got %s", config.Database.Host)
	}

	if config.JWT.Secret != "test_secret" {
		t.Errorf("Expected JWT secret test_secret, got %s", config.JWT.Secret)
	}

	if config.Captcha.Enabled != true {
		t.Errorf("Expected captcha enabled true, got %v", config.Captcha.Enabled)
	}
}
