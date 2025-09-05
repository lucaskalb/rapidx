# Makefile to run the same CI steps locally
# Replicates GitHub Actions workflows: ci.yml, code-quality.yml, pr-validation.yml

.PHONY: help install-tools test test-race test-coverage test-demo test-examples build vet fmt staticcheck lint security tidy clean ci pr-validate

# Variables
GO_VERSION := 1.24
COVERAGE_THRESHOLD := 80
TIMEOUT := 5m

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## Show this help
	@echo "$(BLUE)Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install tools necessary for CI
	@echo "$(BLUE)Installing tools...$(NC)"
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)✅ Tools installed$(NC)"

# =============================================================================
# TESTS (CI Job: test)
# =============================================================================

test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)✅ Tests executed successfully$(NC)"

test-race: ## Run tests with race condition detection
	@echo "$(BLUE)Running tests with race detection...$(NC)"
	@go test -race -v ./...
	@echo "$(GREEN)✅ Tests with race detection executed$(NC)"

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total
	@echo "$(GREEN)✅ Test coverage generated in coverage.out$(NC)"

test-demo: ## Run demonstration tests (expects failures)
	@echo "$(BLUE)Running demonstration tests...$(NC)"
	@go test -tags demo -v ./testfailures/demo/... || echo "$(YELLOW)⚠️  Demo tests failed (expected)$(NC)"

test-examples: ## Run example tests (expects failures)
	@echo "$(BLUE)Running example tests...$(NC)"
	@go test -tags examples -v ./docs/examples/... || echo "$(YELLOW)⚠️  Example tests failed (expected)$(NC)"

# =============================================================================
# BUILD AND VERIFICATIONS (CI Job: build)
# =============================================================================

build: ## Compile the project
	@echo "$(BLUE)Compiling project...$(NC)"
	@go build -v ./...
	@echo "$(GREEN)✅ Project compiled successfully$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ go vet passed$(NC)"

fmt: ## Check code formatting
	@echo "$(BLUE)Checking formatting...$(NC)"
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "$(RED)❌ Code is not formatted. Run 'go fmt ./...' to fix.$(NC)"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Code is properly formatted$(NC)"

fmt-fix: ## Fix code formatting
	@echo "$(BLUE)Fixing formatting...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Formatting fixed$(NC)"

staticcheck: ## Run staticcheck
	@echo "$(BLUE)Running staticcheck...$(NC)"
	@staticcheck ./...
	@echo "$(GREEN)✅ staticcheck passed$(NC)"

# =============================================================================
# LINTING (CI Job: lint)
# =============================================================================

lint: ## Run golangci-lint
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	@golangci-lint run --timeout=$(TIMEOUT)
	@echo "$(GREEN)✅ golangci-lint passed$(NC)"

lint-fix: ## Run golangci-lint with automatic fixes
	@echo "$(BLUE)Running golangci-lint with fixes...$(NC)"
	@golangci-lint run --timeout=$(TIMEOUT) --fix
	@echo "$(GREEN)✅ golangci-lint with fixes executed$(NC)"

# =============================================================================
# SECURITY (CI Job: security)
# =============================================================================

security: ## Run security verification
	@echo "$(BLUE)Running security verification...$(NC)"
	@gosec -no-fail -fmt sarif -out gosec.sarif ./...
	@echo "$(GREEN)✅ Security verification executed$(NC)"

# =============================================================================
# CODE QUALITY (CI Job: code-quality)
# =============================================================================

check-license: ## Check license headers
	@echo "$(BLUE)Checking license headers...$(NC)"
	@find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read file; do \
		if ! head -n 1 "$$file" | grep -q "Copyright\|License"; then \
			echo "$(YELLOW)⚠️  File $$file may be missing license header$(NC)"; \
		fi; \
	done
	@echo "$(GREEN)✅ License verification completed$(NC)"

