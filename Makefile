# ArchTerm - Makefile
# THE ARCHITECT'S CODEX Development Makefile

BINARY_NAME=archterm
GO_FLAGS=-ldflags="-s -w"
GO_PATH=$(shell go env GOPATH)

# Colors for output
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m

.PHONY: all build run clean test dev install help \
        lint fmt vet mod-tidy \
        watch serve \
        py-install py-test \
        web-build web-dev \
        docs \
        dev-setup

all: build

# Build the application
build:
	@echo "$(BLUE)[ArchTerm]$(NC) Building..."
	go build $(GO_FLAGS) -o $(BINARY_NAME) ./cmd/archterm/
	@echo "$(GREEN)[ArchTerm]$(NC) Build complete: ./$($(BINARY_NAME))"

# Run the application
run: build
	@echo "$(BLUE)[ArchTerm]$(NC) Starting server..."
	./$(BINARY_NAME) -port 8080

# Clean build artifacts
clean:
	@echo "$(YELLOW)[ArchTerm]$(NC) Cleaning..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "$(GREEN)[ArchTerm]$(NC) Clean complete"

# Run tests
test:
	@echo "$(BLUE)[ArchTerm]$(NC) Running tests..."
	go test -v ./...

# Development mode (build + run)
dev: build
	@echo "$(GREEN)[ArchTerm]$(NC) Starting development server on :8080..."
	./$(BINARY_NAME) -port 8080

# Install the binary to GOPATH/bin
install:
	@echo "$(BLUE)[ArchTerm]$(NC) Installing..."
	go install ./cmd/archterm/
	@echo "$(GREEN)[ArchTerm]$(NC) Installed to $(GO_PATH)/bin/$(BINARY_NAME)"

# Linting
lint:
	@echo "$(BLUE)[ArchTerm]$(NC) Linting..."
	go vet ./...
	@echo "$(GREEN)[ArchTerm]$(NC) Lint complete"

# Format code
fmt:
	@echo "$(BLUE)[ArchTerm]$(NC) Formatting..."
	go fmt ./...
	@echo "$(GREEN)[ArchTerm]$(NC) Format complete"

# Run go mod tidy
mod-tidy:
	@echo "$(BLUE)[ArchTerm]$(NC) Tidy modules..."
	go mod tidy
	@echo "$(GREEN)[ArchTerm]$(NC) Modules tidy"

# Watch for file changes and rebuild
watch:
	@echo "$(BLUE)[ArchTerm]$(NC) Starting file watcher..."
	go install github.com/githubnemo/Watch@latest
	watch -c make build

# Serve (alias for dev)
serve: dev

# Install Python dependencies
py-install:
	@echo "$(BLUE)[ArchTerm]$(NC) Installing Python dependencies..."
	pip3 install -r python_parser/requirements.txt
	@echo "$(GREEN)[ArchTerm]$(NC) Python dependencies installed"

# Test Python parser
py-test:
	@echo "$(BLUE)[ArchTerm]$(NC) Testing Python parser..."
	python3 -c "import sys; print('Python', sys.version)"
	python3 python_parser/parser.py --help || echo "Parser requires --input and --output arguments"

# Build web assets (placeholder for future bundling)
web-build:
	@echo "$(BLUE)[ArchTerm]$(NC) Web assets are static - no build needed"
	@echo "$(GREEN)[ArchTerm]$(NC) Web: webapp/"

# Development mode for web (alias)
web-dev:
	@echo "$(BLUE)[ArchTerm]$(NC) Web development mode"
	@echo "Open http://localhost:8080 in your browser"

# Generate documentation
docs:
	@echo "$(BLUE)[ArchTerm]$(NC) Generating documentation..."
	@mkdir -p docs/api
	godoc -http=:6060 &
	@echo "$(GREEN)[ArchTerm]$(NC) Docs server started at http://localhost:6060"
	@echo "Press Ctrl+C to stop"

# Full development setup
dev-setup: mod-tidy py-install
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  ArchTerm Development Environment$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(BLUE)  Go:       $(shell go version)$(NC)"
	@echo "$(BLUE)  Python:   $(shell python3 --version)$(NC)"
	@echo "$(BLUE)  Binary:   ./$($(BINARY_NAME))$(NC)"
	@echo "$(BLUE)  Server:   http://localhost:8080$(NC)"
	@echo ""
	@echo "$(YELLOW)  Run 'make dev' to start the server$(NC)"
	@echo "$(YELLOW)  Run 'make help' for more commands$(NC)"
	@echo ""

# Show help
help:
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  ArchTerm - Available Commands$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(BLUE)Build & Run:$(NC)"
	@echo "  make build     - Build the application"
	@echo "  make run       - Build and run"
	@echo "  make dev       - Development mode (build + run)"
	@echo "  make clean     - Clean build artifacts"
	@echo ""
	@echo "$(BLUE)Development:$(NC)"
	@echo "  make test      - Run tests"
	@echo "  make lint      - Run go vet"
	@echo "  make fmt       - Format code"
	@echo "  make mod-tidy  - Tidy Go modules"
	@echo "  make watch     - Watch for file changes"
	@echo ""
	@echo "$(BLUE)Python:$(NC)"
	@echo "  make py-install - Install Python dependencies"
	@echo "  make py-test    - Test Python parser"
	@echo ""
	@echo "$(BLUE)Setup:$(NC)"
	@echo "  make dev-setup - Full development environment setup"
	@echo "  make install   - Install to GOPATH"
	@echo ""
	@echo "$(BLUE)Other:$(NC)"
	@echo "  make help      - Show this help"
	@echo ""
