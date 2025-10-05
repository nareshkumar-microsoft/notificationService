# Notification Service Makefile
# This Makefile provides common tasks for the notification service

.PHONY: help build test clean run fmt vet lint deps demo check-all

# Default target
help: ## Show this help message
	@echo "Notification Service - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

# Build targets
build: ## Build the notification service binary
	@echo "🔨 Building notification service..."
	@go build -o bin/notification-service ./cmd/demo
	@echo "✅ Build complete: bin/notification-service"

demo: ## Run the foundation demo
	@echo "🚀 Running foundation demo..."
	@go run ./cmd/demo

# Development targets
deps: ## Download and verify dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@go mod verify
	@echo "✅ Dependencies updated"

fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet checks passed"

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test ./... -v
	@echo "✅ Tests completed"

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Quality checks
check-all: deps fmt vet test ## Run all quality checks (deps, fmt, vet, test)
	@echo "🎯 All checks completed successfully!"

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "✅ Linting completed"; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Utility targets
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✅ Clean completed"

install-tools: ## Install development tools
	@echo "🛠️  Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed"

# Project info
info: ## Show project information
	@echo "📋 Project Information:"
	@echo "  Name: Notification Service"
	@echo "  Language: Go"
	@echo "  Purpose: Hacktoberfest-friendly notification service"
	@echo "  Features: Email, SMS, Push notifications with mock providers"
	@echo ""
	@echo "📊 Project Stats:"
	@echo "  Go files: $$(find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "  Test files: $$(find . -name '*_test.go' -not -path './vendor/*' | wc -l)"
	@echo "  Lines of code: $$(find . -name '*.go' -not -path './vendor/*' -not -name '*_test.go' | xargs wc -l | tail -1 | awk '{print $$1}')"
	@echo "  Test lines: $$(find . -name '*_test.go' -not -path './vendor/*' | xargs wc -l | tail -1 | awk '{print $$1}')"

# Docker targets (for future use)
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@echo "⚠️  Docker support will be added in future PRs"

# Development workflow
dev: clean deps fmt vet test build demo ## Complete development workflow
	@echo "🎉 Development workflow completed successfully!"