PLUGIN_BINARY=hello-driver
export GO111MODULE=on

default: build

.PHONY: clean test test-unit test-acceptance build help

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

clean: ## Remove build artifacts
	rm -rf ${PLUGIN_BINARY}

test: test-unit test-acceptance ## Run all tests

test-unit: ## Run unit tests
	go test -v ./...

test-acceptance: ## Run acceptance tests using godog
	go test -v -tags=acceptance .

build: test ## Build the plugin binary (runs tests first)
	go build -o ${PLUGIN_BINARY} .
