PLUGIN_NAME=nomad-driver-milo
NOMAD_PLUGIN_DIR ?= /tmp/nomad-plugins
PLUGIN_BINARY=${NOMAD_PLUGIN_DIR}/${PLUGIN_NAME}
export GO111MODULE=on

default: build

.PHONY: clean test test-unit test-acceptance test-bdd test-live-e2e test-all build help

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

clean: ## Remove build artifacts
	@echo "==> Cleaning plugin from ${NOMAD_PLUGIN_DIR}"
	rm -f ${PLUGIN_BINARY}

test: test-unit test-acceptance ## Run core tests (unit and acceptance)

test-all: test-unit test-acceptance test-bdd test-live-e2e ## Run all tests including BDD and live e2e

test-unit: ## Run unit tests
	go test -v ./...

test-acceptance: ## Run acceptance tests using godog
	go test -v -tags=acceptance .

test-bdd: build ## Run BDD acceptance tests with real driver integration
	@echo "==> Running BDD acceptance tests..."
	go test -v ./features -timeout 30m

test-bdd-focus: build ## Run focused BDD test (use SCENARIO env var)
	@echo "==> Running focused BDD test..."
	go test -v ./features -timeout 30m -run "$(SCENARIO)"

test-live-e2e: build ## Run live integration e2e tests
	@echo "==> Running live e2e tests (requires nomad binary)"
	go test -v -tags=live_e2e ./e2e/live/...

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: full
full: clean fmt vet lint build test-all ## Run complete build process

build: test ## Build the plugin binary (runs tests first)
	@echo "==> Building ${PLUGIN_NAME} plugin"
	@mkdir -p ${NOMAD_PLUGIN_DIR}
	go build -o ${PLUGIN_BINARY} .
	@echo "==> Plugin built at ${PLUGIN_BINARY}"

.PHONY: run
run: build ## Build plugin and run Nomad dev server
run:
	@echo "==> Starting Nomad dev server with ${PLUGIN_NAME} plugin"
	@echo "==> Plugin directory: ${NOMAD_PLUGIN_DIR}"
	@echo "==> Once started, run 'nomad run example/example-dev.nomad' in another terminal"
	nomad agent -dev -config=example/agent-dev.hcl -plugin-dir=${NOMAD_PLUGIN_DIR}
