# Load variables from .env (if present) and export them to recipes.
-include .env
export

# Defaults (overridden by .env when present).
POSTGRES_USER     ?= todoapp
POSTGRES_PASSWORD ?= todoapp
POSTGRES_HOST     ?= localhost
POSTGRES_PORT     ?= 5432
POSTGRES_DB       ?= messenger
POSTGRES_SSLMODE  ?= disable

DATABASE_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

BIN_DIR := bin
MAIN    := ./cmd/app

# Append .exe on Windows so the built binary is runnable.
ifeq ($(OS),Windows_NT)
APP := $(BIN_DIR)/app.exe
else
APP := $(BIN_DIR)/app
endif

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

## ---------- Go ----------

.PHONY: tidy
tidy: ## Download and tidy Go module dependencies
	go mod tidy

.PHONY: run
run: ## Run the API server locally (needs Postgres up)
	go run $(MAIN)

.PHONY: build
build: ## Build the server binary into ./bin
	go build -o $(APP) $(MAIN)

.PHONY: test
test: ## Run tests
	go test ./...

.PHONY: fmt
fmt: ## Format the code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

## ---------- Swagger ----------

.PHONY: swagger
swagger: ## Regenerate Swagger docs locally (requires: go install github.com/swaggo/swag/cmd/swag@v1.16.6)
	swag init -g cmd/app/main.go -o docs

.PHONY: swagger-docker
swagger-docker: ## Regenerate Swagger docs using the swagger docker service
	docker compose run --rm swagger

## ---------- Database / migrations ----------

.PHONY: db-up
db-up: ## Start Postgres + port-forwarder in the background
	docker compose up -d todoapp-postgres port-forwarder

.PHONY: db-down
db-down: ## Stop all compose services
	docker compose down

.PHONY: db-reset
db-reset: ## Stop services and delete Postgres data (DESTRUCTIVE)
	docker compose down
	rm -rf out/pgdata/*

.PHONY: migrate-up
migrate-up: ## Apply all migrations via the migrate docker image
	docker compose run --rm todoapp-postgres-migrate

.PHONY: migrate-down
migrate-down: ## Roll back the last migration (local migrate CLI required)
	migrate -path migrations -database "$(DATABASE_URL)" down 1

.PHONY: migrate-create
migrate-create: ## Create a new migration: make migrate-create name=add_something
	migrate create -ext sql -dir migrations -seq $(name)

## ---------- Docker (full stack) ----------

.PHONY: compose-up
compose-up: ## Start Postgres, run migrations and the port-forwarder
	docker compose up -d todoapp-postgres port-forwarder
	docker compose up todoapp-postgres-migrate

.PHONY: compose-down
compose-down: ## Stop and remove all compose services
	docker compose down

.PHONY: up
up: compose-up run ## Bring up the DB stack, then run the server
