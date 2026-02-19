# OAuth Server Build Script for Windows
# Copyright 2024 OAuth Server Authors.

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

function Show-Help {
    Write-Host "OAuth Server - Build Script" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [command]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Green
    Write-Host "  help        - Show this help message"
    Write-Host "  install     - Install dependencies"
    Write-Host "  init        - Initialize database"
    Write-Host "  build       - Build the executable"
    Write-Host "  run         - Build and run the server"
    Write-Host "  dev         - Run in development mode"
    Write-Host "  test        - Run tests"
    Write-Host "  clean       - Clean build artifacts"
    Write-Host "  fmt         - Format code"
    Write-Host ""
}

function Install-Dependencies {
    Write-Host "Installing dependencies..." -ForegroundColor Cyan
    go mod download
    go mod tidy
    Write-Host "Dependencies installed!" -ForegroundColor Green
}

function Initialize-Database {
    Write-Host "Initializing database..." -ForegroundColor Cyan
    
    # Create config if not exists
    if (-not (Test-Path "conf\app.conf")) {
        Write-Host "Creating conf\app.conf from example..." -ForegroundColor Yellow
        Copy-Item "conf\app.conf.example" "conf\app.conf"
    }
    
    go run main.go init
    Write-Host "Database initialized!" -ForegroundColor Green
}

function Build-Server {
    Write-Host "Building OAuth Server..." -ForegroundColor Cyan
    go build -o oauth-server.exe main.go
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build complete: oauth-server.exe" -ForegroundColor Green
        $size = (Get-Item oauth-server.exe).Length / 1MB
        Write-Host ("Size: {0:N2} MB" -f $size) -ForegroundColor Gray
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
}

function Run-Server {
    Build-Server
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Starting OAuth Server..." -ForegroundColor Cyan
        .\oauth-server.exe
    }
}

function Run-Dev {
    Write-Host "Starting in development mode..." -ForegroundColor Cyan
    
    # Create config if not exists
    if (-not (Test-Path "conf\app.conf")) {
        Write-Host "Creating conf\app.conf from example..." -ForegroundColor Yellow
        Copy-Item "conf\app.conf.example" "conf\app.conf"
    }
    
    go run main.go
}

function Run-Tests {
    Write-Host "Running tests..." -ForegroundColor Cyan
    go test -v ./...
}

function Clean-Build {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan
    
    if (Test-Path "oauth-server.exe") {
        Remove-Item "oauth-server.exe" -Force
        Write-Host "Removed oauth-server.exe" -ForegroundColor Gray
    }
    
    Get-ChildItem -Filter "*.db" | Remove-Item -Force
    Get-ChildItem -Filter "*.log" | Remove-Item -Force
    
    Write-Host "Clean complete!" -ForegroundColor Green
}

function Format-Code {
    Write-Host "Formatting code..." -ForegroundColor Cyan
    go fmt ./...
    Write-Host "Code formatted!" -ForegroundColor Green
}

# Main script logic
switch ($Command.ToLower()) {
    "help" { Show-Help }
    "install" { Install-Dependencies }
    "init" { Initialize-Database }
    "build" { Build-Server }
    "run" { Run-Server }
    "dev" { Run-Dev }
    "test" { Run-Tests }
    "clean" { Clean-Build }
    "fmt" { Format-Code }
    default {
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Write-Host ""
        Show-Help
        exit 1
    }
}
