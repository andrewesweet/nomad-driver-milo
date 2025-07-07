PLUGIN_BINARY = nomad-driver-milo
PLUGIN_DIR = /tmp/nomad-plugins

.PHONY: build
build:
	go build -o $(PLUGIN_DIR)/$(PLUGIN_BINARY) .

.PHONY: test
test:
	go test -v ./milo/...

.PHONY: test-acceptance
test-acceptance:
	go test -v ./milo/... -run TestAcceptance

.PHONY: test-integration
test-integration: test-artifacts
	go test -v ./tests/integration/...

.PHONY: test-live-e2e
test-live-e2e:
	go test -v ./e2e/live/...

.PHONY: test-all
test-all: test test-acceptance test-live-e2e test-integration

.PHONY: test-artifacts
test-artifacts:
	@echo "Building test artifacts..."
	@cd tests/fixtures && \
		javac src/*.java -d . && \
		jar cf hello-world.jar HelloWorld.class && \
		jar cf long-running.jar LongRunning.class && \
		jar cf exit-code-test.jar ExitCodeTest.class
	@echo "Test artifacts built successfully"

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run --enable gosec --enable testifylint

.PHONY: clean
clean:
	rm -rf $(PLUGIN_DIR)/$(PLUGIN_BINARY)
	rm -rf tests/fixtures/*.jar
	rm -rf tests/fixtures/*.class

.PHONY: run
run: build
	nomad agent -dev -plugin-dir=$(PLUGIN_DIR) -config=example/agent.hcl

.PHONY: full
full: clean fmt vet lint build test-all