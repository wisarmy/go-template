# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGENERATE=$(GOCMD) generate
BINARY_NAME=go-template
MAIN_PATH=./main.go

# Build directory
BUILD_DIR=./bin

# Default target
.PHONY: all
all: deps generate swagger test build

# Build the project
.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Test the project
.PHONY: test
test:
	$(GOTEST) -v ./...

# Update dependencies
.PHONY: deps
deps:
	$(GOMOD) tidy

# Generate code (for Ent ORM)
.PHONY: generate
generate:
	$(GOGENERATE) ./ent

# Generate Swagger documentation
.PHONY: swagger
swagger:
	swag init -g internal/api/server.go -o docs

# Help command
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make              - Run test and build"
	@echo "  make build        - Build the binary"
	@echo "  make test         - Run tests"
	@echo "  make deps         - Update dependencies"
	@echo "  make generate     - Generate Ent code"
	@echo "  make swagger      - Generate Swagger documentation"
	@echo "  make help         - Show this help"
