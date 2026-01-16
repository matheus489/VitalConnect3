# VitalConnect PEP Agent Makefile

# Build variables
BINARY_NAME=pep-agent
VERSION=1.0.0
BUILD_DIR=./build
MAIN_PATH=./cmd/pep-agent

# Go build flags for portable binary
GO_BUILD_FLAGS=-ldflags="-s -w" -trimpath
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64

.PHONY: all build build-linux test clean run validate help deps

# Default target
all: test build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build portable Linux binary (for hospital deployment)
build-linux:
	@echo "Building portable Linux binary..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f pep-agent-state.json

# Run the agent (development)
run: build
	@echo "Running agent..."
	$(BUILD_DIR)/$(BINARY_NAME) -config mapping.example.yaml

# Validate configuration
validate: build
	@echo "Validating configuration..."
	$(BUILD_DIR)/$(BINARY_NAME) -config mapping.example.yaml -validate

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Show help
help:
	@echo "VitalConnect PEP Agent Build System"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build for current platform"
	@echo "  build-linux  Build portable Linux AMD64 binary"
	@echo "  test         Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  clean        Remove build artifacts"
	@echo "  run          Build and run (development)"
	@echo "  validate     Validate configuration file"
	@echo "  deps         Download and tidy dependencies"
	@echo "  help         Show this help message"
