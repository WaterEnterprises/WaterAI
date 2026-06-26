# ==============================================================================
# WATER AI — UNIFIED BUILD SYSTEM
# ==============================================================================
#
# Produces a SINGLE binary that starts both the gateway/backend service AND
# the Fyne GUI.  CGO_ENABLED=1 is required for Fyne.
#
# Usage:
#   make build              — Build for current platform
#   make build-linux        — Cross-compile for linux/amd64
#   make build-darwin       — Cross-compile for darwin/amd64
#   make build-windows      — Cross-compile for windows/amd64
#   make test               — Run tests with race detection + coverage
#   make release            — Auto-detect OS and build native release artifacts
#   make release-linux      — Build linux .run self-extracting installers
#   make release-darwin     — Build macOS .dmg disk images
#   make release-windows    — Build windows .exe with embedded icon
#   make compress           — UPX-compress binaries in dist/ (optional)
#   make clean              — Remove build artifacts
#   make help               — Show this help
#
# Cross-compilation notes (Fyne / CGO):
#   Linux  → needs gcc, pkg-config, libgl1-mesa-dev, libxcursor-dev, etc.
#   macOS  → needs Xcode Command Line Tools (native) or osxcross (cross)
#   Windows→ needs x86_64-w64-mingw32-gcc (MinGW-w64)
#
# ==============================================================================

# --- Shell -------------------------------------------------------------------
# On Windows (MSYS2/Git Bash), /bin/bash may not exist; use 'bash' from PATH.
ifeq ($(OS),Windows_NT)
    SHELL      := bash
else
    SHELL      := /bin/bash
endif
.SHELLFLAGS := -euo pipefail -c

# --- Project -----------------------------------------------------------------
BINARY     := Water
MODULE     := water-ai
CMD_PKG    := ./cmd/water
APP_ID     := ai.water.app
APP_ICON   := $(CURDIR)/resources/logo-only.png
ASSET_DIR  := $(CURDIR)/resources
ASSET_FILES := logo.png logo-only.png vscode.png

# --- Version info (injected via ldflags) -------------------------------------
VERSION    ?= v0.2.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version 2>/dev/null | awk '{print $$3}' || echo "unknown")

# --- Directories -------------------------------------------------------------
DIST_DIR     := dist
BIN_DIR      := bin
COVERAGE_DIR := coverage

# --- Test settings ------------------------------------------------------------
TEST_TIMEOUT := 15m

# --- Host detection ----------------------------------------------------------
UNAME_S := $(shell uname -s 2>/dev/null || echo "Windows_NT")
UNAME_M := $(shell uname -m 2>/dev/null || echo "x86_64")

ifeq ($(UNAME_S),Linux)
    GOOS_HOST := linux
else ifeq ($(UNAME_S),Darwin)
    GOOS_HOST := darwin
else ifneq (,$(findstring MINGW,$(UNAME_S)))
    GOOS_HOST := windows
else ifneq (,$(findstring MSYS,$(UNAME_S)))
    GOOS_HOST := windows
else ifneq (,$(findstring Windows,$(UNAME_S)))
    GOOS_HOST := windows
else
    GOOS_HOST := linux
endif

ifeq ($(filter $(UNAME_M),x86_64 amd64),)
    GOARCH_HOST := arm64
else
    GOARCH_HOST := amd64
endif

# Allow overriding via environment (useful in CI)
ifdef TARGET_OS
    GOOS_HOST := $(TARGET_OS)
endif
ifdef TARGET_ARCH
    GOARCH_HOST := $(TARGET_ARCH)
endif

# --- Build flags (ultra-optimised) -------------------------------------------
LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.GitCommit=$(GIT_COMMIT)' \
	-X 'main.BuildDate=$(BUILD_DATE)' \
	-X 'main.GoVersion=$(GO_VERSION)'

# Linux "mostly-static" linking flags: statically link the C runtime (libgcc,
# libstdc++) so the binary has no dependency on a specific glibc/libstdc++
# version, but leave libGL/libX11/libXcursor/… as dynamic since OpenGL is
# always a shared library provided by the GPU driver.  Mesa's software
# renderer (llvmpipe/swrast) is bundled alongside the binary as a fallback
# for systems without a hardware OpenGL driver.
LDFLAGS_LINUX := $(LDFLAGS) -linkmode external -extldflags '-static-libgcc -static-libstdc++'

# Windows static linking flags: statically link the C runtime so the binary
# is fully self-contained with no external DLL dependencies beyond system DLLs.
LDFLAGS_WINDOWS_STATIC := $(LDFLAGS) -linkmode external -extldflags '-static -lpthread -lm'

