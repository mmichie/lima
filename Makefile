# Lima - Beancount TUI Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Binary names
BINARY_NAME=lima
CATEGORIZER_DEMO=categorizer-demo

# Build directory
BUILD_DIR=build

# Version info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all build build-all clean test coverage fmt vet install demo help

# Default target
all: test build

# Build the main lima binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/lima
	@echo "✓ Built $(BINARY_NAME)"

# Build all binaries
build-all: build demo
	@echo "✓ Built all binaries"

# Build the categorizer demo
demo:
	@echo "Building $(CATEGORIZER_DEMO)..."
	$(GOBUILD) -o $(CATEGORIZER_DEMO) ./cmd/categorizer-demo
	@echo "✓ Built $(CATEGORIZER_DEMO)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "✓ Code formatted"

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "✓ Vet complete"

# Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "✓ All checks passed"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(CATEGORIZER_DEMO)
	rm -f coverage.out coverage.html
	rm -rf $(BUILD_DIR)
	@echo "✓ Clean complete"

# Install to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) ./cmd/lima
	@echo "✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✓ Dependencies updated"

# Run the demo
run-demo: demo
	@echo "Running categorizer demo..."
	./$(CATEGORIZER_DEMO)

# Run lima with example data (requires testdata/sample.beancount)
run: build
	@if [ -f testdata/sample.beancount ]; then \
		./$(BINARY_NAME) testdata/sample.beancount; \
	else \
		echo "Error: testdata/sample.beancount not found"; \
		echo "Usage: make run LEDGER=/path/to/your/ledger.beancount"; \
	fi

# Run lima with custom ledger file
run-ledger: build
	@if [ -z "$(LEDGER)" ]; then \
		echo "Error: LEDGER variable not set"; \
		echo "Usage: make run-ledger LEDGER=/path/to/your/ledger.beancount"; \
		exit 1; \
	fi
	./$(BINARY_NAME) $(LEDGER)

# Build for multiple platforms
build-cross:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/lima
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/lima
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/lima
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/lima
	@echo "✓ Cross-platform builds complete in $(BUILD_DIR)/"

# Show help
help:
	@echo "Lima - Beancount TUI Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all              Build and test (default)"
	@echo "  build            Build lima binary"
	@echo "  build-all        Build all binaries (lima + demos)"
	@echo "  demo             Build categorizer demo"
	@echo "  test             Run tests"
	@echo "  coverage         Run tests with coverage report"
	@echo "  fmt              Format code with gofmt"
	@echo "  vet              Run go vet"
	@echo "  check            Run fmt, vet, and test"
	@echo "  clean            Remove build artifacts"
	@echo "  install          Install to GOPATH/bin"
	@echo "  deps             Download and tidy dependencies"
	@echo "  run-demo         Build and run categorizer demo"
	@echo "  run              Run lima with testdata/sample.beancount"
	@echo "  run-ledger       Run lima with custom ledger (LEDGER=/path/to/file)"
	@echo "  build-cross      Build for multiple platforms"
	@echo "  help             Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make                          # Build and test"
	@echo "  make build                    # Just build"
	@echo "  make run-demo                 # Run categorizer demo"
	@echo "  make run-ledger LEDGER=~/finances/main.beancount"
	@echo "  make coverage                 # Generate coverage report"
	@echo ""
