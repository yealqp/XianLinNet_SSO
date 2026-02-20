// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package main

import (
	"fmt"
	"os"

	"github.com/beego/beego/v2/server/web"
	"github.com/oauth-server/oauth-server/models"
	"github.com/oauth-server/oauth-server/routers"
	"github.com/oauth-server/oauth-server/services"
)

func main() {
	// Initialize database
	err := models.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	// Auto-sync tables on startup
	err = models.InitTables()
	if err != nil {
		fmt.Printf("Warning: Failed to sync tables: %v\n", err)
	}

	// Auto-initialize data if needed (creates admin if none exists)
	err = models.InitData()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize data: %v\n", err)
	}

	// Initialize RSA keys for encryption
	err = services.InitRSAKeys()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize RSA keys: %v", err))
	}

	// Initialize Redis (optional)
	err = services.InitRedis()
	if err != nil {
		fmt.Printf("Warning: Redis not available: %v\n", err)
		fmt.Println("Continuing without Redis cache...")
	} else {
		fmt.Println("Redis cache initialized successfully!")
	}

	// Check if init command
	if len(os.Args) > 1 && os.Args[1] == "init" {
		fmt.Println("Database initialization completed!")
		return
	}

	// Initialize routers
	routers.Init()

	// Start server
	port := web.AppConfig.DefaultString("httpport", "8080")
	fmt.Printf("OAuth Server starting on port %s...\n", port)
	web.Run()
}