GO_BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"
GO_BUILD_FLAGS_LINUX := -trimpath -ldflags "$(LDFLAGS_LINUX)"
GO_BUILD_FLAGS_WINDOWS_STATIC := -trimpath -ldflags "$(LDFLAGS_WINDOWS_STATIC)"

# CGO is required for Fyne
export CGO_ENABLED := 1


# ==============================================================================
# PHONY
# ==============================================================================
.PHONY: all help build build-linux build-darwin build-windows \
        test test-unit test-race test-vet test-lint test-coverage test-short \
        release release-linux release-linux-local release-darwin release-darwin-local \
        release-windows release-windows-local release-local \
        _bundle-linux-mesa \
        compress checksums \
        clean clean-all \
        deps deps-linux deps-windows deps-update mocks \
        fmt vet lint security \
        version info install run

# --- Default -----------------------------------------------------------------
all: deps test build

# ==============================================================================
# HELP
# ==============================================================================
help:
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║              WATER AI — UNIFIED BUILD SYSTEM               ║"
	@echo "╠══════════════════════════════════════════════════════════════╣"
	@echo "║  make build            Build for current platform          ║"
	@echo "║  make build-linux      Cross-compile linux/amd64           ║"
	@echo "║  make build-darwin     Cross-compile darwin/amd64          ║"
	@echo "║  make build-windows    Cross-compile windows/amd64         ║"
	@echo "║  make test             Run ALL tests (vet+lint+unit+race) ║"
	@echo "║  make test-unit        Unit tests only                     ║"
	@echo "║  make test-race        Tests with race detection           ║"
	@echo "║  make test-vet         Run go vet                          ║"
	@echo "║  make test-lint        Run linter (if available)           ║"
	@echo "║  make test-coverage    Generate coverage report            ║"
	@echo "║  make test-short       Quick tests (no race)               ║"
	@echo "║  make release          Auto-detect OS & build artifacts    ║"
	@echo "║  make release-linux    Build linux .run installers         ║"
	@echo "║  make release-darwin   Build macOS .dmg disk images        ║"
	@echo "║  make release-windows  Build windows .exe (icon embedded)  ║"
	@echo "║  make compress         UPX-compress dist/ binaries         ║"
	@echo "║  make clean            Remove build artifacts              ║"
	@echo "║  make deps             Download Go dependencies            ║"
	@echo "║  make deps-linux       Install Linux system dependencies  ║"
	@echo "║  make deps-update      Update all dependencies             ║"
	@echo "║  make lint             Run golangci-lint                   ║"
	@echo "║  make version          Print version info                  ║"
	@echo "║  make info             Full build environment info         ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"


# ==============================================================================
# VERSION / INFO
# ==============================================================================
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"

info: version
	@echo "Host OS:    $(GOOS_HOST)"
	@echo "Host Arch:  $(GOARCH_HOST)"
	@echo "CGO:        $(CGO_ENABLED)"

# ==============================================================================
# SYSTEM DEPENDENCIES
# ==============================================================================
deps-linux:
	@echo "--> Installing Linux system dependencies for Fyne..."
	@sudo apt-get update -qq
	@sudo apt-get install -y -qq gcc g++ make pkg-config \
		libgl1-mesa-dev libegl1-mesa-dev libgles2-mesa-dev \
		libx11-dev libxcursor-dev libxrandr-dev \
		libxinerama-dev libxi-dev libxxf86vm-dev \
		libasound2-dev \
		gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
		makeself
	@echo "--> Installing fyne CLI..."
	@go install fyne.io/tools/cmd/fyne@latest
	@echo "--> Linux dependencies installed"

deps-linux-static:
	@echo "--> Installing Linux static-linking dependencies for Fyne..."
	@sudo apt-get update -qq
	@sudo apt-get install -y -qq gcc g++ make pkg-config \
		libgl1-mesa-dev libegl1-mesa-dev libgles2-mesa-dev \
		libx11-dev libxcursor-dev libxrandr-dev \
		libxinerama-dev libxi-dev libxxf86vm-dev \
		libasound2-dev \
		libgl-dev libx11-xcb-dev libxkbcommon-dev \
		libwayland-dev libvulkan-dev \
		gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
		libgl1-mesa-dri mesa-utils libosmesa6 \
		makeself
	@echo "--> Installing fyne CLI..."
	@go install fyne.io/tools/cmd/fyne@latest
	@echo "--> Linux static dependencies installed"

