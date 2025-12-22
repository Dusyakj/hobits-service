# PowerShell script for generating proto files

$ErrorActionPreference = "Stop"

# Functions for colored output
function Print-Success {
    param([string]$message)
    Write-Host "[OK] $message" -ForegroundColor Green
}

function Print-Error {
    param([string]$message)
    Write-Host "[ERROR] $message" -ForegroundColor Red
}

function Print-Info {
    param([string]$message)
    Write-Host "[INFO] $message" -ForegroundColor Yellow
}

# Check if protoc is installed
try {
    $null = Get-Command protoc -ErrorAction Stop
} catch {
    Print-Error "protoc is not installed. Please install Protocol Buffers compiler."
    Write-Host "Download from: https://github.com/protocolbuffers/protobuf/releases"
    Write-Host "After downloading, add protoc.exe to your PATH"
    exit 1
}

# Check if protoc-gen-go is installed
try {
    $null = Get-Command protoc-gen-go -ErrorAction Stop
} catch {
    Print-Error "protoc-gen-go is not installed."
    Write-Host "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    Print-Success "Installed protoc-gen-go"
}

# Check if protoc-gen-go-grpc is installed
try {
    $null = Get-Command protoc-gen-go-grpc -ErrorAction Stop
} catch {
    Print-Error "protoc-gen-go-grpc is not installed."
    Write-Host "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    Print-Success "Installed protoc-gen-go-grpc"
}

Print-Info "Generating proto files..."

# Create output directories
New-Item -ItemType Directory -Force -Path "services\user-service\proto\user\v1" | Out-Null
New-Item -ItemType Directory -Force -Path "services\habits-service\proto\habits\v1" | Out-Null
New-Item -ItemType Directory -Force -Path "services\bad-habits-service\proto\bad_habits\v1" | Out-Null
New-Item -ItemType Directory -Force -Path "services\notification-service\proto\events\v1" | Out-Null

# Generate User Service protos
Print-Info "Generating user-service protos..."
Push-Location "services\user-service"
try {
    & protoc `
        --go_out=. `
        --go_opt=paths=source_relative `
        --go-grpc_out=. `
        --go-grpc_opt=paths=source_relative `
        -I ..\..\proto `
        ..\..\proto\user\v1\user.proto

    if ($LASTEXITCODE -eq 0) {
        New-Item -ItemType Directory -Force -Path "proto\user\v1" | Out-Null
        if (Test-Path "user\v1\user.pb.go") {
            Move-Item -Force "user\v1\*.pb.go" "proto\user\v1\"
            Remove-Item -Recurse -Force "user"
        }
        Pop-Location
        Print-Success "Generated user-service protos"
    } else {
        Pop-Location
        Print-Error "Failed to generate user-service protos"
        exit 1
    }
} catch {
    Pop-Location
    Print-Error "Failed to generate user-service protos: $_"
    exit 1
}

# Generate Habits Service protos (if exists)
if (Test-Path "proto\habits\v1\habits.proto") {
    Print-Info "Generating habits-service protos..."
    Push-Location "services\habits-service"
    try {
        & protoc `
            --go_out=. `
            --go_opt=paths=source_relative `
            --go-grpc_out=. `
            --go-grpc_opt=paths=source_relative `
            -I ..\..\proto `
            ..\..\proto\habits\v1\habits.proto

        if ($LASTEXITCODE -eq 0) {
            # Move generated files to proto directory
            New-Item -ItemType Directory -Force -Path "proto\habits\v1" | Out-Null
            if (Test-Path "habits\v1\habits.pb.go") {
                Move-Item -Force "habits\v1\*.pb.go" "proto\habits\v1\"
                Remove-Item -Recurse -Force "habits"
            }
            Pop-Location
            Print-Success "Generated habits-service protos"
        } else {
            Pop-Location
            Print-Error "Failed to generate habits-service protos"
        }
    } catch {
        Pop-Location
        Print-Error "Failed to generate habits-service protos: $_"
    }
}

# Generate Bad Habits Service protos (if exists)
if (Test-Path "proto\bad_habits\v1\bad_habits.proto") {
    Print-Info "Generating bad-habits-service protos..."
    try {
        & protoc `
            --go_out=services\bad-habits-service `
            --go_opt=paths=source_relative `
            --go-grpc_out=services\bad-habits-service `
            --go-grpc_opt=paths=source_relative `
            -I proto `
            proto\bad_habits\v1\bad_habits.proto

        if ($LASTEXITCODE -eq 0) {
            Print-Success "Generated bad-habits-service protos"
        } else {
            Print-Error "Failed to generate bad-habits-service protos"
        }
    } catch {
        Print-Error "Failed to generate bad-habits-service protos: $_"
    }
}

# Generate Events protos for notification-service
if (Test-Path "proto\events\v1\events.proto") {
    Print-Info "Generating events protos for notification-service..."
    Push-Location "services\notification-service"
    try {
        & protoc `
            --go_out=. `
            --go_opt=paths=source_relative `
            -I ..\..\proto `
            ..\..\proto\events\v1\events.proto

        if ($LASTEXITCODE -eq 0) {
            New-Item -ItemType Directory -Force -Path "proto\events\v1" | Out-Null
            if (Test-Path "events\v1\events.pb.go") {
                Move-Item -Force "events\v1\*.pb.go" "proto\events\v1\"
                Remove-Item -Recurse -Force "events"
            }
            Pop-Location
            Print-Success "Generated events protos for notification-service"
        } else {
            Pop-Location
            Print-Error "Failed to generate events protos"
        }
    } catch {
        Pop-Location
        Print-Error "Failed to generate events protos: $_"
    }
}

# Copy user service protos to api-gateway
Print-Info "Copying user service protos to api-gateway..."
try {
    New-Item -ItemType Directory -Force -Path "api-gateway\proto\user\v1" | Out-Null
    Copy-Item -Force "services\user-service\proto\user\v1\*.pb.go" "api-gateway\proto\user\v1\"
    Print-Success "Copied user service protos to api-gateway"
} catch {
    Print-Error "Failed to copy protos to api-gateway: $_"
}

# Copy habits service protos to api-gateway
Print-Info "Copying habits service protos to api-gateway..."
try {
    New-Item -ItemType Directory -Force -Path "api-gateway\proto\habits\v1" | Out-Null
    Copy-Item -Force "services\habits-service\proto\habits\v1\*.pb.go" "api-gateway\proto\habits\v1\"
    Print-Success "Copied habits service protos to api-gateway"
} catch {
    Print-Error "Failed to copy habits protos to api-gateway: $_"
}

Print-Success "All proto files generated successfully!"
