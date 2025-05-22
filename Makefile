.PHONY: build clean test run lint db-migrate db-seed db-reset

# Build variables
BINARY_NAME=serviceability
BUILD_DIR=./build
CMD_DIR=./cmd/api
MIGRATION_DIR=./cmd/migrations

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Version variables
VERSION_PKG=prayog-serviceability-service/pkg/version
MAJOR=1
MINOR=0
PATCH=0
PRE_RELEASE=
BUILD_METADATA=
# Get git commit and build time
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY=$(shell git status --porcelain 2>/dev/null || echo "")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_FLAGS=-ldflags "\
	-X '$(VERSION_PKG).Major=$(MAJOR)' \
	-X '$(VERSION_PKG).Minor=$(MINOR)' \
	-X '$(VERSION_PKG).Patch=$(PATCH)' \
	-X '$(VERSION_PKG).PreRelease=$(PRE_RELEASE)' \
	-X '$(VERSION_PKG).BuildMetadata=$(BUILD_METADATA)' \
	-X '$(VERSION_PKG).Commit=$(GIT_COMMIT)' \
	-X '$(VERSION_PKG).BuildTime=$(BUILD_TIME)' \
	-X '$(VERSION_PKG).Dirty=$(GIT_DIRTY)'"

# Default target
all: test build

# Build the application
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/migrate $(MIGRATION_DIR)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Run the application
run:
	$(GOCMD) run $(BUILD_FLAGS) $(CMD_DIR)

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
	$(GOCMD) run $(BUILD_FLAGS) $(MIGRATION_DIR) -migrate

# Seed database with initial data
db-seed:
	$(GOCMD) run $(BUILD_FLAGS) $(MIGRATION_DIR) -seed

# Reset database: drop all tables and run migrations
db-reset:
	$(GOCMD) run $(BUILD_FLAGS) $(MIGRATION_DIR) -drop -migrate -seed

# Get version info
version:
	@echo "Version: $(MAJOR).$(MINOR).$(PATCH)$(if $(PRE_RELEASE),-$(PRE_RELEASE))$(if $(BUILD_METADATA),+$(BUILD_METADATA))"
	@echo "Commit: $(GIT_COMMIT)$(if $(GIT_DIRTY), (dirty))"
	@echo "Build Time: $(BUILD_TIME)"

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
	@echo "make version    - Show version information"
	@echo "make all        - Run tests and build"
	@echo "make help       - Show this help message" 