deps-windows:
	@echo "--> Installing Windows (MSYS2) dependencies for Fyne..."
	pacman --noconfirm -S mingw-w64-x86_64-toolchain mingw-w64-x86_64-pkg-config
	@echo "--> Windows dependencies installed"
deps-darwin:
	@echo "--> macOS dependencies (Xcode CLT assumed)..."
	@echo "--> Installing fyne CLI..."
	@go install fyne.io/tools/cmd/fyne@latest
	@echo "--> macOS dependencies ready"

deps:
	@echo "--> Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "--> Dependencies ready"

deps-update:
	@echo "--> Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "--> Dependencies updated"

mocks:
	@mkdir -p browser/embed
	@touch browser/embed/OpenSans-Medium.ttf
	@touch browser/embed/findVisibleInteractiveElements.js

# ==============================================================================
# CODE QUALITY
# ==============================================================================
fmt:
	@go fmt ./...

vet:
	@go vet ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout 5m ./...; \
	else \
		echo "WARN: golangci-lint not installed — skipping"; \
	fi

security:
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "WARN: govulncheck not installed — skipping"; \
	fi

# ==============================================================================
# TESTING
# ==============================================================================

# Top-level test target — runs ALL test sub-targets
test: test-vet test-lint test-unit test-race
	@echo "--> All tests passed"

test-unit: mocks
	@echo "--> Running unit tests..."
	@CGO_ENABLED=1 go test -v -count=1 -timeout $(TEST_TIMEOUT) ./...
	@echo "--> Unit tests passed"

test-race: mocks
	@echo "--> Running tests with race detection..."
	@CGO_ENABLED=1 go test -race -v -count=1 -timeout $(TEST_TIMEOUT) ./...
	@echo "--> Race detection tests passed"

test-vet:
	@echo "--> Running go vet..."
	@go vet ./...
	@echo "--> go vet passed"

test-lint:
	@echo "--> Running lint checks..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout 5m ./...; \
	elif command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "WARN: golangci-lint/staticcheck not installed — skipping lint"; \
	fi
	@echo "--> Lint checks passed"

test-coverage: mocks
	@echo "--> Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@PKGS=$$(CGO_ENABLED=1 go list ./... | xargs -I{} sh -c 'go list -f "{{if .TestGoFiles}}{{.ImportPath}}{{end}}" {} 2>/dev/null' | grep -v '^$$'); \
	if [ -n "$$PKGS" ]; then \
		CGO_ENABLED=1 go test -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic -count=1 -timeout $(TEST_TIMEOUT) $$PKGS || true; \
		if [ -f $(COVERAGE_DIR)/coverage.out ]; then \
			go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -n 1; \
		fi; \
	else \
		echo "WARN: no packages with test files found"; \
	fi
	@echo "--> Coverage report: $(COVERAGE_DIR)/coverage.out"

test-short: mocks
	@echo "--> Running short tests..."
	@go test -short -v -count=1 -timeout 5m ./...
	@echo "--> Short tests passed"

# ==============================================================================
# BUILD — single unified binary (gateway + GUI)
# ==============================================================================
build:
	@echo "--> Building $(BINARY) for $(GOOS_HOST)/$(GOARCH_HOST)..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=1 go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD_PKG)
	@echo "--> $(BIN_DIR)/$(BINARY)"

build-linux:
	@echo "--> Building $(BINARY) for linux/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build $(GO_BUILD_FLAGS) \
		-o $(BIN_DIR)/$(BINARY)-linux-amd64 $(CMD_PKG)
	@echo "--> $(BIN_DIR)/$(BINARY)-linux-amd64"

build-darwin:
	@echo "--> Building $(BINARY) for darwin/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build $(GO_BUILD_FLAGS) \
		-o $(BIN_DIR)/$(BINARY)-darwin-amd64 $(CMD_PKG)
	@echo "--> $(BIN_DIR)/$(BINARY)-darwin-amd64"

build-windows:
	@echo "--> Building $(BINARY) for windows/amd64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY)-windows-amd64.exe $(CMD_PKG)
	@echo "--> $(BIN_DIR)/$(BINARY)-windows-amd64.exe"

# Build Go backend only (no Node.js/frontend required)
build-dev: mocks
	@echo "--> Building Go backend (dev mode, no frontend)..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 go build $(GO_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/water
	@echo "Done! Run with: ./$(BIN_DIR)/$(BINARY_NAME)"

