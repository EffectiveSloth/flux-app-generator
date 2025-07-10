.PHONY: build run test lint clean help

BINARY_NAME=flux-app-generator
CMD_PATH=cmd/flux-app-generator
DIST_DIR=dist

build:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY_NAME) ./$(CMD_PATH)

run: build
	./$(DIST_DIR)/$(BINARY_NAME)

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(DIST_DIR)

help:
	@echo "Available targets:"
	@echo "  build   - Build the CLI binary in dist/"
	@echo "  run     - Build and run the CLI from dist/"
	@echo "  test    - Run tests"
	@echo "  lint    - Lint the codebase (requires golangci-lint)"
	@echo "  clean   - Remove build artifacts in dist/"
	@echo "  help    - Show this help message" 