.PHONY: build build-mcp run test clean tidy lint fmt

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/krubenok/toolbox/internal/cli.Version=$(VERSION) \
                     -X github.com/krubenok/toolbox/internal/cli.Commit=$(COMMIT) \
                     -X github.com/krubenok/toolbox/internal/cli.BuildDate=$(BUILD_DATE)"

# Default target
all: build build-mcp

# Build the CLI binary
build:
	go build $(LDFLAGS) -o bin/toolbox ./cmd/toolbox

# Build the MCP server binary
build-mcp:
	go build -o bin/toolbox-mcp ./cmd/toolbox-mcp

# Run the CLI
run:
	go run $(LDFLAGS) ./cmd/toolbox $(ARGS)

# Run tests
test:
	go test -v ./...

# Run linter (requires: brew install golangci-lint)
lint:
	mkdir -p .cache/gocache .cache/gomod .cache/golangci-lint
	GOCACHE="$(PWD)/.cache/gocache" GOMODCACHE="$(PWD)/.cache/gomod" GOLANGCI_LINT_CACHE="$(PWD)/.cache/golangci-lint" golangci-lint run

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/ dist/

# Tidy dependencies
tidy:
	go mod tidy

# Install locally
install: build build-mcp
	go install $(LDFLAGS) ./cmd/toolbox
	go install ./cmd/toolbox-mcp

# Release (dry run, requires: brew install goreleaser)
release-dry:
	goreleaser release --snapshot --clean
