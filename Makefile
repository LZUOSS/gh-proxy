# GitHub Reverse Proxy Makefile

# Variables
BINARY_NAME=github-proxy
MAIN_PATH=./cmd/server
BUILD_DIR=./build
CONFIG_PATH=./configs/config.yaml
DOCKER_IMAGE=github-proxy
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

.PHONY: all build test clean run install lint docker-build docker-run help

# Default target
all: test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH) -config $(CONFIG_PATH)

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed"

# Update dependencies
update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "Dependencies updated"

# Run linters
lint:
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found, installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}
	golangci-lint run ./...
	@echo "Linting complete"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Formatting complete"

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet complete"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -d \
		-p 8080:8080 \
		-p 2222:2222 \
		-p 9090:9090 \
		--name $(BINARY_NAME) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Docker container started"

# Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME)
	docker rm $(BINARY_NAME)
	@echo "Docker container stopped"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"

# Create release packages
release: build-all
	@echo "Creating release packages..."
	@cd $(BUILD_DIR) && \
		tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
		tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release packages created in $(BUILD_DIR)"

# Install the binary to system
install-binary: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Display help
help:
	@echo "GitHub Reverse Proxy - Makefile targets:"
	@echo ""
	@echo "  make build          - Build the application"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make run            - Run the application"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install        - Install dependencies"
	@echo "  make update         - Update dependencies"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make docker-stop    - Stop Docker container"
	@echo "  make build-all      - Build for multiple platforms"
	@echo "  make release        - Create release packages"
	@echo "  make install-binary - Install binary to system"
	@echo "  make help           - Display this help message"
	@echo ""
