// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package models

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/xorm-io/xorm"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var engine *xorm.Engine

func InitDB() error {
	driverName, _ := web.AppConfig.String("driverName")
	dataSourceName, _ := web.AppConfig.String("dataSourceName")

	if driverName == "" {
		driverName = "sqlite"
		dataSourceName = "./oauth_server.db"
	}

	var err error
	engine, err = xorm.NewEngine(driverName, dataSourceName)
	if err != nil {
		return err
	}

	engine.SetMaxIdleConns(10)
	engine.SetMaxOpenConns(100)
	engine.SetConnMaxLifetime(time.Hour)

	// Show SQL for debugging
	engine.ShowSQL(true)

	return engine.Ping()
}

func InitTables() error {
	return engine.Sync2(
		new(User),
		new(Application),
		new(Token),
		new(Organization),
		new(Provider),
		new(Role),
		new(Permission),
		new(RolePermission),
		new(UserRole),
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

	// Initialize default roles and permissions
	err = initializeRolesAndPermissions()
	if err != nil {
		return err
	}

	// Assign admin role to admin user (legacy check - now handled in createDefaultAdmin)
	// This code is kept for backward compatibility with existing databases
	var adminUser User
	has, err := engine.Where("is_admin = ?", true).Get(&adminUser)
	if err != nil {
		return err
	}
	if has {
		// Check if admin role is already assigned
		existsUserRole, err := engine.Where("user_id = ? AND role_owner = ? AND role_name = ?",
			adminUser.Id, "admin", "admin").Exist(&UserRole{})
		if err != nil {
			return err
		}
		// If not exists, add it
		if !existsUserRole {
			_, err = AddUserRole(adminUser.Id, "admin", "admin")
			if err != nil {
				fmt.Printf("Warning: Failed to assign admin role: %v\n", err)
			}
		}
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
	// Get admin credentials from config (REQUIRED)
	adminEmail, _ := web.AppConfig.String("adminEmail")
	adminPassword, _ := web.AppConfig.String("adminPassword")
	adminUsername, _ := web.AppConfig.String("adminUsername")

	// Validate required configuration
	if adminEmail == "" {
		return fmt.Errorf("adminEmail is not configured in app.conf. Please set adminEmail in the configuration file")
	}
	if adminPassword == "" {
		return fmt.Errorf("adminPassword is not configured in app.conf. Please set adminPassword in the configuration file")
	}
	if adminUsername == "" {
		return fmt.Errorf("adminUsername is not configured in app.conf. Please set adminUsername in the configuration file")
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

	// Assign admin role to admin user
	_, err = AddUserRole(admin.Id, "admin", "admin")
	if err != nil {
		fmt.Printf("Warning: Failed to assign admin role: %v\n", err)
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

func initializeRolesAndPermissions() error {
	// Create default permissions
	defaultPermissions := []Permission{
		// User permissions
		{Owner: "admin", Name: "user-read", DisplayName: "Read Users", Resource: "user", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: "admin", Name: "user-write", DisplayName: "Write Users", Resource: "user", Action: "write", Effect: "allow", IsEnabled: true},
		{Owner: "admin", Name: "user-delete", DisplayName: "Delete Users", Resource: "user", Action: "delete", Effect: "allow", IsEnabled: true},

		// Application permissions
		{Owner: "admin", Name: "app-read", DisplayName: "Read Applications", Resource: "application", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: "admin", Name: "app-write", DisplayName: "Write Applications", Resource: "application", Action: "write", Effect: "allow", IsEnabled: true},
		{Owner: "admin", Name: "app-delete", DisplayName: "Delete Applications", Resource: "application", Action: "delete", Effect: "allow", IsEnabled: true},

		// Token permissions
		{Owner: "admin", Name: "token-read", DisplayName: "Read Tokens", Resource: "token", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: "admin", Name: "token-revoke", DisplayName: "Revoke Tokens", Resource: "token", Action: "revoke", Effect: "allow", IsEnabled: true},

		// Organization permissions
		{Owner: "admin", Name: "org-read", DisplayName: "Read Organizations", Resource: "organization", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: "admin", Name: "org-write", DisplayName: "Write Organizations", Resource: "organization", Action: "write", Effect: "allow", IsEnabled: true},

		// Role permissions
		{Owner: "admin", Name: "role-read", DisplayName: "Read Roles", Resource: "role", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: "admin", Name: "role-write", DisplayName: "Write Roles", Resource: "role", Action: "write", Effect: "allow", IsEnabled: true},

		// Permission permissions
		{Owner: "admin", Name: "perm-read", DisplayName: "Read Permissions", Resource: "permission", Action: "read", Effect: "allow", IsEnabled: true},
		{Owner: "admin", Name: "perm-write", DisplayName: "Write Permissions", Resource: "permission", Action: "write", Effect: "allow", IsEnabled: true},

		// Admin permission (all resources, all actions)
		{Owner: "admin", Name: "admin-all", DisplayName: "Admin All", Resource: "*", Action: "*", Effect: "allow", IsEnabled: true},
	}

	for _, perm := range defaultPermissions {
		perm.CreatedTime = time.Now().Format(time.RFC3339)
		exists, _ := engine.Get(&Permission{Owner: perm.Owner, Name: perm.Name})
		if !exists {
			engine.Insert(&perm)
		}
	}

	// Create default roles
	defaultRoles := []Role{
		{Owner: "admin", Name: "admin", DisplayName: "管理员", Description: "系统管理员，拥有所有权限", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "verified-user", DisplayName: "实名用户", Description: "已完成实名认证的用户", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "normal-user", DisplayName: "普通用户", Description: "普通注册用户", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "unverified-user", DisplayName: "未实名用户", Description: "未完成实名认证的用户", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "user-manager", DisplayName: "用户管理员", Description: "管理用户账号", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "app-manager", DisplayName: "应用管理员", Description: "管理应用配置", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "viewer", DisplayName: "只读用户", Description: "只读访问权限", Type: "system", Organization: "built-in", IsEnabled: true},
	}

	for _, role := range defaultRoles {
		role.CreatedTime = time.Now().Format(time.RFC3339)
		exists, _ := engine.Get(&Role{Owner: role.Owner, Name: role.Name})
		if !exists {
			engine.Insert(&role)
		}
	}

	// Assign permissions to roles
	rolePermissions := map[string][]string{
		"admin":           {"admin-all"},
		"verified-user":   {"user-read"},
		"normal-user":     {"user-read"},
		"unverified-user": {"user-read"},
		"user-manager":    {"user-read", "user-write", "user-delete", "role-read"},
		"app-manager":     {"app-read", "app-write", "app-delete", "token-read", "token-revoke"},
		"viewer":          {"user-read", "app-read", "token-read", "org-read", "role-read", "perm-read"},
	}

	for roleName, permNames := range rolePermissions {
		for _, permName := range permNames {
			exists, _ := engine.Get(&RolePermission{
				RoleOwner: "admin",
				RoleName:  roleName,
				PermOwner: "admin",
				PermName:  permName,
			})
			if !exists {
				rp := &RolePermission{
					RoleOwner:   "admin",
					RoleName:    roleName,
					PermOwner:   "admin",
					PermName:    permName,
					CreatedTime: time.Now().Format(time.RFC3339),
				}
				engine.Insert(rp)
			}
		}
	}

	return nil
}

func GetEngine() *xorm.Engine {
	return engine
}