# ==============================================================================
# RELEASE — auto-detect OS and call the appropriate platform target
# ==============================================================================
# Each platform target installs its own dependencies, then builds both amd64
# and arm64 binaries natively (no cross-compilation across OS boundaries).
# CI matrix runners each call `make release` which auto-dispatches.
# ==============================================================================

release: clean
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║           BUILDING RELEASE BINARIES                        ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"
ifeq ($(GOOS_HOST),linux)
	@$(MAKE) release-linux
else ifeq ($(GOOS_HOST),darwin)
	@$(MAKE) release-darwin
else
	@$(MAKE) release-windows
endif
	@$(MAKE) checksums
	@echo "--> Release builds in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

# ------------------------------------------------------------------------------
# release-linux — .run self-extracting installers with bundled Mesa fallback
# ------------------------------------------------------------------------------
release-linux: deps-linux-static
	@echo "--> Building Linux release .run self-extracting installers..."
	@echo "    Validating all required assets..."
	@for asset in $(ASSET_FILES); do \
		test -f "$(ASSET_DIR)/$$asset" || { echo "ERROR: Asset file not found: $(ASSET_DIR)/$$asset"; exit 1; }; \
		echo "      ✓ $$asset"; \
	done
	@test -f scripts/water-launcher.sh || { echo "ERROR: Launcher script not found: scripts/water-launcher.sh"; exit 1; }
	@test -f scripts/linux-install.sh || { echo "ERROR: Install script not found: scripts/linux-install.sh"; exit 1; }
	@mkdir -p $(DIST_DIR)
	@echo "    Building amd64 binary..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		go build -a $(GO_BUILD_FLAGS_LINUX) -o $(DIST_DIR)/$(BINARY)-linux-amd64-bin $(CMD_PKG)
	@echo "    Building arm64 binary..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc \
		go build -a $(GO_BUILD_FLAGS_LINUX) -o $(DIST_DIR)/$(BINARY)-linux-arm64-bin $(CMD_PKG)
	@echo "--> Packaging Linux amd64 with Mesa software renderer fallback..."
	@$(MAKE) _bundle-linux-mesa ARCH=amd64
	@echo "--> Packaging Linux arm64 with Mesa software renderer fallback..."
	@$(MAKE) _bundle-linux-mesa ARCH=arm64
	@echo "--> Linux release .run installers built (with Mesa software renderer fallback)"

