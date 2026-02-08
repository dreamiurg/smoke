.PHONY: build install test lint fmt clean coverage help setup-hooks ci tidy-check vulncheck complexity-check

# Binary name
BINARY=smoke

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOIMPORTS=goimports
GOMOD=$(GOCMD) mod

# Linker flags for version injection
# Support both normal repos and bare repo + worktree setups
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || GIT_WORK_TREE=. GIT_DIR=.git git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || GIT_WORK_TREE=. GIT_DIR=.git git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-buildvcs=false -ldflags "-X github.com/dreamiurg/smoke/internal/cli.Version=$(VERSION) -X github.com/dreamiurg/smoke/internal/cli.Commit=$(COMMIT) -X github.com/dreamiurg/smoke/internal/cli.BuildDate=$(BUILD_DATE)"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/smoke

install: build ## Install the binary to $GOPATH/bin
	$(GOCMD) install $(LDFLAGS) ./cmd/smoke

test: ## Run unit tests
	$(GOTEST) -v -race ./...

test-short: ## Run unit tests (short mode)
	$(GOTEST) -v -short ./...

coverage: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

coverage-check: ## Check coverage meets 70% threshold (MUST), aim for 80% (SHOULD)
	@$(GOTEST) -coverprofile=coverage.out ./... > /dev/null 2>&1
	@COVERAGE=$$($(GOCMD) tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 70" | bc -l) -eq 1 ]; then \
		echo "FAIL: Coverage below 70% threshold"; \
		exit 1; \
	fi

lint: ## Run golangci-lint (includes vet, imports check, etc.)
	golangci-lint run ./...

complexity-check: ## Check cyclomatic complexity thresholds (CCN<=15, length<=60, params<=5)
	@command -v lizard >/dev/null 2>&1 || { echo "Error: lizard not found. Install with: pipx install lizard"; exit 1; }
	lizard -l go -C 15 -L 60 -a 5 -w -x "*_test.go" .

fmt: ## Format code with goimports (matches CI)
	$(GOIMPORTS) -l -w .

fmt-check: ## Check if code is formatted
	@test -z "$$($(GOIMPORTS) -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

tidy: ## Tidy go modules
	$(GOMOD) tidy

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

all: fmt lint test build ## Run all checks and build

pre-commit: fmt-check lint complexity-check ## Pre-commit checks (format, lint, complexity)

pre-push: pre-commit test ## Pre-push checks (format, lint, tests)

ci: fmt-check tidy-check lint complexity-check build test coverage-check ## Run full CI pipeline locally
	@echo "All CI checks passed!"

tidy-check: ## Check if go.mod is tidy
	@$(GOMOD) tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod/go.sum not tidy, run 'make tidy'" && exit 1)

vulncheck: ## Run govulncheck for dependency vulnerabilities
	@command -v govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

setup-hooks: ## Install pre-commit hooks (pre-commit and pre-push stages)
	@command -v pre-commit >/dev/null 2>&1 || { echo "Error: pre-commit not found. Install with: pipx install pre-commit"; exit 1; }
	pre-commit install
	pre-commit install --hook-type pre-push
	@echo "Pre-commit hooks installed (pre-commit + pre-push stages)."
