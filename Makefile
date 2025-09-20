.PHONY: help run build test clean docker-build docker-run dev deps fmt lint

# Default target
help:
	@echo "Available commands:"
	@echo "  run         - Run the application"
	@echo "  dev         - Run in development mode with hot reload"
	@echo "  build       - Build the application"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  lint        - Run linter"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run  - Run with Docker Compose"

# Run the application
run:
	go run main.go

# Development mode with hot reload (requires air)
dev:
	@if ! command -v air > /dev/null; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air

# Build the application
build:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/auth-server main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f auth.db

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t oauth2-auth-server .

# Run with Docker Compose
docker-run:
	docker-compose up --build

# Run Docker Compose in background
docker-up:
	docker-compose up -d --build

# Stop Docker Compose
docker-down:
	docker-compose down

# View Docker logs
docker-logs:
	docker-compose logs -f

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		echo "Please edit .env file with your configuration"; \
	fi
	$(MAKE) deps
	@echo "Setup complete!"

# Database migration (if you add migration files later)
migrate-up:
	@echo "Running database migrations..."
	# Add migration command here

# Database rollback
migrate-down:
	@echo "Rolling back database migrations..."
	# Add rollback command here

# Generate API documentation (if you add swagger later)
docs:
	@echo "Generating API documentation..."
	# Add swagger generate command here

# Security audit
security:
	@if ! command -v gosec > /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...