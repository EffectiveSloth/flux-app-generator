.PHONY: build build-platform build-all test test-coverage-ci test-coverage-html lint clean help

BINARY_NAME=flux-app-generator
CMD_PATH=cmd/flux-app-generator
DIST_DIR=dist
COVERAGE_FILE=coverage.txt

# Default build for current platform
build: clean
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY_NAME) ./$(CMD_PATH)

# Build for specific platform (used by CI matrix)
build-platform:
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		echo "Error: GOOS and GOARCH must be set"; \
		echo "Usage: GOOS=linux GOARCH=amd64 make build-platform"; \
		exit 1; \
	fi
	mkdir -p $(DIST_DIR)
	@if [ "$(GOOS)" = "windows" ]; then \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(DIST_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH).exe ./$(CMD_PATH); \
	else \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(DIST_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./$(CMD_PATH); \
	fi

# Build for all supported platforms
build-all: clean
	mkdir -p $(DIST_DIR)
	@echo "Building for all platforms..."
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "Building for $$os/$$arch..."; \
			if [ "$$os" = "windows" ]; then \
				GOOS=$$os GOARCH=$$arch go build -o $(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch.exe ./$(CMD_PATH); \
			else \
				GOOS=$$os GOARCH=$$arch go build -o $(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch ./$(CMD_PATH); \
			fi; \
		done; \
	done
	@echo "Build complete! Binaries in $(DIST_DIR)/"

run: build
	./$(DIST_DIR)/$(BINARY_NAME)

test:
	go test ./...

test-coverage-ci:
	go test -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...

test-coverage-html: test-coverage-ci
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	golangci-lint run ./...

clean:
	rm -rf $(DIST_DIR)
	rm -f $(COVERAGE_FILE) coverage.html

help:
	@echo "Available targets:"
	@echo "  build              - Build binary for current platform in dist/"
	@echo "  build-platform     - Build for specific platform (requires GOOS and GOARCH)"
	@echo "  build-all          - Build for all supported platforms"
	@echo "  run                - Build and run the CLI from dist/"
	@echo "  test               - Run tests (fast, no coverage)"
	@echo "  test-coverage-ci   - Run tests with coverage (same as CI)"
	@echo "  test-coverage-html - Run tests and generate HTML coverage report"
	@echo "  lint               - Lint the codebase (requires golangci-lint)"
	@echo "  clean              - Remove build artifacts and coverage files"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  GOOS=linux GOARCH=amd64 make build-platform"
	@echo "  GOOS=windows GOARCH=arm64 make build-platform" 