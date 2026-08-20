.PHONY: all build test test-frontend test-parity test-sanitizers lint clean release help

BINARY_NAME=scriptgo
BUILD_DIR=bin

all: build test

## build: Build the scriptgo CLI binary
build:
	@echo "==> Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./cmd/scriptgo

## test: Run full unit and integration test suite
test:
	@echo "==> Running full test suite..."
	go test -v -count=1 ./...

## test-frontend: Run TypeScript-Go frontend adapter tests
test-frontend:
	@echo "==> Running TypeScript-Go frontend tests..."
	go test -v -count=1 ./internal/typescriptgo/...

## test-parity: Run Node.js parity comparison benchmark across the corpus test suite
test-parity:
	@echo "==> Running Node.js parity checker..."
	go run ./cmd/parity

## test-sanitizers: Run native builds with AddressSanitizer & memory checks across the corpus in parallel
test-sanitizers:
	@echo "==> Running AddressSanitizer checks across test corpus in parallel..."
	SCRIPTGO_SANITIZE="address,undefined" go test -v -count=1 -parallel 8 ./internal/compiler -run TestCorpus

## lint: Run Go vet checks
lint:
	@echo "==> Running go vet..."
	go vet ./...
	go vet ./internal/typescriptgo/...

## release: Prepare a new release locally (usage: make release VERSION=0.1.0)
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release VERSION=0.1.0"; \
		exit 1; \
	fi
	@./scripts/prepare-release.sh $(VERSION)

## clean: Clean up build artifacts and temporary files
clean:
	@echo "==> Cleaning up build artifacts..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -B1 -E '^[a-zA-Z_-]+:' $(MAKEFILE_LIST) | awk '/^##/{desc=$$0; sub(/^## /, "", desc)} /^[a-zA-Z_-]+:/{sub(/:.*/, ""); if (desc != "") {printf "  \033[36m%-18s\033[0m %s\n", $$0, desc; desc=""}}'
