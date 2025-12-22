.PHONY: help build up down logs clean migrate-up migrate-down proto

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build all services
	cd deployments && docker-compose -f docker-compose.dev.yml build

up: ## Start all services
	cd deployments && docker-compose -f docker-compose.dev.yml up -d

down: ## Stop all services
	cd deployments && docker-compose -f docker-compose.dev.yml down

logs: ## Show logs from all services
	cd deployments && docker-compose -f docker-compose.dev.yml logs -f

logs-user: ## Show logs from user-service
	cd deployments && docker-compose -f docker-compose.dev.yml logs -f user-service

logs-gateway: ## Show logs from api-gateway
	cd deployments && docker-compose -f docker-compose.dev.yml logs -f api-gateway

clean: ## Stop and remove all containers, volumes, and images
	cd deployments && docker-compose -f docker-compose.dev.yml down -v --rmi all

migrate-up: ## Run migrations for user-service
	./scripts/migrate.sh up user-service

migrate-down: ## Rollback migrations for user-service
	./scripts/migrate.sh down user-service

proto: ## Generate protobuf files
	./scripts/generate-proto.sh

restart: down up ## Restart all services

ps: ## Show running containers
	cd deployments && docker-compose -f docker-compose.dev.yml ps
