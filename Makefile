# Makefile for Vinyl Catalog Application

# Variables
APP_NAME = vinyl-catalog
MAIN_PATH = ./cmd/main.go
MIGRATION_PATH = ./cmd/migrate/main.go
DEV_DB = golang_app
TEST_DB = golang_app_test

# Build the application
build:
	go build -o ${APP_NAME} ${MAIN_PATH}

# Run the application in development mode
run:
	PG_DATABASE=${DEV_DB} go run ${MAIN_PATH}

# Run the application with hot reload (requires air: https://github.com/cosmtrek/air)
dev:
	PG_DATABASE=${DEV_DB} air -c .air.toml

# Clean build files
clean:
	rm -f ${APP_NAME}

# Run all tests
test:
	PG_DATABASE=${TEST_DB} go test -v ./...

# Run model tests specifically
test-models:
	PG_DATABASE=${TEST_DB} go test -v ./db/models/...

# Run API handler tests specifically
test-handlers:
	PG_DATABASE=${TEST_DB} go test -v ./api/handlers/...

# Setup test database
setup-test-db:
	docker exec -it db psql -U postgres -c "DROP DATABASE IF EXISTS ${TEST_DB}"
	docker exec -it db psql -U postgres -c "CREATE DATABASE ${TEST_DB}"
	PG_DATABASE=${TEST_DB} go run ${MIGRATION_PATH} up

# Run migrations up on development database
migrate-up:
	PG_DATABASE=${DEV_DB} go run ${MIGRATION_PATH} up

# Run migrations down on development database
migrate-down:
	PG_DATABASE=${DEV_DB} go run ${MIGRATION_PATH} down

# Run migrations up on test database
migrate-test-up:
	PG_DATABASE=${TEST_DB} go run ${MIGRATION_PATH} up

# Run migrations down on test database
migrate-test-down:
	PG_DATABASE=${TEST_DB} go run ${MIGRATION_PATH} down

# Setup the entire project (create database, run migrations)
setup:
	docker exec -it db psql -U postgres -c "DROP DATABASE IF EXISTS ${DEV_DB}"
	docker exec -it db psql -U postgres -c "CREATE DATABASE ${DEV_DB}"
	$(MAKE) migrate-up
	$(MAKE) setup-test-db

# Install development tools (air for hot reload)
install-tools:
	go install github.com/cosmtrek/air@latest

# List all available make commands
help:
	@echo "Available commands:"
	@echo "  make build            - Build the application"
	@echo "  make run              - Run the application"
	@echo "  make dev              - Run with hot reload (requires air)"
	@echo "  make clean            - Remove build files"
	@echo "  make test             - Run all tests"
	@echo "  make test-models      - Run model tests"
	@echo "  make test-handlers    - Run API handler tests"
	@echo "  make setup-test-db    - Setup test database"
	@echo "  make migrate-up       - Run migrations up (dev)"
	@echo "  make migrate-down     - Run migrations down (dev)"
	@echo "  make migrate-test-up  - Run migrations up (test)"
	@echo "  make migrate-test-down- Run migrations down (test)"
	@echo "  make setup            - Setup the entire project"
	@echo "  make install-tools    - Install development tools"

.PHONY: build run dev clean test test-models test-handlers setup-test-db migrate-up migrate-down migrate-test-up migrate-test-down setup install-tools help