# Internal target: bundle a Linux binary with Mesa software renderer libs,
# icon, desktop entry, and install script; create a .run self-extracting
# installer via makeself that runs the install script on extraction.
_bundle-linux-mesa:
	$(eval BUNDLE_DIR := $(DIST_DIR)/$(BINARY)-linux-$(ARCH))
	$(eval MULTIARCH := $(shell dpkg-architecture -qDEB_HOST_MULTIARCH 2>/dev/null))
	@mkdir -p $(BUNDLE_DIR)/bin $(BUNDLE_DIR)/lib/dri $(BUNDLE_DIR)/assets
	@# ---- Validate ALL asset files exist before proceeding ----
	@echo "    Validating asset files..."
	@for asset in $(ASSET_FILES); do \
		test -f "$(ASSET_DIR)/$$asset" || { echo "ERROR: Asset file not found: $(ASSET_DIR)/$$asset"; exit 1; }; \
		echo "      ✓ $$asset"; \
	done
	@test -f scripts/water-launcher.sh || { echo "ERROR: Launcher script not found: scripts/water-launcher.sh"; exit 1; }
	@test -f scripts/linux-install.sh || { echo "ERROR: Install script not found: scripts/linux-install.sh"; exit 1; }
	@# ---- Copy the binary into the bundle ----
	@mv $(DIST_DIR)/$(BINARY)-linux-$(ARCH)-bin $(BUNDLE_DIR)/bin/$(BINARY)
	@# Copy the icon into the bundle
	@cp $(APP_ICON) $(BUNDLE_DIR)/icon.png
	@# Copy the desktop entry template into the bundle
	@cp scripts/water.desktop $(BUNDLE_DIR)/water.desktop
	@# Copy the install script into the bundle
	@cp scripts/water-install.sh $(BUNDLE_DIR)/install.sh
	@chmod +x $(BUNDLE_DIR)/install.sh
	@# Bundle Mesa software rendering libraries
	@echo "    Copying Mesa software renderer libraries (multiarch=$(MULTIARCH))..."
	@CANDIDATE_PATHS="/usr/lib /usr/lib64"; \
	if [ -n "$(MULTIARCH)" ]; then \
		CANDIDATE_PATHS="$$CANDIDATE_PATHS /usr/lib/$(MULTIARCH)"; \
	fi; \
	CANDIDATE_PATHS="$$CANDIDATE_PATHS /usr/lib/x86_64-linux-gnu /usr/lib/aarch64-linux-gnu"; \
	SEARCH_PATHS=""; \
	for p in $$CANDIDATE_PATHS; do \
		if [ -d "$$p" ]; then SEARCH_PATHS="$$SEARCH_PATHS $$p"; fi; \
	done; \
	if [ -z "$$SEARCH_PATHS" ]; then \
		echo "ERROR: no library search paths found"; exit 1; \
	fi; \
	for lib in libGL.so.1 libGLX.so.0 libGLdispatch.so.0 libEGL.so.1 libGLESv2.so.2; do \
		SRC=""; \
		SRC=$$(find $$SEARCH_PATHS \
			-name "$$lib*" \( -type f -o -type l \) 2>/dev/null | head -1); \
		if [ -n "$$SRC" ]; then \
			if [ -L "$$SRC" ]; then \
				REAL=$$(readlink -f "$$SRC"); \
				cp "$$REAL" "$(BUNDLE_DIR)/lib/$$(basename $$REAL)"; \
				(cd "$(BUNDLE_DIR)/lib" && ln -sf "$$(basename $$REAL)" "$$lib"); \
			else \
				cp "$$SRC" "$(BUNDLE_DIR)/lib/$$lib"; \
			fi; \
			echo "      Bundled $$lib"; \
		else \
			echo "      WARN: $$lib not found on system (searched: $$SEARCH_PATHS)"; \
		fi; \
	done
	@# ---- Bundle the swrast/llvmpipe DRI driver ----
	@CANDIDATE_PATHS="/usr/lib /usr/lib64 /usr/lib/dri"; \
	if [ -n "$(MULTIARCH)" ]; then \
		CANDIDATE_PATHS="$$CANDIDATE_PATHS /usr/lib/$(MULTIARCH) /usr/lib/$(MULTIARCH)/dri"; \
	fi; \
	CANDIDATE_PATHS="$$CANDIDATE_PATHS /usr/lib/x86_64-linux-gnu /usr/lib/aarch64-linux-gnu"; \
	SEARCH_PATHS=""; \
	for p in $$CANDIDATE_PATHS; do \
		if [ -d "$$p" ]; then SEARCH_PATHS="$$SEARCH_PATHS $$p"; fi; \
	done; \
	for drv in swrast_dri.so kms_swrast_dri.so; do \
		SRC=""; \
		if [ -n "$$SEARCH_PATHS" ]; then \
			SRC=$$(find $$SEARCH_PATHS \
				-name "$$drv" -type f 2>/dev/null | head -1); \
		fi; \
		if [ -n "$$SRC" ]; then \
			cp "$$SRC" "$(BUNDLE_DIR)/lib/dri/$$drv"; \
			echo "      Bundled $$drv"; \
		else \
			echo "      WARN: $$drv not found on system (searched: $$SEARCH_PATHS)"; \
		fi; \
	done
	@# ---- Verify bundle contents before creating .run ----
	@echo "    Verifying bundle contents..."
	@test -x $(BUNDLE_DIR)/bin/$(BINARY) || { echo "ERROR: Binary not executable in bundle"; exit 1; }
	@test -x $(BUNDLE_DIR)/$(BINARY) || { echo "ERROR: Launcher not executable in bundle"; exit 1; }
	@test -x $(BUNDLE_DIR)/install.sh || { echo "ERROR: Install script not executable in bundle"; exit 1; }
	@for asset in $(ASSET_FILES); do \
		test -f "$(BUNDLE_DIR)/assets/$$asset" || { echo "ERROR: Asset missing from bundle: $$asset"; exit 1; }; \
	done
	@echo "    Bundle contents verified."
	@# ---- Create the .run self-extracting installer via makeself ----
	@echo "    Creating self-extracting installer $(BINARY)-linux-$(ARCH).run..."
	@command -v makeself >/dev/null 2>&1 || { echo "ERROR: makeself is not installed"; exit 1; }
	@makeself --nox11 $(BUNDLE_DIR) $(DIST_DIR)/$(BINARY)-linux-$(ARCH).run \
		"$(BINARY) Installer" ./install.sh
	@rm -rf $(BUNDLE_DIR)
	@echo "    $(DIST_DIR)/$(BINARY)-linux-$(ARCH).run"