check-todos: ## Check TODO/FIXME comments
	@echo "$(BLUE)Checking TODO/FIXME comments...$(NC)"
	@if grep -r "TODO\|FIXME" --include="*.go" . | grep -v "testfailures/" | grep -v "docs/"; then \
		echo "$(YELLOW)⚠️  TODO/FIXME comments found in production code$(NC)"; \
		echo "Consider resolving these before merging"; \
	fi
	@echo "$(GREEN)✅ TODO verification completed$(NC)"

tidy: ## Check if dependencies are organized
	@echo "$(BLUE)Checking dependencies...$(NC)"
	@go mod tidy
	@if ! git diff --exit-code go.mod go.sum; then \
		echo "$(RED)❌ Dependencies are not organized. Run 'go mod tidy' to fix.$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Dependencies are organized$(NC)"

# =============================================================================
# PR VALIDATION (CI Job: pr-validation)
# =============================================================================

check-coverage: ## Check if coverage is above threshold
	@echo "$(BLUE)Checking test coverage...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Test coverage: $${COVERAGE}%"; \
	if (( $$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l) )); then \
		echo "$(RED)❌ Test coverage is below $(COVERAGE_THRESHOLD)%$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)✅ Test coverage is above $(COVERAGE_THRESHOLD)%$(NC)"

check-adr: ## Check ADR changes
	@echo "$(BLUE)Checking ADR changes...$(NC)"
	@if git diff --name-only origin/main...HEAD 2>/dev/null | grep -q "docs/adr-.*\.md"; then \
		echo "$(YELLOW)📋 ADR changes detected$(NC)"; \
		echo "Please ensure that:"; \
		echo "- ADR follows the template format"; \
		echo "- Status is clearly marked"; \
		echo "- Context and consequences are well documented"; \
	fi
	@echo "$(GREEN)✅ ADR verification completed$(NC)"

# =============================================================================
# COMPOSITE COMMANDS
# =============================================================================

ci: install-tools tidy vet fmt staticcheck lint security test test-race test-coverage build ## Run complete CI pipeline
	@echo "$(GREEN)🎉 CI pipeline executed successfully!$(NC)"

pr-validate: tidy vet fmt lint security test check-coverage check-license check-todos check-adr ## Run complete PR validation
	@echo "$(GREEN)🎉 PR validation executed successfully!$(NC)"

quality: vet fmt staticcheck lint security check-license check-todos ## Run all quality checks
	@echo "$(GREEN)🎉 Quality checks executed successfully!$(NC)"

# =============================================================================
# UTILITIES
# =============================================================================

clean: ## Clean generated files
	@echo "$(BLUE)Cleaning generated files...$(NC)"
	@rm -f coverage.out
	@rm -f gosec.sarif
	@go clean -cache
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

deps: ## Download and verify dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✅ Dependencies downloaded and verified$(NC)"

version: ## Show Go version
	@echo "$(BLUE)Go version:$(NC)"
	@go version

# =============================================================================
# DEVELOPMENT
# =============================================================================

dev-setup: deps install-tools ## Setup development environment
	@echo "$(GREEN)🎉 Development environment configured!$(NC)"
	@echo "$(BLUE)Useful commands:$(NC)"
	@echo "  make test        - Run tests"
	@echo "  make lint        - Run linting"
	@echo "  make ci          - Run complete pipeline"
	@echo "  make pr-validate - Validate PR"

watch-test: ## Run tests in watch mode (requires entr)
	@echo "$(BLUE)Watch mode activated. Press Ctrl+C to exit.$(NC)"
	@find . -name "*.go" | entr -c make test

# =============================================================================
# INFORMATION
# =============================================================================

info: ## Show project information
	@echo "$(BLUE)Project information:$(NC)"
	@echo "  Go version: $$(go version)"
	@echo "  Project path: $$(pwd)"
	@echo "  Go modules: $$(go list -m all | wc -l) modules"
	@echo "  Test files: $$(find . -name "*_test.go" | wc -l) files"
	@echo "  Go files: $$(find . -name "*.go" | wc -l) files"