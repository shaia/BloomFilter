# Go SIMD Bloom Filter Makefile
# Project: github.com/shaia/go-simd-bloomfilter

# Variables
PACKAGE_PATH = .
EXAMPLE_PATH = ./docs/examples/basic
ASM_PATH = ./asm
BIN_DIR = ./bin
DIST_DIR = ./dist

# Build information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_USER ?= $(shell whoami)

# Go build flags
GO = go
GOFMT = gofmt
GOLINT = golangci-lint
GOBENCH = go test -bench=.
LDFLAGS = -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE) -X main.BuildUser=$(BUILD_USER)
BUILD_FLAGS = -ldflags "$(LDFLAGS)"

# Binary names
EXAMPLE_BINARY = basic-example
EXAMPLE_BINARY_VERSIONED = $(EXAMPLE_BINARY)-$(VERSION)

# Default target
.PHONY: all
all: clean fmt lint test build example

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all         - Run clean, fmt, lint, test, build, example"
	@echo "  build       - Build the library"
	@echo "  binaries    - Build all binary artifacts"
	@echo "  example     - Build and run the basic example"
	@echo "  install     - Install binaries to GOPATH/bin"
	@echo "  dist        - Create distribution packages"
	@echo "  release     - Create release artifacts"
	@echo "  test        - Run all tests"
	@echo "  test-pure   - Run tests with pure Go (no SIMD)"
	@echo "  bench       - Run benchmarks"
	@echo "  bench-all   - Run benchmarks for both SIMD and pure Go"
	@echo "  fmt         - Format all Go code"
	@echo "  lint        - Run linter"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Download dependencies"
	@echo "  tidy        - Tidy go.mod"
	@echo "  coverage    - Generate test coverage report"
	@echo "  profile     - Run CPU profiling benchmark"
	@echo "  version     - Show version information"

# Version and build info
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Build User: $(BUILD_USER)"

# Build targets
.PHONY: build
build:
	@echo "Building bloom filter library..."
	cd $(PACKAGE_PATH) && $(GO) build -v .

.PHONY: binaries
binaries: $(BIN_DIR)
	@echo "Building all binary artifacts..."
	@echo "Building basic example..."
	cd $(EXAMPLE_PATH) && $(GO) build $(BUILD_FLAGS) -o ../../../$(BIN_DIR)/$(EXAMPLE_BINARY) .
	@echo "Binary created: $(BIN_DIR)/$(EXAMPLE_BINARY)"

.PHONY: example
example: binaries
	@echo "Running basic example..."
	./$(BIN_DIR)/$(EXAMPLE_BINARY)

# Install targets
.PHONY: install
install:
	@echo "Installing binaries to GOPATH/bin..."
	cd $(EXAMPLE_PATH) && $(GO) install $(BUILD_FLAGS) .

# Create directories
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

dist-dir:
	@mkdir -p $(DIST_DIR)

# Test targets
.PHONY: test
test:
	@echo "Running tests with SIMD optimizations..."
	cd $(PACKAGE_PATH) && $(GO) test -v -race .

.PHONY: test-pure
test-pure:
	@echo "Running tests with pure Go (no SIMD)..."
	cd $(PACKAGE_PATH) && $(GO) test -v -race -tags purego .

.PHONY: test-all
test-all: test test-pure

# Benchmark targets
.PHONY: bench
bench:
	@echo "Running benchmarks with SIMD optimizations..."
	cd $(PACKAGE_PATH) && $(GOBENCH) -benchmem .

.PHONY: bench-pure
bench-pure:
	@echo "Running benchmarks with pure Go (no SIMD)..."
	cd $(PACKAGE_PATH) && $(GOBENCH) -benchmem -tags purego .

.PHONY: bench-all
bench-all:
	@echo "Running benchmarks comparison..."
	@echo "=== SIMD Optimized ==="
	cd $(PACKAGE_PATH) && $(GOBENCH) -benchmem .
	@echo ""
	@echo "=== Pure Go ==="
	cd $(PACKAGE_PATH) && $(GOBENCH) -benchmem -tags purego .

.PHONY: bench-compare
bench-compare:
	@echo "Running benchmark comparison with benchstat..."
	cd $(PACKAGE_PATH) && $(GOBENCH) -benchmem -count=5 . > simd_bench.txt
	cd $(PACKAGE_PATH) && $(GOBENCH) -benchmem -count=5 -tags purego . > pure_bench.txt
	@echo "Benchmark results saved to simd_bench.txt and pure_bench.txt"
	@echo "Install benchstat: go install golang.org/x/perf/cmd/benchstat@latest"
	@echo "Compare with: benchstat pure_bench.txt simd_bench.txt"

# Code quality targets
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) -s -w $(PACKAGE_PATH)
	$(GOFMT) -s -w $(EXAMPLE_PATH)

.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run .
	$(GOLINT) run $(EXAMPLE_PATH)/...