# ------------------------------------------------------------------------------
# release-darwin — .dmg disk images via fyne package + hdiutil
# ------------------------------------------------------------------------------
release-darwin: deps-darwin
	@echo "--> Building macOS release .dmg disk images..."
	@test -f $(APP_ICON) || { echo "ERROR: Icon file not found: $(APP_ICON)"; exit 1; }
	@mkdir -p $(DIST_DIR)
	@echo "    Building $(BINARY)-darwin-amd64.dmg ..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		fyne package -os darwin -name $(BINARY) --app-id $(APP_ID) \
		-icon $(APP_ICON) -src $(CMD_PKG) \
		-release
	@test -d $(BINARY).app || { echo "ERROR: fyne package failed to produce $(BINARY).app"; exit 1; }
	@hdiutil create -volname "$(BINARY)" -srcfolder $(BINARY).app \
		-ov -format UDZO $(DIST_DIR)/$(BINARY)-darwin-amd64.dmg
	@rm -rf $(BINARY).app
	@echo "    Building $(BINARY)-darwin-arm64.dmg ..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		fyne package -os darwin -name $(BINARY) --app-id $(APP_ID) \
		-icon $(APP_ICON) -src $(CMD_PKG) \
		-release
	@test -d $(BINARY).app || { echo "ERROR: fyne package failed to produce $(BINARY).app"; exit 1; }
	@hdiutil create -volname "$(BINARY)" -srcfolder $(BINARY).app \
		-ov -format UDZO $(DIST_DIR)/$(BINARY)-darwin-arm64.dmg
	@rm -rf $(BINARY).app
	@echo "--> macOS release .dmg disk images built"

# ------------------------------------------------------------------------------
# release-windows — .exe with embedded manifest/icon via fyne package
# ------------------------------------------------------------------------------
# _find-windows-exe: helper to locate the .exe produced by fyne package.
# fyne package may name the output based on -name, FyneApp.toml Name, or the
# source directory (e.g. "water.exe" vs "Water.exe").  We search for all
# plausible names so the build succeeds regardless of fyne's behaviour.
define _find-windows-exe
	@FOUND=""; \
	for candidate in "$(BINARY).exe" "$$(echo $(BINARY) | tr '[:upper:]' '[:lower:]').exe" \
	                 "$$(basename $(CMD_PKG)).exe"; do \
		if [ -f "$$candidate" ]; then \
			FOUND="$$candidate"; \
			break; \
		fi; \
	done; \
	if [ -z "$$FOUND" ]; then \
		FOUND=$$(find . -maxdepth 1 -iname '*.exe' -newer Makefile -print -quit 2>/dev/null); \
	fi; \
	if [ -z "$$FOUND" ]; then \
		echo "ERROR: fyne package did not produce an .exe (looked for $(BINARY).exe and variants)"; \
		echo "       Files in current directory:"; ls -la; \
		exit 1; \
	fi; \
	echo "    Found fyne output: $$FOUND"; \
	mv "$$FOUND" "$(1)"
endef

release-windows: deps-windows
	@echo "--> Building Windows release binaries (fyne package)..."
	@test -f $(APP_ICON) || { echo "ERROR: Icon file not found: $(APP_ICON)"; exit 1; }
	@mkdir -p $(DIST_DIR)
	@echo "    Building $(BINARY)-windows-amd64.exe ..."
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
		fyne package -os windows -name $(BINARY) --app-id $(APP_ID) \
		-icon $(APP_ICON) -src $(CMD_PKG) \
		-release
	@if [ -f "$(BINARY).exe" ]; then mv "$(BINARY).exe" "$(DIST_DIR)/$(BINARY)-windows-amd64.exe"; \
	elif [ -f "cmd/water/$(BINARY).exe" ]; then mv "cmd/water/$(BINARY).exe" "$(DIST_DIR)/$(BINARY)-windows-amd64.exe"; \
	else echo "ERROR: $(BINARY).exe not found after fyne package (amd64)"; exit 1; fi
	@echo "    Building $(BINARY)-windows-arm64.exe ..."
	@echo "    (cross-compiling arm64 with CC=$${CC:-aarch64-w64-mingw32-gcc})"
	@CGO_ENABLED=1 GOOS=windows GOARCH=arm64 \
		CC=$${CC:-aarch64-w64-mingw32-gcc} \
		CGO_LDFLAGS="-static -lpthread -lm" \
		fyne package -os windows -name $(BINARY) --app-id $(APP_ID) \
		-icon $(APP_ICON) -src $(CMD_PKG) \
		-release
	@if [ -f "$(BINARY).exe" ]; then mv "$(BINARY).exe" "$(DIST_DIR)/$(BINARY)-windows-arm64.exe"; \
	elif [ -f "cmd/water/$(BINARY).exe" ]; then mv "cmd/water/$(BINARY).exe" "$(DIST_DIR)/$(BINARY)-windows-arm64.exe"; \
	else echo "ERROR: $(BINARY).exe not found after fyne package (arm64)"; exit 1; fi
	@echo "--> Windows release binaries built"

