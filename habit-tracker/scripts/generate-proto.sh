#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    print_error "protoc is not installed. Please install Protocol Buffers compiler."
    echo "Install via:"
    echo "  macOS: brew install protobuf"
    echo "  Linux: apt install -y protobuf-compiler"
    echo "  Or download from: https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    print_error "protoc-gen-go is not installed."
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    print_success "Installed protoc-gen-go"
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    print_error "protoc-gen-go-grpc is not installed."
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    print_success "Installed protoc-gen-go-grpc"
fi

print_info "Generating proto files..."

# Create output directories
mkdir -p services/user-service/proto/user/v1
mkdir -p services/habits-service/proto/habits/v1
mkdir -p services/bad-habits-service/proto/bad_habits/v1
mkdir -p services/notification-service/proto/events/v1

# Generate User Service protos
print_info "Generating user-service protos..."
cd services/user-service
protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    -I ../../proto \
    ../../proto/user/v1/user.proto

if [ $? -eq 0 ]; then
    # Move generated files to proto directory
    mkdir -p proto/user/v1
    if [ -f "user/v1/user.pb.go" ]; then
        mv user/v1/*.pb.go proto/user/v1/
        rm -rf user
    fi
    cd ../..
    print_success "Generated user-service protos"
else
    cd ../..
    print_error "Failed to generate user-service protos"
    exit 1
fi

# Generate Habits Service protos (if exists)
if [ -f "proto/habits/v1/habits.proto" ]; then
    print_info "Generating habits-service protos..."
    protoc \
        --go_out=services/habits-service \
        --go_opt=paths=source_relative \
        --go-grpc_out=services/habits-service \
        --go-grpc_opt=paths=source_relative \
        -I proto \
        proto/habits/v1/habits.proto

    if [ $? -eq 0 ]; then
        print_success "Generated habits-service protos"
    else
        print_error "Failed to generate habits-service protos"
    fi
fi

# Generate Bad Habits Service protos (if exists)
if [ -f "proto/bad_habits/v1/bad_habits.proto" ]; then
    print_info "Generating bad-habits-service protos..."
    protoc \
        --go_out=services/bad-habits-service \
        --go_opt=paths=source_relative \
        --go-grpc_out=services/bad-habits-service \
        --go-grpc_opt=paths=source_relative \
        -I proto \
        proto/bad_habits/v1/bad_habits.proto

    if [ $? -eq 0 ]; then
        print_success "Generated bad-habits-service protos"
    else
        print_error "Failed to generate bad-habits-service protos"
    fi
fi

# Generate Events protos for notification-service
if [ -f "proto/events/v1/events.proto" ]; then
    print_info "Generating events protos for notification-service..."
    cd services/notification-service
    protoc \
        --go_out=. \
        --go_opt=paths=source_relative \
        -I ../../proto \
        ../../proto/events/v1/events.proto

    if [ $? -eq 0 ]; then
        # Move generated files to proto directory
        mkdir -p proto/events/v1
        if [ -f "events/v1/events.pb.go" ]; then
            mv events/v1/*.pb.go proto/events/v1/
            rm -rf events
        fi
        cd ../..
        print_success "Generated events protos for notification-service"
    else
        cd ../..
        print_error "Failed to generate events protos"
    fi
fi

print_success "All proto files generated successfully!"