# Coverage targets
.PHONY: coverage
coverage:
	@echo "Generating test coverage report..."
	cd $(PACKAGE_PATH) && $(GO) test -coverprofile=coverage.out .
	cd $(PACKAGE_PATH) && $(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: $(PACKAGE_PATH)/coverage.html"

.PHONY: coverage-pure
coverage-pure:
	@echo "Generating test coverage report (pure Go)..."
	cd $(PACKAGE_PATH) && $(GO) test -coverprofile=coverage_pure.out -tags purego .
	cd $(PACKAGE_PATH) && $(GO) tool cover -html=coverage_pure.out -o coverage_pure.html
	@echo "Coverage report generated: $(PACKAGE_PATH)/coverage_pure.html"

# Profiling targets
.PHONY: profile
profile:
	@echo "Running CPU profiling benchmark..."
	cd $(PACKAGE_PATH) && $(GO) test -bench=BenchmarkAdd -cpuprofile=cpu.prof -memprofile=mem.prof .
	@echo "CPU profile: $(PACKAGE_PATH)/cpu.prof"
	@echo "Memory profile: $(PACKAGE_PATH)/mem.prof"
	@echo "View with: go tool pprof cpu.prof"

# Dependency management
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

.PHONY: tidy
tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy

.PHONY: vendor
vendor:
	@echo "Vendoring dependencies..."
	$(GO) mod vendor

# Build variations
.PHONY: build-race
build-race:
	@echo "Building with race detector..."
	cd $(PACKAGE_PATH) && $(GO) build -race -v .

.PHONY: build-tags
build-tags:
	@echo "Building with different build tags..."
	@echo "Building with SIMD optimizations..."
	cd $(PACKAGE_PATH) && $(GO) build -v .
	@echo "Building with pure Go..."
	cd $(PACKAGE_PATH) && $(GO) build -tags purego -v .

# Cross-compilation targets
.PHONY: build-all-platforms
build-all-platforms: $(BIN_DIR)
	@echo "Building for all platforms..."
	@echo "Building for Linux AMD64..."
	cd $(EXAMPLE_PATH) && GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o ../../../$(BIN_DIR)/$(EXAMPLE_BINARY)-linux-amd64 .
	@echo "Building for Linux ARM64..."
	cd $(EXAMPLE_PATH) && GOOS=linux GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o ../../../$(BIN_DIR)/$(EXAMPLE_BINARY)-linux-arm64 .
	@echo "Building for macOS AMD64..."
	cd $(EXAMPLE_PATH) && GOOS=darwin GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o ../../../$(BIN_DIR)/$(EXAMPLE_BINARY)-darwin-amd64 .
	@echo "Building for macOS ARM64..."
	cd $(EXAMPLE_PATH) && GOOS=darwin GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o ../../../$(BIN_DIR)/$(EXAMPLE_BINARY)-darwin-arm64 .
	@echo "Building for Windows AMD64..."
	cd $(EXAMPLE_PATH) && GOOS=windows GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o ../../../$(BIN_DIR)/$(EXAMPLE_BINARY)-windows-amd64.exe .

# Distribution targets
.PHONY: dist
dist: build-all-platforms dist-dir
	@echo "Creating distribution packages..."
	@for platform in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64; do \
		echo "Creating package for $$platform..."; \
		if [ "$$platform" = "windows-amd64" ]; then \
			tar -czf $(DIST_DIR)/$(EXAMPLE_BINARY)-$(VERSION)-$$platform.tar.gz -C $(BIN_DIR) $(EXAMPLE_BINARY)-$$platform.exe; \
		else \
			tar -czf $(DIST_DIR)/$(EXAMPLE_BINARY)-$(VERSION)-$$platform.tar.gz -C $(BIN_DIR) $(EXAMPLE_BINARY)-$$platform; \
		fi; \
	done
	@echo "Distribution packages created in $(DIST_DIR)/"

.PHONY: release
release: test lint dist
	@echo "Creating release artifacts..."
	@echo "Release $(VERSION) created successfully!"
	@echo "Artifacts:"
	@ls -la $(DIST_DIR)/
	@echo ""
	@echo "Checksums:"
	@cd $(DIST_DIR) && sha256sum *.tar.gz > checksums.txt
	@cat $(DIST_DIR)/checksums.txt

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean -cache
	rm -rf $(BIN_DIR)
	rm -rf $(DIST_DIR)
	rm -f $(PACKAGE_PATH)/coverage.out $(PACKAGE_PATH)/coverage.html
	rm -f $(PACKAGE_PATH)/coverage_pure.out $(PACKAGE_PATH)/coverage_pure.html
	rm -f $(PACKAGE_PATH)/cpu.prof $(PACKAGE_PATH)/mem.prof
	rm -f $(PACKAGE_PATH)/simd_bench.txt $(PACKAGE_PATH)/pure_bench.txt
	rm -f $(EXAMPLE_PATH)/bloom-example
	find . -name "*.test" -delete

# Development targets
.PHONY: dev
dev:
	@echo "Setting up development environment..."
	$(GO) install golang.org/x/perf/cmd/benchstat@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: check
check: fmt lint test-all
	@echo "All checks passed!"

# Watch target (requires entr: brew install entr)
.PHONY: watch
watch:
	@echo "Watching for changes... (requires 'entr': brew install entr)"
	find $(PACKAGE_PATH) -name "*.go" | entr -c make test

# Docker targets (optional)
.PHONY: docker-test
docker-test:
	@echo "Running tests in Docker..."
	docker run --rm -v $(PWD):/workspace -w /workspace golang:1.23 make test-all

# Documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	cd $(PACKAGE_PATH) && $(GO) doc -all .

.PHONY: godoc
godoc:
	@echo "Starting godoc server..."
	@echo "Visit http://localhost:6060/pkg/github.com/shaia/go-simd-bloomfilter/"
	godoc -http=:6060