# ------------------------------------------------------------------------------
# release-windows-local — build a single .exe for the current arch only (CI use)
# ------------------------------------------------------------------------------
# For arm64 cross-compilation, set CC=aarch64-w64-mingw32-gcc and
# CGO_LDFLAGS="-static -lpthread -lm" in the environment before calling this.
release-windows-local:
	@echo "--> Building Windows .exe for $(GOARCH_HOST) with embedded icon..."
	@test -f $(APP_ICON) || { echo "ERROR: Icon file not found: $(APP_ICON)"; exit 1; }
	@mkdir -p $(DIST_DIR)
ifeq ($(GOARCH_HOST),arm64)
	@echo "    (cross-compiling for arm64 with CC=$${CC:-aarch64-w64-mingw32-gcc})"
	@CGO_ENABLED=1 GOOS=windows GOARCH=arm64 \
		CC=$${CC:-aarch64-w64-mingw32-gcc} \
		fyne package -os windows -icon $(APP_ICON) --app-id $(APP_ID) \
		-name $(BINARY) -src $(CMD_PKG) -release || { \
			echo "WARN: fyne package failed, falling back to go build"; \
			CGO_ENABLED=1 GOOS=windows GOARCH=arm64 \
				CC=$${CC:-aarch64-w64-mingw32-gcc} \
				go build $(GO_BUILD_FLAGS_WINDOWS_STATIC) \
				-o $(DIST_DIR)/$(BINARY)-windows-arm64.exe $(CMD_PKG); \
		}
else
	@CGO_ENABLED=1 GOOS=windows GOARCH=$(GOARCH_HOST) \
		fyne package -os windows -icon $(APP_ICON) --app-id $(APP_ID) \
		-name $(BINARY) -src $(CMD_PKG) -release || { \
			echo "WARN: fyne package failed, falling back to go build"; \
			CGO_ENABLED=1 GOOS=windows GOARCH=$(GOARCH_HOST) \
				go build $(GO_BUILD_FLAGS_WINDOWS_STATIC) \
				-o $(DIST_DIR)/$(BINARY)-windows-$(GOARCH_HOST).exe $(CMD_PKG); \
		}
endif
	@if [ -f "$(BINARY).exe" ]; then \
		mv "$(BINARY).exe" "$(DIST_DIR)/$(BINARY)-windows-$(GOARCH_HOST).exe"; \
	elif [ -f "cmd/water/$(BINARY).exe" ]; then \
		mv "cmd/water/$(BINARY).exe" "$(DIST_DIR)/$(BINARY)-windows-$(GOARCH_HOST).exe"; \
	elif [ ! -f "$(DIST_DIR)/$(BINARY)-windows-$(GOARCH_HOST).exe" ]; then \
		echo "ERROR: Could not find $(BINARY).exe after build"; \
		exit 1; \
	fi
	@echo "--> $(DIST_DIR)/$(BINARY)-windows-$(GOARCH_HOST).exe"

# ------------------------------------------------------------------------------
# release-local — build a single release artifact for the current platform
# ------------------------------------------------------------------------------
# Dispatches to the platform-specific -local target which produces the proper
# artifact format (.dmg, .run, .exe with icon).
release-local:
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║       BUILDING RELEASE ARTIFACT (native platform)          ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"
	@mkdir -p $(DIST_DIR)
	$(eval _OS   := $(or $(GOOS),$(GOOS_HOST)))
	$(eval _ARCH := $(or $(GOARCH),$(GOARCH_HOST)))
	$(eval _EXT  := $(if $(filter windows,$(_OS)),.exe,))
	$(eval _OUT  := $(DIST_DIR)/$(BINARY)-$(_OS)-$(_ARCH)$(_EXT))
	@echo "    Building $(_OUT) ..."
ifeq ($(GOOS_HOST),linux)
	@echo "    (using mostly-static linking for Linux → .run installer)"
	@CGO_ENABLED=1 GOOS=$(_OS) GOARCH=$(_ARCH) \
		go build -a $(GO_BUILD_FLAGS_LINUX) -o $(DIST_DIR)/$(BINARY)-$(_OS)-$(_ARCH)-bin $(CMD_PKG)
	@$(MAKE) _bundle-linux-mesa ARCH=$(_ARCH)
