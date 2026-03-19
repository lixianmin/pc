.PHONY: all build build-plugins test clean run fmt vet lint help

# Build variables
BINARY_NAME=pc
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

all: fmt vet test build build-plugins

build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/pc

PLUGIN_DIR=$(HOME)/.pc/plugins/llm/openai

build-plugins:
	@echo "Building plugins..."
	@mkdir -p $(PLUGIN_DIR)/bin
	@cd examples/plugins/llm/openai/cmd/openai-llm && \
		if ! grep -q "github.com/lixianmin/pc" go.mod 2>/dev/null; then \
			echo "require github.com/lixianmin/pc v0.0.0" >> go.mod && \
			echo "replace github.com/lixianmin/pc => ../../../../../../" >> go.mod; \
		fi && \
		go mod tidy && \
		go build -o $(PLUGIN_DIR)/bin/openai-llm .
	@echo "Plugins built to $(PLUGIN_DIR)/bin/"

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf bin/
	@rm -f coverage.out coverage.html

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Vetting code..."
	go vet ./...

lint:
	@echo "Linting code..."
	@golangci-lint run ./... || echo "golangci-lint not installed, skipping"

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

help:
	@echo "Usage:"
	@echo "  make build        - Build the binary"
	@echo "  make build-plugins- Build all plugins"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make run          - Build and run"
	@echo "  make fmt          - Format code"
	@echo "  make vet          - Vet code"
	@echo "  make lint         - Run linter"
	@echo "  make deps         - Download and tidy dependencies"
	@echo "  make all          - fmt, vet, test, build, build-plugins"
