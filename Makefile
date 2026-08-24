# ============================================
# Comandos para CloudPulse AI
# ============================================

# Colores para los mensajes
GREEN := \033[0;32m
NC := \033[0m

# Muestra todos los comandos disponibles
help: ## Mostrar esta ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# Instala todas las dependencias
install: ## Instalar dependencias
	@echo "📦 Instalando dependencias de Go..."
	cd services/api-go && go mod download
	@echo "✅ Listo"

# Inicia la API en Go
dev-api: ## Iniciar API en modo desarrollo
	@echo "🚀 Iniciando API..."
	cd services/api-go && go run cmd/main.go

# Inicia el frontend en React
dev-frontend: ## Iniciar frontend
	@echo "🚀 Iniciando frontend..."
	cd frontend && npm run dev

# Inicia el collector en Rust
dev-rust: ## Iniciar collector Rust
	@echo "🚀 Iniciando collector..."
	cd services/collector-rust && cargo run

# Inicia todo con Docker
docker-up: ## Iniciar todos los servicios
	@echo "🐳 Levantando servicios..."
	docker-compose up -d

# Detiene los servicios
docker-down: ## Detener servicios
	@echo "🛑 Deteniendo servicios..."
	docker-compose down

# Ejecuta los tests
test: ## Ejecutar tests
	@echo "🧪 Ejecutando tests..."
	cd services/api-go && go test ./...

# Limpia archivos generados
clean: ## Limpiar archivos
	@echo "🧹 Limpiando..."
	rm -rf services/api-go/bin
	rm -rf services/collector-rust/target
	rm -rf frontend/node_modules