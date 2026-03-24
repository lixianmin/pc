.PHONY: all generate build build-plugins test clean run fmt vet lint help

BINARY_NAME=pc
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

all: fmt vet test build build-plugins

generate:
	@echo "Generating BAML client..."
	@$(shell go env GOPATH)/bin/baml-cli generate 2>/dev/null || echo "baml-cli not found, skipping BAML generation"

build: generate
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/pc

PLUGIN_DIR=$(HOME)/.pc/plugins/channel

build-plugins:
	@echo "Building channel plugins..."
	@mkdir -p $(PLUGIN_DIR)/telegram/bin
	@if [ -d examples/plugins/channel/telegram ]; then \
		cd examples/plugins/channel/telegram/cmd/telegram-bot && \
		go build -o $(PLUGIN_DIR)/telegram/bin/telegram-bot .; \
	fi
	@echo "Channel plugins built to $(PLUGIN_DIR)/"

test: generate
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
	@echo "  make generate      - Generate BAML client code"
	@echo "  make build         - Build the binary (includes generate)"
	@echo "  make build-plugins - Build channel plugins only"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make run           - Build and run"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Vet code"
	@echo "  make lint          - Run linter"
	@echo "  make deps          - Download and tidy dependencies"
	@echo "  make all           - fmt, vet, test, build, build-plugins"
