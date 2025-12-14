.PHONY: build run test clean tidy lint fmt

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/kyrubeno/toolbox/internal/cli.Version=$(VERSION) \
                     -X github.com/kyrubeno/toolbox/internal/cli.Commit=$(COMMIT) \
                     -X github.com/kyrubeno/toolbox/internal/cli.BuildDate=$(BUILD_DATE)"

# Default target
all: build

# Build the binary
build:
	go build $(LDFLAGS) -o bin/toolbox ./cmd/toolbox

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
install: build
	go install $(LDFLAGS) ./cmd/toolbox

# Release (dry run, requires: brew install goreleaser)
release-dry:
	goreleaser release --snapshot --clean
