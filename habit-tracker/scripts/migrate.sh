#!/bin/bash

set -e

# Determine script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Configuration
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_SSL_MODE="${DB_SSL_MODE:-disable}"
NETWORK="${NETWORK:-habit-tracker-network}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Function to run migrations for a service using Docker
migrate_service() {
    local service_name=$1
    local db_name=$2
    local migrations_path=$3
    local action=${4:-up}
    local version=$5

    print_info "Running migrations for ${service_name}..."

    # Full path to migrations
    local full_migrations_path="${PROJECT_ROOT}/${migrations_path}"

    if [ ! -d "${full_migrations_path}" ]; then
        print_error "Migrations directory not found: ${full_migrations_path}"
        return 1
    fi

    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${db_name}?sslmode=${DB_SSL_MODE}"

    # Run migrations using Docker
    if [ "$action" == "down" ]; then
        docker run --rm \
            -v "${full_migrations_path}:/migrations" \
            --network "${NETWORK}" \
            migrate/migrate:v4.17.0 \
            -path=/migrations \
            -database "${DATABASE_URL}" \
            down
        print_success "Rolled back migrations for ${service_name}"
    elif [ "$action" == "force" ]; then
        docker run --rm \
            -v "${full_migrations_path}:/migrations" \
            --network "${NETWORK}" \
            migrate/migrate:v4.17.0 \
            -path=/migrations \
            -database "${DATABASE_URL}" \
            force "$version"
        print_success "Forced version $version for ${service_name}"
    else
        docker run --rm \
            -v "${full_migrations_path}:/migrations" \
            --network "${NETWORK}" \
            migrate/migrate:v4.17.0 \
            -path=/migrations \
            -database "${DATABASE_URL}" \
            up
        print_success "Applied migrations for ${service_name}"
    fi
}

# Main script
case "$1" in
    up)
        print_info "Applying all migrations..."
        migrate_service "user-service" "user_service" "services/user-service/migrations"
        migrate_service "habits-service" "habits_service" "services/habits-service/migrations"
        migrate_service "bad-habits-service" "bad_habits_service" "services/bad-habits-service/migrations"
        migrate_service "notification-service" "notification_service" "services/notification-service/migrations"
        print_success "All migrations applied successfully!"
        ;;
    down)
        print_info "Rolling back all migrations..."
        migrate_service "user-service" "user_service" "services/user-service/migrations" "down"
        migrate_service "habits-service" "habits_service" "services/habits-service/migrations" "down"
        migrate_service "bad-habits-service" "bad_habits_service" "services/bad-habits-service/migrations" "down"
        migrate_service "notification-service" "notification_service" "services/notification-service/migrations" "down"
        print_success "All migrations rolled back successfully!"
        ;;
    user)
        migrate_service "user-service" "user_service" "services/user-service/migrations" "$2" "$3"
        ;;
    habits)
        migrate_service "habits-service" "habits_service" "services/habits-service/migrations" "$2" "$3"
        ;;
    bad-habits)
        migrate_service "bad-habits-service" "bad_habits_service" "services/bad-habits-service/migrations" "$2" "$3"
        ;;
    notification)
        migrate_service "notification-service" "notification_service" "services/notification-service/migrations" "$2" "$3"
        ;;
    *)
        echo "Usage: $0 {up|down|user|habits|bad-habits|notification} [up|down|force] [version]"
        echo ""
        echo "Examples:"
        echo "  $0 up                      - Apply all migrations"
        echo "  $0 down                    - Rollback all migrations"
        echo "  $0 user up                 - Apply user-service migrations"
        echo "  $0 user down               - Rollback user-service migrations"
        echo "  $0 user force 1            - Force user-service to version 1"
        echo "  $0 notification up         - Apply notification-service migrations"
        echo ""
        echo "Note: Make sure Docker containers are running (docker compose up -d)"
        exit 1
        ;;
esac
