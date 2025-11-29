# Gotlai Makefile

# Version info
VERSION := $(shell grep 'Version = "' version.go | head -1 | cut -d'"' -f2)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Build flags
LDFLAGS := -X github.com/ZaguanLabs/gotlai.GitCommit=$(GIT_COMMIT) \
           -X github.com/ZaguanLabs/gotlai.GitBranch=$(GIT_BRANCH) \
           -X github.com/ZaguanLabs/gotlai.BuildDate=$(BUILD_DATE) \
           -X github.com/ZaguanLabs/gotlai.GoVersion=$(GO_VERSION)
LDFLAGS_RELEASE := -s -w $(LDFLAGS)

# Output directories
DIST_DIR := dist
BIN_DIR := bin

.PHONY: all build test clean install lint version help
.PHONY: build-linux-amd64 build-linux-arm64 build-darwin-arm64 build-windows-amd64 build-windows-arm64
.PHONY: build-all dist

all: build

## build: Build the binary for current platform
build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/gotlai ./cmd/gotlai

## build-release: Build optimized release binary for current platform
build-release:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS_RELEASE)" -o $(BIN_DIR)/gotlai ./cmd/gotlai

## build-linux-amd64: Build for Linux amd64
build-linux-amd64:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS_RELEASE)" -o $(DIST_DIR)/gotlai-linux-amd64 ./cmd/gotlai

## build-linux-arm64: Build for Linux arm64
build-linux-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS_RELEASE)" -o $(DIST_DIR)/gotlai-linux-arm64 ./cmd/gotlai

## build-darwin-arm64: Build for macOS arm64 (Apple Silicon)
build-darwin-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS_RELEASE)" -o $(DIST_DIR)/gotlai-darwin-arm64 ./cmd/gotlai

## build-windows-amd64: Build for Windows amd64
build-windows-amd64:
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS_RELEASE)" -o $(DIST_DIR)/gotlai-windows-amd64.exe ./cmd/gotlai

## build-windows-arm64: Build for Windows arm64
build-windows-arm64:
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS_RELEASE)" -o $(DIST_DIR)/gotlai-windows-arm64.exe ./cmd/gotlai

## build-all: Build for all platforms
build-all: build-linux-amd64 build-linux-arm64 build-darwin-arm64 build-windows-amd64 build-windows-arm64

## dist: Build all platforms and create release archives
dist: build-all
	@echo "Creating release archives..."
	@cd $(DIST_DIR) && tar -czf gotlai-$(VERSION)-linux-amd64.tar.gz gotlai-linux-amd64
	@cd $(DIST_DIR) && tar -czf gotlai-$(VERSION)-linux-arm64.tar.gz gotlai-linux-arm64
	@cd $(DIST_DIR) && tar -czf gotlai-$(VERSION)-darwin-arm64.tar.gz gotlai-darwin-arm64
	@cd $(DIST_DIR) && zip -q gotlai-$(VERSION)-windows-amd64.zip gotlai-windows-amd64.exe
	@cd $(DIST_DIR) && zip -q gotlai-$(VERSION)-windows-arm64.zip gotlai-windows-arm64.exe
	@echo "Release archives created in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.zip

## test: Run all tests
test:
	go test -v ./...

## test-cover: Run tests with coverage
test-cover:
	go test -cover ./...

## bench: Run benchmarks
bench:
	go test -bench=. -benchmem ./...

## lint: Run linters
lint:
	go vet ./...
	@which golint > /dev/null && golint ./... || echo "golint not installed"

## clean: Remove build artifacts
clean:
	rm -rf $(BIN_DIR)/ $(DIST_DIR)/
	go clean

## install: Install to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/gotlai

## version: Show version info
version:
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
