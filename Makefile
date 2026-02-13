.PHONY: help docker-build docker-up docker-down docker-logs docker-clean docker-ps docker-restart docker-dev docker-prod build run test

help:
	@echo "Fittingly Go Backend - Docker & Build Commands"
	@echo "=============================================="
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-up          Start all services (foreground)"
	@echo "  make docker-up-d        Start all services (background)"
	@echo "  make docker-down        Stop all services"
	@echo "  make docker-clean       Remove all containers and volumes"
	@echo "  make docker-ps          Show running containers"
	@echo "  make docker-logs        View logs from all services"
	@echo "  make docker-logs-api    View API logs only"
	@echo "  make docker-logs-db     View MongoDB logs only"
	@echo "  make docker-build       Build Docker images"
	@echo "  make docker-rebuild     Rebuild without cache"
	@echo "  make docker-restart     Restart services"
	@echo "  make docker-dev         Start with development profile"
	@echo "  make docker-shell-api   Open shell in API container"
	@echo "  make docker-shell-db    Open shell in MongoDB container"
	@echo ""
	@echo "Local Development:"
	@echo "  make build              Build local binary"
	@echo "  make run                Run local binary"
	@echo "  make dev                Run with hot-reload (requires air)"
	@echo "  make test               Run tests"
	@echo "  make clean              Clean build artifacts"
	@echo ""

# Docker Commands
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-rebuild:
	@echo "Rebuilding Docker images (no cache)..."
	docker-compose build --no-cache

docker-up:
	@echo "Starting services in foreground..."
	docker-compose up

docker-up-d:
	@echo "Starting services in background..."
	docker-compose up -d
	@echo "✓ Services started"
	@echo "API: http://localhost:8080"
	@echo "MongoDB Express: http://localhost:8081"

docker-down:
	@echo "Stopping services..."
	docker-compose down
	@echo "✓ Services stopped"

docker-clean:
	@echo "Cleaning up Docker resources..."
	docker-compose down -v --remove-orphans
	docker system prune -f
	@echo "✓ Cleanup complete"

docker-ps:
	@docker-compose ps

docker-logs:
	docker-compose logs -f

docker-logs-api:
	docker-compose logs -f api

docker-logs-db:
	docker-compose logs -f mongodb

docker-restart:
	@echo "Restarting services..."
	docker-compose restart
	@echo "✓ Services restarted"

docker-dev:
	@echo "Starting with development profile..."
	docker-compose -f docker-compose.yml -f docker-compose.override.yml up

docker-shell-api:
	@echo "Opening shell in API container..."
	docker-compose exec api sh

docker-shell-db:
	@echo "Opening MongoDB shell..."
	docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev

# Local Development Commands
build:
	@echo "Building Go binary..."
	go build -o bin/server ./cmd/server
	@echo "✓ Binary built at bin/server"

run: build
	@echo "Running server..."
	./bin/server

dev:
	@echo "Running with hot-reload (requires: go install github.com/cosmtrek/air@latest)..."
	air

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	@echo "✓ Cleanup complete"

# Utility Commands
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go mod download
	@echo "✓ Tools installed"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Code formatted"

vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ Vet check complete"

# Database Commands
db-init:
	@echo "Re-initializing database..."
	docker-compose down -v
	docker-compose up -d
	@sleep 5
	@echo "✓ Database initialized with sample data"

db-reset:
	@echo "Resetting database..."
	docker-compose exec mongodb mongosh -u root -p rootpassword kloset_dev --eval "db.dropDatabase()"
	@echo "✓ Database reset"

# Status Commands
status:
	@echo "=== Service Status ==="
	@docker-compose ps
	@echo ""
	@echo "=== Docker Stats ==="
	@docker stats --no-stream --format "table {{.Container}}\t{{.MemUsage}}\t{{.CPUPerc}}"

version:
	@echo "=== Version Information ==="
	@go version
	@docker --version
	@docker-compose --version
