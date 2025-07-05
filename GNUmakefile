PLUGIN_NAME=nomad-driver-milo
NOMAD_PLUGIN_DIR ?= /tmp/nomad-plugins
PLUGIN_BINARY=${NOMAD_PLUGIN_DIR}/${PLUGIN_NAME}
export GO111MODULE=on

default: build

.PHONY: clean test test-unit test-acceptance build help

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

clean: ## Remove build artifacts
	@echo "==> Cleaning plugin from ${NOMAD_PLUGIN_DIR}"
	rm -f ${PLUGIN_BINARY}

test: test-unit test-acceptance ## Run all tests

test-unit: ## Run unit tests
	go test -v ./...

test-acceptance: ## Run acceptance tests using godog
	go test -v -tags=acceptance .

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: fullbuild
fullbuild: clean fmt vet lint test build ## Run complete build process

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
