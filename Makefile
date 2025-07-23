.PHONY: build clean test run lint docker-build docker-run

# Build variables
BINARY_NAME=serviceability
BUILD_DIR=./build
CMD_DIR=./cmd/api

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

# Database migration functionality has been removed for safety
# If you need to modify the database schema, use database admin tools directly

# Get version info
version:
	@echo "Version: $(MAJOR).$(MINOR).$(PATCH)$(if $(PRE_RELEASE),-$(PRE_RELEASE))$(if $(BUILD_METADATA),+$(BUILD_METADATA))"
	@echo "Commit: $(GIT_COMMIT)$(if $(GIT_DIRTY), (dirty))"
	@echo "Build Time: $(BUILD_TIME)"

# Docker commands
docker-build:
	docker build -t serviceability-service .

docker-run:
	docker run -p 8080:8080 serviceability-service

# Help command
help:
	@echo "Available commands:"
	@echo "make build        - Build the application"
	@echo "make clean        - Remove build artifacts"
	@echo "make test         - Run tests"
	@echo "make run          - Run the application"
	@echo "make tidy         - Tidy go modules"
	@echo "make deps         - Install dependencies"
	@echo "make lint         - Run linter"
	@echo "make version      - Show version information"
	@echo "make docker-build - Build Docker image"
	@echo "make docker-run   - Run Docker container"
	@echo "make all          - Run tests and build"
	@echo "make help         - Show this help message" 