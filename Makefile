# ============================================
# CloudPulse AI commands
# ============================================

# Message colors
GREEN := \033[0;32m
NC := \033[0m

# Show all available commands
help: ## Show this help
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# Install all dependencies
install: ## Install dependencies
	@echo "📦 Installing Go dependencies..."
	cd services/api-go && go mod download
	@echo "✅ Done"

# Start the Go API
dev-api: ## Start API in development mode
	@echo "🚀 Starting API..."
	cd services/api-go && go run cmd/main.go

# Start the React frontend
dev-frontend: ## Start frontend
	@echo "🚀 Starting frontend..."
	cd frontend && npm run dev

# Start the Rust collector
dev-rust: ## Start Rust collector
	@echo "🚀 Starting collector..."
	cd services/collector-rust && cargo run

# Start everything with Docker
docker-up: ## Start all services
	@echo "🐳 Starting services..."
	docker-compose up -d

# Stop the services
docker-down: ## Stop services
	@echo "🛑 Stopping services..."
	docker-compose down

# Run tests
test: ## Run tests
	@echo "🧪 Running tests..."
	cd services/api-go && go test ./...

# Clean generated files
clean: ## Clean files
	@echo "🧹 Cleaning..."
	rm -rf services/api-go/bin
	rm -rf services/collector-rust/target
	rm -rf frontend/node_modules