.PHONY: build clean test run lint db-migrate db-seed db-reset

# Build variables
BINARY_NAME=serviceability
BUILD_DIR=./build
CMD_DIR=./cmd/serviceability
MIGRATION_DIR=./cmd/migrations

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Default target
all: test build

# Build the application
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/migrate $(MIGRATION_DIR)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Run the application
run:
	$(GOCMD) run $(CMD_DIR)

# Tidy go modules
tidy:
	$(GOMOD) tidy

# Install dependencies
deps:
	$(GOMOD) download

# Run linter
lint:
	$(GOLINT) run

# Database migrations
db-migrate:
	$(GOCMD) run $(MIGRATION_DIR) -migrate

# Seed database with initial data
db-seed:
	$(GOCMD) run $(MIGRATION_DIR) -seed

# Reset database: drop all tables and run migrations
db-reset:
	$(GOCMD) run $(MIGRATION_DIR) -drop -migrate -seed

# Help command
help:
	@echo "Available commands:"
	@echo "make build      - Build the application"
	@echo "make clean      - Remove build artifacts"
	@echo "make test       - Run tests"
	@echo "make run        - Run the application"
	@echo "make tidy       - Tidy go modules"
	@echo "make deps       - Install dependencies"
	@echo "make lint       - Run linter"
	@echo "make db-migrate - Run database migrations"
	@echo "make db-seed    - Seed database with initial data"
	@echo "make db-reset   - Reset database (drop, migrate, seed)"
	@echo "make all        - Run tests and build"
	@echo "make help       - Show this help message" 