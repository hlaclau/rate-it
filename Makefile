.PHONY: help install dev dev-api dev-ui stop db-up db-down db-reset lint format

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ── Setup ────────────────────────────────────────────────────────────────────

install: ## Install all dependencies (requires mise)
	mise install
	cd frontend && bun install
	cd backend && go mod download

# ── Dev servers ──────────────────────────────────────────────────────────────

dev: db-up ## Start everything (api + frontend + databases)
	@$(MAKE) -j2 dev-api dev-ui

dev-api: ## Start the Go API (port 8080)
	cd backend && go run ./cmd/...

dev-ui: ## Start the Nuxt frontend (port 3000)
	cd frontend && bun run dev

stop: ## Stop api, frontend and databases
	-lsof -ti:8080 | xargs kill -9 2>/dev/null
	-lsof -ti:3000 | xargs kill -9 2>/dev/null
	docker compose -f docker-compose.dev.yml down

# ── Databases ────────────────────────────────────────────────────────────────

db-up: ## Start PostgreSQL and Redis
	docker compose -f docker-compose.dev.yml up -d

db-down: ## Stop PostgreSQL and Redis
	docker compose -f docker-compose.dev.yml down

db-reset: ## Wipe and restart databases (destroys data)
	docker compose -f docker-compose.dev.yml down -v
	docker compose -f docker-compose.dev.yml up -d

# ── Quality ──────────────────────────────────────────────────────────────────

lint: ## Lint frontend and backend
	cd frontend && bun run lint
	cd backend && go vet ./...

format: ## Format frontend and backend
	cd frontend && bun run format
	cd backend && go fmt ./...
