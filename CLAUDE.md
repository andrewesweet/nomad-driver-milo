# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) and Gemini Code Assist when working with code in this repository.

## AI Tool Usage

### Gemini Integration

For Claude Code, the `gemini` command is available for complex analysis, planning, and debugging:

```bash
# Use Gemini 2.5 Pro (default model, 1M token context)
gemini -p "Your prompt here"

# Include all files in context for full codebase analysis
gemini -a -p "Analyze the entire codebase structure"

# Enable debug mode for troubleshooting
gemini -d -p "Debug this issue"
```

**When to use Gemini:**
- Full codebase analysis and understanding
- Complex problem solving and planning
- Debugging intricate issues
- Architecture reviews
- Long-form analysis requiring large context

**Gemini capabilities:**
- 1 million token context window (excellent for large codebases)
- Reads CLAUDE.md automatically for project context
- Superior at planning and problem-solving
- Can call Claude for additional opinions when needed

### Claude Integration

For Gemini, the `claude` command provides access to different Claude models:

```bash
# Use Sonnet 4 (default, excellent for coding)
claude -p "Your prompt here"

# Use Opus 4 for complex planning and problem-solving
claude --model opus -p "Your prompt here"

# Print mode for scripting/automation
claude -p "Your prompt here" --print
```

**Model Selection:**
- **Sonnet 4**: Best for coding execution, efficient, 200k context
- **Opus 4**: Superior for planning and complex problem-solving, more expensive, 200k context
- Both models are excellent, choose based on task complexity and cost considerations

**When to use Claude:**
- Code generation and editing
- Quick problem-solving
- Integration with existing workflows
- When you need a second opinion on Gemini's suggestions

### Context Management

**Important**: Each `claude -p` and `gemini -p` call starts with **empty context** - no memory of previous calls or conversations. Interactive Claude Code sessions maintain persistent context throughout the entire conversation.

**Context sharing limitations:**
- No built-in context sharing between separate command calls
- Each invocation is stateless and independent
- Manual context passing required (but hits token limits quickly)

### Best Practices

1. **For comprehensive analysis**: Run Gemini from the project root directory with full codebase context:
   ```bash
   cd /path/to/nomad-driver-milo
   gemini -a -p "Analyze the entire codebase and suggest improvements"
   ```

2. **Context-aware workflows:**
   - Use interactive Claude Code sessions for context-heavy work that builds on previous exchanges
   - Use `claude -p` or `gemini -p` for isolated, specific tasks that don't require conversation history
   - Consider copying relevant context manually when using one-off commands

3. **Optimal workflow patterns:**
   - **Planning**: Use Gemini with `-a` flag for comprehensive codebase analysis and planning
   - **Execution**: Use interactive Claude Code sessions to maintain context during implementation
   - **Isolated tasks**: Use `claude -p` or `gemini -p` for specific, standalone questions

4. **Model selection**: Choose the right model for the task (Sonnet for coding, Opus for planning)
5. **Cross-consultation**: Gemini can call Claude for additional perspectives
6. **Cost awareness**: Be mindful of Opus costs for routine tasks

## Repository Overview

This is a Nomad driver plugin called "nomad-driver-milo" that executes Java JAR files in isolated containers using crun. It provides secure, isolated execution of Java applications with Nomad's orchestration capabilities.

## Project History

Originally forked from HashiCorp's skeleton driver template at github.com/hashicorp/nomad-skeleton-driver-plugin. The project has evolved significantly from the template to become a specialized Java JAR execution driver.

## Current Capabilities

The Milo driver provides:
- **Java JAR Execution**: Runs Java JAR files inside OCI containers via crun
- **Automatic Java Detection**: Discovers Java installations on the host system
- **Container Isolation**: Uses Linux namespaces (PID, IPC, UTS, mount) for security
- **Host Java Runtime Mounting**: Mounts system Java into containers to avoid bundling
- **Artifact Support**: Downloads and validates JAR files via Nomad's artifact stanza
- **Standard Nomad Integration**: Works with Nomad CLI, UI, and API

## Build and Development Commands

