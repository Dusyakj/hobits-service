# PowerShell script for running database migrations
# Usage: .\scripts\migrate.ps1 -Action up|down -Service user-service

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("up", "down", "force", "version")]
    [string]$Action,

    [Parameter(Mandatory=$false)]
    [string]$Service = "user-service",

    [Parameter(Mandatory=$false)]
    [int]$Version = 0
)

$ErrorActionPreference = "Stop"

# Service configuration
$services = @{
    "user-service" = @{
        "path" = "services/user-service/migrations"
        "db" = "user_service"
    }
    "habits-service" = @{
        "path" = "services/habits-service/migrations"
        "db" = "habits_service"
    }
    "bad-habits-service" = @{
        "path" = "services/bad-habits-service/migrations"
        "db" = "bad_habits_service"
    }
    "notification-service" = @{
        "path" = "services/notification-service/migrations"
        "db" = "notification_service"
    }
}

# Validate service
if (-not $services.ContainsKey($Service)) {
    Write-Host "Error: Unknown service '$Service'" -ForegroundColor Red
    Write-Host "Available services: $($services.Keys -join ', ')" -ForegroundColor Yellow
    exit 1
}

$serviceConfig = $services[$Service]
$migrationsPath = Join-Path $PSScriptRoot "..\$($serviceConfig.path)" | Resolve-Path
$database = $serviceConfig.db

# Database connection string
# Use host.docker.internal to connect to localhost from Docker container
$dbHost = "host.docker.internal"
$dbPort = "5432"
$dbUser = "postgres"
$dbPassword = "postgres"
$dbName = $database
$dbUrl = "postgres://${dbUser}:${dbPassword}@${dbHost}:${dbPort}/${dbName}?sslmode=disable"

Write-Host "================================" -ForegroundColor Cyan
Write-Host "Running migrations for: $Service" -ForegroundColor Cyan
Write-Host "Action: $Action" -ForegroundColor Cyan
Write-Host "Database: $dbName" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Build docker command
$dockerCmd = "docker run --rm -v `"${migrationsPath}:/migrations`" migrate/migrate -path=/migrations -database `"$dbUrl`""

switch ($Action) {
    "up" {
        $dockerCmd += " up"
    }
    "down" {
        if ($Version -gt 0) {
            $dockerCmd += " down $Version"
        } else {
            $dockerCmd += " down"
        }
    }
    "force" {
        if ($Version -eq 0) {
            Write-Host "Error: Version is required for 'force' action" -ForegroundColor Red
            Write-Host "Usage: .\scripts\migrate.ps1 -Action force -Service $Service -Version 1" -ForegroundColor Yellow
            exit 1
        }
        $dockerCmd += " force $Version"
    }
    "version" {
        $dockerCmd += " version"
    }
}

Write-Host "Executing: $dockerCmd" -ForegroundColor Gray
Write-Host ""

# Execute migration
try {
    Invoke-Expression $dockerCmd
    Write-Host ""
    Write-Host "Migration completed successfully!" -ForegroundColor Green
} catch {
    Write-Host ""
    Write-Host "Migration failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "================================" -ForegroundColor Cyan
Write-Host "To verify migrations:" -ForegroundColor Cyan
Write-Host "docker exec -it habit-tracker-postgres psql -U postgres -d $dbName" -ForegroundColor Yellow
Write-Host ""
Write-Host "Then run: \dt" -ForegroundColor Yellow
Write-Host "================================" -ForegroundColor Cyan
