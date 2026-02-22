// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/oauth-server/oauth-server/config"
	"github.com/xorm-io/xorm"
	"golang.org/x/crypto/bcrypt"
)

var engine *xorm.Engine

// SlowQueryLogger 慢查询日志记录器
// 简化实现：通过 XORM 的 ShowSQL 和自定义日志来记录慢查询
var slowQueryThreshold = 100 * time.Millisecond

func InitDB() error {
	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	dbCfg := cfg.Database

	// Build PostgreSQL connection string
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.Name, dbCfg.SSLMode)

	engine, err = xorm.NewEngine("postgres", dataSourceName)
	if err != nil {
		return err
	}

	// Configure connection pool
	engine.SetMaxIdleConns(dbCfg.MaxIdleConns)
	engine.SetMaxOpenConns(dbCfg.MaxOpenConns)
	engine.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	// Note: SetConnMaxIdleTime is not available in this version of XORM

	// Set slow query threshold for logging
	if dbCfg.QueryTimeout > 0 && dbCfg.QueryTimeout < slowQueryThreshold {
		slowQueryThreshold = dbCfg.QueryTimeout / 2
	}
	log.Printf("[DB] Slow query threshold set to: %v", slowQueryThreshold)

	// Show SQL for debugging (XORM will log queries)
	engine.ShowSQL(true)

	// Test connection with timeout
	return engine.Ping()
}

func InitTables() error {
	return engine.Sync2(
		new(User),
		new(Application),
		new(Token),
		new(Organization),
		new(Provider),
	)
}

func InitData() error {
	// Create default organization
	org := &Organization{
		Owner:        "admin",
		Name:         "built-in",
		CreatedTime:  time.Now().Format(time.RFC3339),
		DisplayName:  "Built-in Organization",
		PasswordType: "bcrypt",
		EnableSignUp: true,
	}

	exists, err := engine.Get(&Organization{Owner: "admin", Name: "built-in"})
	if err != nil {
		return err
	}
	if !exists {
		_, err = engine.Insert(org)
		if err != nil {
			return err
		}
	}

	// Check if any admin user exists
	hasAdmin, err := checkAdminExists()
	if err != nil {
		return err
	}

	// If no admin exists, create default admin user
	if !hasAdmin {
		err = createDefaultAdmin()
		if err != nil {
			return err
		}
	}

	// Create default application
	app := &Application{
		Owner:                "admin",
		Name:                 "app-built-in",
		CreatedTime:          time.Now().Format(time.RFC3339),
		DisplayName:          "Built-in Application",
		Organization:         "built-in",
		ClientId:             GenerateClientId(),
		ClientSecret:         GenerateClientSecret(),
		RedirectUris:         []string{"http://localhost:3000/callback"},
		TokenFormat:          "JWT",
		ExpireInHours:        168, // Access Token: 7 days
		RefreshExpireInHours: 720, // Refresh Token: 30 days
		EnablePassword:       true,
		EnableSignUp:         true,
		GrantTypes:           []string{"authorization_code", "password", "client_credentials", "refresh_token"},
		Scopes:               []string{"openid", "profile", "email"},
	}

	exists, err = engine.Get(&Application{Owner: "admin", Name: "app-built-in"})
	if err != nil {
		return err
	}
	if !exists {
		_, err = engine.Insert(app)
		if err != nil {
			return err
		}
		fmt.Printf("Default application created:\n")
		fmt.Printf("  Client ID: %s\n", app.ClientId)
		fmt.Printf("  Client Secret: %s\n", app.ClientSecret)
	}

	return nil
}

// checkAdminExists checks if any admin user exists in the database
func checkAdminExists() (bool, error) {
	count, err := engine.Where("is_admin = ?", true).Count(&User{})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// createDefaultAdmin creates a default admin user with credentials from config
func createDefaultAdmin() error {
	// Get admin credentials from environment variables (REQUIRED)
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminUsername := os.Getenv("ADMIN_USERNAME")

	// Validate required configuration
	if adminEmail == "" {
		return fmt.Errorf("adminEmail is not configured in .env file. Please set adminEmail in the configuration file")
	}
	if adminPassword == "" {
		return fmt.Errorf("adminPassword is not configured in .env file. Please set adminPassword in the configuration file")
	}
	if adminUsername == "" {
		return fmt.Errorf("adminUsername is not configured in .env file. Please set adminUsername in the configuration file")
	}

	// Validate email format
	if !isValidEmail(adminEmail) {
		return fmt.Errorf("invalid adminEmail format: %s", adminEmail)
	}

	// Validate password strength (minimum 8 characters)
	if len(adminPassword) < 8 {
		return fmt.Errorf("adminPassword must be at least 8 characters long")
	}

	// Hash the password using bcrypt
	hashedPassword, err := hashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %v", err)
	}

	admin := &User{
		Owner:       "built-in",
		CreatedTime: time.Now().Format(time.RFC3339),
		Type:        "normal-user",
		Password:    hashedPassword,
		Username:    adminUsername,
		Email:       adminEmail,
		IsAdmin:     true,
	}

	_, err = engine.Insert(admin)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	fmt.Printf("\n===========================================\n")
	fmt.Printf("Default admin user created:\n")
	fmt.Printf("  Email: %s\n", adminEmail)
	fmt.Printf("  Password: %s (hidden for security)\n", maskPassword(adminPassword))
	fmt.Printf("  Username: %s\n", adminUsername)
	fmt.Printf("===========================================\n")
	fmt.Printf("IMPORTANT: Please change the password after first login!\n\n")

	return nil
}

// isValidEmail checks if an email address is valid
func isValidEmail(email string) bool {
	// Simple email validation
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	// Must contain @ and .
	atIndex := -1
	dotIndex := -1
	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false // Multiple @
			}
			atIndex = i
		}
		if c == '.' && atIndex != -1 {
			dotIndex = i
		}
	}
	return atIndex > 0 && dotIndex > atIndex+1 && dotIndex < len(email)-1
}

// maskPassword masks a password for display (shows first 2 and last 2 characters)
func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func GetEngine() *xorm.Engine {
	return engine
}
