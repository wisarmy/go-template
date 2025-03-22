# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=go-template
MAIN_PATH=./main.go

# Build directory
BUILD_DIR=./bin

# Default target
.PHONY: all
all: test build

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

# Help command
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make              - Run test and build"
	@echo "  make build        - Build the binary"
	@echo "  make test         - Run tests"
	@echo "  make deps         - Update dependencies"
	@echo "  make help         - Show this help"
