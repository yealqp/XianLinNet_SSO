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

	// Create default admin user
	admin := &User{
		Owner:       "built-in",
		Name:        "admin",
		CreatedTime: time.Now().Format(time.RFC3339),
		Id:          "admin",
		Type:        "normal-user",
		Password:    "$2a$10$rQXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", // "admin"
		DisplayName: "Admin",
		Email:       "admin@example.com",
		IsAdmin:     true,
	}

	exists, err = engine.Get(&User{Owner: "built-in", Name: "admin"})
	if err != nil {
		return err
	}
	if !exists {
		_, err = engine.Insert(admin)
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
		ExpireInHours:        168,
		RefreshExpireInHours: 168,
		EnablePassword:       true,
		EnableSignUp:         true,
		GrantTypes:           []string{"authorization_code", "password", "client_credentials", "refresh_token"},
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

	// Assign admin role to admin user
	existsUserRole, err := engine.Get(&UserRole{UserOwner: "built-in", UserName: "admin", RoleOwner: "admin", RoleName: "admin"})
	if err != nil {
		return err
	}
	// If not exists, add it
	if !existsUserRole {
		adminUserRole := &UserRole{
			UserOwner:   "built-in",
			UserName:    "admin",
			RoleOwner:   "admin",
			RoleName:    "admin",
			CreatedTime: time.Now().Format(time.RFC3339),
		}
		engine.Insert(adminUserRole)
	}

	return nil
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
		{Owner: "admin", Name: "admin", DisplayName: "Administrator", Description: "Full system access", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "user-manager", DisplayName: "User Manager", Description: "Manage users", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "app-manager", DisplayName: "Application Manager", Description: "Manage applications", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "viewer", DisplayName: "Viewer", Description: "Read-only access", Type: "system", Organization: "built-in", IsEnabled: true},
		{Owner: "admin", Name: "user", DisplayName: "Regular User", Description: "Basic user access", Type: "system", Organization: "built-in", IsEnabled: true},
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
		"admin":        {"admin-all"},
		"user-manager": {"user-read", "user-write", "user-delete", "role-read"},
		"app-manager":  {"app-read", "app-write", "app-delete", "token-read", "token-revoke"},
		"viewer":       {"user-read", "app-read", "token-read", "org-read", "role-read", "perm-read"},
		"user":         {"user-read"},
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
