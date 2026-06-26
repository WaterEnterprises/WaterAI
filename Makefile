# ==============================================================================
# WATER AI - BUILD CONFIGURATION
# ==============================================================================

BINARY_NAME := water-ai
VERSION     := v0.1.1

# Directories
DIST_DIR     := dist
BIN_DIR      := bin
FRONTEND_DIR := frontend
EMBED_DIR    := cmd/water/embed/out

# Go Build Flags
LDFLAGS  := -s -w -X 'main.Version=$(VERSION)' -extldflags "-static"
GO_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

# Platforms to build for (OS/ARCH)
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# FIX 1: Added 'frontend' here so Make ignores the folder of the same name
.PHONY: all build frontend clean release test test-race test-coverage help mocks

# Default target
all: clean test frontend release

# Help: List available commands
help:
	@echo "Water AI Makefile"
	@echo "-----------------"
	@echo "make all           - Clean, test, build frontend, and build release binaries"
	@echo "make build         - Build binary for current OS (fast dev build)"
	@echo "make release       - Build optimized binaries for all platforms"
	@echo "make frontend      - Build Next.js frontend and copy to embed folder"
	@echo "make test          - Run standard unit tests"
	@echo "make test-race     - Run tests with race detection"
	@echo "make test-coverage - Run tests and generate HTML coverage report"
	@echo "make clean         - Remove build artifacts"

# ==============================================================================
# TESTING
# ==============================================================================

mocks:
	@echo "--> Creating dummy assets for testing..."
	@mkdir -p cmd/water/embed/out
	@touch cmd/water/embed/out/.keep
	@mkdir -p browser/embed
	@touch browser/embed/OpenSans-Medium.ttf
	@touch browser/embed/findVisibleInteractiveElements.js

test: mocks
	@echo "--> Running Go dependencies check..."
	@go mod tidy
	@echo "--> Running static analysis (go vet)..."
	@go vet ./...
	@echo "--> Running unit tests..."
	@go test ./... -v -count=1

test-race: mocks
	@echo "--> Running tests with race detector..."
	@CGO_ENABLED=1 go test ./... -race -v -count=1

test-coverage: mocks
	@echo "--> Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out -count=1
	@if command -v go >/dev/null 2>&1; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "Coverage report generated: coverage.html"; \
	else \
		echo "Coverage report: coverage.out"; \
	fi

# ==============================================================================
# FRONTEND BUILD
# ==============================================================================

frontend:
	@echo "--> Building Next.js frontend..."
	
	@# Safety Check: API routes are incompatible with static export
	@if [ -d "$(FRONTEND_DIR)/app/api" ]; then \
		echo "ERROR: Found API routes in $(FRONTEND_DIR)/app/api"; \
		echo "Next.js 'output: export' does not support API routes. Move logic to Go."; \
		exit 1; \
	fi

	@# Install Dependencies
	@cd $(FRONTEND_DIR) && yarn install --frozen-lockfile --silent

	@# Build Static Export with localhost:7777 as the API/WS target
	@cd $(FRONTEND_DIR) && \
		NEXT_TELEMETRY_DISABLED=1 \
		NEXT_DISABLE_WORKER_THREADS=1 \
		NEXT_PUBLIC_API_URL=http://localhost:7777 \
		yarn build

	@# Prepare Embed Directory
	@echo "--> Embedding assets into Go..."
	@rm -rf $(EMBED_DIR)
	@mkdir -p $(EMBED_DIR)
	@cp -r $(FRONTEND_DIR)/out/. $(EMBED_DIR)/
	@touch $(EMBED_DIR)/.keep

# ==============================================================================
# BACKEND BUILD (LOCAL)
# ==============================================================================

build: frontend
	@echo "--> Building local binary..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 go build $(GO_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/water
	@echo "Done! Run with: ./$(BIN_DIR)/$(BINARY_NAME)"

# ==============================================================================
# RELEASE BUILD (CROSS-PLATFORM)
# ==============================================================================

release: frontend
	@echo "--> Building optimized binaries for platforms: $(PLATFORMS)"
	@mkdir -p $(DIST_DIR)
	@for p in $(PLATFORMS); do \
		os=$$(echo "$$p" | cut -d'/' -f1); \
		arch=$$(echo "$$p" | cut -d'/' -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		target="$(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext"; \
		echo "    Building $$target ..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -a $(GO_FLAGS) -o $$target ./cmd/water; \
	done
	@echo "--> Release builds completed in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

# ==============================================================================
# CLEANUP
# ==============================================================================

clean:
	@echo "--> Cleaning up..."
	@rm -rf $(DIST_DIR)
	@rm -rf $(BIN_DIR)
	@rm -rf $(EMBED_DIR)
	@rm -rf $(FRONTEND_DIR)/out
	@rm -rf $(FRONTEND_DIR)/.next
	@rm -f coverage.out coverage.html
	@echo "Clean complete."