.PHONY: build build-demo build-all test test-race create-package prepare-lib bundle-lib clean package

# Detect operating system for library paths
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Set library path based on OS
ifeq ($(UNAME_S),Darwin)
    LIB_PATH_VAR := DYLD_LIBRARY_PATH
    LIB_PATH := $(PWD)/build/libs/darwin
else ifeq ($(UNAME_S),Linux)
    LIB_PATH_VAR := LD_LIBRARY_PATH
    LIB_PATH := $(PWD)/build/libs/linux/$(UNAME_M)
else
    # Windows - requires different approach
    LIB_PATH_VAR := PATH
    LIB_PATH := $(PWD)/build/libs/windows/x64
endif

# Production build
build:
	go build -o build/gotune ./cmd

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

create-package:
	fyne package -name GoTune -icon Icon.png appVersion 0.0.1

bundle-lib:
	cp -r ./libs GoTune.app/Contents/libs

clean:
	rm -f build/gotune build/gotune-demo build/gotune-*

package: create-package bundle-lib clean


# Linting
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
# Excludes core packages: api/base, internal/core, internal/observability, internal/repo, internal/transport
deadcode:
	@which deadcode > /dev/null || (echo "Installing deadcode..." && go install golang.org/x/tools/cmd/deadcode@latest)
	deadcode ./... #| grep -vE '(api/base|internal/core|internal/observability|internal/repo|internal/transport)' || true

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
# Simulate CI workflow locally (runs all CI checks with filtering)
ci-local: lint deadcode build
	@echo "✅ Local CI simulation complete"