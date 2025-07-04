# Contributing to Nomad Driver Milo

Thank you for your interest in contributing to this project! This guide will help you get started with development and testing.

## Prerequisites

- Go 1.24 or later
- Git
- Make

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/andrewesweet/nomad-driver-milo.git
   cd nomad-driver-milo
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the project:
   ```bash
   make build
   ```

## Testing

This project uses two types of tests:

### Unit Tests

Unit tests are written using Go's built-in testing framework with testify for assertions. They test individual components in isolation.

To run unit tests:
```bash
make test-unit
```

Or directly with Go:
```bash
go test -v ./...
```

Unit tests are located alongside the code they test, following the Go convention of `*_test.go` files.

### Acceptance Tests

Acceptance tests are written using [Godog](https://github.com/cucumber/godog), a Cucumber-style BDD testing framework for Go. These tests validate the behavior of the system from a user's perspective.

To run acceptance tests:
```bash
make test-acceptance
```

Or directly with Go:
```bash
go test -v -tags=acceptance .
```

Acceptance tests are defined in:
- **Feature files**: `features/*.feature` - Written in Gherkin syntax
- **Step definitions**: `acceptance_test.go` - Go code that implements the test steps

### Running All Tests

To run both unit and acceptance tests:
```bash
make test
```

### Test Coverage

To run tests with coverage:
```bash
go test -cover ./...
```

For detailed coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Building

The build process automatically runs all tests before building the binary:
```bash
make build
```

If you want to build without running tests (not recommended for production):
```bash
go build -o hello-driver .
```

## Code Style

- Follow standard Go formatting using `go fmt`
- Use `go vet` to check for common issues
- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Include tests for new functionality

## Submitting Changes

1. Create a new branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and add tests

3. Run the full test suite:
   ```bash
   make test
   ```

4. Ensure the build passes:
   ```bash
   make build
   ```

5. Commit your changes with a clear message:
   ```bash
   git commit -m "Add feature: description of your changes"
   ```

6. Push your branch and create a pull request

## Adding Tests

### Adding Unit Tests

1. Create or update `*_test.go` files in the same package as the code you're testing
2. Write test functions that start with `Test`
3. Use testify assertions for clear test output:
   ```go
   func TestMyFunction(t *testing.T) {
       result := MyFunction("input")
       assert.Equal(t, "expected", result)
   }
   ```

### Adding Acceptance Tests

1. Add scenarios to feature files in the `features/` directory:
   ```gherkin
   Scenario: New functionality
     Given I have the prerequisites
     When I perform an action
     Then I should see the expected result
   ```

2. Implement step definitions in `acceptance_test.go`:
   ```go
   func (ctx *AcceptanceTestContext) iPerformAnAction() error {
       // Implementation
       return nil
   }
   ```

3. Register the step in the `InitializeScenario` function:
   ```go
   ctx.Step(`^I perform an action$`, testCtx.iPerformAnAction)
   ```

## Available Make Targets

- `make help` - Display available targets
- `make build` - Build the plugin binary (runs tests first)
- `make test` - Run all tests
- `make test-unit` - Run unit tests only
- `make test-acceptance` - Run acceptance tests only
- `make clean` - Remove build artifacts

## Debugging Tests

To run tests with verbose output:
```bash
go test -v ./...
```

To run a specific test:
```bash
go test -v -run TestSpecificFunction ./...
```

To run acceptance tests with debugging:
```bash
go test -v -tags=acceptance . -godog.format=progress
```

## Getting Help

If you have questions or run into issues:

1. Check the existing issues in the repository
2. Review the project documentation
3. Create a new issue with details about your problem

Thank you for contributing!