# WazMeow - WhatsApp Session Manager
# Clean Architecture Implementation

.PHONY: help build run test clean deps lint format docker swagger

# Default target
help: ## Show this help message
	@echo "WazMeow - WhatsApp Session Manager"
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build commands
build: ## Build the application
	@echo "Building WazMeow..."
	@mkdir -p bin
	@go build -o bin/wazmeow cmd/server/main.go
	@echo "Build complete: bin/wazmeow"

build-all: ## Build all binaries
	@echo "Building all binaries..."
	@mkdir -p bin
	@go build -o bin/wazmeow cmd/server/main.go
	@echo "Build complete"

# Run commands
run: ## Run the application
	@echo "Starting WazMeow..."
	@go run cmd/server/main.go

dev: ## Run in development mode with auto-reload
	@echo "Starting WazMeow in development mode..."
	@go run cmd/server/main.go

# Test commands
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v ./tests/unit/...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-api: ## Test API endpoints (requires server running)
	@echo "Testing API endpoints..."
	@echo "Make sure the server is running (make run) before running this command"
	@curl -s http://localhost:8080/health | jq . || echo "Health check failed or jq not installed"
	@curl -s http://localhost:8080/sessions | jq . || echo "Sessions endpoint failed or jq not installed"
	@echo "API test complete. Use api-tests.http for detailed testing."

# Development commands
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# Database commands
migrate: ## Run database migrations
	@echo "Running database migrations..."
	@go run cmd/server/main.go --migrate-only

# Clean commands
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

clean-data: ## Clean database and data files
	@echo "Cleaning data files..."
	@rm -rf data/
	@echo "Data files cleaned"

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t wazmeow:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env wazmeow:latest

# Setup commands
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@mkdir -p data
	@go mod download
	@echo "Setup complete. Edit .env file as needed."

api-docs: ## Open API test file for interactive testing
	@echo "Opening API test file..."
	@echo "Use VS Code with REST Client extension to run the tests in api-tests.http"
	@code api-tests.http 2>/dev/null || echo "VS Code not found. Open api-tests.http manually."

# Documentation commands
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@$(HOME)/go/bin/swag init -g cmd/server/main.go -o docs
	@echo "Fixing compatibility issues..."
	@sed -i 's/LeftDelim:.*"{{",//g' docs/docs.go
	@sed -i 's/RightDelim:.*"}}",//g' docs/docs.go
	@echo "Swagger documentation generated in docs/ directory"
	@echo "Access documentation at: http://localhost:8080/swagger/"

swagger-install: ## Install Swagger CLI tool
	@echo "Installing Swagger CLI tool..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Swagger CLI installed. Run 'make swagger' to generate docs."

# Install tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Tools installed"
