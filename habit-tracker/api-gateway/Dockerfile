# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Install swag for Swagger generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN  go mod download

# Copy source code
COPY . .

# Generate Swagger documentation
RUN swag init -g cmd/main.go -o ./docs --parseDependency --parseInternal

# Build binary
RUN CGO_ENABLED=0 GOOS=linux  go build -a -installsuffix cgo -o api-gateway ./cmd

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/api-gateway .

# Copy config files
COPY --from=builder /app/config ./config

# Copy Swagger documentation
COPY --from=builder /app/docs ./docs

# Expose HTTP port
EXPOSE 8080

# Run service
CMD ["./api-gateway"]