else ifeq ($(GOOS_HOST),darwin)
	@echo "    (using fyne package for macOS .app bundle → .dmg)"
	@test -f $(APP_ICON) || { echo "ERROR: Icon file not found: $(APP_ICON)"; exit 1; }
	@CGO_ENABLED=1 GOOS=$(_OS) GOARCH=$(_ARCH) \
		fyne package -os darwin -name $(BINARY) --app-id $(APP_ID) \
		-icon $(APP_ICON) -src $(CMD_PKG) -release
	@test -d $(BINARY).app || { echo "ERROR: fyne package failed to produce $(BINARY).app"; exit 1; }
	@hdiutil create -volname "$(BINARY)" -srcfolder $(BINARY).app \
		-ov -format UDZO $(DIST_DIR)/$(BINARY)-$(_OS)-$(_ARCH).dmg
	@rm -rf $(BINARY).app
else
	@echo "    (using fyne package for Windows .exe)"
	@echo "    CC=$${CC:-default} GOARCH=$(_ARCH)"
	@test -f $(APP_ICON) || { echo "ERROR: Icon file not found: $(APP_ICON)"; exit 1; }
	@CGO_ENABLED=1 GOOS=$(_OS) GOARCH=$(_ARCH) \
		fyne package -os windows -name $(BINARY) --app-id $(APP_ID) \
		-icon $(APP_ICON) -src $(CMD_PKG) -release || { \
			echo "WARN: fyne package failed, falling back to go build"; \
			CGO_ENABLED=1 GOOS=$(_OS) GOARCH=$(_ARCH) \
				go build $(GO_BUILD_FLAGS_WINDOWS_STATIC) -o $(_OUT) $(CMD_PKG); \
		}
	@echo "    Searching for produced .exe files..."
	@if [ -f "$(BINARY).exe" ]; then \
		mv "$(BINARY).exe" "$(_OUT)"; \
	elif [ -f "cmd/water/$(BINARY).exe" ]; then \
		mv "cmd/water/$(BINARY).exe" "$(_OUT)"; \
	elif [ ! -f "$(_OUT)" ]; then \
		echo "ERROR: Could not find $(BINARY).exe after build"; \
		echo "    Files in current directory:"; ls -la *.exe 2>/dev/null || echo "      (no .exe files)"; \
		echo "    Files in dist/:"; ls -la $(DIST_DIR)/ 2>/dev/null || echo "      (empty)"; \
		exit 1; \
	fi
endif
	@echo "--> $(_OUT)"

checksums:
	@if [ -d "$(DIST_DIR)" ] && [ "$$(ls -A $(DIST_DIR) 2>/dev/null)" ]; then \
		cd $(DIST_DIR) && \
		if command -v sha256sum >/dev/null 2>&1; then \
			sha256sum * > checksums.sha256; \
		elif command -v shasum >/dev/null 2>&1; then \
			shasum -a 256 * > checksums.sha256; \
		else \
			echo "WARN: no sha256sum or shasum found — skipping checksums"; \
		fi; \
		echo "--> $(DIST_DIR)/checksums.sha256"; \
	fi

# ==============================================================================
# COMPRESS (optional — requires UPX)
# ==============================================================================
compress:
	@if command -v upx >/dev/null 2>&1; then \
		echo "--> Compressing binaries with UPX..."; \
		for f in $(DIST_DIR)/$(BINARY)-*; do \
			upx --best --lzma "$$f" 2>/dev/null || true; \
		done; \
		echo "--> Compression complete"; \
	else \
		echo "WARN: UPX not found — skipping compression"; \
	fi

# ==============================================================================
# INSTALL / RUN
# ==============================================================================
install: build
	@echo "--> Installing $(BINARY) to $(GOPATH)/bin/"
	@mkdir -p $(GOPATH)/bin
	@cp $(BIN_DIR)/$(BINARY) $(GOPATH)/bin/
	@echo "--> Installed"

run: build
	@./$(BIN_DIR)/$(BINARY)

# ==============================================================================
# CLEAN
# ==============================================================================
clean:
	@echo "--> Cleaning build artifacts..."
	@rm -rf $(DIST_DIR) $(BIN_DIR) $(COVERAGE_DIR)
	@rm -f coverage.out coverage.html
	@echo "--> Clean"

clean-all: clean
	@go clean -cache -testcache -modcache -i -r 2>/dev/null || true
	@echo "--> Deep clean complete"
