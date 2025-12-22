# PowerShell script for testing Habit Tracker API

param(
    [string]$ApiUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

# Test data
$email = "test@example.com"
$username = "testuser"
$password = "testpassword123"

Write-Host "================================" -ForegroundColor Cyan
Write-Host "Testing Habit Tracker API" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# 1. Health Check
Write-Host "1. Health Check..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$ApiUrl/health" -Method Get
    Write-Host "Response: $healthResponse" -ForegroundColor Green
} catch {
    Write-Host "Health check failed: $_" -ForegroundColor Red
}
Write-Host ""

# 2. Register
Write-Host "2. Registering new user..." -ForegroundColor Yellow
$registerBody = @{
    email = $email
    username = $username
    password = $password
    first_name = "Test"
    timezone = "Europe/Moscow"
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "$ApiUrl/api/v1/auth/register" `
        -Method Post `
        -ContentType "application/json" `
        -Body $registerBody

    Write-Host "Response:" -ForegroundColor Green
    $registerResponse | ConvertTo-Json -Depth 10 | Write-Host
} catch {
    $errorDetails = $_.ErrorDetails.Message
    Write-Host "Registration failed: $errorDetails" -ForegroundColor Red
}
Write-Host ""

# 3. Login
Write-Host "3. Logging in..." -ForegroundColor Yellow
$loginBody = @{
    email_or_username = $email
    password = $password
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$ApiUrl/api/v1/auth/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $loginBody

    Write-Host "Response:" -ForegroundColor Green
    $loginResponse | ConvertTo-Json -Depth 10 | Write-Host

    $accessToken = $loginResponse.access_token

    if ([string]::IsNullOrEmpty($accessToken)) {
        Write-Host "Failed to get access token" -ForegroundColor Red
        exit 1
    }

    Write-Host ""
    Write-Host "Access Token: $accessToken" -ForegroundColor Cyan
} catch {
    $errorDetails = $_.ErrorDetails.Message
    Write-Host "Login failed: $errorDetails" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 4. Get Profile
Write-Host "4. Getting user profile..." -ForegroundColor Yellow
try {
    $headers = @{
        "Authorization" = "Bearer $accessToken"
    }

    $profileResponse = Invoke-RestMethod -Uri "$ApiUrl/api/v1/users/profile" `
        -Method Get `
        -Headers $headers

    Write-Host "Response:" -ForegroundColor Green
    $profileResponse | ConvertTo-Json -Depth 10 | Write-Host
} catch {
    $errorDetails = $_.ErrorDetails.Message
    Write-Host "Get profile failed: $errorDetails" -ForegroundColor Red
}
Write-Host ""

# 5. Logout
Write-Host "5. Logging out..." -ForegroundColor Yellow
try {
    $headers = @{
        "Authorization" = "Bearer $accessToken"
    }

    $logoutResponse = Invoke-RestMethod -Uri "$ApiUrl/api/v1/auth/logout" `
        -Method Post `
        -Headers $headers

    Write-Host "Response:" -ForegroundColor Green
    $logoutResponse | ConvertTo-Json -Depth 10 | Write-Host
} catch {
    $errorDetails = $_.ErrorDetails.Message
    Write-Host "Logout failed: $errorDetails" -ForegroundColor Red
}
Write-Host ""

Write-Host "================================" -ForegroundColor Cyan
Write-Host "All tests completed!" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

pause