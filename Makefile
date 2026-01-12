.PHONY: help up down logs restart build clean db-shell swagger

help: ## Show this help message
    @echo 'Usage: make [target]'
    @echo ''
    @echo 'Available targets:'
    @awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Start all services
    docker-compose up -d

down: ## Stop all services
    docker-compose down

logs: ## Show logs
    docker-compose logs -f

logs-app: ## Show app logs only
    docker-compose logs -f app

logs-db: ## Show database logs only
    docker-compose logs -f postgres

restart: ## Restart all services
    docker-compose restart

restart-app: ## Restart app only
    docker-compose restart app

build: ## Build the application
    docker-compose build --no-cache

rebuild: ## Rebuild and restart
    docker-compose down && docker-compose build --no-cache && docker-compose up -d

clean: ## Stop and remove all containers, volumes
    docker-compose down -v

db-shell: ## Access PostgreSQL shell
    docker-compose exec postgres psql -U postgres -d yota_db

app-shell: ## Access app container shell
    docker-compose exec app sh

swagger: ## Generate swagger documentation
    docker-compose exec app swag init -g main.go

test: ## Run tests (when available)
    docker-compose exec app go test ./...