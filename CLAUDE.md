# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a Nomad driver plugin project called "nomad-driver-milo" (originally based on the HashiCorp skeleton driver template). It implements a custom task driver for the Nomad scheduler that can execute tasks with configurable shell commands and greeting messages. The project is evolving to support Java JAR execution via container runtimes.

## Project History

Originally forked from HashiCorp's skeleton driver template at github.com/hashicorp/nomad-skeleton-driver-plugin. Module paths have been updated from the skeleton template to `github.com/andrewesweet/nomad-driver-milo`.

## Future Direction

The driver is planned to evolve into a Java JAR task driver (see docs/spike-001-minimal-jar.md) that will:
- Execute Java JARs inside crun containers
- Mount host Java runtime into containers  
- Support artifact downloading and validation
- Provide proper resource isolation and security
- Integrate with standard Nomad workflows (CLI, UI, API)

## Build and Development Commands

### Build
- `make build` - Build the plugin binary (hello-driver)
- `make fullbuild` - Complete build process: clean, format, vet, lint, then build

### Code Quality
- `make fmt` - Format Go code using `go fmt ./...`
- `make vet` - Run `go vet ./...` to check for issues
- `make lint` - Run golangci-lint with gosec and testifylint enabled
- `make clean` - Remove build artifacts

### Development Environment
- Go 1.18+ required (1.24+ recommended)
- Nomad v0.9+ for running the plugin
- golangci-lint v1.55.2+ for linting with gosec security checks
- Uses Go modules (GO111MODULE=on)

### Development Setup
1. Install Go 1.24 or later
2. Install Make
3. Install golangci-lint v1.55.2 or later
4. Clone the repository
5. Run `make fullbuild` to verify setup

### Testing the Plugin Locally

#### Quick Development Mode (Recommended)
```bash
# Build and start Nomad dev server with the plugin in one command
make run

# In another shell:
nomad run ./example/example-dev.nomad
```

#### Manual Testing
```bash
# Build the plugin to /tmp/nomad-plugins/
make build

# Start Nomad with the plugin
nomad agent -dev -config=./example/agent.hcl -plugin-dir=/tmp/nomad-plugins

# In another shell:
nomad run ./example/example.nomad
```

## Architecture

### Core Components

**main.go**: Entry point that serves the plugin using Nomad's plugin framework. Contains the factory function that creates new plugin instances.

**hello/driver.go**: Main driver implementation with key components:
- `HelloDriverPlugin` struct: Core driver with task management, configuration, and Nomad integration
- Configuration schemas: `configSpec` (agent-level) and `taskConfigSpec` (task-level) using HCL specs
- Plugin capabilities: Currently supports signal sending but not exec
- Fingerprinting: Health checks and node attribute reporting
- Task lifecycle: StartTask, StopTask, DestroyTask, RecoverTask operations

**hello/handle.go**: Task handle management:
- `taskHandle` struct: Stores runtime information for individual tasks (PID, executor, state, timing)
- Task status reporting and state management
- Executor integration for process management

**hello/state.go**: Task storage:
- `taskStore`: Thread-safe in-memory storage for active task handles
- Task ID to handle mapping with RWMutex protection

### Key Patterns

- Uses HashiCorp's executor framework for process management
- Plugin follows Nomad's driver plugin interface requirements
- Configuration uses HCL specs for validation
- Thread-safe task management with proper locking
- Graceful shutdown handling with context cancellation

### Configuration

**Agent-level**: Shell selection (bash/fish) via `plugin` stanza in agent.hcl
**Task-level**: Greeting message via task `config` stanza in job specification

The driver executes shell commands that echo the configured greeting message using the specified shell.

### Example Files

**example/agent.hcl**: Nomad agent configuration with plugin settings
**example/example.nomad**: Sample job specification showing driver usage

## Planned Evolution to Java JAR Driver

The spike document (docs/spike-001-minimal-jar.md) outlines the planned minimal viable implementation:

### Core Features
1. **JAR Execution via crun**: Container runtime integration for process isolation
2. **Host Java Mounting**: Mount system Java runtime into containers
3. **Artifact Support**: Download and validate JAR files via Nomad's artifact stanza
4. **Logging Integration**: Stream JAR output to Nomad logs
5. **Error Handling**: Graceful failures with clear error messages

### Success Criteria
- User submits job with `driver = "milo"` and artifact block
- JAR executes in isolated crun container
- Output visible in Nomad logs
- Standard Nomad operations work (stop, restart, etc.)

## Module Information

- Module path: `github.com/andrewesweet/nomad-driver-milo`
- Nomad version: v1.10.0
- Uses replace directive for go-metrics compatibility
- Go version: 1.24 (with toolchain 1.24.2)

## Additional Resources

- See CONTRIBUTING.md for detailed development workflow
- GitHub Actions workflow in .github/ for CI/CD setup
- Linting configuration in .golangci.yml with security scanning enabled