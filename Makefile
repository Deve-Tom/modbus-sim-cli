
# Modbus-Sim Makefile
# Cross-platform build support for Windows, Linux, and macOS

.PHONY: all build build-all build-windows build-linux build-linux-arm build-linux-arm64 build-macos clean test install uninstall help

# Version information
VERSION ?= 1.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y%m%d)

# Go parameters
GOFLAGS ?=
LDFLAGS ?= -X main.Version=$(VERSION) -X main.Commit=$(COMMIT)

# Output directory
OUTPUT_DIR ?= build

# Binary names
BINARY_WINDOWS := modbus-sim.exe
BINARY_LINUX := modbus-sim
BINARY_MACOS := modbus-sim

# Default target
all: build

# Build for current platform
build:
	@echo "Building for current platform..."
	@mkdir -p $(OUTPUT_DIR)
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_LINUX) .

# Build all platforms
build-all: build-windows build-linux build-linux-arm build-linux-arm64 build-macos
	@echo "All builds completed successfully"

# Build for Windows (amd64)
build-windows:
	@echo "Building for Windows amd64..."
	@mkdir -p $(OUTPUT_DIR)/windows_amd64
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/windows_amd64/$(BINARY_WINDOWS) .

# Build for Linux (amd64)
build-linux:
	@echo "Building for Linux amd64..."
	@mkdir -p $(OUTPUT_DIR)/linux_amd64
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/linux_amd64/$(BINARY_LINUX) .

# Build for Linux ARM (32-bit) - for embedded devices like Raspberry Pi
build-linux-arm:
	@echo "Building for Linux arm (32-bit)..."
	@mkdir -p $(OUTPUT_DIR)/linux_arm
	GOOS=linux GOARCH=arm GOARM=7 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/linux_arm/$(BINARY_LINUX) .

# Build for Linux ARM64 (64-bit) - for modern embedded devices
build-linux-arm64:
	@echo "Building for Linux arm64..."
	@mkdir -p $(OUTPUT_DIR)/linux_arm64
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/linux_arm64/$(BINARY_LINUX) .

# Build for macOS (amd64)
build-macos:
	@echo "Building for macOS amd64..."
	@mkdir -p $(OUTPUT_DIR)/darwin_amd64
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/darwin_amd64/$(BINARY_MACOS) .

# Build for macOS ARM64 (Apple Silicon)
build-macos-arm64:
	@echo "Building for macOS arm64..."
	@mkdir -p $(OUTPUT_DIR)/darwin_arm64
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/darwin_arm64/$(BINARY_MACOS) .

# Run tests
test:
	@echo "Running tests..."
	go test $(GOFLAGS) ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test $(GOFLAGS) -cover ./...

# Install to system (Linux/macOS only)
install: build-linux
	@echo "Installing to /usr/local/bin..."
	sudo cp $(OUTPUT_DIR)/linux_amd64/$(BINARY_LINUX) /usr/local/bin/modbus-sim

# Uninstall from system (Linux/macOS only)
uninstall:
	@echo "Uninstalling..."
	sudo rm -f /usr/local/bin/modbus-sim

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(OUTPUT_DIR)

# Show help
help:
	@echo "Modbus-Sim Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all              Build for current platform (default)"
	@echo "  build            Build for current platform"
	@echo "  build-all        Build for all platforms"
	@echo "  build-windows    Build for Windows amd64"
	@echo "  build-linux      Build for Linux amd64"
	@echo "  build-linux-arm  Build for Linux ARM (32-bit, e.g., Raspberry Pi)"
	@echo "  build-linux-arm64 Build for Linux ARM64"
	@echo "  build-macos      Build for macOS amd64"
	@echo "  build-macos-arm64 Build for macOS ARM64 (Apple Silicon)"
	@echo "  test             Run all tests"
	@echo "  test-coverage    Run tests with coverage report"
	@echo "  install          Install to /usr/local/bin (Linux/macOS)"
	@echo "  uninstall        Remove from /usr/local/bin (Linux/macOS)"
	@echo "  clean            Clean build artifacts"
	@echo "  help             Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION     Version number (default: 1.0.0)"
	@echo "  COMMIT      Git commit hash (auto-detected)"
	@echo "  GOFLAGS     Extra Go build flags"
	@echo "  OUTPUT_DIR  Output directory (default: build)"
