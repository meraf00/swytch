# Binary base name
BINARY_NAME=build

# Output directory
BUILD_DIR=build

DB_USER ?= postgres
DB_PASS ?= root
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= swytch

# Platforms to build for
PLATFORMS=\
	linux/amd64 \
	windows/amd64 \
	darwin/amd64

# Default target
.PHONY: all
all: build

# Build for current OS
.PHONY: build
build:
	go build -o $(BUILD_DIR) ./...

# Cross-compile for multiple platforms
.PHONY: build-all
build-all:
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform##*/}; \
		OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)-$${OS}-$${ARCH}; \
		if [ "$${OS}" = "windows" ]; then OUTPUT=$${OUTPUT}.exe; fi; \
		echo "Building $${OUTPUT}..."; \
		GOOS=$${OS} GOARCH=$${ARCH} go build -o $${OUTPUT} . || exit 1; \
	done

# Database Migrations
POSTGRES_DSN := "postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"

migrate-create: $(MIGRATE)
	@ read -p "Please provide name for the migration: " Name; \
	atlas migrate diff $${Name} \
		--dir "file://core/db/migrations" \
		--to "file://core/db/schema/schema.sql" \
		--dev-url "docker://postgres?search_path=public"

migrate-up: $(MIGRATE)
	atlas migrate apply \
		--dir "file://core/db/migrations" \
		--url "$(POSTGRES_DSN)"

# Run the application
.PHONY: run
run:
	go run .

# Test the application
.PHONY: test
test:
	go test ./...

# Format the code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint the code (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run

# Clean build files
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# Tidy modules
.PHONY: tidy
tidy:
	go mod tidy