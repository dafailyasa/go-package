.PHONY: help fmt vet test test-short test-race coverage coverage-html coverage-func benchmark clean

GO := go
PKG := ./...

COVERAGE_DIR := .coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

fmt: ## Format Go source
	$(GO) fmt $(PKG)

vet: ## Run go vet
	$(GO) vet $(PKG)

test: ## Run all unit tests
	$(GO) test -v $(PKG)

test-short: ## Run short tests
	$(GO) test -short -v $(PKG)

test-race: ## Run tests with race detector
	$(GO) test -race -v $(PKG)

coverage: ## Generate coverage report
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -covermode=atomic -coverprofile=$(COVERAGE_FILE) $(PKG)
	$(GO) tool cover -func=$(COVERAGE_FILE)

coverage-html: coverage ## Generate HTML coverage report
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"

coverage-func: ## Show coverage summary only
	$(GO) tool cover -func=$(COVERAGE_FILE)

benchmark: ## Run benchmarks
	$(GO) test -bench=. -benchmem $(PKG)

clean: ## Remove generated files
	rm -rf $(COVERAGE_DIR)