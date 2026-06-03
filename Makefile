# =========================
# Config
# =========================

APP_NAME := media-jukebox
BACKEND_BIN := ../dist/server
BACKEND_DIR := backend
UI_DIR := frontend
UI_DIST := frontend/dist
GO_MAIN := ./cmd/server/

DOCKER_COMPOSE := docker compose

# =========================
# Default
# =========================

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  dev            Run backend + frontend dev servers"
	@echo "  dev-backend    Run Go backend only"
	@echo "  dev-ui         Run Vite dev server"
	@echo "  build          Build frontend + backend"
	@echo "  build-ui       Build frontend (vite)"
	@echo "  build-backend  Build Go binary"
	@echo "  run            Run backend binary locally"
	@echo "  docker-build   Build docker images"
	@echo "  docker-up      Start full stack (prod)"
	@echo "  docker-down    Stop docker stack"
	@echo "  clean          Remove build artifacts"

# =========================
# Development
# =========================

.PHONY: dev
dev:
	@echo "Starting backend + UI dev mode..."
	DEV_MODE=1 go run $(GO_MAIN)

.PHONY: dev-backend
dev-backend:
	DEV_MODE=1 go run $(GO_MAIN)

.PHONY: dev-ui
dev-ui:
	cd $(UI_DIR) && npm run dev

# =========================
# Build (production)
# =========================

.PHONY: build
build: build-ui build-backend

.PHONY: build-ui
build-ui:
	@echo "Building frontend..."
	cd $(UI_DIR) && npm install && npm run build

.PHONY: build-backend
build-backend:
	@echo "Building backend..."
	mkdir -p dist
	cd $(BACKEND_DIR) && go build -o $(BACKEND_BIN) $(GO_MAIN)

# =========================
# Run locally
# =========================

.PHONY: run
run: build
	@echo "Running backend..."
	./$(BACKEND_BIN)

# =========================
# Docker
# =========================

.PHONY: docker-build
docker-build:
	$(DOCKER_COMPOSE) build

.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) up --build

.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

# =========================
# Clean
# =========================

.PHONY: clean
clean:
	rm -rf dist
	rm -rf $(UI_DIST)
