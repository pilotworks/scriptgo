.PHONY: all build test test-frontend test-parity test-262 test-262-baseline test-sanitizers audit lint clean release help

BINARY_NAME=scriptgo
ALIAS_NAME=scg
BUILD_DIR=bin

all: build test

## build: Build the scriptgo CLI binary and scg alias
build:
	@echo "==> Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./cmd/scriptgo
	@ln -sf $(BINARY_NAME) $(ALIAS_NAME)

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

## test-262: Run the test262 conformance subset (TEST262_ROOT=<checkout>, TEST262_PATHS=<dirs>)
TEST262_PATHS ?= language/expressions,language/statements,built-ins/Math,built-ins/Array,built-ins/String,built-ins/JSON
test-262:
	@test -n "$(TEST262_ROOT)" || (echo "set TEST262_ROOT to a tc39/test262 checkout" && exit 2)
	go run ./cmd/test262 -root $(TEST262_ROOT) -paths $(TEST262_PATHS)

## test-262-baseline: Fail if a test in the recorded passing baseline regresses (CI gate)
TEST262_REV ?= 2e0a56762801e275a9fdf96dc49d90ba0cddcf63
TEST262_BASELINE ?= internal/test262/testdata/baseline.txt
test-262-baseline:
	@test -n "$(TEST262_ROOT)" || (echo "set TEST262_ROOT to a tc39/test262 checkout at $(TEST262_REV)" && exit 2)
	go run ./cmd/test262 -root $(TEST262_ROOT) -list $(TEST262_BASELINE) -require-pass -reasons 5

## audit: Run official Node.js API coverage audit against corpus tests (Source of Truth)
audit:
	@echo "==> Running Node.js official API coverage audit against corpus..."
	go run ./cmd/parity -audit

## test-sanitizers: Run native builds with AddressSanitizer & memory checks across the corpus in parallel
test-sanitizers:
	@echo "==> Running AddressSanitizer checks across test corpus in parallel..."
	ASAN_OPTIONS="detect_leaks=0" SCRIPTGO_SANITIZE="address,undefined" go test -v -count=1 -timeout 30m -parallel 2 ./internal/compiler -run TestCorpus

## lint: Run Go vet checks
lint:
	@echo "==> Running go vet..."
	go vet ./... ./internal/typescriptgo/...

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
	rm -f $(BINARY_NAME) $(ALIAS_NAME)
	rm -rf $(BUILD_DIR)

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -B1 -E '^[a-zA-Z_-]+:' $(MAKEFILE_LIST) | awk '/^##/{desc=$$0; sub(/^## /, "", desc)} /^[a-zA-Z_-]+:/{sub(/:.*/, ""); if (desc != "") {printf "  \033[36m%-18s\033[0m %s\n", $$0, desc; desc=""}}'