### Build
- `make build` - Build the plugin binary to `/tmp/nomad-plugins/`
- `make full` - Run complete build process (clean, fmt, vet, lint, build, test-all)

### Code Quality
- `make fmt` - Format Go code using `go fmt ./...`
- `make vet` - Run `go vet ./...` to check for issues
- `make lint` - Run golangci-lint with gosec and testifylint enabled
- `make clean` - Remove build artifacts

### Testing
- `make test` - Run unit tests
- `make test-acceptance` - Run acceptance tests
- `make test-live-e2e` - Run end-to-end tests (requires running Nomad)
- `make run` - Build and start Nomad dev server with the plugin

### Development Environment
- Go 1.24+ required
- Nomad v1.10+ for running the plugin
- golangci-lint v1.55.2+ for linting with gosec security checks
- crun (OCI runtime) for container execution
- Java runtime installed on the host

### Testing the Plugin Locally

#### Quick Development Mode (Recommended)
```bash
# Build and start Nomad dev server with the plugin in one command
make run

# In another shell, submit a test job:
nomad run ./example/hello-world.nomad
```

#### Example Job Specification
```hcl
job "milo-hello-world" {
  datacenters = ["dc1"]
  type = "batch"

  group "hello" {
    task "java-hello" {
      driver = "milo"

      config {
        dummy = ""  # Currently unused, required by schema
      }

      artifact {
        source = "https://github.com/andrewesweet/nomad-driver-milo/raw/main/test-artifacts/hello-world.jar"
        destination = "local/"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
```

## Architecture

### Core Components

**main.go**: Entry point that serves the plugin using Nomad's plugin framework.

**milo/driver.go**: Main driver implementation:
- `MiloDriverPlugin` struct: Core driver with task management and Nomad integration
- Plugin registration and capabilities
- Fingerprinting for Java runtime detection
- Task lifecycle management (Start, Stop, Destroy, Recover)

**milo/handle.go**: Task handle management:
- `taskHandle` struct: Stores runtime state for tasks
- Container process monitoring
- Status reporting to Nomad

**milo/state.go**: Thread-safe task storage with concurrent access protection

**milo/java.go**: Java runtime detection:
- Searches common Java installation paths
- Validates Java executable availability
- Reports Java version during fingerprinting

**milo/container.go**: OCI container specification:
- Creates container specs for crun
- Configures namespace isolation
- Sets up filesystem mounts (Java runtime, task directory, system libraries)

**milo/artifact.go**: JAR file handling:
- Discovers JAR files in task directory
- Validates artifact presence
- Prepares JAR paths for execution

### Key Implementation Details

- Uses crun as the OCI runtime for container execution
- Mounts host Java runtime to avoid bundling Java in containers
- Implements standard Nomad driver plugin interface
- Thread-safe task management with proper locking
- Graceful shutdown with context cancellation

### Container Isolation

Each task runs in an isolated container with:
- **Namespaces**: PID, IPC, UTS, Mount isolation
- **Filesystem**: Read-only mounts except task directory
- **Java Runtime**: Mounted from host at `/usr/lib/jvm/java`
- **Task Directory**: Mounted at `/app` with read-write access

## Module Information

- Module path: `github.com/andrewesweet/nomad-driver-milo`
- Nomad SDK: v1.10.0
- Go version: 1.24 (with toolchain 1.24.2)
- Uses replace directive for go-metrics compatibility

## Testing

### Unit Tests
Located in `milo/*_test.go`, covering:
- Driver lifecycle
- Task management
- Java detection
- Container specification generation

### End-to-End Tests
Located in `e2e/live/`, testing:
- Full job submission and execution
- JAR artifact handling
- Container isolation
- Resource limits

### Test Artifacts
- `test-artifacts/hello-world.jar`: Simple test JAR for e2e testing
- Example job files in root directory for manual testing

## Additional Resources

- See CONTRIBUTING.md for detailed development workflow
- GitHub Actions workflow in .github/ for CI/CD setup
- Linting configuration in .golangci.yml with security scanning
- Spike documents in docs/ for design decisions

## Important Instruction Reminders

Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.