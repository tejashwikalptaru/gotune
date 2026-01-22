.PHONY: build build-demo build-all test test-race create-package prepare-lib bundle-lib fix-rpath clean package execute run

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

# Production build
build:
	go build -o build/gotune ./

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
run: build execute

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
package: create-package fix-rpath bundle-lib
	@echo "Package created successfully: Go Tune.app"

# -------------------------------------------------------------------------
# Linting
# -------------------------------------------------------------------------
.PHONY: lint lint-ci lint-fix lint-install lint-all deadcode deadcode-unfiltered deadcode-test

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
	deadcode ./...

# Check for dead code including test executables
deadcode-test:
	@which deadcode > /dev/null || (echo "Installing deadcode..." && go install golang.org/x/tools/cmd/deadcode@latest)
	deadcode -test ./...

# Check for dead code without filtering (for debugging/special cases)
deadcode-unfiltered:
	@which deadcode > /dev/null || (echo "Installing deadcode..." && go install golang.org/x/tools/cmd/deadcode@latest)
	deadcode ./...

# Run all linting checks (golangci-lint + deadcode)
lint-all: lint deadcode

# CI Simulation
ci-local: lint deadcode build
	@echo "✅ Local CI simulation complete"