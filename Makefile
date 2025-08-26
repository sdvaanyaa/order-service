include .env.example
export

BINARY_NAME = order-service
BUILD_DIR = build
MIGRATIONS_DIR = migrations
DATABASE_DSN = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

.PHONY: all update linter build start run clean bin-deps up down restart goose-add goose-up goose-down goose-status test test-coverage

all: run

update:
	@echo "Updating dependencies"
	@go mod tidy

linter:
	@echo "Running linters"
	@golangci-lint run ./... --tests=false

build:
	@echo "Building application"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/order-service/main.go

start:
	@echo "Starting application"
	@$(BUILD_DIR)/$(BINARY_NAME)

run: bin-deps up goose-up update linter build start

clean:
	@echo "Cleaning up"
	@rm -rf $(BUILD_DIR)
	@go clean

bin-deps:
	@echo "Installing goose dependencies"
	@go install github.com/pressly/goose/v3/cmd/goose@latest

up:
	@echo "Starting Docker Compose"
	@docker-compose up -d

down:
	@echo "Stopping Docker Compose"
	@docker-compose down

restart: down up

goose-add:
	@echo "Creating new migration"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_DSN)" create $(NAME) sql

goose-up:
	@echo "Waiting for Postgres to be ready"
	@for i in {1..30}; do \
		pg_isready -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) -U $(POSTGRES_USER) && break; \
		echo "Postgres not ready, retry $$i/30"; \
		sleep 1; \
	done
	@echo "Applying migrations"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_DSN)" up

goose-down:
	@echo "Reverting migrations"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_DSN)" down

goose-status:
	@echo "Checking migration status"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_DSN)" status