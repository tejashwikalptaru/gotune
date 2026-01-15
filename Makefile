.PHONY: build create-package prepare-lib bundle-lib clean package

build:
	go build -o build/gotune cmd/main.go

execute:
	./build/gotune

run: build execute

create-package:
	fyne package -name GoTune -icon Icon.png appVersion 0.0.1

prepare-lib:
	install_name_tool -id "@loader_path/../libs/libbass.dylib" ./libs/libbass.dylib

bundle-lib:
	cp -r ./libs GoTune.app/Contents/libs

clean:
	rm ./gotune

package: prepare-lib create-package bundle-lib clean


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
ci-local: lint deadcode test-ci build
	@echo "✅ Local CI simulation complete"