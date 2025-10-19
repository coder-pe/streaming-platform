# Makefile para Streaming Platform

.PHONY: help build run dev test clean update-version deploy

# Variables
APP_NAME=streaming-platform
VERSION=$(shell git rev-parse --short HEAD 2>/dev/null || date +"%Y.%m.%d.%H%M")

help: ## Mostrar esta ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: update-version ## Compilar la aplicación
	@echo "🔨 Compilando aplicación..."
	go build -o bin/$(APP_NAME) cmd/server/main.go
	@echo "✅ Compilación completada: bin/$(APP_NAME)"

run: ## Ejecutar la aplicación
	@echo "🚀 Ejecutando aplicación..."
	go run cmd/server/main.go

dev: update-version run ## Actualizar versión y ejecutar en modo desarrollo

test: ## Ejecutar tests
	@echo "🧪 Ejecutando tests..."
	go test -v ./...

clean: ## Limpiar archivos compilados
	@echo "🧹 Limpiando..."
	rm -rf bin/
	rm -f web/static/js/config/version.js.bak
	@echo "✅ Limpieza completada"

update-version: ## Actualizar versión de cache busting
	@echo "📦 Actualizando versión a: $(VERSION)"
	@./scripts/update-version.sh

deploy: test build ## Ejecutar tests, compilar y preparar para deploy
	@echo "🚀 Aplicación lista para deploy"
	@echo "   Versión: $(VERSION)"
	@echo "   Binary: bin/$(APP_NAME)"

watch: ## Ejecutar con hot reload (requiere air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "❌ 'air' no está instalado. Instalar con: go install github.com/cosmtrek/air@latest"; \
	fi
