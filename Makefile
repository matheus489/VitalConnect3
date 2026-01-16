# VitalConnect Makefile
# Commands for development and deployment

.PHONY: help dev dev-db dev-backend dev-frontend build test clean

# Default target
help:
	@echo "VitalConnect - Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev           - Start all services (requires Docker)"
	@echo "  make dev-db        - Start only database services"
	@echo "  make dev-backend   - Start backend server"
	@echo "  make dev-frontend  - Start frontend dev server"
	@echo ""
	@echo "Build:"
	@echo "  make build         - Build all services"
	@echo "  make build-backend - Build backend Go binary"
	@echo "  make build-frontend - Build frontend Next.js"
	@echo ""
	@echo "Database:"
	@echo "  make migrate       - Run database migrations"
	@echo "  make seed          - Run database seeder"
	@echo ""
	@echo "Testing:"
	@echo "  make test          - Run all tests"
	@echo "  make test-backend  - Run backend tests"
	@echo "  make test-frontend - Run frontend tests"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make clean-db      - Remove database volumes"

# =============================================================================
# Development
# =============================================================================

# Start all services with Docker
dev:
	docker compose --profile dev up -d
	@echo "Services started. Waiting for health checks..."
	@sleep 5
	docker compose ps

# Start only database services
dev-db:
	docker compose up -d postgres redis
	@echo "Database services started."
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

# Start backend server (requires Go installed)
dev-backend:
	cd backend && go run cmd/api/main.go

# Start frontend dev server
dev-frontend:
	cd frontend && npm run dev

# =============================================================================
# Build
# =============================================================================

# Build all
build: build-backend build-frontend

# Build backend
build-backend:
	cd backend && go build -o bin/api ./cmd/api

# Build frontend
build-frontend:
	cd frontend && npm run build

# =============================================================================
# Database
# =============================================================================

# Run migrations
migrate:
	cd backend && go run cmd/migrate/main.go up

# Run seeder
seed:
	cd backend && go run cmd/seed/main.go

# =============================================================================
# Testing
# =============================================================================

# Run all tests
test: test-backend test-frontend

# Run backend tests
test-backend:
	cd backend && go test -v ./...

# Run frontend tests
test-frontend:
	cd frontend && npm test

# =============================================================================
# Cleanup
# =============================================================================

# Clean build artifacts
clean:
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/out

# Remove database volumes (WARNING: destroys data)
clean-db:
	docker compose down -v
	@echo "Database volumes removed."

# Stop all containers
stop:
	docker compose down
