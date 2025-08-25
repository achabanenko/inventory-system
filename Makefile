.PHONY: help dev test build lint clean docker-build docker-up docker-down migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development servers
	@echo "Starting development environment..."
	# @docker compose up postgres -d
	# @sleep 3
	@make -j2 dev-backend dev-frontend

dev-backend: ## Start backend in development mode
	@echo "Starting backend..."
	@cd backend && go run cmd/api/main.go

dev-frontend: ## Start frontend in development mode
	@echo "Starting frontend..."
	@cd frontend && bun run dev

test: ## Run tests
	@echo "Running backend tests..."
	@cd backend && go test ./...
	@echo "Running frontend tests..."
	@cd frontend && bun test

build: ## Build both backend and frontend
	@echo "Building backend..."
	@cd backend && go build -o bin/api cmd/api/main.go
	@echo "Building frontend..."
	@cd frontend && bun run build

lint: ## Run linters
	@echo "Running Go linter..."
	@cd backend && golangci-lint run
	@echo "Running frontend linter..."
	@cd frontend && bun run lint

clean: ## Clean build artifacts
	@echo "Cleaning up..."
	@cd backend && rm -rf bin/
	@cd frontend && rm -rf dist/ node_modules/.vite .bun/

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@docker compose build

docker-up: ## Start all services with Docker
	@echo "Starting all services..."
	@docker compose up -d

docker-down: ## Stop all services
	@echo "Stopping all services..."
	@docker-compose down

docker-logs: ## Show logs from all services
	@docker-compose logs -f

migrate: ## Run database migrations (requires running postgres)
	@echo "Running migrations..."
	@cd backend && go run cmd/migrate/main.go

seed: ## Seed database with sample data
	@echo "Seeding database..."
	@cd backend && go run cmd/seed/main.go

# Development helpers
install-backend: ## Install backend dependencies
	@cd backend && go mod tidy

install-frontend: ## Install frontend dependencies
	@cd frontend && bun install

install: install-backend install-frontend ## Install all dependencies

format: ## Format code
	@cd backend && go fmt ./...
	@cd frontend && bun run format