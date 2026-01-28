.PHONY: generate-credits build build-demo build-all test test-race create-package prepare-lib bundle-lib fix-rpath clean package execute run

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
MODULE := github.com/tejashwikalptaru/gotune

LDFLAGS := -ldflags "\
	-X '$(MODULE)/internal/app.Version=$(VERSION)' \
	-X '$(MODULE)/internal/app.GitCommit=$(GIT_COMMIT)' \
	-X '$(MODULE)/internal/app.GitTag=$(GIT_TAG)' \
	-X '$(MODULE)/internal/app.BuildTime=$(BUILD_TIME)'"

# Detect operating system for library paths
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Set library path based on OS
ifeq ($(UNAME_S),Darwin)
    LIB_PATH_VAR := DYLD_LIBRARY_PATH
    LIB_PATH := $(PWD)/build/libs/darwin
    LIB_FILE := libbass.dylib
else ifeq ($(UNAME_S),Linux)
    LIB_PATH_VAR := LD_LIBRARY_PATH
    LIB_PATH := $(PWD)/build/libs/linux/$(UNAME_M)
    LIB_FILE := libbass.so
else
    # Windows - requires different approach
    LIB_PATH_VAR := PATH
    LIB_PATH := $(PWD)/build/libs/windows/x64
    LIB_FILE := bass.dll
endif

# Generate credit
generate-credits:
	@which fyne-credits-generator > /dev/null || (echo "Installing fyne-credits-generator..." && go install github.com/tejashwikalptaru/fyne-credits-generator/cmd/fyne-credits-generator@latest)
	@echo "Generating credits..."
	fyne-credits-generator > internal/adapter/ui/credits.go
	@echo "Done"

# Production build
build:
	go build $(LDFLAGS) -o build/gotune ./

# Run tests
test:
	$(LIB_PATH_VAR)=$(LIB_PATH) go test ./internal/... -v -count=1

# Run tests with race detection
test-race:
	$(LIB_PATH_VAR)=$(LIB_PATH) go test ./internal/... -race -count=1

# Execute production binary
execute:
	./build/gotune

# Build and run production
run: generate-credits build execute

# Package
install-fyne:
	go install fyne.io/fyne/v2/cmd/fyne@latest

create-package:
	fyne package

# Updated: Copies lib to Contents/MacOS so @executable_path finds it
bundle-lib:
ifeq ($(UNAME_S),Darwin)
	@echo "Bundling $(LIB_FILE) into App Bundle..."
	cp "$(LIB_PATH)/$(LIB_FILE)" "Go Tune.app/Contents/MacOS/"
	# Optional: Remove quarantine if downloaded from web (prevents security warning)
	# xattr -d com.apple.quarantine "Go Tune.app/Contents/MacOS/$(LIB_FILE)" || true
else
	@echo "Bundle logic for $(UNAME_S) not implemented yet"
endif

clean:
	rm -f build/gotune build/gotune-demo build/gotune-*

# Add @executable_path to rpath so the app can find bundled dylibs
fix-rpath:
ifeq ($(UNAME_S),Darwin)
	@echo "Adding @executable_path rpath to executable..."
	install_name_tool -add_rpath @executable_path "Go Tune.app/Contents/MacOS/gotune"
endif

# Full package flow
package: generate-credits create-package fix-rpath bundle-lib
	@echo "Package created successfully: Go Tune.app"

# -------------------------------------------------------------------------
# Linting
# -------------------------------------------------------------------------
.PHONY: lint lint-ci lint-fix lint-install lint-all deadcode

lint:
	golangci-lint run

# Install golangci-lint locally (optional helper)
lint-install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.6.2

# Attempt to auto-fix what can be fixed (formatting, some suggestions)
lint-fix:
	golangci-lint run --fix

# Check for unreachable (dead) code including exported functions not used outside their package
deadcode:
	@which deadcode > /dev/null || (echo "Installing deadcode..." && go install golang.org/x/tools/cmd/deadcode@latest)
	@echo "Running deadcode analysis..."
	@# Capture stdout and stderr (2>&1) so we see if the command fails or finds code
	@output=$$(deadcode -test ./... 2>&1); \
	if [ -n "$$output" ]; then \
		echo "$$output"; \
		echo "Error: Dead code detected"; \
		exit 1; \
	else \
		echo "Success: No dead code found"; \
	fi

# Run all linting checks (golangci-lint + deadcode)
lint-all: lint deadcode

# CI Simulation
ci-local: lint deadcode build
	@echo "Local CI checks complete"